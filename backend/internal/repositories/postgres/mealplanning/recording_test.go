package mealplanning

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	webhookfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexevents"
	"github.com/primandproper/dinnerdonebetter/backend/internal/recordingspine/recordingspinetest"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	auditmock "github.com/primandproper/platform-go/v15/audit/mock"
	"github.com/primandproper/platform-go/v15/outbox"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	searchsync "github.com/primandproper/platform-go/v15/searchsync"
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

// recordingHarness is a repository over mocks of platform's two sinks: the outbox enqueuer the
// Emitter writes through and the audit recorder the Recorder writes through. The dispatcher's
// catalog holds the one event type the harness emits and dispatches into a slice, so a test reads
// the scope the repository chose rather than whether it chose any.
type recordingHarness struct {
	repo       *repository
	eventType  string
	enqueued   []outbox.Message
	dispatched []*webhooks.Delivery
	scopes     []tenancy.Scope
	recorded   []*platformaudit.Entry
	chains     []tenancy.Scope
}

func buildRecordingHarness(t *testing.T) *recordingHarness {
	t.Helper()

	h := &recordingHarness{eventType: webhookfakes.BuildFakeWebhookEventType()}

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

	emitter, err := webhooks.NewEmitter(enqueuer, dispatcher, testDataChangesTopic)
	require.NoError(t, err)

	platformRecorder, err := platformrecording.New(recorder, emitter, sessions.PrincipalFromContext)
	require.NoError(t, err)

	h.repo = &repository{logger: loggingnoop.NewLogger(), emitter: emitter, recorder: platformRecorder}

	return h
}

func txForRecordingTest() database.Tx {
	return database.NewTxForTesting(&mockdatabase.SQLQueryExecutorMock{})
}

// payload decodes the one enqueued message as this application's event, the way the broker
// consumer reads it: platform's envelope names the event, and this application's message is its
// payload.
func (h *recordingHarness) payload(t *testing.T) *datachanges.Message {
	t.Helper()

	require.Len(t, h.enqueued, 1)

	raw, err := json.Marshal(h.enqueued[0].Payload)
	require.NoError(t, err)

	msg := &datachanges.Message{}
	eventType, matched, err := webhooks.Decode(raw, msg, webhooks.EventType(h.eventType))
	require.NoError(t, err)
	require.True(t, matched, "the envelope named %q rather than the harness's event", eventType)

	return msg
}

func TestRepository_emit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := buildRecordingHarness(t)
		accountID := fake.BuildFakeID()
		value := fake.BuildFakeString()

		err := h.repo.emit(t.Context(), txForRecordingTest(), loggingnoop.NewLogger(), h.eventType, accountID, map[string]any{"k": value})
		require.NoError(t, err)

		msg := h.payload(t)
		assert.Equal(t, h.eventType, msg.EventType)
		assert.Equal(t, accountID, msg.AccountID)
		assert.Equal(t, value, msg.Context["k"])

		// The account is the scope the event fans out in, and — by platform's default — the
		// ordering key.
		require.Len(t, h.scopes, 1)
		assert.Equal(t, tenancy.Of(accountID), h.scopes[0])
		assert.Equal(t, accountID, h.enqueued[0].Key)
		assert.Equal(t, webhooks.EventType(h.eventType), h.dispatched[0].EventType)
	})

	T.Run("attributes the event to the session when the write names no account", func(t *testing.T) {
		t.Parallel()

		h := buildRecordingHarness(t)
		userID, accountID := fake.BuildFakeID(), fake.BuildFakeID()

		ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: userID},
			ActiveAccountID: accountID,
		})

		require.NoError(t, h.repo.emit(ctx, txForRecordingTest(), loggingnoop.NewLogger(), h.eventType, "", nil))

		msg := h.payload(t)
		assert.Equal(t, userID, msg.UserID)
		assert.Equal(t, accountID, msg.AccountID)
		require.Len(t, h.scopes, 1)
		assert.Equal(t, tenancy.Of(accountID), h.scopes[0])
	})

	T.Run("with no account at all", func(t *testing.T) {
		t.Parallel()

		// Background jobs emit these. They are published in the global scope, where no
		// endpoint lives, so they reach the broker and no subscriber — but the dispatcher is
		// asked, in that scope, rather than skipped.
		h := buildRecordingHarness(t)

		require.NoError(t, h.repo.emit(t.Context(), txForRecordingTest(), loggingnoop.NewLogger(), h.eventType, "", nil))

		require.Len(t, h.enqueued, 1)
		assert.Empty(t, h.enqueued[0].Key)
		require.Len(t, h.scopes, 1)
		assert.True(t, h.scopes[0].IsGlobal())
	})

	T.Run("with an event type outside the catalog", func(t *testing.T) {
		t.Parallel()

		// Published and not dispatched: platform's gate, not a pre-check here.
		h := buildRecordingHarness(t)

		err := h.repo.emit(t.Context(), txForRecordingTest(), loggingnoop.NewLogger(), fake.BuildFakeID(), fake.BuildFakeID(), nil)
		require.NoError(t, err)

		assert.Len(t, h.enqueued, 1)
		assert.Empty(t, h.dispatched)
	})
}

func TestRepository_record(T *testing.T) {
	T.Parallel()

	T.Run("files the entries on the event's account chain, and publishes beside them", func(t *testing.T) {
		t.Parallel()

		h := buildRecordingHarness(t)
		requesterID, accountID := fake.BuildFakeID(), fake.BuildFakeID()
		first := auditEntry(resourceTypeMealPlanEvents, fake.BuildFakeID(), platformaudit.EventUpdated)
		second := auditEntry(resourceTypeMealPlanEvents, fake.BuildFakeID(), platformaudit.EventUpdated)

		ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: requesterID},
			ActiveAccountID: accountID,
		})

		require.NoError(t, h.repo.record(ctx, txForRecordingTest(), loggingnoop.NewLogger(), h.eventType, accountID, nil, first, second))

		// One chain for both entries, the account's.
		require.Len(t, h.chains, 1)
		assert.Equal(t, tenancy.Of(accountID), h.chains[0])
		require.Len(t, h.recorded, 2)
		assert.Equal(t, first.ResourceType, h.recorded[0].ResourceType)
		assert.Equal(t, first.ResourceID, h.recorded[0].ResourceID)
		assert.Equal(t, platformaudit.EventUpdated, h.recorded[0].EventType)
		assert.Equal(t, second.ResourceID, h.recorded[1].ResourceID)

		// Who did it is the principal on the context.
		assert.Equal(t, requesterID, h.recorded[0].Actor.ID)
		assert.Empty(t, h.recorded[0].Actor.Impersonator)

		// And the event fans out in the scope the entries were filed under.
		assert.Equal(t, accountID, h.payload(t).AccountID)
		require.Len(t, h.scopes, 1)
		assert.Equal(t, tenancy.Of(accountID), h.scopes[0])
	})

	T.Run("takes the account from the session when the write names none", func(t *testing.T) {
		t.Parallel()

		// Meal plan events are written knowing only their meal plan. Their entries belong to the
		// account whose subscribers hear about them, not to the global chain no account reads.
		h := buildRecordingHarness(t)
		accountID := fake.BuildFakeID()

		ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: fake.BuildFakeID()},
			ActiveAccountID: accountID,
		})

		require.NoError(t, h.repo.record(ctx, txForRecordingTest(), loggingnoop.NewLogger(), h.eventType, "", nil,
			auditEntry(resourceTypeMealPlanEvents, fake.BuildFakeID(), platformaudit.EventCreated)))

		require.Len(t, h.chains, 1)
		assert.Equal(t, tenancy.Of(accountID), h.chains[0])
		require.Len(t, h.scopes, 1)
		assert.Equal(t, tenancy.Of(accountID), h.scopes[0])
	})

	T.Run("files the actor's own chain when the write happened in no account", func(t *testing.T) {
		t.Parallel()

		h := buildRecordingHarness(t)
		userID := fake.BuildFakeID()

		ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: userID},
		})

		require.NoError(t, h.repo.record(ctx, txForRecordingTest(), loggingnoop.NewLogger(), h.eventType, "", nil,
			auditEntry(resourceTypeMealPlans, fake.BuildFakeID(), platformaudit.EventCreated)))

		require.Len(t, h.chains, 1)
		assert.Equal(t, tenancy.Of(userID), h.chains[0])
	})

	T.Run("names the operator on an impersonated write", func(t *testing.T) {
		t.Parallel()

		// The entry stays the subject's — it is their data — and the operator is the
		// second slot that stops it saying they did it.
		h := buildRecordingHarness(t)
		subjectID, operatorID, accountID := fake.BuildFakeID(), fake.BuildFakeID(), fake.BuildFakeID()

		ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: subjectID},
			ActiveAccountID: accountID,
			ImpersonatorID:  operatorID,
		})

		require.NoError(t, h.repo.record(ctx, txForRecordingTest(), loggingnoop.NewLogger(), h.eventType, accountID, nil,
			auditEntry(resourceTypeMealPlans, fake.BuildFakeID(), platformaudit.EventUpdated)))

		require.Len(t, h.recorded, 1)
		assert.Equal(t, subjectID, h.recorded[0].Actor.ID)
		assert.Equal(t, operatorID, h.recorded[0].Actor.Impersonator)
	})

	T.Run("records an unattributed write as such", func(t *testing.T) {
		t.Parallel()

		// A background job: nobody on the context, and no account on the write.
		h := buildRecordingHarness(t)

		require.NoError(t, h.repo.record(t.Context(), txForRecordingTest(), loggingnoop.NewLogger(), h.eventType, "", nil,
			auditEntry(resourceTypeMealPlans, fake.BuildFakeID(), platformaudit.EventArchived)))

		require.Len(t, h.recorded, 1)
		// Recorded under the named absence, never with an empty actor, which audit refuses.
		assert.Equal(t, platformaudit.ActorUnattributed, h.recorded[0].Actor.ID)
		assert.True(t, h.chains[0].IsGlobal())
	})

	T.Run("with a nil entry", func(t *testing.T) {
		t.Parallel()

		h := buildRecordingHarness(t)

		err := h.repo.record(t.Context(), txForRecordingTest(), loggingnoop.NewLogger(), h.eventType, "", nil, nil)
		require.ErrorIs(t, err, platformrecording.ErrNilEntry)

		assert.Empty(t, h.recorded)
		assert.Empty(t, h.enqueued)
	})
}

func TestRepository_withRecord(T *testing.T) {
	T.Parallel()

	T.Run("records the entry and the event after the write", func(t *testing.T) {
		t.Parallel()

		h := buildRecordingHarness(t)
		h.repo.Client = &mockdatabase.ClientMock{
			WithTransactionFunc: func(_ context.Context, fn func(database.Tx) error) error {
				return fn(txForRecordingTest())
			},
		}
		accountID, resourceID := fake.BuildFakeID(), fake.BuildFakeID()

		var wrote bool
		require.NoError(t, h.repo.withRecord(t.Context(), loggingnoop.NewLogger(), auditEntry(resourceTypeMealPlans, resourceID, platformaudit.EventCreated), h.eventType, accountID, nil, func(database.Tx) error {
			wrote = true
			return nil
		}))

		assert.True(t, wrote)
		require.Len(t, h.recorded, 1)
		assert.Equal(t, resourceID, h.recorded[0].ResourceID)
		assert.Equal(t, tenancy.Of(accountID), h.chains[0])
		assert.Equal(t, accountID, h.payload(t).AccountID)
	})

	T.Run("records nothing when the write fails", func(t *testing.T) {
		t.Parallel()

		h := buildRecordingHarness(t)
		h.repo.Client = &mockdatabase.ClientMock{
			WithTransactionFunc: func(_ context.Context, fn func(database.Tx) error) error {
				return fn(txForRecordingTest())
			},
		}

		err := h.repo.withRecord(t.Context(), loggingnoop.NewLogger(), auditEntry(resourceTypeMealPlans, fake.BuildFakeID(), platformaudit.EventArchived), h.eventType, "", nil, func(database.Tx) error {
			return sql.ErrNoRows
		})
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, h.recorded)
		assert.Empty(t, h.enqueued)
	})
}

func TestRepository_emitIndex(T *testing.T) {
	T.Parallel()

	T.Run("derives the index event and announces nothing", func(t *testing.T) {
		t.Parallel()

		var (
			topics []string
			events []searchsync.Event
		)

		executor := &mockdatabase.SQLQueryExecutorMock{
			ExecContextFunc: func(_ context.Context, _ string, args ...any) (sql.Result, error) {
				for _, arg := range args {
					switch v := arg.(type) {
					case string:
						if v == mealplanningindexing.IndexTypeRecipes {
							topics = append(topics, v)
						}
					case []byte:
						var event searchsync.Event
						if json.Unmarshal(v, &event) == nil && event.DocumentID != "" {
							events = append(events, event)
						}
					}
				}

				return nil, nil
			},
		}

		writer, err := outbox.NewWriter(dialect.Postgres, recordingspinetest.IndexRules(t))
		require.NoError(t, err)

		repo := &repository{writer: writer}
		recipeID := fake.BuildFakeID()

		require.NoError(t, repo.emitIndex(t.Context(), database.NewTxForTesting(executor), indexevents.RecipeStepCreatedIndexTrigger, map[string]any{
			mealplanningkeys.RecipeIDKey: recipeID,
		}))

		// One row, on the recipes index's topic, naming the recipe — and nothing else, because
		// the trigger is not an event anybody subscribes to.
		require.Len(t, executor.ExecContextCalls(), 1)
		assert.Equal(t, []string{mealplanningindexing.IndexTypeRecipes}, topics)
		require.Len(t, events, 1)
		assert.Equal(t, recipeID, events[0].DocumentID)
		assert.Equal(t, searchsync.OpUpsert, events[0].Op)
	})
}
