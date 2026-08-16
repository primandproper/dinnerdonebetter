package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
)

// BuildFakeRecipeStepIngredientCreationRequestInput builds a faked RecipeStepIngredientCreationRequestInput.
// Note: This now includes bridge table IDs since they are required.
func BuildFakeRecipeStepIngredientCreationRequestInput() *types.RecipeStepIngredientCreationRequestInput {
	recipeStepIngredient := BuildFakeRecipeStepIngredient()
	input := converters.ConvertRecipeStepIngredientToRecipeStepIngredientCreationRequestInput(recipeStepIngredient)
	// Bridge table IDs are now required
	input.ValidIngredientPreparationID = new(BuildFakeID())
	input.ValidIngredientMeasurementUnitID = new(BuildFakeID())
	return input
}

// BuildFakeRecipeStepIngredientCreationRequestInputForRecipeStepProduct builds a faked RecipeStepIngredientCreationRequestInput
// for a recipe step product (no bridge table IDs required).
func BuildFakeRecipeStepIngredientCreationRequestInputForRecipeStepProduct() *types.RecipeStepIngredientCreationRequestInput {
	recipeStepIngredient := BuildFakeRecipeStepIngredient()
	input := converters.ConvertRecipeStepIngredientToRecipeStepIngredientCreationRequestInput(recipeStepIngredient)
	input.ProductOfRecipeStepIndex = new(uint64(0))
	input.ProductOfRecipeStepProductIndex = new(uint64(0))
	return input
}
