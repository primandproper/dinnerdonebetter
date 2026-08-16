package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
)

// BuildFakeRecipeStepInstrumentCreationRequestInput builds a faked RecipeStepInstrumentCreationRequestInput.
// Note: This now includes bridge table IDs since they are required.
func BuildFakeRecipeStepInstrumentCreationRequestInput() *types.RecipeStepInstrumentCreationRequestInput {
	recipeStepInstrument := BuildFakeRecipeStepInstrument()
	input := converters.ConvertRecipeStepInstrumentToRecipeStepInstrumentCreationRequestInput(recipeStepInstrument)
	// Bridge table ID is now required
	input.ValidPreparationInstrumentID = new(BuildFakeID())
	return input
}

// BuildFakeRecipeStepInstrumentCreationRequestInputForRecipeStepProduct builds a faked RecipeStepInstrumentCreationRequestInput
// for a recipe step product (no bridge table IDs required).
func BuildFakeRecipeStepInstrumentCreationRequestInputForRecipeStepProduct() *types.RecipeStepInstrumentCreationRequestInput {
	recipeStepInstrument := BuildFakeRecipeStepInstrument()
	input := converters.ConvertRecipeStepInstrumentToRecipeStepInstrumentCreationRequestInput(recipeStepInstrument)
	input.ProductOfRecipeStepIndex = new(uint64(0))
	input.ProductOfRecipeStepProductIndex = new(uint64(0))
	return input
}
