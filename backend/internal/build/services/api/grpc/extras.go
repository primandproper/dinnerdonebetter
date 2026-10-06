package grpcapi

import (
	"context"
	"fmt"
	"maps"
	"runtime/debug"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identitybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/identity"
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
	textsearchcfg "github.com/primandproper/primitives-go/v2/search/text/config"
	platformgrpc "github.com/primandproper/primitives-go/v2/server/grpc"

	"github.com/samber/do/v2"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

	do.Provide(i, func(i do.Injector) ([]grpc.UnaryServerInterceptor, error) {
		logger := do.MustInvoke[logging.Logger](i)
		authInterceptor := do.MustInvoke[*interceptors.AuthInterceptor](i)

		authzEnforcer, err := ProvideAuthorizationEnforcer(
			do.MustInvoke[interceptors.MethodPermissionsMap](i),
			authInterceptor,
			logger,
			do.MustInvoke[metrics.Provider](i),
			auditOnlyAuthorization,
		)
		if err != nil {
			return nil, err
		}

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

		return BuildUnaryServerInterceptors(logger, authInterceptor, authzEnforcer, idempotencyInterceptor), nil
	})

	do.Provide(i, func(i do.Injector) ([]grpc.StreamServerInterceptor, error) {
		logger := do.MustInvoke[logging.Logger](i)
		authInterceptor := do.MustInvoke[*interceptors.AuthInterceptor](i)
		return BuildStreamServerInterceptors(logger, authInterceptor), nil
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

func BuildUnaryServerInterceptors(
	logger logging.Logger,
	authInterceptor *interceptors.AuthInterceptor,
	authzEnforcer *authzgrpc.Enforcer,
	idempotencyInterceptor grpc.UnaryServerInterceptor,
) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		// recovery must be outermost so it catches panics from downstream interceptors and handlers.
		RecoveryUnaryServerInterceptor(logger),
		// Next, so nothing any interceptor below returns reaches a client with the encoded error
		// chain still attached. See error_details.go.
		StripEncodedErrorDetailUnaryInterceptor(),
		// Outside authentication, so the refusals the interceptors below make are encoded the
		// way a handler's are — with a client-safe reason where the error names one, which is
		// how a forced password change says PASSWORD_CHANGE_REQUIRED.
		errorsgrpc.UnaryErrorEncodingInterceptor(),
		authInterceptor.UnaryServerInterceptor(),
		// Runs after the interceptor above so it sees the session that one established.
		// Both enforce, and they are proven equivalent — see auditOnlyAuthorization.
		authzEnforcer.UnaryServerInterceptor(),
		// after auth, because the key is scoped to the authenticated principal, and inside the
		// error encoder, because it records the handler's status code rather than a rendered one.
		idempotencyInterceptor,
	}
}

func BuildStreamServerInterceptors(logger logging.Logger, authInterceptor *interceptors.AuthInterceptor) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		// recovery must be outermost so it catches panics from downstream interceptors and handlers.
		RecoveryStreamServerInterceptor(logger),
		// Next, so nothing any interceptor below returns reaches a client with the encoded error
		// chain still attached. See error_details.go.
		StripEncodedErrorDetailStreamInterceptor(),
		errorsgrpc.StreamErrorEncodingInterceptor(),
		authInterceptor.StreamServerInterceptor(),
	}
}

// RecoveryUnaryServerInterceptor recovers from panics in unary handlers, logs them, and maps them to codes.Internal
// so a single nil-dereference degrades into a 500 rather than crashing the process.
func RecoveryUnaryServerInterceptor(logger logging.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.WithValue("method", info.FullMethod).WithValue("stack", string(debug.Stack())).Error("recovered from panic in gRPC unary handler", fmt.Errorf("%v", r))
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// RecoveryStreamServerInterceptor recovers from panics in stream handlers, logs them, and maps them to codes.Internal.
func RecoveryStreamServerInterceptor(logger logging.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.WithValue("method", info.FullMethod).WithValue("stack", string(debug.Stack())).Error("recovered from panic in gRPC stream handler", fmt.Errorf("%v", r))
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(srv, ss)
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

// MethodPermissions is the table the server's authorization interceptor enforces: every method
// this deployment serves, and the permissions a caller must hold to make it.
//
// It is assembled here rather than in the injector so that what reads it outside the server —
// the enforcer equivalence proof, and the conformance harness deriving which calls this
// deployment reserves to an operator — reads the table the server enforces rather than a copy.
func MethodPermissions() interceptors.MethodPermissionsMap {
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
