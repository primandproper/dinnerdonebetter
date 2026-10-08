package mcpbuild

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identitybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/identity"
	issuereportsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/issuereports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/queuedmail"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/sagas"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	mealplanningregistration "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/registration"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"
	auditrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identitystore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identitystore"
	issuereportsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	uploadedmediarepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	waitlistsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"
	webhooksstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooksstore"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	platformissuereports "github.com/primandproper/platform-go/v15/issuereports"
	issuereportsmcp "github.com/primandproper/platform-go/v15/issuereports/mcp"
	platformwaitlists "github.com/primandproper/platform-go/v15/waitlists"
	waitlistsmcp "github.com/primandproper/platform-go/v15/waitlists/mcp"
	platformwebhooks "github.com/primandproper/platform-go/v15/webhooks"
	webhooksmcp "github.com/primandproper/platform-go/v15/webhooks/mcp"
	"github.com/primandproper/primitives-go/v2/clock"
	"github.com/primandproper/primitives-go/v2/database"
	databasecfg "github.com/primandproper/primitives-go/v2/database/config"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	loggingcfg "github.com/primandproper/primitives-go/v2/observability/logging/config"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	metricscfg "github.com/primandproper/primitives-go/v2/observability/metrics/config"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	tracingcfg "github.com/primandproper/primitives-go/v2/observability/tracing/config"

	"github.com/samber/do/v2"
)

// BuildInjector creates and configures the dependency injection container for the MCP server.
func BuildInjector(ctx context.Context, cfg *config.MCPServiceConfig) *do.RootScope {
	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue(i, cfg)
	do.ProvideValue[clock.Clock](i, clock.NewClock())

	// config field extraction
	RegisterConfigs(i)

	// platform providers
	observability.RegisterO11yConfigs(i)
	metricscfg.RegisterMetricsProvider(i)
	loggingcfg.RegisterLogger(i)
	tracingcfg.RegisterTracerProvider(i)
	databasecfg.RegisterClientConfig(i)
	postgres.RegisterDatabaseClient(i)

	// authentication (for login credential validation)
	authentication.RegisterAuth(i)

	// repositories
	auditrepo.RegisterAuditLog(i)
	// The platform recorder behind it, which the recording spine registered below files
	// every write's entry through.
	auditrepo.RegisterPlatformRecorder(i)

	// What a role grants, read from the policy tables the migrator seeds. The identity
	// repository resolves a principal's role names through it when it builds a session,
	// and the session builder renders a tool call's caller through it.
	authorization.RegisterPolicyResolver(i)
	identitystore.RegisterIdentityStore(i)
	// The mailer the identity service hands an invitation to, and sign-in, password
	// reset and waitlists their mail. See internal/build/queuedmail.
	queuedmail.Register(i)
	identitybuild.RegisterSessionBuilder(i)
	// The upload registry, because both repositories above read media through it —
	// a user's avatar, a recipe step's images.
	uploadedmediarepo.RegisterUploadedMediaRepository(i)

	// The gate every tool call passes: who is calling, rendered as this application's
	// session from the token the bearer middleware verified, and what they may do.
	do.Provide[*mcptools.Gate](i, func(i do.Injector) (*mcptools.Gate, error) {
		authenticate, err := mcptools.NewAuthenticator(
			do.MustInvoke[platformidentity.Store](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*identitybuild.SessionBuilder](i),
		)
		if err != nil {
			return nil, err
		}

		return mcptools.NewGate(authenticate, sessions.PrincipalFromContext, sessions.GrantsFromContext)
	})

	// The finalization saga's store and runner, which the meal planning manager starts
	// sagas through. No tool here starts one; the manager is the API's, and it is built
	// whole rather than with a write path that refuses. Registered before the domain, which
	// puts its definitions on the registry.
	sagas.RegisterSagaStore(i)
	sagas.RegisterSagas(i)

	// Domain: mealplanning
	//
	// The manager the tools read through, and the repository under it. The recording spine
	// every store here records through — the outbox writer, the webhook emitter, and the
	// recorder over them — is assembled in there too, as it is for the scheduler and the
	// async handler, for the reason config.SchedulerConfig.OutboxRelay gives.
	mealplanningregistration.RegisterForMCP(i)

	webhooksstore.RegisterWebhooksStore(i)
	waitlistsrepo.RegisterWaitlistsRepository(i)
	issuereportsrepo.RegisterIssueReportsRepository(i)

	// Every tool surface, as the one list the server mounts.
	do.Provide[[]mcptools.Toolset](i, Toolsets)

	return i
}

// Toolsets is every tool surface the server mounts, resolved from i in the order it mounts
// them: each domain's, then platform's own over the stores every deployment of this server
// has.
func Toolsets(i do.Injector) ([]mcptools.Toolset, error) {
	var toolsets []mcptools.Toolset

	for _, contribute := range []func(do.Injector) (mcptools.Toolset, error){
		// Domain: mealplanning
		mealplanningregistration.MCPTools,
		webhookTools,
		waitlistTools,
		issueReportTools,
	} {
		toolset, err := contribute(i)
		if err != nil {
			return nil, err
		}

		toolsets = append(toolsets, toolset)
	}

	return toolsets, nil
}

// webhookTools is platform's read-only surface over webhook endpoints and the event catalog.
//
// Account-scoped: an endpoint is its account's, as it is over gRPC. See
// sessions.AccountScopedPrincipal for why handing it the global principal would be quiet and
// wrong.
func webhookTools(i do.Injector) (mcptools.Toolset, error) {
	gate := do.MustInvoke[*mcptools.Gate](i)

	return webhooksmcp.NewTools(
		do.MustInvoke[platformwebhooks.Dispatcher](i),
		do.MustInvoke[platformwebhooks.Store](i),
		do.MustInvoke[database.Client](i),
		gate.Authenticate(),
		sessions.AccountScopedPrincipalFromContext,
		gate.Grants(),
		webhooksmcp.WithLogger(do.MustInvoke[logging.Logger](i)),
		webhooksmcp.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
		webhooksmcp.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
	)
}

// waitlistTools is platform's read-only surface over the waitlist catalog — the lists and
// never the signups, which is platform's ruling and the one this server made for itself
// before adopting it.
func waitlistTools(i do.Injector) (mcptools.Toolset, error) {
	gate := do.MustInvoke[*mcptools.Gate](i)

	return waitlistsmcp.NewTools(
		do.MustInvoke[platformwaitlists.Store](i),
		do.MustInvoke[database.Client](i),
		gate.Authenticate(),
		// Global: a waitlist is the deployment's. See internal/build/waitlists.
		sessions.PrincipalFromContext,
		gate.Grants(),
		waitlistsmcp.WithLogger(do.MustInvoke[logging.Logger](i)),
		waitlistsmcp.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
		waitlistsmcp.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
	)
}

// issueReportTools is platform's read-only surface over the report queue, under the same rule
// the gRPC surface applies: a report is its reporter's, and the administrator's who triages.
func issueReportTools(i do.Injector) (mcptools.Toolset, error) {
	gate := do.MustInvoke[*mcptools.Gate](i)

	return issuereportsmcp.NewTools(
		do.MustInvoke[platformissuereports.Store](i),
		do.MustInvoke[database.Client](i),
		gate.Authenticate(),
		// Global: every report is filed under the one scope. See internal/build/issuereports.
		sessions.PrincipalFromContext,
		gate.Grants(),
		issuereportsbuild.OwnReportOrAdmin(gate.Grants()),
		issuereportsmcp.WithLogger(do.MustInvoke[logging.Logger](i)),
		issuereportsmcp.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
		issuereportsmcp.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
	)
}
