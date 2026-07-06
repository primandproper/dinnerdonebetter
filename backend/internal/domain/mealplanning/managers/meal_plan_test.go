package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	mealplanningworkers "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v3/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMealPlanningManager_ListMealPlans(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlansList()
		exampleOwnerID := fakes.BuildFakeID()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlansForAccount), testutils.ContextMatcher, exampleOwnerID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.ListMealPlans(ctx, exampleOwnerID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_CreateMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		ownerID := fakes.BuildFakeID()
		creatorID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlan()
		fakeInput := fakes.BuildFakeMealPlanCreationRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.CreateMealPlan), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.MealPlanCreatedServiceEventType: {mealplanningkeys.MealPlanIDKey},
			},
		)

		actual, err := mpm.CreateMealPlan(ctx, ownerID, creatorID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})

	T.Run("invokes workers when meal plan is created finalized", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		groceryWorker := &mealplanningworkers.MockWorker{}
		taskWorker := &mealplanningworkers.MockWorker{}
		groceryWorker.On("Work", testutils.ContextMatcher).Return(nil)
		taskWorker.On("Work", testutils.ContextMatcher).Return(nil)

		mpm := buildMealPlanManagerForTestWithWorkers(t, groceryWorker, taskWorker)

		ownerID := fakes.BuildFakeID()
		creatorID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlan()
		expected.Status = string(types.MealPlanStatusFinalized)
		fakeInput := fakes.BuildFakeMealPlanCreationRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.CreateMealPlan), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.MealPlanCreatedServiceEventType: {mealplanningkeys.MealPlanIDKey},
			},
		)

		actual, err := mpm.CreateMealPlan(ctx, ownerID, creatorID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, append(expectations, groceryWorker, taskWorker)...)
	})
}

func TestMealPlanningManager_ReadMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlan()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlan), testutils.ContextMatcher, exampleMealPlanID, expected.ID).Return(expected, nil)
			},
		)

		actual, err := mpm.ReadMealPlan(ctx, exampleMealPlanID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_UpdateMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlan := fakes.BuildFakeMealPlan()
		ownerID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanUpdateRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlan), testutils.ContextMatcher, exampleMealPlan.ID, ownerID).Return(exampleMealPlan, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateMealPlan), testutils.ContextMatcher, testutils.MatchType[*types.MealPlan]()).Return(nil)
			},
			map[string][]string{
				types.MealPlanUpdatedServiceEventType: {mealplanningkeys.MealPlanIDKey},
			},
		)

		assert.NoError(t, mpm.UpdateMealPlan(ctx, exampleMealPlan.ID, ownerID, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ArchiveMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlan()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.ArchiveMealPlan), testutils.ContextMatcher, expected.ID, expected.CreatedByUser).Return(nil)
			},
			map[string][]string{
				types.MealPlanArchivedServiceEventType: {mealplanningkeys.MealPlanIDKey},
			},
		)

		err := mpm.ArchiveMealPlan(ctx, expected.ID, expected.CreatedByUser)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_FinalizeMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlan()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.AttemptToFinalizeMealPlan), testutils.ContextMatcher, expected.ID, expected.CreatedByUser).Return(true, nil)
			},
			map[string][]string{
				types.MealPlanFinalizedServiceEventType: {mealplanningkeys.MealPlanIDKey},
			},
		)

		finalized, err := mpm.FinalizeMealPlan(ctx, expected.ID, expected.CreatedByUser)
		assert.True(t, finalized)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})

	T.Run("invokes workers when finalized", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		groceryWorker := &mealplanningworkers.MockWorker{}
		taskWorker := &mealplanningworkers.MockWorker{}
		groceryWorker.On("Work", testutils.ContextMatcher).Return(nil)
		taskWorker.On("Work", testutils.ContextMatcher).Return(nil)

		mpm := buildMealPlanManagerForTestWithWorkers(t, groceryWorker, taskWorker)

		expected := fakes.BuildFakeMealPlan()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.AttemptToFinalizeMealPlan), testutils.ContextMatcher, expected.ID, expected.CreatedByUser).Return(true, nil)
			},
			map[string][]string{
				types.MealPlanFinalizedServiceEventType: {mealplanningkeys.MealPlanIDKey},
			},
		)

		finalized, err := mpm.FinalizeMealPlan(ctx, expected.ID, expected.CreatedByUser)
		assert.True(t, finalized)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, append(expectations, groceryWorker, taskWorker)...)
	})
}
