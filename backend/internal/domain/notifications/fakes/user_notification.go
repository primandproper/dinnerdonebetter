package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
)

// BuildFakeUserNotification builds a faked user notification.
func BuildFakeUserNotification() *types.UserNotification {
	notification := fake.BuildFakeRecord[types.UserNotification]()

	// A notification is unread until something reads it, and read is the state a test
	// arrives at rather than starts from.
	notification.Status = types.UserNotificationStatusTypeUnread

	return notification
}

// BuildFakeUserNotificationsList builds a faked UserNotificationList.
func BuildFakeUserNotificationsList() *filtering.QueryFilteredResult[types.UserNotification] {
	return fake.BuildFakePage(BuildFakeUserNotification)
}

// BuildFakeUserNotificationUpdateRequestInput builds a faked UserNotificationUpdateRequestInput.
func BuildFakeUserNotificationUpdateRequestInput() *types.UserNotificationUpdateRequestInput {
	userNotification := BuildFakeUserNotification()

	return converters.ConvertUserNotificationToUserNotificationUpdateRequestInput(userNotification)
}

// BuildFakeUserNotificationCreationRequestInput builds a faked UserNotificationCreationRequestInput.
func BuildFakeUserNotificationCreationRequestInput() *types.UserNotificationCreationRequestInput {
	userNotification := BuildFakeUserNotification()

	return converters.ConvertUserNotificationToUserNotificationCreationRequestInput(userNotification)
}
