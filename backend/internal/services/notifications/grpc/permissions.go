package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	notificationssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/notifications"
)

// NotificationsMethodPermissions is a named type for Wire dependency injection.
type NotificationsMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the notifications service's method permissions.
func ProvideMethodPermissions() NotificationsMethodPermissions {
	return NotificationsMethodPermissions{
		notificationssvc.UserNotificationsService_GetUserNotification_FullMethodName: {
			authorization.ReadUserNotificationsPermission,
		},
		notificationssvc.UserNotificationsService_GetUserNotifications_FullMethodName: {
			authorization.ReadUserNotificationsPermission,
		},
		notificationssvc.UserNotificationsService_UpdateUserNotification_FullMethodName: {
			authorization.UpdateUserNotificationsPermission,
		},
		notificationssvc.UserNotificationsService_RegisterDeviceToken_FullMethodName: {
			authorization.CreateUserDeviceTokensPermission,
		},
		notificationssvc.UserNotificationsService_GetUserDeviceToken_FullMethodName: {
			authorization.ReadUserDeviceTokensPermission,
		},
		notificationssvc.UserNotificationsService_GetUserDeviceTokens_FullMethodName: {
			authorization.ReadUserDeviceTokensPermission,
		},
		notificationssvc.UserNotificationsService_ArchiveUserDeviceToken_FullMethodName: {
			authorization.ArchiveUserDeviceTokensPermission,
		},
	}
}
