/*
Package indexevents says which writes feed which search index.

A write that changes an indexed row owes the index an event. That obligation used to be an
option on the emit call — every repository method that touched an indexed entity passed
WithIndexUpsert or WithIndexDelete by hand, at forty call sites — and an obligation expressed as
an option is one a call site can forget. A repository method that omitted it compiled, reviewed
clean, and left the index stale until the next scheduled rebuild, with no test that would catch
it and no metric that would show it.

So the obligation is no longer a parameter. The rules below are registered on the outbox Writer
once, as platform's searchsync side effect, which runs inside every Enqueue and derives the index
events from the data change messages the caller was already sending. The table is the whole of
what it knows, and it is the one place to edit when an entity becomes indexed.

The matching, the refusal of a message with no document ID, and the per-document ordering key
are platform's (searchsync.NewSideEffect). What is this application's is the table: which of its
event types feed which index, and under which context key the document's ID travels.

# What the table says

Each row maps an event type to the index it feeds, whether the document is written or removed,
and the key in the message's context that holds the document's ID. Everything a row needs is
already in the message: the event type says what happened, and the context map carries the ID.
datachanges.Message answers both through searchsync.Change.

An event type absent from the table produces no index event, which is right — most of them
should not.

# Two things that are easy to get wrong

Archiving a *sub-entity* of an indexed document is an upsert, not a delete. Archiving a recipe
step leaves the recipe indexed and changes what it says, so RecipeStepArchived upserts the
recipe. Only archiving the indexed entity itself is a delete.

The document ID is not always the ID of the thing that changed. Every recipe step, ingredient,
instrument and vessel write reindexes its parent *recipe*, because the indexed recipe document
embeds their names — so those rows read the recipe's ID out of the context, not the sub-entity's.
*/
package indexevents

import (
	"slices"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/platform-go/v15/searchsync"
)

// SideEffectName identifies this effect on the Writer. It appears in the error a duplicate or
// nil registration is refused with.
const SideEffectName = "search-index"

// RecipeStepCreatedIndexTrigger stands in for an event type that does not exist.
//
// Creating a recipe step reindexes its recipe but announces nothing: there is no
// recipe_step_created data change event, and there deliberately is not one. The constant existed
// once and put the event in the generated webhook catalog, where it was subscribable and could
// never fire; it was removed rather than made to fire, because a step creation reaches
// subscribers as the recipe event that accompanies it.
//
// That decision stands, so this is a trigger rather than an event type: Emitter.EmitIndex passes
// it to derive the index event without putting anything on the wire. It is not in the webhook
// catalog, and nothing publishes it.
const RecipeStepCreatedIndexTrigger = "recipe_step_created.index_only"

// rule is one row: the index an event feeds, where to find the document's ID, and what to do to
// the document. The index's name is the rule's topic, because platform says which index an event
// belongs to by where it arrived.
func rule(eventType, index, idKey string, op searchsync.Op) searchsync.Rule {
	return searchsync.Rule{EventType: eventType, Topic: index, IDKey: idKey, Op: op}
}

// rules is the table. Adding an indexed entity means adding rows here and nothing in the
// repository layer.
var rules = []searchsync.Rule{
	// Users.
	rule(identity.UserSignedUpServiceEventType, identityindexing.IndexTypeUsers, identitykeys.UserIDKey, searchsync.OpUpsert),
	rule(identity.UsernameChangedEventType, identityindexing.IndexTypeUsers, identitykeys.UserIDKey, searchsync.OpUpsert),
	rule(identity.EmailAddressChangedEventType, identityindexing.IndexTypeUsers, identitykeys.UserIDKey, searchsync.OpUpsert),
	rule(identity.UserDetailsChangedEventType, identityindexing.IndexTypeUsers, identitykeys.UserIDKey, searchsync.OpUpsert),
	rule(identity.UserArchivedServiceEventType, identityindexing.IndexTypeUsers, identitykeys.UserIDKey, searchsync.OpDelete),

	// Meals.
	rule(types.MealCreatedServiceEventType, mealplanningindexing.IndexTypeMeals, mealplanningkeys.MealIDKey, searchsync.OpUpsert),
	rule(types.MealArchivedServiceEventType, mealplanningindexing.IndexTypeMeals, mealplanningkeys.MealIDKey, searchsync.OpDelete),

	// Recipes, and everything under them. The indexed recipe document embeds each step's
	// preparation name and the names of its ingredients, instruments and vessels, so a write to
	// any of those reindexes the recipe — which is why every row here reads RecipeIDKey, and why
	// archiving a sub-entity is an upsert.
	rule(types.RecipeCreatedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
	rule(types.RecipeUpdatedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
	rule(types.RecipeArchivedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpDelete),

	rule(RecipeStepCreatedIndexTrigger, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
	rule(types.RecipeStepUpdatedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
	rule(types.RecipeStepArchivedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),

	rule(types.RecipeStepIngredientCreatedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
	rule(types.RecipeStepIngredientUpdatedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
	rule(types.RecipeStepIngredientArchivedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),

	rule(types.RecipeStepInstrumentCreatedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
	rule(types.RecipeStepInstrumentUpdatedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
	rule(types.RecipeStepInstrumentArchivedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),

	rule(types.RecipeStepVesselCreatedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
	rule(types.RecipeStepVesselUpdatedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
	rule(types.RecipeStepVesselArchivedServiceEventType, mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),

	// The catalog entities, each its own index and its own document.
	rule(types.ValidIngredientCreatedServiceEventType, mealplanningindexing.IndexTypeValidIngredients, mealplanningkeys.ValidIngredientIDKey, searchsync.OpUpsert),
	rule(types.ValidIngredientUpdatedServiceEventType, mealplanningindexing.IndexTypeValidIngredients, mealplanningkeys.ValidIngredientIDKey, searchsync.OpUpsert),
	rule(types.ValidIngredientArchivedServiceEventType, mealplanningindexing.IndexTypeValidIngredients, mealplanningkeys.ValidIngredientIDKey, searchsync.OpDelete),

	rule(types.ValidIngredientStateCreatedServiceEventType, mealplanningindexing.IndexTypeValidIngredientStates, mealplanningkeys.ValidIngredientStateIDKey, searchsync.OpUpsert),
	rule(types.ValidIngredientStateUpdatedServiceEventType, mealplanningindexing.IndexTypeValidIngredientStates, mealplanningkeys.ValidIngredientStateIDKey, searchsync.OpUpsert),
	rule(types.ValidIngredientStateArchivedServiceEventType, mealplanningindexing.IndexTypeValidIngredientStates, mealplanningkeys.ValidIngredientStateIDKey, searchsync.OpDelete),

	rule(types.ValidInstrumentCreatedServiceEventType, mealplanningindexing.IndexTypeValidInstruments, mealplanningkeys.ValidInstrumentIDKey, searchsync.OpUpsert),
	rule(types.ValidInstrumentUpdatedServiceEventType, mealplanningindexing.IndexTypeValidInstruments, mealplanningkeys.ValidInstrumentIDKey, searchsync.OpUpsert),
	rule(types.ValidInstrumentArchivedServiceEventType, mealplanningindexing.IndexTypeValidInstruments, mealplanningkeys.ValidInstrumentIDKey, searchsync.OpDelete),

	rule(types.ValidMeasurementUnitCreatedServiceEventType, mealplanningindexing.IndexTypeValidMeasurementUnits, mealplanningkeys.ValidMeasurementUnitIDKey, searchsync.OpUpsert),
	rule(types.ValidMeasurementUnitUpdatedServiceEventType, mealplanningindexing.IndexTypeValidMeasurementUnits, mealplanningkeys.ValidMeasurementUnitIDKey, searchsync.OpUpsert),
	rule(types.ValidMeasurementUnitArchivedServiceEventType, mealplanningindexing.IndexTypeValidMeasurementUnits, mealplanningkeys.ValidMeasurementUnitIDKey, searchsync.OpDelete),

	rule(types.ValidPreparationCreatedServiceEventType, mealplanningindexing.IndexTypeValidPreparations, mealplanningkeys.ValidPreparationIDKey, searchsync.OpUpsert),
	rule(types.ValidPreparationUpdatedServiceEventType, mealplanningindexing.IndexTypeValidPreparations, mealplanningkeys.ValidPreparationIDKey, searchsync.OpUpsert),
	rule(types.ValidPreparationArchivedServiceEventType, mealplanningindexing.IndexTypeValidPreparations, mealplanningkeys.ValidPreparationIDKey, searchsync.OpDelete),

	rule(types.ValidVesselCreatedServiceEventType, mealplanningindexing.IndexTypeValidVessels, mealplanningkeys.ValidVesselIDKey, searchsync.OpUpsert),
	rule(types.ValidVesselUpdatedServiceEventType, mealplanningindexing.IndexTypeValidVessels, mealplanningkeys.ValidVesselIDKey, searchsync.OpUpsert),
	rule(types.ValidVesselArchivedServiceEventType, mealplanningindexing.IndexTypeValidVessels, mealplanningkeys.ValidVesselIDKey, searchsync.OpDelete),
}

// Rules is the table, as a fresh slice: platform validates and keeps what it is handed, and a
// shared slice would let one caller's mutation change what every other caller registered.
func Rules() []searchsync.Rule {
	return slices.Clone(rules)
}

// NewSideEffect builds the outbox side effect that derives this application's index events.
//
// Register it once on the outbox Writer. It then runs inside every Enqueue, on the caller's
// executor, so the index events are written by the same statement as the row change and commit
// with it — an index event cannot outlive a rolled-back write, and a committed write cannot lose
// its index event. The one failure it refuses rather than skips is a tabled event with no
// document ID under the key its rule names; skipping that would put back the failure this
// package removes.
func NewSideEffect() (outbox.SideEffect, error) {
	return searchsync.NewSideEffect(Rules())
}

// RulesFor returns the rows for an event type, in table order. An untabled event type has none.
func RulesFor(eventType string) []searchsync.Rule {
	var matched []searchsync.Rule
	for i := range rules {
		if rules[i].EventType == eventType {
			matched = append(matched, rules[i])
		}
	}

	return matched
}

// EventTypes reports every event type the table covers, each once, for tests and for anything
// that wants to assert the set rather than read it.
func EventTypes() []string {
	out := make([]string, 0, len(rules))
	for i := range rules {
		if !slices.Contains(out, rules[i].EventType) {
			out = append(out, rules[i].EventType)
		}
	}

	return out
}
