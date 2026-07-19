package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v5/filtering"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.UpdateMealListItem), testutils.ContextMatcher, testutils.MatchType[*types.MealListItem]()).Return(nil)
			},
		)

		assert.NoError(t, mpm.UpdateMealListItem(ctx, itemID, listID, mealID, input))

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.MealExistsInMealList), testutils.ContextMatcher, listID, mealID).Return(false, nil)
				db.On(reflection.GetMethodName(mpm.db.CreateMealListItem), testutils.ContextMatcher, testutils.MatchType[*types.MealListItemDatabaseCreationInput]()).Return(expected, nil)
			},
		)

		actual, err := mpm.AddMealToMealList(ctx, listID, mealID, expected.Notes)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})

	T.Run("returns ErrDuplicateMealInList when MealExistsInMealList returns true", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		listID := fakes.BuildFakeID()
		mealID := fakes.BuildFakeID()
		notes := t.Name()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.MealExistsInMealList), testutils.ContextMatcher, listID, mealID).Return(true, nil)
			},
		)

		actual, err := mpm.AddMealToMealList(ctx, listID, mealID, notes)
		assert.ErrorIs(t, err, types.ErrDuplicateMealInList)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.ArchiveMealListItem), testutils.ContextMatcher, itemID, listID).Return(nil)
			},
		)

		assert.NoError(t, mpm.RemoveMealFromMealList(ctx, listID, itemID))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ListMealListItems(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		listID := fakes.BuildFakeID()
		expectedItem := &types.MealListItem{
			ID:                fakes.BuildFakeID(),
			BelongsToMealList: listID,
			Notes:             t.Name(),
			Meal:              types.Meal{ID: fakes.BuildFakeID()},
		}
		expected := &filtering.QueryFilteredResult[types.MealListItem]{Data: []*types.MealListItem{expectedItem}}

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealListItems), testutils.ContextMatcher, listID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.ListMealListItems(ctx, listID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
