package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	mcpbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/services/mcp"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	waitlistsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"

	"github.com/primandproper/platform-go/v9/authentication/totp"
	"github.com/primandproper/platform-go/v9/encoding"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/routing"
	routingcfg "github.com/primandproper/platform-go/v9/routing/config"
	"github.com/primandproper/platform-go/v9/version"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/samber/do/v2"
	"github.com/spf13/pflag"
)

const (
	defaultMcpServerConfigurationFilepath = "deploy/environments/localdev/config_files/mcp_server_config.json"

	transportStdio = "stdio"
	transportSSE   = "sse"
	transportHTTP  = "http"

	defaultPort    = 8888
	defaultBaseURL = "http://localhost:8888"
)

func main() {
	// Parse command-line flags
	transport := pflag.String("transport", transportHTTP, fmt.Sprintf("Transport method: %s, %s, or %s", transportStdio, transportSSE, transportHTTP))
	baseURL := pflag.String("base-url", defaultBaseURL, "Public base URL of the MCP server (used for OAuth2 metadata)")
	pflag.Parse()

	// Validate transport flag
	validTransports := map[string]bool{transportStdio: true, transportSSE: true, transportHTTP: true}
	if !validTransports[*transport] {
		log.Fatalf("Invalid transport method: %s. Allowed values are: %s, %s, %s", *transport, transportStdio, transportSSE, transportHTTP)
	}

	// Allow override via env var
	if envBase := os.Getenv("MCP_BASE_URL"); envBase != "" {
		*baseURL = envBase
	}

	ctx := context.Background()

	configFilepath := os.Getenv(config.ConfigurationFilePathEnvVarKey)
	if configFilepath == "" {
		configFilepath = defaultMcpServerConfigurationFilepath
	}

	cfg, err := config.LoadConfigFromPath[config.MCPServiceConfig](configFilepath)
	if err != nil {
		log.Fatal(err)
	}

	pillars, err := cfg.Observability.NewPillars(ctx)
	if err != nil {
		log.Fatal(err)
	}
	logger := pillars.Logger

	if err = cfg.ValidateWithContext(ctx); err != nil {
		log.Fatal(err)
	}

	port := cfg.HTTPServer.Port
	if port == 0 {
		port = defaultPort
	}

	// Build DI container with repositories and auth.
	injector := mcpbuild.BuildInjector(ctx, cfg)

	mealplanningRepo := do.MustInvoke[mealplanning.Repository](injector)
	webhooksRepo := do.MustInvoke[webhooks.Repository](injector)
	waitlistRepo := do.MustInvoke[*waitlistsrepo.Repository](injector)
	issueReportsRepo := do.MustInvoke[issuereports.Repository](injector)
	identityRepo := do.MustInvoke[identity.Repository](injector)
	authenticator := do.MustInvoke[authentication.Authenticator](injector)
	totpVerifier := do.MustInvoke[totp.Verifier](injector)

	// Create token store for per-user auth.
	tokens := newTokenStore()
	tokens.startCleanupLoop(ctx)

	helper := &mcpToolManager{
		tokens:           tokens,
		mealplanningRepo: mealplanningRepo,
		webhooksRepo:     webhooksRepo,
		waitlistsRepo:    waitlistRepo,
		issueReportsRepo: issueReportsRepo,
	}
	server := helper.setupServer()

	log.Printf("serving now with transport: %s", *transport)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM,
	)

	go func() {
		<-signalChan
		os.Exit(0)
	}()

	switch *transport {
	case transportStdio:
		if err = server.Run(ctx, &mcp.StdioTransport{}); err != nil {
			logger.Error("serving MCP server via stdio", err)
			log.Fatal(err)
		}
	case transportSSE:
		sseHandler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
			return server
		}, &mcp.SSEOptions{})

		router, routerErr := buildRouter(ctx, sseHandler, tokens, pillars, &cfg.Routing, *baseURL, identityRepo, authenticator, totpVerifier)
		if routerErr != nil {
			log.Fatalf("failed to build router: %v", routerErr)
		}

		srv := &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           router.Handler(),
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
		}
		if err = srv.ListenAndServe(); err != nil {
			logger.Error("starting MCP server via SSE", err)
		}
	case transportHTTP:
		handlerOpts := &mcp.StreamableHTTPOptions{
			Stateless:      true,
			JSONResponse:   true,
			Logger:         slog.New(&slog.JSONHandler{}),
			EventStore:     mcp.NewMemoryEventStore(nil),
			SessionTimeout: 0,
		}
		streamHandler := mcp.NewStreamableHTTPHandler(func(request *http.Request) *mcp.Server {
			return server
		}, handlerOpts)

		router, routerErr := buildRouter(ctx, streamHandler, tokens, pillars, &cfg.Routing, *baseURL, identityRepo, authenticator, totpVerifier)
		if routerErr != nil {
			log.Fatalf("failed to build router: %v", routerErr)
		}

		srv := &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           router.Handler(),
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
		}
		if err = srv.ListenAndServe(); err != nil {
			logger.Error("starting MCP server via HTTP", err)
		}
	}
}

// buildRouter creates a router with OAuth2 routes (unauthenticated) and the MCP handler (authenticated).
func buildRouter(ctx context.Context, mcpHandler http.Handler, tokens *tokenStore, pillars *observability.Pillars, routingCfg *routingcfg.Config, baseURL string, identityRepo identity.Repository, authenticator authentication.Authenticator, totpVerifier totp.Verifier) (*routing.Router, error) {
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

	// Register OAuth2 routes (no auth middleware — these handle authentication themselves).
	registerOAuth2Routes(router, tokens, baseURL, identityRepo, authenticator, totpVerifier)

	// Wrap the MCP handler with bearer token auth middleware. The MCP transport
	// serves multiple methods (GET for streaming, POST for messages, DELETE to
	// terminate a session), so register the handler for each.
	authMiddleware := auth.RequireBearerToken(tokens.verifyToken, &auth.RequireBearerTokenOptions{
		ResourceMetadataURL: baseURL + "/.well-known/oauth-protected-resource",
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
	tokens           *tokenStore
	mealplanningRepo mealplanning.Repository
	webhooksRepo     webhooks.Repository
	waitlistsRepo    *waitlistsrepo.Repository
	issueReportsRepo issuereports.Repository
}

// userFromRequest resolves the authenticated user's account from the MCP request's auth token.
func (h *mcpToolManager) userFromRequest(req *mcp.CallToolRequest) (accountID string, err error) {
	if req.Extra == nil || req.Extra.TokenInfo == nil {
		return "", fmt.Errorf("not authenticated")
	}
	rawToken, ok := req.Extra.TokenInfo.Extra["raw_token"].(string)
	if !ok || rawToken == "" {
		return "", fmt.Errorf("bearer token not found in request")
	}
	_, accountID, err = h.tokens.userContextForToken(rawToken)
	return accountID, err
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
	mcp.AddTool(mcpServer, getIssueReportsForAccountTool, h.GetIssueReportsForAccount())

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
