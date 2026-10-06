package grpcapi

import (
	"context"
	"fmt"

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
	passkeysbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/passkeys"
	passwordresetbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/passwordreset"
	paymentsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/queuedmail"
	"github.com/primandproper/dinnerdonebetter/backend/internal/build/sagas"
	settingsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/settings"
	signinbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/signin"
	waitlistsbuild2 "github.com/primandproper/dinnerdonebetter/backend/internal/build/waitlists"
	webhooksbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	mealplanningregistration "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/registration"
	paymentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/manager"
	appentitlements "github.com/primandproper/dinnerdonebetter/backend/internal/entitlements"
	appmetering "github.com/primandproper/dinnerdonebetter/backend/internal/metering"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories"
	auditrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	authrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auth"
	commentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/comments"
	identitystore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identitystore"
	internalopsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/internalops"
	issuereportsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/issuereports"
	notificationsstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notificationsstore"
	oauth2clientsstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/oauth2clientsstore"
	paymentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/payments"
	settingsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/settings"
	uploadedmediarepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/uploadedmedia"
	waitlistsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/waitlists"
	webhooksstore "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooksstore"
	analyticssvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/analytics/grpc"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"
	authhttpsvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/handlers/authentication"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"
	internalopssvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/grpc"
	paymentsadapters "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/adapters"
	uploadedmediacfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/config"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc"

	operationscfg "github.com/primandproper/platform-go/v15/operations/config"
	"github.com/primandproper/platform-go/v15/service"
	"github.com/primandproper/primitives-go/v2/analytics/multisource"
	tokenscfg "github.com/primandproper/primitives-go/v2/authentication/tokens/config"
	"github.com/primandproper/primitives-go/v2/qrcodes"
	"github.com/primandproper/primitives-go/v2/random"
	uploadscfg "github.com/primandproper/primitives-go/v2/uploads/config"
	"github.com/primandproper/primitives-go/v2/uploads/objectstorage"

	"github.com/samber/do/v2"
)

// BuildInjector validates cfg and composes the API server from it.
//
// service.Register builds what platform owns from cfg.Service: the pillars, the database, the
// broker, the outbound HTTP client, the feature flag manager, both servers and the encoding they
// speak, the health registry those servers mount, and the transport mappings for every platform
// sentinel — the one errormappers.Register call this composition root used to make itself. What
// follows it is this application's: the repositories, the services, the domain.
//
// Validation comes first and is not optional, for the reason the scheduler's BuildInjector gives.
func BuildInjector(
	ctx context.Context,
	cfg *config.APIServiceConfig,
) (*do.RootScope, error) {
	if err := cfg.ValidateWithContext(ctx); err != nil {
		return nil, fmt.Errorf("validating API server config: %w", err)
	}

	i := do.New()

	do.ProvideValue(i, ctx)
	do.ProvideValue(i, cfg)

	service.Register(i, &cfg.Service)

	// config field extraction
	RegisterConfigs(i)

	repositories.RegisterMigrator(i)
	random.RegisterGenerator(i)
	do.ProvideValue(i, qrcodes.Issuer(branding.CompanyName))
	qrcodes.RegisterBuilder(i)
	uploadscfg.RegisterStorageConfig(i)
	objectstorage.RegisterUploadManager(i)
	// The operations tier a privacy request is submitted as. Only the enqueue-and-read half
	// runs here — the worker that claims and runs operations is in the scheduler — but the
	// registry is not optional on this side: Start looks a kind up in the registry of the
	// process calling it, so an API server without the privacy kinds registered would refuse
	// every request with ErrUnknownKind.
	dataprivacybuild.RegisterRegistry(i)
	dataprivacybuild.RegisterOperationsRegistry(i)
	dataprivacybuild.RegisterCompletionNotifier(i)
	operationscfg.RegisterStore(i)
	operationscfg.RegisterQueue(i)
	operationscfg.RegisterService(i)
	// The loop that pushes operation snapshots to the event stream operations/http serves,
	// which is where a privacy request's receipt tells a client to follow it. service.New
	// resolves and runs it.
	operationscfg.RegisterWatcher(i)

	// The request store, the artifact storage — a bucket and a keyring of its own, pointed at
	// the user data bucket rather than the media bucket the ambient upload manager above serves
	// — and the service a subject's request is submitted to. All platform's, built the way the
	// scheduler builds them, from the same block; see RegisterRequestService for why by hand.
	dataprivacycfg.RegisterRequestService(i, &cfg.Services.DataPrivacy)
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
	// The mailer the identity service hands an invitation to, and sign-in, password
	// reset and waitlists their mail. See internal/build/queuedmail.
	queuedmail.Register(i)
	identitybuild.RegisterSessionBuilder(i)
	issuereportsrepo.RegisterIssueReportsRepository(i)
	uploadedmediarepo.RegisterUploadedMediaRepository(i)
	webhooksstore.RegisterWebhooksStore(i)
	notificationsstore.RegisterNotificationsStore(i)
	oauth2clientsstore.RegisterOAuth2ClientsStore(i)
	paymentsrepo.RegisterPaymentsRepository(i)
	internalopsrepo.RegisterInternalOpsRepository(i)

	// managers
	paymentsmanager.RegisterPaymentsDataManager(i)
	settingsrepo.RegisterSettingsRepository(i)
	waitlistsrepo.RegisterWaitlistsRepository(i)
	paymentsadapters.RegisterPaymentProcessorRegistry(i)

	// services
	authhttpsvc.RegisterAuthHTTPService(i)
	analyticssvc.RegisterAnalyticsService(i)
	auditrepo.RegisterPlatformReader(i)
	auditrepo.RegisterPlatformRecorder(i)
	auditbuild.RegisterAuditService(i)
	commentstargets.RegisterCommentsService(i)
	identitybuild.RegisterIdentityService(i)
	internalopssvc.RegisterInternalOpsService(i)
	issuereportsbuild.RegisterIssueReportsService(i)
	notificationsbuild.RegisterNotificationsService(i)
	settingsbuild.RegisterSettingsService(i)
	signinbuild.RegisterSignInService(i)
	passwordresetbuild.RegisterPasswordResetService(i)
	passkeysbuild.RegisterPasskeysService(i)
	uploadedmediasvc.RegisterUploadedMediaService(i)
	webhooksbuild.RegisterWebhooksService(i)
	oauth2clientsbuild.RegisterOAuth2ClientsService(i)
	paymentsbuild.RegisterPaymentsService(i)
	waitlistsbuild2.RegisterWaitlistsService(i)
	uploadedmediacfg.RegisterUploadedMediaConfig(i)

	// The saga machinery, minus the worker: this process starts durable processes and does not
	// advance them. Registered before the domain, which puts its definitions on the registry.
	sagas.RegisterSagaStore(i)
	sagas.RegisterSagas(i)

	// Domain: mealplanning
	mealplanningregistration.RegisterForGRPCAPI(i)

	// extras (functions from extras.go)
	RegisterExtras(i)

	return i, nil
}
