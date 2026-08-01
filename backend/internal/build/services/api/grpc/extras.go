package grpcapi

import (
	"context"
	"fmt"
	"maps"
	"runtime/debug"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	analyticspb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/analytics"
	auditsvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/audit"
	authsvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	commentssvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
	dataprivacysvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/dataprivacy"
	identitysvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	internalopssvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	issuereportssvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	mealplanningsvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	notificationssvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/notifications"
	oauthsvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/oauth"
	paymentssvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
	settingssvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/settings"
	uploadedmediasvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	waitlistssvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"
	webhookssvcpb "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"
	analyticsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/analytics/grpc"
	auditgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/audit/grpc"
	authgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/auth/grpc"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"
	commentsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/comments/grpc"
	dataprivacygrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/dataprivacy/grpc"
	identitygrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/grpc"
	identityindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/indexing"
	internalopsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/internalops/grpc"
	issuereportsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/issuereports/grpc"
	mealplanninggrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/grpc"
	notificationsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/notifications/grpc"
	oauthgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/oauth/grpc"
	paymentsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/payments/grpc"
	settingsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/settings/grpc"
	uploadedmediagrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc"
	waitlistsgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/waitlists/grpc"
	webhooksgrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/webhooks/grpc"

	analyticscfg "github.com/primandproper/platform-go/v8/analytics/config"
	authzgrpc "github.com/primandproper/platform-go/v8/authorization/grpc"
	"github.com/primandproper/platform-go/v8/database"
	errorsgrpc "github.com/primandproper/platform-go/v8/errors/grpc"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/metrics"
	"github.com/primandproper/platform-go/v8/observability/tracing"
	textsearchcfg "github.com/primandproper/platform-go/v8/search/text/config"
	platformgrpc "github.com/primandproper/platform-go/v8/server/grpc"

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
		tracerProvider := do.MustInvoke[tracing.TracerProvider](i)
		metricsProvider := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideUserTextSearcher(ctx, logger, tracerProvider, metricsProvider, cfg)
	})

	do.Provide(i, func(i do.Injector) (interceptors.MethodPermissionsMap, error) {
		return AggregateMethodPermissions(
			do.MustInvoke[analyticsgrpc.AnalyticsMethodPermissions](i),
			do.MustInvoke[auditgrpc.AuditMethodPermissions](i),
			do.MustInvoke[authgrpc.AuthMethodPermissions](i),
			do.MustInvoke[commentsgrpc.CommentsMethodPermissions](i),
			do.MustInvoke[dataprivacygrpc.DataPrivacyMethodPermissions](i),
			do.MustInvoke[identitygrpc.IdentityMethodPermissions](i),
			do.MustInvoke[internalopsgrpc.InternalOpsMethodPermissions](i),
			do.MustInvoke[issuereportsgrpc.IssueReportsMethodPermissions](i),
			do.MustInvoke[mealplanninggrpc.MealPlanningMethodPermissions](i),
			do.MustInvoke[notificationsgrpc.NotificationsMethodPermissions](i),
			do.MustInvoke[oauthgrpc.OAuthMethodPermissions](i),
			do.MustInvoke[paymentsgrpc.PaymentsMethodPermissions](i),
			do.MustInvoke[settingsgrpc.SettingsMethodPermissions](i),
			do.MustInvoke[uploadedmediagrpc.UploadedMediaMethodPermissions](i),
			do.MustInvoke[waitlistsgrpc.WaitlistsMethodPermissions](i),
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
			do.MustInvoke[tracing.TracerProvider](i),
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
			do.MustInvoke[auditsvcpb.AuditServiceServer](i),
			do.MustInvoke[authsvcpb.AuthServiceServer](i),
			do.MustInvoke[commentssvcpb.CommentsServiceServer](i),
			do.MustInvoke[dataprivacysvcpb.DataPrivacyServiceServer](i),
			do.MustInvoke[identitysvcpb.IdentityServiceServer](i),
			do.MustInvoke[internalopssvcpb.InternalOperationsServer](i),
			do.MustInvoke[issuereportssvcpb.IssueReportsServiceServer](i),
			do.MustInvoke[mealplanningsvcpb.MealPlanningServiceServer](i),
			do.MustInvoke[notificationssvcpb.UserNotificationsServiceServer](i),
			do.MustInvoke[oauthsvcpb.OAuthServiceServer](i),
			do.MustInvoke[paymentssvcpb.PaymentsServiceServer](i),
			do.MustInvoke[settingssvcpb.SettingsServiceServer](i),
			do.MustInvoke[uploadedmediasvcpb.UploadedMediaServiceServer](i),
			do.MustInvoke[waitlistssvcpb.WaitlistsServiceServer](i),
			do.MustInvoke[webhookssvcpb.WebhooksServiceServer](i),
		), nil
	})

	do.Provide(i, func(i do.Injector) (*GRPCService, error) {
		return NewGRPCService(
			do.MustInvoke[auditsvcpb.AuditServiceServer](i),
			do.MustInvoke[authsvcpb.AuthServiceServer](i),
			do.MustInvoke[dataprivacysvcpb.DataPrivacyServiceServer](i),
			do.MustInvoke[identitysvcpb.IdentityServiceServer](i),
			do.MustInvoke[internalopssvcpb.InternalOperationsServer](i),
			do.MustInvoke[issuereportssvcpb.IssueReportsServiceServer](i),
			do.MustInvoke[mealplanningsvcpb.MealPlanningServiceServer](i),
			do.MustInvoke[notificationssvcpb.UserNotificationsServiceServer](i),
			do.MustInvoke[oauthsvcpb.OAuthServiceServer](i),
			do.MustInvoke[paymentssvcpb.PaymentsServiceServer](i),
			do.MustInvoke[settingssvcpb.SettingsServiceServer](i),
			do.MustInvoke[uploadedmediasvcpb.UploadedMediaServiceServer](i),
			do.MustInvoke[webhookssvcpb.WebhooksServiceServer](i),
			do.MustInvoke[waitlistssvcpb.WaitlistsServiceServer](i),
			do.MustInvoke[*platformgrpc.Server](i),
		), nil
	})
}

func BuildRegistrationFuncs(
	analyticsService analyticspb.AnalyticsServiceServer,
	auditLogService auditsvcpb.AuditServiceServer,
	authService authsvcpb.AuthServiceServer,
	commentsService commentssvcpb.CommentsServiceServer,
	dataPrivacyServer dataprivacysvcpb.DataPrivacyServiceServer,
	identityServiceServer identitysvcpb.IdentityServiceServer,
	internalOpsService internalopssvcpb.InternalOperationsServer,
	issueReportsService issuereportssvcpb.IssueReportsServiceServer,
	mealPlanningService mealplanningsvcpb.MealPlanningServiceServer,
	notificationsService notificationssvcpb.UserNotificationsServiceServer,
	oauthService oauthsvcpb.OAuthServiceServer,
	paymentsService paymentssvcpb.PaymentsServiceServer,
	settingsService settingssvcpb.SettingsServiceServer,
	uploadedMediaService uploadedmediasvcpb.UploadedMediaServiceServer,
	waitlistsService waitlistssvcpb.WaitlistsServiceServer,
	webhooksService webhookssvcpb.WebhooksServiceServer,
) []platformgrpc.RegistrationFunc {
	return []platformgrpc.RegistrationFunc{
		func(server *grpc.Server) {
			analyticspb.RegisterAnalyticsServiceServer(server, analyticsService)
			auditsvcpb.RegisterAuditServiceServer(server, auditLogService)
			authsvcpb.RegisterAuthServiceServer(server, authService)
			commentssvcpb.RegisterCommentsServiceServer(server, commentsService)
			dataprivacysvcpb.RegisterDataPrivacyServiceServer(server, dataPrivacyServer)
			identitysvcpb.RegisterIdentityServiceServer(server, identityServiceServer)
			internalopssvcpb.RegisterInternalOperationsServer(server, internalOpsService)
			issuereportssvcpb.RegisterIssueReportsServiceServer(server, issueReportsService)
			mealplanningsvcpb.RegisterMealPlanningServiceServer(server, mealPlanningService)
			notificationssvcpb.RegisterUserNotificationsServiceServer(server, notificationsService)
			oauthsvcpb.RegisterOAuthServiceServer(server, oauthService)
			paymentssvcpb.RegisterPaymentsServiceServer(server, paymentsService)
			settingssvcpb.RegisterSettingsServiceServer(server, settingsService)
			uploadedmediasvcpb.RegisterUploadedMediaServiceServer(server, uploadedMediaService)
			waitlistssvcpb.RegisterWaitlistsServiceServer(server, waitlistsService)
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
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (identityindexing.UserTextSearcher, error) {
	return textsearchcfg.NewIndex[identityindexing.UserSearchSubset](
		ctx,
		logger,
		tracerProvider, metricsProvider,
		cfg,
		identityindexing.IndexTypeUsers,
	)
}

// AggregateMethodPermissions combines method permissions from all services into a single map.
func AggregateMethodPermissions(
	analyticsPermissions analyticsgrpc.AnalyticsMethodPermissions,
	auditPermissions auditgrpc.AuditMethodPermissions,
	authPermissions authgrpc.AuthMethodPermissions,
	commentsPermissions commentsgrpc.CommentsMethodPermissions,
	dataprivacyPermissions dataprivacygrpc.DataPrivacyMethodPermissions,
	identityPermissions identitygrpc.IdentityMethodPermissions,
	internalopsPermissions internalopsgrpc.InternalOpsMethodPermissions,
	issuereportsPermissions issuereportsgrpc.IssueReportsMethodPermissions,
	mealplanningPermissions mealplanninggrpc.MealPlanningMethodPermissions,
	notificationsPermissions notificationsgrpc.NotificationsMethodPermissions,
	oauthPermissions oauthgrpc.OAuthMethodPermissions,
	paymentsPermissions paymentsgrpc.PaymentsMethodPermissions,
	settingsPermissions settingsgrpc.SettingsMethodPermissions,
	uploadedmediaPermissions uploadedmediagrpc.UploadedMediaMethodPermissions,
	waitlistsPermissions waitlistsgrpc.WaitlistsMethodPermissions,
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
