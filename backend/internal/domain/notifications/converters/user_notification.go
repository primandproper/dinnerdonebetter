package converters

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"

	"github.com/primandproper/platform-go/v11/identifiers"
)

// ConvertUserNotificationToUserNotificationUpdateRequestInput creates a UserNotificationUpdateRequestInput from a UserNotification.
func ConvertUserNotificationToUserNotificationUpdateRequestInput(x *types.UserNotification) *types.UserNotificationUpdateRequestInput {
	out := &types.UserNotificationUpdateRequestInput{
		Status: &x.Status,
	}

	return out
}

// ConvertUserNotificationCreationRequestInputToUserNotificationDatabaseCreationInput creates a UserNotificationDatabaseCreationInput from a UserNotificationCreationRequestInput.
func ConvertUserNotificationCreationRequestInputToUserNotificationDatabaseCreationInput(x *types.UserNotificationCreationRequestInput) *types.UserNotificationDatabaseCreationInput {
	out := &types.UserNotificationDatabaseCreationInput{
		ID:            identifiers.New(),
		Content:       x.Content,
		BelongsToUser: x.BelongsToUser,
	}

	return out
}

// ConvertUserNotificationToUserNotificationCreationRequestInput builds a UserNotification from a UserNotificationCreationRequestInput.
func ConvertUserNotificationToUserNotificationCreationRequestInput(x *types.UserNotification) *types.UserNotificationCreationRequestInput {
	return &types.UserNotificationCreationRequestInput{
		Content:       x.Content,
		Status:        x.Status,
		BelongsToUser: x.BelongsToUser,
	}
}

// ConvertUserNotificationToUserNotificationDatabaseCreationInput builds a UserNotificationDatabaseCreationInput from a UserNotification.
func ConvertUserNotificationToUserNotificationDatabaseCreationInput(x *types.UserNotification) *types.UserNotificationDatabaseCreationInput {
	return &types.UserNotificationDatabaseCreationInput{
		ID:            x.ID,
		Content:       x.Content,
		BelongsToUser: x.BelongsToUser,
	}
}
