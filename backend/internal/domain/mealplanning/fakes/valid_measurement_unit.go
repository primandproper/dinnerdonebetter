package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v9/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeValidMeasurementUnit builds a faked valid ingredient.
func BuildFakeValidMeasurementUnit() *types.ValidMeasurementUnit {
	return &types.ValidMeasurementUnit{
		ID:          BuildFakeID(),
		Name:        buildUniqueString(),
		Description: buildUniqueString(),
		Volumetric:  fake.Bool(),
		IconPath:    buildUniqueString(),
		Universal:   fake.Bool(),
		Metric:      true,
		Imperial:    false,
		PluralName:  buildUniqueString(),
		Slug:        buildUniqueString(),
		CreatedAt:   BuildFakeTime(),
	}
}

// BuildFakeValidMeasurementUnitsList builds a faked ValidMeasurementUnitList.
func BuildFakeValidMeasurementUnitsList() *filtering.QueryFilteredResult[types.ValidMeasurementUnit] {
	var examples []*types.ValidMeasurementUnit
	for range exampleQuantity {
		examples = append(examples, BuildFakeValidMeasurementUnit())
	}

	return &filtering.QueryFilteredResult[types.ValidMeasurementUnit]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeValidMeasurementUnitUpdateRequestInput builds a faked ValidMeasurementUnitUpdateRequestInput from a valid ingredient.
func BuildFakeValidMeasurementUnitUpdateRequestInput() *types.ValidMeasurementUnitUpdateRequestInput {
	validMeasurementUnit := BuildFakeValidMeasurementUnit()
	return converters.ConvertValidMeasurementUnitToValidMeasurementUnitUpdateRequestInput(validMeasurementUnit)
}

// BuildFakeValidMeasurementUnitCreationRequestInput builds a faked ValidMeasurementUnitCreationRequestInput.
func BuildFakeValidMeasurementUnitCreationRequestInput() *types.ValidMeasurementUnitCreationRequestInput {
	validMeasurementUnit := BuildFakeValidMeasurementUnit()
	return converters.ConvertValidMeasurementUnitToValidMeasurementUnitCreationRequestInput(validMeasurementUnit)
}
