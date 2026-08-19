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

	"github.com/primandproper/platform-go/v11/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createValidPreparationForTest(t *testing.T, ctx context.Context, exampleValidPreparation *types.ValidPreparation, dbc *repository) *types.ValidPreparation {
	t.Helper()

	// create
	if exampleValidPreparation == nil {
		exampleValidPreparation = fakes.BuildFakeValidPreparation()
	}
	dbInput := converters.ConvertValidPreparationToValidPreparationDatabaseCreationInput(exampleValidPreparation)

	created, err := dbc.CreateValidPreparation(ctx, dbInput)
	exampleValidPreparation.CreatedAt = created.CreatedAt
	require.NoError(t, err)
	assert.Equal(t, exampleValidPreparation, created)

	validPreparation, err := dbc.GetValidPreparation(ctx, created.ID)
	exampleValidPreparation.CreatedAt = validPreparation.CreatedAt

	require.NoError(t, err)
	assert.Equal(t, validPreparation, exampleValidPreparation)

	return validPreparation
}

func TestQuerier_Integration_ValidPreparations(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	exampleValidPreparation := fakes.BuildFakeValidPreparation()
	createdValidPreparations := []*types.ValidPreparation{}

	// create
	createdValidPreparations = append(createdValidPreparations, createValidPreparationForTest(t, ctx, exampleValidPreparation, dbc))

	// update
	updatedValidPreparation := fakes.BuildFakeValidPreparation()
	updatedValidPreparation.ID = createdValidPreparations[0].ID
	require.NoError(t, dbc.UpdateValidPreparation(ctx, updatedValidPreparation))

	// create more
	for i := range exampleQuantity {
		input := fakes.BuildFakeValidPreparation()
		input.Name = fmt.Sprintf("%s %d", updatedValidPreparation.Name, i)
		createdValidPreparations = append(createdValidPreparations, createValidPreparationForTest(t, ctx, input, dbc))
	}

	// fetch as list
	validPreparations, err := dbc.GetValidPreparations(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, validPreparations.Data)
	assert.Len(t, validPreparations.Data, len(createdValidPreparations))

	// fetch as list of IDs
	validPreparationIDs := []string{}
	for _, validPreparation := range createdValidPreparations {
		validPreparationIDs = append(validPreparationIDs, validPreparation.ID)
	}

	byIDs, err := dbc.GetValidPreparationsWithIDs(ctx, validPreparationIDs)
	require.NoError(t, err)
	assert.Equal(t, validPreparations.Data, byIDs)

	// fetch via name search
	byName, err := dbc.SearchForValidPreparations(ctx, updatedValidPreparation.Name, nil)
	require.NoError(t, err)
	assert.Equal(t, validPreparations, byName)

	whatever, err := dbc.ScanValidPreparationIDsForReindex(ctx, "", 100)
	require.NoError(t, err)
	assert.NotEmpty(t, whatever)

	require.NoError(t, dbc.MarkValidPreparationsAsIndexed(ctx, []string{updatedValidPreparation.ID}))

	randomPreparation, err := dbc.GetRandomValidPreparation(ctx)
	require.NoError(t, err)
	assert.NotNil(t, randomPreparation)

	// delete
	for _, validPreparation := range createdValidPreparations {
		require.NoError(t, dbc.ArchiveValidPreparation(ctx, validPreparation.ID))

		var exists bool
		exists, err = dbc.ValidPreparationExists(ctx, validPreparation.ID)
		require.NoError(t, err)
		assert.False(t, exists)

		var y *types.ValidPreparation
		y, err = dbc.GetValidPreparation(ctx, validPreparation.ID)
		assert.Nil(t, y)
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_ValidPreparationExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid preparation ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.ValidPreparationExists(ctx, "")
		require.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid preparation ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidPreparation(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_SearchForValidPreparations(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid preparation ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.SearchForValidPreparations(ctx, "", nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetValidPreparationsWithIDs(T *testing.T) {
	T.Parallel()

	T.Run("with nil IDs", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidPreparationsWithIDs(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateValidPreparation(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateValidPreparation(ctx, nil))
	})
}

func TestQuerier_ArchiveValidPreparation(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid preparation ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveValidPreparation(ctx, ""))
	})
}

func TestQuerier_MarkValidPreparationsAsIndexed(T *testing.T) {
	T.Parallel()

	T.Run("with no ids", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		// The client is inert, so a nil error is the assertion that nothing was
		// executed: an empty flush must not reach the database at all.
		assert.NoError(t, c.MarkValidPreparationsAsIndexed(ctx, nil))
	})
}

func TestQuerier_Integration_ValidPreparations_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ValidPreparation]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "valid preparation",
		CreateItem: func(ctx context.Context, i int) *types.ValidPreparation {
			validPreparation := fakes.BuildFakeValidPreparation()
			validPreparation.Name = fmt.Sprintf("Valid Preparation %02d", i)
			return createValidPreparationForTest(t, ctx, validPreparation, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparation], error) {
			return dbc.GetValidPreparations(ctx, filter)
		},
		GetID: func(validPreparation *types.ValidPreparation) string {
			return validPreparation.ID
		},
		CleanupItem: func(ctx context.Context, validPreparation *types.ValidPreparation) error {
			return dbc.ArchiveValidPreparation(ctx, validPreparation.ID)
		},
	})
}
