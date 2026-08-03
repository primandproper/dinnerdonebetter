package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createValidIngredientPreparationForTest(t *testing.T, ctx context.Context, exampleValidIngredientPreparation *types.ValidIngredientPreparation, dbc *repository) *types.ValidIngredientPreparation {
	t.Helper()

	// create
	if exampleValidIngredientPreparation == nil {
		exampleValidIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
		exampleValidPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
		exampleValidIngredientPreparation = fakes.BuildFakeValidIngredientPreparation()
		exampleValidIngredientPreparation.Ingredient = *exampleValidIngredient
		exampleValidIngredientPreparation.Preparation = *exampleValidPreparation
	}

	dbInput := converters.ConvertValidIngredientPreparationToValidIngredientPreparationDatabaseCreationInput(exampleValidIngredientPreparation)

	created, err := dbc.CreateValidIngredientPreparation(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)
	exampleValidIngredientPreparation.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleValidIngredientPreparation, created)

	validIngredientPreparation, err := dbc.GetValidIngredientPreparation(ctx, created.ID)
	require.NotNil(t, validIngredientPreparation)
	exampleValidIngredientPreparation.CreatedAt = validIngredientPreparation.CreatedAt
	exampleValidIngredientPreparation.Preparation = validIngredientPreparation.Preparation
	exampleValidIngredientPreparation.Ingredient = validIngredientPreparation.Ingredient

	assert.NoError(t, err)
	assert.Equal(t, validIngredientPreparation, exampleValidIngredientPreparation)

	return created
}

func TestQuerier_Integration_ValidIngredientPreparations(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	exampleValidIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
	exampleValidPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
	exampleValidIngredientPreparation := fakes.BuildFakeValidIngredientPreparation()
	exampleValidIngredientPreparation.Preparation = *exampleValidPreparation
	exampleValidIngredientPreparation.Ingredient = *exampleValidIngredient
	createdValidIngredientPreparations := []*types.ValidIngredientPreparation{}

	// create
	createdValidIngredientPreparations = append(createdValidIngredientPreparations, createValidIngredientPreparationForTest(t, ctx, exampleValidIngredientPreparation, dbc))

	// update
	updatedValidIngredientPreparation := fakes.BuildFakeValidIngredientPreparation()
	updatedValidIngredientPreparation.ID = createdValidIngredientPreparations[0].ID
	updatedValidIngredientPreparation.Preparation = createdValidIngredientPreparations[0].Preparation
	updatedValidIngredientPreparation.Ingredient = createdValidIngredientPreparations[0].Ingredient
	assert.NoError(t, dbc.UpdateValidIngredientPreparation(ctx, updatedValidIngredientPreparation))

	// create more (each must have unique prep+ingredient per active row)
	for range exampleQuantity {
		extraPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
		input := fakes.BuildFakeValidIngredientPreparation()
		input.Preparation = *extraPreparation
		input.Ingredient = createdValidIngredientPreparations[0].Ingredient
		createdValidIngredientPreparations = append(createdValidIngredientPreparations, createValidIngredientPreparationForTest(t, ctx, input, dbc))
	}

	// fetch as list
	validIngredientPreparations, err := dbc.GetValidIngredientPreparations(ctx, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, validIngredientPreparations.Data)
	assert.Len(t, validIngredientPreparations.Data, len(createdValidIngredientPreparations))

	forIngredient, err := dbc.GetValidIngredientPreparationsForIngredient(ctx, createdValidIngredientPreparations[0].Ingredient.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, forIngredient.Data)

	forPreparation, err := dbc.GetValidIngredientPreparationsForPreparation(ctx, createdValidIngredientPreparations[0].Preparation.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, forPreparation.Data)

	// delete
	for _, validIngredientPreparation := range createdValidIngredientPreparations {
		assert.NoError(t, dbc.ArchiveValidIngredientPreparation(ctx, validIngredientPreparation.ID))

		var exists bool
		exists, err = dbc.ValidIngredientPreparationExists(ctx, validIngredientPreparation.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.ValidIngredientPreparation
		y, err = dbc.GetValidIngredientPreparation(ctx, validIngredientPreparation.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_ValidIngredientPreparationExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient preparation MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.ValidIngredientPreparationExists(ctx, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient preparation MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidIngredientPreparation(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetValidIngredientPreparationsForIngredient(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient preparation MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidIngredientPreparationsForIngredient(ctx, "", nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetValidIngredientPreparationsForPreparation(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient preparation MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidIngredientPreparationsForPreparation(ctx, "", nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateValidIngredientPreparation(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateValidIngredientPreparation(ctx, nil))
	})
}

func TestQuerier_ArchiveValidIngredientPreparation(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient preparation MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveValidIngredientPreparation(ctx, ""))
	})
}

func TestQuerier_GetValidIngredientPreparationsByIDs(T *testing.T) {
	T.Parallel()

	T.Run("with empty list", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidIngredientPreparationsByIDs(ctx, []string{})
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Empty(t, actual)
	})
}

func TestQuerier_Integration_GetValidIngredientPreparationsByIDs(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Create multiple valid ingredient preparations
	created1 := createValidIngredientPreparationForTest(t, ctx, nil, dbc)
	created2 := createValidIngredientPreparationForTest(t, ctx, nil, dbc)
	created3 := createValidIngredientPreparationForTest(t, ctx, nil, dbc)

	// Test fetching by IDs
	ids := []string{created1.ID, created2.ID, created3.ID}
	results, err := dbc.GetValidIngredientPreparationsByIDs(ctx, ids)
	assert.NoError(t, err)
	assert.Len(t, results, 3)
	assert.NotNil(t, results[created1.ID])
	assert.NotNil(t, results[created2.ID])
	assert.NotNil(t, results[created3.ID])

	// Test with partial IDs (some exist, some don't)
	partialIDs := []string{created1.ID, "nonexistent-id"}
	partialResults, err := dbc.GetValidIngredientPreparationsByIDs(ctx, partialIDs)
	assert.NoError(t, err)
	assert.Len(t, partialResults, 1)
	assert.NotNil(t, partialResults[created1.ID])

	// Test with empty list
	emptyResults, err := dbc.GetValidIngredientPreparationsByIDs(ctx, []string{})
	assert.NoError(t, err)
	assert.Empty(t, emptyResults)

	// Cleanup
	assert.NoError(t, dbc.ArchiveValidIngredientPreparation(ctx, created1.ID))
	assert.NoError(t, dbc.ArchiveValidIngredientPreparation(ctx, created2.ID))
	assert.NoError(t, dbc.ArchiveValidIngredientPreparation(ctx, created3.ID))
}

func TestQuerier_Integration_ValidIngredientPreparations_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Create different ingredients and preparations for each item to ensure uniqueness
	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ValidIngredientPreparation]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "valid ingredient preparation",
		CreateItem: func(ctx context.Context, i int) *types.ValidIngredientPreparation {
			// Create unique ingredient and preparation for each item
			exampleValidIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
			exampleValidPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
			exampleValidIngredientPreparation := fakes.BuildFakeValidIngredientPreparation()
			exampleValidIngredientPreparation.Ingredient = *exampleValidIngredient
			exampleValidIngredientPreparation.Preparation = *exampleValidPreparation
			return createValidIngredientPreparationForTest(t, ctx, exampleValidIngredientPreparation, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientPreparation], error) {
			return dbc.GetValidIngredientPreparations(ctx, filter)
		},
		GetID: func(validIngredientPreparation *types.ValidIngredientPreparation) string {
			return validIngredientPreparation.ID
		},
		CleanupItem: func(ctx context.Context, validIngredientPreparation *types.ValidIngredientPreparation) error {
			return dbc.ArchiveValidIngredientPreparation(ctx, validIngredientPreparation.ID)
		},
	})
}
