package fakes

import (
	"crypto/rand"
	"encoding/hex"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"

	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"
)

// BuildFakeUserDeviceToken builds a faked user device token.
func BuildFakeUserDeviceToken() *types.UserDeviceToken {
	token := fake.BuildFakeRecord[types.UserDeviceToken]()

	// An APNs token is sixty-four hex characters, and the registration path checks that
	// before it will store one. Thirty-two random bytes is a valid placeholder rather
	// than a real token.
	//
	// Random and not a constant, which this was until it cost an afternoon. A
	// registration converges on (scope, platform, token) — a handset re-registering is
	// meant to keep the id it already had — and this deployment scopes notifications
	// globally, so one literal here made every user in the suite share a single row
	// owned by whoever registered first. The read-back then handed a caller an id
	// belonging to somebody else, and revoking it was correctly refused. See
	// fake.BuildFakeRecord's own reasoning: a fake that collides is a fake that tests
	// something other than what it says.
	token.DeviceToken = fakeDeviceToken()
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

// fakeDeviceToken builds sixty-four hex characters, the shape of an APNs token.
//
// crypto/rand rather than the faker's own randomness, because the uniqueness is the
// point and rand.Read does not fail: it panics or fills, so there is no error to
// swallow into a collision.
func fakeDeviceToken() string {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		panic(err)
	}

	return hex.EncodeToString(raw)
}
