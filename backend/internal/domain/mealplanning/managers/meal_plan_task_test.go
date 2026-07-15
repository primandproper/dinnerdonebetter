package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v4/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMealPlanningManager_ListMealPlanTasksByMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanTasksList()
		exampleMealPlanID := fakes.BuildFakeID()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanTasksForMealPlan), testutils.ContextMatcher, exampleMealPlanID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.ListMealPlanTasksByMealPlan(ctx, exampleMealPlanID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanTask), testutils.ContextMatcher, expected.ID).Return(expected, nil)
			},
		)

		actual, err := mpm.ReadMealPlanTask(ctx, exampleMealPlanID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.CreateMealPlanTask), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanTaskDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.MealPlanTaskCreatedServiceEventType: {mealplanningkeys.MealPlanTaskIDKey},
			},
		)

		actual, err := mpm.CreateMealPlanTask(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_MealPlanTaskStatusChange(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleInput := fakes.BuildFakeMealPlanTaskStatusChangeRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.ChangeMealPlanTaskStatus), testutils.ContextMatcher, exampleInput).Return(nil)
			},
			map[string][]string{
				types.MealPlanTaskStatusChangedServiceEventType: {
					mealplanningkeys.MealPlanTaskIDKey,
				},
			},
		)

		assert.NoError(t, mpm.MealPlanTaskStatusChange(ctx, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
