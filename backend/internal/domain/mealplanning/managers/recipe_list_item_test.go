package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipeManager_UpdateRecipeListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		itemID := fake.BuildFakeID()
		listID := fake.BuildFakeID()
		recipeID := fake.BuildFakeID()
		notes := new(t.Name())
		input := &types.RecipeListItemUpdateRequestInput{
			Notes: notes,
		}

		db := &mealplanningmock.RepositoryMock{
			UpdateRecipeListItemFunc: func(_ context.Context, _ *types.RecipeListItem) error {
				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.UpdateRecipeListItem(ctx, itemID, listID, recipeID, input))

		assert.Len(t, db.UpdateRecipeListItemCalls(), 1)
	})
}

func TestRecipeManager_AddRecipeToRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		listID := fake.BuildFakeID()
		recipeID := fake.BuildFakeID()
		expected := &types.RecipeListItem{
			ID:                  fake.BuildFakeID(),
			BelongsToRecipeList: listID,
			Notes:               t.Name(),
			Recipe:              types.Recipe{ID: recipeID},
		}

		db := &mealplanningmock.RepositoryMock{
			CreateRecipeListItemFunc: func(_ context.Context, _ *types.RecipeListItemDatabaseCreationInput) (*types.RecipeListItem, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.AddRecipeToRecipeList(ctx, listID, recipeID, expected.Notes)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateRecipeListItemCalls(), 1)
	})
}

func TestRecipeManager_RemoveRecipeFromRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		listID := fake.BuildFakeID()
		itemID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			ArchiveRecipeListItemFunc: func(_ context.Context, recipeListItemID string, recipeListID string) error {
				assert.Equal(t, itemID, recipeListItemID)
				assert.Equal(t, listID, recipeListID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.RemoveRecipeFromRecipeList(ctx, listID, itemID))

		assert.Len(t, db.ArchiveRecipeListItemCalls(), 1)
	})
}

func TestRecipeManager_ListRecipeListItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		listID := fake.BuildFakeID()
		expectedItem := &types.RecipeListItem{
			ID:                  fake.BuildFakeID(),
			BelongsToRecipeList: listID,
			Notes:               t.Name(),
			Recipe:              types.Recipe{ID: fake.BuildFakeID()},
		}
		expected := &filtering.QueryFilteredResult[types.RecipeListItem]{Data: []*types.RecipeListItem{expectedItem}}

		db := &mealplanningmock.RepositoryMock{
			GetRecipeListItemsFunc: func(_ context.Context, recipeListID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeListItem], error) {
				assert.Equal(t, listID, recipeListID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipeListItems(ctx, listID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeListItemsCalls(), 1)
	})
}
