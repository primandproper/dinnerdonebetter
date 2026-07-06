package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v3/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMealPlanningManager_ListMeals(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealsList()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMeals), testutils.ContextMatcher, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.ListMeals(ctx, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_CreateMeal(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		creator := fakes.BuildFakeID()
		expected := fakes.BuildFakeMeal()
		fakeInput := fakes.BuildFakeMealCreationRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.FindMealWithSameComponents), testutils.ContextMatcher, creator, testutils.MatchType[*types.MealCreationRequestInput]()).Return(nil, types.ErrNoMatchingMeal)
				db.On(reflection.GetMethodName(mpm.db.CreateMeal), testutils.ContextMatcher, testutils.MatchType[*types.MealDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.MealCreatedServiceEventType: {mealplanningkeys.MealIDKey},
			},
		)

		actual, err := mpm.CreateMeal(ctx, creator, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})

	T.Run("returns ErrDuplicateMeal when FindMealWithSameComponents returns existing meal", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		creator := fakes.BuildFakeID()
		existingMeal := fakes.BuildFakeMeal()
		fakeInput := fakes.BuildFakeMealCreationRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.FindMealWithSameComponents), testutils.ContextMatcher, creator, testutils.MatchType[*types.MealCreationRequestInput]()).Return(existingMeal, nil)
			},
		)

		actual, err := mpm.CreateMeal(ctx, creator, fakeInput)
		assert.ErrorIs(t, err, types.ErrDuplicateMeal)
		assert.Nil(t, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ReadMeal(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMeal()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMeal), testutils.ContextMatcher, expected.ID).Return(expected, nil)
			},
		)

		actual, err := mpm.ReadMeal(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_SearchMeals(T *testing.T) {
	T.Parallel()

	T.Run("useSearchService false uses database", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealsList()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.SearchForMeals), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.SearchMeals(ctx, exampleQuery, false, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})

	T.Run("useSearchService true falls back to database when search returns empty", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealsList()
		exampleQuery := fakes.BuildFakeID()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.SearchForMeals), testutils.ContextMatcher, exampleQuery, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.SearchMeals(ctx, exampleQuery, true, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ArchiveMeal(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMeal()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.ArchiveMeal), testutils.ContextMatcher, expected.ID, expected.CreatedByUser).Return(nil)
			},
			map[string][]string{
				types.MealArchivedServiceEventType: {mealplanningkeys.MealIDKey},
			},
		)

		err := mpm.ArchiveMeal(ctx, expected.ID, expected.CreatedByUser)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
