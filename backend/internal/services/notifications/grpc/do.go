package grpc

import (
	notificationsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/manager"
	notificationssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/notifications"

	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterNotificationsService registers the notifications gRPC service with the injector.
func RegisterNotificationsService(i do.Injector) {
	do.Provide[NotificationsMethodPermissions](i, func(i do.Injector) (NotificationsMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[notificationssvc.UserNotificationsServiceServer](i, func(i do.Injector) (notificationssvc.UserNotificationsServiceServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[notificationsmanager.NotificationsDataManager](i),
		), nil
	})
}
