package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v7/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeRecipeStepIngredient builds a faked recipe step ingredient.
// NOTE: this currently represents a typical recipe step ingredient with a valid ingredient and not a product.
func BuildFakeRecipeStepIngredient() *types.RecipeStepIngredient {
	minQty, maxQty := BuildFakeFloat32WithOptionalMax()
	return &types.RecipeStepIngredient{
		ID:                     BuildFakeID(),
		Name:                   buildUniqueString(),
		Ingredient:             BuildFakeValidIngredient(),
		MeasurementUnit:        *BuildFakeValidMeasurementUnit(),
		MinQuantity:            minQty,
		MaxQuantity:            maxQty,
		QuantityNotes:          buildUniqueString(),
		Optional:               fake.Bool(),
		IngredientNotes:        buildUniqueString(),
		CreatedAt:              BuildFakeTime(),
		BelongsToRecipeStep:    BuildFakeID(),
		Index:                  0, // Will be set from array index during recipe creation (via converter)
		OptionIndex:            0, // Default to 0 for single-option items
		VesselIndex:            new(fake.Uint16()),
		ToTaste:                fake.Bool(),
		ProductPercentageToUse: new(float32(buildFakeNumber())),
		ScaleFactor:            1.0,
	}
}

// BuildFakeRecipeStepIngredientsList builds a faked RecipeStepIngredientList.
func BuildFakeRecipeStepIngredientsList() *filtering.QueryFilteredResult[types.RecipeStepIngredient] {
	var examples []*types.RecipeStepIngredient
	for range exampleQuantity {
		examples = append(examples, BuildFakeRecipeStepIngredient())
	}

	return &filtering.QueryFilteredResult[types.RecipeStepIngredient]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeRecipeStepIngredientUpdateRequestInput builds a faked RecipeStepIngredientUpdateRequestInput from a recipe step ingredient.
func BuildFakeRecipeStepIngredientUpdateRequestInput() *types.RecipeStepIngredientUpdateRequestInput {
	recipeStepIngredient := BuildFakeRecipeStepIngredient()
	return converters.ConvertRecipeStepIngredientToRecipeStepIngredientUpdateRequestInput(recipeStepIngredient)
}

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
