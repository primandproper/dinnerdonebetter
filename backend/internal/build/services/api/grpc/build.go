package grpcapi

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	auditbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/auditlog"
	commentstargets "github.com/primandproper/dinnerdonebetter/backend/internal/build/comments"
	dataprivacybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/dataprivacy"
	identitybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/identity"
	issuereportsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/issuereports"
	notificationsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/notifications"
	oauth2clientsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/oauth2clients"
	paymentsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/sagas"
	settingsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/settings"
	signinbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/signin"
	waitlistsbuild2 "github.com/primandproper/dinnerdonebetter/backend/internal/build/waitlists"
	webhooksbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	auditmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/manager"
	authmgr "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/managers"
	mealplanningregistration "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/registration"
	notificationsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/manager"
	paymentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/manager"
	webhooksmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/manager"
	appentitlements "github.com/primandproper/dinnerdonebetter/backend/internal/entitlements"
	appmetering "github.com/primandproper/dinnerdonebetter/backend/internal/metering"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories"
	auditrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	authrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auth"
	commentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/comments"
	identitystore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identitystore"
	internalopsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	issuereportsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	oauth2clientsstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/oauth2clientsstore"
	paymentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/payments"
	settingsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/settings"
	uploadedmediarepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	waitlistsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"
	webhooksstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooksstore"
	analyticssvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/analytics/grpc"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"
	authhttpsvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/handlers/authentication"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"
	dataprivacysvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/grpc"
	internalopssvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/grpc"
	paymentsadapters "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/adapters"
	uploadedmediacfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/config"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc"

	platformerrormappers "github.com/primandproper/platform-go/v14/errormappers"
	operationscfg "github.com/primandproper/platform-go/v14/operations/config"
	"github.com/primandproper/primitives-go/v2/analytics/multisource"
	tokenscfg "github.com/primandproper/primitives-go/v2/authentication/tokens/config"
	databasecfg "github.com/primandproper/primitives-go/v2/database/config"
	featureflagscfg "github.com/primandproper/primitives-go/v2/featureflags/config"
	"github.com/primandproper/primitives-go/v2/httpclient"
	msgconfig "github.com/primandproper/primitives-go/v2/messagequeue/config"
	"github.com/primandproper/primitives-go/v2/observability"
	loggingcfg "github.com/primandproper/primitives-go/v2/observability/logging/config"
	metricscfg "github.com/primandproper/primitives-go/v2/observability/metrics/config"
	tracingcfg "github.com/primandproper/primitives-go/v2/observability/tracing/config"
	"github.com/primandproper/primitives-go/v2/qrcodes"
	"github.com/primandproper/primitives-go/v2/random"
	"github.com/primandproper/primitives-go/v2/server/grpc"
	uploadscfg "github.com/primandproper/primitives-go/v2/uploads/config"
	"github.com/primandproper/primitives-go/v2/uploads/objectstorage"

	"github.com/samber/do/v2"
)

// BuildInjector creates and configures the dependency injection container.
func BuildInjector(
	ctx context.Context,
	cfg *config.APIServiceConfig,
) *do.RootScope {
	i := do.New()

	// The transport mappings for every platform-go sentinel, installed before anything can
	// raise one. As of v14 the mappers do not register themselves — a package that installs
	// itself into a process-wide registry by being linked in is a side effect a consumer
	// cannot opt out of — so the composition root makes the one call. Without it every
	// platform error reaches a client as whatever default code the handler happened to name,
	// which compiles and passes every test that does not assert on a code.
	//
	// This repository's own mappers still register from an init, because their packages are
	// imported for nothing else and there is no root they could be called from.
	platformerrormappers.Register()

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

	// The operations tier a privacy request is submitted as. Only the enqueue-and-read half
	// runs here — the worker that claims and runs operations is in the scheduler — but the
	// registry is not optional on this side: Start looks a kind up in the registry of the
	// process calling it, so an API server without the privacy kinds registered would refuse
	// every request with ErrUnknownKind.
	dataprivacybuild.RegisterRegistry(i)
	dataprivacybuild.RegisterOperationsRegistry(i)
	operationscfg.RegisterStore(i)
	operationscfg.RegisterQueue(i)
	operationscfg.RegisterService(i)

	dataprivacycfg.RegisterRequestService(i)
	featureflagscfg.RegisterFeatureFlagManager(i)
	multisource.RegisterMultiSourceEventReporter(i)

	// Usage metering. Only the ingest half runs here: the flusher that posts usage to a
	// billing provider is a scheduled pass in the scheduler process. The enforcer is
	// registered but nothing consults it yet — see its registration for what has to change
	// before the first limit goes on.
	appmetering.RegisterRegistry(i)
	appmetering.RegisterStore(i)
	appmetering.RegisterRecorder(i)
	appmetering.RegisterEnforcer(i)

	// What each account may use. Registration order does not matter — these providers are
	// lazy, and the two halves refer to each other: the enforcer above resolves the quota
	// source registered here, and the checker registered here resolves that enforcer. What
	// matters is that there is one quota source, so the limit a check reports and the limit
	// enforced against an account cannot drift apart. Nothing consults the checker yet.
	appentitlements.RegisterFeatures(i)
	appentitlements.RegisterPlanSource(i)
	appentitlements.RegisterCatalog(i)
	appentitlements.RegisterQuotaSource(i)
	appentitlements.RegisterChecker(i)

	// authentication
	authentication.RegisterAuth(i)
	tokenscfg.RegisterTokenIssuer(i)
	interceptors.RegisterAuthInterceptor(i)

	// The catalog of things this application accepts comments on, carrying an
	// existence check per type. It is registered before the store that enforces
	// it because a store built without one accepts no writes at all.
	commentstargets.RegisterTargets(i)

	// repositories (core)
	auditrepo.RegisterAuditLogRepository(i)
	authrepo.RegisterAuthRepository(i)
	commentsrepo.RegisterCommentsRepository(i)
	// What a role grants, read from the policy tables the migrator seeds. The
	// identity repository resolves a principal's role names through it when it
	// builds a session.
	authorization.RegisterPolicyResolver(i)
	identitystore.RegisterIdentityStore(i)
	identitybuild.RegisterSessionBuilder(i)
	issuereportsrepo.RegisterIssueReportsRepository(i)
	uploadedmediarepo.RegisterUploadedMediaRepository(i)
	webhooksstore.RegisterWebhooksStore(i)
	oauth2clientsstore.RegisterOAuth2ClientsStore(i)
	paymentsrepo.RegisterPaymentsRepository(i)
	internalopsrepo.RegisterInternalOpsRepository(i)

	// managers
	auditmanager.RegisterAuditDataManager(i)
	authmgr.RegisterAuthManager(i)
	notificationsmanager.RegisterNotificationsDataManager(i)
	paymentsmanager.RegisterPaymentsDataManager(i)
	webhooksmanager.RegisterWebhookDataManager(i)
	settingsrepo.RegisterSettingsRepository(i)
	waitlistsrepo.RegisterWaitlistsRepository(i)
	paymentsadapters.RegisterPaymentProcessorRegistry(i)

	// services
	authsvc.RegisterAuthService(i)
	authhttpsvc.RegisterAuthHTTPService(i)
	analyticssvc.RegisterAnalyticsService(i)
	auditrepo.RegisterPlatformReader(i)
	auditbuild.RegisterAuditService(i)
	commentstargets.RegisterCommentsService(i)
	dataprivacysvc.RegisterDataPrivacyService(i)
	do.Provide[dataprivacysvc.DataPrivacyMethodPermissions](i, func(i do.Injector) (dataprivacysvc.DataPrivacyMethodPermissions, error) {
		return dataprivacysvc.ProvideMethodPermissions(), nil
	})
	identitybuild.RegisterIdentityService(i)
	internalopssvc.RegisterInternalOpsService(i)
	issuereportsbuild.RegisterIssueReportsService(i)
	notificationsbuild.RegisterNotificationsService(i)
	settingsbuild.RegisterSettingsService(i)
	signinbuild.RegisterSignInService(i)
	uploadedmediasvc.RegisterUploadedMediaService(i)
	webhooksbuild.RegisterWebhooksService(i)
	oauth2clientsbuild.RegisterOAuth2ClientsService(i)
	paymentsbuild.RegisterPaymentsService(i)
	waitlistsbuild2.RegisterWaitlistsService(i)
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
