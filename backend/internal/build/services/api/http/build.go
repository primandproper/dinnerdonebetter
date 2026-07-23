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

	analyticscfg "github.com/primandproper/platform-go/v5/analytics/config"
	"github.com/primandproper/platform-go/v5/database"
	databasecfg "github.com/primandproper/platform-go/v5/database/config"
	"github.com/primandproper/platform-go/v5/encoding"
	"github.com/primandproper/platform-go/v5/healthcheck"
	msgconfig "github.com/primandproper/platform-go/v5/messagequeue/config"
	"github.com/primandproper/platform-go/v5/observability"
	loggingcfg "github.com/primandproper/platform-go/v5/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v5/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v5/observability/tracing/config"
	"github.com/primandproper/platform-go/v5/qrcodes"
	"github.com/primandproper/platform-go/v5/random"
	"github.com/primandproper/platform-go/v5/server/http"

	"github.com/samber/do/v2"
)

// RegisterHTTPServerServices registers the providers the HTTP API server needs beyond
// what the gRPC API injector already provides. It is safe to call on the shared gRPC
// API injector: none of these registrations overlap with that container's contents.
func RegisterHTTPServerServices(i do.Injector) {
	encoding.RegisterServerEncoderDecoder(i)
	analyticscfg.RegisterEventReporter(i)
	do.Provide[healthcheck.Registry](i, func(i do.Injector) (healthcheck.Registry, error) {
		registry := healthcheck.NewRegistry()
		dbClient := do.MustInvoke[database.Client](i)
		if checker, ok := dbClient.(healthcheck.DatabaseReadyChecker); ok {
			registry.Register(healthcheck.NewDatabaseChecker("database", checker))
		}
		return registry, nil
	})
	http.RegisterHTTPServer(i, "api_server")

	// services
	paymentshttp.RegisterPaymentsHTTP(i)

	// routes
	RegisterAPIRouter(i)
}

// BuildInjector creates and configures a standalone dependency injection container for
// the HTTP API server. The combined HTTP+gRPC server does not use this; it registers
// RegisterHTTPServerServices onto the shared gRPC injector instead.
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
	msgconfig.RegisterMessageQueue(i)
	repositories.RegisterMigrator(i)
	databasecfg.RegisterDatabase(i)
	random.RegisterGenerator(i)
	do.ProvideValue(i, qrcodes.Issuer("Dinner Done Better"))
	qrcodes.RegisterBuilder(i)

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

	// searchers
	RegisterSearchers(i)

	// HTTP-server-specific providers (shared with the combined-server path)
	RegisterHTTPServerServices(i)

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
