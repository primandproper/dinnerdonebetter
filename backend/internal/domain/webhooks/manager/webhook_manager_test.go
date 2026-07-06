package manager

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/fakes"
	webhookkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/keys"
	webhookmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v3/database/filtering"
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

func buildWebhookManagerForTest(t *testing.T) (*webhookManager, *webhookmock.Repository) {
	t.Helper()

	ctx := t.Context()
	repo := &webhookmock.Repository{}
	queueCfg := &msgconfig.QueuesConfig{DataChangesTopicName: t.Name()}

	mpp := &mockpublishers.PublisherProviderMock{
		ProvidePublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{
				PublishAsyncFunc: func(_ context.Context, _ any) {},
			}, nil
		},
	}

	m, err := NewWebhookDataManager(ctx, tracingnoop.NewTracerProvider(), loggingnoop.NewLogger(), repo, queueCfg, mpp)
	require.NoError(t, err)

	return m.(*webhookManager), repo
}

func setupExpectationsForWebhookManager(
	manager *webhookManager,
	repoSetupFunc func(repo *webhookmock.Repository),
	eventTypeMaps ...map[string][]string,
) []any {
	repo := &webhookmock.Repository{}
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

func TestWebhookDataManager_CreateWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWebhookManagerForTest(t)

		userID := "user-1"
		accountID := "account-1"
		input := &webhooks.WebhookCreationRequestInput{
			Name:        "test webhook",
			ContentType: "application/json",
			URL:         "https://example.com/hook",
			Method:      http.MethodPost,
			Events:      []*webhooks.WebhookTriggerEventCreationRequestInput{{ID: "event-id-1"}},
		}

		expectedWebhook := fakes.BuildFakeWebhook()

		expectations := setupExpectationsForWebhookManager(
			manager,
			func(repo *webhookmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateWebhook), testutils.ContextMatcher, mock.MatchedBy(func(in *webhooks.WebhookDatabaseCreationInput) bool {
					return in.Name == input.Name && in.URL == input.URL && in.CreatedByUser == userID && in.BelongsToAccount == accountID && len(in.TriggerConfigs) == 1 && in.TriggerConfigs[0].TriggerEventID == "event-id-1"
				})).Return(expectedWebhook, nil)
			},
			map[string][]string{
				webhooks.WebhookCreatedServiceEventType: {webhookkeys.WebhookIDKey},
			},
		)

		created, err := manager.CreateWebhook(ctx, userID, accountID, input)

		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, expectedWebhook.ID, created.ID)
		mock.AssertExpectationsForObjects(t, expectations...)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildWebhookManagerForTest(t)

		created, err := manager.CreateWebhook(ctx, "user-1", "account-1", nil)

		assert.Error(t, err)
		assert.Nil(t, created)
		repo.AssertNotCalled(t, reflection.GetMethodName(repo.CreateWebhook))
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildWebhookManagerForTest(t)

		input := &webhooks.WebhookCreationRequestInput{
			Name:   "", // invalid
			URL:    "https://example.com",
			Method: http.MethodPost,
			Events: []*webhooks.WebhookTriggerEventCreationRequestInput{{ID: "e1"}},
		}

		created, err := manager.CreateWebhook(ctx, "user-1", "account-1", input)

		assert.Error(t, err)
		assert.Nil(t, created)
		repo.AssertNotCalled(t, reflection.GetMethodName(repo.CreateWebhook))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWebhookManagerForTest(t)

		input := fakes.BuildFakeWebhookCreationRequestInput()

		expectations := setupExpectationsForWebhookManager(
			manager,
			func(repo *webhookmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateWebhook), testutils.ContextMatcher, mock.Anything).Return(nil, errors.New("db error"))
			},
		)

		created, err := manager.CreateWebhook(ctx, "user-1", "account-1", input)

		assert.Error(t, err)
		assert.Nil(t, created)
		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestWebhookDataManager_GetWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildWebhookManagerForTest(t)

		expected := fakes.BuildFakeWebhook()
		repo.On(reflection.GetMethodName(repo.GetWebhook), testutils.ContextMatcher, expected.ID, expected.BelongsToAccount).Return(expected, nil)

		result, err := manager.GetWebhook(ctx, expected.ID, expected.BelongsToAccount)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, repo)
	})
}

func TestWebhookDataManager_GetWebhooks(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildWebhookManagerForTest(t)

		accountID := "account-1"
		filter := filtering.DefaultQueryFilter()
		expected := fakes.BuildFakeWebhooksList()
		repo.On(reflection.GetMethodName(repo.GetWebhooks), testutils.ContextMatcher, accountID, filter).Return(expected, nil)

		result, err := manager.GetWebhooks(ctx, accountID, filter)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, repo)
	})
}

func TestWebhookDataManager_ArchiveWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWebhookManagerForTest(t)

		webhookID := "wh-1"
		accountID := "account-1"

		expectations := setupExpectationsForWebhookManager(
			manager,
			func(repo *webhookmock.Repository) {
				repo.On(reflection.GetMethodName(repo.ArchiveWebhook), testutils.ContextMatcher, webhookID, accountID).Return(nil)
			},
			map[string][]string{
				webhooks.WebhookArchivedServiceEventType: {webhookkeys.WebhookIDKey},
			},
		)

		err := manager.ArchiveWebhook(ctx, webhookID, accountID)

		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestWebhookDataManager_AddWebhookTriggerConfig(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWebhookManagerForTest(t)

		accountID := "account-1"
		input := &webhooks.WebhookTriggerConfigCreationRequestInput{
			BelongsToWebhook: "webhook-1",
			TriggerEventID:   "event-1",
		}
		expectedConfig := fakes.BuildFakeWebhookTriggerConfig()

		expectations := setupExpectationsForWebhookManager(
			manager,
			func(repo *webhookmock.Repository) {
				repo.On(reflection.GetMethodName(repo.AddWebhookTriggerConfig), testutils.ContextMatcher, accountID, mock.MatchedBy(func(in *webhooks.WebhookTriggerConfigDatabaseCreationInput) bool {
					return in.BelongsToWebhook == input.BelongsToWebhook && in.TriggerEventID == input.TriggerEventID
				})).Return(expectedConfig, nil)
			},
			map[string][]string{
				webhooks.WebhookTriggerConfigCreatedServiceEventType: {webhookkeys.WebhookIDKey, webhookkeys.WebhookTriggerConfigIDKey},
			},
		)

		result, err := manager.AddWebhookTriggerConfig(ctx, accountID, input)

		require.NoError(t, err)
		assert.Equal(t, expectedConfig, result)
		mock.AssertExpectationsForObjects(t, expectations...)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildWebhookManagerForTest(t)

		result, err := manager.AddWebhookTriggerConfig(ctx, "account-1", nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		repo.AssertNotCalled(t, reflection.GetMethodName(repo.AddWebhookTriggerConfig))
	})
}

func TestWebhookDataManager_ArchiveWebhookTriggerConfig(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, _ := buildWebhookManagerForTest(t)

		webhookID := "wh-1"
		configID := "config-1"

		expectations := setupExpectationsForWebhookManager(
			manager,
			func(repo *webhookmock.Repository) {
				repo.On(reflection.GetMethodName(repo.ArchiveWebhookTriggerConfig), testutils.ContextMatcher, webhookID, configID).Return(nil)
			},
			map[string][]string{
				webhooks.WebhookTriggerConfigArchivedServiceEventType: {webhookkeys.WebhookIDKey, webhookkeys.WebhookTriggerConfigIDKey},
			},
		)

		err := manager.ArchiveWebhookTriggerConfig(ctx, webhookID, configID)

		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestWebhookDataManager_CreateWebhookTriggerEvent(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildWebhookManagerForTest(t)

		input := &webhooks.WebhookTriggerEventCreationRequestInput{
			Name:        "webhook_created",
			Description: "Fired when a webhook is created",
		}
		expected := fakes.BuildFakeWebhookTriggerEvent()
		repo.On(reflection.GetMethodName(repo.CreateWebhookTriggerEvent), testutils.ContextMatcher, mock.MatchedBy(func(in *webhooks.WebhookTriggerEventDatabaseCreationInput) bool {
			return in.Name == input.Name && in.Description == input.Description
		})).Return(expected, nil)

		result, err := manager.CreateWebhookTriggerEvent(ctx, input)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, repo)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildWebhookManagerForTest(t)

		result, err := manager.CreateWebhookTriggerEvent(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		repo.AssertNotCalled(t, reflection.GetMethodName(repo.CreateWebhookTriggerEvent))
	})
}

func TestWebhookDataManager_WebhookExists(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildWebhookManagerForTest(t)

		repo.On(reflection.GetMethodName(repo.WebhookExists), testutils.ContextMatcher, "wh-1", "account-1").Return(true, nil)

		exists, err := manager.WebhookExists(ctx, "wh-1", "account-1")

		require.NoError(t, err)
		assert.True(t, exists)
		mock.AssertExpectationsForObjects(t, repo)
	})
}
