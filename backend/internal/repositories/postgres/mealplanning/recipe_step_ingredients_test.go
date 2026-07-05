package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v3/database/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRecipeStepIngredientForTest(t *testing.T, ctx context.Context, recipeID string, exampleRecipeStepIngredient *types.RecipeStepIngredient, dbc *repository) *types.RecipeStepIngredient {
	t.Helper()

	// create
	if exampleRecipeStepIngredient == nil {
		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
		exampleRecipe := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
		createdRecipe := createRecipeForTest(t, ctx, exampleRecipe, dbc, true)
		exampleRecipeStep := createdRecipe.Steps[0]

		exampleRecipeStepIngredient = fakes.BuildFakeRecipeStepIngredient()
		exampleRecipeStepIngredient.BelongsToRecipeStep = exampleRecipeStep.ID
	}
	dbInput := converters.ConvertRecipeStepIngredientToRecipeStepIngredientDatabaseCreationInput(exampleRecipeStepIngredient)

	created, err := dbc.CreateRecipeStepIngredient(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleRecipeStepIngredient.CreatedAt = created.CreatedAt
	exampleRecipeStepIngredient.Ingredient.CreatedAt = created.Ingredient.CreatedAt
	exampleRecipeStepIngredient.Ingredient = created.Ingredient
	exampleRecipeStepIngredient.MeasurementUnit = created.MeasurementUnit
	assert.Equal(t, exampleRecipeStepIngredient, created)

	recipeStepIngredient, err := dbc.GetRecipeStepIngredient(ctx, recipeID, exampleRecipeStepIngredient.BelongsToRecipeStep, exampleRecipeStepIngredient.ID)
	require.NotNil(t, recipeStepIngredient)
	require.NoError(t, err)

	exampleRecipeStepIngredient.CreatedAt = recipeStepIngredient.CreatedAt
	exampleRecipeStepIngredient.Ingredient.CreatedAt = recipeStepIngredient.Ingredient.CreatedAt
	exampleRecipeStepIngredient.Ingredient = recipeStepIngredient.Ingredient
	exampleRecipeStepIngredient.MeasurementUnit = recipeStepIngredient.MeasurementUnit

	assert.Equal(t, exampleRecipeStepIngredient, recipeStepIngredient)

	return created
}

func TestQuerier_Integration_RecipeStepIngredients(t *testing.T) {
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

	validMeasurementUnit := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
	validIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
	exampleRecipeStepIngredient := fakes.BuildFakeRecipeStepIngredient()
	exampleRecipeStepIngredient.Ingredient = validIngredient
	exampleRecipeStepIngredient.MeasurementUnit = *validMeasurementUnit
	exampleRecipeStepIngredient.BelongsToRecipeStep = exampleRecipeStep.ID
	// Set unique index - first ingredient from recipe creation has index 0, so start at 1
	exampleRecipeStepIngredient.Index = 1
	exampleRecipeStepIngredient.OptionIndex = 0
	createdRecipeStepIngredients := []*types.RecipeStepIngredient{
		exampleRecipeStep.Ingredients[0],
	}

	// create
	createdRecipeStepIngredients = append(createdRecipeStepIngredients, createRecipeStepIngredientForTest(t, ctx, exampleRecipe.ID, exampleRecipeStepIngredient, dbc))

	// create more
	for i := range exampleQuantity {
		validMeasurementUnit = createValidMeasurementUnitForTest(t, ctx, nil, dbc)
		validIngredient = createValidIngredientForTest(t, ctx, nil, dbc)
		input := fakes.BuildFakeRecipeStepIngredient()
		input.Ingredient = validIngredient
		input.MeasurementUnit = *validMeasurementUnit
		input.BelongsToRecipeStep = exampleRecipeStep.ID
		// Set unique index - start from 2 since 0 and 1 are already taken
		input.Index = uint16(i + 2)
		input.OptionIndex = 0
		createdRecipeStepIngredients = append(createdRecipeStepIngredients, createRecipeStepIngredientForTest(t, ctx, exampleRecipe.ID, input, dbc))
	}

	// fetch as list
	recipeStepIngredients, err := dbc.GetRecipeStepIngredients(ctx, exampleRecipe.ID, createdRecipeStepIngredients[0].BelongsToRecipeStep, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, recipeStepIngredients.Data)
	assert.Equal(t, len(createdRecipeStepIngredients), len(recipeStepIngredients.Data))

	// delete
	for _, recipeStepIngredient := range createdRecipeStepIngredients {
		assert.NoError(t, dbc.ArchiveRecipeStepIngredient(ctx, exampleRecipeStep.ID, recipeStepIngredient.ID))

		var exists bool
		exists, err = dbc.RecipeStepIngredientExists(ctx, exampleRecipe.ID, recipeStepIngredient.BelongsToRecipeStep, recipeStepIngredient.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.RecipeStepIngredient
		y, err = dbc.GetRecipeStepIngredient(ctx, exampleRecipe.ID, recipeStepIngredient.BelongsToRecipeStep, recipeStepIngredient.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_RecipeStepIngredientExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepIngredient := fakes.BuildFakeRecipeStepIngredient()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepIngredientExists(ctx, "", exampleRecipeStepID, exampleRecipeStepIngredient.ID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid recipe step ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepIngredient := fakes.BuildFakeRecipeStepIngredient()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepIngredientExists(ctx, exampleRecipeID, "", exampleRecipeStepIngredient.ID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid recipe step ingredient ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepIngredientExists(ctx, exampleRecipeID, exampleRecipeStepID, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe ID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepIngredient := fakes.BuildFakeRecipeStepIngredient()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepIngredient(ctx, "", exampleRecipeStepID, exampleRecipeStepIngredient.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step ID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepIngredient := fakes.BuildFakeRecipeStepIngredient()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepIngredient(ctx, exampleRecipeID, "", exampleRecipeStepIngredient.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step ingredient ID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_getRecipeStepIngredientsForRecipe(T *testing.T) {
	T.Parallel()

	T.Run("with missing recipe ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.getRecipeStepIngredientsForRecipe(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetRecipeStepIngredients(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe ID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()
		exampleRecipeStepID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepIngredients(ctx, "", exampleRecipeStepID, filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step ID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()
		exampleRecipeID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepIngredients(ctx, exampleRecipeID, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateRecipeStepIngredient(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestSQLQuerier_createRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.createRecipeStepIngredient(ctx, c.writeDB, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateRecipeStepIngredient(ctx, nil))
	})
}

func TestQuerier_ArchiveRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe step ID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepIngredient := fakes.BuildFakeRecipeStepIngredient()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeStepIngredient(ctx, "", exampleRecipeStepIngredient.ID))
	})

	T.Run("with invalid recipe step ingredient ID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeStepIngredient(ctx, exampleRecipeStepID, ""))
	})
}

func TestQuerier_Integration_RecipeStepIngredients_CursorBasedPagination(t *testing.T) {
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
	// Clear the default ingredients from the step so we start fresh
	for _, step := range recipeStruct.Steps {
		step.Ingredients = nil // Use nil instead of empty slice to match database behavior
	}
	recipe := createRecipeForTest(t, ctx, recipeStruct, dbc, false)
	recipeStep := recipe.Steps[0]

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.RecipeStepIngredient]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "recipe step ingredient",
		CreateItem: func(ctx context.Context, i int) *types.RecipeStepIngredient {
			ingredient := createValidIngredientForTest(t, ctx, nil, dbc)
			measurementUnit := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
			recipeStepIngredient := fakes.BuildFakeRecipeStepIngredient()
			recipeStepIngredient.BelongsToRecipeStep = recipeStep.ID
			recipeStepIngredient.Ingredient = ingredient
			recipeStepIngredient.MeasurementUnit = *measurementUnit
			// Set unique index for each ingredient to avoid constraint violations
			recipeStepIngredient.Index = uint16(i)
			recipeStepIngredient.OptionIndex = 0
			return createRecipeStepIngredientForTest(t, ctx, recipe.ID, recipeStepIngredient, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepIngredient], error) {
			return dbc.GetRecipeStepIngredients(ctx, recipe.ID, recipeStep.ID, filter)
		},
		GetID: func(recipeStepIngredient *types.RecipeStepIngredient) string {
			return recipeStepIngredient.ID
		},
		CleanupItem: func(ctx context.Context, recipeStepIngredient *types.RecipeStepIngredient) error {
			return dbc.ArchiveRecipeStepIngredient(ctx, recipeStep.ID, recipeStepIngredient.ID)
		},
	})
}
