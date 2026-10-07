package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	platformwebhooks "github.com/primandproper/platform-go/v15/webhooks"
	webhookspb "github.com/primandproper/platform-go/v15/webhooks/webhookspb"
	"github.com/primandproper/primitives-go/v2/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The webhooks surface's behavior is asserted by platform's webhooks conformance suite, run
// against this deployment in conformance_test.go. What remains here is this application's own:
// the audit entries its store records, and its event type catalog.
//
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

	// A literal address in TEST-NET-1 rather than a hostname. platform resolves an
	// endpoint's host before it will register one — an SSRF check, since a delivery is
	// this deployment making a request somebody else chose the target of — so a made-up
	// name fails on DNS rather than on anything this test is about. 192.0.2.0/24 is
	// documentation space: routable-looking, in none of the ranges checkIP refuses, and
	// nothing is listening. The path carries the uniqueness instead.
	return &webhookspb.WebhookEndpointInput{
		Name:        t.Name(),
		Url:         "https://192.0.2.1/webhook/" + identifiers.New(),
		ContentType: "application/json",
		EventTypes:  []string{platformwebhooks.EventEndpointCreated.String()},
	}
}

// createWebhookForTest registers one endpoint. What the save answers, and that it reads back
// the same, is conformance/webhooks'.
func createWebhookForTest(t *testing.T, testClient client.Client) *webhookspb.WebhookEndpoint {
	t.Helper()

	saved, err := testClient.WebhooksService().SaveEndpoint(t.Context(), &webhookspb.SaveEndpointRequest{
		Endpoint:    endpointInputForTest(t),
		SigningKeys: signingSecretForTest(),
	})
	require.NoError(t, err)
	require.NotNil(t, saved.GetResult())

	return saved.GetResult()
}

func TestWebhooks_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createWebhookForTest(t, testClient)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: platformwebhooks.ResourceTypeEndpoint, RelevantID: created.GetId()},
		})
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
			{EventType: "archived", ResourceType: platformwebhooks.ResourceTypeEndpoint, RelevantID: created.GetId()},
		})
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
			EventType:  platformwebhooks.EventEndpointArchived.String(),
		})
		require.NoError(t, err)
		require.NotNil(t, added.GetResult())

		// The entry is platform's, named with platform's resource type: the store's
		// RecordingHooks write it, on the same transaction as the subscription.
		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: platformwebhooks.ResourceTypeSubscription, RelevantID: added.GetResult().GetId()},
		})
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
		assert.Contains(t, byType, platformwebhooks.EventEndpointCreated.String())
		assert.Contains(t, byType, platformwebhooks.EventEndpointArchived.String())
	})
}
