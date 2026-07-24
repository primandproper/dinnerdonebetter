package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v6/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMealPlanningManager_ListMealPlanGroceryListItemsByMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanGroceryListItemsList()
		exampleMealPlanID := fakes.BuildFakeID()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanGroceryListItemsForMealPlan), testutils.ContextMatcher, exampleMealPlanID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.ListMealPlanGroceryListItemsByMealPlan(ctx, exampleMealPlanID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_CreateMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanGroceryListItem()
		fakeInput := fakes.BuildFakeMealPlanGroceryListItemCreationRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.CreateMealPlanGroceryListItem), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanGroceryListItemDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.MealPlanGroceryListItemCreatedServiceEventType: {mealplanningkeys.MealPlanGroceryListItemIDKey},
			},
		)

		actual, err := mpm.CreateMealPlanGroceryListItem(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ReadMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanGroceryListItem()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanGroceryListItem), testutils.ContextMatcher, exampleMealPlanID, expected.ID).Return(expected, nil)
			},
		)

		actual, err := mpm.ReadMealPlanGroceryListItem(ctx, exampleMealPlanID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_UpdateMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanGroceryListItem := fakes.BuildFakeMealPlanGroceryListItem()
		exampleMealPlanID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanGroceryListItemUpdateRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanGroceryListItem), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanGroceryListItem.ID).Return(exampleMealPlanGroceryListItem, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateMealPlanGroceryListItem), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanGroceryListItem]()).Return(nil)
			},
			map[string][]string{
				types.MealPlanGroceryListItemUpdatedServiceEventType: {
					mealplanningkeys.MealPlanIDKey,
					mealplanningkeys.MealPlanGroceryListItemIDKey,
				},
			},
		)

		assert.NoError(t, mpm.UpdateMealPlanGroceryListItem(ctx, exampleMealPlanID, exampleMealPlanGroceryListItem.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ArchiveMealPlanGroceryListItem(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanGroceryListItem()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.ArchiveMealPlanGroceryListItem), testutils.ContextMatcher, expected.ID).Return(nil)
			},
			map[string][]string{
				types.MealPlanGroceryListItemArchivedServiceEventType: {
					mealplanningkeys.MealPlanIDKey,
					mealplanningkeys.MealPlanGroceryListItemIDKey,
				},
			},
		)

		err := mpm.ArchiveMealPlanGroceryListItem(ctx, mealPlanID, expected.ID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
