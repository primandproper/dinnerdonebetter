package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"
)

// BuildFakeValidPreparation builds a faked valid preparation.
func BuildFakeValidPreparation() *types.ValidPreparation {
	validPreparation := fake.BuildFakeRecord[types.ValidPreparation]()

	validPreparation.MinIngredientCount, validPreparation.MaxIngredientCount = BuildFakeUint16WithOptionalMax()
	validPreparation.MinInstrumentCount, validPreparation.MaxInstrumentCount = BuildFakeUint16WithOptionalMax()
	validPreparation.MinVesselCount, validPreparation.MaxVesselCount = BuildFakeUint16WithOptionalMax()

	return validPreparation
}

// BuildFakeValidPreparationsList builds a faked ValidPreparationList.
func BuildFakeValidPreparationsList() *filtering.QueryFilteredResult[types.ValidPreparation] {
	return fake.BuildFakePage(BuildFakeValidPreparation)
}

// BuildFakeValidPreparationUpdateRequestInput builds a faked ValidPreparationUpdateRequestInput from a valid preparation.
func BuildFakeValidPreparationUpdateRequestInput() *types.ValidPreparationUpdateRequestInput {
	validPreparation := BuildFakeValidPreparation()

	return converters.ConvertValidPreparationToValidPreparationUpdateRequestInput(validPreparation)
}

// BuildFakeValidPreparationCreationRequestInput builds a faked ValidPreparationCreationRequestInput.
func BuildFakeValidPreparationCreationRequestInput() *types.ValidPreparationCreationRequestInput {
	validPreparation := BuildFakeValidPreparation()

	return converters.ConvertValidPreparationToValidPreparationCreationRequestInput(validPreparation)
}
