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

	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/identifiers"

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
