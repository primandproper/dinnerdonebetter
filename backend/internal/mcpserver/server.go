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
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	issuereports "github.com/primandproper/platform-go/v15/issuereports"
	waitlists "github.com/primandproper/platform-go/v15/waitlists"
	platformwebhooks "github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/encoding"
	"github.com/primandproper/primitives-go/v2/healthcheck"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/routing"
	routingcfg "github.com/primandproper/primitives-go/v2/routing/config"
	"github.com/primandproper/primitives-go/v2/version"

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
//
// health is what /_ops_/ready asks. Every check it holds must pass for the probe to answer 200.
func buildRouter(ctx context.Context, mcpHandler http.Handler, authServer *oauth2server.Server, resourceMetadata *oauth2server.ResourceMetadata, loginThrottle routing.Middleware, health healthcheck.Registry, pillars *observability.Pillars, routingCfg *routingcfg.Config, baseURL string) (*routing.Router, error) {
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
		// Readiness: every registered component reported up. A replica whose database is
		// unreachable can serve neither a token nor a tool, so it is taken out of rotation
		// rather than left answering 500s.
		opsRouter.Handle(http.MethodGet, "/ready", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			result := health.CheckAll(req.Context())

			status := http.StatusOK
			if result.Status != healthcheck.StatusUp {
				status = http.StatusServiceUnavailable
			}

			encoder.EncodeResponseWithStatus(req.Context(), res, result, status)
		}))
		opsRouter.Handle(http.MethodGet, "/version", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.Header().Set("Content-Type", "application/json")
			encoder.EncodeResponseWithStatus(req.Context(), res, version.Get(), http.StatusOK)
		}))
	})

	// The six authorization server endpoints, plus the protected resource document.
	// No auth middleware: these are how a caller gets a token in the first place. The login
	// form's POST is throttled; a nil throttle is a router built in a test about the rest.
	var authMountMiddleware []routing.Middleware
	if loginThrottle != nil {
		authMountMiddleware = append(authMountMiddleware, throttleLoginForm(loginThrottle))
	}
	authServer.Mount(router, authMountMiddleware...)
	resourceMetadata.Mount(router)

	// Wrap the MCP handler with bearer token auth middleware. The MCP transport
	// serves multiple methods (GET for streaming, POST for messages, DELETE to
	// terminate a session), so register the handler for each.
	verifier, err := oauth2server.NewVerifier(resourceMetadata, authServer,
		oauth2server.WithVerifierLogger(pillars.Logger),
		oauth2server.WithVerifierTracerProvider(pillars.TracerProvider),
		oauth2server.WithVerifierMetricsProvider(pillars.MetricsProvider),
	)
	if err != nil {
		return nil, err
	}

	authMiddleware := auth.RequireBearerToken(newTokenVerifier(verifier), &auth.RequireBearerTokenOptions{
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

// throttleLoginForm applies throttle to the login form's POST alone.
//
// Mount hands one middleware list to every endpoint it routes, and the POST to /authorize is the
// only one of them that tests a password: the GET renders the form, and /token and /revoke
// authenticate a client rather than a person. The API server throttles the same one route.
func throttleLoginForm(throttle routing.Middleware) routing.Middleware {
	return func(next http.Handler) http.Handler {
		throttled := throttle(next)

		return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodPost && req.URL.Path == oauth2server.PathAuthorize {
				throttled.ServeHTTP(res, req)
				return
			}

			next.ServeHTTP(res, req)
		})
	}
}

// mcpToolManager holds the tools over platform's stores: the ones every deployment of this
// server has, whatever domain it serves. A domain's own tools arrive as toolsets, and the server
// names none of them.
type mcpToolManager struct {
	// reader is the read executor every platform store call now takes. The MCP
	// tools are all reads, so one executor settled here is enough: none of them
	// has a caller transaction to join.
	reader       database.SQLQueryExecutor
	webhooks     platformwebhooks.Store
	waitlists    waitlists.Store
	issueReports issuereports.Store
	toolsets     []mcptools.Toolset
}

// userFromRequest resolves the authenticated user's account from the MCP request's auth token.
func (h *mcpToolManager) userFromRequest(req *mcp.CallToolRequest) (accountID string, err error) {
	return mcptools.AccountFromRequest(req)
}

func (h *mcpToolManager) setupServer() *mcp.Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: fmt.Sprintf("%s-mcp", branding.CompanyNameSlug), Version: "v1.0.0"}, nil)

	// Each domain's tools, in the order the build listed them. The SDK keeps the last tool
	// registered under a name and says nothing about the first, so the build's test over the
	// built server is what holds two domains to distinct names.
	for _, toolset := range h.toolsets {
		toolset.RegisterOn(mcpServer)
	}

	// Issue Reports (read-only, and only the caller's own: the queue is a service administrator's)
	mcp.AddTool(mcpServer, getIssueReportTool, h.GetIssueReport())
	mcp.AddTool(mcpServer, getIssueReportsTool, h.GetIssueReports())

	// Webhooks (read-only)
	mcp.AddTool(mcpServer, getWebhookTool, h.GetWebhook())
	mcp.AddTool(mcpServer, getWebhooksTool, h.GetWebhooks())
	mcp.AddTool(mcpServer, getWebhookEventTypesTool, h.GetWebhookEventTypes())

	// Waitlists (read-only, catalog only — see waitlists_waitlists.go for why the
	// signups are not here)
	mcp.AddTool(mcpServer, getWaitlistTool, h.GetWaitlist())
	mcp.AddTool(mcpServer, getWaitlistsTool, h.GetWaitlists())
	mcp.AddTool(mcpServer, getOpenWaitlistsTool, h.GetOpenWaitlists())

	return mcpServer
}
