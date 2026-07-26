package manager

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/fakes"
	webhookmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/mock"

	platformerrors "github.com/primandproper/platform-go/v6/errors"
	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v6/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v6/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v6/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v6/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildWebhookManagerForTest builds a manager backed by the given repository mock. A nil repo gets
// an unconfigured mock, which panics if any of its methods are called.
func buildWebhookManagerForTest(t *testing.T, repo *webhookmock.RepositoryMock) *webhookManager {
	t.Helper()

	if repo == nil {
		repo = &webhookmock.RepositoryMock{}
	}

	ctx := t.Context()
	queueCfg := &msgconfig.QueuesConfig{DataChangesTopicName: t.Name()}

	mpp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{
				PublishAsyncFunc: func(_ context.Context, _ any) {},
			}, nil
		},
	}

	m, err := NewWebhookDataManager(ctx, tracingnoop.NewTracerProvider(), loggingnoop.NewLogger(), repo, queueCfg, mpp)
	require.NoError(t, err)

	manager := m.(*webhookManager)
	manager.dataChangesPublisher = &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}

	return manager
}

func TestWebhookDataManager_CreateWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

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

		repo := &webhookmock.RepositoryMock{
			CreateWebhookFunc: func(_ context.Context, in *webhooks.WebhookDatabaseCreationInput) (*webhooks.Webhook, error) {
				assert.Equal(t, input.Name, in.Name)
				assert.Equal(t, input.URL, in.URL)
				assert.Equal(t, userID, in.CreatedByUser)
				assert.Equal(t, accountID, in.BelongsToAccount)
				require.Len(t, in.TriggerConfigs, 1)
				assert.Equal(t, "event-id-1", in.TriggerConfigs[0].TriggerEventID)

				return expectedWebhook, nil
			},
		}
		manager := buildWebhookManagerForTest(t, repo)

		created, err := manager.CreateWebhook(ctx, userID, accountID, input)

		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, expectedWebhook.ID, created.ID)
		assert.Len(t, repo.CreateWebhookCalls(), 1)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		repo := &webhookmock.RepositoryMock{}
		manager := buildWebhookManagerForTest(t, repo)

		created, err := manager.CreateWebhook(ctx, "user-1", "account-1", nil)

		assert.Error(t, err)
		assert.Nil(t, created)
		assert.Empty(t, repo.CreateWebhookCalls())
	})

	t.Run("validation error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		repo := &webhookmock.RepositoryMock{}
		manager := buildWebhookManagerForTest(t, repo)

		input := &webhooks.WebhookCreationRequestInput{
			Name:   "", // invalid
			URL:    "https://example.com",
			Method: http.MethodPost,
			Events: []*webhooks.WebhookTriggerEventCreationRequestInput{{ID: "e1"}},
		}

		created, err := manager.CreateWebhook(ctx, "user-1", "account-1", input)

		assert.Error(t, err)
		assert.Nil(t, created)
		assert.Empty(t, repo.CreateWebhookCalls())
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		input := fakes.BuildFakeWebhookCreationRequestInput()

		repo := &webhookmock.RepositoryMock{
			CreateWebhookFunc: func(_ context.Context, _ *webhooks.WebhookDatabaseCreationInput) (*webhooks.Webhook, error) {
				return nil, errors.New("db error")
			},
		}
		manager := buildWebhookManagerForTest(t, repo)

		created, err := manager.CreateWebhook(ctx, "user-1", "account-1", input)

		assert.Error(t, err)
		assert.Nil(t, created)
		assert.Len(t, repo.CreateWebhookCalls(), 1)
	})
}

func TestWebhookDataManager_GetWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		expected := fakes.BuildFakeWebhook()

		repo := &webhookmock.RepositoryMock{
			GetWebhookFunc: func(_ context.Context, webhookID, accountID string) (*webhooks.Webhook, error) {
				assert.Equal(t, expected.ID, webhookID)
				assert.Equal(t, expected.BelongsToAccount, accountID)

				return expected, nil
			},
		}
		manager := buildWebhookManagerForTest(t, repo)

		result, err := manager.GetWebhook(ctx, expected.ID, expected.BelongsToAccount)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.GetWebhookCalls(), 1)
	})
}

func TestWebhookDataManager_GetWebhooks(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		accountID := "account-1"
		filter := filtering.DefaultQueryFilter()
		expected := fakes.BuildFakeWebhooksList()

		repo := &webhookmock.RepositoryMock{
			GetWebhooksFunc: func(_ context.Context, actualAccountID string, actualFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[webhooks.Webhook], error) {
				assert.Equal(t, accountID, actualAccountID)
				assert.Equal(t, filter, actualFilter)

				return expected, nil
			},
		}
		manager := buildWebhookManagerForTest(t, repo)

		result, err := manager.GetWebhooks(ctx, accountID, filter)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.GetWebhooksCalls(), 1)
	})
}

func TestWebhookDataManager_ArchiveWebhook(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		webhookID := "wh-1"
		accountID := "account-1"

		repo := &webhookmock.RepositoryMock{
			ArchiveWebhookFunc: func(_ context.Context, actualWebhookID, actualAccountID string) error {
				assert.Equal(t, webhookID, actualWebhookID)
				assert.Equal(t, accountID, actualAccountID)

				return nil
			},
		}
		manager := buildWebhookManagerForTest(t, repo)

		err := manager.ArchiveWebhook(ctx, webhookID, accountID)

		require.NoError(t, err)
		assert.Len(t, repo.ArchiveWebhookCalls(), 1)
	})
}

func TestWebhookDataManager_AddWebhookTriggerConfig(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		accountID := "account-1"
		input := &webhooks.WebhookTriggerConfigCreationRequestInput{
			BelongsToWebhook: "webhook-1",
			TriggerEventID:   "event-1",
		}
		expectedConfig := fakes.BuildFakeWebhookTriggerConfig()

		repo := &webhookmock.RepositoryMock{
			AddWebhookTriggerConfigFunc: func(_ context.Context, actualAccountID string, in *webhooks.WebhookTriggerConfigDatabaseCreationInput) (*webhooks.WebhookTriggerConfig, error) {
				assert.Equal(t, accountID, actualAccountID)
				assert.Equal(t, input.BelongsToWebhook, in.BelongsToWebhook)
				assert.Equal(t, input.TriggerEventID, in.TriggerEventID)

				return expectedConfig, nil
			},
		}
		manager := buildWebhookManagerForTest(t, repo)

		result, err := manager.AddWebhookTriggerConfig(ctx, accountID, input)

		require.NoError(t, err)
		assert.Equal(t, expectedConfig, result)
		assert.Len(t, repo.AddWebhookTriggerConfigCalls(), 1)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		repo := &webhookmock.RepositoryMock{}
		manager := buildWebhookManagerForTest(t, repo)

		result, err := manager.AddWebhookTriggerConfig(ctx, "account-1", nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Empty(t, repo.AddWebhookTriggerConfigCalls())
	})
}

func TestWebhookDataManager_ArchiveWebhookTriggerConfig(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		webhookID := "wh-1"
		configID := "config-1"

		repo := &webhookmock.RepositoryMock{
			ArchiveWebhookTriggerConfigFunc: func(_ context.Context, actualWebhookID, actualConfigID string) error {
				assert.Equal(t, webhookID, actualWebhookID)
				assert.Equal(t, configID, actualConfigID)

				return nil
			},
		}
		manager := buildWebhookManagerForTest(t, repo)

		err := manager.ArchiveWebhookTriggerConfig(ctx, webhookID, configID)

		require.NoError(t, err)
		assert.Len(t, repo.ArchiveWebhookTriggerConfigCalls(), 1)
	})
}

func TestWebhookDataManager_CreateWebhookTriggerEvent(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		input := &webhooks.WebhookTriggerEventCreationRequestInput{
			Name:        "webhook_created",
			Description: "Fired when a webhook is created",
		}
		expected := fakes.BuildFakeWebhookTriggerEvent()

		repo := &webhookmock.RepositoryMock{
			CreateWebhookTriggerEventFunc: func(_ context.Context, in *webhooks.WebhookTriggerEventDatabaseCreationInput) (*webhooks.WebhookTriggerEvent, error) {
				assert.Equal(t, input.Name, in.Name)
				assert.Equal(t, input.Description, in.Description)

				return expected, nil
			},
		}
		manager := buildWebhookManagerForTest(t, repo)

		result, err := manager.CreateWebhookTriggerEvent(ctx, input)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.CreateWebhookTriggerEventCalls(), 1)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		repo := &webhookmock.RepositoryMock{}
		manager := buildWebhookManagerForTest(t, repo)

		result, err := manager.CreateWebhookTriggerEvent(ctx, nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Empty(t, repo.CreateWebhookTriggerEventCalls())
	})
}

func TestWebhookDataManager_UpdateWebhookTriggerEvent(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		triggerEventID := fakes.BuildFakeID()
		input := &webhooks.WebhookTriggerEventUpdateRequestInput{}

		repo := &webhookmock.RepositoryMock{
			UpdateWebhookTriggerEventFunc: func(_ context.Context, id string, in *webhooks.WebhookTriggerEventUpdateRequestInput) error {
				assert.Equal(t, triggerEventID, id)
				assert.Equal(t, input, in)

				return nil
			},
		}
		manager := buildWebhookManagerForTest(t, repo)

		err := manager.UpdateWebhookTriggerEvent(ctx, triggerEventID, input)

		require.NoError(t, err)
		assert.Len(t, repo.UpdateWebhookTriggerEventCalls(), 1)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		repo := &webhookmock.RepositoryMock{}
		manager := buildWebhookManagerForTest(t, repo)

		err := manager.UpdateWebhookTriggerEvent(ctx, fakes.BuildFakeID(), nil)

		require.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
		assert.Empty(t, repo.UpdateWebhookTriggerEventCalls())
	})
}

func TestWebhookDataManager_WebhookExists(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		repo := &webhookmock.RepositoryMock{
			WebhookExistsFunc: func(_ context.Context, webhookID, accountID string) (bool, error) {
				assert.Equal(t, "wh-1", webhookID)
				assert.Equal(t, "account-1", accountID)

				return true, nil
			},
		}
		manager := buildWebhookManagerForTest(t, repo)

		exists, err := manager.WebhookExists(ctx, "wh-1", "account-1")

		require.NoError(t, err)
		assert.True(t, exists)
		assert.Len(t, repo.WebhookExistsCalls(), 1)
	})
}
