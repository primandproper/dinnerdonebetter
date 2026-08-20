package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"
)

// BuildFakeValidMeasurementUnit builds a faked valid measurement unit.
func BuildFakeValidMeasurementUnit() *types.ValidMeasurementUnit {
	validMeasurementUnit := fake.BuildFakeRecord[types.ValidMeasurementUnit]()

	// A unit belongs to one system or the other, and the conversion path reads both
	// flags — so a unit that is neither, or both, is one no conversion knows what to do
	// with.
	validMeasurementUnit.Metric = true
	validMeasurementUnit.Imperial = false

	return validMeasurementUnit
}

// BuildFakeValidMeasurementUnitsList builds a faked ValidMeasurementUnitList.
func BuildFakeValidMeasurementUnitsList() *filtering.QueryFilteredResult[types.ValidMeasurementUnit] {
	return fake.BuildFakePage(BuildFakeValidMeasurementUnit)
}

// BuildFakeValidMeasurementUnitUpdateRequestInput builds a faked ValidMeasurementUnitUpdateRequestInput from a valid measurement unit.
func BuildFakeValidMeasurementUnitUpdateRequestInput() *types.ValidMeasurementUnitUpdateRequestInput {
	validMeasurementUnit := BuildFakeValidMeasurementUnit()

	return converters.ConvertValidMeasurementUnitToValidMeasurementUnitUpdateRequestInput(validMeasurementUnit)
}

// BuildFakeValidMeasurementUnitCreationRequestInput builds a faked ValidMeasurementUnitCreationRequestInput.
func BuildFakeValidMeasurementUnitCreationRequestInput() *types.ValidMeasurementUnitCreationRequestInput {
	validMeasurementUnit := BuildFakeValidMeasurementUnit()

	return converters.ConvertValidMeasurementUnitToValidMeasurementUnitCreationRequestInput(validMeasurementUnit)
}
