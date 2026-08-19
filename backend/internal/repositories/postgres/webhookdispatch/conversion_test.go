package webhookdispatch

import (
	"testing"

	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/tenancy"
	"github.com/primandproper/platform-go/v11/webhooks"
	webhooksmock "github.com/primandproper/platform-go/v11/webhooks/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToPlatformEndpoint(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		secret := []byte("the-current-signing-secret-bytes")

		endpoint := toPlatformEndpoint(&Registration{
			ID:          exampleWebhookID,
			AccountID:   exampleAccountID,
			URL:         "https://example.com/hook",
			ContentType: "application/json",
			EventTypes:  []string{exampleEventType},
		}, secret)

		require.NotNil(t, endpoint)
		assert.Equal(t, exampleWebhookID, endpoint.ID)
		assert.Equal(t, "https://example.com/hook", endpoint.URL)
		assert.Equal(t, "application/json", endpoint.ContentType)
		assert.Equal(t, secret, endpoint.Secret.Current)
		assert.Empty(t, endpoint.Secret.Previous)
		assert.Equal(t, tenancy.Global(), endpoint.Scope)
	})

	T.Run("leaves the endpoint subscribed to nothing", func(t *testing.T) {
		t.Parallel()

		// Registration's event types are deliberately not read here. Subscriptions are
		// qualified, and the qualification rule has one implementation —
		// setAccountSubscriptions — so that creating an endpoint and amending one cannot
		// namespace event types differently.
		endpoint := toPlatformEndpoint(&Registration{
			ID:         exampleWebhookID,
			AccountID:  exampleAccountID,
			EventTypes: []string{exampleEventType},
		}, nil)

		require.NotNil(t, endpoint)
		assert.Empty(t, endpoint.Events)
	})

	T.Run("with no content type", func(t *testing.T) {
		t.Parallel()

		// EnsureDefaults is applied here rather than left to the caller: an endpoint stored
		// with no Content-Type delivers requests a subscriber's framework may refuse to parse.
		endpoint := toPlatformEndpoint(&Registration{ID: exampleWebhookID}, nil)

		require.NotNil(t, endpoint)
		assert.NotEmpty(t, endpoint.ContentType)
	})
}

func TestDispatcher_setAccountSubscriptions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		d := buildDispatcherForTest(t, &webhooksmock.StoreMock{})
		endpoint := &webhooks.Endpoint{ID: exampleWebhookID}

		require.NoError(t, d.setAccountSubscriptions(endpoint, exampleAccountID, []string{exampleEventType}))

		// The account is inside the subscription string; that is the whole tenancy mechanism.
		assert.Equal(t, []webhooks.EventType{exampleAccountID + ":" + exampleEventType}, endpoint.Events)
	})

	T.Run("replaces this account's subscriptions and keeps every other account's", func(t *testing.T) {
		t.Parallel()

		otherAccountSubscription := webhooks.EventType("acct_other:meal_plan_updated")

		d := buildDispatcherForTest(t, &webhooksmock.StoreMock{})
		endpoint := &webhooks.Endpoint{
			ID: exampleWebhookID,
			Events: []webhooks.EventType{
				exampleAccountID + ":" + exampleEventType,
				otherAccountSubscription,
			},
		}

		require.NoError(t, d.setAccountSubscriptions(endpoint, exampleAccountID, []string{"meal_plan_updated"}))

		// SaveEndpoint replaces an endpoint's whole subscription set, so a rewrite built from
		// one account's event types alone would silently drop another account's.
		assert.ElementsMatch(t, []webhooks.EventType{
			otherAccountSubscription,
			exampleAccountID + ":meal_plan_updated",
		}, endpoint.Events)
	})

	T.Run("with no event types", func(t *testing.T) {
		t.Parallel()

		d := buildDispatcherForTest(t, &webhooksmock.StoreMock{})
		endpoint := &webhooks.Endpoint{
			ID:     exampleWebhookID,
			Events: []webhooks.EventType{exampleAccountID + ":" + exampleEventType},
		}

		// Unsubscribing from the last event type is a real state, not an error: the endpoint
		// stays registered and goes inert, because fan-out reads the subscriptions table.
		require.NoError(t, d.setAccountSubscriptions(endpoint, exampleAccountID, nil))
		assert.Empty(t, endpoint.Events)
	})

	T.Run("with event type outside the catalog", func(t *testing.T) {
		t.Parallel()

		d := buildDispatcherForTest(t, &webhooksmock.StoreMock{})
		endpoint := &webhooks.Endpoint{ID: exampleWebhookID}

		err := d.setAccountSubscriptions(endpoint, exampleAccountID, []string{"reciped_created"})
		require.ErrorIs(t, err, webhooks.ErrUnknownEventType)
		// The endpoint is left as it was, so a caller that ignores the error does not save a
		// half-applied subscription set.
		assert.Empty(t, endpoint.Events)
	})

	T.Run("with no account", func(t *testing.T) {
		t.Parallel()

		// An unqualified subscription string is one every account's fan-out would match, so
		// there is no sensible thing to write without an account.
		d := buildDispatcherForTest(t, &webhooksmock.StoreMock{})
		endpoint := &webhooks.Endpoint{ID: exampleWebhookID}

		err := d.setAccountSubscriptions(endpoint, "", []string{exampleEventType})
		require.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
		assert.Empty(t, endpoint.Events)
	})
}

func TestQualification(T *testing.T) {
	T.Parallel()

	T.Run("round trips", func(t *testing.T) {
		t.Parallel()

		subscription := qualify(exampleAccountID, exampleEventType)
		assert.Equal(t, webhooks.EventType(exampleAccountID+":"+exampleEventType), subscription)

		eventType, ok := unqualify(exampleAccountID, subscription)
		assert.True(t, ok)
		assert.Equal(t, exampleEventType, eventType)
	})

	T.Run("rejects another account's subscription", func(t *testing.T) {
		t.Parallel()

		// The prefix must match in full. A partial match here would leak one account's
		// subscriptions into another's set on the read-modify-write in SetEventTypes.
		_, ok := unqualify("acct_abc", qualify(exampleAccountID, exampleEventType))
		assert.False(t, ok)
	})
}

func TestDispatcher_checkKnown(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		d := buildDispatcherForTest(t, &webhooksmock.StoreMock{})
		require.NoError(t, d.checkKnown([]string{exampleEventType, "meal_plan_updated"}))
	})

	T.Run("with event type outside the catalog", func(t *testing.T) {
		t.Parallel()

		// The gate is on the unqualified type, which is the one a catalog can hold: the
		// qualified string qualify produces would need one catalog entry per account.
		d := buildDispatcherForTest(t, &webhooksmock.StoreMock{})

		err := d.checkKnown([]string{exampleEventType, "reciped_created"})
		require.ErrorIs(t, err, webhooks.ErrUnknownEventType)
		require.ErrorIs(t, err, ErrUnknownEventType)
	})
}
