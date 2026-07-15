package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/converters"

	"github.com/primandproper/platform-go/v4/filtering"
)

// BuildFakeUserNotification builds a faked valid ingredient.
func BuildFakeUserNotification() *types.UserNotification {
	return &types.UserNotification{
		CreatedAt:     BuildFakeTime(),
		ID:            BuildFakeID(),
		Content:       buildUniqueString(),
		Status:        types.UserNotificationStatusTypeUnread,
		BelongsToUser: BuildFakeID(),
	}
}

// BuildFakeUserNotificationsList builds a faked UserNotificationList.
func BuildFakeUserNotificationsList() *filtering.QueryFilteredResult[types.UserNotification] {
	var notifications []*types.UserNotification
	for range exampleQuantity {
		notifications = append(notifications, BuildFakeUserNotification())
	}

	return &filtering.QueryFilteredResult[types.UserNotification]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: notifications,
	}
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
