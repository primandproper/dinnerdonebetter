package indexevents

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"

	"github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/platform-go/v15/searchsync"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func message(eventType string, context map[string]any) outbox.Message {
	return outbox.Message{Topic: "data_changes", Payload: &datachanges.Message{EventType: eventType, Context: context}}
}

// ruleFor is the one row an event type has. The table has no event type feeding two indexes
// today; a test that needs the second row reads RulesFor directly.
func ruleFor(t *testing.T, eventType string) searchsync.Rule {
	t.Helper()

	matched := RulesFor(eventType)
	require.Len(t, matched, 1, eventType)

	return matched[0]
}

// The tests here drive the merged table through platform's side effect on identity's rows,
// which this package names rather than a domain's: what a domain's rows say is that domain's
// to test, beside the rows. What is asserted here is the merge and the registration.
func TestNewSideEffect(T *testing.T) {
	T.Parallel()

	T.Run("derives an event keyed on the document", func(t *testing.T) {
		t.Parallel()

		effect, err := NewSideEffect()
		require.NoError(t, err)

		r := ruleFor(t, identity.EventUserRegistered.String())

		derived, err := effect(t.Context(), nil, []outbox.Message{
			message(identity.EventUserRegistered.String(), map[string]any{r.IDKey: "user_123"}),
		})
		require.NoError(t, err)
		require.Len(t, derived, 1)

		// The topic is the index: platform says which index an event belongs to by where it
		// arrived.
		assert.Equal(t, r.Topic, derived[0].Topic)

		// The key is the document ID, which is what buys per-document ordering — at most one
		// event per document in flight, however many relays are running.
		assert.Equal(t, "user_123", derived[0].Key)

		event, ok := derived[0].Payload.(searchsync.Event)
		require.True(t, ok)
		assert.Equal(t, "user_123", event.DocumentID)
		assert.Equal(t, searchsync.OpUpsert, event.Op)
	})

	T.Run("derives one event per message", func(t *testing.T) {
		t.Parallel()

		effect, err := NewSideEffect()
		require.NoError(t, err)

		registered := ruleFor(t, identity.EventUserRegistered.String())
		archived := ruleFor(t, identity.EventUserArchived.String())

		derived, err := effect(t.Context(), nil, []outbox.Message{
			message(identity.EventUserRegistered.String(), map[string]any{registered.IDKey: "a"}),
			message(identity.EventUserArchived.String(), map[string]any{archived.IDKey: "b"}),
		})
		require.NoError(t, err)
		require.Len(t, derived, 2)

		assert.Equal(t, searchsync.OpUpsert, derived[0].Payload.(searchsync.Event).Op)
		assert.Equal(t, searchsync.OpDelete, derived[1].Payload.(searchsync.Event).Op)
	})

	T.Run("an untabled event type derives nothing", func(t *testing.T) {
		t.Parallel()

		effect, err := NewSideEffect()
		require.NoError(t, err)

		derived, err := effect(t.Context(), nil, []outbox.Message{
			message("webhook_created", map[string]any{"webhook_id": "wh_1"}),
		})
		require.NoError(t, err)
		assert.Empty(t, derived)
	})

	T.Run("a payload that is not a data change message derives nothing", func(t *testing.T) {
		t.Parallel()

		effect, err := NewSideEffect()
		require.NoError(t, err)

		derived, err := effect(t.Context(), nil, []outbox.Message{
			{Topic: "users", Payload: searchsync.NewEvent(searchsync.OpUpsert, "already_an_index_event")},
		})
		require.NoError(t, err)
		assert.Empty(t, derived)
	})

	T.Run("with no document ID under the key the table names", func(t *testing.T) {
		t.Parallel()

		effect, err := NewSideEffect()
		require.NoError(t, err)

		// Refused rather than skipped. Skipping would put back the failure this package
		// removes: a write commits, the index never hears about it, and nothing says so.
		_, err = effect(t.Context(), nil, []outbox.Message{
			message(identity.EventUserRegistered.String(), map[string]any{"some_other_key": "user_123"}),
		})
		require.ErrorIs(t, err, searchsync.ErrMissingDocumentID)
	})

	T.Run("with no messages", func(t *testing.T) {
		t.Parallel()

		effect, err := NewSideEffect()
		require.NoError(t, err)

		derived, err := effect(t.Context(), nil, nil)
		require.NoError(t, err)
		assert.Empty(t, derived)
	})
}

func TestRules(T *testing.T) {
	T.Parallel()

	T.Run("merges every domain's table", func(t *testing.T) {
		t.Parallel()

		expected := 0
		for _, table := range tables() {
			expected += len(table)
		}

		require.NotZero(t, expected, "no domain contributed a table")
		assert.Len(t, Rules(), expected)
	})

	T.Run("every row names an index and a key", func(t *testing.T) {
		t.Parallel()

		for _, eventType := range EventTypes() {
			for _, r := range RulesFor(eventType) {
				assert.NotEmpty(t, r.Topic, eventType)
				assert.NotEmpty(t, r.IDKey, eventType)
				assert.Contains(t, []searchsync.Op{searchsync.OpUpsert, searchsync.OpDelete}, r.Op, eventType)
			}
		}
	})

	T.Run("the merged table is one platform accepts", func(t *testing.T) {
		t.Parallel()

		// platform refuses a duplicate rule and an incomplete one at construction, so this is
		// the merged table's validity in one call — including that no two domains tabled the
		// same event type for the same index.
		_, err := NewSideEffect()
		require.NoError(t, err)
	})

	T.Run("Rules hands out a copy", func(t *testing.T) {
		t.Parallel()

		first := Rules()
		first[0].Topic = "mutated"

		assert.NotEqual(t, "mutated", Rules()[0].Topic)
	})
}
