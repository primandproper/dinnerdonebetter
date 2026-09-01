package mcpserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	mcpbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/services/mcp"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	waitlistsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"

	"github.com/primandproper/platform-go/v13/authentication/oauth2server"
	oauth2servercfg "github.com/primandproper/platform-go/v13/authentication/oauth2server/config"
	"github.com/primandproper/platform-go/v13/authentication/totp"
	"github.com/primandproper/platform-go/v13/database"
	issuereports "github.com/primandproper/platform-go/v13/issuereports"
	"github.com/primandproper/platform-go/v13/observability"
	routingcfg "github.com/primandproper/platform-go/v13/routing/config"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/samber/do/v2"
)

// Service is a built MCP server: the tool server, the authorization server that guards
// it, and the dependency container holding the database handles both read through.
//
// It is the seam between "the process starts" and "the process serves". Run used to build
// all of this inline and then block on ListenAndServe, so whether a deployment's config,
// DI container and schema agree with one another was answerable only by starting the
// binary and watching. A Service can be built, served on a listener the caller owns, and
// shut down — and two of them can be pointed at one database, which is the property the
// durable authorization server was adopted for and the one no in-process test of the
// router can demonstrate.
type Service struct {
	_ struct{} `json:"-"`

	injector         *do.RootScope
	pillars          *observability.Pillars
	mcpServer        *mcp.Server
	authServer       *oauth2server.Server
	resourceMetadata *oauth2server.ResourceMetadata
	routingConfig    *routingcfg.Config
	baseURL          string
}

// NewService builds the MCP server's dependencies from an already-loaded configuration.
//
// baseURL is the server's public address — the fleet's, not this process's — which the
// two discovery documents advertise and which an access token's audience names.
func NewService(ctx context.Context, cfg *config.MCPServiceConfig, baseURL string) (svc *Service, err error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil config")
	}

	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	pillars, err := cfg.Observability.NewPillars(ctx)
	if err != nil {
		return nil, fmt.Errorf("creating observability pillars: %w", err)
	}

	if err = cfg.ValidateWithContext(ctx); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

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

	injector := mcpbuild.BuildInjector(ctx, cfg)

	// The container owns a database pool from the first resolution below onward. A build
	// that fails after that and simply returns would leak it, which a process that builds
	// one of these and exits would never notice and a test that builds several would.
	defer func() {
		if err != nil {
			if report := injector.ShutdownWithContext(ctx); report != nil && !report.Succeed {
				pillars.Logger.Error("releasing MCP service container after a failed build", report)
			}
		}
	}()

	// Resolved rather than must-resolved: a container missing one of these is the failure
	// this whole seam exists to make visible, and it is worth more as a named error than
	// as a panic in whatever goroutine happened to be building the server.
	mealplanningRepo, err := do.Invoke[mealplanning.Repository](injector)
	if err != nil {
		return nil, fmt.Errorf("resolving meal planning repository: %w", err)
	}

	webhooksRepo, err := do.Invoke[webhooks.Repository](injector)
	if err != nil {
		return nil, fmt.Errorf("resolving webhooks repository: %w", err)
	}

	waitlistRepo, err := do.Invoke[*waitlistsrepo.Repository](injector)
	if err != nil {
		return nil, fmt.Errorf("resolving waitlists repository: %w", err)
	}

	issueReportsStore, err := do.Invoke[issuereports.Store](injector)
	if err != nil {
		return nil, fmt.Errorf("resolving issue reports store: %w", err)
	}

	identityRepo, err := do.Invoke[identity.Repository](injector)
	if err != nil {
		return nil, fmt.Errorf("resolving identity repository: %w", err)
	}

	authenticator, err := do.Invoke[authentication.Authenticator](injector)
	if err != nil {
		return nil, fmt.Errorf("resolving authenticator: %w", err)
	}

	totpVerifier, err := do.Invoke[totp.Verifier](injector)
	if err != nil {
		return nil, fmt.Errorf("resolving TOTP verifier: %w", err)
	}

	dbClient, err := do.Invoke[database.Client](injector)
	if err != nil {
		return nil, fmt.Errorf("resolving database client: %w", err)
	}

	// The authorization server this process runs. Its records live in the four
	// ddb_oauth2_* tables rather than in this process's memory, which is what makes
	// a second replica and a restart survivable: an authorization code issued by one
	// replica is redeemed at whichever one serves /token, and a registered MCP client
	// outlives a deploy.
	authServer, err := oauth2servercfg.NewServer(ctx, &cfg.OAuth2, dbClient,
		&subjectAuthenticator{
			identityRepo:  identityRepo,
			authenticator: authenticator,
			totpVerifier:  totpVerifier,
		},
		oauth2servercfg.WithPillars(pillars),
		oauth2servercfg.WithServerOptions(oauth2server.WithLoginRenderer(newLoginRenderer(pillars.Logger))),
	)
	if err != nil {
		return nil, fmt.Errorf("building authorization server: %w", err)
	}

	// The RFC 9728 document, published by the resource server rather than by the
	// authorization server — they are the same process here, but the document is
	// what tells a client with no token where to go, so it is mounted separately.
	resourceMetadata, err := oauth2server.NewResourceMetadata(baseURL, []string{baseURL},
		oauth2server.WithResourceName(fmt.Sprintf("%s MCP Server", branding.CompanyName)),
	)
	if err != nil {
		return nil, fmt.Errorf("building protected resource metadata: %w", err)
	}

	helper := &mcpToolManager{
		mealplanningRepo: mealplanningRepo,
		webhooksRepo:     webhooksRepo,
		waitlistsRepo:    waitlistRepo,
		issueReports:     issueReportsStore,
	}

	return &Service{
		injector:         injector,
		pillars:          pillars,
		mcpServer:        helper.setupServer(),
		authServer:       authServer,
		resourceMetadata: resourceMetadata,
		routingConfig:    &cfg.Routing,
		baseURL:          baseURL,
	}, nil
}

// Handler builds the service's HTTP surface for the named transport: the authorization
// server's endpoints and both discovery documents unauthenticated, and the MCP endpoint
// behind bearer authentication.
//
// It does not bind anything. The listener is the caller's, which is what lets a test
// serve two of these at once.
func (s *Service) Handler(ctx context.Context, transport string) (http.Handler, error) {
	var mcpHandler http.Handler

	switch transport {
	case TransportSSE:
		mcpHandler = mcp.NewSSEHandler(func(*http.Request) *mcp.Server { return s.mcpServer }, &mcp.SSEOptions{})
	case TransportHTTP:
		mcpHandler = mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return s.mcpServer }, &mcp.StreamableHTTPOptions{
			Stateless:      true,
			JSONResponse:   true,
			Logger:         slog.New(&slog.JSONHandler{}),
			EventStore:     mcp.NewMemoryEventStore(nil),
			SessionTimeout: 0,
		})
	default:
		return nil, fmt.Errorf("transport %q is not served over HTTP", transport)
	}

	router, err := buildRouter(ctx, mcpHandler, s.authServer, s.resourceMetadata, s.pillars, s.routingConfig, s.baseURL)
	if err != nil {
		return nil, fmt.Errorf("building router: %w", err)
	}

	return router.Handler(), nil
}

// ServeStdio serves MCP over stdin and stdout, blocking until the transport closes.
//
// Nothing here is authenticated, and the authorization server is not mounted: the client
// is whichever process holds the other end of the pipe, and it got there by being started
// by the user rather than by presenting a token.
func (s *Service) ServeStdio(ctx context.Context) error {
	return s.mcpServer.Run(ctx, &mcp.StdioTransport{})
}

// Shutdown releases what the service holds, principally the container's database pool.
//
// It stops no HTTP server: the listener belongs to whoever bound it, and shutting that
// down first is how a caller drains in-flight requests before the pool goes away.
func (s *Service) Shutdown(ctx context.Context) error {
	if report := s.injector.ShutdownWithContext(ctx); report != nil && !report.Succeed {
		return fmt.Errorf("shutting down MCP service container: %w", *report)
	}

	return nil
}
