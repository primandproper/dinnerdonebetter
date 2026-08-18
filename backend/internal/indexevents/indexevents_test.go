package indexevents

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	"github.com/primandproper/platform-go/v10/outbox"
	searchsync "github.com/primandproper/platform-go/v10/search/sync"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func message(eventType string, context map[string]any) outbox.Message {
	return outbox.Message{Topic: "data_changes", Payload: &audit.DataChangeMessage{EventType: eventType, Context: context}}
}

func TestSideEffect(T *testing.T) {
	T.Parallel()

	T.Run("derives an event keyed on the document", func(t *testing.T) {
		t.Parallel()

		spec, ok := SpecFor("valid_instrument_created")
		require.True(t, ok)

		derived, err := SideEffect(t.Context(), nil, []outbox.Message{
			message("valid_instrument_created", map[string]any{spec.IDKey: "instrument_123"}),
		})
		require.NoError(t, err)
		require.Len(t, derived, 1)

		// The topic is the index: platform says which index an event belongs to by where it
		// arrived.
		assert.Equal(t, spec.Index, derived[0].Topic)

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

		instrument, ok := SpecFor("valid_instrument_created")
		require.True(t, ok)
		vessel, ok := SpecFor("valid_vessel_created")
		require.True(t, ok)

		derived, err := SideEffect(t.Context(), nil, []outbox.Message{
			message("valid_instrument_created", map[string]any{instrument.IDKey: "a"}),
			message("valid_vessel_created", map[string]any{vessel.IDKey: "b"}),
		})
		require.NoError(t, err)
		assert.Len(t, derived, 2)
	})

	T.Run("archiving an indexed entity deletes its document", func(t *testing.T) {
		t.Parallel()

		spec, ok := SpecFor("valid_instrument_archived")
		require.True(t, ok)

		derived, err := SideEffect(t.Context(), nil, []outbox.Message{
			message("valid_instrument_archived", map[string]any{spec.IDKey: "gone"}),
		})
		require.NoError(t, err)
		require.Len(t, derived, 1)

		event, ok := derived[0].Payload.(searchsync.Event)
		require.True(t, ok)
		assert.Equal(t, searchsync.OpDelete, event.Op)
	})

	T.Run("archiving a sub-entity reindexes its parent instead", func(t *testing.T) {
		t.Parallel()

		// An archived recipe step leaves the recipe indexed and changes what it says, so this
		// is an upsert of the recipe rather than a delete of anything.
		step, ok := SpecFor("recipe_step_archived")
		require.True(t, ok)
		recipe, ok := SpecFor("recipe_updated")
		require.True(t, ok)

		// The row reads the parent recipe's ID, which is the same key a recipe's own events use.
		require.Equal(t, recipe.IDKey, step.IDKey)

		derived, err := SideEffect(t.Context(), nil, []outbox.Message{
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

	T.Run("an untabled event type derives nothing", func(t *testing.T) {
		t.Parallel()

		derived, err := SideEffect(t.Context(), nil, []outbox.Message{
			message("webhook_created", map[string]any{"webhook_id": "wh_1"}),
		})
		require.NoError(t, err)
		assert.Empty(t, derived)
	})

	T.Run("a payload that is not a data change message derives nothing", func(t *testing.T) {
		t.Parallel()

		derived, err := SideEffect(t.Context(), nil, []outbox.Message{
			{Topic: "recipes", Payload: searchsync.NewEvent(searchsync.OpUpsert, "already_an_index_event")},
		})
		require.NoError(t, err)
		assert.Empty(t, derived)
	})

	T.Run("with no document ID under the key the table names", func(t *testing.T) {
		t.Parallel()

		// Refused rather than skipped. Skipping would put back the failure this package
		// removes: a write commits, the index never hears about it, and nothing says so.
		_, err := SideEffect(t.Context(), nil, []outbox.Message{
			message("valid_instrument_created", map[string]any{"some_other_key": "instrument_123"}),
		})
		assert.Error(t, err)
	})

	T.Run("with no messages", func(t *testing.T) {
		t.Parallel()

		derived, err := SideEffect(t.Context(), nil, nil)
		require.NoError(t, err)
		assert.Empty(t, derived)
	})
}

func TestTable(T *testing.T) {
	T.Parallel()

	T.Run("every row names an index and a key", func(t *testing.T) {
		t.Parallel()

		for _, eventType := range EventTypes() {
			spec, ok := SpecFor(eventType)
			require.True(t, ok, eventType)

			assert.NotEmpty(t, spec.Index, eventType)
			assert.NotEmpty(t, spec.IDKey, eventType)
			assert.Contains(t, []searchsync.Op{searchsync.OpUpsert, searchsync.OpDelete}, spec.Op, eventType)
		}
	})

	T.Run("the index-only trigger is not a published event type", func(t *testing.T) {
		t.Parallel()

		// It stands in for an event that deliberately does not exist. If it ever collides with
		// a real event type, every write of that type would derive a recipe reindex.
		spec, ok := SpecFor(RecipeStepCreatedIndexTrigger)
		require.True(t, ok)
		assert.Equal(t, "recipes", spec.Index)
		assert.Contains(t, RecipeStepCreatedIndexTrigger, ".index_only")
	})
}
