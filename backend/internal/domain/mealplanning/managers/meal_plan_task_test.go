package managers

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMealPlanningManager_ListMealPlanTasksByMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanTasksList()
		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			GetMealPlanTasksForMealPlanFunc: func(_ context.Context, mealPlanID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanTask], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlanTasksByMealPlan(ctx, exampleMealPlanID, exampleOwnerID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanTasksForMealPlanCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlanTasksByMealPlan(ctx, exampleMealPlanID, exampleOwnerID, nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetMealPlanTasksForMealPlanCalls())
	})
}

func TestMealPlanningManager_ReadMealPlanTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanTask()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:     mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanTaskExistsFunc: mealPlanTaskExistsStub(t, exampleMealPlanID, expected.ID, true),
			GetMealPlanTaskFunc: func(_ context.Context, mealPlanTaskID string) (*types.MealPlanTask, error) {
				assert.Equal(t, expected.ID, mealPlanTaskID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanTask(ctx, exampleMealPlanID, expected.ID, exampleOwnerID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanTaskCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanTask(ctx, exampleMealPlanID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetMealPlanTaskCalls())
	})

	// The read is keyed by the task alone, so the task has to be shown to be the plan's.
	T.Run("with a task from another meal plan", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleTaskID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:     mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanTaskExistsFunc: mealPlanTaskExistsStub(t, exampleMealPlanID, exampleTaskID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanTask(ctx, exampleMealPlanID, exampleTaskID, exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetMealPlanTaskCalls())
	})
}

func TestMealPlanningManager_CreateMealPlanTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanTask()
		fakeInput := fakes.BuildFakeMealPlanTaskCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:                 mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanOptionBelongsToAccountFunc: mealPlanOptionBelongsToAccountStub(t, fakeInput.MealPlanOptionID, exampleOwnerID, true),
			CreateMealPlanTaskFunc: func(_ context.Context, input *types.MealPlanTaskDatabaseCreationInput) (*types.MealPlanTask, error) {
				assert.Equal(t, fakeInput.MealPlanOptionID, input.MealPlanOptionID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanTask(ctx, exampleMealPlanID, exampleOwnerID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanTaskCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanTask(ctx, exampleMealPlanID, exampleOwnerID, fakes.BuildFakeMealPlanTaskCreationRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.MealPlanOptionBelongsToAccountCalls())
		assert.Empty(t, db.CreateMealPlanTaskCalls())
	})

	T.Run("with a meal plan option outside the caller's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeMealPlanTaskCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:                 mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanOptionBelongsToAccountFunc: mealPlanOptionBelongsToAccountStub(t, fakeInput.MealPlanOptionID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanTask(ctx, exampleMealPlanID, exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.CreateMealPlanTaskCalls())
	})
}

func TestMealPlanningManager_MealPlanTaskStatusChange(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanTaskStatusChangeRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:     mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanTaskExistsFunc: mealPlanTaskExistsStub(t, exampleMealPlanID, exampleInput.MealPlanTaskID, true),
			ChangeMealPlanTaskStatusFunc: func(_ context.Context, input *types.MealPlanTaskStatusChangeRequestInput) error {
				assert.Equal(t, exampleInput, input)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		require.NoError(t, mpm.MealPlanTaskStatusChange(ctx, exampleMealPlanID, exampleOwnerID, exampleInput))

		assert.Len(t, db.ChangeMealPlanTaskStatusCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.MealPlanTaskStatusChange(ctx, exampleMealPlanID, exampleOwnerID, fakes.BuildFakeMealPlanTaskStatusChangeRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.MealPlanTaskExistsCalls())
		assert.Empty(t, db.ChangeMealPlanTaskStatusCalls())
	})

	T.Run("with a task from another meal plan", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanTaskStatusChangeRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:     mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanTaskExistsFunc: mealPlanTaskExistsStub(t, exampleMealPlanID, exampleInput.MealPlanTaskID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.MealPlanTaskStatusChange(ctx, exampleMealPlanID, exampleOwnerID, exampleInput)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ChangeMealPlanTaskStatusCalls())
	})
}
