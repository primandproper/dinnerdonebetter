package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v10/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMealPlanningManager_ListMealPlans(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlansList()
		exampleOwnerID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlansForAccountFunc: func(_ context.Context, accountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlan], error) {
				assert.Equal(t, exampleOwnerID, accountID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlans(ctx, exampleOwnerID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlansForAccountCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			CreateMealPlanFunc: func(_ context.Context, _ *types.MealPlanDatabaseCreationInput) (*types.MealPlan, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlan(ctx, ownerID, creatorID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanCalls(), 1)
	})

	T.Run("starts the finalization saga when meal plan is created finalized", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		starter := &fakeFinalizationStarter{}

		mpm := buildMealPlanManagerForTestWithStarter(t, starter)

		ownerID := fakes.BuildFakeID()
		creatorID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlan()
		expected.Status = string(types.MealPlanStatusFinalized)
		fakeInput := fakes.BuildFakeMealPlanCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateMealPlanFunc: func(_ context.Context, _ *types.MealPlanDatabaseCreationInput) (*types.MealPlan, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlan(ctx, ownerID, creatorID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanCalls(), 1)
		assert.Equal(t, []string{expected.ID}, starter.calls)
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

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanFunc: func(_ context.Context, mealPlanID, accountID string) (*types.MealPlan, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, expected.ID, accountID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlan(ctx, exampleMealPlanID, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanFunc: func(_ context.Context, mealPlanID, accountID string) (*types.MealPlan, error) {
				assert.Equal(t, exampleMealPlan.ID, mealPlanID)
				assert.Equal(t, ownerID, accountID)

				return exampleMealPlan, nil
			},
			UpdateMealPlanFunc: func(_ context.Context, _ *types.MealPlan) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		require.NoError(t, mpm.UpdateMealPlan(ctx, exampleMealPlan.ID, ownerID, exampleInput))

		assert.Len(t, db.GetMealPlanCalls(), 1)
		assert.Len(t, db.UpdateMealPlanCalls(), 1)
	})
}

func TestMealPlanningManager_ArchiveMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlan()

		db := &mealplanningmock.RepositoryMock{
			ArchiveMealPlanFunc: func(_ context.Context, mealPlanID, accountID string) error {
				assert.Equal(t, expected.ID, mealPlanID)
				assert.Equal(t, expected.CreatedByUser, accountID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlan(ctx, expected.ID, expected.CreatedByUser)
		require.NoError(t, err)

		assert.Len(t, db.ArchiveMealPlanCalls(), 1)
	})
}

func TestMealPlanningManager_FinalizeMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlan()

		db := &mealplanningmock.RepositoryMock{
			AttemptToFinalizeMealPlanFunc: func(_ context.Context, mealPlanID, accountID string) (bool, error) {
				assert.Equal(t, expected.ID, mealPlanID)
				assert.Equal(t, expected.CreatedByUser, accountID)

				return true, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		finalized, err := mpm.FinalizeMealPlan(ctx, expected.ID, expected.CreatedByUser)
		assert.True(t, finalized)
		require.NoError(t, err)

		assert.Len(t, db.AttemptToFinalizeMealPlanCalls(), 1)
	})

	T.Run("starts the finalization saga when finalized", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		starter := &fakeFinalizationStarter{}

		mpm := buildMealPlanManagerForTestWithStarter(t, starter)

		expected := fakes.BuildFakeMealPlan()

		db := &mealplanningmock.RepositoryMock{
			AttemptToFinalizeMealPlanFunc: func(_ context.Context, mealPlanID, accountID string) (bool, error) {
				assert.Equal(t, expected.ID, mealPlanID)
				assert.Equal(t, expected.CreatedByUser, accountID)

				return true, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		finalized, err := mpm.FinalizeMealPlan(ctx, expected.ID, expected.CreatedByUser)
		assert.True(t, finalized)
		require.NoError(t, err)

		assert.Len(t, db.AttemptToFinalizeMealPlanCalls(), 1)
		assert.Equal(t, []string{expected.ID}, starter.calls)
	})
}
