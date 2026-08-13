package mealplanning

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v10/database"
	"github.com/primandproper/platform-go/v10/identifiers"
	searchsync "github.com/primandproper/platform-go/v10/search/sync"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// outboxRow is the subset of an outbox_messages row these tests read back.
type outboxRow struct {
	topic        string
	partitionKey string
	payload      []byte
}

// fetchOutboxRows reads every outbox row for a topic, oldest first.
func fetchOutboxRows(ctx context.Context, t *testing.T, db database.SQLQueryExecutor, topic string) []outboxRow {
	t.Helper()

	rows, err := db.QueryContext(ctx, `SELECT topic, partition_key, payload FROM outbox_messages WHERE topic = $1 ORDER BY created_at, id`, topic)
	require.NoError(t, err)
	defer func() { assert.NoError(t, rows.Close()) }()

	var out []outboxRow
	for rows.Next() {
		var r outboxRow
		require.NoError(t, rows.Scan(&r.topic, &r.partitionKey, &r.payload))
		out = append(out, r)
	}
	require.NoError(t, rows.Err())

	return out
}

// decodeDataChangeMessages decodes each row's payload as the message the async handler consumes.
// The relay republishes these bytes verbatim, so what decodes here is what a consumer receives.
func decodeDataChangeMessages(t *testing.T, rows []outboxRow) []*audit.DataChangeMessage {
	t.Helper()

	out := make([]*audit.DataChangeMessage, 0, len(rows))
	for i := range rows {
		var msg audit.DataChangeMessage
		require.NoError(t, json.Unmarshal(rows[i].payload, &msg))
		out = append(out, &msg)
	}

	return out
}

func findEvent(msgs []*audit.DataChangeMessage, eventType string) *audit.DataChangeMessage {
	for _, msg := range msgs {
		if msg.EventType == eventType {
			return msg
		}
	}

	return nil
}

func TestQuerier_Integration_MealPlanOutboxEvents(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	recipe := createRecipeForTest(t, ctx, nil, dbc, true)
	meal := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)

	exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
	exampleMealPlan.BelongsToAccount = account.ID

	// create — the event is another statement in CreateMealPlan's transaction
	created := createMealPlanForTest(t, ctx, exampleMealPlan, dbc)

	msgs := decodeDataChangeMessages(t, fetchOutboxRows(ctx, t, dbc.writeDB, testDataChangesTopic))
	createdEvent := findEvent(msgs, types.MealPlanCreatedServiceEventType)
	require.NotNil(t, createdEvent, "no created event was enqueued")
	assert.Equal(t, account.ID, createdEvent.AccountID)
	assert.Equal(t, created.ID, createdEvent.Context[mealplanningkeys.MealPlanIDKey])

	// The partition key is what preserves per-account ordering, so a row that names an account
	// must be keyed by it. Rows for meal-plan child entities carry no account and are keyed by
	// nothing, which the relay treats as unordered — correct, since nothing orders them.
	for _, row := range fetchOutboxRows(ctx, t, dbc.writeDB, testDataChangesTopic) {
		var msg audit.DataChangeMessage
		require.NoError(t, json.Unmarshal(row.payload, &msg))

		if msg.AccountID != "" {
			assert.Equal(t, msg.AccountID, row.partitionKey)
		}
	}

	// archive
	require.NoError(t, dbc.ArchiveMealPlan(ctx, created.ID, account.ID))

	msgs = decodeDataChangeMessages(t, fetchOutboxRows(ctx, t, dbc.writeDB, testDataChangesTopic))
	archivedEvent := findEvent(msgs, types.MealPlanArchivedServiceEventType)
	require.NotNil(t, archivedEvent, "no archived event was enqueued")
	assert.Equal(t, account.ID, archivedEvent.AccountID)
	assert.Equal(t, created.ID, archivedEvent.Context[mealplanningkeys.MealPlanIDKey])
}

func TestQuerier_Integration_MealPlanOutboxRollsBackWithItsWrite(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	// Archiving a meal plan that does not exist affects no rows, so the transaction rolls
	// back. This is the whole point of the outbox: the event rolls back with it, rather than
	// announcing a change that never happened.
	require.Error(t, dbc.ArchiveMealPlan(ctx, "nonexistent", account.ID))

	assert.Empty(t, fetchOutboxRows(ctx, t, dbc.writeDB, testDataChangesTopic))
}

func TestQuerier_Integration_GroceryListInitializationEmitsOneEventPerItem(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	recipe := createRecipeForTest(t, ctx, nil, dbc, true)
	meal := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)

	exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
	exampleMealPlan.BelongsToAccount = account.ID
	mealPlan := createMealPlanForTest(t, ctx, exampleMealPlan, dbc)

	ingredient := createValidIngredientForTest(t, ctx, nil, dbc)
	unit := createValidMeasurementUnitForTest(t, ctx, nil, dbc)

	buildInput := func() *types.MealPlanGroceryListItemDatabaseCreationInput {
		return &types.MealPlanGroceryListItemDatabaseCreationInput{
			ID:                     identifiers.New(),
			BelongsToMealPlan:      mealPlan.ID,
			ValidIngredientID:      ingredient.ID,
			ValidMeasurementUnitID: unit.ID,
			Status:                 types.MealPlanGroceryListItemStatusNeeds,
			MinQuantityNeeded:      1,
		}
	}

	countCreatedEvents := func() int {
		msgs := decodeDataChangeMessages(t, fetchOutboxRows(ctx, t, dbc.writeDB, testDataChangesTopic))

		var found int
		for _, msg := range msgs {
			if msg.EventType == types.MealPlanGroceryListItemCreatedServiceEventType {
				found++
				assert.Equal(t, account.ID, msg.AccountID,
					"a background job has no session, so the account has to come from the caller")
			}
		}

		return found
	}

	// a batch that fails announces nothing: the events are statements in the transaction that
	// rolled back.
	doomed := buildInput()
	doomed.ValidIngredientID = identifiers.New()

	_, err := dbc.InitializeMealPlanGroceryList(ctx, mealPlan.ID, account.ID, []*types.MealPlanGroceryListItemDatabaseCreationInput{buildInput(), doomed})
	require.Error(t, err)
	assert.Zero(t, countCreatedEvents())

	// and a batch that succeeds announces each item exactly once. The initializer job used to
	// publish these itself on top of the repository's transactional emit, so every item announced
	// itself twice.
	inputs := []*types.MealPlanGroceryListItemDatabaseCreationInput{buildInput(), buildInput()}
	created, err := dbc.InitializeMealPlanGroceryList(ctx, mealPlan.ID, account.ID, inputs)
	require.NoError(t, err)
	require.Len(t, created, len(inputs))

	assert.Equal(t, len(inputs), countCreatedEvents())
}

func TestQuerier_Integration_CatalogOutboxEvents(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Catalog entities are global — owned by no account — so their events carry no account ID.
	created := createValidVesselForTest(t, ctx, nil, dbc)

	msgs := decodeDataChangeMessages(t, fetchOutboxRows(ctx, t, dbc.writeDB, testDataChangesTopic))
	createdEvent := findEvent(msgs, types.ValidVesselCreatedServiceEventType)
	require.NotNil(t, createdEvent, "no created event was enqueued")
	assert.Equal(t, created.ID, createdEvent.Context[mealplanningkeys.ValidVesselIDKey])

	require.NoError(t, dbc.ArchiveValidVessel(ctx, created.ID))

	msgs = decodeDataChangeMessages(t, fetchOutboxRows(ctx, t, dbc.writeDB, testDataChangesTopic))
	archivedEvent := findEvent(msgs, types.ValidVesselArchivedServiceEventType)
	require.NotNil(t, archivedEvent, "no archived event was enqueued")
	assert.Equal(t, created.ID, archivedEvent.Context[mealplanningkeys.ValidVesselIDKey])
}

func TestQuerier_Integration_RecipeCloneEmitsBothEvents(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	original := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, true)

	// A clone is one write, so both the created and the cloned event commit with it.
	cloneInput := converters.ConvertRecipeToRecipeDatabaseCreationInput(
		buildRecipeForTestCreation(t, ctx, user.ID, dbc))
	cloneInput.ClonedFromRecipeID = &original.ID

	clone, err := dbc.CreateRecipe(ctx, cloneInput)
	require.NoError(t, err)
	require.NotNil(t, clone)

	msgs := decodeDataChangeMessages(t, fetchOutboxRows(ctx, t, dbc.writeDB, testDataChangesTopic))

	cloned := findEvent(msgs, types.RecipeClonedServiceEventType)
	require.NotNil(t, cloned, "no cloned event was enqueued")
	assert.Equal(t, original.ID, cloned.Context[mealplanningkeys.RecipeIDKey],
		"the cloned event should name the recipe that was copied from")

	var createdForClone bool
	for _, msg := range msgs {
		if msg.EventType == types.RecipeCreatedServiceEventType && msg.Context[mealplanningkeys.RecipeIDKey] == clone.ID {
			createdForClone = true
		}
	}
	assert.True(t, createdForClone, "the clone should also announce itself as created")
}

// decodeIndexEvents decodes each row's payload as the index event a searchsync.Syncer consumes.
func decodeIndexEvents(t *testing.T, rows []outboxRow) []searchsync.Event {
	t.Helper()

	out := make([]searchsync.Event, 0, len(rows))
	for i := range rows {
		var event searchsync.Event
		require.NoError(t, json.Unmarshal(rows[i].payload, &event))
		out = append(out, event)
	}

	return out
}

func TestQuerier_Integration_IndexEventsCommitWithTheirWrite(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	indexRows := func() []outboxRow {
		return fetchOutboxRows(ctx, t, dbc.writeDB, mealplanningindexing.IndexTypeValidVessels)
	}

	// create
	created := createValidVesselForTest(t, ctx, nil, dbc)

	rows := indexRows()
	require.Len(t, rows, 1, "creating a vessel should enqueue exactly one index event")
	// The document ID is the partition key, which is what buys per-document ordering: the
	// relay admits a keyed message only while no older one with that key is pending, so two
	// edits to one vessel can never be applied out of order.
	assert.Equal(t, created.ID, rows[0].partitionKey)

	events := decodeIndexEvents(t, rows)
	assert.Equal(t, created.ID, events[0].DocumentID)
	assert.Equal(t, searchsync.OpUpsert, events[0].Op)

	// update
	require.NoError(t, dbc.UpdateValidVessel(ctx, created))

	events = decodeIndexEvents(t, indexRows())
	require.Len(t, events, 2)
	assert.Equal(t, created.ID, events[1].DocumentID)
	assert.Equal(t, searchsync.OpUpsert, events[1].Op)

	// archive — an archived row is one search must stop returning, so it is a delete
	require.NoError(t, dbc.ArchiveValidVessel(ctx, created.ID))

	events = decodeIndexEvents(t, indexRows())
	require.Len(t, events, 3)
	assert.Equal(t, created.ID, events[2].DocumentID)
	assert.Equal(t, searchsync.OpDelete, events[2].Op)
}

func TestQuerier_Integration_IndexEventRollsBackWithItsWrite(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Archiving a vessel that does not exist affects no rows, so the transaction rolls back.
	// The index event rolls back with it — which is the whole reason it is enqueued here
	// rather than published by a consumer reading the data change event downstream, where a
	// row that never changed could still have been announced to the index.
	require.Error(t, dbc.ArchiveValidVessel(ctx, "nonexistent"))

	assert.Empty(t, fetchOutboxRows(ctx, t, dbc.writeDB, mealplanningindexing.IndexTypeValidVessels))
}

func TestQuerier_Integration_RecipeCreationEnqueuesOneIndexEvent(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Recipes emit inside a larger transaction rather than through withEvent, so this covers
	// the other of the two shapes a write takes in this package.
	created := createRecipeForTest(t, ctx, nil, dbc, false)

	events := decodeIndexEvents(t, fetchOutboxRows(ctx, t, dbc.writeDB, mealplanningindexing.IndexTypeRecipes))
	require.Len(t, events, 1)
	assert.Equal(t, created.ID, events[0].DocumentID)
	assert.Equal(t, searchsync.OpUpsert, events[0].Op)
}

func TestQuerier_Integration_RecipeStepWritesReindexTheirRecipe(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)

	recipeEvents := func() []searchsync.Event {
		return decodeIndexEvents(t, fetchOutboxRows(ctx, t, dbc.writeDB, mealplanningindexing.IndexTypeRecipes))
	}

	// The recipe's own creation is one event; every write below is a change to the same
	// document, because the indexed subset holds each step's preparation name and the names
	// of its ingredients, instruments and vessels.
	require.Len(t, recipeEvents(), 1)

	step := createRecipeStepForTest(t, ctx, recipe.ID, buildRecipeStepForTestCreation(t, ctx, recipe.ID, dbc), dbc)
	require.Len(t, recipeEvents(), 2, "adding a step should reindex the recipe")

	step.Notes = "updated"
	require.NoError(t, dbc.UpdateRecipeStep(ctx, step))
	require.Len(t, recipeEvents(), 3, "updating a step should reindex the recipe")

	require.NoError(t, dbc.ArchiveRecipeStep(ctx, recipe.ID, step.ID))
	events := recipeEvents()
	require.Len(t, events, 4, "archiving a step should reindex the recipe")

	// Every one of them names the recipe, and every one is an upsert: the step went away, the
	// recipe did not, so the document is rewritten rather than removed.
	for _, event := range events {
		assert.Equal(t, recipe.ID, event.DocumentID)
		assert.Equal(t, searchsync.OpUpsert, event.Op)
	}
}

func TestQuerier_Integration_RecipeStepCreationIsAtomicWithItsIndexEvent(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)

	before := len(decodeIndexEvents(t, fetchOutboxRows(ctx, t, dbc.writeDB, mealplanningindexing.IndexTypeRecipes)))

	// A step naming a preparation that does not exist fails partway through the write. It used
	// to run outside a transaction, so the step row survived a failure among its children;
	// now the whole thing rolls back, and the index event with it.
	doomed := converters.ConvertRecipeStepToRecipeStepDatabaseCreationInput(
		buildRecipeStepForTestCreation(t, ctx, recipe.ID, dbc))
	doomed.PreparationID = identifiers.New()

	_, err := dbc.CreateRecipeStep(ctx, doomed)
	require.Error(t, err)

	assert.Len(t, decodeIndexEvents(t, fetchOutboxRows(ctx, t, dbc.writeDB, mealplanningindexing.IndexTypeRecipes)), before)

	exists, err := dbc.RecipeStepExists(ctx, recipe.ID, doomed.ID)
	require.NoError(t, err)
	assert.False(t, exists, "the step row should have rolled back with its index event")
}
