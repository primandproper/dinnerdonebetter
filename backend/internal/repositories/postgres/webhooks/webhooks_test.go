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

	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"

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

	created, err := dbc.CreateWebhook(ctx, dbInput)
	assert.NoError(t, err)
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

	assert.NoError(t, err)
	assert.Equal(t, webhook, exampleWebhook)

	return created
}

func TestQuerier_Integration_Webhooks(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	// Create catalog trigger events so webhook trigger configs can reference them
	catalogEvent, err := dbc.CreateWebhookTriggerEvent(ctx, &types.WebhookTriggerEventDatabaseCreationInput{
		ID:          identifiers.New(),
		Name:        "webhook_created",
		Description: "test",
	})
	require.NoError(t, err)
	require.NotNil(t, catalogEvent)

	exampleWebhook := fakes.BuildFakeWebhook()
	exampleWebhook.BelongsToAccount = account.ID
	exampleWebhook.CreatedByUser = user.ID
	exampleWebhook.TriggerConfigs[0].TriggerEventID = catalogEvent.ID
	createdWebhooks := []*types.Webhook{}

	// create
	createdWebhooks = append(createdWebhooks, createWebhookForTest(t, ctx, exampleWebhook, dbc))

	// create more
	for i := range exampleQuantity {
		input := fakes.BuildFakeWebhook()
		input.Name = fmt.Sprintf("%s %d", exampleWebhook.Name, i)
		input.BelongsToAccount = account.ID
		input.CreatedByUser = user.ID
		input.TriggerConfigs[0].TriggerEventID = catalogEvent.ID
		createdWebhooks = append(createdWebhooks, createWebhookForTest(t, ctx, input, dbc))
	}

	// fetch as list
	webhooks, err := dbc.GetWebhooks(ctx, account.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, webhooks.Data)
	assert.Equal(t, len(createdWebhooks), len(webhooks.Data))

	// fetch as list (by trigger event ID from first webhook's first config)
	triggerEventID := createdWebhooks[0].TriggerConfigs[0].TriggerEventID
	webhooksByAccountAndEvent, err := dbc.GetWebhooksForAccountAndEvent(ctx, account.ID, triggerEventID)
	assert.NoError(t, err)
	assert.NotEmpty(t, webhooksByAccountAndEvent)

	// Create a catalog trigger event and add a trigger config for it
	catalogEvent2, err := dbc.CreateWebhookTriggerEvent(ctx, &types.WebhookTriggerEventDatabaseCreationInput{
		ID:          identifiers.New(),
		Name:        "test_event",
		Description: "test",
	})
	require.NoError(t, err)
	require.NotNil(t, catalogEvent2)

	createdConfig, err := dbc.AddWebhookTriggerConfig(ctx, account.ID, &types.WebhookTriggerConfigDatabaseCreationInput{
		ID:               identifiers.New(),
		BelongsToWebhook: createdWebhooks[0].ID,
		TriggerEventID:   catalogEvent2.ID,
	})
	assert.NoError(t, err)
	assert.NotNil(t, createdConfig)

	createdWebhooks[0].TriggerConfigs = append(createdWebhooks[0].TriggerConfigs, createdConfig)

	// Assert audit log entries were written for creates (pre-cleanup)
	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
		{EventType: string(audit.EventCreated), ResourceType: resourceTypeWebhooks, RelevantID: createdWebhooks[0].ID},
		{EventType: string(audit.EventCreated), ResourceType: resourceTypeWebhookTriggerConfigs, RelevantID: createdWebhooks[0].TriggerConfigs[0].ID},
		{EventType: string(audit.EventCreated), ResourceType: resourceTypeWebhookTriggerConfigs, RelevantID: createdConfig.ID},
	})

	// delete: archive trigger configs then archive webhook; archive catalog event if needed
	for _, webhook := range createdWebhooks {
		for _, cfg := range webhook.TriggerConfigs {
			assert.NoError(t, dbc.ArchiveWebhookTriggerConfig(ctx, webhook.ID, cfg.ID))
		}

		assert.NoError(t, dbc.ArchiveWebhook(ctx, webhook.ID, account.ID))
	}

	// Assert audit log entries were written for webhook archives (ArchiveWebhookTriggerConfig
	// does not set BelongsToAccount, so those entries are not returned by GetAuditLogEntriesForAccount)
	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
		{EventType: string(audit.EventArchived), ResourceType: resourceTypeWebhooks, RelevantID: createdWebhooks[0].ID},
	})

	for _, webhook := range createdWebhooks {
		var exists bool
		exists, err = dbc.WebhookExists(ctx, webhook.ID, account.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.Webhook
		y, err = dbc.GetWebhook(ctx, webhook.ID, account.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_GetWebhook(T *testing.T) {
	T.Parallel()

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetWebhook(ctx, "", exampleAccountID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhook := fakes.BuildFakeWebhook()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetWebhook(ctx, exampleWebhook.ID, "")
		assert.Error(t, err)
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
		assert.Error(t, err)
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
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateWebhookTriggerEvent(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		created, err := c.CreateWebhookTriggerEvent(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	// "with valid input" requires a real DB and is covered by integration tests
}

func TestQuerier_ArchiveWebhook(T *testing.T) {
	T.Parallel()

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveWebhook(ctx, "", exampleAccountID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhookID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveWebhook(ctx, exampleWebhookID, ""))
	})
}

func TestQuerier_ArchiveWebhookTriggerConfig(T *testing.T) {
	T.Parallel()

	T.Run("with invalid webhook ID", func(t *testing.T) {
		t.Parallel()

		exampleConfigID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveWebhookTriggerConfig(ctx, "", exampleConfigID))
	})

	T.Run("with invalid webhook trigger config ID", func(t *testing.T) {
		t.Parallel()

		exampleWebhookID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveWebhookTriggerConfig(ctx, exampleWebhookID, ""))
	})
}

func TestQuerier_ArchiveWebhookTriggerEvent(T *testing.T) {
	T.Parallel()

	T.Run("with invalid catalog event ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveWebhookTriggerEvent(ctx, ""))
	})
}

func TestQuerier_Integration_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	catalogEvent, err := dbc.CreateWebhookTriggerEvent(ctx, &types.WebhookTriggerEventDatabaseCreationInput{
		ID:          identifiers.New(),
		Name:        "pagination_event",
		Description: "for pagination test",
	})
	require.NoError(t, err)
	require.NotNil(t, catalogEvent)

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
			webhook.TriggerConfigs[0].TriggerEventID = catalogEvent.ID
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
				if err = dbc.ArchiveWebhookTriggerConfig(ctx, webhook.ID, cfg.ID); err != nil {
					return err
				}
			}
			return dbc.ArchiveWebhook(ctx, webhook.ID, account.ID)
		},
	})
}
