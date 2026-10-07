package managers

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recipeListOwnershipFunc answers RecipeListExists for exactly one (list, owner) pair.
func recipeListOwnershipFunc(t *testing.T, listID, ownerID string) func(context.Context, string, string) (bool, error) {
	t.Helper()

	return func(_ context.Context, recipeListID, userID string) (bool, error) {
		return recipeListID == listID && userID == ownerID, nil
	}
}

func TestRecipeManager_UpdateRecipeListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		itemID := fake.BuildFakeID()
		listID := fake.BuildFakeID()
		userID := fake.BuildFakeID()
		recipeID := fake.BuildFakeID()
		notes := new(t.Name())
		input := &types.RecipeListItemUpdateRequestInput{
			Notes: notes,
		}

		db := &mealplanningmock.RepositoryMock{
			RecipeListExistsFunc: recipeListOwnershipFunc(t, listID, userID),
			UpdateRecipeListItemFunc: func(_ context.Context, _ *types.RecipeListItem) error {
				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.UpdateRecipeListItem(ctx, itemID, listID, userID, recipeID, input))

		assert.Len(t, db.RecipeListExistsCalls(), 1)
		assert.Len(t, db.UpdateRecipeListItemCalls(), 1)
	})

	T.Run("with another user's list", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		listID := fake.BuildFakeID()
		input := &types.RecipeListItemUpdateRequestInput{
			Notes: new(t.Name()),
		}

		db := &mealplanningmock.RepositoryMock{
			RecipeListExistsFunc: recipeListOwnershipFunc(t, listID, fake.BuildFakeID()),
		}
		attachRepositoryToManager(rm, db)

		err := rm.UpdateRecipeListItem(ctx, fake.BuildFakeID(), listID, fake.BuildFakeID(), fake.BuildFakeID(), input)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.UpdateRecipeListItemCalls())
	})
}

func TestRecipeManager_AddRecipeToRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		listID := fake.BuildFakeID()
		userID := fake.BuildFakeID()
		recipeID := fake.BuildFakeID()
		expected := &types.RecipeListItem{
			ID:                  fake.BuildFakeID(),
			BelongsToRecipeList: listID,
			Notes:               t.Name(),
			Recipe:              types.Recipe{ID: recipeID},
		}

		db := &mealplanningmock.RepositoryMock{
			RecipeListExistsFunc: recipeListOwnershipFunc(t, listID, userID),
			CreateRecipeListItemFunc: func(_ context.Context, _ *types.RecipeListItemDatabaseCreationInput) (*types.RecipeListItem, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.AddRecipeToRecipeList(ctx, listID, userID, recipeID, expected.Notes)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.RecipeListExistsCalls(), 1)
		assert.Len(t, db.CreateRecipeListItemCalls(), 1)
	})

	T.Run("with another user's list", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		listID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			RecipeListExistsFunc: recipeListOwnershipFunc(t, listID, fake.BuildFakeID()),
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.AddRecipeToRecipeList(ctx, listID, fake.BuildFakeID(), fake.BuildFakeID(), t.Name())
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.CreateRecipeListItemCalls())
	})
}

func TestRecipeManager_RemoveRecipeFromRecipeList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		listID := fake.BuildFakeID()
		userID := fake.BuildFakeID()
		itemID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			RecipeListExistsFunc: recipeListOwnershipFunc(t, listID, userID),
			ArchiveRecipeListItemFunc: func(_ context.Context, recipeListItemID string, recipeListID string) error {
				assert.Equal(t, itemID, recipeListItemID)
				assert.Equal(t, listID, recipeListID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.RemoveRecipeFromRecipeList(ctx, listID, userID, itemID))

		assert.Len(t, db.RecipeListExistsCalls(), 1)
		assert.Len(t, db.ArchiveRecipeListItemCalls(), 1)
	})

	T.Run("with another user's list", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		listID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			RecipeListExistsFunc: recipeListOwnershipFunc(t, listID, fake.BuildFakeID()),
		}
		attachRepositoryToManager(rm, db)

		err := rm.RemoveRecipeFromRecipeList(ctx, listID, fake.BuildFakeID(), fake.BuildFakeID())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveRecipeListItemCalls())
	})
}

func TestRecipeManager_ListRecipeListItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		listID := fake.BuildFakeID()
		userID := fake.BuildFakeID()
		expectedItem := &types.RecipeListItem{
			ID:                  fake.BuildFakeID(),
			BelongsToRecipeList: listID,
			Notes:               t.Name(),
			Recipe:              types.Recipe{ID: fake.BuildFakeID()},
		}
		expected := &filtering.QueryFilteredResult[types.RecipeListItem]{Data: []*types.RecipeListItem{expectedItem}}

		db := &mealplanningmock.RepositoryMock{
			RecipeListExistsFunc: recipeListOwnershipFunc(t, listID, userID),
			GetRecipeListItemsFunc: func(_ context.Context, recipeListID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeListItem], error) {
				assert.Equal(t, listID, recipeListID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipeListItems(ctx, listID, userID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.RecipeListExistsCalls(), 1)
		assert.Len(t, db.GetRecipeListItemsCalls(), 1)
	})

	T.Run("with another user's list", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		listID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			RecipeListExistsFunc: recipeListOwnershipFunc(t, listID, fake.BuildFakeID()),
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipeListItems(ctx, listID, fake.BuildFakeID(), nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetRecipeListItemsCalls())
	})
}
