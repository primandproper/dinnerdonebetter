package mcpbuild

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	auditrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	issuereportsrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	mealplanningrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	waitlistsrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"
	webhooksrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/webhooks"

	databasecfg "github.com/primandproper/platform-go/v8/database/config"
	"github.com/primandproper/platform-go/v8/database/postgres"
	"github.com/primandproper/platform-go/v8/observability"
	loggingcfg "github.com/primandproper/platform-go/v8/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v8/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v8/observability/tracing/config"

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
	identityrepo.RegisterIdentityRepository(i)
	mealplanningrepo.RegisterMealPlanningRepository(i)
	webhooksrepo.RegisterWebhooksRepository(i)
	waitlistsrepo.RegisterWaitlistsRepository(i)
	issuereportsrepo.RegisterIssueReportsRepository(i)

	return i
}
