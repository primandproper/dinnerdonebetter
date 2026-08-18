package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
)

// BuildFakeUserDeviceToken builds a faked user device token.
func BuildFakeUserDeviceToken() *types.UserDeviceToken {
	token := fake.BuildFakeRecord[types.UserDeviceToken]()

	// An APNs token is sixty-four hex characters, and the registration path checks that
	// before it will store one. This is a valid placeholder rather than a real token.
	token.DeviceToken = "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
	token.Platform = types.UserDeviceTokenPlatformIOS

	return token
}

// BuildFakeUserDeviceTokenDatabaseCreationInput builds a faked UserDeviceTokenDatabaseCreationInput.
func BuildFakeUserDeviceTokenDatabaseCreationInput() *types.UserDeviceTokenDatabaseCreationInput {
	token := BuildFakeUserDeviceToken()

	return &types.UserDeviceTokenDatabaseCreationInput{
		ID:            token.ID,
		DeviceToken:   token.DeviceToken,
		Platform:      token.Platform,
		BelongsToUser: token.BelongsToUser,
	}
}

// BuildFakeUserDeviceTokensList builds a faked list of user device tokens.
func BuildFakeUserDeviceTokensList() *filtering.QueryFilteredResult[types.UserDeviceToken] {
	return fake.BuildFakePage(BuildFakeUserDeviceToken)
}
