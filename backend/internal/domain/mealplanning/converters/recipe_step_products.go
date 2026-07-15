package converters

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v4/identifiers"
)

// ConvertRecipeStepProductToRecipeStepProductUpdateRequestInput creates a RecipeStepProductUpdateRequestInput from a RecipeStepProduct.
func ConvertRecipeStepProductToRecipeStepProductUpdateRequestInput(input *types.RecipeStepProduct) *types.RecipeStepProductUpdateRequestInput {
	if input == nil {
		return nil
	}

	x := &types.RecipeStepProductUpdateRequestInput{
		Name:                           &input.Name,
		Type:                           &input.Type,
		MeasurementUnitID:              &input.MeasurementUnit.ID,
		QuantityNotes:                  &input.QuantityNotes,
		BelongsToRecipeStep:            &input.BelongsToRecipeStep,
		Compostable:                    &input.Compostable,
		MinMeasurementQuantity:         input.MinMeasurementQuantity,
		MaxMeasurementQuantity:         input.MaxMeasurementQuantity,
		MinItemQuantity:                input.MinItemQuantity,
		MaxItemQuantity:                input.MaxItemQuantity,
		MinStorageDurationInSeconds:    input.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    input.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		StorageInstructions:            &input.StorageInstructions,
		IsWaste:                        &input.IsWaste,
		IsLiquid:                       &input.IsLiquid,
		Index:                          &input.Index,
		ContainedInVesselIndex:         input.ContainedInVesselIndex,
	}

	return x
}

// ConvertRecipeStepProductCreationInputToRecipeStepProductDatabaseCreationInput creates a RecipeStepProductDatabaseCreationInput from a RecipeStepProductCreationRequestInput.
func ConvertRecipeStepProductCreationInputToRecipeStepProductDatabaseCreationInput(input *types.RecipeStepProductCreationRequestInput) *types.RecipeStepProductDatabaseCreationInput {
	if input == nil {
		return nil
	}

	x := &types.RecipeStepProductDatabaseCreationInput{
		ID:                             identifiers.New(),
		Name:                           input.Name,
		Type:                           input.Type,
		MeasurementUnitID:              input.MeasurementUnitID,
		QuantityNotes:                  input.QuantityNotes,
		Compostable:                    input.Compostable,
		MinMeasurementQuantity:         input.MinMeasurementQuantity,
		MaxMeasurementQuantity:         input.MaxMeasurementQuantity,
		MinItemQuantity:                input.MinItemQuantity,
		MaxItemQuantity:                input.MaxItemQuantity,
		MinStorageDurationInSeconds:    input.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    input.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		StorageInstructions:            input.StorageInstructions,
		IsWaste:                        input.IsWaste,
		IsLiquid:                       input.IsLiquid,
		Index:                          input.Index,
		ContainedInVesselIndex:         input.ContainedInVesselIndex,
	}

	return x
}

// ConvertRecipeStepProductToRecipeStepProductCreationRequestInput builds a RecipeStepProductCreationRequestInput from a RecipeStepProduct.
func ConvertRecipeStepProductToRecipeStepProductCreationRequestInput(recipeStepProduct *types.RecipeStepProduct) *types.RecipeStepProductCreationRequestInput {
	return &types.RecipeStepProductCreationRequestInput{
		Name:                           recipeStepProduct.Name,
		Type:                           recipeStepProduct.Type,
		QuantityNotes:                  recipeStepProduct.QuantityNotes,
		MeasurementUnitID:              &recipeStepProduct.MeasurementUnit.ID,
		Compostable:                    recipeStepProduct.Compostable,
		MinMeasurementQuantity:         recipeStepProduct.MinMeasurementQuantity,
		MaxMeasurementQuantity:         recipeStepProduct.MaxMeasurementQuantity,
		MinItemQuantity:                recipeStepProduct.MinItemQuantity,
		MaxItemQuantity:                recipeStepProduct.MaxItemQuantity,
		MinStorageDurationInSeconds:    recipeStepProduct.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    recipeStepProduct.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: recipeStepProduct.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: recipeStepProduct.MaxStorageTemperatureInCelsius,
		StorageInstructions:            recipeStepProduct.StorageInstructions,
		IsWaste:                        recipeStepProduct.IsWaste,
		IsLiquid:                       recipeStepProduct.IsLiquid,
		Index:                          recipeStepProduct.Index,
		ContainedInVesselIndex:         recipeStepProduct.ContainedInVesselIndex,
	}
}

// ConvertRecipeStepProductToRecipeStepProductDatabaseCreationInput builds a RecipeStepProductDatabaseCreationInput from a RecipeStepProduct.
func ConvertRecipeStepProductToRecipeStepProductDatabaseCreationInput(recipeStepProduct *types.RecipeStepProduct) *types.RecipeStepProductDatabaseCreationInput {
	var measurementUnitID *string
	if recipeStepProduct.MeasurementUnit != nil {
		measurementUnitID = &recipeStepProduct.MeasurementUnit.ID
	}

	return &types.RecipeStepProductDatabaseCreationInput{
		ID:                             recipeStepProduct.ID,
		Name:                           recipeStepProduct.Name,
		Type:                           recipeStepProduct.Type,
		QuantityNotes:                  recipeStepProduct.QuantityNotes,
		MeasurementUnitID:              measurementUnitID,
		BelongsToRecipeStep:            recipeStepProduct.BelongsToRecipeStep,
		Compostable:                    recipeStepProduct.Compostable,
		MinMeasurementQuantity:         recipeStepProduct.MinMeasurementQuantity,
		MaxMeasurementQuantity:         recipeStepProduct.MaxMeasurementQuantity,
		MinItemQuantity:                recipeStepProduct.MinItemQuantity,
		MaxItemQuantity:                recipeStepProduct.MaxItemQuantity,
		MinStorageDurationInSeconds:    recipeStepProduct.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    recipeStepProduct.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: recipeStepProduct.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: recipeStepProduct.MaxStorageTemperatureInCelsius,
		StorageInstructions:            recipeStepProduct.StorageInstructions,
		IsWaste:                        recipeStepProduct.IsWaste,
		IsLiquid:                       recipeStepProduct.IsLiquid,
		Index:                          recipeStepProduct.Index,
		ContainedInVesselIndex:         recipeStepProduct.ContainedInVesselIndex,
	}
}
