package notifications

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v12/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createUserDeviceTokenForTest(t *testing.T, ctx context.Context, userID string, exampleToken *types.UserDeviceToken, dbc *Repository) *types.UserDeviceToken {
	t.Helper()

	if userID == "" {
		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
		userID = user.ID
	}

	if exampleToken == nil {
		exampleToken = fakes.BuildFakeUserDeviceToken()
	}
	exampleToken.BelongsToUser = userID
	dbInput := converters.ConvertUserDeviceTokenToUserDeviceTokenDatabaseCreationInput(exampleToken)

	created, err := dbc.CreateUserDeviceToken(ctx, dbInput)
	require.NoError(t, err)
	exampleToken.CreatedAt = created.CreatedAt
	exampleToken.LastUpdatedAt = created.LastUpdatedAt
	assert.Equal(t, exampleToken.ID, created.ID)
	assert.Equal(t, exampleToken.DeviceToken, created.DeviceToken)
	assert.Equal(t, exampleToken.Platform, created.Platform)
	assert.Equal(t, exampleToken.BelongsToUser, created.BelongsToUser)

	token, err := dbc.GetUserDeviceToken(ctx, created.BelongsToUser, created.ID)
	require.NoError(t, err)
	assert.Equal(t, token.ID, created.ID)

	return created
}

func TestQuerier_Integration_UserDeviceTokens(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	exampleToken := fakes.BuildFakeUserDeviceToken()
	created := createUserDeviceTokenForTest(t, ctx, user.ID, exampleToken, dbc)

	// Note: Create/Upsert does not create audit log entries (unlike user notifications).

	// upsert same token - should update last_updated_at
	created2 := createUserDeviceTokenForTest(t, ctx, user.ID, exampleToken, dbc)
	assert.Equal(t, created.ID, created2.ID)

	// list
	tokens, err := dbc.GetUserDeviceTokens(ctx, user.ID, nil, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, tokens.Data)
	assert.GreaterOrEqual(t, len(tokens.Data), 1)

	// archive
	err = dbc.ArchiveUserDeviceToken(ctx, user.ID, created.ID)
	require.NoError(t, err)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeUserDeviceTokens, RelevantID: created.ID},
	})

	// get after archive should fail
	_, err = dbc.GetUserDeviceToken(ctx, user.ID, created.ID)
	assert.Error(t, err)
}

func TestQuerier_UserDeviceTokenExists(t *testing.T) {
	t.Parallel()

	t.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.UserDeviceTokenExists(ctx, fake.BuildFakeID(), "")
		require.Error(t, err)
		assert.False(t, actual)
	})

	t.Run("with invalid token ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.UserDeviceTokenExists(ctx, "", fake.BuildFakeID())
		require.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetUserDeviceToken(t *testing.T) {
	t.Parallel()

	t.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUserDeviceToken(ctx, "", fake.BuildFakeID())
		require.Error(t, err)
		assert.Nil(t, actual)
	})

	t.Run("with invalid token ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUserDeviceToken(ctx, fake.BuildFakeID(), "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateUserDeviceToken(t *testing.T) {
	t.Parallel()

	t.Run("with nil input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateUserDeviceToken(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateUserDeviceToken(t *testing.T) {
	t.Parallel()

	t.Run("with nil input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateUserDeviceToken(ctx, nil))
	})
}
