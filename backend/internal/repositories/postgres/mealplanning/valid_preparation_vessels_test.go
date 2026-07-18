package mealplanning

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createValidPreparationVesselForTest(t *testing.T, ctx context.Context, exampleValidPreparationVessel *types.ValidPreparationVessel, dbc *repository) *types.ValidPreparationVessel {
	t.Helper()

	// create
	if exampleValidPreparationVessel == nil {
		exampleValidVessel := createValidVesselForTest(t, ctx, nil, dbc)
		exampleValidPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
		exampleValidPreparationVessel = fakes.BuildFakeValidPreparationVessel()
		exampleValidPreparationVessel.Vessel = *exampleValidVessel
		exampleValidPreparationVessel.Preparation = *exampleValidPreparation
	}

	dbInput := converters.ConvertValidPreparationVesselToValidPreparationVesselDatabaseCreationInput(exampleValidPreparationVessel)

	created, err := dbc.CreateValidPreparationVessel(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)
	exampleValidPreparationVessel.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleValidPreparationVessel, created)

	validPreparationVessel, err := dbc.GetValidPreparationVessel(ctx, created.ID)
	exampleValidPreparationVessel.CreatedAt = validPreparationVessel.CreatedAt
	exampleValidPreparationVessel.Preparation = validPreparationVessel.Preparation
	exampleValidPreparationVessel.Vessel = validPreparationVessel.Vessel

	assert.NoError(t, err)
	assert.Equal(t, validPreparationVessel, exampleValidPreparationVessel)

	return created
}

func TestQuerier_Integration_ValidPreparationVessels(t *testing.T) {
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

	exampleValidVessel := createValidVesselForTest(t, ctx, nil, dbc)
	exampleValidPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
	exampleValidPreparationVessel := fakes.BuildFakeValidPreparationVessel()
	exampleValidPreparationVessel.Preparation = *exampleValidPreparation
	exampleValidPreparationVessel.Vessel = *exampleValidVessel
	createdValidPreparationVessels := []*types.ValidPreparationVessel{}

	// create
	createdValidPreparationVessels = append(createdValidPreparationVessels, createValidPreparationVesselForTest(t, ctx, exampleValidPreparationVessel, dbc))

	// update
	updatedValidPreparationVessel := fakes.BuildFakeValidPreparationVessel()
	updatedValidPreparationVessel.ID = createdValidPreparationVessels[0].ID
	updatedValidPreparationVessel.Preparation = createdValidPreparationVessels[0].Preparation
	updatedValidPreparationVessel.Vessel = createdValidPreparationVessels[0].Vessel
	assert.NoError(t, dbc.UpdateValidPreparationVessel(ctx, updatedValidPreparationVessel))

	// create more (each must have unique prep+vessel per active row)
	for range exampleQuantity {
		extraVessel := createValidVesselForTest(t, ctx, nil, dbc)
		input := fakes.BuildFakeValidPreparationVessel()
		input.Preparation = createdValidPreparationVessels[0].Preparation
		input.Vessel = *extraVessel
		createdValidPreparationVessels = append(createdValidPreparationVessels, createValidPreparationVesselForTest(t, ctx, input, dbc))
	}

	// fetch as list
	validPreparationVessels, err := dbc.GetValidPreparationVessels(ctx, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, validPreparationVessels.Data)
	assert.Equal(t, len(createdValidPreparationVessels), len(validPreparationVessels.Data))

	forPreparation, err := dbc.GetValidPreparationVesselsForPreparation(ctx, createdValidPreparationVessels[0].Preparation.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, forPreparation.Data)

	forVessel, err := dbc.GetValidPreparationVesselsForVessel(ctx, createdValidPreparationVessels[0].Vessel.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, forVessel.Data)

	// delete
	for _, validPreparationVessel := range createdValidPreparationVessels {
		assert.NoError(t, dbc.ArchiveValidPreparationVessel(ctx, validPreparationVessel.ID))

		var exists bool
		exists, err = dbc.ValidPreparationVesselExists(ctx, validPreparationVessel.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.ValidPreparationVessel
		y, err = dbc.GetValidPreparationVessel(ctx, validPreparationVessel.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_ValidPreparationVesselExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid preparation instrument MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.ValidPreparationVesselExists(ctx, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid preparation instrument MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidPreparationVessel(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateValidPreparationVessel(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateValidPreparationVessel(ctx, nil))
	})
}

func TestQuerier_ArchiveValidPreparationVessel(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid preparation instrument MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveValidPreparationVessel(ctx, ""))
	})
}

func TestQuerier_GetValidPreparationVesselsByIDs(T *testing.T) {
	T.Parallel()

	T.Run("with empty list", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidPreparationVesselsByIDs(ctx, []string{})
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.Empty(t, actual)
	})
}

func TestQuerier_Integration_GetValidPreparationVesselsByIDs(t *testing.T) {
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

	// Create multiple valid preparation vessels
	created1 := createValidPreparationVesselForTest(t, ctx, nil, dbc)
	created2 := createValidPreparationVesselForTest(t, ctx, nil, dbc)
	created3 := createValidPreparationVesselForTest(t, ctx, nil, dbc)

	// Test fetching by IDs
	ids := []string{created1.ID, created2.ID, created3.ID}
	results, err := dbc.GetValidPreparationVesselsByIDs(ctx, ids)
	assert.NoError(t, err)
	assert.Len(t, results, 3)
	assert.NotNil(t, results[created1.ID])
	assert.NotNil(t, results[created2.ID])
	assert.NotNil(t, results[created3.ID])

	// Test with partial IDs (some exist, some don't)
	partialIDs := []string{created1.ID, "nonexistent-id"}
	partialResults, err := dbc.GetValidPreparationVesselsByIDs(ctx, partialIDs)
	assert.NoError(t, err)
	assert.Len(t, partialResults, 1)
	assert.NotNil(t, partialResults[created1.ID])

	// Test with empty list
	emptyResults, err := dbc.GetValidPreparationVesselsByIDs(ctx, []string{})
	assert.NoError(t, err)
	assert.Empty(t, emptyResults)

	// Cleanup
	assert.NoError(t, dbc.ArchiveValidPreparationVessel(ctx, created1.ID))
	assert.NoError(t, dbc.ArchiveValidPreparationVessel(ctx, created2.ID))
	assert.NoError(t, dbc.ArchiveValidPreparationVessel(ctx, created3.ID))
}

func TestQuerier_Integration_ValidPreparationVessels_CursorBasedPagination(t *testing.T) {
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

	// Create different preparations and vessels for each item to ensure uniqueness
	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ValidPreparationVessel]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "valid preparation vessel",
		CreateItem: func(ctx context.Context, i int) *types.ValidPreparationVessel {
			// Create unique preparation and vessel for each item
			exampleValidVessel := createValidVesselForTest(t, ctx, nil, dbc)
			exampleValidPreparation := createValidPreparationForTest(t, ctx, nil, dbc)
			exampleValidPreparationVessel := fakes.BuildFakeValidPreparationVessel()
			exampleValidPreparationVessel.Preparation = *exampleValidPreparation
			exampleValidPreparationVessel.Vessel = *exampleValidVessel
			return createValidPreparationVesselForTest(t, ctx, exampleValidPreparationVessel, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationVessel], error) {
			return dbc.GetValidPreparationVessels(ctx, filter)
		},
		GetID: func(validPreparationVessel *types.ValidPreparationVessel) string {
			return validPreparationVessel.ID
		},
		CleanupItem: func(ctx context.Context, validPreparationVessel *types.ValidPreparationVessel) error {
			return dbc.ArchiveValidPreparationVessel(ctx, validPreparationVessel.ID)
		},
	})
}
