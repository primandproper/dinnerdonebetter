package mealplanning

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildMealForIntegrationTest(userID string, recipe *types.Recipe) *types.Meal {
	exampleMeal := fakes.BuildFakeMeal()
	exampleMeal.CreatedByUser = userID
	exampleMeal.Components = []*types.MealComponent{
		{
			ComponentType: types.MealComponentTypesMain,
			Recipe:        *recipe,
			RecipeScale:   1,
		},
	}

	return exampleMeal
}

func createMealForTest(t *testing.T, ctx context.Context, exampleMeal *types.Meal, dbc *repository) *types.Meal {
	t.Helper()

	// create
	if exampleMeal == nil {
		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
		recipe := createRecipeForTest(t, ctx, nil, dbc, true)
		exampleMeal = buildMealForIntegrationTest(user.ID, recipe)
	}
	dbInput := converters.ConvertMealToMealDatabaseCreationInput(exampleMeal)

	created, err := dbc.CreateMeal(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	meal, err := dbc.GetMeal(ctx, created.ID)
	assert.NoError(t, err)
	require.NotNil(t, meal)

	return meal
}

func TestQuerier_Integration_Meals(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)

	exampleMeal := buildMealForIntegrationTest(user.ID, recipe)
	createdMeals := []*types.Meal{}

	// create
	createdMeals = append(createdMeals, createMealForTest(t, ctx, exampleMeal, dbc))

	// create more
	for i := range exampleQuantity {
		input := buildMealForIntegrationTest(user.ID, recipe)
		input.Name = fmt.Sprintf("%s %d", exampleMeal.Name, i)
		createdMeals = append(createdMeals, createMealForTest(t, ctx, input, dbc))
	}

	// fetch as list
	meals, err := dbc.GetMeals(ctx, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, meals.Data)
	assert.Equal(t, len(createdMeals), len(meals.Data))

	results, err := dbc.GetMealsWithIDs(ctx, []string{createdMeals[0].ID})
	assert.NoError(t, err)
	assert.NotEmpty(t, results)

	ids, err := dbc.GetMealIDsThatNeedSearchIndexing(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, ids)

	// delete
	for _, meal := range createdMeals {
		assert.NoError(t, dbc.ArchiveMeal(ctx, meal.ID, user.ID))

		var exists bool
		exists, err = dbc.MealExists(ctx, meal.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.Meal
		y, err = dbc.GetMeal(ctx, meal.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_Integration_GetMealsWithIDs(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)
	defer func() {
	}()

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)

	meal1 := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)
	meal2 := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)

	results, err := dbc.GetMealsWithIDs(ctx, []string{meal1.ID, meal2.ID})
	require.NoError(t, err)
	require.Len(t, results, 2)

	found := map[string]*types.Meal{}
	for _, m := range results {
		found[m.ID] = m
		require.NotEmpty(t, m.Components)
	}

	assert.Contains(t, found, meal1.ID)
	assert.Contains(t, found, meal2.ID)
}

func TestQuerier_Integration_FindMealWithSameComponents(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)
	defer func() {
	}()

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)
	recipe2 := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)

	meal := createMealForTest(t, ctx, buildMealForIntegrationTest(user.ID, recipe), dbc)
	input := converters.ConvertMealToMealCreationRequestInput(meal)

	// Same name and components: should find existing meal
	found, err := dbc.FindMealWithSameComponents(ctx, user.ID, input)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, meal.ID, found.ID)

	// Different name: should not find (returns ErrNoMatchingMeal)
	inputDifferentName := converters.ConvertMealToMealCreationRequestInput(meal)
	inputDifferentName.Name = "Different Name"
	found, err = dbc.FindMealWithSameComponents(ctx, user.ID, inputDifferentName)
	require.ErrorIs(t, err, types.ErrNoMatchingMeal)
	assert.Nil(t, found)

	// Different components: should not find (returns ErrNoMatchingMeal)
	inputDifferentComponents := converters.ConvertMealToMealCreationRequestInput(meal)
	inputDifferentComponents.Components = []*types.MealComponentCreationRequestInput{
		{RecipeID: recipe2.ID, ComponentType: types.MealComponentTypesMain, RecipeScale: 1.0},
	}
	found, err = dbc.FindMealWithSameComponents(ctx, user.ID, inputDifferentComponents)
	require.ErrorIs(t, err, types.ErrNoMatchingMeal)
	assert.Nil(t, found)
}

func TestQuerier_MealExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.MealExists(ctx, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetMeal(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetMeal(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateMeal(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateMeal(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateMealRecipe(T *testing.T) {
	T.Parallel()

	T.Run("with missing meal ID", func(t *testing.T) {
		t.Parallel()

		exampleMeal := fakes.BuildFakeMeal()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		exampleInput := converters.ConvertMealComponentToMealComponentDatabaseCreationInput(exampleMeal.Components[0])

		err := c.CreateMealComponent(ctx, c.writeDB, "", exampleInput)
		assert.Error(t, err)
	})

	T.Run("with missing input", func(t *testing.T) {
		t.Parallel()

		exampleMeal := fakes.BuildFakeMeal()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.CreateMealComponent(ctx, c.writeDB, exampleMeal.ID, nil)
		assert.Error(t, err)
	})
}

func TestQuerier_ArchiveMeal(T *testing.T) {
	T.Parallel()

	T.Run("with invalid meal ID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveMeal(ctx, "", exampleAccountID))
	})

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		exampleMeal := fakes.BuildFakeMeal()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveMeal(ctx, exampleMeal.ID, ""))
	})
}

func TestQuerier_MarkMealAsIndexed(T *testing.T) {
	T.Parallel()

	T.Run("with invalid ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.MarkMealAsIndexed(ctx, ""))
	})
}

func TestQuerier_Integration_Meals_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipe := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.Meal]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "meal",
		CreateItem: func(ctx context.Context, i int) *types.Meal {
			meal := buildMealForIntegrationTest(user.ID, recipe)
			meal.Name = fmt.Sprintf("Meal %02d", i)
			return createMealForTest(t, ctx, meal, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Meal], error) {
			return dbc.GetMeals(ctx, filter)
		},
		GetID: func(meal *types.Meal) string {
			return meal.ID
		},
		CleanupItem: func(ctx context.Context, meal *types.Meal) error {
			return dbc.ArchiveMeal(ctx, meal.ID, user.ID)
		},
	})
}
