package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v6/filtering"
)

// BuildFakeValidPreparationInstrument builds a faked valid preparation instrument.
func BuildFakeValidPreparationInstrument() *types.ValidPreparationInstrument {
	return &types.ValidPreparationInstrument{
		ID:          BuildFakeID(),
		Notes:       buildUniqueString(),
		Preparation: *BuildFakeValidPreparation(),
		Instrument:  *BuildFakeValidInstrument(),
		CreatedAt:   BuildFakeTime(),
	}
}

// BuildFakeValidPreparationInstrumentsList builds a faked ValidPreparationInstrumentList.
func BuildFakeValidPreparationInstrumentsList() *filtering.QueryFilteredResult[types.ValidPreparationInstrument] {
	var examples []*types.ValidPreparationInstrument
	for range exampleQuantity {
		examples = append(examples, BuildFakeValidPreparationInstrument())
	}

	return &filtering.QueryFilteredResult[types.ValidPreparationInstrument]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
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
