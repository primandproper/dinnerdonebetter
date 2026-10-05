package integration

import (
	"crypto/rand"
	"encoding/hex"
	"testing"

	platformnotifications "github.com/primandproper/platform-go/v15/notifications"
	notificationspb "github.com/primandproper/platform-go/v15/notifications/notificationspb"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The device surface's behavior is asserted by platform's notifications conformance suite, run
// against this deployment in conformance_test.go. What remains here is the audit entry this
// application's store records when a device is revoked.

// createUserDeviceForTest registers a handset through the decorated registry.
func createUserDeviceForTest(t *testing.T, forUser string) *platformnotifications.Device {
	t.Helper()

	ctx := t.Context()

	// An APNs token is 32 bytes, rendered as 64 hex characters.
	token := make([]byte, 32)
	_, err := rand.Read(token)
	require.NoError(t, err)

	var created *platformnotifications.Device
	require.NoError(t, databaseClient.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		created, writeErr = notifsRegistry.RegisterDevice(ctx, tx, tenancy.Global(), &platformnotifications.Device{
			ID:        identifiers.New(),
			Principal: forUser,
			Token:     hex.EncodeToString(token),
			Platform:  platformnotifications.PlatformIOS,
		})

		return writeErr
	}))
	assert.NotNil(t, created)

	return created
}

func TestUserDeviceTokens_Archive(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		created := createUserDeviceForTest(t, user.ID)

		_, err := testClient.RevokeDevice(ctx, &notificationspb.RevokeDeviceRequest{DeviceId: created.ID})
		require.NoError(t, err)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "archived", ResourceType: "user_device_tokens", RelevantID: created.ID},
		})
	})
}
