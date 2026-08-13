package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v10/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeRecipeStepInstrument builds a faked recipe step instrument.
func BuildFakeRecipeStepInstrument() *types.RecipeStepInstrument {
	minQty, maxQty := BuildFakeUint32WithOptionalMax()
	return &types.RecipeStepInstrument{
		ID:                  BuildFakeID(),
		Instrument:          BuildFakeValidInstrument(),
		Name:                buildUniqueString(),
		RecipeStepProductID: nil,
		Notes:               buildUniqueString(),
		PreferenceRank:      fake.Uint8(),
		CreatedAt:           BuildFakeTime(),
		BelongsToRecipeStep: fake.UUID(),
		Index:               0,
		OptionIndex:         0,
		Optional:            fake.Bool(),
		MinQuantity:         minQty,
		MaxQuantity:         maxQty,
		ScaleFactor:         1.0,
	}
}

// BuildFakeRecipeStepInstrumentsList builds a faked RecipeStepInstrumentList.
func BuildFakeRecipeStepInstrumentsList() *filtering.QueryFilteredResult[types.RecipeStepInstrument] {
	var examples []*types.RecipeStepInstrument
	for range exampleQuantity {
		examples = append(examples, BuildFakeRecipeStepInstrument())
	}

	return &filtering.QueryFilteredResult[types.RecipeStepInstrument]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeRecipeStepInstrumentUpdateRequestInput builds a faked RecipeStepInstrumentUpdateRequestInput from a recipe step instrument.
func BuildFakeRecipeStepInstrumentUpdateRequestInput() *types.RecipeStepInstrumentUpdateRequestInput {
	recipeStepInstrument := BuildFakeRecipeStepInstrument()
	return converters.ConvertRecipeStepInstrumentToRecipeStepInstrumentUpdateRequestInput(recipeStepInstrument)
}

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
