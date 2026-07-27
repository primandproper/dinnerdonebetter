package managers

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v7/filtering"

	"github.com/stretchr/testify/assert"
)

func TestRecipeManager_UpdateRecipeListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		itemID := fakes.BuildFakeID()
		listID := fakes.BuildFakeID()
		recipeID := fakes.BuildFakeID()
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

		assert.NoError(t, rm.UpdateRecipeListItem(ctx, itemID, listID, recipeID, input))

		assert.Len(t, db.UpdateRecipeListItemCalls(), 1)
	})
}

func TestRecipeManager_AddRecipeToRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		listID := fakes.BuildFakeID()
		recipeID := fakes.BuildFakeID()
		expected := &types.RecipeListItem{
			ID:                  fakes.BuildFakeID(),
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
		assert.NoError(t, err)
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

		listID := fakes.BuildFakeID()
		itemID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			ArchiveRecipeListItemFunc: func(_ context.Context, recipeListItemID string, recipeListID string) error {
				assert.Equal(t, itemID, recipeListItemID)
				assert.Equal(t, listID, recipeListID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		assert.NoError(t, rm.RemoveRecipeFromRecipeList(ctx, listID, itemID))

		assert.Len(t, db.ArchiveRecipeListItemCalls(), 1)
	})
}

func TestRecipeManager_ListRecipeListItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		listID := fakes.BuildFakeID()
		expectedItem := &types.RecipeListItem{
			ID:                  fakes.BuildFakeID(),
			BelongsToRecipeList: listID,
			Notes:               t.Name(),
			Recipe:              types.Recipe{ID: fakes.BuildFakeID()},
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
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeListItemsCalls(), 1)
	})
}
