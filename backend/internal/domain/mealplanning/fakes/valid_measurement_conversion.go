package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v5/filtering"
)

// BuildFakeValidMeasurementUnitConversion builds a faked valid preparation.
func BuildFakeValidMeasurementUnitConversion() *types.ValidMeasurementUnitConversion {
	return &types.ValidMeasurementUnitConversion{
		ID:                BuildFakeID(),
		From:              *BuildFakeValidMeasurementUnit(),
		To:                *BuildFakeValidMeasurementUnit(),
		OnlyForIngredient: nil,
		Modifier:          float32(buildFakeNumber()),
		Notes:             buildUniqueString(),
		CreatedAt:         BuildFakeTime(),
	}
}

// BuildFakeValidMeasurementUnitConversionsList builds a faked ValidMeasurementUnitConversionList.
func BuildFakeValidMeasurementUnitConversionsList() *filtering.QueryFilteredResult[types.ValidMeasurementUnitConversion] {
	var examples []*types.ValidMeasurementUnitConversion
	for range exampleQuantity {
		examples = append(examples, BuildFakeValidMeasurementUnitConversion())
	}

	return &filtering.QueryFilteredResult[types.ValidMeasurementUnitConversion]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

func BuildFakeValidMeasurementUnitConversionUpdateRequestInput() *types.ValidMeasurementUnitConversionUpdateRequestInput {
	return &types.ValidMeasurementUnitConversionUpdateRequestInput{
		From:              new(BuildFakeID()),
		To:                new(BuildFakeID()),
		OnlyForIngredient: new(BuildFakeID()),
		Modifier:          new(float32(buildFakeNumber())),
		Notes:             new(BuildFakeID()),
	}
}

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

// BuildFakeValidMeasurementUnitConversionCreationRequestInput builds a faked ValidMeasurementUnitConversionCreationRequestInput.
func BuildFakeValidMeasurementUnitConversionCreationRequestInput() *types.ValidMeasurementUnitConversionCreationRequestInput {
	validMeasurementUnitConversion := BuildFakeValidMeasurementUnitConversion()
	return converters.ConvertValidMeasurementUnitConversionToValidMeasurementUnitConversionCreationRequestInput(validMeasurementUnitConversion)
}
