package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
)

// BuildFakeValidInstrument builds a faked valid instrument.
func BuildFakeValidInstrument() *types.ValidInstrument {
	return fake.BuildFakeRecord[types.ValidInstrument]()
}

// BuildFakeValidInstrumentsList builds a faked ValidInstrumentList.
func BuildFakeValidInstrumentsList() *filtering.QueryFilteredResult[types.ValidInstrument] {
	return fake.BuildFakePage(BuildFakeValidInstrument)
}

// BuildFakeValidInstrumentUpdateRequestInput builds a faked ValidInstrumentUpdateRequestInput from a valid instrument.
func BuildFakeValidInstrumentUpdateRequestInput() *types.ValidInstrumentUpdateRequestInput {
	validInstrument := BuildFakeValidInstrument()

	return converters.ConvertValidInstrumentToValidInstrumentUpdateRequestInput(validInstrument)
}

// BuildFakeValidInstrumentCreationRequestInput builds a faked ValidInstrumentCreationRequestInput.
func BuildFakeValidInstrumentCreationRequestInput() *types.ValidInstrumentCreationRequestInput {
	validInstrument := BuildFakeValidInstrument()

	return converters.ConvertValidInstrumentToValidInstrumentCreationRequestInput(validInstrument)
}
