package converters

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
)

// ConvertUserDeviceTokenToUserDeviceTokenDatabaseCreationInput builds a UserDeviceTokenDatabaseCreationInput from a UserDeviceToken.
func ConvertUserDeviceTokenToUserDeviceTokenDatabaseCreationInput(x *types.UserDeviceToken) *types.UserDeviceTokenDatabaseCreationInput {
	return &types.UserDeviceTokenDatabaseCreationInput{
		ID:            x.ID,
		DeviceToken:   x.DeviceToken,
		Platform:      x.Platform,
		BelongsToUser: x.BelongsToUser,
	}
}
