package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/pointer"
)

// BuildFakeRecipeStepInstrument builds a faked recipe step instrument.
func BuildFakeRecipeStepInstrument() *types.RecipeStepInstrument {
	instrument := fake.BuildFakeRecord[types.RecipeStepInstrument]()

	instrument.Instrument = BuildFakeValidInstrument()
	instrument.MinQuantity, instrument.MaxQuantity = BuildFakeUint32WithOptionalMax()

	// Assigned from position by the converter during recipe creation, and option zero
	// for a single-option instrument.
	instrument.Index = 0
	instrument.OptionIndex = 0

	instrument.ScaleFactor = 1.0

	return instrument
}

// BuildFakeRecipeStepInstrumentsList builds a faked RecipeStepInstrumentList.
func BuildFakeRecipeStepInstrumentsList() *filtering.QueryFilteredResult[types.RecipeStepInstrument] {
	return fake.BuildFakePage(BuildFakeRecipeStepInstrument)
}

// BuildFakeRecipeStepInstrumentUpdateRequestInput builds a faked RecipeStepInstrumentUpdateRequestInput from a recipe step instrument.
func BuildFakeRecipeStepInstrumentUpdateRequestInput() *types.RecipeStepInstrumentUpdateRequestInput {
	recipeStepInstrument := BuildFakeRecipeStepInstrument()

	return converters.ConvertRecipeStepInstrumentToRecipeStepInstrumentUpdateRequestInput(recipeStepInstrument)
}

// BuildFakeRecipeStepInstrumentCreationRequestInput builds a faked RecipeStepInstrumentCreationRequestInput.
//
// Hand-written past the conversion: the bridge table ID is required on the input and
// has no counterpart on the record it was converted from.
func BuildFakeRecipeStepInstrumentCreationRequestInput() *types.RecipeStepInstrumentCreationRequestInput {
	recipeStepInstrument := BuildFakeRecipeStepInstrument()
	input := converters.ConvertRecipeStepInstrumentToRecipeStepInstrumentCreationRequestInput(recipeStepInstrument)
	input.ValidPreparationInstrumentID = pointer.To(fake.BuildFakeID())

	return input
}

// BuildFakeRecipeStepInstrumentCreationRequestInputForRecipeStepProduct builds a faked RecipeStepInstrumentCreationRequestInput
// for a recipe step product (no bridge table IDs required).
func BuildFakeRecipeStepInstrumentCreationRequestInputForRecipeStepProduct() *types.RecipeStepInstrumentCreationRequestInput {
	recipeStepInstrument := BuildFakeRecipeStepInstrument()
	input := converters.ConvertRecipeStepInstrumentToRecipeStepInstrumentCreationRequestInput(recipeStepInstrument)
	input.ProductOfRecipeStepIndex = pointer.To(uint64(0))
	input.ProductOfRecipeStepProductIndex = pointer.To(uint64(0))

	return input
}
