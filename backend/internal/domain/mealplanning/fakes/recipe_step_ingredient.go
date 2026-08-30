package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/pointer"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeRecipeStepIngredient builds a faked recipe step ingredient.
//
// NOTE: this currently represents a typical recipe step ingredient with a valid ingredient and not a product.
func BuildFakeRecipeStepIngredient() *types.RecipeStepIngredient {
	ingredient := fake.BuildFakeRecord[types.RecipeStepIngredient]()

	// What is being measured and what it is measured in. Both are optional on the type
	// and neither is optional in practice: the scaling code multiplies a quantity by a
	// unit it reads from here.
	ingredient.Ingredient = BuildFakeValidIngredient()
	ingredient.MeasurementUnit = *BuildFakeValidMeasurementUnit()
	ingredient.MinQuantity, ingredient.MaxQuantity = BuildFakeFloat32WithOptionalMax()
	ingredient.ProductPercentageToUse = pointer.To(float32(fake.BuildFakeNumber()))
	ingredient.VesselIndex = pointer.To(gofakeit.Uint16())

	// Both indices are assigned from position by the converter during recipe creation,
	// and a single-option ingredient is option zero.
	ingredient.Index = 0
	ingredient.OptionIndex = 0

	// Unscaled, so a test that scales a recipe measures its own factor.
	ingredient.ScaleFactor = 1.0

	return ingredient
}

// BuildFakeRecipeStepIngredientsList builds a faked RecipeStepIngredientList.
func BuildFakeRecipeStepIngredientsList() *filtering.QueryFilteredResult[types.RecipeStepIngredient] {
	return fake.BuildFakePage(BuildFakeRecipeStepIngredient)
}

// BuildFakeRecipeStepIngredientUpdateRequestInput builds a faked RecipeStepIngredientUpdateRequestInput from a recipe step ingredient.
func BuildFakeRecipeStepIngredientUpdateRequestInput() *types.RecipeStepIngredientUpdateRequestInput {
	recipeStepIngredient := BuildFakeRecipeStepIngredient()

	return converters.ConvertRecipeStepIngredientToRecipeStepIngredientUpdateRequestInput(recipeStepIngredient)
}

// BuildFakeRecipeStepIngredientCreationRequestInput builds a faked RecipeStepIngredientCreationRequestInput.
//
// Hand-written past the conversion: the bridge table IDs are required on the input and
// have no counterpart on the record it was converted from.
func BuildFakeRecipeStepIngredientCreationRequestInput() *types.RecipeStepIngredientCreationRequestInput {
	recipeStepIngredient := BuildFakeRecipeStepIngredient()
	input := converters.ConvertRecipeStepIngredientToRecipeStepIngredientCreationRequestInput(recipeStepIngredient)
	input.ValidIngredientPreparationID = pointer.To(fake.BuildFakeID())
	input.ValidIngredientMeasurementUnitID = pointer.To(fake.BuildFakeID())

	return input
}

// BuildFakeRecipeStepIngredientCreationRequestInputForRecipeStepProduct builds a faked RecipeStepIngredientCreationRequestInput
// for a recipe step product (no bridge table IDs required).
func BuildFakeRecipeStepIngredientCreationRequestInputForRecipeStepProduct() *types.RecipeStepIngredientCreationRequestInput {
	recipeStepIngredient := BuildFakeRecipeStepIngredient()
	input := converters.ConvertRecipeStepIngredientToRecipeStepIngredientCreationRequestInput(recipeStepIngredient)
	input.ProductOfRecipeStepIndex = pointer.To(uint64(0))
	input.ProductOfRecipeStepProductIndex = pointer.To(uint64(0))

	return input
}
