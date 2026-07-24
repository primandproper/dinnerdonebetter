package searchdataindexscheduler

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	mealplanningrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"

	databasecfg "github.com/primandproper/platform-go/v6/database/config"
	"github.com/primandproper/platform-go/v6/database/postgres"
	msgconfig "github.com/primandproper/platform-go/v6/messagequeue/config"
	"github.com/primandproper/platform-go/v6/observability"
	loggingcfg "github.com/primandproper/platform-go/v6/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v6/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v6/observability/tracing/config"
	"github.com/primandproper/platform-go/v6/search/text/indexing"

	"github.com/samber/do/v2"
)

// BuildInjector creates and configures the dependency injection container.
func BuildInjector(
	ctx context.Context,
	cfg *config.SearchDataIndexSchedulerConfig,
) *do.RootScope {
	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue(i, cfg)

	RegisterConfigs(i)

	observability.RegisterO11yConfigs(i)
	tracingcfg.RegisterTracerProvider(i)
	loggingcfg.RegisterLogger(i)
	metricscfg.RegisterMetricsProvider(i)
	msgconfig.RegisterMessageQueue(i)
	databasecfg.RegisterClientConfig(i)
	postgres.RegisterDatabaseClient(i)
	auditlogentries.RegisterAuditLogRepository(i)
	identityrepo.RegisterIdentityRepository(i)
	// Domain: mealplanning
	mealplanningrepo.RegisterMealPlanningRepository(i)

	do.Provide[map[string]indexing.Function](i, func(i do.Injector) (map[string]indexing.Function, error) {
		identityRepo := do.MustInvoke[identity.Repository](i)
		mealPlanningRepo := do.MustInvoke[mealplanning.Repository](i)
		return ProvideIndexFunctions(identityRepo, mealPlanningRepo), nil
	})

	indexing.RegisterIndexScheduler(i)

	return i
}

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.SearchDataIndexSchedulerConfig,
) (*indexing.IndexScheduler, error) {
	i := BuildInjector(ctx, cfg)
	return do.MustInvoke[*indexing.IndexScheduler](i), nil
}
