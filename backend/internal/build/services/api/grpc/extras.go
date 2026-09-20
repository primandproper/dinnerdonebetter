package grpcapi

import (
	"context"
	"fmt"
	"maps"
	"runtime/debug"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	waitlistsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	analyticspb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/analytics"
	authsvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	dataprivacysvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"
	identitysvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	internalopssvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	mealplanningsvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	notificationssvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/notifications"
	oauthsvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/oauth"
	paymentssvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
	uploadedmediasvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	webhookssvcpb "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"
	analyticsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/analytics/grpc"
	authgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"
	dataprivacygrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/grpc"
	identitygrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/grpc"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	internalopsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/grpc"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc"
	notificationsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/notifications/grpc"
	oauthgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/oauth/grpc"
	paymentsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/grpc"
	uploadedmediagrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc"
	webhooksgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/webhooks/grpc"
	auditpb "github.com/primandproper/platform-go/v14/audit/auditpb"
	auditgrpc "github.com/primandproper/platform-go/v14/audit/grpc"
	commentspb "github.com/primandproper/platform-go/v14/comments/commentspb"
	commentsgrpc "github.com/primandproper/platform-go/v14/comments/grpc"
	issuereportsgrpc "github.com/primandproper/platform-go/v14/issuereports/grpc"
	issuereportspb "github.com/primandproper/platform-go/v14/issuereports/issuereportspb"
	settingsgrpc "github.com/primandproper/platform-go/v14/settings/grpc"
	settingspb "github.com/primandproper/platform-go/v14/settings/settingspb"
	waitlistspb "github.com/primandproper/platform-go/v14/waitlists/waitlistspb"

	analyticscfg "github.com/primandproper/primitives-go/v2/analytics/config"
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
	do.Provide(i, func(i do.Injector) (map[string]*analyticscfg.SourceConfig, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return ProvideAnalyticsProxySources(cfg), nil
	})

	do.Provide(i, func(i do.Injector) (identityindexing.UserTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tracerProvider := do.MustInvoke[tracing.Provider](i)
		metricsProvider := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideUserTextSearcher(ctx, logger, tracerProvider, metricsProvider, cfg)
	})

	do.Provide(i, func(i do.Injector) (interceptors.MethodPermissionsMap, error) {
		return AggregateMethodPermissions(
			do.MustInvoke[analyticsgrpc.AnalyticsMethodPermissions](i),
			auditgrpc.Permissions(),
			do.MustInvoke[authgrpc.AuthMethodPermissions](i),
			commentsgrpc.Permissions(),
			do.MustInvoke[dataprivacygrpc.DataPrivacyMethodPermissions](i),
			do.MustInvoke[identitygrpc.IdentityMethodPermissions](i),
			do.MustInvoke[internalopsgrpc.InternalOpsMethodPermissions](i),
			issuereportsgrpc.Permissions(),
			do.MustInvoke[mealplanninggrpc.MealPlanningMethodPermissions](i),
			do.MustInvoke[notificationsgrpc.NotificationsMethodPermissions](i),
			do.MustInvoke[oauthgrpc.OAuthMethodPermissions](i),
			do.MustInvoke[paymentsgrpc.PaymentsMethodPermissions](i),
			settingsgrpc.Permissions(),
			do.MustInvoke[uploadedmediagrpc.UploadedMediaMethodPermissions](i),
			waitlistsbuild.Permissions(),
			do.MustInvoke[webhooksgrpc.WebhooksMethodPermissions](i),
		), nil
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
			do.MustInvoke[analyticspb.AnalyticsServiceServer](i),
			do.MustInvoke[auditpb.AuditServiceServer](i),
			do.MustInvoke[authsvcpb.AuthServiceServer](i),
			do.MustInvoke[commentspb.CommentsServiceServer](i),
			do.MustInvoke[dataprivacysvcpb.DataPrivacyServiceServer](i),
			do.MustInvoke[identitysvcpb.IdentityServiceServer](i),
			do.MustInvoke[internalopssvcpb.InternalOperationsServer](i),
			do.MustInvoke[issuereportspb.IssueReportsServiceServer](i),
			do.MustInvoke[mealplanningsvcpb.MealPlanningServiceServer](i),
			do.MustInvoke[notificationssvcpb.UserNotificationsServiceServer](i),
			do.MustInvoke[oauthsvcpb.OAuthServiceServer](i),
			do.MustInvoke[paymentssvcpb.PaymentsServiceServer](i),
			do.MustInvoke[settingspb.SettingsServiceServer](i),
			do.MustInvoke[uploadedmediasvcpb.UploadedMediaServiceServer](i),
			do.MustInvoke[waitlistspb.WaitlistsServiceServer](i),
			do.MustInvoke[webhookssvcpb.WebhooksServiceServer](i),
		), nil
	})

	do.Provide(i, func(i do.Injector) (*GRPCService, error) {
		return NewGRPCService(
			do.MustInvoke[auditpb.AuditServiceServer](i),
			do.MustInvoke[authsvcpb.AuthServiceServer](i),
			do.MustInvoke[dataprivacysvcpb.DataPrivacyServiceServer](i),
			do.MustInvoke[identitysvcpb.IdentityServiceServer](i),
			do.MustInvoke[internalopssvcpb.InternalOperationsServer](i),
			do.MustInvoke[issuereportspb.IssueReportsServiceServer](i),
			do.MustInvoke[mealplanningsvcpb.MealPlanningServiceServer](i),
			do.MustInvoke[notificationssvcpb.UserNotificationsServiceServer](i),
			do.MustInvoke[oauthsvcpb.OAuthServiceServer](i),
			do.MustInvoke[paymentssvcpb.PaymentsServiceServer](i),
			do.MustInvoke[settingspb.SettingsServiceServer](i),
			do.MustInvoke[uploadedmediasvcpb.UploadedMediaServiceServer](i),
			do.MustInvoke[webhookssvcpb.WebhooksServiceServer](i),
			do.MustInvoke[waitlistspb.WaitlistsServiceServer](i),
			do.MustInvoke[*platformgrpc.Server](i),
		), nil
	})
}

func BuildRegistrationFuncs(
	analyticsService analyticspb.AnalyticsServiceServer,
	auditLogService auditpb.AuditServiceServer,
	authService authsvcpb.AuthServiceServer,
	commentsService commentspb.CommentsServiceServer,
	dataPrivacyServer dataprivacysvcpb.DataPrivacyServiceServer,
	identityServiceServer identitysvcpb.IdentityServiceServer,
	internalOpsService internalopssvcpb.InternalOperationsServer,
	issueReportsService issuereportspb.IssueReportsServiceServer,
	mealPlanningService mealplanningsvcpb.MealPlanningServiceServer,
	notificationsService notificationssvcpb.UserNotificationsServiceServer,
	oauthService oauthsvcpb.OAuthServiceServer,
	paymentsService paymentssvcpb.PaymentsServiceServer,
	settingsService settingspb.SettingsServiceServer,
	uploadedMediaService uploadedmediasvcpb.UploadedMediaServiceServer,
	waitlistsService waitlistspb.WaitlistsServiceServer,
	webhooksService webhookssvcpb.WebhooksServiceServer,
) []platformgrpc.RegistrationFunc {
	return []platformgrpc.RegistrationFunc{
		func(server *grpc.Server) {
			analyticspb.RegisterAnalyticsServiceServer(server, analyticsService)
			auditpb.RegisterAuditServiceServer(server, auditLogService)
			authsvcpb.RegisterAuthServiceServer(server, authService)
			commentspb.RegisterCommentsServiceServer(server, commentsService)
			dataprivacysvcpb.RegisterDataPrivacyServiceServer(server, dataPrivacyServer)
			identitysvcpb.RegisterIdentityServiceServer(server, identityServiceServer)
			internalopssvcpb.RegisterInternalOperationsServer(server, internalOpsService)
			issuereportspb.RegisterIssueReportsServiceServer(server, issueReportsService)
			mealplanningsvcpb.RegisterMealPlanningServiceServer(server, mealPlanningService)
			notificationssvcpb.RegisterUserNotificationsServiceServer(server, notificationsService)
			oauthsvcpb.RegisterOAuthServiceServer(server, oauthService)
			paymentssvcpb.RegisterPaymentsServiceServer(server, paymentsService)
			settingspb.RegisterSettingsServiceServer(server, settingsService)
			uploadedmediasvcpb.RegisterUploadedMediaServiceServer(server, uploadedMediaService)
			waitlistspb.RegisterWaitlistsServiceServer(server, waitlistsService)
			webhookssvcpb.RegisterWebhooksServiceServer(server, webhooksService)
		},
	}
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
		authInterceptor.UnaryServerInterceptor(),
		// Runs after the interceptor above so it sees the session that one established.
		// Both enforce, and they are proven equivalent — see auditOnlyAuthorization.
		authzEnforcer.UnaryServerInterceptor(),
		// after auth, because the key is scoped to the authenticated principal, and before the
		// error encoder, because it records the handler's status code rather than a rendered one.
		idempotencyInterceptor,
		errorsgrpc.UnaryErrorEncodingInterceptor(),
	}
}

func BuildStreamServerInterceptors(logger logging.Logger, authInterceptor *interceptors.AuthInterceptor) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		// recovery must be outermost so it catches panics from downstream interceptors and handlers.
		RecoveryStreamServerInterceptor(logger),
		authInterceptor.StreamServerInterceptor(),
		errorsgrpc.StreamErrorEncodingInterceptor(),
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

// ProvideAnalyticsProxySources extracts proxy sources config for the multisource reporter.
func ProvideAnalyticsProxySources(cfg *config.APIServiceConfig) map[string]*analyticscfg.SourceConfig {
	return cfg.Analytics.ProxySources.ToMap()
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

// AggregateMethodPermissions combines method permissions from all services into a single map.
func AggregateMethodPermissions(
	analyticsPermissions analyticsgrpc.AnalyticsMethodPermissions,
	auditPermissions map[string][]authorization.Permission,
	authPermissions authgrpc.AuthMethodPermissions,
	commentsPermissions map[string][]authorization.Permission,
	dataprivacyPermissions dataprivacygrpc.DataPrivacyMethodPermissions,
	identityPermissions identitygrpc.IdentityMethodPermissions,
	internalopsPermissions internalopsgrpc.InternalOpsMethodPermissions,
	issuereportsPermissions map[string][]authorization.Permission,
	mealplanningPermissions mealplanninggrpc.MealPlanningMethodPermissions,
	notificationsPermissions notificationsgrpc.NotificationsMethodPermissions,
	oauthPermissions oauthgrpc.OAuthMethodPermissions,
	paymentsPermissions paymentsgrpc.PaymentsMethodPermissions,
	settingsPermissions map[string][]authorization.Permission,
	uploadedmediaPermissions uploadedmediagrpc.UploadedMediaMethodPermissions,
	waitlistsPermissions map[string][]authorization.Permission,
	webhooksPermissions webhooksgrpc.WebhooksMethodPermissions,
) interceptors.MethodPermissionsMap {
	result := make(interceptors.MethodPermissionsMap)

	maps.Copy(result, analyticsPermissions)
	maps.Copy(result, auditPermissions)
	maps.Copy(result, authPermissions)
	maps.Copy(result, commentsPermissions)
	maps.Copy(result, dataprivacyPermissions)
	maps.Copy(result, identityPermissions)
	maps.Copy(result, internalopsPermissions)
	maps.Copy(result, issuereportsPermissions)
	maps.Copy(result, mealplanningPermissions)
	maps.Copy(result, notificationsPermissions)
	maps.Copy(result, oauthPermissions)
	maps.Copy(result, paymentsPermissions)
	maps.Copy(result, settingsPermissions)
	maps.Copy(result, uploadedmediaPermissions)
	maps.Copy(result, waitlistsPermissions)
	maps.Copy(result, webhooksPermissions)

	return result
}
