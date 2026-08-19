package events

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/fakes"

	"github.com/primandproper/platform-go/v11/database"
	"github.com/primandproper/platform-go/v11/database/dialect"
	mockdatabase "github.com/primandproper/platform-go/v11/database/mock"
	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/outbox"
	"github.com/primandproper/platform-go/v11/tenancy"
	"github.com/primandproper/platform-go/v11/webhooks"
	webhooksmock "github.com/primandproper/platform-go/v11/webhooks/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildEmitterForTest builds an Emitter over a dispatcher and nothing else. The outbox half is
// exercised by the repositories that write through it; this is about the webhook half.
func buildEmitterForTest(dispatcher webhooks.Dispatcher) *Emitter {
	return &Emitter{dispatcher: dispatcher}
}

func TestEmitter_dispatchWebhooks(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		var dispatched *webhooks.Delivery

		executor := &mockdatabase.SQLQueryExecutorMock{}
		dispatcher := &webhooksmock.DispatcherMock{
			DispatchFunc: func(_ context.Context, q database.SQLQueryExecutor, delivery *webhooks.Delivery) error {
				assert.Same(t, executor, q)
				dispatched = delivery

				return nil
			},
		}

		msg := &audit.DataChangeMessage{
			AccountID: fake.BuildFakeID(),
			EventType: fakes.BuildFakeWebhookEventType(),
		}

		err := buildEmitterForTest(dispatcher).dispatchWebhooks(t.Context(), executor, msg, &emitConfig{orderingKey: msg.AccountID})
		require.NoError(t, err)
		require.NotNil(t, dispatched)

		// The account is the delivery's scope, and the event type is the plain catalog
		// string. Both together are the point: the account used to be a prefix on the event
		// type, which no scope column could index and no catalog could hold.
		assert.Equal(t, tenancy.Of(msg.AccountID), dispatched.Scope)
		assert.Equal(t, webhooks.EventType(msg.EventType), dispatched.EventType)
		assert.Equal(t, msg.AccountID, dispatched.OrderingKey)

		// The payload is the message itself, marshaled once, so a subscriber and a queue
		// consumer see byte-identical bodies.
		var payload audit.DataChangeMessage
		require.NoError(t, json.Unmarshal(dispatched.Payload, &payload))
		assert.Equal(t, msg.EventType, payload.EventType)
		assert.Equal(t, msg.AccountID, payload.AccountID)
	})

	T.Run("with an excluded event type", func(t *testing.T) {
		t.Parallel()

		// Skipped before Dispatch, not rejected by it. This runs inside the transaction that
		// wrote the row the event describes, so an error would not fail a webhook — it would
		// fail the sign-in. Dispatch is unconfigured, so calling it would panic.
		dispatcher := &webhooksmock.DispatcherMock{}

		msg := &audit.DataChangeMessage{
			AccountID: fake.BuildFakeID(),
			EventType: identity.UserLoggedInServiceEventType,
		}

		err := buildEmitterForTest(dispatcher).dispatchWebhooks(t.Context(), &mockdatabase.SQLQueryExecutorMock{}, msg, &emitConfig{})
		require.NoError(t, err)
		assert.Empty(t, dispatcher.DispatchCalls())
	})

	T.Run("with an event type outside the catalog", func(t *testing.T) {
		t.Parallel()

		dispatcher := &webhooksmock.DispatcherMock{}

		msg := &audit.DataChangeMessage{
			AccountID: fake.BuildFakeID(),
			EventType: "reciped_created",
		}

		err := buildEmitterForTest(dispatcher).dispatchWebhooks(t.Context(), &mockdatabase.SQLQueryExecutorMock{}, msg, &emitConfig{})
		require.NoError(t, err)
		assert.Empty(t, dispatcher.DispatchCalls())
	})

	T.Run("with no account", func(t *testing.T) {
		t.Parallel()

		// Background jobs emit these, and they belong to no subscriber. Reaching Dispatch
		// with one would be worse than a no-op: tenancy.Of refuses the empty identifier, so
		// the delivery would be rejected and the transaction that caused it would fail.
		dispatcher := &webhooksmock.DispatcherMock{}

		msg := &audit.DataChangeMessage{EventType: fakes.BuildFakeWebhookEventType()}

		err := buildEmitterForTest(dispatcher).dispatchWebhooks(t.Context(), &mockdatabase.SQLQueryExecutorMock{}, msg, &emitConfig{})
		require.NoError(t, err)
		assert.Empty(t, dispatcher.DispatchCalls())
	})

	T.Run("with no dispatcher", func(t *testing.T) {
		t.Parallel()

		// A process wired without webhooks still writes rows and emits events.
		msg := &audit.DataChangeMessage{
			AccountID: fake.BuildFakeID(),
			EventType: fakes.BuildFakeWebhookEventType(),
		}

		err := buildEmitterForTest(nil).dispatchWebhooks(t.Context(), &mockdatabase.SQLQueryExecutorMock{}, msg, &emitConfig{})
		require.NoError(t, err)
	})
}

func TestNewEmitter(T *testing.T) {
	T.Parallel()

	T.Run("with no writer", func(t *testing.T) {
		t.Parallel()

		// A nil Emitter is inert, which is what lets a repository be built in a process with
		// no outbox without threading a mock through every call site.
		assert.Nil(t, NewEmitter(nil, "data_changes", &webhooksmock.DispatcherMock{}, nil))
	})

	T.Run("with no topic", func(t *testing.T) {
		t.Parallel()

		writer, err := outbox.NewWriter(dialect.Postgres)
		require.NoError(t, err)

		// A process with no queues config has no topic to emit to — the MCP server, and the
		// one-shot CLI tools, register repositories but publish nothing.
		assert.Nil(t, NewEmitter(writer, "", &webhooksmock.DispatcherMock{}, nil))
	})
}
