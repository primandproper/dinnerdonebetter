package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v9/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeRecipeStepVessel builds a faked recipe step vessel.
func BuildFakeRecipeStepVessel() *types.RecipeStepVessel {
	minQty, maxQty := BuildFakeUint16WithOptionalMax()
	return &types.RecipeStepVessel{
		ID:                   BuildFakeID(),
		Vessel:               BuildFakeValidVessel(),
		Name:                 buildUniqueString(),
		RecipeStepProductID:  nil,
		Notes:                buildUniqueString(),
		CreatedAt:            BuildFakeTime(),
		BelongsToRecipeStep:  fake.UUID(),
		Index:                0, // Will be set from array index during recipe creation
		OptionIndex:          0, // Default to 0 for single-option items
		MinQuantity:          minQty,
		MaxQuantity:          maxQty,
		VesselPreposition:    buildUniqueString(),
		UnavailableAfterStep: fake.Bool(),
		ScaleFactor:          1.0,
	}
}

// BuildFakeRecipeStepVesselsList builds a faked RecipeStepVesselList.
func BuildFakeRecipeStepVesselsList() *filtering.QueryFilteredResult[types.RecipeStepVessel] {
	var examples []*types.RecipeStepVessel
	for range exampleQuantity {
		examples = append(examples, BuildFakeRecipeStepVessel())
	}

	return &filtering.QueryFilteredResult[types.RecipeStepVessel]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeRecipeStepVesselUpdateRequestInput builds a faked RecipeStepVesselUpdateRequestInput from a recipe step vessel.
func BuildFakeRecipeStepVesselUpdateRequestInput() *types.RecipeStepVesselUpdateRequestInput {
	recipeStepInstrument := BuildFakeRecipeStepVessel()
	return converters.ConvertRecipeStepVesselToRecipeStepVesselUpdateRequestInput(recipeStepInstrument)
}

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
