package grpcapi

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/sagas"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	auditmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/manager"
	authmgr "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/managers"
	commentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/manager"
	identitymgr "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager"
	issuereportsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/manager"
	mealplanningregistration "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/registration"
	notificationsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/manager"
	oauthmgr "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/manager"
	paymentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/manager"
	settingsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/manager"
	uploadedmediamanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/manager"
	waitlistsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/manager"
	webhooksmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/manager"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories"
	auditrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	authrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auth"
	commentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/comments"
	identityrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	internalopsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	issuereportsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	oauthrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/oauth"
	paymentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/payments"
	uploadedmediarepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhookdispatch"
	webhooksrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooks"
	analyticssvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/analytics/grpc"
	auditsvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/audit/grpc"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"
	authhttpsvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/handlers/authentication"
	commentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/comments/grpc"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"
	dataprivacysvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/grpc"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/grpc"
	internalopssvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/grpc"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/issuereports/grpc"
	notificationssvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/notifications/grpc"
	oauthsvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/oauth/grpc"
	paymentsadapters "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/adapters"
	paymentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/grpc"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/settings/grpc"
	uploadedmediacfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/config"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc"
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/waitlists/grpc"
	webhookssvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/webhooks/grpc"

	"github.com/primandproper/platform-go/v9/analytics/multisource"
	tokenscfg "github.com/primandproper/platform-go/v9/authentication/tokens/config"
	databasecfg "github.com/primandproper/platform-go/v9/database/config"
	featureflagscfg "github.com/primandproper/platform-go/v9/featureflags/config"
	"github.com/primandproper/platform-go/v9/httpclient"
	msgconfig "github.com/primandproper/platform-go/v9/messagequeue/config"
	"github.com/primandproper/platform-go/v9/observability"
	loggingcfg "github.com/primandproper/platform-go/v9/observability/logging/config"
	metricscfg "github.com/primandproper/platform-go/v9/observability/metrics/config"
	tracingcfg "github.com/primandproper/platform-go/v9/observability/tracing/config"
	"github.com/primandproper/platform-go/v9/qrcodes"
	"github.com/primandproper/platform-go/v9/random"
	"github.com/primandproper/platform-go/v9/server/grpc"
	uploadscfg "github.com/primandproper/platform-go/v9/uploads/config"
	"github.com/primandproper/platform-go/v9/uploads/objectstorage"

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
	do.ProvideValue(i, qrcodes.Issuer(branding.CompanyName))
	qrcodes.RegisterBuilder(i)
	uploadscfg.RegisterStorageConfig(i)
	objectstorage.RegisterUploadManager(i)
	// Export artifacts get an upload manager of their own, pointed at the user data bucket
	// rather than the media bucket the ambient one above serves, plus the request store the
	// Service reads and writes.
	dataprivacycfg.RegisterArtifactStorage(i)
	dataprivacycfg.RegisterRequestService(i)
	featureflagscfg.RegisterFeatureFlagManager(i)
	multisource.RegisterMultiSourceEventReporter(i)

	// authentication
	authentication.RegisterAuth(i)
	tokenscfg.RegisterTokenIssuer(i)
	interceptors.RegisterAuthInterceptor(i)

	// repositories (core)
	auditrepo.RegisterAuditLogRepository(i)
	authrepo.RegisterAuthRepository(i)
	commentsrepo.RegisterCommentsRepository(i)
	identityrepo.RegisterIdentityRepository(i)
	issuereportsrepo.RegisterIssueReportsRepository(i)
	uploadedmediarepo.RegisterUploadedMediaRepository(i)
	webhookdispatch.RegisterWebhookDispatch(i)
	webhooksrepo.RegisterWebhooksRepository(i)
	oauthrepo.RegisterOAuthRepository(i)
	paymentsrepo.RegisterPaymentsRepository(i)
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
	paymentsadapters.RegisterPaymentProcessorRegistry(i)

	// services
	authsvc.RegisterAuthService(i)
	authhttpsvc.RegisterAuthHTTPService(i)
	analyticssvc.RegisterAnalyticsService(i)
	auditsvc.RegisterAuditService(i)
	commentssvc.RegisterCommentsService(i)
	dataprivacysvc.RegisterDataPrivacyService(i)
	do.Provide[dataprivacysvc.DataPrivacyMethodPermissions](i, func(i do.Injector) (dataprivacysvc.DataPrivacyMethodPermissions, error) {
		return dataprivacysvc.ProvideMethodPermissions(), nil
	})
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

	// The saga machinery, minus the worker: this process starts durable processes and does not
	// advance them. Registered before the domain, which puts its definitions on the registry.
	sagas.RegisterSagas(i)

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
