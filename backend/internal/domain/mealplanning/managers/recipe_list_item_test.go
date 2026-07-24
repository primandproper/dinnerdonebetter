package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.UpdateRecipeListItem), testutils.ContextMatcher, testutils.MatchType[*types.RecipeListItem]()).Return(nil)
			},
		)

		assert.NoError(t, rm.UpdateRecipeListItem(ctx, itemID, listID, recipeID, input))

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.CreateRecipeListItem), testutils.ContextMatcher, testutils.MatchType[*types.RecipeListItemDatabaseCreationInput]()).Return(expected, nil)
			},
		)

		actual, err := rm.AddRecipeToRecipeList(ctx, listID, recipeID, expected.Notes)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.ArchiveRecipeListItem), testutils.ContextMatcher, itemID, listID).Return(nil)
			},
		)

		assert.NoError(t, rm.RemoveRecipeFromRecipeList(ctx, listID, itemID))

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipeListItems), testutils.ContextMatcher, listID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := rm.ListRecipeListItems(ctx, listID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
