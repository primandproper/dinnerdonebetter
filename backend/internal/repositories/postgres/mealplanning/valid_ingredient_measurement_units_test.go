package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v11/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createValidIngredientMeasurementUnitForTest(t *testing.T, ctx context.Context, exampleValidIngredientMeasurementUnit *types.ValidIngredientMeasurementUnit, dbc *repository) *types.ValidIngredientMeasurementUnit {
	t.Helper()

	// create
	if exampleValidIngredientMeasurementUnit == nil {
		exampleValidIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
		exampleValidMeasurementUnit := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
		exampleValidIngredientMeasurementUnit = fakes.BuildFakeValidIngredientMeasurementUnit()
		exampleValidIngredientMeasurementUnit.Ingredient = *exampleValidIngredient
		exampleValidIngredientMeasurementUnit.MeasurementUnit = *exampleValidMeasurementUnit
	}

	dbInput := converters.ConvertValidIngredientMeasurementUnitToValidIngredientMeasurementUnitDatabaseCreationInput(exampleValidIngredientMeasurementUnit)

	created, err := dbc.CreateValidIngredientMeasurementUnit(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)
	exampleValidIngredientMeasurementUnit.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleValidIngredientMeasurementUnit, created)

	validIngredientMeasurementUnit, err := dbc.GetValidIngredientMeasurementUnit(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, validIngredientMeasurementUnit)
	exampleValidIngredientMeasurementUnit.CreatedAt = validIngredientMeasurementUnit.CreatedAt
	exampleValidIngredientMeasurementUnit.MeasurementUnit = validIngredientMeasurementUnit.MeasurementUnit
	exampleValidIngredientMeasurementUnit.Ingredient = validIngredientMeasurementUnit.Ingredient

	require.NoError(t, err)
	assert.Equal(t, validIngredientMeasurementUnit, exampleValidIngredientMeasurementUnit)

	return created
}

func TestQuerier_Integration_ValidIngredientMeasurementUnits(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	exampleValidIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
	exampleValidMeasurementUnit := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
	exampleValidIngredientMeasurementUnit := fakes.BuildFakeValidIngredientMeasurementUnit()
	exampleValidIngredientMeasurementUnit.MeasurementUnit = *exampleValidMeasurementUnit
	exampleValidIngredientMeasurementUnit.Ingredient = *exampleValidIngredient
	createdValidIngredientMeasurementUnits := []*types.ValidIngredientMeasurementUnit{}

	// create
	createdValidIngredientMeasurementUnits = append(createdValidIngredientMeasurementUnits, createValidIngredientMeasurementUnitForTest(t, ctx, exampleValidIngredientMeasurementUnit, dbc))

	// update
	updatedValidIngredientMeasurementUnit := fakes.BuildFakeValidIngredientMeasurementUnit()
	updatedValidIngredientMeasurementUnit.ID = createdValidIngredientMeasurementUnits[0].ID
	updatedValidIngredientMeasurementUnit.MeasurementUnit = createdValidIngredientMeasurementUnits[0].MeasurementUnit
	updatedValidIngredientMeasurementUnit.Ingredient = createdValidIngredientMeasurementUnits[0].Ingredient
	require.NoError(t, dbc.UpdateValidIngredientMeasurementUnit(ctx, updatedValidIngredientMeasurementUnit))

	// create more (each must have unique ingredient+unit per active row)
	for range exampleQuantity {
		extraIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
		input := fakes.BuildFakeValidIngredientMeasurementUnit()
		input.MeasurementUnit = createdValidIngredientMeasurementUnits[0].MeasurementUnit
		input.Ingredient = *extraIngredient
		createdValidIngredientMeasurementUnits = append(createdValidIngredientMeasurementUnits, createValidIngredientMeasurementUnitForTest(t, ctx, input, dbc))
	}

	// fetch as list
	validIngredientMeasurementUnits, err := dbc.GetValidIngredientMeasurementUnits(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, validIngredientMeasurementUnits.Data)
	assert.Len(t, validIngredientMeasurementUnits.Data, len(createdValidIngredientMeasurementUnits))

	forIngredient, err := dbc.GetValidIngredientMeasurementUnitsForIngredient(ctx, createdValidIngredientMeasurementUnits[0].Ingredient.ID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, forIngredient.Data)

	forMeasurementUnit, err := dbc.GetValidIngredientMeasurementUnitsForMeasurementUnit(ctx, createdValidIngredientMeasurementUnits[0].MeasurementUnit.ID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, forMeasurementUnit.Data)

	// delete
	for _, validIngredientMeasurementUnit := range createdValidIngredientMeasurementUnits {
		require.NoError(t, dbc.ArchiveValidIngredientMeasurementUnit(ctx, validIngredientMeasurementUnit.ID))

		var exists bool
		exists, err = dbc.ValidIngredientMeasurementUnitExists(ctx, validIngredientMeasurementUnit.ID)
		require.NoError(t, err)
		assert.False(t, exists)

		var y *types.ValidIngredientMeasurementUnit
		y, err = dbc.GetValidIngredientMeasurementUnit(ctx, validIngredientMeasurementUnit.ID)
		assert.Nil(t, y)
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_ValidIngredientMeasurementUnitExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient measurement unit MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.ValidIngredientMeasurementUnitExists(ctx, "")
		require.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient measurement unit MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidIngredientMeasurementUnit(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateValidIngredientMeasurementUnit(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateValidIngredientMeasurementUnit(ctx, nil))
	})
}

func TestQuerier_ArchiveValidIngredientMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient measurement unit MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveValidIngredientMeasurementUnit(ctx, ""))
	})
}

func TestQuerier_GetValidIngredientMeasurementUnitsByIDs(T *testing.T) {
	T.Parallel()

	T.Run("with empty list", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidIngredientMeasurementUnitsByIDs(ctx, []string{})
		require.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Empty(t, actual)
	})
}

func TestQuerier_Integration_GetValidIngredientMeasurementUnitsByIDs(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Create multiple valid ingredient measurement units
	created1 := createValidIngredientMeasurementUnitForTest(t, ctx, nil, dbc)
	created2 := createValidIngredientMeasurementUnitForTest(t, ctx, nil, dbc)
	created3 := createValidIngredientMeasurementUnitForTest(t, ctx, nil, dbc)

	// Test fetching by IDs
	ids := []string{created1.ID, created2.ID, created3.ID}
	results, err := dbc.GetValidIngredientMeasurementUnitsByIDs(ctx, ids)
	require.NoError(t, err)
	assert.Len(t, results, 3)
	assert.NotNil(t, results[created1.ID])
	assert.NotNil(t, results[created2.ID])
	assert.NotNil(t, results[created3.ID])

	// Test with partial IDs (some exist, some don't)
	partialIDs := []string{created1.ID, "nonexistent-id"}
	partialResults, err := dbc.GetValidIngredientMeasurementUnitsByIDs(ctx, partialIDs)
	require.NoError(t, err)
	assert.Len(t, partialResults, 1)
	assert.NotNil(t, partialResults[created1.ID])

	// Test with empty list
	emptyResults, err := dbc.GetValidIngredientMeasurementUnitsByIDs(ctx, []string{})
	require.NoError(t, err)
	assert.Empty(t, emptyResults)

	// Cleanup
	assert.NoError(t, dbc.ArchiveValidIngredientMeasurementUnit(ctx, created1.ID))
	assert.NoError(t, dbc.ArchiveValidIngredientMeasurementUnit(ctx, created2.ID))
	assert.NoError(t, dbc.ArchiveValidIngredientMeasurementUnit(ctx, created3.ID))
}

func TestQuerier_Integration_ValidIngredientMeasurementUnits_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Create different ingredients and measurement units for each item to ensure uniqueness
	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ValidIngredientMeasurementUnit]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "valid ingredient measurement unit",
		CreateItem: func(ctx context.Context, i int) *types.ValidIngredientMeasurementUnit {
			// Create unique ingredient and measurement unit for each item
			exampleValidIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
			exampleValidMeasurementUnit := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
			exampleValidIngredientMeasurementUnit := fakes.BuildFakeValidIngredientMeasurementUnit()
			exampleValidIngredientMeasurementUnit.Ingredient = *exampleValidIngredient
			exampleValidIngredientMeasurementUnit.MeasurementUnit = *exampleValidMeasurementUnit
			return createValidIngredientMeasurementUnitForTest(t, ctx, exampleValidIngredientMeasurementUnit, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientMeasurementUnit], error) {
			return dbc.GetValidIngredientMeasurementUnits(ctx, filter)
		},
		GetID: func(validIngredientMeasurementUnit *types.ValidIngredientMeasurementUnit) string {
			return validIngredientMeasurementUnit.ID
		},
		CleanupItem: func(ctx context.Context, validIngredientMeasurementUnit *types.ValidIngredientMeasurementUnit) error {
			return dbc.ArchiveValidIngredientMeasurementUnit(ctx, validIngredientMeasurementUnit.ID)
		},
	})
}
