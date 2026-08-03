package integration

import (
	"fmt"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/fakes"
	notificationssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/notifications"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/notifications/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

		user, testClient := createUserAndClientForTest(t)
		exampleToken := fakes.BuildFakeUserDeviceToken()

		response, err := testClient.RegisterDeviceToken(ctx, &notificationssvc.RegisterDeviceTokenRequest{
			Input: &notificationssvc.UserDeviceTokenCreationRequestInput{
				DeviceToken: exampleToken.DeviceToken,
				Platform:    exampleToken.Platform,
			},
		})
		assert.NoError(t, err)
		require.NotNil(t, response)
		require.NotNil(t, response.Created)
		assert.NotEmpty(t, response.Created.Id)
		assert.Equal(t, exampleToken.DeviceToken, response.Created.DeviceToken)
		assert.Equal(t, exampleToken.Platform, response.Created.Platform)
		assert.Equal(t, user.ID, response.Created.BelongsToUser)

		retrieved, err := testClient.GetUserDeviceToken(ctx, &notificationssvc.GetUserDeviceTokenRequest{
			UserDeviceTokenId: response.Created.Id,
		})
		assert.NoError(t, err)
		require.NotNil(t, retrieved)
		converted := grpcconverters.ConvertGRPCUserDeviceTokenToUserDeviceToken(retrieved.Result)
		assert.Equal(t, response.Created.Id, converted.ID)
		assert.Equal(t, exampleToken.DeviceToken, converted.DeviceToken)
		assert.Equal(t, exampleToken.Platform, converted.Platform)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)
		created := createUserDeviceTokenForTest(t, user.ID)

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetUserDeviceToken(ctx, &notificationssvc.GetUserDeviceTokenRequest{
			UserDeviceTokenId: created.ID,
		})
		assert.Error(t, err)
	})

	T.Run("invalid token ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetUserDeviceToken(ctx, &notificationssvc.GetUserDeviceTokenRequest{
			UserDeviceTokenId: nonexistentID,
		})
		assert.Error(t, err)
	})
}

func TestUserDeviceTokens_Listing(T *testing.T) {
	T.Parallel()

	u, testClient := createUserAndClientForTest(T)
	createdTokens := []*notifications.UserDeviceToken{}
	for i := range exampleQuantity {
		// Use unique device token per creation; CreateUserDeviceToken upserts on (user, token), so
		// duplicate tokens would result in a single DB row.
		uniqueToken := fmt.Sprintf("a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef%06x", i)
		created := createUserDeviceTokenForTest(T, u.ID, uniqueToken)
		createdTokens = append(createdTokens, created)
	}

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		retrieved, err := testClient.GetUserDeviceTokens(ctx, &notificationssvc.GetUserDeviceTokensRequest{})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		assert.GreaterOrEqual(t, len(retrieved.Results), len(createdTokens))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetUserDeviceTokens(ctx, &notificationssvc.GetUserDeviceTokensRequest{})
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

		_, err := testClient.ArchiveUserDeviceToken(ctx, &notificationssvc.ArchiveUserDeviceTokenRequest{
			UserDeviceTokenId: created.ID,
		})
		assert.NoError(t, err)

		_, err = testClient.GetUserDeviceToken(ctx, &notificationssvc.GetUserDeviceTokenRequest{
			UserDeviceTokenId: created.ID,
		})
		assert.Error(t, err)

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

		_, err := c.ArchiveUserDeviceToken(ctx, &notificationssvc.ArchiveUserDeviceTokenRequest{
			UserDeviceTokenId: created.ID,
		})
		assert.Error(t, err)
	})
}
