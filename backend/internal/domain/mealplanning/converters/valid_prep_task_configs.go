package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v10/identifiers"
)

// ConvertValidPrepTaskConfigCreationRequestInputToValidPrepTaskConfigDatabaseCreationInput creates a ValidPrepTaskConfigDatabaseCreationInput from a ValidPrepTaskConfigCreationRequestInput.
func ConvertValidPrepTaskConfigCreationRequestInputToValidPrepTaskConfigDatabaseCreationInput(input *mealplanning.ValidPrepTaskConfigCreationRequestInput) *mealplanning.ValidPrepTaskConfigDatabaseCreationInput {
	return &mealplanning.ValidPrepTaskConfigDatabaseCreationInput{
		ID:                             identifiers.New(),
		MinStorageDurationInSeconds:    input.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    input.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		StorageType:                    input.StorageType,
		StorageInstructions:            input.StorageInstructions,
		Notes:                          input.Notes,
		Source:                         input.Source,
		ValidPreparationID:             input.ValidPreparationID,
		ValidIngredientID:              input.ValidIngredientID,
	}
}

// ConvertValidPrepTaskConfigToValidPrepTaskConfigUpdateRequestInput builds a ValidPrepTaskConfigUpdateRequestInput from a ValidPrepTaskConfig.
func ConvertValidPrepTaskConfigToValidPrepTaskConfigUpdateRequestInput(validPrepTaskConfig *mealplanning.ValidPrepTaskConfig) *mealplanning.ValidPrepTaskConfigUpdateRequestInput {
	return &mealplanning.ValidPrepTaskConfigUpdateRequestInput{
		MinStorageDurationInSeconds:    &validPrepTaskConfig.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    validPrepTaskConfig.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: validPrepTaskConfig.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: validPrepTaskConfig.MaxStorageTemperatureInCelsius,
		StorageType:                    &validPrepTaskConfig.StorageType,
		StorageInstructions:            &validPrepTaskConfig.StorageInstructions,
		Notes:                          &validPrepTaskConfig.Notes,
		Source:                         &validPrepTaskConfig.Source,
		ValidPreparationID:             &validPrepTaskConfig.Preparation.ID,
		ValidIngredientID:              &validPrepTaskConfig.Ingredient.ID,
	}
}

// ConvertValidPrepTaskConfigToValidPrepTaskConfigCreationRequestInput builds a ValidPrepTaskConfigCreationRequestInput from a ValidPrepTaskConfig.
func ConvertValidPrepTaskConfigToValidPrepTaskConfigCreationRequestInput(validPrepTaskConfig *mealplanning.ValidPrepTaskConfig) *mealplanning.ValidPrepTaskConfigCreationRequestInput {
	return &mealplanning.ValidPrepTaskConfigCreationRequestInput{
		MinStorageDurationInSeconds:    validPrepTaskConfig.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    validPrepTaskConfig.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: validPrepTaskConfig.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: validPrepTaskConfig.MaxStorageTemperatureInCelsius,
		StorageType:                    validPrepTaskConfig.StorageType,
		StorageInstructions:            validPrepTaskConfig.StorageInstructions,
		Notes:                          validPrepTaskConfig.Notes,
		Source:                         validPrepTaskConfig.Source,
		ValidPreparationID:             validPrepTaskConfig.Preparation.ID,
		ValidIngredientID:              validPrepTaskConfig.Ingredient.ID,
	}
}

// ConvertValidPrepTaskConfigToValidPrepTaskConfigDatabaseCreationInput builds a ValidPrepTaskConfigDatabaseCreationInput from a ValidPrepTaskConfig.
func ConvertValidPrepTaskConfigToValidPrepTaskConfigDatabaseCreationInput(validPrepTaskConfig *mealplanning.ValidPrepTaskConfig) *mealplanning.ValidPrepTaskConfigDatabaseCreationInput {
	return &mealplanning.ValidPrepTaskConfigDatabaseCreationInput{
		ID:                             validPrepTaskConfig.ID,
		MinStorageDurationInSeconds:    validPrepTaskConfig.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    validPrepTaskConfig.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: validPrepTaskConfig.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: validPrepTaskConfig.MaxStorageTemperatureInCelsius,
		StorageType:                    validPrepTaskConfig.StorageType,
		StorageInstructions:            validPrepTaskConfig.StorageInstructions,
		Notes:                          validPrepTaskConfig.Notes,
		Source:                         validPrepTaskConfig.Source,
		ValidPreparationID:             validPrepTaskConfig.Preparation.ID,
		ValidIngredientID:              validPrepTaskConfig.Ingredient.ID,
	}
}
