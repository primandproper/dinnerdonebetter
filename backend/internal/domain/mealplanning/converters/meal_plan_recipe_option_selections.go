package converters

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/identifiers"
)

// ConvertMealPlanRecipeOptionSelectionToMealPlanRecipeOptionSelectionDatabaseCreationInput builds a MealPlanRecipeOptionSelectionDatabaseCreationInput from a MealPlanRecipeOptionSelection.
func ConvertMealPlanRecipeOptionSelectionToMealPlanRecipeOptionSelectionDatabaseCreationInput(input *mealplanning.MealPlanRecipeOptionSelection) *mealplanning.MealPlanRecipeOptionSelectionDatabaseCreationInput {
	return &mealplanning.MealPlanRecipeOptionSelectionDatabaseCreationInput{
		ID:                      input.ID,
		BelongsToMealPlanOption: input.BelongsToMealPlanOption,
		RecipeID:                input.RecipeID,
		RecipeStepID:            input.RecipeStepID,
		IngredientIndex:         input.IngredientIndex,
		SelectedOptionIndex:     input.SelectedOptionIndex,
		SelectionType:           input.SelectionType,
	}
}

// ConvertMealPlanRecipeOptionSelectionDatabaseCreationInputToMealPlanRecipeOptionSelectionDatabaseCreationInput creates a new DatabaseCreationInput with a new ID.
func ConvertMealPlanRecipeOptionSelectionDatabaseCreationInputToMealPlanRecipeOptionSelectionDatabaseCreationInput(input *mealplanning.MealPlanRecipeOptionSelectionCreationRequestInput, mealPlanOptionID string) *mealplanning.MealPlanRecipeOptionSelectionDatabaseCreationInput {
	return &mealplanning.MealPlanRecipeOptionSelectionDatabaseCreationInput{
		ID:                      identifiers.New(),
		BelongsToMealPlanOption: mealPlanOptionID,
		RecipeID:                input.RecipeID,
		RecipeStepID:            input.RecipeStepID,
		IngredientIndex:         input.IngredientIndex,
		SelectedOptionIndex:     input.SelectedOptionIndex,
		SelectionType:           input.SelectionType,
	}
}
