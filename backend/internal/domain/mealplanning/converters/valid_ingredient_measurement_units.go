package converters

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v4/identifiers"
)

// ConvertValidIngredientMeasurementUnitToValidIngredientMeasurementUnitUpdateRequestInput creates a ValidIngredientMeasurementUnitUpdateRequestInput from a ValidIngredientMeasurementUnit.
func ConvertValidIngredientMeasurementUnitToValidIngredientMeasurementUnitUpdateRequestInput(input *mealplanning.ValidIngredientMeasurementUnit) *mealplanning.ValidIngredientMeasurementUnitUpdateRequestInput {
	x := &mealplanning.ValidIngredientMeasurementUnitUpdateRequestInput{
		Notes:                  &input.Notes,
		ValidMeasurementUnitID: &input.MeasurementUnit.ID,
		ValidIngredientID:      &input.Ingredient.ID,
		MinAllowableQuantity:   &input.MinAllowableQuantity,
		MaxAllowableQuantity:   input.MaxAllowableQuantity,
	}

	return x
}

// ConvertValidIngredientMeasurementUnitCreationRequestInputToValidIngredientMeasurementUnitDatabaseCreationInput creates a ValidIngredientMeasurementUnitDatabaseCreationInput from a ValidIngredientMeasurementUnitCreationRequestInput.
func ConvertValidIngredientMeasurementUnitCreationRequestInputToValidIngredientMeasurementUnitDatabaseCreationInput(input *mealplanning.ValidIngredientMeasurementUnitCreationRequestInput) *mealplanning.ValidIngredientMeasurementUnitDatabaseCreationInput {
	x := &mealplanning.ValidIngredientMeasurementUnitDatabaseCreationInput{
		ID:                     identifiers.New(),
		Notes:                  input.Notes,
		ValidMeasurementUnitID: input.ValidMeasurementUnitID,
		ValidIngredientID:      input.ValidIngredientID,
		MinAllowableQuantity:   input.MinAllowableQuantity,
		MaxAllowableQuantity:   input.MaxAllowableQuantity,
	}

	return x
}

// ConvertValidIngredientMeasurementUnitToValidIngredientMeasurementUnitCreationRequestInput builds a ValidIngredientMeasurementUnitCreationRequestInput from a ValidIngredientMeasurementUnit.
func ConvertValidIngredientMeasurementUnitToValidIngredientMeasurementUnitCreationRequestInput(validIngredientMeasurementUnit *mealplanning.ValidIngredientMeasurementUnit) *mealplanning.ValidIngredientMeasurementUnitCreationRequestInput {
	return &mealplanning.ValidIngredientMeasurementUnitCreationRequestInput{
		Notes:                  validIngredientMeasurementUnit.Notes,
		ValidMeasurementUnitID: validIngredientMeasurementUnit.MeasurementUnit.ID,
		ValidIngredientID:      validIngredientMeasurementUnit.Ingredient.ID,
		MinAllowableQuantity:   validIngredientMeasurementUnit.MinAllowableQuantity,
		MaxAllowableQuantity:   validIngredientMeasurementUnit.MaxAllowableQuantity,
	}
}

// ConvertValidIngredientMeasurementUnitToValidIngredientMeasurementUnitDatabaseCreationInput builds a ValidIngredientMeasurementUnitDatabaseCreationInput from a ValidIngredientMeasurementUnit.
func ConvertValidIngredientMeasurementUnitToValidIngredientMeasurementUnitDatabaseCreationInput(validIngredientMeasurementUnit *mealplanning.ValidIngredientMeasurementUnit) *mealplanning.ValidIngredientMeasurementUnitDatabaseCreationInput {
	return &mealplanning.ValidIngredientMeasurementUnitDatabaseCreationInput{
		ID:                     validIngredientMeasurementUnit.ID,
		Notes:                  validIngredientMeasurementUnit.Notes,
		ValidMeasurementUnitID: validIngredientMeasurementUnit.MeasurementUnit.ID,
		ValidIngredientID:      validIngredientMeasurementUnit.Ingredient.ID,
		MinAllowableQuantity:   validIngredientMeasurementUnit.MinAllowableQuantity,
		MaxAllowableQuantity:   validIngredientMeasurementUnit.MaxAllowableQuantity,
	}
}
