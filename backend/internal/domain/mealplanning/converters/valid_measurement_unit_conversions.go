package converters

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v6/identifiers"
)

// ConvertValidMeasurementUnitConversionToValidMeasurementUnitConversionUpdateRequestInput creates a ValidMeasurementUnitConversionUpdateRequestInput from a ValidMeasurementUnitConversion.
func ConvertValidMeasurementUnitConversionToValidMeasurementUnitConversionUpdateRequestInput(input *types.ValidMeasurementUnitConversion) *types.ValidMeasurementUnitConversionUpdateRequestInput {
	var onlyForIngredient *string
	if input.OnlyForIngredient != nil {
		onlyForIngredient = &input.OnlyForIngredient.ID
	}

	x := &types.ValidMeasurementUnitConversionUpdateRequestInput{
		From:              &input.From.ID,
		To:                &input.To.ID,
		Modifier:          &input.Modifier,
		Notes:             &input.Notes,
		OnlyForIngredient: onlyForIngredient,
	}

	return x
}

// ConvertValidMeasurementUnitConversionCreationRequestInputToValidMeasurementUnitConversionDatabaseCreationInput creates a ValidMeasurementUnitConversionDatabaseCreationInput from a ValidMeasurementUnitConversionCreationRequestInput.
func ConvertValidMeasurementUnitConversionCreationRequestInputToValidMeasurementUnitConversionDatabaseCreationInput(input *types.ValidMeasurementUnitConversionCreationRequestInput) *types.ValidMeasurementUnitConversionDatabaseCreationInput {
	x := &types.ValidMeasurementUnitConversionDatabaseCreationInput{
		ID:                identifiers.New(),
		From:              input.From,
		To:                input.To,
		OnlyForIngredient: input.OnlyForIngredient,
		Modifier:          input.Modifier,
		Notes:             input.Notes,
	}

	return x
}

// ConvertValidMeasurementUnitConversionToValidMeasurementUnitConversionCreationRequestInput builds a ValidMeasurementUnitConversionCreationRequestInput from a ValidMeasurementUnitConversion.
func ConvertValidMeasurementUnitConversionToValidMeasurementUnitConversionCreationRequestInput(validMeasurementUnitConversion *types.ValidMeasurementUnitConversion) *types.ValidMeasurementUnitConversionCreationRequestInput {
	var onlyForIngredient *string
	if validMeasurementUnitConversion.OnlyForIngredient != nil {
		onlyForIngredient = &validMeasurementUnitConversion.OnlyForIngredient.ID
	}

	x := &types.ValidMeasurementUnitConversionCreationRequestInput{
		From:              validMeasurementUnitConversion.From.ID,
		To:                validMeasurementUnitConversion.To.ID,
		Modifier:          validMeasurementUnitConversion.Modifier,
		Notes:             validMeasurementUnitConversion.Notes,
		OnlyForIngredient: onlyForIngredient,
	}

	return x
}

// ConvertValidMeasurementUnitConversionToValidMeasurementUnitConversionDatabaseCreationInput builds a ValidMeasurementUnitConversionDatabaseCreationInput from a ValidMeasurementUnitConversion.
func ConvertValidMeasurementUnitConversionToValidMeasurementUnitConversionDatabaseCreationInput(validMeasurementUnitConversion *types.ValidMeasurementUnitConversion) *types.ValidMeasurementUnitConversionDatabaseCreationInput {
	var onlyForIngredient *string
	if validMeasurementUnitConversion.OnlyForIngredient != nil {
		onlyForIngredient = &validMeasurementUnitConversion.OnlyForIngredient.ID
	}

	x := &types.ValidMeasurementUnitConversionDatabaseCreationInput{
		ID:                validMeasurementUnitConversion.ID,
		From:              validMeasurementUnitConversion.From.ID,
		To:                validMeasurementUnitConversion.To.ID,
		Modifier:          validMeasurementUnitConversion.Modifier,
		Notes:             validMeasurementUnitConversion.Notes,
		OnlyForIngredient: onlyForIngredient,
	}

	return x
}
