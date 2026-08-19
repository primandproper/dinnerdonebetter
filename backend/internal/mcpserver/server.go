package mcpserver

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"strings"
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

	"github.com/primandproper/platform-go/v11/authentication/oauth2server"
	oauth2servercfg "github.com/primandproper/platform-go/v11/authentication/oauth2server/config"
	"github.com/primandproper/platform-go/v11/authentication/totp"
	"github.com/primandproper/platform-go/v11/database"
	"github.com/primandproper/platform-go/v11/encoding"
	"github.com/primandproper/platform-go/v11/observability"
	"github.com/primandproper/platform-go/v11/routing"
	routingcfg "github.com/primandproper/platform-go/v11/routing/config"
	"github.com/primandproper/platform-go/v11/version"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/samber/do/v2"
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

	pillars, err := cfg.Observability.NewPillars(ctx)
	if err != nil {
		return fmt.Errorf("creating observability pillars: %w", err)
	}
	logger := pillars.Logger

	if err = cfg.ValidateWithContext(ctx); err != nil {
		return fmt.Errorf("validating config: %w", err)
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
	dbClient := do.MustInvoke[database.Client](injector)

	// The authorization server this process runs. Its records live in the four
	// ddb_oauth2_* tables rather than in this process's memory, which is what makes
	// a second replica and a restart survivable: an authorization code issued by one
	// replica is redeemed at whichever one serves /token, and a registered MCP client
	// outlives a deploy.
	//
	// Issuer and Resources default to this server's own public URL rather than being
	// rendered into the config file, because only the deployment knows it — it arrives
	// as MCP_BASE_URL. They are the same string by default on purpose: the issuer names
	// the authorization server, the resource names the protected resource, and here
	// they are one process.
	//
	// Both defer to a configured value rather than overwriting it, so that
	// DINNER_DONE_BETTER_OAUTH2_ISSUER and _RESOURCES mean what the generated env var
	// constants say they mean. A deployment fronting this server at a different issuer
	// is the case that needs them.
	if cfg.OAuth2.Issuer == "" {
		cfg.OAuth2.Issuer = baseURL
	}
	if len(cfg.OAuth2.Resources) == 0 {
		cfg.OAuth2.Resources = []string{baseURL}
	}

	authServer, err := oauth2servercfg.NewServer(ctx, &cfg.OAuth2, dbClient,
		&subjectAuthenticator{
			identityRepo:  identityRepo,
			authenticator: authenticator,
			totpVerifier:  totpVerifier,
		},
		oauth2servercfg.WithPillars(pillars),
		oauth2servercfg.WithServerOptions(oauth2server.WithLoginRenderer(newLoginRenderer(logger))),
	)
	if err != nil {
		return fmt.Errorf("building authorization server: %w", err)
	}

	// The RFC 9728 document, published by the resource server rather than by the
	// authorization server — they are the same process here, but the document is
	// what tells a client with no token where to go, so it is mounted separately.
	resourceMetadata, err := oauth2server.NewResourceMetadata(baseURL, []string{baseURL},
		oauth2server.WithResourceName(fmt.Sprintf("%s MCP Server", branding.CompanyName)),
	)
	if err != nil {
		return fmt.Errorf("building protected resource metadata: %w", err)
	}

	helper := &mcpToolManager{
		mealplanningRepo: mealplanningRepo,
		webhooksRepo:     webhooksRepo,
		waitlistsRepo:    waitlistRepo,
		issueReportsRepo: issueReportsRepo,
	}
	server := helper.setupServer()

	log.Printf("serving now with transport: %s", transport)

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

	switch transport {
	case TransportStdio:
		if err = server.Run(ctx, &mcp.StdioTransport{}); err != nil {
			logger.Error("serving MCP server via stdio", err)
			return err
		}
	case TransportSSE:
		sseHandler := mcp.NewSSEHandler(func(request *http.Request) *mcp.Server {
			return server
		}, &mcp.SSEOptions{})

		router, routerErr := buildRouter(ctx, sseHandler, authServer, resourceMetadata, pillars, &cfg.Routing, baseURL)
		if routerErr != nil {
			return fmt.Errorf("building router: %w", routerErr)
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
			return err
		}
	case TransportHTTP:
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

		router, routerErr := buildRouter(ctx, streamHandler, authServer, resourceMetadata, pillars, &cfg.Routing, baseURL)
		if routerErr != nil {
			return fmt.Errorf("building router: %w", routerErr)
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
			return err
		}
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
	issueReportsRepo issuereports.Repository
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
