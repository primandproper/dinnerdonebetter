package searchindex

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"

	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/platform-go/v15/searchsync"
	"github.com/primandproper/platform-go/v15/webhooks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func message(eventType string, context map[string]any) outbox.Message {
	return outbox.Message{Topic: "data_changes", Payload: &datachanges.Message{EventType: eventType, Context: context}}
}

// ruleFor is the one row an event type has in this domain's table.
func ruleFor(t *testing.T, eventType string) searchsync.Rule {
	t.Helper()

	var matched []searchsync.Rule

	rules := IndexRules()
	for i := range rules {
		if rules[i].EventType == eventType {
			matched = append(matched, rules[i])
		}
	}

	require.Len(t, matched, 1, eventType)

	return matched[0]
}

// effect is platform's side effect over this domain's table alone.
func effect(t *testing.T) outbox.SideEffect {
	t.Helper()

	built, err := searchsync.NewSideEffect(IndexRules())
	require.NoError(t, err)

	return built
}

func TestIndexRules(T *testing.T) {
	T.Parallel()

	T.Run("every row names one of this domain's indexes and a key", func(t *testing.T) {
		t.Parallel()

		rules := IndexRules()
		require.NotEmpty(t, rules)

		for _, r := range rules {
			assert.Contains(t, IndexNames(), r.Topic, r.EventType)
			assert.NotEmpty(t, r.IDKey, r.EventType)
			assert.Contains(t, []searchsync.Op{searchsync.OpUpsert, searchsync.OpDelete}, r.Op, r.EventType)
		}
	})

	T.Run("every index is fed by at least one row", func(t *testing.T) {
		t.Parallel()

		fed := map[string]bool{}
		for _, r := range IndexRules() {
			fed[r.Topic] = true
		}

		for _, name := range IndexNames() {
			assert.True(t, fed[name], "index %q has a rebuild and no row feeding it", name)
		}
	})

	T.Run("the table is one platform accepts", func(t *testing.T) {
		t.Parallel()

		// platform refuses a duplicate rule and an incomplete one at construction.
		_, err := searchsync.NewSideEffect(IndexRules())
		require.NoError(t, err)
	})

	T.Run("archiving an indexed entity deletes its document", func(t *testing.T) {
		t.Parallel()

		r := ruleFor(t, mealplanning.ValidInstrumentArchivedServiceEventType)

		derived, err := effect(t)(t.Context(), nil, []outbox.Message{
			message(mealplanning.ValidInstrumentArchivedServiceEventType, map[string]any{r.IDKey: "gone"}),
		})
		require.NoError(t, err)
		require.Len(t, derived, 1)

		assert.Equal(t, IndexTypeValidInstruments, derived[0].Topic)

		event, ok := derived[0].Payload.(searchsync.Event)
		require.True(t, ok)
		assert.Equal(t, searchsync.OpDelete, event.Op)
	})

	T.Run("archiving a sub-entity reindexes its parent instead", func(t *testing.T) {
		t.Parallel()

		// An archived recipe step leaves the recipe indexed and changes what it says, so this
		// is an upsert of the recipe rather than a delete of anything.
		step := ruleFor(t, mealplanning.RecipeStepArchivedServiceEventType)
		recipe := ruleFor(t, mealplanning.RecipeUpdatedServiceEventType)

		// The row reads the parent recipe's ID, which is the same key a recipe's own events use.
		require.Equal(t, recipe.IDKey, step.IDKey)

		derived, err := effect(t)(t.Context(), nil, []outbox.Message{
			message(mealplanning.RecipeStepArchivedServiceEventType, map[string]any{step.IDKey: "recipe_1"}),
		})
		require.NoError(t, err)
		require.Len(t, derived, 1)

		assert.Equal(t, IndexTypeRecipes, derived[0].Topic)
		assert.Equal(t, "recipe_1", derived[0].Key)

		event, ok := derived[0].Payload.(searchsync.Event)
		require.True(t, ok)
		assert.Equal(t, searchsync.OpUpsert, event.Op)
		assert.Equal(t, "recipe_1", event.DocumentID)
	})

	T.Run("every write under a recipe reindexes the recipe", func(t *testing.T) {
		t.Parallel()

		recipe := ruleFor(t, mealplanning.RecipeUpdatedServiceEventType)

		for _, r := range IndexRules() {
			if r.Topic == IndexTypeRecipes {
				assert.Equal(t, recipe.IDKey, r.IDKey, r.EventType)
			}
		}
	})

	T.Run("the index-only trigger is not a published event type", func(t *testing.T) {
		t.Parallel()

		// It stands in for an event that deliberately does not exist. If it ever collides with
		// a real event type, every write of that type would derive a recipe reindex — and if
		// it were in the catalog, it would be subscribable and never fire.
		r := ruleFor(t, RecipeStepCreatedIndexTrigger)
		assert.Equal(t, IndexTypeRecipes, r.Topic)
		assert.Contains(t, RecipeStepCreatedIndexTrigger, ".index_only")

		_, published := catalog.Catalog()[webhooks.EventType(RecipeStepCreatedIndexTrigger)]
		assert.False(t, published)
	})
}
