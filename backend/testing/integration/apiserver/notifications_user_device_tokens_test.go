package integration

import (
	"fmt"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/fakes"

	notificationspb "github.com/primandproper/platform-go/v14/notifications/notificationspb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The device surface is platform's now, and it is three RPCs where this application's was
// four: register, list, revoke. There is no read-one, because platform's registry has none
// — a handset is identified by the token it registered with, and a caller holding an id
// already got it from the list. The cases that read one back by id are gone with it.
//
// A registered Device does not carry its token back either. The token is the credential a
// push is sent with, so rendering it over an API would hand anybody who can list devices
// the ability to push to them. That is why the assertions below check the platform and the
// id rather than the token they sent.

func createUserDeviceTokenForTest(t *testing.T, forUser string, deviceTokenOverride ...string) *notifications.UserDeviceToken {
	t.Helper()

	ctx := t.Context()

	creationInput := fakes.BuildFakeUserDeviceToken()
	input := converters.ConvertUserDeviceTokenToUserDeviceTokenDatabaseCreationInput(creationInput)
	input.BelongsToUser = forUser
	if len(deviceTokenOverride) > 0 {
		input.DeviceToken = deviceTokenOverride[0]
	}

	created, err := notifsRepo.CreateUserDeviceToken(ctx, input)
	require.NoError(t, err)
	assert.NotNil(t, created)

	return created
}

func TestUserDeviceTokens_RegisterAndRead(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		exampleToken := fakes.BuildFakeUserDeviceToken()

		response, err := testClient.RegisterDevice(ctx, &notificationspb.RegisterDeviceRequest{
			Input: &notificationspb.DeviceRegistrationInput{
				Token:    exampleToken.DeviceToken,
				Platform: notificationspb.DevicePlatform_DEVICE_PLATFORM_IOS,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, response)
		require.NotNil(t, response.GetResult())
		assert.NotEmpty(t, response.GetResult().GetId())
		assert.Equal(t, notificationspb.DevicePlatform_DEVICE_PLATFORM_IOS, response.GetResult().GetPlatform())
		assert.NotNil(t, response.GetResult().GetCreatedAt())

		// The registration is visible on the caller's own list, which is the read that
		// replaced the read-one. It is scoped to the session's user by the server, so a
		// list is never a list of somebody else's handsets.
		listed, err := testClient.ListDevices(ctx, &notificationspb.ListDevicesRequest{})
		require.NoError(t, err)
		require.NotEmpty(t, listed.GetResults())

		var found bool
		for _, device := range listed.GetResults() {
			if device.GetId() == response.GetResult().GetId() {
				found = true

				assert.Equal(t, notificationspb.DevicePlatform_DEVICE_PLATFORM_IOS, device.GetPlatform())
			}
		}
		assert.True(t, found, "the registered device was not on the caller's own list")
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ListDevices(ctx, &notificationspb.ListDevicesRequest{})
		assert.Error(t, err)
	})

	T.Run("revoking a device nobody registered is refused", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.RevokeDevice(ctx, &notificationspb.RevokeDeviceRequest{DeviceId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestUserDeviceTokens_Listing(T *testing.T) {
	T.Parallel()

	u, testClient := createUserAndClientForTest(T)
	createdTokens := []*notifications.UserDeviceToken{}
	for i := range exampleQuantity {
		// Use unique device token per creation; registering upserts on (user, token), so
		// duplicate tokens would result in a single row.
		uniqueToken := fmt.Sprintf("a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef%06x", i)
		created := createUserDeviceTokenForTest(T, u.ID, uniqueToken)
		createdTokens = append(createdTokens, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.ListDevices(ctx, &notificationspb.ListDevicesRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.GreaterOrEqual(t, len(retrieved.GetResults()), len(createdTokens))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ListDevices(ctx, &notificationspb.ListDevicesRequest{})
		assert.Error(t, err)
	})
}

func TestUserDeviceTokens_Archive(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		created := createUserDeviceTokenForTest(t, user.ID)

		_, err := testClient.RevokeDevice(ctx, &notificationspb.RevokeDeviceRequest{DeviceId: created.ID})
		require.NoError(t, err)

		// Gone from the caller's list, which is the only read there is.
		listed, err := testClient.ListDevices(ctx, &notificationspb.ListDevicesRequest{})
		require.NoError(t, err)
		for _, device := range listed.GetResults() {
			assert.NotEqual(t, created.ID, device.GetId(), "a revoked device is still listed")
		}

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "archived", ResourceType: "user_device_tokens", RelevantID: created.ID},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		created := createUserDeviceTokenForTest(t, user.ID)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.RevokeDevice(ctx, &notificationspb.RevokeDeviceRequest{DeviceId: created.ID})
		assert.Error(t, err)
	})
}
