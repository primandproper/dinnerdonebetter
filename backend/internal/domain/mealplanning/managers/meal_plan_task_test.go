package managers

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v8/filtering"

	"github.com/stretchr/testify/assert"
)

func TestMealPlanningManager_ListMealPlanTasksByMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanTasksList()
		exampleMealPlanID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanTasksForMealPlanFunc: func(_ context.Context, mealPlanID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanTask], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlanTasksByMealPlan(ctx, exampleMealPlanID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanTasksForMealPlanCalls(), 1)
	})
}

func TestMealPlanningManager_ReadMealPlanTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanTask()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanTaskFunc: func(_ context.Context, mealPlanTaskID string) (*types.MealPlanTask, error) {
				assert.Equal(t, expected.ID, mealPlanTaskID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanTask(ctx, exampleMealPlanID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanTaskCalls(), 1)
	})
}

func TestMealPlanningManager_CreateMealPlanTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanTask()
		fakeInput := fakes.BuildFakeMealPlanTaskCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateMealPlanTaskFunc: func(_ context.Context, _ *types.MealPlanTaskDatabaseCreationInput) (*types.MealPlanTask, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanTask(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanTaskCalls(), 1)
	})
}

func TestMealPlanningManager_MealPlanTaskStatusChange(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleInput := fakes.BuildFakeMealPlanTaskStatusChangeRequestInput()

		db := &mealplanningmock.RepositoryMock{
			ChangeMealPlanTaskStatusFunc: func(_ context.Context, input *types.MealPlanTaskStatusChangeRequestInput) error {
				assert.Equal(t, exampleInput, input)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		assert.NoError(t, mpm.MealPlanTaskStatusChange(ctx, exampleInput))

		assert.Len(t, db.ChangeMealPlanTaskStatusCalls(), 1)
	})
}
