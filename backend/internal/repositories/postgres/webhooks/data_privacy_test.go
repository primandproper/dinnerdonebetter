package webhooks

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v6/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Unit tests ---

func TestCollectUserData(T *testing.T) {
	T.Parallel()

	T.Run("with nil account IDs", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		result, err := c.CollectUserData(ctx, nil)
		assert.NoError(t, err)
		require.NotNil(t, result)
		assert.NotNil(t, result.Data)
		assert.Empty(t, result.Data)
	})

	T.Run("with empty account IDs", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		result, err := c.CollectUserData(ctx, []string{})
		assert.NoError(t, err)
		require.NotNil(t, result)
		assert.NotNil(t, result.Data)
		assert.Empty(t, result.Data)
	})
}

// --- Integration tests ---

func TestQuerier_Integration_CollectUserData(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, container := buildDatabaseClientForTest(t)

	_, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	// Create catalog trigger event for webhook
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
	dbInput := converters.ConvertWebhookToWebhookDatabaseCreationInput(exampleWebhook)

	created, err := dbc.CreateWebhook(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)

	// Collect user data for account
	result, err := dbc.CollectUserData(ctx, []string{account.ID})
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Data)
	assert.Contains(t, result.Data, account.ID)
	assert.Len(t, result.Data[account.ID], 1)
	assert.Equal(t, created.ID, result.Data[account.ID][0].ID)
	assert.Equal(t, account.ID, result.Data[account.ID][0].BelongsToAccount)
}

func TestQuerier_Integration_CollectUserData_MultipleAccounts(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, container := buildDatabaseClientForTest(t)

	_, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	user1 := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account1 := pgtesting.CreateAccountForTest(t, nil, user1.ID, dbc.writeDB)

	user2 := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account2 := pgtesting.CreateAccountForTest(t, nil, user2.ID, dbc.writeDB)

	catalogEvent, err := dbc.CreateWebhookTriggerEvent(ctx, &types.WebhookTriggerEventDatabaseCreationInput{
		ID:          identifiers.New(),
		Name:        "webhook_created",
		Description: "test",
	})
	require.NoError(t, err)
	require.NotNil(t, catalogEvent)

	// Create webhook for account1 only
	webhook1 := fakes.BuildFakeWebhook()
	webhook1.BelongsToAccount = account1.ID
	webhook1.CreatedByUser = user1.ID
	webhook1.TriggerConfigs[0].TriggerEventID = catalogEvent.ID
	dbInput1 := converters.ConvertWebhookToWebhookDatabaseCreationInput(webhook1)
	created1, err := dbc.CreateWebhook(ctx, dbInput1)
	require.NoError(t, err)
	require.NotNil(t, created1)

	// Collect for both accounts - only accounts with webhooks appear in result
	result, err := dbc.CollectUserData(ctx, []string{account1.ID, account2.ID})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Contains(t, result.Data, account1.ID)
	assert.Len(t, result.Data[account1.ID], 1)
}
