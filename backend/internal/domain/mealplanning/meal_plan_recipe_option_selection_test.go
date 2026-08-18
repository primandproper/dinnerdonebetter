package mealplanning

import (
	"testing"

	"github.com/primandproper/platform-go/v11/identifiers"

	"github.com/stretchr/testify/assert"
)

func TestMealPlanRecipeOptionSelectionCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionCreationRequestInput{
			RecipeID:            t.Name() + "_recipe",
			RecipeStepID:        t.Name() + "_step",
			IngredientIndex:     0,
			SelectionType:       MealPlanRecipeOptionSelectionTypeIngredient,
			SelectedOptionIndex: 0,
		}

		assert.NoError(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with ingredient selection type", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionCreationRequestInput{
			RecipeID:            t.Name() + "_recipe",
			RecipeStepID:        t.Name() + "_step",
			IngredientIndex:     1,
			SelectionType:       MealPlanRecipeOptionSelectionTypeIngredient,
			SelectedOptionIndex: 2,
		}

		assert.NoError(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with instrument selection type", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionCreationRequestInput{
			RecipeID:            t.Name() + "_recipe",
			RecipeStepID:        t.Name() + "_step",
			IngredientIndex:     0,
			SelectionType:       MealPlanRecipeOptionSelectionTypeInstrument,
			SelectedOptionIndex: 1,
		}

		assert.NoError(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with vessel selection type", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionCreationRequestInput{
			RecipeID:            t.Name() + "_recipe",
			RecipeStepID:        t.Name() + "_step",
			IngredientIndex:     2,
			SelectionType:       MealPlanRecipeOptionSelectionTypeVessel,
			SelectedOptionIndex: 0,
		}

		assert.NoError(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionCreationRequestInput{}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with missing recipe ID", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionCreationRequestInput{
			RecipeStepID:        t.Name() + "_step",
			IngredientIndex:     0,
			SelectionType:       MealPlanRecipeOptionSelectionTypeIngredient,
			SelectedOptionIndex: 0,
		}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with missing recipe step ID", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionCreationRequestInput{
			RecipeID:            t.Name() + "_recipe",
			IngredientIndex:     0,
			SelectionType:       MealPlanRecipeOptionSelectionTypeIngredient,
			SelectedOptionIndex: 0,
		}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with missing selection type", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionCreationRequestInput{
			RecipeID:            t.Name() + "_recipe",
			RecipeStepID:        t.Name() + "_step",
			IngredientIndex:     0,
			SelectedOptionIndex: 0,
		}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with invalid selection type", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionCreationRequestInput{
			RecipeID:            t.Name() + "_recipe",
			RecipeStepID:        t.Name() + "_step",
			IngredientIndex:     0,
			SelectionType:       "invalid_type",
			SelectedOptionIndex: 0,
		}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})
}

func TestMealPlanRecipeOptionSelectionDatabaseCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionDatabaseCreationInput{
			ID:                      identifiers.New(),
			BelongsToMealPlanOption: t.Name() + "_option",
			RecipeID:                t.Name() + "_recipe",
			RecipeStepID:            t.Name() + "_step",
			IngredientIndex:         0,
			SelectionType:           MealPlanRecipeOptionSelectionTypeIngredient,
			SelectedOptionIndex:     0,
		}

		assert.NoError(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with ingredient selection type", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionDatabaseCreationInput{
			ID:                      identifiers.New(),
			BelongsToMealPlanOption: t.Name() + "_option",
			RecipeID:                t.Name() + "_recipe",
			RecipeStepID:            t.Name() + "_step",
			IngredientIndex:         1,
			SelectionType:           MealPlanRecipeOptionSelectionTypeIngredient,
			SelectedOptionIndex:     2,
		}

		assert.NoError(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with instrument selection type", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionDatabaseCreationInput{
			ID:                      identifiers.New(),
			BelongsToMealPlanOption: t.Name() + "_option",
			RecipeID:                t.Name() + "_recipe",
			RecipeStepID:            t.Name() + "_step",
			IngredientIndex:         0,
			SelectionType:           MealPlanRecipeOptionSelectionTypeInstrument,
			SelectedOptionIndex:     1,
		}

		assert.NoError(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with vessel selection type", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionDatabaseCreationInput{
			ID:                      identifiers.New(),
			BelongsToMealPlanOption: t.Name() + "_option",
			RecipeID:                t.Name() + "_recipe",
			RecipeStepID:            t.Name() + "_step",
			IngredientIndex:         2,
			SelectionType:           MealPlanRecipeOptionSelectionTypeVessel,
			SelectedOptionIndex:     0,
		}

		assert.NoError(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionDatabaseCreationInput{}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with missing ID", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionDatabaseCreationInput{
			BelongsToMealPlanOption: t.Name() + "_option",
			RecipeID:                t.Name() + "_recipe",
			RecipeStepID:            t.Name() + "_step",
			IngredientIndex:         0,
			SelectionType:           MealPlanRecipeOptionSelectionTypeIngredient,
			SelectedOptionIndex:     0,
		}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with missing belongs to meal plan option", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionDatabaseCreationInput{
			ID:                  identifiers.New(),
			RecipeID:            t.Name() + "_recipe",
			RecipeStepID:        t.Name() + "_step",
			IngredientIndex:     0,
			SelectionType:       MealPlanRecipeOptionSelectionTypeIngredient,
			SelectedOptionIndex: 0,
		}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with missing recipe ID", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionDatabaseCreationInput{
			ID:                      identifiers.New(),
			BelongsToMealPlanOption: t.Name() + "_option",
			RecipeStepID:            t.Name() + "_step",
			IngredientIndex:         0,
			SelectionType:           MealPlanRecipeOptionSelectionTypeIngredient,
			SelectedOptionIndex:     0,
		}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with missing recipe step ID", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionDatabaseCreationInput{
			ID:                      identifiers.New(),
			BelongsToMealPlanOption: t.Name() + "_option",
			RecipeID:                t.Name() + "_recipe",
			IngredientIndex:         0,
			SelectionType:           MealPlanRecipeOptionSelectionTypeIngredient,
			SelectedOptionIndex:     0,
		}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with missing selection type", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionDatabaseCreationInput{
			ID:                      identifiers.New(),
			BelongsToMealPlanOption: t.Name() + "_option",
			RecipeID:                t.Name() + "_recipe",
			RecipeStepID:            t.Name() + "_step",
			IngredientIndex:         0,
			SelectedOptionIndex:     0,
		}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with invalid selection type", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanRecipeOptionSelectionDatabaseCreationInput{
			ID:                      identifiers.New(),
			BelongsToMealPlanOption: t.Name() + "_option",
			RecipeID:                t.Name() + "_recipe",
			RecipeStepID:            t.Name() + "_step",
			IngredientIndex:         0,
			SelectionType:           "invalid_type",
			SelectedOptionIndex:     0,
		}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})
}
