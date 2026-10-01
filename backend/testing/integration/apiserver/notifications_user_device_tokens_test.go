package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/fakes"

	notificationspb "github.com/primandproper/platform-go/v14/notifications/notificationspb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The device surface's behavior is asserted by platform's notifications conformance suite, run
// against this deployment in conformance_test.go. What remains here is the audit entry this
// application's store records when a device is revoked.

func createUserDeviceTokenForTest(t *testing.T, forUser string) *notifications.UserDeviceToken {
	t.Helper()

	ctx := t.Context()

	creationInput := fakes.BuildFakeUserDeviceToken()
	input := converters.ConvertUserDeviceTokenToUserDeviceTokenDatabaseCreationInput(creationInput)
	input.BelongsToUser = forUser

	created, err := notifsRepo.CreateUserDeviceToken(ctx, input)
	require.NoError(t, err)
	assert.NotNil(t, created)

	return created
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

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "archived", ResourceType: "user_device_tokens", RelevantID: created.ID},
		})
	})
}
