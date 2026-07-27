package managers

import (
	"context"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v7/filtering"

	"github.com/stretchr/testify/assert"
)

func TestRecipeManager_ListRecipeStepIngredients(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		expected := fakes.BuildFakeRecipeStepIngredientsList()
		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepIngredientsFunc: func(_ context.Context, recipeID string, recipeStepID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepIngredient], error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ListRecipeStepIngredients(ctx, exampleRecipeID, exampleRecipeStepID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepIngredientsCalls(), 1)
	})
}

func TestRecipeManager_CreateRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepIngredient()
		fakeInput := fakes.BuildFakeRecipeStepIngredientCreationRequestInput()
		fakeInput.Index = new(uint16(0))

		fakeValidIngredientPreparation := fakes.BuildFakeValidIngredientPreparation()
		fakeValidIngredientMeasurementUnit := fakes.BuildFakeValidIngredientMeasurementUnit()

		db := &mealplanningmock.RepositoryMock{
			GetValidIngredientPreparationFunc: func(_ context.Context, validIngredientPreparationID string) (*types.ValidIngredientPreparation, error) {
				assert.Equal(t, *fakeInput.ValidIngredientPreparationID, validIngredientPreparationID)

				return fakeValidIngredientPreparation, nil
			},
			GetValidIngredientMeasurementUnitFunc: func(_ context.Context, validIngredientMeasurementUnitID string) (*types.ValidIngredientMeasurementUnit, error) {
				assert.Equal(t, *fakeInput.ValidIngredientMeasurementUnitID, validIngredientMeasurementUnitID)

				return fakeValidIngredientMeasurementUnit, nil
			},
			CreateRecipeStepIngredientFunc: func(_ context.Context, _ *types.RecipeStepIngredientDatabaseCreationInput) (*types.RecipeStepIngredient, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.CreateRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetValidIngredientPreparationCalls(), 1)
		assert.Len(t, db.GetValidIngredientMeasurementUnitCalls(), 1)
		assert.Len(t, db.CreateRecipeStepIngredientCalls(), 1)
	})
}

func TestRecipeManager_ReadRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepIngredient()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepIngredientFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepIngredientID string) (*types.RecipeStepIngredient, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepIngredientID)

				return expected, nil
			},
		}
		attachRepositoryToManager(rm, db)

		actual, err := rm.ReadRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetRecipeStepIngredientCalls(), 1)
	})
}

func TestRecipeManager_UpdateRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		exampleRecipeStepIngredient := fakes.BuildFakeRecipeStepIngredient()
		exampleInput := fakes.BuildFakeRecipeStepIngredientUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetRecipeStepIngredientFunc: func(_ context.Context, recipeID string, recipeStepID string, recipeStepIngredientID string) (*types.RecipeStepIngredient, error) {
				assert.Equal(t, exampleRecipeID, recipeID)
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, exampleRecipeStepIngredient.ID, recipeStepIngredientID)

				return exampleRecipeStepIngredient, nil
			},
			UpdateRecipeStepIngredientFunc: func(_ context.Context, _ *types.RecipeStepIngredient) error {
				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		assert.NoError(t, rm.UpdateRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, exampleRecipeStepIngredient.ID, exampleInput))

		assert.Len(t, db.GetRecipeStepIngredientCalls(), 1)
		assert.Len(t, db.UpdateRecipeStepIngredientCalls(), 1)
	})
}

func TestRecipeManager_ArchiveRecipeStepIngredient(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		rm := buildRecipeManagerForTest(t)

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipeStepID := fakes.BuildFakeID()
		expected := fakes.BuildFakeRecipeStepIngredient()

		db := &mealplanningmock.RepositoryMock{
			ArchiveRecipeStepIngredientFunc: func(_ context.Context, recipeStepID string, recipeStepIngredientID string) error {
				assert.Equal(t, exampleRecipeStepID, recipeStepID)
				assert.Equal(t, expected.ID, recipeStepIngredientID)

				return nil
			},
		}
		attachRepositoryToManager(rm, db)

		assert.NoError(t, rm.ArchiveRecipeStepIngredient(ctx, exampleRecipeID, exampleRecipeStepID, expected.ID))

		assert.Len(t, db.ArchiveRecipeStepIngredientCalls(), 1)
	})
}
