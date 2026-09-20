package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	webhookspb "github.com/primandproper/platform-go/v14/webhooks/webhookspb"
	"github.com/primandproper/primitives-go/v2/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The webhooks surface is platform's eleven RPCs now, and the model beneath it is a different
// shape rather than a renamed one.
//
// A webhook was one row carrying a method, a creator and a list of trigger configs. An endpoint
// is a URL and a name in an account, and what it hears about is a subscription — its own row,
// with its own id, addable and archivable without touching the endpoint. The method went because
// every delivery is a POST; the creator went because an endpoint belongs to an account rather
// than to a person, which is also why there is no longer a privacy collector over them.
//
// The caller supplies the signing secret. SaveEndpoint refuses an endpoint without one rather
// than minting it, which is the honest direction: the subscriber is the party that has to verify
// signatures, so the secret has to reach them regardless, and a server-minted one has to be read
// back out over an API to get there.
//
// Reached through WebhooksService() rather than as an embedded client: billing and webhooks both
// name three RPCs Subscription, so a type embedding both has ambiguous selectors and does not
// compile. See pkg/client.

// signingSecretForTest mints the secret a subscriber would verify deliveries with.
func signingSecretForTest() *webhookspb.WebhookSigningKeys {
	return &webhookspb.WebhookSigningKeys{Current: []byte(identifiers.New())}
}

// endpointInputForTest builds a registerable endpoint subscribed to one event type.
func endpointInputForTest(t *testing.T) *webhookspb.WebhookEndpointInput {
	t.Helper()

	return &webhookspb.WebhookEndpointInput{
		Name:        t.Name(),
		Url:         "https://" + identifiers.New() + ".example.com/webhook",
		ContentType: "application/json",
		EventTypes:  []string{webhooks.WebhookCreatedServiceEventType},
	}
}

func checkEndpointEquality(t *testing.T, expected *webhookspb.WebhookEndpointInput, actual *webhookspb.WebhookEndpoint) {
	t.Helper()

	assert.NotEmpty(t, actual.GetId(), "expected endpoint to have ID")
	assert.NotNil(t, actual.GetCreatedAt(), "expected endpoint to have CreatedAt")

	assert.Equal(t, expected.GetName(), actual.GetName(), "expected endpoint Name")
	assert.Equal(t, expected.GetUrl(), actual.GetUrl(), "expected endpoint URL")

	// The subscriptions the input asked for are the subscriptions the endpoint has, each with
	// an id of its own — which is what makes one removable without rewriting the endpoint.
	require.Len(t, actual.GetSubscriptions(), len(expected.GetEventTypes()), "expected endpoint subscriptions length")
	for i, eventType := range expected.GetEventTypes() {
		if i >= len(actual.GetSubscriptions()) {
			continue
		}

		subscription := actual.GetSubscriptions()[i]
		assert.NotEmpty(t, subscription.GetId(), "expected subscription %d to have ID", i)
		assert.NotNil(t, subscription.GetCreatedAt(), "expected subscription %d to have CreatedAt", i)
		assert.Equal(t, eventType, subscription.GetEventType(), "expected subscription %d EventType", i)
		assert.Equal(t, actual.GetId(), subscription.GetEndpointId(), "expected subscription %d EndpointID", i)
	}
}

// createWebhookForTest registers one endpoint and reads it back.
func createWebhookForTest(t *testing.T, testClient client.Client) *webhookspb.WebhookEndpoint {
	t.Helper()
	ctx := t.Context()

	input := endpointInputForTest(t)

	saved, err := testClient.WebhooksService().SaveEndpoint(ctx, &webhookspb.SaveEndpointRequest{
		Endpoint:    input,
		SigningKeys: signingSecretForTest(),
	})
	require.NoError(t, err)
	checkEndpointEquality(t, input, saved.GetResult())

	retrieved, err := testClient.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{
		EndpointId: saved.GetResult().GetId(),
	})
	require.NoError(t, err)
	require.NotNil(t, retrieved.GetResult())
	checkEndpointEquality(t, input, retrieved.GetResult())

	return retrieved.GetResult()
}

func TestWebhooks_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createWebhookForTest(t, testClient)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "webhooks", RelevantID: created.GetId()},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.WebhooksService().SaveEndpoint(ctx, &webhookspb.SaveEndpointRequest{})
		require.Error(t, err)
	})

	// An endpoint with no signing secret is refused rather than given one, because a
	// subscriber that cannot verify a delivery cannot tell it from an attacker's.
	T.Run("without a signing secret", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.WebhooksService().SaveEndpoint(ctx, &webhookspb.SaveEndpointRequest{
			Endpoint: endpointInputForTest(t),
		})
		assert.Error(t, err)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.WebhooksService().SaveEndpoint(ctx, &webhookspb.SaveEndpointRequest{
			Endpoint: &webhookspb.WebhookEndpointInput{
				Name:        t.Name(),
				Url:         "invalid protocol :\\ neato.ai",
				ContentType: "application/whatever",
			},
			SigningKeys: signingSecretForTest(),
		})
		assert.Error(t, err)
	})
}

func TestWebhooks_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createWebhookForTest(t, testClient)

		retrieved, err := testClient.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{
			EndpointId: created.GetId(),
		})
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrieved, err := testClient.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{
			EndpointId: nonexistentID,
		})
		require.Error(t, err)
		assert.Nil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{})
		assert.Error(t, err)
	})
}

func TestWebhooks_Listing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		created := []*webhookspb.WebhookEndpoint{}
		for range exampleQuantity {
			created = append(created, createWebhookForTest(t, testClient))
		}

		results, err := testClient.WebhooksService().ListEndpoints(ctx, &webhookspb.ListEndpointsRequest{})
		require.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.GetResults()), len(created))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.WebhooksService().ListEndpoints(ctx, &webhookspb.ListEndpointsRequest{})
		assert.Error(t, err)
	})
}

func TestWebhooks_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createWebhookForTest(t, testClient)

		_, err := testClient.WebhooksService().ArchiveEndpoint(ctx, &webhookspb.ArchiveEndpointRequest{
			EndpointId: created.GetId(),
		})
		require.NoError(t, err)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 10, []*ExpectedAuditEntry{
			{EventType: "archived", ResourceType: "webhooks", RelevantID: created.GetId()},
		})
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createWebhookForTest(t, testClient)

		_, err := testClient.WebhooksService().ArchiveEndpoint(ctx, &webhookspb.ArchiveEndpointRequest{
			EndpointId: nonexistentID,
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.WebhooksService().ArchiveEndpoint(ctx, &webhookspb.ArchiveEndpointRequest{})
		assert.Error(t, err)
	})
}

// TestWebhooks_RotatingSecrets covers the RPC the local service had no equivalent for.
//
// A rotation is its own call rather than a field on the endpoint, because it is the one write
// that invalidates every signature a subscriber has already learned to check — and the keyring
// keeps the previous secret so a subscriber that has not yet picked up the new one still
// verifies.
func TestWebhooks_RotatingSecrets(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createWebhookForTest(t, testClient)

		_, err := testClient.WebhooksService().RotateSecret(ctx, &webhookspb.RotateSecretRequest{
			EndpointId: created.GetId(),
			SigningKey: []byte(identifiers.New()),
		})
		require.NoError(t, err)

		// The secret does not come back on a read, which is the property that makes an
		// endpoint safe for an account member to look at.
		retrieved, err := testClient.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{
			EndpointId: created.GetId(),
		})
		require.NoError(t, err)
		assert.NotNil(t, retrieved.GetResult())
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.WebhooksService().RotateSecret(ctx, &webhookspb.RotateSecretRequest{
			EndpointId: nonexistentID,
			SigningKey: []byte(identifiers.New()),
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.WebhooksService().RotateSecret(ctx, &webhookspb.RotateSecretRequest{})
		assert.Error(t, err)
	})
}

func TestWebhookSubscriptions_Adding(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createWebhookForTest(t, testClient)

		added, err := testClient.WebhooksService().AddSubscription(ctx, &webhookspb.AddSubscriptionRequest{
			EndpointId: created.GetId(),
			EventType:  webhooks.WebhookArchivedServiceEventType,
		})
		require.NoError(t, err)
		require.NotNil(t, added.GetResult())
		assert.Equal(t, webhooks.WebhookArchivedServiceEventType, added.GetResult().GetEventType())
		assert.Equal(t, created.GetId(), added.GetResult().GetEndpointId())

		// Readable on its own, which is what having an id of its own buys.
		fetched, err := testClient.WebhooksService().GetSubscription(ctx, &webhookspb.GetSubscriptionRequest{
			SubscriptionId: added.GetResult().GetId(),
		})
		require.NoError(t, err)
		assert.Equal(t, added.GetResult().GetId(), fetched.GetResult().GetId())

		// And on the endpoint's list, alongside the one the registration asked for.
		listed, err := testClient.WebhooksService().ListSubscriptions(ctx, &webhookspb.ListSubscriptionsRequest{
			EndpointId: created.GetId(),
		})
		require.NoError(t, err)
		assert.Len(t, listed.GetResults(), 2)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "webhook_subscriptions", RelevantID: added.GetResult().GetId()},
		})
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createWebhookForTest(t, testClient)

		_, err := testClient.WebhooksService().AddSubscription(ctx, &webhookspb.AddSubscriptionRequest{
			EndpointId: nonexistentID,
			EventType:  webhooks.WebhookArchivedServiceEventType,
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.WebhooksService().AddSubscription(ctx, &webhookspb.AddSubscriptionRequest{})
		assert.Error(t, err)
	})
}

func TestWebhookSubscriptions_Removing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createWebhookForTest(t, testClient)

		added, err := testClient.WebhooksService().AddSubscription(ctx, &webhookspb.AddSubscriptionRequest{
			EndpointId: created.GetId(),
			EventType:  webhooks.WebhookArchivedServiceEventType,
		})
		require.NoError(t, err)

		// Archiving a subscription leaves the endpoint standing, which is the whole reason
		// a subscription is a row rather than a field.
		_, err = testClient.WebhooksService().ArchiveSubscription(ctx, &webhookspb.ArchiveSubscriptionRequest{
			SubscriptionId: added.GetResult().GetId(),
		})
		require.NoError(t, err)

		stillThere, err := testClient.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{
			EndpointId: created.GetId(),
		})
		require.NoError(t, err)
		assert.NotNil(t, stillThere.GetResult())

		listed, err := testClient.WebhooksService().ListSubscriptions(ctx, &webhookspb.ListSubscriptionsRequest{
			EndpointId: created.GetId(),
		})
		require.NoError(t, err)
		for _, subscription := range listed.GetResults() {
			assert.NotEqual(t, added.GetResult().GetId(), subscription.GetId(), "an archived subscription is still listed")
		}
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createWebhookForTest(t, testClient)

		_, err := testClient.WebhooksService().ArchiveSubscription(ctx, &webhookspb.ArchiveSubscriptionRequest{
			SubscriptionId: nonexistentID,
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.WebhooksService().ArchiveSubscription(ctx, &webhookspb.ArchiveSubscriptionRequest{})
		assert.Error(t, err)
	})
}

// TestWebhookEventTypes_Listing covers the eleventh RPC, which is what a client building a
// subscription UI reads to know what it may subscribe to.
//
// It takes no filter and no scope, both deliberate upstream: the catalog is a Go value
// identical for every tenant rather than a table, so there is nothing to page and nothing to
// narrow. What is in it here is this application's catalog, generated from the event type
// constants — see internal/domain/webhooks/catalog.
func TestWebhookEventTypes_Listing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		results, err := testClient.WebhooksService().ListEventTypes(ctx, &webhookspb.ListEventTypesRequest{})
		require.NoError(t, err)
		require.NotEmpty(t, results.GetResults())

		// Every event type carries the prose saying when it fires, which is the reason the
		// catalog is a surface at all rather than a constant a client hardcodes.
		byType := map[string]string{}
		for _, definition := range results.GetResults() {
			assert.NotEmpty(t, definition.GetEventType())
			assert.NotEmpty(t, definition.GetDescription(), "event type %q has no description", definition.GetEventType())
			byType[definition.GetEventType()] = definition.GetDescription()
		}

		// The one a subscription in this file is made against has to be in it, or the
		// subscription above was accepted against an event nothing will ever fire.
		assert.Contains(t, byType, webhooks.WebhookCreatedServiceEventType)
		assert.Contains(t, byType, webhooks.WebhookArchivedServiceEventType)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.WebhooksService().ListEventTypes(ctx, &webhookspb.ListEventTypesRequest{})
		assert.Error(t, err)
	})
}
