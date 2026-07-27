package mealplanning

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v7/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createValidMeasurementUnitForTest(t *testing.T, ctx context.Context, exampleValidMeasurementUnit *types.ValidMeasurementUnit, dbc *repository) *types.ValidMeasurementUnit {
	t.Helper()

	// create
	if exampleValidMeasurementUnit == nil {
		exampleValidMeasurementUnit = fakes.BuildFakeValidMeasurementUnit()
	}
	dbInput := converters.ConvertValidMeasurementUnitToValidMeasurementUnitDatabaseCreationInput(exampleValidMeasurementUnit)

	created, err := dbc.CreateValidMeasurementUnit(ctx, dbInput)
	exampleValidMeasurementUnit.CreatedAt = created.CreatedAt
	assert.NoError(t, err)
	assert.Equal(t, exampleValidMeasurementUnit, created)

	validMeasurementUnit, err := dbc.GetValidMeasurementUnit(ctx, created.ID)
	exampleValidMeasurementUnit.CreatedAt = validMeasurementUnit.CreatedAt

	assert.NoError(t, err)
	assert.Equal(t, validMeasurementUnit, exampleValidMeasurementUnit)

	return validMeasurementUnit
}

func TestQuerier_Integration_ValidMeasurementUnits(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	exampleValidMeasurementUnit := fakes.BuildFakeValidMeasurementUnit()
	createdValidMeasurementUnits := []*types.ValidMeasurementUnit{}

	// create
	createdValidMeasurementUnits = append(createdValidMeasurementUnits, createValidMeasurementUnitForTest(t, ctx, exampleValidMeasurementUnit, dbc))

	// update
	updatedValidMeasurementUnit := fakes.BuildFakeValidMeasurementUnit()
	updatedValidMeasurementUnit.ID = createdValidMeasurementUnits[0].ID
	assert.NoError(t, dbc.UpdateValidMeasurementUnit(ctx, updatedValidMeasurementUnit))

	// create more
	for i := range exampleQuantity {
		input := fakes.BuildFakeValidMeasurementUnit()
		input.Name = fmt.Sprintf("%s %d", updatedValidMeasurementUnit.Name, i)
		createdValidMeasurementUnits = append(createdValidMeasurementUnits, createValidMeasurementUnitForTest(t, ctx, input, dbc))
	}

	// fetch as list
	validMeasurementUnits, err := dbc.GetValidMeasurementUnits(ctx, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, validMeasurementUnits.Data)
	assert.GreaterOrEqual(t, len(validMeasurementUnits.Data), len(createdValidMeasurementUnits))

	// fetch as list of IDs
	validMeasurementUnitIDs := []string{}
	for _, validMeasurementUnit := range createdValidMeasurementUnits {
		validMeasurementUnitIDs = append(validMeasurementUnitIDs, validMeasurementUnit.ID)
	}

	byIDs, err := dbc.GetValidMeasurementUnitsWithIDs(ctx, validMeasurementUnitIDs)
	assert.NoError(t, err)
	assert.Subset(t, validMeasurementUnits.Data, byIDs)

	// fetch via name search
	byName, err := dbc.SearchForValidMeasurementUnits(ctx, updatedValidMeasurementUnit.Name, nil)
	assert.NoError(t, err)
	assert.Subset(t, validMeasurementUnits.Data, byName.Data)

	random, err := dbc.GetRandomValidMeasurementUnit(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, random)

	needingIndexing, err := dbc.GetValidMeasurementUnitIDsThatNeedSearchIndexing(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, needingIndexing)

	ingredient := createValidIngredientForTest(t, ctx, nil, dbc)
	ingredientMeasurementUnit := fakes.BuildFakeValidIngredientMeasurementUnit()
	ingredientMeasurementUnit.MeasurementUnit = *createdValidMeasurementUnits[0]
	ingredientMeasurementUnit.Ingredient = *ingredient
	dbInput := converters.ConvertValidIngredientMeasurementUnitToValidIngredientMeasurementUnitDatabaseCreationInput(ingredientMeasurementUnit)
	createdIngredientMeasurementUnit, err := dbc.CreateValidIngredientMeasurementUnit(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, createdIngredientMeasurementUnit)

	validIngredientMeasurementUnits, err := dbc.ValidMeasurementUnitsForIngredientID(ctx, ingredient.ID, nil)
	require.NoError(t, err)
	require.NotEmpty(t, validIngredientMeasurementUnits)

	// delete
	for _, validMeasurementUnit := range createdValidMeasurementUnits {
		assert.NoError(t, dbc.MarkValidMeasurementUnitAsIndexed(ctx, validMeasurementUnit.ID))
		assert.NoError(t, dbc.ArchiveValidMeasurementUnit(ctx, validMeasurementUnit.ID))

		var exists bool
		exists, err = dbc.ValidMeasurementUnitExists(ctx, validMeasurementUnit.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.ValidMeasurementUnit
		y, err = dbc.GetValidMeasurementUnit(ctx, validMeasurementUnit.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_ValidMeasurementUnitExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.ValidMeasurementUnitExists(ctx, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidMeasurementUnit(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_SearchForValidMeasurementUnits(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.SearchForValidMeasurementUnits(ctx, "", nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_ValidMeasurementUnitsForIngredientID(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		filter := filtering.DefaultQueryFilter()
		c := buildInertClientForTest(t)

		actual, err := c.ValidMeasurementUnitsForIngredientID(ctx, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateValidMeasurementUnit(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateValidMeasurementUnit(ctx, nil))
	})
}

func TestQuerier_ArchiveValidMeasurementUnit(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveValidMeasurementUnit(ctx, ""))
	})
}

func TestQuerier_MarkValidMeasurementUnitAsIndexed(T *testing.T) {
	T.Parallel()

	T.Run("with invalid MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.MarkValidMeasurementUnitAsIndexed(ctx, ""))
	})
}

func TestQuerier_Integration_ValidMeasurementUnits_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ValidMeasurementUnit]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "valid measurement unit",
		CreateItem: func(ctx context.Context, i int) *types.ValidMeasurementUnit {
			validMeasurementUnit := fakes.BuildFakeValidMeasurementUnit()
			validMeasurementUnit.Name = fmt.Sprintf("Valid Measurement Unit %02d", i)
			return createValidMeasurementUnitForTest(t, ctx, validMeasurementUnit, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnit], error) {
			return dbc.GetValidMeasurementUnits(ctx, filter)
		},
		GetID: func(validMeasurementUnit *types.ValidMeasurementUnit) string {
			return validMeasurementUnit.ID
		},
		CleanupItem: func(ctx context.Context, validMeasurementUnit *types.ValidMeasurementUnit) error {
			return dbc.ArchiveValidMeasurementUnit(ctx, validMeasurementUnit.ID)
		},
	})
}
