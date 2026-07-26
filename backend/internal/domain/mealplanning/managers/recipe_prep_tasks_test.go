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

func TestRecipeManager_ListRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipePrepTasksList()
		exampleRecipeID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetRecipePrepTasksFunc: func(_ context.Context, recipeID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipePrepTask], error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipePrepTask(ctx, exampleRecipeID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipePrepTasksCalls(), 1)
	})
}

func TestRecipeManager_CreateRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipePrepTask()
		fakeInput := fakes.BuildFakeRecipePrepTaskCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateRecipePrepTaskFunc: func(_ context.Context, _ *types.RecipePrepTaskDatabaseCreationInput) (*types.RecipePrepTask, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipePrepTask(ctx, exampleRecipeID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateRecipePrepTaskCalls(), 1)
	})
}

func TestRecipeManager_ReadRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipePrepTask()

		db := &mealplanningmock.RepositoryMock{
			GetRecipePrepTaskFunc: func(_ context.Context, recipeID string, recipePrepTaskID string) (*types.RecipePrepTask, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, expected.ID, recipePrepTaskID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ReadRecipePrepTask(ctx, exampleRecipeID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipePrepTaskCalls(), 1)
	})
}

func TestRecipeManager_UpdateRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipePrepTask := fakes.BuildFakeRecipePrepTask()
		exampleInput := fakes.BuildFakeRecipePrepTaskUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetRecipePrepTaskFunc: func(_ context.Context, recipeID string, recipePrepTaskID string) (*types.RecipePrepTask, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipePrepTask.ID, recipePrepTaskID)

				return exampleRecipePrepTask, nil
			},
			UpdateRecipePrepTaskFunc: func(_ context.Context, _ *types.RecipePrepTask) error {
				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		assert.NoError(t, rm.UpdateRecipePrepTask(ctx, exampleRecipeID, exampleRecipePrepTask.ID, exampleInput))

		assert.Len(t, db.GetRecipePrepTaskCalls(), 1)
		assert.Len(t, db.UpdateRecipePrepTaskCalls(), 1)
	})
}

func TestRecipeManager_ArchiveRecipePrepTask(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipePrepTask()

		db := &mealplanningmock.RepositoryMock{
			ArchiveRecipePrepTaskFunc: func(_ context.Context, recipeID string, recipePrepTaskID string) error {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, expected.ID, recipePrepTaskID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		assert.NoError(t, rm.ArchiveRecipePrepTask(ctx, exampleRecipeID, expected.ID))

		assert.Len(t, db.ArchiveRecipePrepTaskCalls(), 1)
	})
}
