package manager

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/fakes"
	notificationsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/mock"

	loggingnoop "github.com/primandproper/platform-go/v11/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v11/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildNotificationsManagerForTest(t *testing.T) *notificationsManager {
	t.Helper()

	ctx := t.Context()
	m, err := NewNotificationsDataManager(
		ctx,
		tracingnoop.NewTracerProvider(),
		loggingnoop.NewLogger(),
		&notificationsmock.RepositoryMock{},
	)
	require.NoError(t, err)

	return m.(*notificationsManager)
}

// attachRepositoryToNotificationsManager wires a configured repository mock
// into the manager under test.
func attachRepositoryToNotificationsManager(manager *notificationsManager, repo *notificationsmock.RepositoryMock) {
	manager.repo = repo
}

func TestNotificationsManager_CreateUserNotification(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		nm := buildNotificationsManagerForTest(t)

		expected := fakes.BuildFakeUserNotification()
		input := converters.ConvertUserNotificationToUserNotificationDatabaseCreationInput(expected)

		repo := &notificationsmock.RepositoryMock{
			CreateUserNotificationFunc: func(_ context.Context, _ *notifications.UserNotificationDatabaseCreationInput) (*notifications.UserNotification, error) {
				return expected, nil
			},
		}
		attachRepositoryToNotificationsManager(nm, repo)

		actual, err := nm.CreateUserNotification(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, repo.CreateUserNotificationCalls(), 1)
	})
}

func TestNotificationsManager_UpdateUserNotification(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		nm := buildNotificationsManagerForTest(t)

		updated := fakes.BuildFakeUserNotification()
		updated.Status = notifications.UserNotificationStatusTypeRead

		repo := &notificationsmock.RepositoryMock{
			UpdateUserNotificationFunc: func(_ context.Context, input *notifications.UserNotification) error {
				assert.Equal(t, updated, input)

				return nil
			},
		}
		attachRepositoryToNotificationsManager(nm, repo)

		err := nm.UpdateUserNotification(ctx, updated)
		require.NoError(t, err)

		assert.Len(t, repo.UpdateUserNotificationCalls(), 1)
	})
}

func TestNotificationsManager_CreateUserDeviceToken(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		nm := buildNotificationsManagerForTest(t)

		expected := fakes.BuildFakeUserDeviceToken()
		input := converters.ConvertUserDeviceTokenToUserDeviceTokenDatabaseCreationInput(expected)

		repo := &notificationsmock.RepositoryMock{
			CreateUserDeviceTokenFunc: func(_ context.Context, _ *notifications.UserDeviceTokenDatabaseCreationInput) (*notifications.UserDeviceToken, error) {
				return expected, nil
			},
		}
		attachRepositoryToNotificationsManager(nm, repo)

		actual, err := nm.CreateUserDeviceToken(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, repo.CreateUserDeviceTokenCalls(), 1)
	})
}

func TestNotificationsManager_ArchiveUserDeviceToken(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		nm := buildNotificationsManagerForTest(t)

		userID := fakes.BuildFakeID()
		tokenID := fakes.BuildFakeID()

		repo := &notificationsmock.RepositoryMock{
			ArchiveUserDeviceTokenFunc: func(_ context.Context, archivedUserID, archivedTokenID string) error {
				assert.Equal(t, userID, archivedUserID)
				assert.Equal(t, tokenID, archivedTokenID)

				return nil
			},
		}
		attachRepositoryToNotificationsManager(nm, repo)

		err := nm.ArchiveUserDeviceToken(ctx, userID, tokenID)
		require.NoError(t, err)

		assert.Len(t, repo.ArchiveUserDeviceTokenCalls(), 1)
	})

	t.Run("with empty user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		nm := buildNotificationsManagerForTest(t)

		err := nm.ArchiveUserDeviceToken(ctx, "", fakes.BuildFakeID())
		assert.Error(t, err)
	})

	t.Run("with empty token ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		nm := buildNotificationsManagerForTest(t)

		err := nm.ArchiveUserDeviceToken(ctx, fakes.BuildFakeID(), "")
		assert.Error(t, err)
	})
}
