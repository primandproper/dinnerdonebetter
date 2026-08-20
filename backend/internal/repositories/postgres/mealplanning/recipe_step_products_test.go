package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRecipeStepProductForTest(t *testing.T, ctx context.Context, recipeID string, exampleRecipeStepProduct *types.RecipeStepProduct, dbc *repository) *types.RecipeStepProduct {
	t.Helper()

	// create
	if exampleRecipeStepProduct == nil {
		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
		exampleRecipe := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
		createdRecipe := createRecipeForTest(t, ctx, exampleRecipe, dbc, true)
		exampleRecipeStep := createdRecipe.Steps[0]

		exampleRecipeStepProduct = fakes.BuildFakeRecipeStepProduct()
		exampleRecipeStepProduct.BelongsToRecipeStep = exampleRecipeStep.ID
	}
	dbInput := converters.ConvertRecipeStepProductToRecipeStepProductDatabaseCreationInput(exampleRecipeStepProduct)

	created, err := dbc.CreateRecipeStepProduct(ctx, recipeID, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)

	exampleRecipeStepProduct.CreatedAt = created.CreatedAt
	exampleRecipeStepProduct.MeasurementUnit = created.MeasurementUnit
	assert.Equal(t, exampleRecipeStepProduct, created)

	recipeStepProduct, err := dbc.GetRecipeStepProduct(ctx, recipeID, exampleRecipeStepProduct.BelongsToRecipeStep, exampleRecipeStepProduct.ID)

	exampleRecipeStepProduct.CreatedAt = recipeStepProduct.CreatedAt
	exampleRecipeStepProduct.MeasurementUnit = recipeStepProduct.MeasurementUnit

	require.Equal(t, exampleRecipeStepProduct, recipeStepProduct)

	require.NoError(t, err)
	assert.Equal(t, recipeStepProduct, exampleRecipeStepProduct)

	return created
}

func TestQuerier_Integration_RecipeStepProducts(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	exampleRecipe := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
	createdRecipe := createRecipeForTest(t, ctx, exampleRecipe, dbc, true)
	exampleRecipeStep := createdRecipe.Steps[0]

	validMeasurementUnit := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
	exampleRecipeStepProduct := fakes.BuildFakeRecipeStepProduct()
	exampleRecipeStepProduct.MeasurementUnit = validMeasurementUnit
	exampleRecipeStepProduct.BelongsToRecipeStep = exampleRecipeStep.ID
	createdRecipeStepProducts := []*types.RecipeStepProduct{
		exampleRecipeStep.Products[0],
	}

	// create
	createdRecipeStepProducts = append(createdRecipeStepProducts, createRecipeStepProductForTest(t, ctx, exampleRecipe.ID, exampleRecipeStepProduct, dbc))

	// create more
	for range exampleQuantity {
		validMeasurementUnit = createValidMeasurementUnitForTest(t, ctx, nil, dbc)
		input := fakes.BuildFakeRecipeStepProduct()
		input.MeasurementUnit = validMeasurementUnit
		input.BelongsToRecipeStep = exampleRecipeStep.ID
		createdRecipeStepProducts = append(createdRecipeStepProducts, createRecipeStepProductForTest(t, ctx, exampleRecipe.ID, input, dbc))
	}

	// fetch as list
	recipeStepProducts, err := dbc.GetRecipeStepProducts(ctx, exampleRecipe.ID, createdRecipeStepProducts[0].BelongsToRecipeStep, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, recipeStepProducts.Data)
	assert.Len(t, recipeStepProducts.Data, len(createdRecipeStepProducts))

	// delete
	for _, recipeStepProduct := range createdRecipeStepProducts {
		require.NoError(t, dbc.ArchiveRecipeStepProduct(ctx, createdRecipe.ID, exampleRecipeStep.ID, recipeStepProduct.ID))

		var exists bool
		exists, err = dbc.RecipeStepProductExists(ctx, exampleRecipe.ID, recipeStepProduct.BelongsToRecipeStep, recipeStepProduct.ID)
		require.NoError(t, err)
		assert.False(t, exists)

		var y *types.RecipeStepProduct
		y, err = dbc.GetRecipeStepProduct(ctx, exampleRecipe.ID, recipeStepProduct.BelongsToRecipeStep, recipeStepProduct.ID)
		assert.Nil(t, y)
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_RecipeStepProductExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeStepID := fake.BuildFakeID()
		exampleRecipeStepProduct := fakes.BuildFakeRecipeStepProduct()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepProductExists(ctx, "", exampleRecipeStepID, exampleRecipeStepProduct.ID)
		require.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepProduct := fakes.BuildFakeRecipeStepProduct()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepProductExists(ctx, exampleRecipeID, "", exampleRecipeStepProduct.ID)
		require.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid recipe step product MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepProductExists(ctx, exampleRecipeID, exampleRecipeStepID, "")
		require.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepID := fake.BuildFakeID()
		exampleRecipeStepProduct := fakes.BuildFakeRecipeStepProduct()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepProduct(ctx, "", exampleRecipeStepID, exampleRecipeStepProduct.ID)
		require.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepProduct := fakes.BuildFakeRecipeStepProduct()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepProduct(ctx, exampleRecipeID, "", exampleRecipeStepProduct.ID)
		require.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step product MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepProduct(ctx, exampleRecipeID, exampleRecipeStepID, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_getRecipeStepProductsForRecipe(T *testing.T) {
	T.Parallel()

	T.Run("with missing recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.getRecipeStepProductsForRecipe(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetRecipeStepProducts(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()
		exampleRecipeStepID := fake.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepProducts(ctx, "", exampleRecipeStepID, filter)
		require.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()
		exampleRecipeID := fake.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStepProducts(ctx, exampleRecipeID, "", filter)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateRecipeStepProduct(ctx, fake.BuildFakeID(), nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateRecipeStepProduct(ctx, fake.BuildFakeID(), nil))
	})
}

func TestQuerier_ArchiveRecipeStepProduct(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepProduct := fakes.BuildFakeRecipeStepProduct()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeStepProduct(ctx, fake.BuildFakeID(), "", exampleRecipeStepProduct.ID))
	})

	T.Run("with invalid recipe step product MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStepID := fake.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeStepProduct(ctx, fake.BuildFakeID(), exampleRecipeStepID, ""))
	})
}

func TestQuerier_Integration_RecipeStepProducts_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipeStruct := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
	// Clear the default products from the step so we start fresh
	for _, step := range recipeStruct.Steps {
		step.Products = nil // Use nil instead of empty slice to match database behavior
	}
	recipe := createRecipeForTest(t, ctx, recipeStruct, dbc, false)
	recipeStep := recipe.Steps[0]

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.RecipeStepProduct]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "recipe step product",
		CreateItem: func(ctx context.Context, i int) *types.RecipeStepProduct {
			measurementUnit := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
			recipeStepProduct := fakes.BuildFakeRecipeStepProduct()
			recipeStepProduct.BelongsToRecipeStep = recipeStep.ID
			recipeStepProduct.MeasurementUnit = measurementUnit
			return createRecipeStepProductForTest(t, ctx, recipe.ID, recipeStepProduct, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepProduct], error) {
			return dbc.GetRecipeStepProducts(ctx, recipe.ID, recipeStep.ID, filter)
		},
		GetID: func(recipeStepProduct *types.RecipeStepProduct) string {
			return recipeStepProduct.ID
		},
		CleanupItem: func(ctx context.Context, recipeStepProduct *types.RecipeStepProduct) error {
			return dbc.ArchiveRecipeStepProduct(ctx, recipe.ID, recipeStep.ID, recipeStepProduct.ID)
		},
	})
}
