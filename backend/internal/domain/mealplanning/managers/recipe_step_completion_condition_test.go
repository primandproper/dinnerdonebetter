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

func TestRecipeManager_ListRecipeStepCompletionConditions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipeStepCompletionConditionsList()
		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepCompletionConditionsFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepCompletionCondition], error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipeStepCompletionConditions(ctx, exampleRecipeID, exampleRecipeStepID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepCompletionConditionsCalls(), 1)
	})
}

func TestRecipeManager_CreateRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepCompletionCondition()
		fakeInput := fakes.BuildFakeRecipeStepCompletionConditionForExistingRecipeCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, true),
			CreateRecipeStepCompletionConditionFunc: func(_ context.Context, _ string, _ *types.RecipeStepCompletionConditionDatabaseCreationInput) (*types.RecipeStepCompletionCondition, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateRecipeStepCompletionConditionCalls(), 1)
	})

	T.Run("with a caller who does not own the recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeRecipeStepCompletionConditionForExistingRecipeCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc: recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, false),
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.CreateRecipeStepCompletionConditionCalls())
	})

	T.Run("with a step from another recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeRecipeStepCompletionConditionForExistingRecipeCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, false),
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.CreateRecipeStepCompletionConditionCalls())
	})
}

func TestRecipeManager_ReadRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepCompletionCondition()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepCompletionConditionFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepIngredientID string) (*types.RecipeStepCompletionCondition, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepIngredientID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ReadRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepCompletionConditionCalls(), 1)
	})
}

func TestRecipeManager_UpdateRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		exampleRecipeStepCompletionCondition := fakes.BuildFakeRecipeStepCompletionCondition()
		exampleInput := fakes.BuildFakeRecipeStepCompletionConditionUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, true),
			GetRecipeStepCompletionConditionFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepIngredientID string) (*types.RecipeStepCompletionCondition, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, exampleRecipeStepCompletionCondition.ID, recipeStepIngredientID)

				return exampleRecipeStepCompletionCondition, nil
			},
			UpdateRecipeStepCompletionConditionFunc: func(_ context.Context, _ string, _ *types.RecipeStepCompletionCondition) error {
				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.UpdateRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepCompletionCondition.ID, exampleOwnerID, exampleInput))

		assert.Len(t, db.GetRecipeStepCompletionConditionCalls(), 1)
		assert.Len(t, db.UpdateRecipeStepCompletionConditionCalls(), 1)
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

		err := rm.UpdateRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID, fakes.BuildFakeRecipeStepCompletionConditionUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.GetRecipeStepCompletionConditionCalls())
		assert.Empty(t, db.UpdateRecipeStepCompletionConditionCalls())
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

		err := rm.UpdateRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID, fakes.BuildFakeRecipeStepCompletionConditionUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.GetRecipeStepCompletionConditionCalls())
		assert.Empty(t, db.UpdateRecipeStepCompletionConditionCalls())
	})
}

func TestRecipeManager_ArchiveRecipeStepCompletionCondition(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleRecipeStepID := fake.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepCompletionCondition()

		db := &mealplanningmock.RepositoryMock{
			RecipeIsOwnedByFunc:  recipeIsOwnedByStub(t, exampleRecipeID, exampleOwnerID, true),
			RecipeStepExistsFunc: recipeStepExistsStub(t, exampleRecipeID, exampleRecipeStepID, true),
			ArchiveRecipeStepCompletionConditionFunc: func(_ context.Context, _ string, recipeStepID string, recipeStepIngredientID string) error {
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepIngredientID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		require.NoError(t, rm.ArchiveRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID, exampleOwnerID))

		assert.Len(t, db.ArchiveRecipeStepCompletionConditionCalls(), 1)
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

		err := rm.ArchiveRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.RecipeStepExistsCalls())
		assert.Empty(t, db.ArchiveRecipeStepCompletionConditionCalls())
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

		err := rm.ArchiveRecipeStepCompletionCondition(ctx, exampleRecipeID, exampleRecipeStepID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveRecipeStepCompletionConditionCalls())
	})
}
