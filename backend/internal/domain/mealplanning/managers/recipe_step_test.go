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
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStep()
		fakeInput := fakes.BuildFakeRecipeStepCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc: recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			CreateRecipeStepFunc: func(_ context.Context, _ *types.RecipeStepDatabaseCreationInput) (*types.RecipeStep, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStep(ctx, exampleRecipeID, exampleOwnerID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateRecipeStepCalls(), 1)
	})

	T.Run("with a caller who does not own the recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc: recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, false),
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStep(ctx, exampleRecipeID, exampleOwnerID, fakes.BuildFakeRecipeStepCreationRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.CreateRecipeStepCalls())
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
		exampleOwnerID := fake.BuildFakeID()
		exampleRecipeStep := fakes.BuildFakeRecipeStep()
		exampleInput := fakes.BuildFakeRecipeStepUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStep.ID, true),
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

		require.NoError(t, rm.UpdateRecipeStep(ctx, exampleRecipeID, exampleRecipeStep.ID, exampleOwnerID, exampleInput))

		assert.Len(t, db.GetRecipeStepCalls(), 1)
		assert.Len(t, db.UpdateRecipeStepCalls(), 1)
	})

	T.Run("with a caller who does not own the recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc: recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, false),
		}
		attachRepositoryToManager(rm, db)

		err := rm.UpdateRecipeStep(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakes.BuildFakeRecipeStepUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.GetRecipeStepCalls())
		assert.Empty(t, db.UpdateRecipeStepCalls())
	})

	T.Run("with a step from another recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, false),
		}
		attachRepositoryToManager(rm, db)

		err := rm.UpdateRecipeStep(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakes.BuildFakeRecipeStepUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.GetRecipeStepCalls())
		assert.Empty(t, db.UpdateRecipeStepCalls())
	})
}

func TestRecipeManager_ArchiveRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStep()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, expected.ID, true),
			ArchiveRecipeStepFunc: func(_ context.Context, recipeID string, recipeStepID string) error {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, expected.ID, recipeStepID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.ArchiveRecipeStep(ctx, exampleRecipeID, expected.ID, exampleOwnerID))

		assert.Len(t, db.ArchiveRecipeStepCalls(), 1)
	})

	T.Run("with a caller who does not own the recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc: recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, false),
		}
		attachRepositoryToManager(rm, db)

		err := rm.ArchiveRecipeStep(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.ArchiveRecipeStepCalls())
	})

	T.Run("with a step from another recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, false),
		}
		attachRepositoryToManager(rm, db)

		err := rm.ArchiveRecipeStep(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveRecipeStepCalls())
	})
}
