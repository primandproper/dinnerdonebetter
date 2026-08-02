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

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/stretchr/testify/assert"
)

func createValidIngredientStateForTest(t *testing.T, ctx context.Context, exampleValidIngredientState *types.ValidIngredientState, dbc *repository) *types.ValidIngredientState {
	t.Helper()

	// create
	if exampleValidIngredientState == nil {
		exampleValidIngredientState = fakes.BuildFakeValidIngredientState()
	}
	dbInput := converters.ConvertValidIngredientStateToValidIngredientStateDatabaseCreationInput(exampleValidIngredientState)

	created, err := dbc.CreateValidIngredientState(ctx, dbInput)
	exampleValidIngredientState.CreatedAt = created.CreatedAt
	assert.NoError(t, err)
	assert.Equal(t, exampleValidIngredientState, created)

	validIngredientState, err := dbc.GetValidIngredientState(ctx, created.ID)
	exampleValidIngredientState.CreatedAt = validIngredientState.CreatedAt

	assert.NoError(t, err)
	assert.Equal(t, validIngredientState, exampleValidIngredientState)

	return validIngredientState
}

func TestQuerier_Integration_ValidIngredientStates(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	exampleValidIngredientState := fakes.BuildFakeValidIngredientState()
	createdValidIngredientStates := []*types.ValidIngredientState{}

	// create
	createdValidIngredientStates = append(createdValidIngredientStates, createValidIngredientStateForTest(t, ctx, exampleValidIngredientState, dbc))

	// update
	updatedValidIngredientState := fakes.BuildFakeValidIngredientState()
	updatedValidIngredientState.ID = createdValidIngredientStates[0].ID
	assert.NoError(t, dbc.UpdateValidIngredientState(ctx, updatedValidIngredientState))

	// create more
	for i := range exampleQuantity {
		input := fakes.BuildFakeValidIngredientState()
		input.Name = fmt.Sprintf("%s %d", updatedValidIngredientState.Name, i)
		createdValidIngredientStates = append(createdValidIngredientStates, createValidIngredientStateForTest(t, ctx, input, dbc))
	}

	// fetch as list
	validIngredientStates, err := dbc.GetValidIngredientStates(ctx, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, validIngredientStates.Data)
	assert.Equal(t, len(createdValidIngredientStates), len(validIngredientStates.Data))

	// fetch as list of IDs
	validIngredientStateIDs := []string{}
	for _, validIngredientState := range createdValidIngredientStates {
		validIngredientStateIDs = append(validIngredientStateIDs, validIngredientState.ID)
	}

	byIDs, err := dbc.GetValidIngredientStatesWithIDs(ctx, validIngredientStateIDs)
	assert.NoError(t, err)
	assert.Equal(t, validIngredientStates.Data, byIDs)

	// fetch via name search
	byName, err := dbc.SearchForValidIngredientStates(ctx, updatedValidIngredientState.Name, nil)
	assert.NoError(t, err)
	assert.Equal(t, validIngredientStates, byName)

	needingIndexing, err := dbc.GetValidIngredientStateIDsThatNeedSearchIndexing(ctx)
	assert.NoError(t, err)
	assert.NotEmpty(t, needingIndexing)

	// delete
	for _, validIngredientState := range createdValidIngredientStates {
		assert.NoError(t, dbc.MarkValidIngredientStateAsIndexed(ctx, validIngredientState.ID))
		assert.NoError(t, dbc.ArchiveValidIngredientState(ctx, validIngredientState.ID))

		var exists bool
		exists, err = dbc.ValidIngredientStateExists(ctx, validIngredientState.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.ValidIngredientState
		y, err = dbc.GetValidIngredientState(ctx, validIngredientState.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_ValidIngredientStateExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient state ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.ValidIngredientStateExists(ctx, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient state ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidIngredientState(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_SearchForValidIngredientStates(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient state ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.SearchForValidIngredientStates(ctx, "", nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateValidIngredientState(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateValidIngredientState(ctx, nil))
	})
}

func TestQuerier_ArchiveValidIngredientState(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient state ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveValidIngredientState(ctx, ""))
	})
}

func TestQuerier_MarkValidIngredientStateAsIndexed(T *testing.T) {
	T.Parallel()

	T.Run("with invalid ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.MarkValidIngredientStateAsIndexed(ctx, ""))
	})
}

func TestQuerier_Integration_ValidIngredientStates_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ValidIngredientState]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "valid ingredient state",
		CreateItem: func(ctx context.Context, i int) *types.ValidIngredientState {
			validIngredientState := fakes.BuildFakeValidIngredientState()
			validIngredientState.Name = fmt.Sprintf("Valid Ingredient State %02d", i)
			return createValidIngredientStateForTest(t, ctx, validIngredientState, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientState], error) {
			return dbc.GetValidIngredientStates(ctx, filter)
		},
		GetID: func(validIngredientState *types.ValidIngredientState) string {
			return validIngredientState.ID
		},
		CleanupItem: func(ctx context.Context, validIngredientState *types.ValidIngredientState) error {
			return dbc.ArchiveValidIngredientState(ctx, validIngredientState.ID)
		},
	})
}
