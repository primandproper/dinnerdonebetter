package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/pointer"
)

// BuildFakeRecipeStepVessel builds a faked recipe step vessel.
func BuildFakeRecipeStepVessel() *types.RecipeStepVessel {
	vessel := fake.BuildFakeRecord[types.RecipeStepVessel]()

	vessel.Vessel = BuildFakeValidVessel()
	vessel.MinQuantity, vessel.MaxQuantity = BuildFakeUint16WithOptionalMax()

	// Assigned from position by the converter during recipe creation, and option zero
	// for a single-option vessel.
	vessel.Index = 0
	vessel.OptionIndex = 0

	vessel.ScaleFactor = 1.0

	return vessel
}

// BuildFakeRecipeStepVesselsList builds a faked RecipeStepVesselList.
func BuildFakeRecipeStepVesselsList() *filtering.QueryFilteredResult[types.RecipeStepVessel] {
	return fake.BuildFakePage(BuildFakeRecipeStepVessel)
}

// BuildFakeRecipeStepVesselUpdateRequestInput builds a faked RecipeStepVesselUpdateRequestInput from a recipe step vessel.
func BuildFakeRecipeStepVesselUpdateRequestInput() *types.RecipeStepVesselUpdateRequestInput {
	recipeStepVessel := BuildFakeRecipeStepVessel()

	return converters.ConvertRecipeStepVesselToRecipeStepVesselUpdateRequestInput(recipeStepVessel)
}

// BuildFakeRecipeStepVesselCreationRequestInput builds a faked RecipeStepVesselCreationRequestInput.
//
// Hand-written past the conversion: the bridge table ID is required on the input and
// has no counterpart on the record it was converted from.
func BuildFakeRecipeStepVesselCreationRequestInput() *types.RecipeStepVesselCreationRequestInput {
	recipeStepVessel := BuildFakeRecipeStepVessel()
	input := converters.ConvertRecipeStepVesselToRecipeStepVesselCreationRequestInput(recipeStepVessel)
	input.ValidPreparationVesselID = pointer.To(fake.BuildFakeID())

	return input
}

// BuildFakeRecipeStepVesselCreationRequestInputForRecipeStepProduct builds a faked RecipeStepVesselCreationRequestInput
// for a recipe step product (no bridge table IDs required).
func BuildFakeRecipeStepVesselCreationRequestInputForRecipeStepProduct() *types.RecipeStepVesselCreationRequestInput {
	recipeStepVessel := BuildFakeRecipeStepVessel()
	input := converters.ConvertRecipeStepVesselToRecipeStepVesselCreationRequestInput(recipeStepVessel)
	input.ProductOfRecipeStepIndex = pointer.To(uint64(0))
	input.ProductOfRecipeStepProductIndex = pointer.To(uint64(0))

	return input
}
