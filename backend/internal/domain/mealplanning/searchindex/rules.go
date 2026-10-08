package searchindex

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	"github.com/primandproper/platform-go/v15/searchsync"
)

// RecipeStepCreatedIndexTrigger stands in for an event type that does not exist.
//
// Creating a recipe step reindexes its recipe but announces nothing: there is no
// recipe_step_created data change event, and there deliberately is not one. The constant existed
// once and put the event in the generated webhook catalog, where it was subscribable and could
// never fire; it was removed rather than made to fire, because a step creation reaches
// subscribers as the recipe event that accompanies it.
//
// That decision stands, so this is a trigger rather than an event type: the repository's
// emitIndex passes it to derive the index event without putting anything on the wire. It is not
// in the webhook catalog, and nothing publishes it.
const RecipeStepCreatedIndexTrigger = "recipe_step_created.index_only"

// rule is one row: the index an event feeds, where to find the document's ID, and what to do to
// the document. The index's name is the rule's topic, because platform says which index an event
// belongs to by where it arrived.
func rule(eventType, index, idKey string, op searchsync.Op) searchsync.Rule {
	return searchsync.Rule{EventType: eventType, Topic: index, IDKey: idKey, Op: op}
}

// IndexRules is this domain's entry in the table internal/indexevents registers on the outbox
// writer: which of its writes feed which of its indexes. Adding an indexed entity means adding
// rows here and nothing in the repository layer.
//
// Each row maps an event type to the index it feeds, whether the document is written or
// removed, and the key under which the document's ID travels in the event's context. Two things
// are easy to get wrong:
//
// Archiving a sub-entity of an indexed document is an upsert, not a delete. Archiving a recipe
// step leaves the recipe indexed and changes what it says, so RecipeStepArchived upserts the
// recipe. Only archiving the indexed entity itself is a delete.
//
// The document ID is not always the ID of the thing that changed. Every recipe step, ingredient,
// instrument and vessel write reindexes its parent recipe, because the indexed recipe document
// embeds their names — so those rows read the recipe's ID out of the context, not the
// sub-entity's.
func IndexRules() []searchsync.Rule {
	return []searchsync.Rule{
		// Meals.
		rule(mealplanning.MealCreatedServiceEventType, IndexTypeMeals, mealplanningkeys.MealIDKey, searchsync.OpUpsert),
		rule(mealplanning.MealArchivedServiceEventType, IndexTypeMeals, mealplanningkeys.MealIDKey, searchsync.OpDelete),

		// Recipes, and everything under them. The indexed recipe document embeds each step's
		// preparation name and the names of its ingredients, instruments and vessels, so a
		// write to any of those reindexes the recipe — which is why every row here reads
		// RecipeIDKey, and why archiving a sub-entity is an upsert.
		rule(mealplanning.RecipeCreatedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
		rule(mealplanning.RecipeUpdatedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
		rule(mealplanning.RecipeArchivedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpDelete),

		rule(RecipeStepCreatedIndexTrigger, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
		rule(mealplanning.RecipeStepUpdatedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
		rule(mealplanning.RecipeStepArchivedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),

		rule(mealplanning.RecipeStepIngredientCreatedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
		rule(mealplanning.RecipeStepIngredientUpdatedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
		rule(mealplanning.RecipeStepIngredientArchivedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),

		rule(mealplanning.RecipeStepInstrumentCreatedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
		rule(mealplanning.RecipeStepInstrumentUpdatedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
		rule(mealplanning.RecipeStepInstrumentArchivedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),

		rule(mealplanning.RecipeStepVesselCreatedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
		rule(mealplanning.RecipeStepVesselUpdatedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),
		rule(mealplanning.RecipeStepVesselArchivedServiceEventType, IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert),

		// The catalog entities, each its own index and its own document.
		rule(mealplanning.ValidIngredientCreatedServiceEventType, IndexTypeValidIngredients, mealplanningkeys.ValidIngredientIDKey, searchsync.OpUpsert),
		rule(mealplanning.ValidIngredientUpdatedServiceEventType, IndexTypeValidIngredients, mealplanningkeys.ValidIngredientIDKey, searchsync.OpUpsert),
		rule(mealplanning.ValidIngredientArchivedServiceEventType, IndexTypeValidIngredients, mealplanningkeys.ValidIngredientIDKey, searchsync.OpDelete),

		rule(mealplanning.ValidIngredientStateCreatedServiceEventType, IndexTypeValidIngredientStates, mealplanningkeys.ValidIngredientStateIDKey, searchsync.OpUpsert),
		rule(mealplanning.ValidIngredientStateUpdatedServiceEventType, IndexTypeValidIngredientStates, mealplanningkeys.ValidIngredientStateIDKey, searchsync.OpUpsert),
		rule(mealplanning.ValidIngredientStateArchivedServiceEventType, IndexTypeValidIngredientStates, mealplanningkeys.ValidIngredientStateIDKey, searchsync.OpDelete),

		rule(mealplanning.ValidInstrumentCreatedServiceEventType, IndexTypeValidInstruments, mealplanningkeys.ValidInstrumentIDKey, searchsync.OpUpsert),
		rule(mealplanning.ValidInstrumentUpdatedServiceEventType, IndexTypeValidInstruments, mealplanningkeys.ValidInstrumentIDKey, searchsync.OpUpsert),
		rule(mealplanning.ValidInstrumentArchivedServiceEventType, IndexTypeValidInstruments, mealplanningkeys.ValidInstrumentIDKey, searchsync.OpDelete),

		rule(mealplanning.ValidMeasurementUnitCreatedServiceEventType, IndexTypeValidMeasurementUnits, mealplanningkeys.ValidMeasurementUnitIDKey, searchsync.OpUpsert),
		rule(mealplanning.ValidMeasurementUnitUpdatedServiceEventType, IndexTypeValidMeasurementUnits, mealplanningkeys.ValidMeasurementUnitIDKey, searchsync.OpUpsert),
		rule(mealplanning.ValidMeasurementUnitArchivedServiceEventType, IndexTypeValidMeasurementUnits, mealplanningkeys.ValidMeasurementUnitIDKey, searchsync.OpDelete),

		rule(mealplanning.ValidPreparationCreatedServiceEventType, IndexTypeValidPreparations, mealplanningkeys.ValidPreparationIDKey, searchsync.OpUpsert),
		rule(mealplanning.ValidPreparationUpdatedServiceEventType, IndexTypeValidPreparations, mealplanningkeys.ValidPreparationIDKey, searchsync.OpUpsert),
		rule(mealplanning.ValidPreparationArchivedServiceEventType, IndexTypeValidPreparations, mealplanningkeys.ValidPreparationIDKey, searchsync.OpDelete),

		rule(mealplanning.ValidVesselCreatedServiceEventType, IndexTypeValidVessels, mealplanningkeys.ValidVesselIDKey, searchsync.OpUpsert),
		rule(mealplanning.ValidVesselUpdatedServiceEventType, IndexTypeValidVessels, mealplanningkeys.ValidVesselIDKey, searchsync.OpUpsert),
		rule(mealplanning.ValidVesselArchivedServiceEventType, IndexTypeValidVessels, mealplanningkeys.ValidVesselIDKey, searchsync.OpDelete),
	}
}

// IndexNames is every index this domain keeps, each of which IndexRules feeds.
func IndexNames() []string {
	return []string{
		IndexTypeRecipes,
		IndexTypeMeals,
		IndexTypeValidIngredients,
		IndexTypeValidInstruments,
		IndexTypeValidMeasurementUnits,
		IndexTypeValidPreparations,
		IndexTypeValidIngredientStates,
		IndexTypeValidVessels,
	}
}
