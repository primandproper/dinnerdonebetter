package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
)

// BuildFakeValidPreparationInstrument builds a faked valid preparation instrument.
func BuildFakeValidPreparationInstrument() *types.ValidPreparationInstrument {
	validPreparationInstrument := fake.BuildFakeRecord[types.ValidPreparationInstrument]()

	validPreparationInstrument.Preparation = *BuildFakeValidPreparation()
	validPreparationInstrument.Instrument = *BuildFakeValidInstrument()

	return validPreparationInstrument
}

// BuildFakeValidPreparationInstrumentsList builds a faked ValidPreparationInstrumentList.
func BuildFakeValidPreparationInstrumentsList() *filtering.QueryFilteredResult[types.ValidPreparationInstrument] {
	return fake.BuildFakePage(BuildFakeValidPreparationInstrument)
}

// BuildFakeValidPreparationInstrumentUpdateRequestInput builds a faked ValidPreparationInstrumentUpdateRequestInput from a valid preparation instrument.
func BuildFakeValidPreparationInstrumentUpdateRequestInput() *types.ValidPreparationInstrumentUpdateRequestInput {
	validPreparationInstrument := BuildFakeValidPreparationInstrument()

	return &types.ValidPreparationInstrumentUpdateRequestInput{
		Notes:              &validPreparationInstrument.Notes,
		ValidPreparationID: &validPreparationInstrument.Preparation.ID,
		ValidInstrumentID:  &validPreparationInstrument.Instrument.ID,
	}
}

// BuildFakeValidPreparationInstrumentCreationRequestInput builds a faked ValidPreparationInstrumentCreationRequestInput.
func BuildFakeValidPreparationInstrumentCreationRequestInput() *types.ValidPreparationInstrumentCreationRequestInput {
	validPreparationInstrument := BuildFakeValidPreparationInstrument()

	return converters.ConvertValidPreparationInstrumentToValidPreparationInstrumentCreationRequestInput(validPreparationInstrument)
}
