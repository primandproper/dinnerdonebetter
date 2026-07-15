package mobilenotificationscheduler

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"

	databasecfg "github.com/primandproper/platform-go/v4/database/config"
	"github.com/primandproper/platform-go/v4/database/postgres"
	msgconfig "github.com/primandproper/platform-go/v4/messagequeue/config"
	"github.com/primandproper/platform-go/v4/observability"
	loggingcfg "github.com/primandproper/platform-go/v4/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v4/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v4/observability/tracing/config"

	"github.com/samber/do/v2"
)

// BuildInjector creates and configures the dependency injection container.
func BuildInjector(
	ctx context.Context,
	cfg *config.MobileNotificationSchedulerConfig,
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
	identity.RegisterIdentityRepository(i)
	mealplanning.RegisterMealPlanningRepository(i)

	return i
}

// Build builds a mobile notification scheduler.
func Build(
	ctx context.Context,
	cfg *config.MobileNotificationSchedulerConfig,
) (*Scheduler, error) {
	i := BuildInjector(ctx, cfg)
	return do.MustInvoke[*Scheduler](i), nil
}
