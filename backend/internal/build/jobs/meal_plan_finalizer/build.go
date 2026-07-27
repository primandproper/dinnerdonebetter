package mealplanfinalizer

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	mealplanfinalizer "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalizer"

	databasecfg "github.com/primandproper/platform-go/v7/database/config"
	"github.com/primandproper/platform-go/v7/database/postgres"
	msgconfig "github.com/primandproper/platform-go/v7/messagequeue/config"
	"github.com/primandproper/platform-go/v7/observability"
	loggingcfg "github.com/primandproper/platform-go/v7/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v7/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v7/observability/tracing/config"

	"github.com/samber/do/v2"
)

// BuildInjector creates and configures the dependency injection container.
func BuildInjector(
	ctx context.Context,
	cfg *config.MealPlanFinalizerConfig,
) *do.RootScope {
	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue(i, cfg)

	RegisterConfigs(i)

	observability.RegisterO11yConfigs(i)
	tracingcfg.RegisterTracerProvider(i)
	loggingcfg.RegisterLogger(i)
	metricscfg.RegisterMetricsProvider(i)
	databasecfg.RegisterClientConfig(i)
	postgres.RegisterDatabaseClient(i)
	msgconfig.RegisterMessageQueue(i)
	auditlogentries.RegisterAuditLogRepository(i)
	identity.RegisterIdentityRepository(i)
	mealplanning.RegisterMealPlanningRepository(i)
	mealplanfinalizer.RegisterMealPlanFinalizer(i)

	return i
}

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.MealPlanFinalizerConfig,
) (*mealplanfinalizer.Worker, error) {
	i := BuildInjector(ctx, cfg)
	return do.MustInvoke[*mealplanfinalizer.Worker](i), nil
}
