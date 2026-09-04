package mcpbuild

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	auditrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	identityrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	issuereportsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	uploadedmediarepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	waitlistsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"
	webhooksrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooks"

	databasecfg "github.com/primandproper/platform-go/v13/database/config"
	"github.com/primandproper/platform-go/v13/database/postgres"
	"github.com/primandproper/platform-go/v13/observability"
	loggingcfg "github.com/primandproper/platform-go/v13/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v13/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v13/observability/tracing/config"

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
	auditrepo.RegisterAuditLogRepository(i)
	// What a role grants, read from the policy tables the migrator seeds. The
	// identity repository resolves a principal's role names through it when it
	// builds a session.
	identityrepo.RegisterPolicyResolver(i)
	identityrepo.RegisterIdentityRepository(i)
	events.RegisterOutboxEmitter(i)

	// The upload registry, because both repositories above read media through it —
	// a user's avatar, a recipe step's images.
	uploadedmediarepo.RegisterUploadedMediaRepository(i)
	mealplanningrepo.RegisterMealPlanningRepository(i)
	webhooksrepo.RegisterWebhooksRepository(i)
	waitlistsrepo.RegisterWaitlistsRepository(i)
	issuereportsrepo.RegisterIssueReportsRepository(i)

	return i
}
