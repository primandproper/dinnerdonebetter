package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v9/filtering"
)

// BuildFakeValidPreparationVessel builds a faked valid preparation instrument.
func BuildFakeValidPreparationVessel() *types.ValidPreparationVessel {
	return &types.ValidPreparationVessel{
		ID:          BuildFakeID(),
		Notes:       buildUniqueString(),
		Preparation: *BuildFakeValidPreparation(),
		Vessel:      *BuildFakeValidVessel(),
		CreatedAt:   BuildFakeTime(),
	}
}

// BuildFakeValidPreparationVesselsList builds a faked ValidPreparationVesselList.
func BuildFakeValidPreparationVesselsList() *filtering.QueryFilteredResult[types.ValidPreparationVessel] {
	var examples []*types.ValidPreparationVessel
	for range exampleQuantity {
		examples = append(examples, BuildFakeValidPreparationVessel())
	}

	return &filtering.QueryFilteredResult[types.ValidPreparationVessel]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeValidPreparationVesselUpdateRequestInput builds a faked ValidPreparationVesselUpdateRequestInput from a valid preparation instrument.
func BuildFakeValidPreparationVesselUpdateRequestInput() *types.ValidPreparationVesselUpdateRequestInput {
	validPreparationVessel := BuildFakeValidPreparationVessel()
	return &types.ValidPreparationVesselUpdateRequestInput{
		Notes:              &validPreparationVessel.Notes,
		ValidPreparationID: &validPreparationVessel.Preparation.ID,
		ValidVesselID:      &validPreparationVessel.Vessel.ID,
	}
}

// BuildFakeValidPreparationVesselCreationRequestInput builds a faked ValidPreparationVesselCreationRequestInput.
func BuildFakeValidPreparationVesselCreationRequestInput() *types.ValidPreparationVesselCreationRequestInput {
	validPreparationVessel := BuildFakeValidPreparationVessel()
	return converters.ConvertValidPreparationVesselToValidPreparationVesselCreationRequestInput(validPreparationVessel)
}
