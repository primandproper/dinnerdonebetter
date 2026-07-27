package notifications

import (
	"context"
	"fmt"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createUserNotificationForTest(t *testing.T, ctx context.Context, userID string, exampleUserNotification *types.UserNotification, dbc *Repository) *types.UserNotification {
	t.Helper()

	if userID == "" {
		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
		userID = user.ID
	}

	// create
	if exampleUserNotification == nil {
		exampleUserNotification = fakes.BuildFakeUserNotification()
	}
	exampleUserNotification.BelongsToUser = userID
	dbInput := converters.ConvertUserNotificationToUserNotificationDatabaseCreationInput(exampleUserNotification)

	created, err := dbc.CreateUserNotification(ctx, dbInput)
	require.NoError(t, err)
	exampleUserNotification.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleUserNotification, created)

	userNotification, err := dbc.GetUserNotification(ctx, created.BelongsToUser, created.ID)
	exampleUserNotification.CreatedAt = userNotification.CreatedAt

	assert.NoError(t, err)
	assert.Equal(t, userNotification, exampleUserNotification)

	return created
}

func TestQuerier_Integration_UserNotifications(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	exampleUserNotification := fakes.BuildFakeUserNotification()
	createdUserNotifications := []*types.UserNotification{}

	// create
	createdUserNotifications = append(createdUserNotifications, createUserNotificationForTest(t, ctx, user.ID, exampleUserNotification, dbc))

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeUserNotifications, RelevantID: createdUserNotifications[0].ID},
	})

	// update
	updatedUserNotification := fakes.BuildFakeUserNotification()
	updatedUserNotification.ID = createdUserNotifications[0].ID
	updatedUserNotification.BelongsToUser = user.ID
	assert.NoError(t, dbc.UpdateUserNotification(ctx, updatedUserNotification))
	createdUserNotifications[0] = updatedUserNotification

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeUserNotifications, RelevantID: createdUserNotifications[0].ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeUserNotifications, RelevantID: createdUserNotifications[0].ID},
	})

	// create more
	for i := range exampleQuantity {
		input := fakes.BuildFakeUserNotification()
		input.Content = fmt.Sprintf("%s %d", updatedUserNotification.Content, i)
		createdUserNotifications = append(createdUserNotifications, createUserNotificationForTest(t, ctx, user.ID, input, dbc))
	}

	// fetch as list
	userNotifications, err := dbc.GetUserNotifications(ctx, user.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, userNotifications.Data)
	assert.Equal(t, len(createdUserNotifications), len(userNotifications.Data))

	// delete
	for _, userNotification := range createdUserNotifications {
		var exists bool
		exists, err = dbc.UserNotificationExists(ctx, user.ID, userNotification.ID)
		assert.NoError(t, err)
		assert.True(t, exists)
	}
}

func TestQuerier_UserNotificationExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.UserNotificationExists(ctx, fakes.BuildFakeID(), "")
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid user notification MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.UserNotificationExists(ctx, "", fakes.BuildFakeID())
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetUserNotification(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUserNotification(ctx, "", fakes.BuildFakeID())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid user notification MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUserNotification(ctx, fakes.BuildFakeID(), "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateUserNotification(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateUserNotification(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateUserNotification(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateUserNotification(ctx, nil))
	})
}
