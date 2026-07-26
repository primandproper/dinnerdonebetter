package managers

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v6/filtering"

	"github.com/stretchr/testify/assert"
)

func TestMealPlanningManager_UpdateMealListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		itemID := fakes.BuildFakeID()
		listID := fakes.BuildFakeID()
		mealID := fakes.BuildFakeID()
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

		assert.NoError(t, mpm.UpdateMealListItem(ctx, itemID, listID, mealID, input))

		assert.Len(t, db.UpdateMealListItemCalls(), 1)
	})
}

func TestMealPlanningManager_AddMealToMealList(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		listID := fakes.BuildFakeID()
		mealID := fakes.BuildFakeID()
		expected := &types.MealListItem{
			ID:                fakes.BuildFakeID(),
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
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.MealExistsInMealListCalls(), 1)
		assert.Len(t, db.CreateMealListItemCalls(), 1)
	})

	T.Run("returns ErrDuplicateMealInList when MealExistsInMealList returns true", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		listID := fakes.BuildFakeID()
		mealID := fakes.BuildFakeID()
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
		assert.ErrorIs(t, err, types.ErrDuplicateMealInList)
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

		listID := fakes.BuildFakeID()
		itemID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			ArchiveMealListItemFunc: func(_ context.Context, mealListItemID string, mealListID string) error {
				assert.Equal(t, itemID, mealListItemID)
				assert.Equal(t, listID, mealListID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		assert.NoError(t, mpm.RemoveMealFromMealList(ctx, listID, itemID))

		assert.Len(t, db.ArchiveMealListItemCalls(), 1)
	})
}

func TestMealPlanningManager_ListMealListItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		listID := fakes.BuildFakeID()
		userID := fakes.BuildFakeID()
		expectedItem := &types.MealListItem{
			ID:                fakes.BuildFakeID(),
			BelongsToMealList: listID,
			Notes:             t.Name(),
			Meal:              types.Meal{ID: fakes.BuildFakeID()},
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
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealListItemsCalls(), 1)
	})
}
