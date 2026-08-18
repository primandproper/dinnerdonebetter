package webhooks

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createWebhookForTest(t *testing.T, ctx context.Context, exampleWebhook *types.Webhook, dbc *repository) *types.Webhook {
	t.Helper()

	// create
	if exampleWebhook == nil {
		exampleWebhook = fakes.BuildFakeWebhook()
	}
	dbInput := converters.ConvertWebhookToWebhookDatabaseCreationInput(exampleWebhook)

	response, err := dbc.CreateWebhook(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, response)

	// The signing secret is returned here and nowhere else, so this is the only place it can
	// be asserted on at all.
	assert.NotEmpty(t, response.Secret)

	created := response.Webhook
	require.NotNil(t, created)

	exampleWebhook.CreatedAt = created.CreatedAt
	for i := range created.TriggerConfigs {
		exampleWebhook.TriggerConfigs[i].CreatedAt = created.TriggerConfigs[i].CreatedAt
	}
	assert.Equal(t, exampleWebhook, created)

	webhook, err := dbc.GetWebhook(ctx, created.ID, created.BelongsToAccount)
	exampleWebhook.CreatedAt = webhook.CreatedAt
	for i := range created.TriggerConfigs {
		exampleWebhook.TriggerConfigs[i].CreatedAt = webhook.TriggerConfigs[i].CreatedAt
	}

	require.NoError(t, err)
	assert.Equal(t, webhook, exampleWebhook)

	return created
}

func TestQuerier_Integration_Webhooks(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	exampleWebhook := fakes.BuildFakeWebhook()
	exampleWebhook.BelongsToAccount = account.ID
	exampleWebhook.CreatedByUser = user.ID
	createdWebhooks := []*types.Webhook{}

	// create
	createdWebhooks = append(createdWebhooks, createWebhookForTest(t, ctx, exampleWebhook, dbc))

	// create more
	for i := range exampleQuantity {
		input := fakes.BuildFakeWebhook()
		input.Name = fmt.Sprintf("%s %d", exampleWebhook.Name, i)
		input.BelongsToAccount = account.ID
		input.CreatedByUser = user.ID
		createdWebhooks = append(createdWebhooks, createWebhookForTest(t, ctx, input, dbc))
	}

	// fetch as list
	webhooks, err := dbc.GetWebhooks(ctx, account.ID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, webhooks.Data)
	assert.Len(t, webhooks.Data, len(createdWebhooks))

	// Subscribe the webhook to a second event type. It has to differ from the one the fake
	// already carries: (trigger_event, belongs_to_webhook, archived_at) is unique.
	secondEventType := fakes.BuildFakeWebhookEventType()
	for secondEventType == createdWebhooks[0].TriggerConfigs[0].EventType {
		secondEventType = fakes.BuildFakeWebhookEventType()
	}

	createdConfig, err := dbc.AddWebhookTriggerConfig(ctx, account.ID, &types.WebhookTriggerConfigDatabaseCreationInput{
		ID:               identifiers.New(),
		BelongsToWebhook: createdWebhooks[0].ID,
		EventType:        secondEventType,
	})
	require.NoError(t, err)
	assert.NotNil(t, createdConfig)

	createdWebhooks[0].TriggerConfigs = append(createdWebhooks[0].TriggerConfigs, createdConfig)

	// Assert audit log entries were written for creates (pre-cleanup)
	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeWebhooks, RelevantID: createdWebhooks[0].ID},
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeWebhookTriggerConfigs, RelevantID: createdWebhooks[0].TriggerConfigs[0].ID},
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeWebhookTriggerConfigs, RelevantID: createdConfig.ID},
	})

	// delete: archive trigger configs then archive webhook; archive catalog event if needed
	for _, webhook := range createdWebhooks {
		for _, cfg := range webhook.TriggerConfigs {
			require.NoError(t, dbc.ArchiveWebhookTriggerConfig(ctx, webhook.ID, account.ID, cfg.ID))
		}

		require.NoError(t, dbc.ArchiveWebhook(ctx, webhook.ID, account.ID))
	}

	// Assert audit log entries were written for webhook archives (ArchiveWebhookTriggerConfig
	// does not set BelongsToAccount, so those entries are not returned by GetAuditLogEntriesForAccount)
	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeWebhooks, RelevantID: createdWebhooks[0].ID},
	})

	for _, webhook := range createdWebhooks {
		var exists bool
		exists, err = dbc.WebhookExists(ctx, webhook.ID, account.ID)
		require.NoError(t, err)
		assert.False(t, exists)

		var y *types.Webhook
		y, err = dbc.GetWebhook(ctx, webhook.ID, account.ID)
		assert.Nil(t, y)
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fake.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetWebhook(ctx, "", exampleAccountID)
		require.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetWebhook(ctx, exampleWebhook.ID, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		filter := filtering.DefaultQueryFilter()
		c := buildInertClientForTest(t)

		actual, err := c.GetWebhooks(ctx, "", filter)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateWebhook(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateWebhook(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fake.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveWebhook(ctx, "", exampleAccountID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhookID := fake.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveWebhook(ctx, exampleWebhookID, ""))
	})
}

func TestQuerier_ArchiveWebhookTriggerConfig(T *testing.T) {
	T.Parallel()

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		exampleConfigID := fake.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveWebhookTriggerConfig(ctx, "", fake.BuildFakeID(), exampleConfigID))
	})

	T.Run("with invalid webhook trigger config ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhookID := fake.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveWebhookTriggerConfig(ctx, exampleWebhookID, fake.BuildFakeID(), ""))
	})
}

func TestQuerier_Integration_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.Webhook]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "webhook",
		CreateItem: func(ctx context.Context, i int) *types.Webhook {
			webhook := fakes.BuildFakeWebhook()
			webhook.Name = fmt.Sprintf("Webhook %02d", i) // Use zero-padded numbers for consistent sorting
			webhook.BelongsToAccount = account.ID
			webhook.CreatedByUser = user.ID
			return createWebhookForTest(t, ctx, webhook, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Webhook], error) {
			return dbc.GetWebhooks(ctx, account.ID, filter)
		},
		GetID: func(webhook *types.Webhook) string {
			return webhook.ID
		},
		CleanupItem: func(ctx context.Context, webhook *types.Webhook) error {
			for _, cfg := range webhook.TriggerConfigs {
				if err := dbc.ArchiveWebhookTriggerConfig(ctx, webhook.ID, account.ID, cfg.ID); err != nil {
					return err
				}
			}
			return dbc.ArchiveWebhook(ctx, webhook.ID, account.ID)
		},
	})
}
