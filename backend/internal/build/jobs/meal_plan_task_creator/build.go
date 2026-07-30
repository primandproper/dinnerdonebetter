package mealplantaskcreator

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	mealplantaskcreator "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_creator"

	databasecfg "github.com/primandproper/platform-go/v8/database/config"
	"github.com/primandproper/platform-go/v8/database/postgres"
	msgconfig "github.com/primandproper/platform-go/v8/messagequeue/config"
	"github.com/primandproper/platform-go/v8/observability"
	loggingcfg "github.com/primandproper/platform-go/v8/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v8/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v8/observability/tracing/config"

	"github.com/samber/do/v2"
)

// BuildInjector creates and configures the dependency injection container.
func BuildInjector(
	ctx context.Context,
	cfg *config.MealPlanTaskCreatorConfig,
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
	recipeanalysis.RegisterRecipeAnalyzer(i)
	auditlogentries.RegisterAuditLogRepository(i)
	identity.RegisterIdentityRepository(i)
	mealplanning.RegisterMealPlanningRepository(i)
	mealplantaskcreator.RegisterMealPlanTaskCreator(i)

	return i
}

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.MealPlanTaskCreatorConfig,
) (*mealplantaskcreator.Worker, error) {
	i := BuildInjector(ctx, cfg)
	return do.MustInvoke[*mealplantaskcreator.Worker](i), nil
}
