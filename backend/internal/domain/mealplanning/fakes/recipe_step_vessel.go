package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
)

// BuildFakeRecipeStepVesselCreationRequestInput builds a faked RecipeStepVesselCreationRequestInput.
// Note: This now includes bridge table IDs since they are required.
func BuildFakeRecipeStepVesselCreationRequestInput() *types.RecipeStepVesselCreationRequestInput {
	recipeStepVessel := BuildFakeRecipeStepVessel()
	input := converters.ConvertRecipeStepVesselToRecipeStepVesselCreationRequestInput(recipeStepVessel)
	// Bridge table ID is now required
	input.ValidPreparationVesselID = new(BuildFakeID())
	return input
}

// BuildFakeRecipeStepVesselCreationRequestInputForRecipeStepProduct builds a faked RecipeStepVesselCreationRequestInput
// for a recipe step product (no bridge table IDs required).
func BuildFakeRecipeStepVesselCreationRequestInputForRecipeStepProduct() *types.RecipeStepVesselCreationRequestInput {
	recipeStepVessel := BuildFakeRecipeStepVessel()
	input := converters.ConvertRecipeStepVesselToRecipeStepVesselCreationRequestInput(recipeStepVessel)
	input.ProductOfRecipeStepIndex = new(uint64(0))
	input.ProductOfRecipeStepProductIndex = new(uint64(0))
	return input
}
