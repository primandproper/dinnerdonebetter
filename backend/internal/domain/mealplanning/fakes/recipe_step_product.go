package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeRecipeStepProduct builds a faked recipe step product.
func BuildFakeRecipeStepProduct() *types.RecipeStepProduct {
	measurementMin := float32(buildFakeNumber())
	measurementMax := measurementMin + float32(buildFakeNumber())
	itemMin := float32(buildFakeNumber())
	itemMax := itemMin + float32(buildFakeNumber())
	storageTempMin := float32(buildFakeNumber())
	storageTempMax := storageTempMin + float32(buildFakeNumber())
	storageDurationMax := uint32(buildFakeNumber())

	p := &types.RecipeStepProduct{
		ID:                             BuildFakeID(),
		Name:                           buildUniqueString(),
		Type:                           types.RecipeStepProductIngredientType,
		QuantityNotes:                  buildUniqueString(),
		MeasurementUnit:                BuildFakeValidMeasurementUnit(),
		CreatedAt:                      BuildFakeTime(),
		BelongsToRecipeStep:            fake.UUID(),
		Compostable:                    fake.Bool(),
		IsLiquid:                       fake.Bool(),
		IsWaste:                        fake.Bool(),
		MinMeasurementQuantity:         &measurementMin,
		MaxMeasurementQuantity:         &measurementMax,
		MinItemQuantity:                &itemMin,
		MaxItemQuantity:                &itemMax,
		MinStorageTemperatureInCelsius: &storageTempMin,
		MaxStorageTemperatureInCelsius: &storageTempMax,
		MaxStorageDurationInSeconds:    &storageDurationMax,
		StorageInstructions:            buildUniqueString(),
		Index:                          fake.Uint16(),
		ContainedInVesselIndex:         new(fake.Uint16()),
	}

	return p
}

// BuildFakeRecipeStepProductsList builds a faked RecipeStepProductList.
func BuildFakeRecipeStepProductsList() *filtering.QueryFilteredResult[types.RecipeStepProduct] {
	var examples []*types.RecipeStepProduct
	for range exampleQuantity {
		examples = append(examples, BuildFakeRecipeStepProduct())
	}

	return &filtering.QueryFilteredResult[types.RecipeStepProduct]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
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
