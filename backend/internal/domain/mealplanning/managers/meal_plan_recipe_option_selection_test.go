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

func TestMealPlanningManager_GetMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fakes.BuildFakeID()
		recipeStepID := fakes.BuildFakeID()
		ingredientIndex := uint16(0)
		selectionType := types.MealPlanRecipeOptionSelectionTypeIngredient
		expected := fakes.BuildFakeMealPlanRecipeOptionSelection()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanRecipeOptionSelectionFunc: func(_ context.Context, actualMealPlanOptionID string, actualRecipeStepID string, actualIngredientIndex uint16, actualSelectionType string) (*types.MealPlanRecipeOptionSelection, error) {
				assert.Equal(t, mealPlanOptionID, actualMealPlanOptionID)
				assert.Equal(t, recipeStepID, actualRecipeStepID)
				assert.Equal(t, ingredientIndex, actualIngredientIndex)
				assert.Equal(t, selectionType, actualSelectionType)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.GetMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanRecipeOptionSelectionCalls(), 1)
	})
}

func TestMealPlanningManager_GetMealPlanRecipeOptionSelectionsForMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanRecipeOptionSelectionsList()
		mealPlanOptionID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetSelectionsForMealPlanOptionFunc: func(_ context.Context, actualMealPlanOptionID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanRecipeOptionSelection], error) {
				assert.Equal(t, mealPlanOptionID, actualMealPlanOptionID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.GetMealPlanRecipeOptionSelectionsForMealPlanOption(ctx, mealPlanOptionID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetSelectionsForMealPlanOptionCalls(), 1)
	})
}

func TestMealPlanningManager_CreateMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanRecipeOptionSelection()
		fakeInput := fakes.BuildFakeMealPlanRecipeOptionSelectionCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateMealPlanRecipeOptionSelectionFunc: func(_ context.Context, _ *types.MealPlanRecipeOptionSelectionDatabaseCreationInput) (*types.MealPlanRecipeOptionSelection, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanRecipeOptionSelectionCalls(), 1)
	})
}

func TestMealPlanningManager_UpdateMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		existing := fakes.BuildFakeMealPlanRecipeOptionSelection()
		mealPlanOptionID := fakes.BuildFakeID()
		recipeStepID := fakes.BuildFakeID()
		ingredientIndex := uint16(0)
		selectionType := types.MealPlanRecipeOptionSelectionTypeIngredient
		fakeInput := fakes.BuildFakeMealPlanRecipeOptionSelectionUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanRecipeOptionSelectionFunc: func(_ context.Context, actualMealPlanOptionID string, actualRecipeStepID string, actualIngredientIndex uint16, actualSelectionType string) (*types.MealPlanRecipeOptionSelection, error) {
				assert.Equal(t, mealPlanOptionID, actualMealPlanOptionID)
				assert.Equal(t, recipeStepID, actualRecipeStepID)
				assert.Equal(t, ingredientIndex, actualIngredientIndex)
				assert.Equal(t, selectionType, actualSelectionType)

				return existing, nil
			},
			UpdateMealPlanRecipeOptionSelectionFunc: func(_ context.Context, actualMealPlanOptionID string, actualRecipeStepID string, actualIngredientIndex uint16, actualSelectionType string, _ *types.MealPlanRecipeOptionSelectionUpdateRequestInput) error {
				assert.Equal(t, mealPlanOptionID, actualMealPlanOptionID)
				assert.Equal(t, recipeStepID, actualRecipeStepID)
				assert.Equal(t, ingredientIndex, actualIngredientIndex)
				assert.Equal(t, selectionType, actualSelectionType)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.UpdateMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType, fakeInput)
		assert.NoError(t, err)

		assert.Len(t, db.GetMealPlanRecipeOptionSelectionCalls(), 1)
		assert.Len(t, db.UpdateMealPlanRecipeOptionSelectionCalls(), 1)
	})
}

func TestMealPlanningManager_ArchiveMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fakes.BuildFakeID()
		recipeStepID := fakes.BuildFakeID()
		ingredientIndex := uint16(0)
		selectionType := types.MealPlanRecipeOptionSelectionTypeIngredient

		db := &mealplanningmock.RepositoryMock{
			ArchiveMealPlanRecipeOptionSelectionFunc: func(_ context.Context, actualMealPlanOptionID string, actualRecipeStepID string, actualIngredientIndex uint16, actualSelectionType string) error {
				assert.Equal(t, mealPlanOptionID, actualMealPlanOptionID)
				assert.Equal(t, recipeStepID, actualRecipeStepID)
				assert.Equal(t, ingredientIndex, actualIngredientIndex)
				assert.Equal(t, selectionType, actualSelectionType)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType)
		assert.NoError(t, err)

		assert.Len(t, db.ArchiveMealPlanRecipeOptionSelectionCalls(), 1)
	})
}
