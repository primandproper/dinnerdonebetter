package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/pointer"
)

// BuildFakeValidMeasurementUnitConversion builds a faked valid measurement unit conversion.
func BuildFakeValidMeasurementUnitConversion() *types.ValidMeasurementUnitConversion {
	conversion := fake.BuildFakeRecord[types.ValidMeasurementUnitConversion]()

	// The two units the conversion is between. They are whole records rather than IDs,
	// and a conversion whose ends are random is one no unit in the test converts through.
	conversion.From = *BuildFakeValidMeasurementUnit()
	conversion.To = *BuildFakeValidMeasurementUnit()

	return conversion
}

// BuildFakeValidMeasurementUnitConversionsList builds a faked ValidMeasurementUnitConversionList.
func BuildFakeValidMeasurementUnitConversionsList() *filtering.QueryFilteredResult[types.ValidMeasurementUnitConversion] {
	return fake.BuildFakePage(BuildFakeValidMeasurementUnitConversion)
}

// BuildFakeValidMeasurementUnitConversionUpdateRequestInput builds a faked
// ValidMeasurementUnitConversionUpdateRequestInput whose every field is set.
//
// Each field on this input is optional, and an update whose fields are all absent
// updates nothing — so they are filled here rather than left to BuildFakeRecord.
func BuildFakeValidMeasurementUnitConversionUpdateRequestInput() *types.ValidMeasurementUnitConversionUpdateRequestInput {
	return &types.ValidMeasurementUnitConversionUpdateRequestInput{
		From:              pointer.To(fake.BuildFakeID()),
		To:                pointer.To(fake.BuildFakeID()),
		OnlyForIngredient: pointer.To(fake.BuildFakeID()),
		Modifier:          pointer.To(float32(fake.BuildFakeNumber())),
		Notes:             pointer.To(fake.BuildFakeID()),
	}
}

// BuildFakeValidMeasurementUnitConversionUnitUpdateRequestInput builds a faked ValidMeasurementUnitConversionUpdateRequestInput from a conversion.
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

// BuildFakeValidMeasurementUnitConversionCreationRequestInput builds a faked ValidMeasurementUnitConversionCreationRequestInput.
func BuildFakeValidMeasurementUnitConversionCreationRequestInput() *types.ValidMeasurementUnitConversionCreationRequestInput {
	validMeasurementUnitConversion := BuildFakeValidMeasurementUnitConversion()

	return converters.ConvertValidMeasurementUnitConversionToValidMeasurementUnitConversionCreationRequestInput(validMeasurementUnitConversion)
}
