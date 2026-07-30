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

	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createValidIngredientGroupForTest(t *testing.T, ctx context.Context, exampleValidIngredientGroup *types.ValidIngredientGroup, dbc *repository) *types.ValidIngredientGroup {
	t.Helper()

	// create
	if exampleValidIngredientGroup == nil {
		exampleValidIngredientGroup = fakes.BuildFakeValidIngredientGroup()
	}
	dbInput := converters.ConvertValidIngredientGroupToValidIngredientGroupDatabaseCreationInput(exampleValidIngredientGroup)

	created, err := dbc.CreateValidIngredientGroup(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)
	exampleValidIngredientGroup.CreatedAt = created.CreatedAt
	for i := range exampleValidIngredientGroup.Members {
		exampleValidIngredientGroup.Members[i].CreatedAt = created.Members[i].CreatedAt
		exampleValidIngredientGroup.Members[i].ValidIngredient = created.Members[i].ValidIngredient
	}
	assert.Equal(t, exampleValidIngredientGroup, created)

	validIngredientGroup, err := dbc.GetValidIngredientGroup(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, validIngredientGroup)

	exampleValidIngredientGroup.CreatedAt = validIngredientGroup.CreatedAt
	for i := range exampleValidIngredientGroup.Members {
		exampleValidIngredientGroup.Members[i].CreatedAt = validIngredientGroup.Members[i].CreatedAt
		exampleValidIngredientGroup.Members[i].ValidIngredient = validIngredientGroup.Members[i].ValidIngredient
	}

	assert.NoError(t, err)
	assert.Equal(t, validIngredientGroup, exampleValidIngredientGroup)

	return created
}

func TestQuerier_Integration_ValidIngredientGroups(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	exampleValidIngredientGroup := fakes.BuildFakeValidIngredientGroup()
	exampleValidIngredientGroup.Members = []*types.ValidIngredientGroupMember{}
	for range exampleQuantity {
		exampleValidIngredientGroup.Members = append(exampleValidIngredientGroup.Members, &types.ValidIngredientGroupMember{
			ID:              identifiers.New(),
			BelongsToGroup:  exampleValidIngredientGroup.ID,
			ValidIngredient: *createValidIngredientForTest(t, ctx, nil, dbc),
		})
	}
	createdValidIngredientGroups := []*types.ValidIngredientGroup{}

	// create
	createdValidIngredientGroups = append(createdValidIngredientGroups, createValidIngredientGroupForTest(t, ctx, exampleValidIngredientGroup, dbc))

	// update
	updatedValidIngredientGroup := fakes.BuildFakeValidIngredientGroup()
	updatedValidIngredientGroup.ID = createdValidIngredientGroups[0].ID
	assert.NoError(t, dbc.UpdateValidIngredientGroup(ctx, updatedValidIngredientGroup))

	// fetch as list
	validIngredientGroups, err := dbc.GetValidIngredientGroups(ctx, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, validIngredientGroups.Data)
	assert.Equal(t, len(createdValidIngredientGroups), len(validIngredientGroups.Data))

	// fetch via name search
	byName, err := dbc.SearchForValidIngredientGroups(ctx, updatedValidIngredientGroup.Name, nil)
	assert.NoError(t, err)
	assert.Equal(t, validIngredientGroups, byName)

	// delete
	for _, validIngredientGroup := range createdValidIngredientGroups {
		assert.NoError(t, dbc.ArchiveValidIngredientGroup(ctx, validIngredientGroup.ID))

		var exists bool
		exists, err = dbc.ValidIngredientGroupExists(ctx, validIngredientGroup.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.ValidIngredientGroup
		y, err = dbc.GetValidIngredientGroup(ctx, validIngredientGroup.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_ValidIngredientGroupExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient group MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.ValidIngredientGroupExists(ctx, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient group MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetValidIngredientGroup(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_SearchForValidIngredientGroups(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient group MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)
		filter := filtering.DefaultQueryFilter()

		actual, err := c.SearchForValidIngredientGroups(ctx, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateValidIngredientGroup(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateValidIngredientGroup(ctx, nil))
	})
}

func TestQuerier_ArchiveValidIngredientGroup(T *testing.T) {
	T.Parallel()

	T.Run("with invalid valid ingredient group MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveValidIngredientGroup(ctx, ""))
	})
}

func TestQuerier_Integration_ValidIngredientGroups_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.ValidIngredientGroup]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "valid ingredient group",
		CreateItem: func(ctx context.Context, i int) *types.ValidIngredientGroup {
			validIngredient := createValidIngredientForTest(t, ctx, nil, dbc)
			validIngredientGroup := fakes.BuildFakeValidIngredientGroup()
			validIngredientGroup.Name = fmt.Sprintf("Valid Ingredient Group %02d", i)
			validIngredientGroup.Members = []*types.ValidIngredientGroupMember{
				{
					ID:              identifiers.New(),
					BelongsToGroup:  validIngredientGroup.ID,
					ValidIngredient: *validIngredient,
				},
			}
			return createValidIngredientGroupForTest(t, ctx, validIngredientGroup, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientGroup], error) {
			return dbc.GetValidIngredientGroups(ctx, filter)
		},
		GetID: func(validIngredientGroup *types.ValidIngredientGroup) string {
			return validIngredientGroup.ID
		},
		CleanupItem: func(ctx context.Context, validIngredientGroup *types.ValidIngredientGroup) error {
			return dbc.ArchiveValidIngredientGroup(ctx, validIngredientGroup.ID)
		},
	})
}
