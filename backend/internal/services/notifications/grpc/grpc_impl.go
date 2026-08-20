package grpc

import (
	notificationsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/manager"
	notificationssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/notifications"

	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

const (
	o11yName = "notifications_service"
)

var _ notificationssvc.UserNotificationsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		notificationssvc.UnimplementedUserNotificationsServiceServer
		tracer               tracing.Tracer
		logger               logging.Logger
		notificationsManager notificationsmanager.NotificationsDataManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	notificationsManager notificationsmanager.NotificationsDataManager,
) notificationssvc.UserNotificationsServiceServer {
	return &serviceImpl{
		logger:               logging.NewNamedLogger(logger, o11yName),
		tracer:               tracing.NewNamedTracer(tracerProvider, o11yName),
		notificationsManager: notificationsManager,
	}
}
