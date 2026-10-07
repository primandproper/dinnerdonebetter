package grpcapi

import (
	"context"
	"maps"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identitybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/identity"
	issuereportsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/issuereports"
	mediaregistrybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/mediaregistry"
	oauth2clientsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/oauth2clients"
	passkeysbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/passkeys"
	signinbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/signin"
	waitlistsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	internalopssvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	mealplanningsvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	internalopsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/grpc"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc"

	auditpb "github.com/primandproper/platform-go/v15/audit/auditpb"
	auditgrpc "github.com/primandproper/platform-go/v15/audit/grpc"
	"github.com/primandproper/platform-go/v15/authentication/oauth2clients/oauth2clientspb"
	"github.com/primandproper/platform-go/v15/authentication/passkeys/passkeyspb"
	"github.com/primandproper/platform-go/v15/authentication/passwordreset/passwordresetpb"
	"github.com/primandproper/platform-go/v15/authentication/signin/signinpb"
	billingpb "github.com/primandproper/platform-go/v15/billing/billingpb"
	paymentsgrpc "github.com/primandproper/platform-go/v15/billing/grpc"
	commentspb "github.com/primandproper/platform-go/v15/comments/commentspb"
	commentsgrpc "github.com/primandproper/platform-go/v15/comments/grpc"
	"github.com/primandproper/platform-go/v15/identity/identitypb"
	issuereportsgrpc "github.com/primandproper/platform-go/v15/issuereports/grpc"
	issuereportspb "github.com/primandproper/platform-go/v15/issuereports/issuereportspb"
	mediaregistrygrpc "github.com/primandproper/platform-go/v15/mediaregistry/grpc"
	notificationsgrpc "github.com/primandproper/platform-go/v15/notifications/grpc"
	notificationspb "github.com/primandproper/platform-go/v15/notifications/notificationspb"
	settingsgrpc "github.com/primandproper/platform-go/v15/settings/grpc"
	settingspb "github.com/primandproper/platform-go/v15/settings/settingspb"
	waitlistspb "github.com/primandproper/platform-go/v15/waitlists/waitlistspb"
	webhooksgrpc "github.com/primandproper/platform-go/v15/webhooks/grpc"
	webhookspb "github.com/primandproper/platform-go/v15/webhooks/webhookspb"
	authzgrpc "github.com/primandproper/primitives-go/v2/authorization/grpc"
	"github.com/primandproper/primitives-go/v2/database"
	errorsgrpc "github.com/primandproper/primitives-go/v2/errors/grpc"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/ratelimiting"
	textsearchcfg "github.com/primandproper/primitives-go/v2/search/text/config"
	platformgrpc "github.com/primandproper/primitives-go/v2/server/grpc"

	"github.com/samber/do/v2"
	grpc "google.golang.org/grpc"
)

// RegisterExtras registers the helper functions with the injector.
func RegisterExtras(i do.Injector) {
	do.Provide(i, func(i do.Injector) (identityindexing.UserTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tracerProvider := do.MustInvoke[tracing.Provider](i)
		metricsProvider := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideUserTextSearcher(ctx, logger, tracerProvider, metricsProvider, cfg)
	})

	do.Provide(i, func(do.Injector) (interceptors.MethodPermissionsMap, error) {
		return MethodPermissions(), nil
	})

	// One enforcer, shared by both chains, so a stream and a unary call are refused by the same
	// table.
	do.Provide(i, func(i do.Injector) (*authzgrpc.Enforcer, error) {
		return ProvideAuthorizationEnforcer(
			MethodPermissionFragments(),
			MethodPermissionOverrides(),
			do.MustInvoke[*interceptors.AuthInterceptor](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[metrics.Provider](i),
			auditOnlyAuthorization,
		)
	})

	do.Provide(i, func(i do.Injector) ([]grpc.UnaryServerInterceptor, error) {
		logger := do.MustInvoke[logging.Logger](i)
		authInterceptor := do.MustInvoke[*interceptors.AuthInterceptor](i)
		authzEnforcer := do.MustInvoke[*authzgrpc.Enforcer](i)

		idempotencyInterceptor, err := ProvideIdempotencyInterceptor(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[*config.APIServiceConfig](i),
			logger,
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[database.Client](i),
		)
		if err != nil {
			return nil, err
		}

		throttle, err := interceptors.NewAnonymousDoorThrottle(
			do.MustInvoke[ratelimiting.RateLimiter](i),
			logger,
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
		if err != nil {
			return nil, err
		}

		return BuildUnaryServerInterceptors(throttle, authInterceptor, authzEnforcer, idempotencyInterceptor), nil
	})

	do.Provide(i, func(i do.Injector) ([]grpc.StreamServerInterceptor, error) {
		return BuildStreamServerInterceptors(
			do.MustInvoke[*interceptors.AuthInterceptor](i),
			do.MustInvoke[*authzgrpc.Enforcer](i),
		), nil
	})

	do.Provide(i, func(i do.Injector) ([]platformgrpc.RegistrationFunc, error) {
		return BuildRegistrationFuncs(
			do.MustInvoke[auditpb.AuditServiceServer](i),
			do.MustInvoke[commentspb.CommentsServiceServer](i),
			do.MustInvoke[identitypb.IdentityServiceServer](i),
			do.MustInvoke[internalopssvcpb.InternalOperationsServer](i),
			do.MustInvoke[issuereportspb.IssueReportsServiceServer](i),
			do.MustInvoke[mealplanningsvcpb.MealPlanningServiceServer](i),
			do.MustInvoke[notificationspb.NotificationsServiceServer](i),
			do.MustInvoke[oauth2clientspb.OAuth2ClientsServiceServer](i),
			do.MustInvoke[billingpb.BillingServiceServer](i),
			do.MustInvoke[settingspb.SettingsServiceServer](i),
			do.MustInvoke[signinpb.SignInServiceServer](i),
			do.MustInvoke[passwordresetpb.PasswordResetServiceServer](i),
			do.MustInvoke[passkeyspb.PasskeysServiceServer](i),
			do.MustInvoke[*mediaregistrygrpc.Server](i),
			do.MustInvoke[waitlistspb.WaitlistsServiceServer](i),
			do.MustInvoke[webhookspb.WebhooksServiceServer](i),
		), nil
	})

	do.Provide(i, func(i do.Injector) (*GRPCService, error) {
		return NewGRPCService(
			do.MustInvoke[auditpb.AuditServiceServer](i),
			do.MustInvoke[identitypb.IdentityServiceServer](i),
			do.MustInvoke[internalopssvcpb.InternalOperationsServer](i),
			do.MustInvoke[issuereportspb.IssueReportsServiceServer](i),
			do.MustInvoke[mealplanningsvcpb.MealPlanningServiceServer](i),
			do.MustInvoke[notificationspb.NotificationsServiceServer](i),
			do.MustInvoke[oauth2clientspb.OAuth2ClientsServiceServer](i),
			do.MustInvoke[billingpb.BillingServiceServer](i),
			do.MustInvoke[settingspb.SettingsServiceServer](i),
			do.MustInvoke[webhookspb.WebhooksServiceServer](i),
			do.MustInvoke[waitlistspb.WaitlistsServiceServer](i),
			do.MustInvoke[*platformgrpc.Server](i),
		), nil
	})
}

func BuildRegistrationFuncs(
	auditLogService auditpb.AuditServiceServer,
	commentsService commentspb.CommentsServiceServer,
	identityServiceServer identitypb.IdentityServiceServer,
	internalOpsService internalopssvcpb.InternalOperationsServer,
	issueReportsService issuereportspb.IssueReportsServiceServer,
	mealPlanningService mealplanningsvcpb.MealPlanningServiceServer,
	notificationsService notificationspb.NotificationsServiceServer,
	oauth2ClientsService oauth2clientspb.OAuth2ClientsServiceServer,
	paymentsService billingpb.BillingServiceServer,
	settingsService settingspb.SettingsServiceServer,
	signInService signinpb.SignInServiceServer,
	passwordResetService passwordresetpb.PasswordResetServiceServer,
	passkeysService passkeyspb.PasskeysServiceServer,
	mediaRegistryService *mediaregistrygrpc.Server,
	waitlistsService waitlistspb.WaitlistsServiceServer,
	webhooksService webhookspb.WebhooksServiceServer,
) []platformgrpc.RegistrationFunc {
	return []platformgrpc.RegistrationFunc{
		func(server *grpc.Server) {
			registerWithAdministration(server, auditLogService, func(server *grpc.Server) {
				auditpb.RegisterAuditServiceServer(server, auditLogService)
			})
			commentspb.RegisterCommentsServiceServer(server, commentsService)
			identitypb.RegisterIdentityServiceServer(server, identityServiceServer)
			internalopssvcpb.RegisterInternalOperationsServer(server, internalOpsService)
			issuereportspb.RegisterIssueReportsServiceServer(server, issueReportsService)
			mealplanningsvcpb.RegisterMealPlanningServiceServer(server, mealPlanningService)
			notificationspb.RegisterNotificationsServiceServer(server, notificationsService)
			oauth2clientspb.RegisterOAuth2ClientsServiceServer(server, oauth2ClientsService)
			billingpb.RegisterBillingServiceServer(server, paymentsService)
			settingspb.RegisterSettingsServiceServer(server, settingsService)
			registerWithAdministration(server, signInService, func(server *grpc.Server) {
				signinpb.RegisterSignInServiceServer(server, signInService)
			})
			passwordresetpb.RegisterPasswordResetServiceServer(server, passwordResetService)
			passkeyspb.RegisterPasskeysServiceServer(server, passkeysService)
			mediaRegistryService.RegisterOn(server)
			waitlistspb.RegisterWaitlistsServiceServer(server, waitlistsService)
			webhookspb.RegisterWebhooksServiceServer(server, webhooksService)
		},
	}
}

// registerWithAdministration mounts a platform surface through its own RegisterOn where it has
// one, which registers the surface's operator half beside it — AuditAdministrationService beside
// AuditService, SignInAdministrationService beside SignInService. Mounting the half grants
// nothing: each of its methods requires a permission only an operator holds. A server without
// RegisterOn, a test double say, is registered as the plain service.
func registerWithAdministration(server *grpc.Server, impl any, plain func(*grpc.Server)) {
	if registrar, ok := impl.(interface{ RegisterOn(*grpc.Server) }); ok {
		registrar.RegisterOn(server)
		return
	}

	plain(server)
}

// The interceptors this server adds, inside the two primitives-go's server installs ahead of every
// list it is given: RecoveryInterceptor, outermost, and the logging interceptor after it. Recovery
// is not repeated here. A second one would sit inside the first and catch nothing it does not.
//
// The first of this server's own strips the encoded error chain. The error encoding interceptors
// put a failure on the wire twice: once as a client-safe status message, and once as the whole
// wrapped chain, encoded into the status details so a trusted peer can reconstruct it with
// errorsgrpc.DecodeErrorFromStatus. That second copy is unredacted — table names, the rule a query
// broke, which of two refusals signin deliberately answers identically — and primitives-go is
// explicit that a server reachable by untrusted clients must strip it at the edge.
//
// This server is that edge. The iOS app and the web frontend dial it directly, and nothing between
// the handler and their transport removes a detail. Nor is there a trusted peer on the far side to
// keep it for: nothing that calls this server decodes the chain. So it is stripped here, for every
// method, rather than per service or behind a flag somebody has to remember to set.
//
// Only the encoded chain goes. The code, the message and the google.rpc.ErrorInfo a client branches
// on are left exactly as the encoder built them — see errorsgrpc.StripEncodedErrorDetail.

// BuildUnaryServerInterceptors is the unary chain this server adds, outermost first.
func BuildUnaryServerInterceptors(
	throttle grpc.UnaryServerInterceptor,
	authInterceptor *interceptors.AuthInterceptor,
	authzEnforcer *authzgrpc.Enforcer,
	idempotencyInterceptor grpc.UnaryServerInterceptor,
) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		// First, so nothing any interceptor below returns reaches a client with the encoded error
		// chain still attached. It must sit outside the error encoder, which is what attaches it.
		errorsgrpc.StripEncodedErrorDetailUnaryServerInterceptor(),
		// Outside authentication, so the refusals the interceptors below make are encoded the
		// way a handler's are — with a client-safe reason where the error names one, which is
		// how a forced password change says PASSWORD_CHANGE_REQUIRED.
		errorsgrpc.UnaryErrorEncodingInterceptor(),
		// Ahead of authentication: the doors it throttles read no credential, and a caller it
		// refuses should cost nothing past the bucket.
		throttle,
		authInterceptor.UnaryServerInterceptor(),
		// Runs after the interceptor above so it sees the session that one established. It is
		// the only permission check: the interceptor above decides who is calling, not what
		// they may do.
		authzEnforcer.UnaryServerInterceptor(),
		// after auth, because the key is scoped to the authenticated principal, and inside the
		// error encoder, because it records the handler's status code rather than a rendered one.
		idempotencyInterceptor,
	}
}

// BuildStreamServerInterceptors is the stream chain this server adds, outermost first.
//
// It enforces the method permission table exactly as the unary chain does. A stream —
// MediaRegistryService.UploadObject is one — is no less a call than a unary one, and a chain
// without the enforcer would admit any signed-in caller to it.
func BuildStreamServerInterceptors(
	authInterceptor *interceptors.AuthInterceptor,
	authzEnforcer *authzgrpc.Enforcer,
) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		// First, for the reason the unary chain's strip is.
		errorsgrpc.StripEncodedErrorDetailStreamServerInterceptor(),
		errorsgrpc.StreamErrorEncodingInterceptor(),
		authInterceptor.StreamServerInterceptor(),
		// After authentication, for the reason the unary chain's enforcer is.
		authzEnforcer.StreamServerInterceptor(),
	}
}

func ProvideUserTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (identityindexing.UserTextSearcher, error) {
	return textsearchcfg.NewIndex[identityindexing.UserSearchSubset](
		ctx,
		cfg,
		identityindexing.IndexTypeUsers,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

// MethodPermissions is the table the server's authentication interceptor enforces: every method
// this deployment serves, and the permissions a caller must hold to make it. It is each surface's
// fragment with this deployment's overrides applied.
//
// It is assembled here rather than in the injector so that what reads it outside the server —
// the enforcer equivalence proof, and the conformance harness deriving which calls this
// deployment reserves to an operator — reads the table the server enforces rather than a copy.
//
// The authorization enforcer does not read it. It is built from the two halves separately —
// MethodPermissionFragments declared, MethodPermissionOverrides applied through
// RequirementsBuilder.Override — so the equivalence proof compares two derivations of the table
// rather than one table with itself.
func MethodPermissions() interceptors.MethodPermissionsMap {
	out := MethodPermissionFragments()
	maps.Copy(out, MethodPermissionOverrides())

	return out
}

// MethodPermissionFragments is every surface's table as it ships, before this deployment amends
// any of it.
func MethodPermissionFragments() interceptors.MethodPermissionsMap {
	return AggregateMethodPermissions(
		auditgrpc.Permissions(),
		commentsgrpc.Permissions(),
		identitybuild.Permissions(),
		internalopsgrpc.ProvideMethodPermissions(),
		issuereportsgrpc.Permissions(),
		mealplanninggrpc.ProvideMethodPermissions(),
		notificationsgrpc.Permissions(),
		oauth2clientsbuild.Permissions(),
		passkeysbuild.Permissions(),
		paymentsgrpc.Permissions(),
		settingsgrpc.Permissions(),
		signinbuild.Permissions(),
		mediaregistrybuild.Permissions(),
		waitlistsbuild.Permissions(),
		webhooksgrpc.Permissions(),
	)
}

// MethodPermissionOverrides is where this deployment amends a platform surface's fragment: each
// entry replaces what a method the fragment already declares demands.
//
// They are kept apart from the fragments rather than written over a copy of them because the
// requirements builder checks an override against what was declared: an override naming a
// method no fragment declares — a typo, or an RPC the surface renamed — fails the build, where a
// copy edited in place would declare it quietly and leave the real method on its default.
func MethodPermissionOverrides() map[string][]authorization.Permission {
	out := map[string][]authorization.Permission{}

	maps.Copy(out, identitybuild.PermissionOverrides())
	maps.Copy(out, issuereportsbuild.PermissionOverrides())
	maps.Copy(out, waitlistsbuild.PermissionOverrides())

	return out
}

// AggregateMethodPermissions combines method permissions from all services into a single map.
func AggregateMethodPermissions(
	auditPermissions map[string][]authorization.Permission,
	commentsPermissions map[string][]authorization.Permission,
	identityPermissions map[string][]authorization.Permission,
	internalopsPermissions internalopsgrpc.InternalOpsMethodPermissions,
	issuereportsPermissions map[string][]authorization.Permission,
	mealplanningPermissions mealplanninggrpc.MealPlanningMethodPermissions,
	notificationsPermissions map[string][]authorization.Permission,
	oauth2ClientsPermissions map[string][]authorization.Permission,
	passkeysPermissions map[string][]authorization.Permission,
	paymentsPermissions map[string][]authorization.Permission,
	settingsPermissions map[string][]authorization.Permission,
	signInPermissions map[string][]authorization.Permission,
	mediaRegistryPermissions map[string][]authorization.Permission,
	waitlistsPermissions map[string][]authorization.Permission,
	webhooksPermissions map[string][]authorization.Permission,
) interceptors.MethodPermissionsMap {
	result := make(interceptors.MethodPermissionsMap)

	maps.Copy(result, auditPermissions)
	maps.Copy(result, commentsPermissions)
	maps.Copy(result, identityPermissions)
	maps.Copy(result, internalopsPermissions)
	maps.Copy(result, issuereportsPermissions)
	maps.Copy(result, mealplanningPermissions)
	maps.Copy(result, notificationsPermissions)
	maps.Copy(result, oauth2ClientsPermissions)
	maps.Copy(result, passkeysPermissions)
	maps.Copy(result, paymentsPermissions)
	maps.Copy(result, settingsPermissions)
	maps.Copy(result, signInPermissions)
	maps.Copy(result, mediaRegistryPermissions)
	maps.Copy(result, waitlistsPermissions)
	maps.Copy(result, webhooksPermissions)

	return result
}
