package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"

	"github.com/primandproper/platform-go/v5/filtering"
)

// BuildFakeUserDeviceToken builds a faked user device token.
// iOS APNs tokens must be 64 hex chars; this uses a valid placeholder.
func BuildFakeUserDeviceToken() *types.UserDeviceToken {
	return &types.UserDeviceToken{
		CreatedAt:     BuildFakeTime(),
		ID:            BuildFakeID(),
		DeviceToken:   "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456",
		Platform:      types.UserDeviceTokenPlatformIOS,
		BelongsToUser: BuildFakeID(),
	}
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
	var tokens []*types.UserDeviceToken
	for range exampleQuantity {
		tokens = append(tokens, BuildFakeUserDeviceToken())
	}

	return &filtering.QueryFilteredResult[types.UserDeviceToken]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: tokens,
	}
}
