package mealplanning

import (
	"database/sql"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v9/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildMealListItemForTest(listID, mealID string) *mealplanning.MealListItemDatabaseCreationInput {
	return &mealplanning.MealListItemDatabaseCreationInput{
		ID:                identifiers.New(),
		MealID:            mealID,
		Notes:             "note1",
		BelongsToMealList: listID,
	}
}

func TestIntegration_MealListItems(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)
	defer func() {
	}()

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)
	meal := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)
	listInput := buildMealListForTest(user.ID)
	createdList, err := dbc.CreateMealList(ctx, listInput)
	require.NoError(t, err)

	itemInput := buildMealListItemForTest(createdList.ID, meal.ID)
	createdItem, err := dbc.CreateMealListItem(ctx, itemInput)
	require.NoError(t, err)
	require.Equal(t, itemInput.ID, createdItem.ID)

	items, err := dbc.GetMealListItems(ctx, createdList.ID, user.ID, nil)
	require.NoError(t, err)
	require.Len(t, items.Data, 1)
	assert.Equal(t, meal.ID, items.Data[0].Meal.ID)

	require.NoError(t, dbc.ArchiveMealListItem(ctx, createdItem.ID, createdList.ID))

	afterArchive, err := dbc.GetMealListItems(ctx, createdList.ID, user.ID, nil)
	require.NoError(t, err)
	assert.Len(t, afterArchive.Data, 0)

	err = dbc.ArchiveMealListItem(ctx, createdItem.ID, createdList.ID)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestIntegration_MealExistsInMealList(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)
	defer func() {
	}()

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)
	meal := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)
	listInput := buildMealListForTest(user.ID)
	createdList, err := dbc.CreateMealList(ctx, listInput)
	require.NoError(t, err)

	itemInput := buildMealListItemForTest(createdList.ID, meal.ID)
	_, err = dbc.CreateMealListItem(ctx, itemInput)
	require.NoError(t, err)

	exists, err := dbc.MealExistsInMealList(ctx, createdList.ID, meal.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = dbc.MealExistsInMealList(ctx, createdList.ID, "nonexistent-meal-id")
	require.NoError(t, err)
	assert.False(t, exists)

	anotherMeal := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)
	exists, err = dbc.MealExistsInMealList(ctx, createdList.ID, anotherMeal.ID)
	require.NoError(t, err)
	assert.False(t, exists)
}
