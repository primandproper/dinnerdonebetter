package grpcapi

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	auditmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit/manager"
	authmgr "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth/managers"
	commentsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/manager"
	dataprivacymanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy/manager"
	identitymgr "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager"
	issuereportsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports/manager"
	mealplanningregistration "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/registration"
	notificationsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/manager"
	oauthmgr "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/manager"
	paymentsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments/manager"
	settingsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings/manager"
	uploadedmediamanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/manager"
	waitlistsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/manager"
	webhooksmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/manager"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories"
	auditrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	authrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auth"
	commentsrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/comments"
	dataprivacyrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/dataprivacy"
	identityrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	internalopsrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	issuereportsrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	oauthrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/oauth"
	paymentsrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/payments"
	uploadedmediarepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	webhooksrepo "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/webhooks"
	analyticssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/analytics/grpc"
	auditsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/audit/grpc"
	authsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/auth/grpc"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"
	authhttpsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/auth/handlers/authentication"
	commentssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/comments/grpc"
	dataprivacysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/dataprivacy/grpc"
	identitysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/grpc"
	internalopssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/internalops/grpc"
	issuereportssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/issuereports/grpc"
	notificationssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/notifications/grpc"
	oauthsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/oauth/grpc"
	paymentsadapters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/payments/adapters"
	paymentssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/payments/grpc"
	settingssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/settings/grpc"
	uploadedmediacfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/uploadedmedia/config"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc"
	waitlistssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/waitlists/grpc"
	webhookssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/webhooks/grpc"

	"github.com/primandproper/platform-go/v3/analytics/multisource"
	tokenscfg "github.com/primandproper/platform-go/v3/authentication/tokens/config"
	databasecfg "github.com/primandproper/platform-go/v3/database/config"
	featureflagscfg "github.com/primandproper/platform-go/v3/featureflags/config"
	"github.com/primandproper/platform-go/v3/httpclient"
	msgconfig "github.com/primandproper/platform-go/v3/messagequeue/config"
	"github.com/primandproper/platform-go/v3/observability"
	loggingcfg "github.com/primandproper/platform-go/v3/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v3/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v3/observability/tracing/config"
	"github.com/primandproper/platform-go/v3/qrcodes"
	"github.com/primandproper/platform-go/v3/random"
	"github.com/primandproper/platform-go/v3/server/grpc"
	uploadscfg "github.com/primandproper/platform-go/v3/uploads/config"
	"github.com/primandproper/platform-go/v3/uploads/objectstorage"

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
	metricscfg.RegisterMetricsProvider(i)
	loggingcfg.RegisterLogger(i)
	tracingcfg.RegisterTracerProvider(i)
	httpclient.RegisterHTTPClient(i)
	msgconfig.RegisterMessageQueue(i)
	random.RegisterGenerator(i)
	repositories.RegisterMigrator(i)
	databasecfg.RegisterDatabase(i)
	grpc.RegisterGRPCServer(i)
	do.ProvideValue(i, qrcodes.Issuer("Dinner Done Better"))
	qrcodes.RegisterBuilder(i)
	uploadscfg.RegisterStorageConfig(i)
	objectstorage.RegisterUploadManager(i)
	featureflagscfg.RegisterFeatureFlagManager(i)
	multisource.RegisterMultiSourceEventReporter(i)

	// authentication
	authentication.RegisterAuth(i)
	sessions.RegisterSessionProviders(i)
	tokenscfg.RegisterTokenIssuer(i)
	interceptors.RegisterAuthInterceptor(i)

	// repositories (core)
	auditrepo.RegisterAuditLogRepository(i)
	authrepo.RegisterAuthRepository(i)
	commentsrepo.RegisterCommentsRepository(i)
	identityrepo.RegisterIdentityRepository(i)
	issuereportsrepo.RegisterIssueReportsRepository(i)
	uploadedmediarepo.RegisterUploadedMediaRepository(i)
	webhooksrepo.RegisterWebhooksRepository(i)
	oauthrepo.RegisterOAuthRepository(i)
	paymentsrepo.RegisterPaymentsRepository(i)
	dataprivacyrepo.RegisterDataPrivacyRepository(i)
	internalopsrepo.RegisterInternalOpsRepository(i)

	// managers
	auditmanager.RegisterAuditDataManager(i)
	authmgr.RegisterAuthManager(i)
	commentsmanager.RegisterCommentsDataManager(i)
	identitymgr.RegisterIdentityDataManager(i)
	notificationsmanager.RegisterNotificationsDataManager(i)
	settingsmanager.RegisterSettingsDataManager(i)
	paymentsmanager.RegisterPaymentsDataManager(i)
	oauthmgr.RegisterOAuth2Manager(i)
	webhooksmanager.RegisterWebhookDataManager(i)
	waitlistsmanager.RegisterWaitlistDataManager(i)
	issuereportsmanager.RegisterIssueReportsDataManager(i)
	uploadedmediamanager.RegisterUploadedMediaManager(i)
	dataprivacymanager.RegisterDataPrivacyManager(i)
	paymentsadapters.RegisterPaymentProcessorRegistry(i)

	// services
	authsvc.RegisterAuthService(i)
	authhttpsvc.RegisterAuthHTTPService(i)
	analyticssvc.RegisterAnalyticsService(i)
	auditsvc.RegisterAuditService(i)
	commentssvc.RegisterCommentsService(i)
	dataprivacysvc.RegisterDataPrivacyService(i)
	identitysvc.RegisterIdentityService(i)
	internalopssvc.RegisterInternalOpsService(i)
	issuereportssvc.RegisterIssueReportsService(i)
	notificationssvc.RegisterNotificationsService(i)
	settingssvc.RegisterSettingsService(i)
	uploadedmediasvc.RegisterUploadedMediaService(i)
	webhookssvc.RegisterWebhooksService(i)
	oauthsvc.RegisterOAuthService(i)
	paymentssvc.RegisterPaymentsService(i)
	waitlistssvc.RegisterWaitlistsService(i)
	uploadedmediacfg.RegisterUploadedMediaConfig(i)

	// Domain: mealplanning
	mealplanningregistration.RegisterForGRPCAPI(i)

	// extras (functions from extras.go)
	RegisterExtras(i)

	return i
}

// Build builds a server.
func Build(
	ctx context.Context,
	cfg *config.APIServiceConfig,
) (*GRPCService, error) {
	i := BuildInjector(ctx, cfg)
	return do.MustInvoke[*GRPCService](i), nil
}
