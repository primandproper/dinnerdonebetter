package mcpserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	waitlistsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"

	"github.com/primandproper/platform-go/v13/authentication/oauth2server"
	"github.com/primandproper/platform-go/v13/encoding"
	issuereports "github.com/primandproper/platform-go/v13/issuereports"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/routing"
	routingcfg "github.com/primandproper/platform-go/v13/routing/config"
	"github.com/primandproper/platform-go/v13/version"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultMcpServerConfigurationFilepath = "deploy/environments/localdev/config_files/mcp_server_config.json"

	// TransportStdio serves MCP over stdin and stdout.
	TransportStdio = "stdio"
	// TransportSSE serves MCP over server-sent events.
	TransportSSE = "sse"
	// TransportHTTP serves MCP over streamable HTTP.
	TransportHTTP = "http"

	// DefaultBaseURL is the public base URL assumed when the caller supplies none.
	DefaultBaseURL = "http://localhost:8888"

	defaultPort = 8888

	// shutdownTimeout bounds the drain: how long in-flight requests have to finish once
	// the server has been told to stop, and how long the container has to release its
	// database pool afterwards.
	shutdownTimeout = 30 * time.Second
)

// ValidTransports returns every transport Run accepts.
func ValidTransports() []string {
	return []string{TransportStdio, TransportSSE, TransportHTTP}
}

// Run serves the MCP server over the named transport, blocking until it is signaled to stop.
// baseURL is the server's public address, which the OAuth2 metadata documents advertise.
func Run(ctx context.Context, transport, baseURL string) error {
	if !slices.Contains(ValidTransports(), transport) {
		return fmt.Errorf("invalid transport method %q: allowed values are %s", transport, strings.Join(ValidTransports(), ", "))
	}

	configFilepath := os.Getenv(config.ConfigurationFilePathEnvVarKey)
	if configFilepath == "" {
		configFilepath = defaultMcpServerConfigurationFilepath
	}

	cfg, err := config.LoadConfigFromPath[config.MCPServiceConfig](configFilepath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	svc, err := NewService(ctx, cfg, baseURL)
	if err != nil {
		return err
	}

	logger := svc.pillars.Logger

	// Whichever transport serves, the container's database pool is released on the way
	// out. What this replaced was a goroutine that called os.Exit on a signal, which
	// released nothing, reported nothing, and could not be run inside a test process at
	// all — the first thing in the way of ever exercising this server as a server.
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
		defer cancel()

		if shutdownErr := svc.Shutdown(shutdownCtx); shutdownErr != nil {
			logger.Error("shutting down MCP service", shutdownErr)
		}
	}()

	logger.WithValue("transport", transport).Info("serving MCP server")

	if transport == TransportStdio {
		if err = svc.ServeStdio(ctx); err != nil {
			return fmt.Errorf("serving MCP server via stdio: %w", err)
		}

		return nil
	}

	handler, err := svc.Handler(ctx, transport)
	if err != nil {
		return err
	}

	port := cfg.HTTPServer.Port
	if port == 0 {
		port = defaultPort
	}

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           handler,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)
	defer signal.Stop(signalChan)

	serveErrors := make(chan error, 1)
	go func() {
		serveErrors <- srv.ListenAndServe()
	}()

	select {
	case err = <-serveErrors:
		// ErrServerClosed cannot arrive here — nothing has called Shutdown yet — so
		// whatever this is, the server stopped serving on its own.
		return fmt.Errorf("serving MCP server over %s: %w", transport, err)
	case sig := <-signalChan:
		logger.WithValue("signal", sig.String()).Info("stopping MCP server")
	case <-ctx.Done():
		logger.Info("stopping MCP server: context canceled")
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()

	if err = srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutting down HTTP server: %w", err)
	}

	return nil
}

// buildRouter creates a router with OAuth2 routes (unauthenticated) and the MCP handler (authenticated).
func buildRouter(ctx context.Context, mcpHandler http.Handler, authServer *oauth2server.Server, resourceMetadata *oauth2server.ResourceMetadata, pillars *observability.Pillars, routingCfg *routingcfg.Config, baseURL string) (*routing.Router, error) {
	encoder := encoding.NewServerEncoderDecoder(encoding.ContentTypeJSON, encoding.WithLogger(pillars.Logger), encoding.WithTracerProvider(pillars.TracerProvider))

	router, err := routingcfg.NewRouter(ctx, routingCfg, encoder, routingcfg.WithPillars(pillars))
	if err != nil {
		return nil, err
	}

	// Ops routes (unauthenticated).
	router.Group("/_ops_", func(opsRouter *routing.Router) {
		opsRouter.Handle(http.MethodGet, "/live", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusOK)
		}))
		opsRouter.Handle(http.MethodGet, "/ready", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusOK)
		}))
		opsRouter.Handle(http.MethodGet, "/version", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.Header().Set("Content-Type", "application/json")
			encoder.EncodeResponseWithStatus(req.Context(), res, version.Get(), http.StatusOK)
		}))
	})

	// The six authorization server endpoints, plus the protected resource document.
	// No auth middleware: these are how a caller gets a token in the first place.
	authServer.Mount(router)
	resourceMetadata.Mount(router)

	// Wrap the MCP handler with bearer token auth middleware. The MCP transport
	// serves multiple methods (GET for streaming, POST for messages, DELETE to
	// terminate a session), so register the handler for each.
	authMiddleware := auth.RequireBearerToken(newTokenVerifier(authServer, baseURL), &auth.RequireBearerTokenOptions{
		ResourceMetadataURL: baseURL + oauth2server.PathProtectedResourceMetadata,
	})
	mcpWrapped := authMiddleware(mcpHandler)
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		router.Handle(method, "/mcp", mcpWrapped)
	}

	if err = router.Err(); err != nil {
		return nil, err
	}

	return router, nil
}

type mcpToolManager struct {
	mealplanningRepo mealplanning.Repository
	webhooksRepo     webhooks.Repository
	waitlistsRepo    *waitlistsrepo.Repository
	issueReports     issuereports.Store
}

// userFromRequest resolves the authenticated user's account from the MCP request's auth token.
//
// The account is read off the token rather than looked up again: it was resolved
// once at /authorize and travels in the access token's Subject claims, so a tool
// call costs the one store read the bearer middleware already made.
func (h *mcpToolManager) userFromRequest(req *mcp.CallToolRequest) (accountID string, err error) {
	if req.Extra == nil || req.Extra.TokenInfo == nil {
		return "", fmt.Errorf("not authenticated")
	}

	accountID, ok := req.Extra.TokenInfo.Extra[claimAccountID].(string)
	if !ok || accountID == "" {
		return "", fmt.Errorf("no account on token")
	}

	return accountID, nil
}

func (h *mcpToolManager) setupServer() *mcp.Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: fmt.Sprintf("%s-mcp", branding.CompanyNameSlug), Version: "v1.0.0"}, nil)

	// Valid Ingredients (read-only)
	mcp.AddTool(mcpServer, getValidIngredientTool, h.GetValidIngredient())
	mcp.AddTool(mcpServer, searchForValidIngredientsTool, h.SearchForValidIngredients())

	// Valid Preparations (read-only)
	mcp.AddTool(mcpServer, getValidPreparationTool, h.GetValidPreparation())
	mcp.AddTool(mcpServer, searchForValidPreparationsTool, h.SearchForValidPreparations())

	// Valid Measurement Units (read-only)
	mcp.AddTool(mcpServer, getValidMeasurementUnitTool, h.GetValidMeasurementUnit())
	mcp.AddTool(mcpServer, searchForValidMeasurementUnitsTool, h.SearchForValidMeasurementUnits())

	// Valid Ingredient Preparations (read-only)
	mcp.AddTool(mcpServer, getValidIngredientPreparationTool, h.GetValidIngredientPreparation())
	mcp.AddTool(mcpServer, getValidIngredientPreparationsTool, h.GetValidIngredientPreparations())

	// Valid Prep Task Configs (read-only)
	mcp.AddTool(mcpServer, getValidPrepTaskConfigTool, h.GetValidPrepTaskConfig())
	mcp.AddTool(mcpServer, getValidPrepTaskConfigsTool, h.GetValidPrepTaskConfigs())
	mcp.AddTool(mcpServer, getValidPrepTaskConfigsByIngredientTool, h.GetValidPrepTaskConfigsByIngredient())
	mcp.AddTool(mcpServer, getValidPrepTaskConfigsByPreparationTool, h.GetValidPrepTaskConfigsByPreparation())
	mcp.AddTool(mcpServer, getValidPrepTaskConfigsByIngredientAndPreparationTool, h.GetValidPrepTaskConfigsByIngredientAndPreparation())

	// Valid Ingredient Measurement Units (read-only)
	mcp.AddTool(mcpServer, getValidIngredientMeasurementUnitTool, h.GetValidIngredientMeasurementUnit())
	mcp.AddTool(mcpServer, getValidIngredientMeasurementUnitsTool, h.GetValidIngredientMeasurementUnits())

	// Valid Vessels (read-only)
	mcp.AddTool(mcpServer, getValidVesselTool, h.GetValidVessel())
	mcp.AddTool(mcpServer, searchForValidVesselsTool, h.SearchForValidVessels())

	// Valid Measurement Unit Conversions (read-only)
	mcp.AddTool(mcpServer, getValidMeasurementUnitConversionTool, h.GetValidMeasurementUnitConversion())
	mcp.AddTool(mcpServer, getValidMeasurementUnitConversionsForUnitTool, h.GetValidMeasurementUnitConversionsForUnit())
	mcp.AddTool(mcpServer, getValidMeasurementUnitConversionsForIngredientsTool, h.GetValidMeasurementUnitConversionsForIngredients())

	// Valid Ingredient States (read-only)
	mcp.AddTool(mcpServer, getValidIngredientStateTool, h.GetValidIngredientState())
	mcp.AddTool(mcpServer, searchForValidIngredientStatesTool, h.SearchForValidIngredientStates())

	// Valid Ingredient State Ingredients (read-only)
	mcp.AddTool(mcpServer, getValidIngredientStateIngredientTool, h.GetValidIngredientStateIngredient())
	mcp.AddTool(mcpServer, getValidIngredientStateIngredientsTool, h.GetValidIngredientStateIngredients())

	// Valid Instruments (read-only)
	mcp.AddTool(mcpServer, getValidInstrumentTool, h.GetValidInstrument())
	mcp.AddTool(mcpServer, searchForValidInstrumentsTool, h.SearchForValidInstruments())

	// Valid Preparation Instruments (read-only)
	mcp.AddTool(mcpServer, getValidPreparationInstrumentTool, h.GetValidPreparationInstrument())
	mcp.AddTool(mcpServer, getValidPreparationInstrumentsTool, h.GetValidPreparationInstruments())

	// Valid Preparation Vessels (read-only)
	mcp.AddTool(mcpServer, getValidPreparationVesselTool, h.GetValidPreparationVessel())
	mcp.AddTool(mcpServer, getValidPreparationVesselsTool, h.GetValidPreparationVessels())

	// Recipe Step Instruments (read-only)
	mcp.AddTool(mcpServer, getRecipeStepInstrumentTool, h.GetRecipeStepInstrument())
	mcp.AddTool(mcpServer, getRecipeStepInstrumentsTool, h.GetRecipeStepInstruments())

	// Recipe Step Products (read-only)
	mcp.AddTool(mcpServer, getRecipeStepProductTool, h.GetRecipeStepProduct())
	mcp.AddTool(mcpServer, getRecipeStepProductsTool, h.GetRecipeStepProducts())

	// Recipe Step Ingredients (read-only)
	mcp.AddTool(mcpServer, getRecipeStepIngredientTool, h.GetRecipeStepIngredient())
	mcp.AddTool(mcpServer, getRecipeStepIngredientsTool, h.GetRecipeStepIngredients())

	// Recipe Prep Tasks (read-only)
	mcp.AddTool(mcpServer, getRecipePrepTaskTool, h.GetRecipePrepTask())
	mcp.AddTool(mcpServer, getRecipePrepTasksTool, h.GetRecipePrepTasks())

	// Recipe Step Vessels (read-only)
	mcp.AddTool(mcpServer, getRecipeStepVesselTool, h.GetRecipeStepVessel())
	mcp.AddTool(mcpServer, getRecipeStepVesselsTool, h.GetRecipeStepVessels())

	// Recipe Step Completion Conditions (read-only)
	mcp.AddTool(mcpServer, getRecipeStepCompletionConditionTool, h.GetRecipeStepCompletionCondition())
	mcp.AddTool(mcpServer, getRecipeStepCompletionConditionsTool, h.GetRecipeStepCompletionConditions())

	// Recipe Steps (read-only)
	mcp.AddTool(mcpServer, getRecipeStepTool, h.GetRecipeStep())
	mcp.AddTool(mcpServer, getRecipeStepsTool, h.GetRecipeSteps())

	// Recipes (read-only)
	mcp.AddTool(mcpServer, getRecipeTool, h.GetRecipe())
	mcp.AddTool(mcpServer, getRecipesTool, h.GetRecipes())
	mcp.AddTool(mcpServer, searchForRecipesTool, h.SearchForRecipes())

	// Issue Reports (read-only)
	mcp.AddTool(mcpServer, getIssueReportTool, h.GetIssueReport())
	mcp.AddTool(mcpServer, getIssueReportsTool, h.GetIssueReports())
	mcp.AddTool(mcpServer, getIssueReportsByStatusTool, h.GetIssueReportsByStatus())

	// Webhooks (read-only)
	mcp.AddTool(mcpServer, getWebhookTool, h.GetWebhook())
	mcp.AddTool(mcpServer, getWebhooksTool, h.GetWebhooks())
	mcp.AddTool(mcpServer, getWebhookEventTypesTool, h.GetWebhookEventTypes())

	// Waitlists (read-only)
	mcp.AddTool(mcpServer, getWaitlistTool, h.GetWaitlist())
	mcp.AddTool(mcpServer, getWaitlistsTool, h.GetWaitlists())
	mcp.AddTool(mcpServer, getActiveWaitlistsTool, h.GetActiveWaitlists())
	mcp.AddTool(mcpServer, getWaitlistSignupTool, h.GetWaitlistSignup())
	mcp.AddTool(mcpServer, getWaitlistSignupsForWaitlistTool, h.GetWaitlistSignupsForWaitlist())

	return mcpServer
}
