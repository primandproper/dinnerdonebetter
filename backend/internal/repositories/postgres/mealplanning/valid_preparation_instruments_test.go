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

func createValidPreparationInstrumentForTest(t *testing.T, ctx context.Context, exampleValidPreparationInstrument *types.ValidPreparationInstrument, dbc *repository) *types.ValidPreparationInstrument {
	t.Helper()

	// create
	if exampleValidPreparationInstrument == nil {
		exampleValidInstrument := createValidInstrumentForTest(t, ctx, nil, dbc)
		exampleValidPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
		exampleValidPreparationInstrument = fakes.BuildFakeValidPreparationInstrument()
		exampleValidPreparationInstrument.Instrument = *exampleValidInstrument
		exampleValidPreparationInstrument.Preparation = *exampleValidPreparation
	}

	dbInput := converters.ConvertValidPreparationInstrumentToValidPreparationInstrumentDatabaseCreationInput(exampleValidPreparationInstrument)

	created, err := dbc.CreateValidPreparationInstrument(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)
	exampleValidPreparationInstrument.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleValidPreparationInstrument, created)

	validPreparationInstrument, err := dbc.GetValidPreparationInstrument(ctx, created.ID)
	exampleValidPreparationInstrument.CreatedAt = validPreparationInstrument.CreatedAt
	exampleValidPreparationInstrument.Preparation = validPreparationInstrument.Preparation
	exampleValidPreparationInstrument.Instrument = validPreparationInstrument.Instrument

	require.NoError(t, err)
	assert.Equal(t, validPreparationInstrument, exampleValidPreparationInstrument)

	return created
}

func TestQuerier_Integration_ValidPreparationInstruments(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	exampleValidInstrument := createValidInstrumentForTest(t, ctx, nil, dbc)
	exampleValidPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
	exampleValidPreparationInstrument := fakes.BuildFakeValidPreparationInstrument()
	exampleValidPreparationInstrument.Preparation = *exampleValidPreparation
	exampleValidPreparationInstrument.Instrument = *exampleValidInstrument
	createdValidPreparationInstruments := []*types.ValidPreparationInstrument{}

	// create
	createdValidPreparationInstruments = append(createdValidPreparationInstruments, createValidPreparationInstrumentForTest(t, ctx, exampleValidPreparationInstrument, dbc))

	// update
	updatedValidPreparationInstrument := fakes.BuildFakeValidPreparationInstrument()
	updatedValidPreparationInstrument.ID = createdValidPreparationInstruments[0].ID
	updatedValidPreparationInstrument.Preparation = createdValidPreparationInstruments[0].Preparation
	updatedValidPreparationInstrument.Instrument = createdValidPreparationInstruments[0].Instrument
	require.NoError(t, dbc.UpdateValidPreparationInstrument(ctx, updatedValidPreparationInstrument))

	// create more (each with unique prep+instrument pair to satisfy uniqueness constraint)
	for range exampleQuantity {
		tempValidInstrument := createValidInstrumentForTest(t, ctx, nil, dbc)
		tempValidPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
		input := fakes.BuildFakeValidPreparationInstrument()
		input.Preparation = *tempValidPreparation
		input.Instrument = *tempValidInstrument
		createdValidPreparationInstruments = append(createdValidPreparationInstruments, createValidPreparationInstrumentForTest(t, ctx, input, dbc))
	}

	// fetch as list
	validPreparationInstruments, err := dbc.GetValidPreparationInstruments(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, validPreparationInstruments.Data)
	assert.Len(t, validPreparationInstruments.Data, len(createdValidPreparationInstruments))

	forPreparation, err := dbc.GetValidPreparationInstrumentsForPreparation(ctx, createdValidPreparationInstruments[0].Preparation.ID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, forPreparation.Data)

	forInstrument, err := dbc.GetValidPreparationInstrumentsForInstrument(ctx, createdValidPreparationInstruments[0].Instrument.ID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, forInstrument.Data)

	// delete
	for _, validPreparationInstrument := range createdValidPreparationInstruments {
		require.NoError(t, dbc.ArchiveValidPreparationInstrument(ctx, validPreparationInstrument.ID))

		var exists bool
		exists, err = dbc.ValidPreparationInstrumentExists(ctx, validPreparationInstrument.ID)
		require.NoError(t, err)
		assert.False(t, exists)

		var y *types.ValidPreparationInstrument
		y, err = dbc.GetValidPreparationInstrument(ctx, validPreparationInstrument.ID)
		assert.Nil(t, y)
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_ValidPreparationInstrumentExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid preparation instrument MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.ValidPreparationInstrumentExists(ctx, "")
		require.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid preparation instrument MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidPreparationInstrument(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateValidPreparationInstrument(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateValidPreparationInstrument(ctx, nil))
	})
}

func TestQuerier_ArchiveValidPreparationInstrument(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid preparation instrument MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveValidPreparationInstrument(ctx, ""))
	})
}

func TestQuerier_GetValidPreparationInstrumentsByIDs(T *testing.T) {
	T.Parallel()

	T.Run("with empty list", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidPreparationInstrumentsByIDs(ctx, []string{})
		require.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Empty(t, actual)
	})
}

func TestQuerier_Integration_GetValidPreparationInstrumentsByIDs(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Create multiple valid preparation instruments
	created1 := createValidPreparationInstrumentForTest(t, ctx, nil, dbc)
	created2 := createValidPreparationInstrumentForTest(t, ctx, nil, dbc)
	created3 := createValidPreparationInstrumentForTest(t, ctx, nil, dbc)

	// Test fetching by IDs
	ids := []string{created1.ID, created2.ID, created3.ID}
	results, err := dbc.GetValidPreparationInstrumentsByIDs(ctx, ids)
	require.NoError(t, err)
	assert.Len(t, results, 3)
	assert.NotNil(t, results[created1.ID])
	assert.NotNil(t, results[created2.ID])
	assert.NotNil(t, results[created3.ID])

	// Test with partial IDs (some exist, some don't)
	partialIDs := []string{created1.ID, "nonexistent-id"}
	partialResults, err := dbc.GetValidPreparationInstrumentsByIDs(ctx, partialIDs)
	require.NoError(t, err)
	assert.Len(t, partialResults, 1)
	assert.NotNil(t, partialResults[created1.ID])

	// Test with empty list
	emptyResults, err := dbc.GetValidPreparationInstrumentsByIDs(ctx, []string{})
	require.NoError(t, err)
	assert.Empty(t, emptyResults)

	// Cleanup
	assert.NoError(t, dbc.ArchiveValidPreparationInstrument(ctx, created1.ID))
	assert.NoError(t, dbc.ArchiveValidPreparationInstrument(ctx, created2.ID))
	assert.NoError(t, dbc.ArchiveValidPreparationInstrument(ctx, created3.ID))
}

func TestQuerier_Integration_ValidPreparationInstruments_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ValidPreparationInstrument]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "valid preparation instrument",
		CreateItem: func(ctx context.Context, i int) *types.ValidPreparationInstrument {
			// Create unique preparation and instrument for each item
			exampleValidInstrument := createValidInstrumentForTest(t, ctx, nil, dbc)
			exampleValidPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
			exampleValidPreparationInstrument := fakes.BuildFakeValidPreparationInstrument()
			exampleValidPreparationInstrument.Preparation = *exampleValidPreparation
			exampleValidPreparationInstrument.Instrument = *exampleValidInstrument
			return createValidPreparationInstrumentForTest(t, ctx, exampleValidPreparationInstrument, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationInstrument], error) {
			return dbc.GetValidPreparationInstruments(ctx, filter)
		},
		GetID: func(validPreparationInstrument *types.ValidPreparationInstrument) string {
			return validPreparationInstrument.ID
		},
		CleanupItem: func(ctx context.Context, validPreparationInstrument *types.ValidPreparationInstrument) error {
			return dbc.ArchiveValidPreparationInstrument(ctx, validPreparationInstrument.ID)
		},
	})
}
