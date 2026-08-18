package mealplanning

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v11/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMealPlanGroceryListItemForTest(t *testing.T, ctx context.Context, exampleMealPlanGroceryListItem *types.MealPlanGroceryListItem, dbc *repository) *types.MealPlanGroceryListItem {
	t.Helper()

	// create
	dbInput := converters.ConvertMealPlanGroceryListItemToMealPlanGroceryListItemDatabaseCreationInput(exampleMealPlanGroceryListItem)

	created, err := dbc.CreateMealPlanGroceryListItem(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)

	exampleMealPlanGroceryListItem.CreatedAt = created.CreatedAt
	require.Equal(t, exampleMealPlanGroceryListItem.MeasurementUnit.ID, created.MeasurementUnit.ID)
	exampleMealPlanGroceryListItem.MeasurementUnit = created.MeasurementUnit
	require.Equal(t, exampleMealPlanGroceryListItem.PurchasedMeasurementUnit.ID, created.PurchasedMeasurementUnit.ID)
	exampleMealPlanGroceryListItem.PurchasedMeasurementUnit = created.PurchasedMeasurementUnit
	require.Equal(t, exampleMealPlanGroceryListItem.Ingredient.ID, created.Ingredient.ID)
	exampleMealPlanGroceryListItem.Ingredient = created.Ingredient
	assert.Equal(t, exampleMealPlanGroceryListItem, created)

	mealPlanGroceryListItem, err := dbc.GetMealPlanGroceryListItem(ctx, created.BelongsToMealPlan, created.ID)
	require.NoError(t, err)

	exampleMealPlanGroceryListItem.CreatedAt = mealPlanGroceryListItem.CreatedAt
	require.Equal(t, exampleMealPlanGroceryListItem.MeasurementUnit.ID, mealPlanGroceryListItem.MeasurementUnit.ID)
	exampleMealPlanGroceryListItem.MeasurementUnit = mealPlanGroceryListItem.MeasurementUnit
	require.Equal(t, exampleMealPlanGroceryListItem.PurchasedMeasurementUnit.ID, mealPlanGroceryListItem.PurchasedMeasurementUnit.ID)
	exampleMealPlanGroceryListItem.PurchasedMeasurementUnit = mealPlanGroceryListItem.PurchasedMeasurementUnit
	require.Equal(t, exampleMealPlanGroceryListItem.Ingredient.ID, mealPlanGroceryListItem.Ingredient.ID)
	exampleMealPlanGroceryListItem.Ingredient = mealPlanGroceryListItem.Ingredient
	require.Equal(t, exampleMealPlanGroceryListItem.CreatedAt, mealPlanGroceryListItem.CreatedAt)
	require.Equal(t, exampleMealPlanGroceryListItem.LastUpdatedAt, mealPlanGroceryListItem.LastUpdatedAt)
	require.Equal(t, exampleMealPlanGroceryListItem.ID, mealPlanGroceryListItem.ID)

	assert.Equal(t, exampleMealPlanGroceryListItem, mealPlanGroceryListItem)

	return mealPlanGroceryListItem
}

func TestQuerier_Integration_MealPlanGroceryListItems(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	recipe := createRecipeForTest(t, ctx, nil, dbc, true)
	buildMealForIntegrationTest(user.ID, recipe)
	meal := createMealForTest(t, ctx, nil, dbc)

	exampleMealPlan := buildMealPlanForIntegrationTest(user.ID, meal)
	exampleMealPlan.BelongsToAccount = account.ID
	mealPlan := createMealPlanForTest(t, ctx, exampleMealPlan, dbc)

	ingredient := createValidIngredientForTest(t, ctx, nil, dbc)
	measurmentUnit := createValidMeasurementUnitForTest(t, ctx, nil, dbc)

	exampleMealPlanGroceryListItem := fakes.BuildFakeMealPlanGroceryListItem()
	exampleMealPlanGroceryListItem.BelongsToMealPlan = mealPlan.ID
	exampleMealPlanGroceryListItem.Ingredient = *ingredient
	exampleMealPlanGroceryListItem.MeasurementUnit = *measurmentUnit
	exampleMealPlanGroceryListItem.PurchasedMeasurementUnit = measurmentUnit

	// create
	createdMealPlanGroceryListItems := []*types.MealPlanGroceryListItem{}
	createdMealPlanGroceryListItems = append(createdMealPlanGroceryListItems, createMealPlanGroceryListItemForTest(t, ctx, exampleMealPlanGroceryListItem, dbc))

	// update
	require.NoError(t, dbc.UpdateMealPlanGroceryListItem(ctx, createdMealPlanGroceryListItems[0]))

	// fetch as list
	mealPlanGroceryListItems, err := dbc.GetMealPlanGroceryListItemsForMealPlan(ctx, mealPlan.ID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, mealPlanGroceryListItems)
	assert.Len(t, mealPlanGroceryListItems.Data, len(createdMealPlanGroceryListItems))

	// a batch that fails partway through commits nothing. The retry regenerates the whole list
	// with fresh IDs and cannot tell which items it already wrote, so anything left behind by a
	// partial failure would be written a second time on the next run.
	goodInput := func() *types.MealPlanGroceryListItemDatabaseCreationInput {
		return &types.MealPlanGroceryListItemDatabaseCreationInput{
			ID:                     fake.BuildFakeID(),
			BelongsToMealPlan:      mealPlan.ID,
			ValidIngredientID:      ingredient.ID,
			ValidMeasurementUnitID: measurmentUnit.ID,
			Status:                 types.MealPlanGroceryListItemStatusNeeds,
			MinQuantityNeeded:      1,
		}
	}

	doomedItem := goodInput()
	// a nonexistent ingredient trips the foreign key, so this input fails after the one before it
	// has already been inserted.
	doomedItem.ValidIngredientID = fake.BuildFakeID()

	initialized, err := dbc.InitializeMealPlanGroceryList(ctx, mealPlan.ID, account.ID, []*types.MealPlanGroceryListItemDatabaseCreationInput{goodInput(), doomedItem})
	require.Error(t, err)
	assert.Nil(t, initialized)

	mealPlanGroceryListItems, err = dbc.GetMealPlanGroceryListItemsForMealPlan(ctx, mealPlan.ID, nil)
	require.NoError(t, err)
	assert.Len(t, mealPlanGroceryListItems.Data, len(createdMealPlanGroceryListItems), "the item preceding the failure must have rolled back with it")

	unmarkedMealPlan, err := dbc.GetMealPlan(ctx, mealPlan.ID, account.ID)
	require.NoError(t, err)
	assert.False(t, unmarkedMealPlan.GroceryListInitialized)

	// a batch that succeeds writes every item and the flag together.
	batch := []*types.MealPlanGroceryListItemDatabaseCreationInput{goodInput(), goodInput()}
	initialized, err = dbc.InitializeMealPlanGroceryList(ctx, mealPlan.ID, account.ID, batch)
	require.NoError(t, err)
	assert.Len(t, initialized, len(batch))
	createdMealPlanGroceryListItems = append(createdMealPlanGroceryListItems, initialized...)

	mealPlanGroceryListItems, err = dbc.GetMealPlanGroceryListItemsForMealPlan(ctx, mealPlan.ID, nil)
	require.NoError(t, err)
	assert.Len(t, mealPlanGroceryListItems.Data, len(createdMealPlanGroceryListItems))

	markedMealPlan, err := dbc.GetMealPlan(ctx, mealPlan.ID, account.ID)
	require.NoError(t, err)
	assert.True(t, markedMealPlan.GroceryListInitialized)

	// delete
	for _, mealPlanGroceryListItem := range createdMealPlanGroceryListItems {
		require.NoError(t, dbc.ArchiveMealPlanGroceryListItem(ctx, mealPlanGroceryListItem.ID))

		var exists bool
		exists, err = dbc.MealPlanGroceryListItemExists(ctx, mealPlanGroceryListItem.ID, account.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	}
}

func TestQuerier_MealPlanGroceryListItemExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan grocery list item ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleMealPlan := fakes.BuildFakeMealPlan()
		c := buildInertClientForTest(t)

		actual, err := c.MealPlanGroceryListItemExists(ctx, exampleMealPlan.ID, "")
		require.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_fleshOutMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.fleshOutMealPlanGroceryListItem(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan grocery list item ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleMealPlan := fakes.BuildFakeMealPlan()
		c := buildInertClientForTest(t)

		actual, err := c.GetMealPlanGroceryListItem(ctx, exampleMealPlan.ID, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateMealPlanGroceryListItem(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateMealPlanGroceryListItem(ctx, nil))
	})
}

func TestQuerier_ArchiveMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal plan grocery list item ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveMealPlanGroceryListItem(ctx, ""))
	})
}
