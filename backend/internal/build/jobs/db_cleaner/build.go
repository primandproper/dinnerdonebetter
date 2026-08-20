package dbcleaner

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	dbcleaner "github.com/primandproper/dinnerdonebetter/backend/internal/services/oauth/workers/db_cleaner"

	oauth2servercfg "github.com/primandproper/platform-go/v12/authentication/oauth2server/config"
	databasecfg "github.com/primandproper/platform-go/v12/database/config"
	"github.com/primandproper/platform-go/v12/database/postgres"
	"github.com/primandproper/platform-go/v12/observability"
	loggingcfg "github.com/primandproper/platform-go/v12/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v12/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v12/observability/tracing/config"

	"github.com/samber/do/v2"
)

// BuildInjector creates and configures the dependency injection container.
func BuildInjector(
	ctx context.Context,
	cfg *config.DBCleanerConfig,
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
	internalops.RegisterInternalOpsRepository(i)
	oauth2servercfg.RegisterStore(i)
	dbcleaner.RegisterDBCleaner(i)

	return i
}

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.DBCleanerConfig,
) (*dbcleaner.Job, error) {
	i := BuildInjector(ctx, cfg)
	return do.MustInvoke[*dbcleaner.Job](i), nil
}
