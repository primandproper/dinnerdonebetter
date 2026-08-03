package webhookdispatch

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/primandproper/platform-go/v9/database"
	mockdatabase "github.com/primandproper/platform-go/v9/database/mock"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v9/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"
	"github.com/primandproper/platform-go/v9/webhooks"
	webhooksmock "github.com/primandproper/platform-go/v9/webhooks/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleAccountID = "acct_abc123"
	exampleEventType = "meal_plan_created"
	exampleWebhookID = "wh_def456"
)

// testCatalog is a two-entry stand-in for the generated catalog. Using the real one would make
// these tests fail whenever a domain constant was added, which tests nothing about dispatch.
func testCatalog() webhooks.Catalog {
	return webhooks.Catalog{
		exampleEventType:    {Description: "A meal plan was created."},
		"meal_plan_updated": {Description: "A meal plan was updated."},
	}
}

func buildDispatcherForTest(t *testing.T, store webhooks.Store) *Dispatcher {
	t.Helper()

	d, err := NewDispatcher(
		store,
		testCatalog(),
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
	)
	require.NoError(t, err)

	return d
}

func TestNewDispatcher(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		d, err := NewDispatcher(
			&webhooksmock.StoreMock{},
			testCatalog(),
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			metricsnoop.NewMetricsProvider(),
		)
		require.NoError(t, err)
		assert.NotNil(t, d)
	})

	T.Run("with nil store", func(t *testing.T) {
		t.Parallel()

		d, err := NewDispatcher(nil, testCatalog(), loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider())
		require.Error(t, err)
		assert.Nil(t, d)
	})

	T.Run("with empty catalog", func(t *testing.T) {
		t.Parallel()

		// An empty catalog rejects every event type, which would present as a total webhook
		// outage made of individually plausible rejections. Refused at construction instead.
		d, err := NewDispatcher(&webhooksmock.StoreMock{}, webhooks.Catalog{}, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider())
		require.Error(t, err)
		assert.Nil(t, d)
	})
}

func TestDispatcher_Dispatch(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		executor := &mockdatabase.SQLQueryExecutorMock{}

		store := &webhooksmock.StoreMock{
			EndpointsForEventFunc: func(_ context.Context, q database.SQLQueryExecutor, eventType string) ([]*webhooks.Endpoint, error) {
				// The account is inside the subscription string; that is the whole
				// tenancy mechanism, so it is what this asserts.
				assert.Equal(t, exampleAccountID+":"+exampleEventType, eventType)
				assert.Same(t, executor, q)

				return []*webhooks.Endpoint{{ID: exampleWebhookID}}, nil
			},
			EnqueueFunc: func(_ context.Context, q database.SQLQueryExecutor, delivery *webhooks.Delivery, endpointIDs []string, _ time.Time) error {
				assert.Same(t, executor, q)
				assert.Equal(t, []string{exampleWebhookID}, endpointIDs)
				// The delivery carries the unqualified type: that is what reaches the
				// subscriber in the X-Platform-Event header.
				assert.Equal(t, exampleEventType, delivery.EventType)
				// Store.Enqueue does not mint this — webhooks.Dispatcher did, and this
				// package replaced it. Left empty, every delivery shares the empty string
				// as its primary key and the header a subscriber deduplicates on is blank.
				assert.NotEmpty(t, delivery.ID)
				assert.Equal(t, "meal_plan_1", delivery.OrderingKey)
				assert.JSONEq(t, `{"hello":"world"}`, string(delivery.Payload))

				return nil
			},
		}

		d := buildDispatcherForTest(t, store)

		err := d.Dispatch(ctx, executor, exampleAccountID, exampleEventType, "meal_plan_1", json.RawMessage(`{"hello":"world"}`))
		require.NoError(t, err)
		assert.Len(t, store.EndpointsForEventCalls(), 1)
		assert.Len(t, store.EnqueueCalls(), 1)
	})

	T.Run("with no subscribers", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		store := &webhooksmock.StoreMock{
			EndpointsForEventFunc: func(context.Context, database.SQLQueryExecutor, string) ([]*webhooks.Endpoint, error) {
				return nil, nil
			},
		}

		d := buildDispatcherForTest(t, store)

		// An event nobody subscribes to is the common case and writes nothing. Enqueue is
		// unconfigured, so calling it would panic — which is the assertion.
		err := d.Dispatch(ctx, &mockdatabase.SQLQueryExecutorMock{}, exampleAccountID, exampleEventType, "", json.RawMessage(`{}`))
		require.NoError(t, err)
		assert.Empty(t, store.EnqueueCalls())
	})

	T.Run("with event type outside the catalog", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		store := &webhooksmock.StoreMock{}

		d := buildDispatcherForTest(t, store)

		// Skipped, not rejected. Dispatch runs inside the transaction that wrote the row the
		// event describes, so returning an error here would not fail a webhook — it would
		// fail the write. Most of what lands here is a deliberate exclusion (a sign-in, a
		// two-factor change) that the application publishes and no webhook may receive.
		//
		// A typo is caught at registration, where a human types one, and by the catalog's own
		// test — at build time rather than by taking down a write at runtime.
		err := d.Dispatch(ctx, &mockdatabase.SQLQueryExecutorMock{}, exampleAccountID, "reciped_created", "", json.RawMessage(`{}`))
		require.NoError(t, err)
		assert.Empty(t, store.EndpointsForEventCalls())
		assert.Empty(t, store.EnqueueCalls())
	})

	T.Run("with no account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		store := &webhooksmock.StoreMock{}

		d := buildDispatcherForTest(t, store)

		// Background jobs emit events with no account. Those belong to no subscriber.
		err := d.Dispatch(ctx, &mockdatabase.SQLQueryExecutorMock{}, "", exampleEventType, "", json.RawMessage(`{}`))
		require.NoError(t, err)
		assert.Empty(t, store.EndpointsForEventCalls())
	})

	T.Run("with nil executor", func(t *testing.T) {
		t.Parallel()

		d := buildDispatcherForTest(t, &webhooksmock.StoreMock{})

		err := d.Dispatch(t.Context(), nil, exampleAccountID, exampleEventType, "", json.RawMessage(`{}`))
		assert.Error(t, err)
	})

	T.Run("with nil dispatcher", func(t *testing.T) {
		t.Parallel()

		// Inert rather than an error: a process wired without webhooks must still be able to
		// write rows and emit events, and dispatch is a side effect of a write it must not
		// fail.
		var d *Dispatcher

		err := d.Dispatch(t.Context(), &mockdatabase.SQLQueryExecutorMock{}, exampleAccountID, exampleEventType, "", json.RawMessage(`{}`))
		assert.NoError(t, err)
	})

	T.Run("with error resolving endpoints", func(t *testing.T) {
		t.Parallel()

		expected := errors.New("blah")
		store := &webhooksmock.StoreMock{
			EndpointsForEventFunc: func(context.Context, database.SQLQueryExecutor, string) ([]*webhooks.Endpoint, error) {
				return nil, expected
			},
		}

		d := buildDispatcherForTest(t, store)

		err := d.Dispatch(t.Context(), &mockdatabase.SQLQueryExecutorMock{}, exampleAccountID, exampleEventType, "", json.RawMessage(`{}`))
		assert.Error(t, err)
	})
}

func TestDispatcher_Register(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		var saved *webhooks.Endpoint

		store := &webhooksmock.StoreMock{
			SaveEndpointFunc: func(_ context.Context, endpoint *webhooks.Endpoint) error {
				saved = endpoint

				return nil
			},
		}

		d := buildDispatcherForTest(t, store)

		secret, err := d.Register(t.Context(), &Registration{
			ID:          exampleWebhookID,
			AccountID:   exampleAccountID,
			URL:         "https://example.com/hook",
			ContentType: "application/json",
			EventTypes:  []string{exampleEventType},
		})
		require.NoError(t, err)

		// The returned secret is hex of exactly the bytes stored to sign with — a subscriber
		// that decodes it must be able to reproduce the signature.
		decoded, err := hex.DecodeString(secret)
		require.NoError(t, err)
		require.NotNil(t, saved)
		assert.Equal(t, decoded, saved.Secret.Current)
		assert.Len(t, decoded, secretBytes)
		assert.Empty(t, saved.Secret.Previous)

		assert.Equal(t, exampleWebhookID, saved.ID)
		assert.Equal(t, []string{exampleAccountID + ":" + exampleEventType}, saved.Events)
	})

	T.Run("with event type outside the catalog", func(t *testing.T) {
		t.Parallel()

		store := &webhooksmock.StoreMock{}
		d := buildDispatcherForTest(t, store)

		// A subscription to an event that does not exist produces an endpoint that never
		// fires, and diagnosing that means noticing an absence. Refused at registration.
		_, err := d.Register(t.Context(), &Registration{
			ID:         exampleWebhookID,
			AccountID:  exampleAccountID,
			URL:        "https://example.com/hook",
			EventTypes: []string{"reciped_created"},
		})
		require.Error(t, err)
		require.ErrorIs(t, err, webhooks.ErrUnknownEventType)
		assert.Empty(t, store.SaveEndpointCalls())
	})

	T.Run("with no event types", func(t *testing.T) {
		t.Parallel()

		var saved *webhooks.Endpoint

		store := &webhooksmock.StoreMock{
			SaveEndpointFunc: func(_ context.Context, endpoint *webhooks.Endpoint) error {
				saved = endpoint

				return nil
			},
		}

		d := buildDispatcherForTest(t, store)

		// Registered and subscribed to nothing is a real state, not an error: it is where a
		// webhook adopted from before delivery worked starts, and where unsubscribing from
		// the last event type returns to. It is inert, because fan-out reads subscriptions.
		// "A webhook is created subscribed to something" is a request-validation rule.
		_, err := d.Register(t.Context(), &Registration{
			ID:        exampleWebhookID,
			AccountID: exampleAccountID,
			URL:       "https://example.com/hook",
		})
		require.NoError(t, err)
		require.NotNil(t, saved)
		assert.Empty(t, saved.Events)
	})

	T.Run("with a URL that is not publicly routable", func(t *testing.T) {
		t.Parallel()

		store := &webhooksmock.StoreMock{}
		d := buildDispatcherForTest(t, store)

		// The SSRF check runs before the endpoint is stored, so the rejection reaches
		// whoever submitted the URL rather than a log line days later.
		_, err := d.Register(t.Context(), &Registration{
			ID:         exampleWebhookID,
			AccountID:  exampleAccountID,
			URL:        "https://127.0.0.1/hook",
			EventTypes: []string{exampleEventType},
		})
		require.Error(t, err)
		assert.Empty(t, store.SaveEndpointCalls())
	})

	T.Run("with a plaintext URL", func(t *testing.T) {
		t.Parallel()

		store := &webhooksmock.StoreMock{}
		d := buildDispatcherForTest(t, store)

		_, err := d.Register(t.Context(), &Registration{
			ID:         exampleWebhookID,
			AccountID:  exampleAccountID,
			URL:        "http://example.com/hook",
			EventTypes: []string{exampleEventType},
		})
		require.Error(t, err)
		assert.Empty(t, store.SaveEndpointCalls())
	})

	T.Run("with nil dispatcher", func(t *testing.T) {
		t.Parallel()

		// Loud, unlike Dispatch: a user who asked for a webhook and was told it was created
		// would otherwise never learn it does not exist.
		var d *Dispatcher

		_, err := d.Register(t.Context(), &Registration{})
		assert.ErrorIs(t, err, ErrNoDispatcher)
	})
}

func TestDispatcher_SetEventTypes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		otherAccountSubscription := "acct_other:meal_plan_updated"

		var saved *webhooks.Endpoint

		store := &webhooksmock.StoreMock{
			GetEndpointFunc: func(context.Context, string) (*webhooks.Endpoint, error) {
				return &webhooks.Endpoint{
					ID: exampleWebhookID,
					Events: []string{
						exampleAccountID + ":" + exampleEventType,
						otherAccountSubscription,
					},
				}, nil
			},
			SaveEndpointFunc: func(_ context.Context, endpoint *webhooks.Endpoint) error {
				saved = endpoint

				return nil
			},
		}

		d := buildDispatcherForTest(t, store)

		err := d.SetEventTypes(t.Context(), exampleWebhookID, exampleAccountID, []string{"meal_plan_updated"})
		require.NoError(t, err)
		require.NotNil(t, saved)

		// This account's subscriptions are replaced; another account's survive untouched.
		// SaveEndpoint replaces the whole set, so rewriting it from the event types alone
		// would silently drop the other account's.
		assert.ElementsMatch(t, []string{
			otherAccountSubscription,
			exampleAccountID + ":meal_plan_updated",
		}, saved.Events)
	})

	T.Run("with event type outside the catalog", func(t *testing.T) {
		t.Parallel()

		store := &webhooksmock.StoreMock{}
		d := buildDispatcherForTest(t, store)

		err := d.SetEventTypes(t.Context(), exampleWebhookID, exampleAccountID, []string{"reciped_created"})
		require.ErrorIs(t, err, webhooks.ErrUnknownEventType)
		// Refused before the read, so a bad request costs no query.
		assert.Empty(t, store.GetEndpointCalls())
	})
}

func TestDispatcher_RotateSecret(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		current := []byte("the-current-signing-secret-bytes")

		var saved *webhooks.Endpoint

		store := &webhooksmock.StoreMock{
			GetEndpointFunc: func(context.Context, string) (*webhooks.Endpoint, error) {
				return &webhooks.Endpoint{ID: exampleWebhookID, Secret: webhooks.Secret{Current: current}}, nil
			},
			SaveEndpointFunc: func(_ context.Context, endpoint *webhooks.Endpoint) error {
				saved = endpoint

				return nil
			},
		}

		d := buildDispatcherForTest(t, store)

		secret, err := d.RotateSecret(t.Context(), exampleWebhookID, nil)
		require.NoError(t, err)
		require.NotNil(t, saved)

		decoded, err := hex.DecodeString(secret)
		require.NoError(t, err)

		assert.Equal(t, decoded, saved.Secret.Current)
		// The outgoing key is retained so deliveries are signed under both and a subscriber
		// can accept either while it switches. Dropping it here would break every subscriber
		// at the instant of rotation, which is the failure per-endpoint secrets exist to
		// avoid.
		assert.Equal(t, current, saved.Secret.Previous)
		assert.NotEqual(t, current, saved.Secret.Current)
	})
}

func TestQualification(T *testing.T) {
	T.Parallel()

	T.Run("round trips", func(t *testing.T) {
		t.Parallel()

		subscription := qualify(exampleAccountID, exampleEventType)
		assert.Equal(t, exampleAccountID+":"+exampleEventType, subscription)

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
