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

func TestNewSideEffect(T *testing.T) {
	T.Parallel()

	T.Run("derives an event keyed on the document", func(t *testing.T) {
		t.Parallel()

		effect, err := NewSideEffect()
		require.NoError(t, err)

		r := ruleFor(t, "valid_instrument_created")

		derived, err := effect(t.Context(), nil, []outbox.Message{
			message("valid_instrument_created", map[string]any{r.IDKey: "instrument_123"}),
		})
		require.NoError(t, err)
		require.Len(t, derived, 1)

		// The topic is the index: platform says which index an event belongs to by where it
		// arrived.
		assert.Equal(t, r.Topic, derived[0].Topic)

		// The key is the document ID, which is what buys per-document ordering — at most one
		// event per document in flight, however many relays are running.
		assert.Equal(t, "instrument_123", derived[0].Key)

		event, ok := derived[0].Payload.(searchsync.Event)
		require.True(t, ok)
		assert.Equal(t, "instrument_123", event.DocumentID)
		assert.Equal(t, searchsync.OpUpsert, event.Op)
	})

	T.Run("derives one event per message", func(t *testing.T) {
		t.Parallel()

		effect, err := NewSideEffect()
		require.NoError(t, err)

		instrument := ruleFor(t, "valid_instrument_created")
		vessel := ruleFor(t, "valid_vessel_created")

		derived, err := effect(t.Context(), nil, []outbox.Message{
			message("valid_instrument_created", map[string]any{instrument.IDKey: "a"}),
			message("valid_vessel_created", map[string]any{vessel.IDKey: "b"}),
		})
		require.NoError(t, err)
		assert.Len(t, derived, 2)
	})

	T.Run("archiving an indexed entity deletes its document", func(t *testing.T) {
		t.Parallel()

		effect, err := NewSideEffect()
		require.NoError(t, err)

		r := ruleFor(t, "valid_instrument_archived")

		derived, err := effect(t.Context(), nil, []outbox.Message{
			message("valid_instrument_archived", map[string]any{r.IDKey: "gone"}),
		})
		require.NoError(t, err)
		require.Len(t, derived, 1)

		event, ok := derived[0].Payload.(searchsync.Event)
		require.True(t, ok)
		assert.Equal(t, searchsync.OpDelete, event.Op)
	})

	T.Run("archiving a sub-entity reindexes its parent instead", func(t *testing.T) {
		t.Parallel()

		effect, err := NewSideEffect()
		require.NoError(t, err)

		// An archived recipe step leaves the recipe indexed and changes what it says, so this
		// is an upsert of the recipe rather than a delete of anything.
		step := ruleFor(t, "recipe_step_archived")
		recipe := ruleFor(t, "recipe_updated")

		// The row reads the parent recipe's ID, which is the same key a recipe's own events use.
		require.Equal(t, recipe.IDKey, step.IDKey)

		derived, err := effect(t.Context(), nil, []outbox.Message{
			message("recipe_step_archived", map[string]any{step.IDKey: "recipe_1"}),
		})
		require.NoError(t, err)
		require.Len(t, derived, 1)

		assert.Equal(t, "recipe_1", derived[0].Key)

		event, ok := derived[0].Payload.(searchsync.Event)
		require.True(t, ok)
		assert.Equal(t, searchsync.OpUpsert, event.Op)
		assert.Equal(t, "recipe_1", event.DocumentID)
	})

	T.Run("a user write feeds the users index under platform's event name", func(t *testing.T) {
		t.Parallel()

		effect, err := NewSideEffect()
		require.NoError(t, err)

		registered := ruleFor(t, identity.EventUserRegistered.String())
		archived := ruleFor(t, identity.EventUserArchived.String())
		require.Equal(t, registered.IDKey, archived.IDKey)

		derived, err := effect(t.Context(), nil, []outbox.Message{
			message(identity.EventUserRegistered.String(), map[string]any{registered.IDKey: "user_1"}),
			message(identity.EventUserArchived.String(), map[string]any{archived.IDKey: "user_1"}),
		})
		require.NoError(t, err)
		require.Len(t, derived, 2)

		assert.Equal(t, "users", derived[0].Topic)
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
			{Topic: "recipes", Payload: searchsync.NewEvent(searchsync.OpUpsert, "already_an_index_event")},
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
			message("valid_instrument_created", map[string]any{"some_other_key": "instrument_123"}),
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

	T.Run("the table is one platform accepts", func(t *testing.T) {
		t.Parallel()

		// platform refuses a duplicate rule and an incomplete one at construction, so this is
		// the table's validity in one call.
		_, err := NewSideEffect()
		require.NoError(t, err)
	})

	T.Run("the index-only trigger is not a published event type", func(t *testing.T) {
		t.Parallel()

		// It stands in for an event that deliberately does not exist. If it ever collides with
		// a real event type, every write of that type would derive a recipe reindex.
		r := ruleFor(t, RecipeStepCreatedIndexTrigger)
		assert.Equal(t, "recipes", r.Topic)
		assert.Contains(t, RecipeStepCreatedIndexTrigger, ".index_only")
	})

	T.Run("Rules hands out a copy", func(t *testing.T) {
		t.Parallel()

		first := Rules()
		first[0].Topic = "mutated"

		assert.NotEqual(t, "mutated", Rules()[0].Topic)
	})
}
