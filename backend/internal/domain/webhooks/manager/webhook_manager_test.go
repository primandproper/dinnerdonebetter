package manager

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/fakes"
	webhookmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/mock"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"
	loggingnoop "github.com/primandproper/platform-go/v12/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v12/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildWebhookManagerForTest builds a manager backed by the given repository mock. A nil repo gets
// an unconfigured mock, which panics if any of its methods are called.
// exampleEventType is a real catalog entry: validation rejects anything else, so a made-up
// string would only ever exercise the rejection path.
var exampleEventType = fakes.BuildFakeWebhookEventType()

const exampleSecret = "6465616462656566"

func buildWebhookManagerForTest(t *testing.T, repo *webhookmock.RepositoryMock) *webhookManager {
	t.Helper()

	if repo == nil {
		repo = &webhookmock.RepositoryMock{}
	}

	ctx := t.Context()

	m, err := NewWebhookDataManager(ctx, tracingnoop.NewTracerProvider(), loggingnoop.NewLogger(), repo)
	require.NoError(t, err)

	manager := m.(*webhookManager)

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
			Events:      []string{exampleEventType},
		}

		expectedWebhook := fakes.BuildFakeWebhook()

		repo := &webhookmock.RepositoryMock{
			CreateWebhookFunc: func(_ context.Context, in *webhooks.WebhookDatabaseCreationInput) (*webhooks.WebhookCreationResponse, error) {
				assert.Equal(t, input.Name, in.Name)
				assert.Equal(t, input.URL, in.URL)
				assert.Equal(t, userID, in.CreatedByUser)
				assert.Equal(t, accountID, in.BelongsToAccount)
				require.Len(t, in.TriggerConfigs, 1)
				assert.Equal(t, exampleEventType, in.TriggerConfigs[0].EventType)

				return &webhooks.WebhookCreationResponse{Webhook: expectedWebhook, Secret: exampleSecret}, nil
			},
		}
		manager := buildWebhookManagerForTest(t, repo)

		created, err := manager.CreateWebhook(ctx, userID, accountID, input)

		require.NoError(t, err)
		assert.NotNil(t, created)
		assert.Equal(t, expectedWebhook.ID, created.Webhook.ID)
		// The secret reaches the caller from here and from nowhere else.
		assert.Equal(t, exampleSecret, created.Secret)
		assert.Len(t, repo.CreateWebhookCalls(), 1)
	})

	t.Run("nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		repo := &webhookmock.RepositoryMock{}
		manager := buildWebhookManagerForTest(t, repo)

		created, err := manager.CreateWebhook(ctx, "user-1", "account-1", nil)

		require.Error(t, err)
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
			Events: []string{exampleEventType},
		}

		created, err := manager.CreateWebhook(ctx, "user-1", "account-1", input)

		require.Error(t, err)
		assert.Nil(t, created)
		assert.Empty(t, repo.CreateWebhookCalls())
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		input := fakes.BuildFakeWebhookCreationRequestInput()

		repo := &webhookmock.RepositoryMock{
			CreateWebhookFunc: func(_ context.Context, _ *webhooks.WebhookDatabaseCreationInput) (*webhooks.WebhookCreationResponse, error) {
				return nil, errors.New("db error")
			},
		}
		manager := buildWebhookManagerForTest(t, repo)

		created, err := manager.CreateWebhook(ctx, "user-1", "account-1", input)

		require.Error(t, err)
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
			EventType:        exampleEventType,
		}
		expectedConfig := fakes.BuildFakeWebhookTriggerConfig()

		repo := &webhookmock.RepositoryMock{
			AddWebhookTriggerConfigFunc: func(_ context.Context, actualAccountID string, in *webhooks.WebhookTriggerConfigDatabaseCreationInput) (*webhooks.WebhookTriggerConfig, error) {
				assert.Equal(t, accountID, actualAccountID)
				assert.Equal(t, input.BelongsToWebhook, in.BelongsToWebhook)
				assert.Equal(t, input.EventType, in.EventType)

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

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Empty(t, repo.AddWebhookTriggerConfigCalls())
	})
}

func TestWebhookDataManager_ArchiveWebhookTriggerConfig(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		webhookID := fake.BuildFakeID()
		accountID := fake.BuildFakeID()
		configID := fake.BuildFakeID()

		repo := &webhookmock.RepositoryMock{
			ArchiveWebhookTriggerConfigFunc: func(_ context.Context, actualWebhookID, actualAccountID, actualConfigID string) error {
				assert.Equal(t, webhookID, actualWebhookID)
				// The account travels down to the query, which scopes the archive by it.
				// Without that, two IDs would be enough to silence another account's
				// subscription.
				assert.Equal(t, accountID, actualAccountID)
				assert.Equal(t, configID, actualConfigID)

				return nil
			},
		}
		manager := buildWebhookManagerForTest(t, repo)

		err := manager.ArchiveWebhookTriggerConfig(ctx, webhookID, accountID, configID)

		require.NoError(t, err)
		assert.Len(t, repo.ArchiveWebhookTriggerConfigCalls(), 1)
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
