package authorization

import (
	notificationsgrpc "github.com/primandproper/platform-go/v14/notifications/grpc"
)

// The notification permissions are platform's, re-exported under the names this
// application's policy already spells. See comments_permissions.go.
//
// Two local grants have no platform counterpart, and their absence is the
// point. CreateUserNotifications gated nothing on the wire — this application
// has never had an RPC for it, because a notification is written by the system
// that has something to say rather than requested by the person it is about,
// and platform's surface takes the same position. UpdateUserNotifications was
// the deleted UpdateUserNotification RPC, whose one real effect was marking a
// notification read; that is MarkInboxRead here, which says what it does.
const (
	// ReadUserNotificationsPermission is a permission.
	ReadUserNotificationsPermission = notificationsgrpc.PermissionReadInbox
	// MarkUserNotificationsReadPermission is a permission.
	MarkUserNotificationsReadPermission = notificationsgrpc.PermissionMarkInboxRead
	// ArchiveUserNotificationsPermission is a permission.
	ArchiveUserNotificationsPermission = notificationsgrpc.PermissionArchiveInbox
	// CreateUserDeviceTokensPermission registers a handset.
	CreateUserDeviceTokensPermission = notificationsgrpc.PermissionRegisterDevices
	// ReadUserDeviceTokensPermission is a permission.
	ReadUserDeviceTokensPermission = notificationsgrpc.PermissionReadDevices
	// ArchiveUserDeviceTokensPermission revokes a handset.
	ArchiveUserDeviceTokensPermission = notificationsgrpc.PermissionRevokeDevices
)
