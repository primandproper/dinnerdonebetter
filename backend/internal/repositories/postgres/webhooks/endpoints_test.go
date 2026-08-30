package webhooks

import (
	"database/sql"
	"encoding/hex"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"
	generated "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooks/generated"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/tenancy"
	"github.com/primandproper/platform-go/v13/webhooks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createWebhookWithAccountForTest creates a user, an account, and one webhook belonging to it.
func createWebhookWithAccountForTest(t *testing.T, dbc *repository) (webhook *types.Webhook, accountID, secret string) {
	t.Helper()

	ctx := t.Context()

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	exampleWebhook := fakes.BuildFakeWebhook()
	exampleWebhook.BelongsToAccount = account.ID
	exampleWebhook.CreatedByUser = user.ID

	response, err := dbc.CreateWebhook(ctx, converters.ConvertWebhookToWebhookDatabaseCreationInput(exampleWebhook))
	require.NoError(t, err)
	require.NotNil(t, response)

	return response.Webhook, account.ID, response.Secret
}

// TestIntegration_WebhookEndpoint_Scope is the tenancy property, asserted on the row rather than
// on a delivery.
//
// The account used to travel inside the subscription's event type, as <accountID>:<eventType>,
// because platform-go's endpoints had no owner. They carry a tenancy.Scope now, so the account is
// a column and the event type is the plain catalog string — which is the whole of this change.
func TestIntegration_WebhookEndpoint_Scope(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	webhook, accountID, _ := createWebhookWithAccountForTest(t, dbc)

	endpoint, err := dbc.endpoints.GetEndpoint(ctx, tenancy.Of(accountID), webhook.ID)
	require.NoError(t, err)

	assert.Equal(t, tenancy.Of(accountID), endpoint.Scope)
	assert.Equal(t, webhook.ID, endpoint.ID)
	assert.Equal(t, webhook.URL, endpoint.URL)
	assert.Equal(t, webhook.ContentType, endpoint.ContentType)

	// Unqualified: the same string the catalog holds, and the same one a subscriber reads out
	// of X-Platform-Event.
	assert.Equal(t, []webhooks.EventType{webhooks.EventType(webhook.TriggerConfigs[0].EventType)}, endpoint.EventTypes())

	// Another account's read of the same ID finds nothing, which is what the scope buys — and
	// so does the global scope, which matches only itself.
	_, err = dbc.endpoints.GetEndpoint(ctx, tenancy.Of(fake.BuildFakeID()), webhook.ID)
	require.ErrorIs(t, err, sql.ErrNoRows)

	_, err = dbc.endpoints.GetEndpoint(ctx, tenancy.Global(), webhook.ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

// TestIntegration_WebhookEndpoint_Subscriptions covers the derivation: trigger configs are the
// account-facing record, and the endpoint's subscriptions are rewritten from them.
func TestIntegration_WebhookEndpoint_Subscriptions(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	webhook, accountID, _ := createWebhookWithAccountForTest(t, dbc)
	scope := tenancy.Of(accountID)

	firstEventType := webhook.TriggerConfigs[0].EventType

	// It has to differ from the one the fake already carries:
	// (trigger_event, belongs_to_webhook, archived_at) is unique.
	secondEventType := fakes.BuildFakeWebhookEventType()
	for secondEventType == firstEventType {
		secondEventType = fakes.BuildFakeWebhookEventType()
	}

	secondConfig, err := dbc.AddWebhookTriggerConfig(ctx, accountID, &types.WebhookTriggerConfigDatabaseCreationInput{
		ID:               identifiers.New(),
		BelongsToWebhook: webhook.ID,
		EventType:        secondEventType,
	})
	require.NoError(t, err)

	endpoint, err := dbc.endpoints.GetEndpoint(ctx, scope, webhook.ID)
	require.NoError(t, err)
	assert.ElementsMatch(t, []webhooks.EventType{
		webhooks.EventType(firstEventType),
		webhooks.EventType(secondEventType),
	}, endpoint.EventTypes())

	require.NoError(t, dbc.ArchiveWebhookTriggerConfig(ctx, webhook.ID, accountID, secondConfig.ID))

	endpoint, err = dbc.endpoints.GetEndpoint(ctx, scope, webhook.ID)
	require.NoError(t, err)
	assert.Equal(t, []webhooks.EventType{webhooks.EventType(firstEventType)}, endpoint.EventTypes())

	// Unsubscribing from the last event type leaves an endpoint that is registered and
	// subscribed to nothing. It is a real state — inert, because fan-out reads the
	// subscriptions table — and not one Dispatcher.Register would accept, which is why the
	// subscription set is written through the Store.
	require.NoError(t, dbc.ArchiveWebhookTriggerConfig(ctx, webhook.ID, accountID, webhook.TriggerConfigs[0].ID))

	endpoint, err = dbc.endpoints.GetEndpoint(ctx, scope, webhook.ID)
	require.NoError(t, err)
	assert.Empty(t, endpoint.EventTypes())
}

// TestIntegration_RotateWebhookSecret covers the rotation window: both keys sign every delivery
// while Previous is set, so a subscriber accepts either signature while it switches.
func TestIntegration_RotateWebhookSecret(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	webhook, accountID, original := createWebhookWithAccountForTest(t, dbc)

	rotated, err := dbc.RotateWebhookSecret(ctx, webhook.ID, accountID)
	require.NoError(t, err)
	assert.NotEqual(t, original, rotated)

	endpoint, err := dbc.endpoints.GetEndpoint(ctx, tenancy.Of(accountID), webhook.ID)
	require.NoError(t, err)

	assert.Equal(t, decodeSecret(t, rotated), endpoint.Secret.Current)
	assert.Equal(t, decodeSecret(t, original), endpoint.Secret.Previous)

	// Another account naming the same ID is refused, so knowing an ID is not enough to roll
	// somebody else's signing key and break every delivery to them.
	_, err = dbc.RotateWebhookSecret(ctx, webhook.ID, fake.BuildFakeID())
	require.Error(t, err)
}

// TestIntegration_RotateWebhookSecret_RegistersMissingEndpoint covers adoption.
//
// CreateWebhook commits the webhook row before it registers the endpoint, deliberately: a
// failure between the two leaves a webhook that exists and does not yet deliver, rather than an
// endpoint receiving an account's events with nothing in the API that says it exists. Rotating
// the secret is how that webhook is adopted, and the only operation whose response can hand its
// owner a secret.
func TestIntegration_RotateWebhookSecret_RegistersMissingEndpoint(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	exampleWebhook := fakes.BuildFakeWebhook()
	exampleWebhook.BelongsToAccount = account.ID
	exampleWebhook.CreatedByUser = user.ID

	// The rows CreateWebhook's transaction writes, without the registration that follows it.
	require.NoError(t, dbc.generatedQuerier.CreateWebhook(ctx, dbc.writeDB, &generated.CreateWebhookParams{
		ID:               exampleWebhook.ID,
		Name:             exampleWebhook.Name,
		ContentType:      generated.WebhookContentType(exampleWebhook.ContentType),
		URL:              exampleWebhook.URL,
		Method:           generated.WebhookMethod(exampleWebhook.Method),
		CreatedByUser:    exampleWebhook.CreatedByUser,
		BelongsToAccount: exampleWebhook.BelongsToAccount,
	}))
	require.NoError(t, dbc.generatedQuerier.CreateWebhookTriggerConfig(ctx, dbc.writeDB, &generated.CreateWebhookTriggerConfigParams{
		ID:               exampleWebhook.TriggerConfigs[0].ID,
		TriggerEvent:     exampleWebhook.TriggerConfigs[0].EventType,
		BelongsToWebhook: exampleWebhook.ID,
	}))

	_, err := dbc.endpoints.GetEndpoint(ctx, tenancy.Of(account.ID), exampleWebhook.ID)
	require.ErrorIs(t, err, sql.ErrNoRows)

	secret, err := dbc.RotateWebhookSecret(ctx, exampleWebhook.ID, account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, secret)

	endpoint, err := dbc.endpoints.GetEndpoint(ctx, tenancy.Of(account.ID), exampleWebhook.ID)
	require.NoError(t, err)

	assert.Equal(t, decodeSecret(t, secret), endpoint.Secret.Current)
	// Registered, not rotated: there is no outgoing key to keep signing under.
	assert.Empty(t, endpoint.Secret.Previous)
	assert.Equal(t, []webhooks.EventType{webhooks.EventType(exampleWebhook.TriggerConfigs[0].EventType)}, endpoint.EventTypes())
}

// TestIntegration_ArchiveWebhook_RetiresEndpoint asserts the endpoint stops matching fan-out when
// the webhook is archived.
func TestIntegration_ArchiveWebhook_RetiresEndpoint(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	webhook, accountID, _ := createWebhookWithAccountForTest(t, dbc)
	scope := tenancy.Of(accountID)
	eventType := webhooks.EventType(webhook.TriggerConfigs[0].EventType)

	subscribed, err := dbc.endpoints.EndpointsForEvent(ctx, dbc.readDB, scope, eventType)
	require.NoError(t, err)
	require.Len(t, subscribed, 1)
	assert.Equal(t, webhook.ID, subscribed[0].ID)

	// Another account's fan-out for the same event type never saw it.
	subscribed, err = dbc.endpoints.EndpointsForEvent(ctx, dbc.readDB, tenancy.Of(fake.BuildFakeID()), eventType)
	require.NoError(t, err)
	assert.Empty(t, subscribed)

	require.NoError(t, dbc.ArchiveWebhook(ctx, webhook.ID, accountID))

	subscribed, err = dbc.endpoints.EndpointsForEvent(ctx, dbc.readDB, scope, eventType)
	require.NoError(t, err)
	assert.Empty(t, subscribed)
}

// decodeSecret turns the hex a rotation returns back into the bytes deliveries are signed with.
func decodeSecret(t *testing.T, secret string) []byte {
	t.Helper()

	decoded, err := hex.DecodeString(secret)
	require.NoError(t, err)

	return decoded
}
