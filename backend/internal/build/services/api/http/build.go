package api

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	authcfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	identitymgr "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager"
	paymentsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments/manager"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories"
	auditrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	oauthrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/oauth"
	paymentsrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/payments"
	authservice "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/auth/handlers/authentication"
	paymentsadapters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/payments/adapters"
	paymentshttp "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/payments/http"

	analyticscfg "github.com/primandproper/platform-go/v4/analytics/config"
	"github.com/primandproper/platform-go/v4/database"
	databasecfg "github.com/primandproper/platform-go/v4/database/config"
	"github.com/primandproper/platform-go/v4/encoding"
	"github.com/primandproper/platform-go/v4/healthcheck"
	msgconfig "github.com/primandproper/platform-go/v4/messagequeue/config"
	"github.com/primandproper/platform-go/v4/observability"
	loggingcfg "github.com/primandproper/platform-go/v4/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v4/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v4/observability/tracing/config"
	"github.com/primandproper/platform-go/v4/random"
	routingcfg "github.com/primandproper/platform-go/v4/routing/config"
	"github.com/primandproper/platform-go/v4/server/http"

	"github.com/samber/do/v2"
)

// BuildInjector creates and configures the dependency injection container.
func BuildInjector(
	ctx context.Context,
	cfg *config.APIServiceConfig,
) *do.RootScope {
	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue(i, cfg)

	// config field extraction
	RegisterConfigs(i)

	// platform providers
	observability.RegisterO11yConfigs(i)
	loggingcfg.RegisterLogger(i)
	tracingcfg.RegisterTracerProvider(i)
	metricscfg.RegisterMetricsProvider(i)
	encoding.RegisterServerEncoderDecoder(i)
	msgconfig.RegisterMessageQueue(i)
	analyticscfg.RegisterEventReporter(i)
	repositories.RegisterMigrator(i)
	databasecfg.RegisterDatabase(i)
	do.Provide[healthcheck.Registry](i, func(i do.Injector) (healthcheck.Registry, error) {
		registry := healthcheck.NewRegistry()
		dbClient := do.MustInvoke[database.Client](i)
		if checker, ok := dbClient.(healthcheck.DatabaseReadyChecker); ok {
			registry.Register(healthcheck.NewDatabaseChecker("database", checker))
		}
		return registry, nil
	})
	routingcfg.RegisterRouteParamManager(i)
	random.RegisterGenerator(i)
	http.RegisterHTTPServer(i, "api_server")

	// authentication
	authentication.RegisterAuth(i)
	authcfg.RegisterConfigs(i)

	// repos
	auditrepo.RegisterAuditLogRepository(i)
	identityrepo.RegisterIdentityRepository(i)
	oauthrepo.RegisterOAuthRepository(i)
	paymentsrepo.RegisterPaymentsRepository(i)

	// managers
	identitymgr.RegisterIdentityDataManager(i)
	paymentsmanager.RegisterPaymentsDataManager(i)
	paymentsadapters.RegisterPaymentProcessorRegistry(i)

	// services
	authservice.RegisterAuthHTTPService(i)
	paymentshttp.RegisterPaymentsHTTP(i)

	// searchers & routes
	RegisterSearchers(i)
	RegisterAPIRouter(i)

	return i
}

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.APIServiceConfig,
) (http.Server, error) {
	i := BuildInjector(ctx, cfg)
	return do.MustInvoke[http.Server](i), nil
}
