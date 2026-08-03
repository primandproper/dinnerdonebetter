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

func createValidIngredientStateIngredientForTest(t *testing.T, ctx context.Context, exampleValidIngredientStateIngredient *types.ValidIngredientStateIngredient, dbc *repository) *types.ValidIngredientStateIngredient {
	t.Helper()

	// create
	if exampleValidIngredientStateIngredient == nil {
		exampleValidIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
		exampleValidIngredientState := createValidIngredientStateForTest(t, ctx, nil, dbc)
		exampleValidIngredientStateIngredient = fakes.BuildFakeValidIngredientStateIngredient()
		exampleValidIngredientStateIngredient.Ingredient = *exampleValidIngredient
		exampleValidIngredientStateIngredient.IngredientState = *exampleValidIngredientState
	}

	dbInput := converters.ConvertValidIngredientStateIngredientToValidIngredientStateIngredientDatabaseCreationInput(exampleValidIngredientStateIngredient)

	created, err := dbc.CreateValidIngredientStateIngredient(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)
	exampleValidIngredientStateIngredient.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleValidIngredientStateIngredient, created)

	validIngredientStateIngredient, err := dbc.GetValidIngredientStateIngredient(ctx, created.ID)
	exampleValidIngredientStateIngredient.CreatedAt = validIngredientStateIngredient.CreatedAt
	exampleValidIngredientStateIngredient.IngredientState = validIngredientStateIngredient.IngredientState
	exampleValidIngredientStateIngredient.Ingredient = validIngredientStateIngredient.Ingredient

	require.NoError(t, err)
	assert.Equal(t, validIngredientStateIngredient, exampleValidIngredientStateIngredient)

	return validIngredientStateIngredient
}

func TestQuerier_Integration_ValidIngredientStateIngredients(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	exampleValidIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
	exampleValidIngredientState := createValidIngredientStateForTest(t, ctx, nil, dbc)
	exampleValidIngredientStateIngredient := fakes.BuildFakeValidIngredientStateIngredient()
	exampleValidIngredientStateIngredient.IngredientState = *exampleValidIngredientState
	exampleValidIngredientStateIngredient.Ingredient = *exampleValidIngredient
	createdValidIngredientStateIngredients := []*types.ValidIngredientStateIngredient{}

	// create
	createdValidIngredientStateIngredients = append(createdValidIngredientStateIngredients, createValidIngredientStateIngredientForTest(t, ctx, exampleValidIngredientStateIngredient, dbc))

	// update
	updatedValidIngredientStateIngredient := fakes.BuildFakeValidIngredientStateIngredient()
	updatedValidIngredientStateIngredient.ID = createdValidIngredientStateIngredients[0].ID
	updatedValidIngredientStateIngredient.IngredientState = createdValidIngredientStateIngredients[0].IngredientState
	updatedValidIngredientStateIngredient.Ingredient = createdValidIngredientStateIngredients[0].Ingredient
	require.NoError(t, dbc.UpdateValidIngredientStateIngredient(ctx, updatedValidIngredientStateIngredient))

	// create more (each must have unique ingredient+state per active row)
	for range exampleQuantity {
		extraState := createValidIngredientStateForTest(t, ctx, nil, dbc)
		input := fakes.BuildFakeValidIngredientStateIngredient()
		input.IngredientState = *extraState
		input.Ingredient = createdValidIngredientStateIngredients[0].Ingredient
		createdValidIngredientStateIngredients = append(createdValidIngredientStateIngredients, createValidIngredientStateIngredientForTest(t, ctx, input, dbc))
	}

	// fetch as list
	validIngredientStateIngredients, err := dbc.GetValidIngredientStateIngredients(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, validIngredientStateIngredients.Data)
	assert.Len(t, validIngredientStateIngredients.Data, len(createdValidIngredientStateIngredients))

	forIngredientState, err := dbc.GetValidIngredientStateIngredientsForIngredientState(ctx, createdValidIngredientStateIngredients[0].IngredientState.ID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, forIngredientState.Data)

	forIngredient, err := dbc.GetValidIngredientStateIngredientsForIngredient(ctx, createdValidIngredientStateIngredients[0].Ingredient.ID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, forIngredient.Data)

	// delete
	for _, validIngredientStateIngredient := range createdValidIngredientStateIngredients {
		require.NoError(t, dbc.ArchiveValidIngredientStateIngredient(ctx, validIngredientStateIngredient.ID))

		var exists bool
		exists, err = dbc.ValidIngredientStateIngredientExists(ctx, validIngredientStateIngredient.ID)
		require.NoError(t, err)
		assert.False(t, exists)

		var y *types.ValidIngredientStateIngredient
		y, err = dbc.GetValidIngredientStateIngredient(ctx, validIngredientStateIngredient.ID)
		assert.Nil(t, y)
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_ValidIngredientStateIngredientExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient preparation ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.ValidIngredientStateIngredientExists(ctx, "")
		require.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient preparation ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidIngredientStateIngredient(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateValidIngredientStateIngredient(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateValidIngredientStateIngredient(ctx, nil))
	})
}

func TestQuerier_ArchiveValidIngredientStateIngredient(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient preparation ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveValidIngredientStateIngredient(ctx, ""))
	})
}

func TestQuerier_Integration_ValidIngredientStateIngredients_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Create different ingredients and ingredient states for each item to ensure uniqueness
	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ValidIngredientStateIngredient]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "valid ingredient state ingredient",
		CreateItem: func(ctx context.Context, i int) *types.ValidIngredientStateIngredient {
			// Create unique ingredient and ingredient state for each item
			exampleValidIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
			exampleValidIngredientState := createValidIngredientStateForTest(t, ctx, nil, dbc)
			exampleValidIngredientStateIngredient := fakes.BuildFakeValidIngredientStateIngredient()
			exampleValidIngredientStateIngredient.Ingredient = *exampleValidIngredient
			exampleValidIngredientStateIngredient.IngredientState = *exampleValidIngredientState
			return createValidIngredientStateIngredientForTest(t, ctx, exampleValidIngredientStateIngredient, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientStateIngredient], error) {
			return dbc.GetValidIngredientStateIngredients(ctx, filter)
		},
		GetID: func(validIngredientStateIngredient *types.ValidIngredientStateIngredient) string {
			return validIngredientStateIngredient.ID
		},
		CleanupItem: func(ctx context.Context, validIngredientStateIngredient *types.ValidIngredientStateIngredient) error {
			return dbc.ArchiveValidIngredientStateIngredient(ctx, validIngredientStateIngredient.ID)
		},
	})
}
