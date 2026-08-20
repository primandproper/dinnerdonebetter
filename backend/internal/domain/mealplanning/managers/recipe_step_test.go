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

func TestRecipeManager_ListRecipeSteps(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipeStepsList()
		exampleRecipeID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepsFunc: func(_ context.Context, recipeID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStep], error) {
				assert.Equal(t, exampleRecipeID, recipeID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipeSteps(ctx, exampleRecipeID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepsCalls(), 1)
	})
}

func TestRecipeManager_CreateRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStep()
		fakeInput := fakes.BuildFakeRecipeStepCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateRecipeStepFunc: func(_ context.Context, _ *types.RecipeStepDatabaseCreationInput) (*types.RecipeStep, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStep(ctx, exampleRecipeID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateRecipeStepCalls(), 1)
	})
}

func TestRecipeManager_ReadRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStep()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepFunc: func(_ context.Context, recipeID string, recipeStepID string) (*types.RecipeStep, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, expected.ID, recipeStepID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ReadRecipeStep(ctx, exampleRecipeID, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepCalls(), 1)
	})
}

func TestRecipeManager_UpdateRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStep := fakes.BuildFakeRecipeStep()
		exampleInput := fakes.BuildFakeRecipeStepUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepFunc: func(_ context.Context, recipeID string, recipeStepID string) (*types.RecipeStep, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStep.ID, recipeStepID)

				return exampleRecipeStep, nil
			},
			UpdateRecipeStepFunc: func(_ context.Context, _ *types.RecipeStep) error {
				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.UpdateRecipeStep(ctx, exampleRecipeID, exampleRecipeStep.ID, exampleInput))

		assert.Len(t, db.GetRecipeStepCalls(), 1)
		assert.Len(t, db.UpdateRecipeStepCalls(), 1)
	})
}

func TestRecipeManager_ArchiveRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStep()

		db := &mealplanningmock.RepositoryMock{
			ArchiveRecipeStepFunc: func(_ context.Context, recipeID string, recipeStepID string) error {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, expected.ID, recipeStepID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.ArchiveRecipeStep(ctx, exampleRecipeID, expected.ID))

		assert.Len(t, db.ArchiveRecipeStepCalls(), 1)
	})
}
