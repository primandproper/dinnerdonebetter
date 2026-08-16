package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
)

// BuildFakeValidMeasurementUnitConversionUnitUpdateRequestInput builds a faked ValidMeasurementUnitConversionUpdateRequestInput from a valid preparation.
func BuildFakeValidMeasurementUnitConversionUnitUpdateRequestInput() *types.ValidMeasurementUnitConversionUpdateRequestInput {
	validMeasurementUnitConversion := BuildFakeValidMeasurementUnitConversion()

	x := &types.ValidMeasurementUnitConversionUpdateRequestInput{
		From:     &validMeasurementUnitConversion.From.ID,
		To:       &validMeasurementUnitConversion.To.ID,
		Modifier: &validMeasurementUnitConversion.Modifier,
		Notes:    &validMeasurementUnitConversion.Notes,
	}

	if validMeasurementUnitConversion.OnlyForIngredient != nil {
		x.OnlyForIngredient = &validMeasurementUnitConversion.OnlyForIngredient.ID
	}

	return x
}
