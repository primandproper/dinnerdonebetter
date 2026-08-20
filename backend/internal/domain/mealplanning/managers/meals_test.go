package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMealPlanningManager_ListMeals(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealsList()

		db := &mealplanningmock.RepositoryMock{
			GetMealsFunc: func(_ context.Context, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Meal], error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMeals(ctx, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealsCalls(), 1)
	})
}

func TestMealPlanningManager_CreateMeal(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		creator := fake.BuildFakeID()
		expected := fakes.BuildFakeMeal()
		fakeInput := fakes.BuildFakeMealCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			FindMealWithSameComponentsFunc: func(_ context.Context, creatorID string, _ *types.MealCreationRequestInput) (*types.Meal, error) {
				assert.Equal(t, creator, creatorID)

				return nil, types.ErrNoMatchingMeal
			},
			CreateMealFunc: func(_ context.Context, _ *types.MealDatabaseCreationInput) (*types.Meal, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMeal(ctx, creator, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.FindMealWithSameComponentsCalls(), 1)
		assert.Len(t, db.CreateMealCalls(), 1)
	})

	T.Run("returns ErrDuplicateMeal when FindMealWithSameComponents returns existing meal", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		creator := fake.BuildFakeID()
		existingMeal := fakes.BuildFakeMeal()
		fakeInput := fakes.BuildFakeMealCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			FindMealWithSameComponentsFunc: func(_ context.Context, creatorID string, _ *types.MealCreationRequestInput) (*types.Meal, error) {
				assert.Equal(t, creator, creatorID)

				return existingMeal, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMeal(ctx, creator, fakeInput)
		require.ErrorIs(t, err, types.ErrDuplicateMeal)
		assert.Nil(t, actual)

		assert.Len(t, db.FindMealWithSameComponentsCalls(), 1)
	})
}

func TestMealPlanningManager_ReadMeal(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMeal()

		db := &mealplanningmock.RepositoryMock{
			GetMealFunc: func(_ context.Context, mealID string) (*types.Meal, error) {
				assert.Equal(t, expected.ID, mealID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMeal(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealCalls(), 1)
	})
}

func TestMealPlanningManager_SearchMeals(T *testing.T) {
	T.Parallel()

	T.Run("useSearchService false uses database", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealsList()
		exampleQuery := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			SearchForMealsFunc: func(_ context.Context, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Meal], error) {
				assert.Equal(t, exampleQuery, query)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.SearchMeals(ctx, exampleQuery, false, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForMealsCalls(), 1)
	})

	T.Run("useSearchService true falls back to database when search returns empty", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealsList()
		exampleQuery := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			SearchForMealsFunc: func(_ context.Context, query string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Meal], error) {
				assert.Equal(t, exampleQuery, query)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.SearchMeals(ctx, exampleQuery, true, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.SearchForMealsCalls(), 1)
	})
}

func TestMealPlanningManager_ArchiveMeal(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMeal()

		db := &mealplanningmock.RepositoryMock{
			ArchiveMealFunc: func(_ context.Context, mealID string, userID string) error {
				assert.Equal(t, expected.ID, mealID)
				assert.Equal(t, expected.CreatedByUser, userID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMeal(ctx, expected.ID, expected.CreatedByUser)
		require.NoError(t, err)

		assert.Len(t, db.ArchiveMealCalls(), 1)
	})
}
