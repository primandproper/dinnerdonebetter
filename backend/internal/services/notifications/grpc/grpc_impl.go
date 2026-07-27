package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	notificationsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/manager"
	notificationssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/notifications"

	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/tracing"
)

const (
	o11yName = "notifications_service"
)

var _ notificationssvc.UserNotificationsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		notificationssvc.UnimplementedUserNotificationsServiceServer
		tracer                    tracing.Tracer
		logger                    logging.Logger
		sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error)
		notificationsManager      notificationsmanager.NotificationsDataManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	notificationsManager notificationsmanager.NotificationsDataManager,
) notificationssvc.UserNotificationsServiceServer {
	return &serviceImpl{
		logger:                    logging.NewNamedLogger(logger, o11yName),
		tracer:                    tracing.NewNamedTracer(tracerProvider, o11yName),
		notificationsManager:      notificationsManager,
		sessionContextDataFetcher: sessions.FetchContextDataFromContext,
	}
}
