package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
)

// BuildFakeValidPreparationVessel builds a faked valid preparation vessel.
func BuildFakeValidPreparationVessel() *types.ValidPreparationVessel {
	validPreparationVessel := fake.BuildFakeRecord[types.ValidPreparationVessel]()

	validPreparationVessel.Preparation = *BuildFakeValidPreparation()
	validPreparationVessel.Vessel = *BuildFakeValidVessel()

	return validPreparationVessel
}

// BuildFakeValidPreparationVesselsList builds a faked ValidPreparationVesselList.
func BuildFakeValidPreparationVesselsList() *filtering.QueryFilteredResult[types.ValidPreparationVessel] {
	return fake.BuildFakePage(BuildFakeValidPreparationVessel)
}

// BuildFakeValidPreparationVesselUpdateRequestInput builds a faked ValidPreparationVesselUpdateRequestInput from a valid preparation vessel.
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
