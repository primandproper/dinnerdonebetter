package manager

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/fakes"
	notificationkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/keys"
	notificationsmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v3/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v3/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v3/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v3/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v3/observability/tracing/noop"
	"github.com/primandproper/platform-go/v3/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildNotificationsManagerForTest(t *testing.T) *notificationsManager {
	t.Helper()

	ctx := t.Context()
	queueCfg := &msgconfig.QueuesConfig{
		DataChangesTopicName: t.Name(),
	}

	mpp := &mockpublishers.PublisherProviderMock{
		ProvidePublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{
				PublishAsyncFunc: func(_ context.Context, _ any) {},
			}, nil
		},
	}

	m, err := NewNotificationsDataManager(
		ctx,
		tracingnoop.NewTracerProvider(),
		loggingnoop.NewLogger(),
		&notificationsmock.Repository{},
		queueCfg,
		mpp,
	)
	require.NoError(t, err)

	return m.(*notificationsManager)
}

func setupExpectationsForNotificationsManager(
	manager *notificationsManager,
	repoSetupFunc func(repo *notificationsmock.Repository),
	eventTypeMaps ...map[string][]string,
) []any {
	repo := &notificationsmock.Repository{}
	if repoSetupFunc != nil {
		repoSetupFunc(repo)
	}
	manager.repo = repo

	mp := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}
	manager.dataChangesPublisher = mp

	return []any{repo}
}

func TestNotificationsManager_CreateUserNotification(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		nm := buildNotificationsManagerForTest(t)

		expected := fakes.BuildFakeUserNotification()
		input := converters.ConvertUserNotificationToUserNotificationDatabaseCreationInput(expected)

		expectations := setupExpectationsForNotificationsManager(
			nm,
			func(repo *notificationsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateUserNotification), testutils.ContextMatcher, testutils.MatchType[*notifications.UserNotificationDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				notifications.UserNotificationCreatedServiceEventType: {notificationkeys.UserNotificationIDKey},
			},
		)

		actual, err := nm.CreateUserNotification(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForNotificationsManager(
			nm,
			func(repo *notificationsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.UpdateUserNotification), testutils.ContextMatcher, updated).Return(nil)
			},
			map[string][]string{
				notifications.UserNotificationUpdatedServiceEventType: {notificationkeys.UserNotificationIDKey},
			},
		)

		err := nm.UpdateUserNotification(ctx, updated)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForNotificationsManager(
			nm,
			func(repo *notificationsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateUserDeviceToken), testutils.ContextMatcher, testutils.MatchType[*notifications.UserDeviceTokenDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				notifications.UserDeviceTokenCreatedServiceEventType: {notificationkeys.UserDeviceTokenIDKey},
			},
		)

		actual, err := nm.CreateUserDeviceToken(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForNotificationsManager(
			nm,
			func(repo *notificationsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.ArchiveUserDeviceToken), testutils.ContextMatcher, userID, tokenID).Return(nil)
			},
			map[string][]string{
				notifications.UserDeviceTokenArchivedServiceEventType: {notificationkeys.UserDeviceTokenIDKey},
			},
		)

		err := nm.ArchiveUserDeviceToken(ctx, userID, tokenID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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
