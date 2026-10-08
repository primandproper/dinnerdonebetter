package mcpbuild

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identitybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/queuedmail"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	mealplanningregistration "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/registration"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"
	auditrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identitystore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identitystore"
	issuereportsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	uploadedmediarepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	waitlistsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"
	webhooksstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooksstore"

	databasecfg "github.com/primandproper/primitives-go/v2/database/config"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	"github.com/primandproper/primitives-go/v2/observability"
	loggingcfg "github.com/primandproper/primitives-go/v2/observability/logging/config"
	metricscfg "github.com/primandproper/primitives-go/v2/observability/metrics/config"
	tracingcfg "github.com/primandproper/primitives-go/v2/observability/tracing/config"

	"github.com/samber/do/v2"
)

// BuildInjector creates and configures the dependency injection container for the MCP server.
func BuildInjector(ctx context.Context, cfg *config.MCPServiceConfig) *do.RootScope {
	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue(i, cfg)

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
	// What a role grants, read from the policy tables the migrator seeds. The
	// identity repository resolves a principal's role names through it when it
	// builds a session.
	authorization.RegisterPolicyResolver(i)
	identitystore.RegisterIdentityStore(i)
	// The mailer the identity service hands an invitation to, and sign-in, password
	// reset and waitlists their mail. See internal/build/queuedmail.
	queuedmail.Register(i)
	identitybuild.RegisterSessionBuilder(i)

	// The upload registry, because both repositories above read media through it —
	// a user's avatar, a recipe step's images.
	uploadedmediarepo.RegisterUploadedMediaRepository(i)
	// Domain: mealplanning
	//
	// The repository the tools read through. The recording spine every store here records
	// through — the outbox writer, the webhook emitter, and the recorder over them — is
	// assembled in there too, as it is for the scheduler and the async handler, for the
	// reason config.SchedulerConfig.OutboxRelay gives.
	mealplanningregistration.RegisterForMCP(i)
	webhooksstore.RegisterWebhooksStore(i)
	waitlistsrepo.RegisterWaitlistsRepository(i)
	issuereportsrepo.RegisterIssueReportsRepository(i)

	// The domains' tools, as the one list the server mounts.
	do.Provide[[]mcptools.Toolset](i, Toolsets)

	return i
}

// Toolsets is every domain's tool surface, resolved from i in the order the server mounts them.
// The server's own tools over platform's stores are not in it; those are the server's.
func Toolsets(i do.Injector) ([]mcptools.Toolset, error) {
	var toolsets []mcptools.Toolset

	for _, contribute := range []func(do.Injector) (mcptools.Toolset, error){
		// Domain: mealplanning
		mealplanningregistration.MCPTools,
	} {
		toolset, err := contribute(i)
		if err != nil {
			return nil, err
		}

		toolsets = append(toolsets, toolset)
	}

	return toolsets, nil
}
