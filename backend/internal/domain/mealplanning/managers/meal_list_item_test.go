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

func TestMealPlanningManager_UpdateMealListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		itemID := fake.BuildFakeID()
		listID := fake.BuildFakeID()
		mealID := fake.BuildFakeID()
		notes := new(t.Name())
		input := &types.MealListItemUpdateRequestInput{
			Notes: notes,
		}

		db := &mealplanningmock.RepositoryMock{
			UpdateMealListItemFunc: func(_ context.Context, _ *types.MealListItem) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		require.NoError(t, mpm.UpdateMealListItem(ctx, itemID, listID, mealID, input))

		assert.Len(t, db.UpdateMealListItemCalls(), 1)
	})
}

func TestMealPlanningManager_AddMealToMealList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		listID := fake.BuildFakeID()
		mealID := fake.BuildFakeID()
		expected := &types.MealListItem{
			ID:                fake.BuildFakeID(),
			BelongsToMealList: listID,
			Notes:             t.Name(),
			Meal:              types.Meal{ID: mealID},
		}

		db := &mealplanningmock.RepositoryMock{
			MealExistsInMealListFunc: func(_ context.Context, mealListID string, actualMealID string) (bool, error) {
				assert.Equal(t, listID, mealListID)
				assert.Equal(t, mealID, actualMealID)

				return false, nil
			},
			CreateMealListItemFunc: func(_ context.Context, _ *types.MealListItemDatabaseCreationInput) (*types.MealListItem, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.AddMealToMealList(ctx, listID, mealID, expected.Notes)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.MealExistsInMealListCalls(), 1)
		assert.Len(t, db.CreateMealListItemCalls(), 1)
	})

	T.Run("returns ErrDuplicateMealInList when MealExistsInMealList returns true", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		listID := fake.BuildFakeID()
		mealID := fake.BuildFakeID()
		notes := t.Name()

		db := &mealplanningmock.RepositoryMock{
			MealExistsInMealListFunc: func(_ context.Context, mealListID string, actualMealID string) (bool, error) {
				assert.Equal(t, listID, mealListID)
				assert.Equal(t, mealID, actualMealID)

				return true, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.AddMealToMealList(ctx, listID, mealID, notes)
		require.ErrorIs(t, err, types.ErrDuplicateMealInList)
		assert.Nil(t, actual)

		assert.Len(t, db.MealExistsInMealListCalls(), 1)
	})
}

func TestMealPlanningManager_RemoveMealFromMealList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		listID := fake.BuildFakeID()
		itemID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			ArchiveMealListItemFunc: func(_ context.Context, mealListItemID string, mealListID string) error {
				assert.Equal(t, itemID, mealListItemID)
				assert.Equal(t, listID, mealListID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		require.NoError(t, mpm.RemoveMealFromMealList(ctx, listID, itemID))

		assert.Len(t, db.ArchiveMealListItemCalls(), 1)
	})
}

func TestMealPlanningManager_ListMealListItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		listID := fake.BuildFakeID()
		userID := fake.BuildFakeID()
		expectedItem := &types.MealListItem{
			ID:                fake.BuildFakeID(),
			BelongsToMealList: listID,
			Notes:             t.Name(),
			Meal:              types.Meal{ID: fake.BuildFakeID()},
		}
		expected := &filtering.QueryFilteredResult[types.MealListItem]{Data: []*types.MealListItem{expectedItem}}

		db := &mealplanningmock.RepositoryMock{
			GetMealListItemsFunc: func(_ context.Context, mealListID string, actualUserID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealListItem], error) {
				assert.Equal(t, listID, mealListID)
				assert.Equal(t, userID, actualUserID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealListItems(ctx, listID, userID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealListItemsCalls(), 1)
	})
}
