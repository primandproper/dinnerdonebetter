package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v11/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

		db := &mealplanningmock.RepositoryMock{
			GetRecipeListsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeList], error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipeLists(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeListsCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			CreateRecipeListFunc: func(_ context.Context, _ *types.RecipeListDatabaseCreationInput) (*types.RecipeList, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeList(ctx, userID, input)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateRecipeListCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			ArchiveRecipeListFunc: func(_ context.Context, recipeListID string, actualUserID string) error {
				assert.Equal(t, listID, recipeListID)
				assert.Equal(t, userID, actualUserID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.ArchiveRecipeList(ctx, listID, userID))

		assert.Len(t, db.ArchiveRecipeListCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			UpdateRecipeListFunc: func(_ context.Context, _ *types.RecipeList) error {
				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.UpdateRecipeList(ctx, listID, userID, input))

		assert.Len(t, db.UpdateRecipeListCalls(), 1)
	})
}
