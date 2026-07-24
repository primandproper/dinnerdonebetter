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

func TestRecipeManager_ListRecipeLists(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		recipeList := &types.RecipeList{
			ID:            fakes.BuildFakeID(),
			Name:          t.Name(),
			Description:   t.Name(),
			BelongsToUser: fakes.BuildFakeID(),
		}
		expected := &filtering.QueryFilteredResult[types.RecipeList]{Data: []*types.RecipeList{recipeList}}

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.GetRecipeLists), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := rm.ListRecipeLists(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_CreateRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		userID := fakes.BuildFakeID()
		input := &types.RecipeListCreationRequestInput{
			Name:        t.Name(),
			Description: t.Name(),
		}
		expected := &types.RecipeList{ID: fakes.BuildFakeID(), Name: input.Name, Description: input.Description, BelongsToUser: userID}

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.CreateRecipeList), testutils.ContextMatcher, testutils.MatchType[*types.RecipeListDatabaseCreationInput]()).Return(expected, nil)
			},
		)

		actual, err := rm.CreateRecipeList(ctx, userID, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_ArchiveRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		userID := fakes.BuildFakeID()
		listID := fakes.BuildFakeID()

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.ArchiveRecipeList), testutils.ContextMatcher, listID, userID).Return(nil)
			},
		)

		assert.NoError(t, rm.ArchiveRecipeList(ctx, listID, userID))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestRecipeManager_UpdateRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		listID := fakes.BuildFakeID()
		userID := fakes.BuildFakeID()
		name := t.Name()
		desc := "desc"
		input := &types.RecipeListUpdateRequestInput{
			Name:        &name,
			Description: &desc,
		}

		expectations := setupExpectationsForRecipeManager(
			rm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(rm.db.UpdateRecipeList), testutils.ContextMatcher, testutils.MatchType[*types.RecipeList]()).Return(nil)
			},
		)

		assert.NoError(t, rm.UpdateRecipeList(ctx, listID, userID, input))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
