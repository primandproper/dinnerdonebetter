package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/pointer"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeRecipeStepProduct builds a faked recipe step product.
func BuildFakeRecipeStepProduct() *types.RecipeStepProduct {
	product := fake.BuildFakeRecord[types.RecipeStepProduct]()

	// One of the three kinds of thing a step can produce, which the type validates.
	product.Type = types.RecipeStepProductIngredientType
	product.MeasurementUnit = BuildFakeValidMeasurementUnit()

	// Three ranges rather than six independent numbers.
	product.MinMeasurementQuantity, product.MaxMeasurementQuantity = BuildFakeOptionalFloat32MinMax()
	product.MinItemQuantity, product.MaxItemQuantity = BuildFakeOptionalFloat32MinMax()
	product.MinStorageTemperatureInCelsius, product.MaxStorageTemperatureInCelsius = BuildFakeOptionalFloat32MinMax()
	product.MaxStorageDurationInSeconds = pointer.To(uint32(fake.BuildFakeNumber()))

	product.ContainedInVesselIndex = pointer.To(gofakeit.Uint16())

	return product
}

// BuildFakeRecipeStepProductsList builds a faked RecipeStepProductList.
func BuildFakeRecipeStepProductsList() *filtering.QueryFilteredResult[types.RecipeStepProduct] {
	return fake.BuildFakePage(BuildFakeRecipeStepProduct)
}

// BuildFakeRecipeStepProductUpdateRequestInput builds a faked RecipeStepProductUpdateRequestInput from a recipe step product.
func BuildFakeRecipeStepProductUpdateRequestInput() *types.RecipeStepProductUpdateRequestInput {
	recipeStepProduct := BuildFakeRecipeStepProduct()

	return converters.ConvertRecipeStepProductToRecipeStepProductUpdateRequestInput(recipeStepProduct)
}

// BuildFakeRecipeStepProductCreationRequestInput builds a faked RecipeStepProductCreationRequestInput.
func BuildFakeRecipeStepProductCreationRequestInput() *types.RecipeStepProductCreationRequestInput {
	recipeStepProduct := BuildFakeRecipeStepProduct()

	return converters.ConvertRecipeStepProductToRecipeStepProductCreationRequestInput(recipeStepProduct)
}
