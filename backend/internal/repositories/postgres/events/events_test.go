package events

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexevents"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	auditmock "github.com/primandproper/platform-go/v15/audit/mock"
	"github.com/primandproper/platform-go/v15/outbox"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/platform-go/v15/webhooks"
	webhooksmock "github.com/primandproper/platform-go/v15/webhooks/mock"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/fake"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// harness is an Emitter over mocks of the two platform sinks: the outbox enqueuer and the audit
// recorder. The dispatcher's catalog holds the one event type the harness emits and dispatches
// into a slice, so a test reads the scope and key the Emitter chose rather than whether it chose
// any.
type harness struct {
	emitter    *Emitter
	eventType  string
	enqueued   []outbox.Message
	dispatched []*webhooks.Delivery
	scopes     []tenancy.Scope
	recorded   []*platformaudit.Entry
	chains     []tenancy.Scope
}

func buildHarness(t *testing.T) *harness {
	t.Helper()

	h := &harness{eventType: fakes.BuildFakeWebhookEventType()}

	enqueuer := &webhooksmock.EnqueuerMock{
		EnqueueFunc: func(_ context.Context, _ database.Tx, msgs ...outbox.Message) error {
			h.enqueued = append(h.enqueued, msgs...)
			return nil
		},
	}

	dispatcher := &webhooksmock.DispatcherMock{
		CatalogFunc: func() webhooks.Catalog {
			return webhooks.Catalog{webhooks.EventType(h.eventType): {}}
		},
		DispatchFunc: func(_ context.Context, _ database.Tx, scope tenancy.Scope, delivery *webhooks.Delivery) error {
			h.dispatched = append(h.dispatched, delivery)
			h.scopes = append(h.scopes, scope)
			return nil
		},
	}

	recorder := &auditmock.RecorderMock{
		RecordFunc: func(_ context.Context, _ database.Tx, scope tenancy.Scope, entries ...*platformaudit.Entry) error {
			h.recorded = append(h.recorded, entries...)
			h.chains = append(h.chains, scope)
			return nil
		},
	}

	emitter, err := webhooks.NewEmitter(enqueuer, dispatcher, "data_changes")
	require.NoError(t, err)

	platformRecorder, err := platformrecording.New(recorder, emitter, sessions.PrincipalFromContext)
	require.NoError(t, err)

	effect, err := indexevents.NewSideEffect()
	require.NoError(t, err)

	writer, err := outbox.NewWriter(dialect.Postgres, outbox.WithWriterSideEffect(indexevents.SideEffectName, effect))
	require.NoError(t, err)

	h.emitter, err = NewEmitter(emitter, platformRecorder, writer)
	require.NoError(t, err)

	return h
}

func txForTest() database.Tx {
	return database.NewTxForTesting(&mockdatabase.SQLQueryExecutorMock{})
}

// payload decodes the one enqueued message as this application's event.
func (h *harness) payload(t *testing.T) *datachanges.Message {
	t.Helper()

	require.Len(t, h.enqueued, 1)

	// Rendered the way the outbox renders it, then read the way the broker consumer reads it:
	// platform wraps the payload in an envelope naming the event, and Decode is the consumer's
	// half of that.
	raw, err := json.Marshal(h.enqueued[0].Payload)
	require.NoError(t, err)

	msg := &datachanges.Message{}
	eventType, matched, err := webhooks.Decode(raw, msg, webhooks.EventType(h.eventType))
	require.NoError(t, err)
	require.True(t, matched, "the envelope named %q rather than the harness's event", eventType)

	return msg
}

func TestEmitter_Emit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildHarness(t)
		accountID := fake.BuildFakeID()

		err := h.emitter.Emit(t.Context(), txForTest(), loggingnoop.NewLogger(), h.eventType, accountID, map[string]any{"k": "v"})
		require.NoError(t, err)

		// The payload is this application's message, inside platform's envelope: the writer's
		// side effects read it by type through the envelope's delegation, and the consumer
		// decodes it out of the envelope.
		msg := h.payload(t)
		assert.Equal(t, h.eventType, msg.EventType)
		assert.Equal(t, accountID, msg.AccountID)
		assert.Equal(t, "v", msg.Context["k"])

		// The account is the scope the event fans out in, and — by platform's default — the
		// ordering key.
		require.Len(t, h.scopes, 1)
		assert.Equal(t, tenancy.Of(accountID), h.scopes[0])
		assert.Equal(t, accountID, h.enqueued[0].Key)
		assert.Equal(t, webhooks.EventType(h.eventType), h.dispatched[0].EventType)
	})

	T.Run("with an ordering key", func(t *testing.T) {
		t.Parallel()

		h := buildHarness(t)

		err := h.emitter.Emit(t.Context(), txForTest(), loggingnoop.NewLogger(), h.eventType, fake.BuildFakeID(), nil, WithOrderingKey("resource_1"))
		require.NoError(t, err)

		require.Len(t, h.enqueued, 1)
		require.Len(t, h.dispatched, 1)
		assert.Equal(t, "resource_1", h.enqueued[0].Key)
		assert.Equal(t, "resource_1", h.dispatched[0].OrderingKey)
	})

	T.Run("with no account", func(t *testing.T) {
		t.Parallel()

		// Background jobs emit these. They are published in the global scope, where no
		// endpoint lives, so they reach the broker and no subscriber — but the dispatcher is
		// asked, in that scope, rather than skipped.
		h := buildHarness(t)

		err := h.emitter.Emit(t.Context(), txForTest(), loggingnoop.NewLogger(), h.eventType, "", nil)
		require.NoError(t, err)

		require.Len(t, h.enqueued, 1)
		assert.Empty(t, h.enqueued[0].Key)
		require.Len(t, h.scopes, 1)
		assert.True(t, h.scopes[0].IsGlobal())
	})

	T.Run("with an event type outside the catalog", func(t *testing.T) {
		t.Parallel()

		// Published and not dispatched: platform's gate, not a pre-check here.
		h := buildHarness(t)

		err := h.emitter.Emit(t.Context(), txForTest(), loggingnoop.NewLogger(), "reciped_created", fake.BuildFakeID(), nil)
		require.NoError(t, err)

		assert.Len(t, h.enqueued, 1)
		assert.Empty(t, h.dispatched)
	})

	T.Run("with a user named for a write made before any session", func(t *testing.T) {
		t.Parallel()

		h := buildHarness(t)
		userID := fake.BuildFakeID()

		err := h.emitter.Emit(t.Context(), txForTest(), loggingnoop.NewLogger(), h.eventType, "", nil, WithUserID(userID))
		require.NoError(t, err)

		assert.Equal(t, userID, h.payload(t).UserID)
	})
}

func TestEmitter_Record(T *testing.T) {
	T.Parallel()

	T.Run("files the entry on the chain the attribution rule chose, and publishes beside it", func(t *testing.T) {
		t.Parallel()

		h := buildHarness(t)
		accountID := fake.BuildFakeID()
		entry := audit.NewEntry("", accountID, "meal_plans", fake.BuildFakeID(), platformaudit.EventCreated)
		entry.Metadata = map[string]string{"why": "because"}

		ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: "requester"},
			ActiveAccountID: accountID,
		})

		err := h.emitter.Record(ctx, txForTest(), loggingnoop.NewLogger(), entry, h.eventType, accountID, nil)
		require.NoError(t, err)

		require.Len(t, h.recorded, 1)
		assert.Equal(t, tenancy.Of(accountID), h.chains[0])
		assert.Equal(t, entry.ResourceType, h.recorded[0].ResourceType)
		assert.Equal(t, entry.ResourceID, h.recorded[0].ResourceID)
		assert.Equal(t, platformaudit.EventCreated, h.recorded[0].EventType)
		assert.Equal(t, "because", h.recorded[0].Metadata["why"])

		// Who did it is the principal on the context, not whoever the entry was built for.
		assert.Equal(t, "requester", h.recorded[0].Actor.ID)

		assert.Equal(t, accountID, h.payload(t).AccountID)
		require.Len(t, h.scopes, 1)
		assert.Equal(t, tenancy.Of(accountID), h.scopes[0])
	})

	T.Run("names the entry's user as the actor when the context carries nobody", func(t *testing.T) {
		t.Parallel()

		// A sign-in: the write that establishes who is acting, so the request has no principal
		// yet and the entry says who.
		h := buildHarness(t)
		userID := fake.BuildFakeID()
		entry := audit.NewEntry(userID, "", "users", userID, platformaudit.EventOther)

		err := h.emitter.Record(t.Context(), txForTest(), loggingnoop.NewLogger(), entry, h.eventType, "", nil, WithUserID(userID))
		require.NoError(t, err)

		require.Len(t, h.recorded, 1)
		assert.Equal(t, userID, h.recorded[0].Actor.ID)
		assert.Equal(t, platformaudit.ActorUser, h.recorded[0].Actor.Type)
		assert.Equal(t, tenancy.Of(userID), h.chains[0])
		assert.Equal(t, userID, h.payload(t).UserID)
	})

	T.Run("records an unattributed write as such", func(t *testing.T) {
		t.Parallel()

		h := buildHarness(t)
		entry := audit.NewEntry("", "", "valid_instruments", fake.BuildFakeID(), platformaudit.EventCreated)

		err := h.emitter.Record(t.Context(), txForTest(), loggingnoop.NewLogger(), entry, h.eventType, "", nil)
		require.NoError(t, err)

		require.Len(t, h.recorded, 1)
		// Recorded under the named absence, never with an empty actor, which audit refuses.
		assert.Equal(t, platformaudit.ActorUnattributed, h.recorded[0].Actor.ID)
		assert.True(t, h.chains[0].IsGlobal())
	})

	T.Run("records an entry and announces nothing when the write names no event", func(t *testing.T) {
		t.Parallel()

		// An account's membership row on registration: the log is owed the entry, and the
		// registration event already announces the whole write.
		h := buildHarness(t)
		accountID := fake.BuildFakeID()
		entry := audit.NewEntry("", accountID, "account_user_memberships", fake.BuildFakeID(), platformaudit.EventCreated)

		err := h.emitter.Record(t.Context(), txForTest(), loggingnoop.NewLogger(), entry, "", accountID, nil)
		require.NoError(t, err)

		require.Len(t, h.recorded, 1)
		assert.Empty(t, h.enqueued)
		assert.Empty(t, h.dispatched)
	})

	T.Run("with a nil entry", func(t *testing.T) {
		t.Parallel()

		h := buildHarness(t)

		err := h.emitter.Record(t.Context(), txForTest(), loggingnoop.NewLogger(), nil, h.eventType, "", nil)
		require.Error(t, err)
		assert.Empty(t, h.recorded)
		assert.Empty(t, h.enqueued)
	})
}

func TestNewEmitter(T *testing.T) {
	T.Parallel()

	T.Run("refuses every missing part", func(t *testing.T) {
		t.Parallel()

		// Nothing here is optional any more: the emitter this replaced was nil-inert for a
		// process with no topic, which made a process with no broker a process whose writes
		// announced nothing.
		_, err := NewEmitter(nil, nil, nil)
		require.Error(t, err)
	})
}
