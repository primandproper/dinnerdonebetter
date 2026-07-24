package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRecipeStepCompletionConditionForTest(t *testing.T, ctx context.Context, recipeID string, exampleRecipeStepCompletionCondition *types.RecipeStepCompletionCondition, dbc *repository) *types.RecipeStepCompletionCondition {
	t.Helper()

	// create
	require.NotNil(t, exampleRecipeStepCompletionCondition, "exampleRecipeStepCompletionCondition must not be nil")
	cond := exampleRecipeStepCompletionCondition

	recipe, getRecipeErr := dbc.GetRecipe(ctx, recipeID)
	require.NoError(t, getRecipeErr)
	require.NotNil(t, recipe)

	recipeStep, getRecipeStepErr := dbc.GetRecipeStep(ctx, recipeID, cond.BelongsToRecipeStep)
	require.NoError(t, getRecipeStepErr)
	require.NotNil(t, recipeStep)

	ingredientState, getIngredientStateErr := dbc.GetValidIngredientState(ctx, cond.IngredientState.ID)
	require.NoError(t, getIngredientStateErr)
	require.NotNil(t, ingredientState)

	dbInput := converters.ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionDatabaseCreationInput(cond)
	dbInput.Ingredients = []*types.RecipeStepCompletionConditionIngredientDatabaseCreationInput{
		{
			ID:                                     identifiers.New(),
			BelongsToRecipeStepCompletionCondition: dbInput.ID,
			RecipeStepIngredient:                   recipe.Steps[0].Ingredients[0].ID,
		},
	}

	created, err := dbc.CreateRecipeStepCompletionCondition(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	cond.CreatedAt = created.CreatedAt
	cond.IngredientState = created.IngredientState
	cond.Ingredients = created.Ingredients
	assert.Equal(t, cond, created)

	recipeStepCompletionCondition, err := dbc.GetRecipeStepCompletionCondition(ctx, recipe.ID, recipeStep.ID, created.ID)
	require.NoError(t, err)
	require.NotNil(t, recipeStepCompletionCondition)

	cond.CreatedAt = recipeStepCompletionCondition.CreatedAt
	cond.IngredientState = recipeStepCompletionCondition.IngredientState
	cond.Ingredients = recipeStepCompletionCondition.Ingredients

	require.Equal(t, cond, recipeStepCompletionCondition)

	return recipeStepCompletionCondition
}

func TestQuerier_Integration_RecipeStepCompletionConditions(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, container := buildDatabaseClientForTest(t)

	databaseURI, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, databaseURI)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	exampleRecipe := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
	createdRecipe := createRecipeForTest(t, ctx, exampleRecipe, dbc, true)
	exampleRecipeStep := createdRecipe.Steps[0]

	ingredientState := createValidIngredientStateForTest(t, ctx, nil, dbc)
	exampleRecipeStepCompletionCondition := fakes.BuildFakeRecipeStepCompletionCondition()
	exampleRecipeStepCompletionCondition.IngredientState = *ingredientState
	exampleRecipeStepCompletionCondition.BelongsToRecipeStep = exampleRecipeStep.ID
	exampleRecipeStepCompletionCondition.Ingredients = []*types.RecipeStepCompletionConditionIngredient{
		{
			ID:                                     identifiers.New(),
			BelongsToRecipeStepCompletionCondition: exampleRecipeStepCompletionCondition.ID,
			RecipeStepIngredient:                   exampleRecipeStep.Ingredients[0].ID,
		},
	}

	// create
	createdRecipeStepCompletionConditions := []*types.RecipeStepCompletionCondition{}
	createdRecipeStepCompletionConditions = append(createdRecipeStepCompletionConditions, createRecipeStepCompletionConditionForTest(t, ctx, exampleRecipe.ID, exampleRecipeStepCompletionCondition, dbc))

	// fetch as list
	recipeStepCompletionConditions, err := dbc.GetRecipeStepCompletionConditions(ctx, exampleRecipe.ID, createdRecipeStepCompletionConditions[0].BelongsToRecipeStep, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, recipeStepCompletionConditions.Data)
	assert.Equal(t, len(createdRecipeStepCompletionConditions), len(recipeStepCompletionConditions.Data))

	// delete
	for _, recipeStepCompletionCondition := range createdRecipeStepCompletionConditions {
		assert.NoError(t, dbc.ArchiveRecipeStepCompletionCondition(ctx, exampleRecipeStep.ID, recipeStepCompletionCondition.ID))

		var exists bool
		exists, err = dbc.RecipeStepCompletionConditionExists(ctx, exampleRecipe.ID, recipeStepCompletionCondition.BelongsToRecipeStep, recipeStepCompletionCondition.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.RecipeStepCompletionCondition
		y, err = dbc.GetRecipeStepCompletionCondition(ctx, exampleRecipe.ID, recipeStepCompletionCondition.BelongsToRecipeStep, recipeStepCompletionCondition.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_RecipeStepCompletionConditionExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepCompletionCondition := fakes.BuildFakeRecipeStepCompletionCondition()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepCompletionConditionExists(ctx, "", exampleRecipeStepID, exampleRecipeStepCompletionCondition.ID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepCompletionCondition := fakes.BuildFakeRecipeStepCompletionCondition()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepCompletionConditionExists(ctx, exampleRecipeID, "", exampleRecipeStepCompletionCondition.ID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid recipe step completion condition MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepCompletionConditionExists(ctx, exampleRecipeID, exampleRecipeStepID, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepCompletionCondition := fakes.BuildFakeRecipeStepCompletionCondition()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepCompletionCondition(ctx, "", exampleRecipeStepID, exampleRecipeStepCompletionCondition.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepCompletionCondition := fakes.BuildFakeRecipeStepCompletionCondition()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepCompletionCondition(ctx, exampleRecipeID, "", exampleRecipeStepCompletionCondition.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step completion condition MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetRecipeStepCompletionConditions(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()
		exampleRecipeStepID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepCompletionConditions(ctx, "", exampleRecipeStepID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateRecipeStepCompletionCondition(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestSQLQuerier_createRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.createRecipeStepCompletionCondition(ctx, c.writeDB, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateRecipeStepCompletionCondition(ctx, nil))
	})
}

func TestQuerier_ArchiveRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepCompletionCondition := fakes.BuildFakeRecipeStepCompletionCondition()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeStepCompletionCondition(ctx, "", exampleRecipeStepCompletionCondition.ID))
	})

	T.Run("with invalid recipe step completion condition MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeStepCompletionCondition(ctx, exampleRecipeStepID, ""))
	})
}

func TestQuerier_Integration_RecipeStepCompletionConditions_CursorBasedPagination(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, container := buildDatabaseClientForTest(t)

	databaseURI, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, databaseURI)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipeStruct := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
	// Clear the default completion conditions from the step so we start fresh
	for _, step := range recipeStruct.Steps {
		step.CompletionConditions = nil // Use nil instead of empty slice to match database behavior
	}
	recipe := createRecipeForTest(t, ctx, recipeStruct, dbc, false)
	recipeStep := recipe.Steps[0]

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.RecipeStepCompletionCondition]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "recipe step completion condition",
		CreateItem: func(ctx context.Context, i int) *types.RecipeStepCompletionCondition {
			ingredientState := createValidIngredientStateForTest(t, ctx, nil, dbc)
			recipeStepCompletionCondition := fakes.BuildFakeRecipeStepCompletionCondition()
			recipeStepCompletionCondition.BelongsToRecipeStep = recipeStep.ID
			recipeStepCompletionCondition.IngredientState = *ingredientState
			return createRecipeStepCompletionConditionForTest(t, ctx, recipe.ID, recipeStepCompletionCondition, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepCompletionCondition], error) {
			return dbc.GetRecipeStepCompletionConditions(ctx, recipe.ID, recipeStep.ID, filter)
		},
		GetID: func(recipeStepCompletionCondition *types.RecipeStepCompletionCondition) string {
			return recipeStepCompletionCondition.ID
		},
		CleanupItem: func(ctx context.Context, recipeStepCompletionCondition *types.RecipeStepCompletionCondition) error {
			return dbc.ArchiveRecipeStepCompletionCondition(ctx, recipeStep.ID, recipeStepCompletionCondition.ID)
		},
	})
}
