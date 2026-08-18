/*
Package indexevents says which writes feed which search index.

A write that changes an indexed row owes the index an event. That obligation used to be an
option on the emit call — every repository method that touched an indexed entity passed
WithIndexUpsert or WithIndexDelete by hand, at forty call sites — and an obligation expressed as
an option is one a call site can forget. A repository method that omitted it compiled, reviewed
clean, and left the index stale until the next scheduled rebuild, with no test that would catch
it and no metric that would show it.

So the obligation is no longer a parameter. SideEffect is registered on the outbox Writer once,
runs inside every Enqueue, and derives the index events from the data change messages the caller
was already sending. The table below is the whole of what it knows, and it is the one place to
edit when an entity becomes indexed.

# What the table says

Each row maps an event type to the index it feeds, whether the document is written or removed,
and the key in the message's context that holds the document's ID. Everything a row needs is
already in the message: the event type says what happened, and the context map carries the ID.

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
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v10/database"
	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/outbox"
	searchsync "github.com/primandproper/platform-go/v10/search/sync"
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

// Spec is one row: the index an event feeds, what it does to the document, and where to find
// the document's ID.
type Spec struct {
	// Index is the index's name, which the indexing packages declare as their IndexType
	// constants. It is also the topic, because platform-go says which index an event belongs to
	// by where it arrived.
	Index string

	// IDKey is the key in the message's context holding the ID the index holds the document
	// under. It is not always the ID of the entity that changed — see the package doc.
	IDKey string

	// Op is whether the document is written or removed. Archival of the indexed entity counts
	// as removal; archival of a sub-entity does not.
	Op searchsync.Op
}

// byEventType is the table. Adding an indexed entity means adding rows here and nothing in the
// repository layer.
var byEventType = map[string]Spec{
	// Users.
	identity.UserSignedUpServiceEventType: {identityindexing.IndexTypeUsers, identitykeys.UserIDKey, searchsync.OpUpsert},
	identity.UsernameChangedEventType:     {identityindexing.IndexTypeUsers, identitykeys.UserIDKey, searchsync.OpUpsert},
	identity.EmailAddressChangedEventType: {identityindexing.IndexTypeUsers, identitykeys.UserIDKey, searchsync.OpUpsert},
	identity.UserDetailsChangedEventType:  {identityindexing.IndexTypeUsers, identitykeys.UserIDKey, searchsync.OpUpsert},
	identity.UserArchivedServiceEventType: {identityindexing.IndexTypeUsers, identitykeys.UserIDKey, searchsync.OpDelete},

	// Meals.
	types.MealCreatedServiceEventType:  {mealplanningindexing.IndexTypeMeals, mealplanningkeys.MealIDKey, searchsync.OpUpsert},
	types.MealArchivedServiceEventType: {mealplanningindexing.IndexTypeMeals, mealplanningkeys.MealIDKey, searchsync.OpDelete},

	// Recipes, and everything under them. The indexed recipe document embeds each step's
	// preparation name and the names of its ingredients, instruments and vessels, so a write to
	// any of those reindexes the recipe — which is why every row here reads RecipeIDKey, and why
	// archiving a sub-entity is an upsert.
	types.RecipeCreatedServiceEventType:  {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},
	types.RecipeUpdatedServiceEventType:  {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},
	types.RecipeArchivedServiceEventType: {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpDelete},

	RecipeStepCreatedIndexTrigger:            {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},
	types.RecipeStepUpdatedServiceEventType:  {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},
	types.RecipeStepArchivedServiceEventType: {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},

	types.RecipeStepIngredientCreatedServiceEventType:  {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},
	types.RecipeStepIngredientUpdatedServiceEventType:  {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},
	types.RecipeStepIngredientArchivedServiceEventType: {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},

	types.RecipeStepInstrumentCreatedServiceEventType:  {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},
	types.RecipeStepInstrumentUpdatedServiceEventType:  {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},
	types.RecipeStepInstrumentArchivedServiceEventType: {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},

	types.RecipeStepVesselCreatedServiceEventType:  {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},
	types.RecipeStepVesselUpdatedServiceEventType:  {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},
	types.RecipeStepVesselArchivedServiceEventType: {mealplanningindexing.IndexTypeRecipes, mealplanningkeys.RecipeIDKey, searchsync.OpUpsert},

	// The catalog entities, each its own index and its own document.
	types.ValidIngredientCreatedServiceEventType:  {mealplanningindexing.IndexTypeValidIngredients, mealplanningkeys.ValidIngredientIDKey, searchsync.OpUpsert},
	types.ValidIngredientUpdatedServiceEventType:  {mealplanningindexing.IndexTypeValidIngredients, mealplanningkeys.ValidIngredientIDKey, searchsync.OpUpsert},
	types.ValidIngredientArchivedServiceEventType: {mealplanningindexing.IndexTypeValidIngredients, mealplanningkeys.ValidIngredientIDKey, searchsync.OpDelete},

	types.ValidIngredientStateCreatedServiceEventType:  {mealplanningindexing.IndexTypeValidIngredientStates, mealplanningkeys.ValidIngredientStateIDKey, searchsync.OpUpsert},
	types.ValidIngredientStateUpdatedServiceEventType:  {mealplanningindexing.IndexTypeValidIngredientStates, mealplanningkeys.ValidIngredientStateIDKey, searchsync.OpUpsert},
	types.ValidIngredientStateArchivedServiceEventType: {mealplanningindexing.IndexTypeValidIngredientStates, mealplanningkeys.ValidIngredientStateIDKey, searchsync.OpDelete},

	types.ValidInstrumentCreatedServiceEventType:  {mealplanningindexing.IndexTypeValidInstruments, mealplanningkeys.ValidInstrumentIDKey, searchsync.OpUpsert},
	types.ValidInstrumentUpdatedServiceEventType:  {mealplanningindexing.IndexTypeValidInstruments, mealplanningkeys.ValidInstrumentIDKey, searchsync.OpUpsert},
	types.ValidInstrumentArchivedServiceEventType: {mealplanningindexing.IndexTypeValidInstruments, mealplanningkeys.ValidInstrumentIDKey, searchsync.OpDelete},

	types.ValidMeasurementUnitCreatedServiceEventType:  {mealplanningindexing.IndexTypeValidMeasurementUnits, mealplanningkeys.ValidMeasurementUnitIDKey, searchsync.OpUpsert},
	types.ValidMeasurementUnitUpdatedServiceEventType:  {mealplanningindexing.IndexTypeValidMeasurementUnits, mealplanningkeys.ValidMeasurementUnitIDKey, searchsync.OpUpsert},
	types.ValidMeasurementUnitArchivedServiceEventType: {mealplanningindexing.IndexTypeValidMeasurementUnits, mealplanningkeys.ValidMeasurementUnitIDKey, searchsync.OpDelete},

	types.ValidPreparationCreatedServiceEventType:  {mealplanningindexing.IndexTypeValidPreparations, mealplanningkeys.ValidPreparationIDKey, searchsync.OpUpsert},
	types.ValidPreparationUpdatedServiceEventType:  {mealplanningindexing.IndexTypeValidPreparations, mealplanningkeys.ValidPreparationIDKey, searchsync.OpUpsert},
	types.ValidPreparationArchivedServiceEventType: {mealplanningindexing.IndexTypeValidPreparations, mealplanningkeys.ValidPreparationIDKey, searchsync.OpDelete},

	types.ValidVesselCreatedServiceEventType:  {mealplanningindexing.IndexTypeValidVessels, mealplanningkeys.ValidVesselIDKey, searchsync.OpUpsert},
	types.ValidVesselUpdatedServiceEventType:  {mealplanningindexing.IndexTypeValidVessels, mealplanningkeys.ValidVesselIDKey, searchsync.OpUpsert},
	types.ValidVesselArchivedServiceEventType: {mealplanningindexing.IndexTypeValidVessels, mealplanningkeys.ValidVesselIDKey, searchsync.OpDelete},
}

// SideEffect derives the index events implied by the messages a caller enqueued.
//
// Register it once on the outbox Writer. It then runs inside every Enqueue, on the caller's
// executor, so the index events are written by the same statement as the row change and commit
// with it — an index event cannot outlive a rolled-back write, and a committed write cannot lose
// its index event.
//
// The executor is unused: every row this needs is already in the message. It is in the signature
// because an effect that has to read the database to decide is the reason the signature has one.
func SideEffect(_ context.Context, _ database.SQLQueryExecutor, msgs []outbox.Message) ([]outbox.Message, error) {
	var derived []outbox.Message

	for i := range msgs {
		// The outbox never looks inside a Payload, so the assertion is ours to make. A message
		// that is not a data change message is another effect's business or nobody's.
		msg, ok := msgs[i].Payload.(*audit.DataChangeMessage)
		if !ok {
			continue
		}

		spec, tabled := byEventType[msg.EventType]
		if !tabled {
			continue
		}

		documentID, isString := msg.Context[spec.IDKey].(string)
		if !isString || documentID == "" {
			// Refused rather than skipped, and the caller's transaction rolls back with it.
			// Skipping would put back exactly the failure this package removes: a write that
			// commits, an index that does not hear about it, and nothing anywhere that says so.
			return nil, platformerrors.Errorf(
				"deriving index event for %q: context has no document ID under %q",
				msg.EventType, spec.IDKey,
			)
		}

		// Built through searchsync so the outbox key is the document ID, which is what buys
		// per-document ordering: at most one event per document in flight, however many relays
		// are running.
		derived = append(derived, searchsync.NewEvent(spec.Op, documentID).Message(spec.Index))
	}

	return derived, nil
}

// EventTypes reports every event type the table covers, for tests and for anything that wants to
// assert the set rather than read it.
func EventTypes() []string {
	out := make([]string, 0, len(byEventType))
	for eventType := range byEventType {
		out = append(out, eventType)
	}

	return out
}

// SpecFor returns the row for an event type, if it has one.
func SpecFor(eventType string) (Spec, bool) {
	spec, ok := byEventType[eventType]

	return spec, ok
}
