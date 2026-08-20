package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"
)

// BuildFakeValidVessel builds a faked valid vessel.
func BuildFakeValidVessel() *types.ValidVessel {
	validVessel := fake.BuildFakeRecord[types.ValidVessel]()

	// A shape from the enumeration the domain validates against.
	validVessel.Shape = types.VesselShapeOther

	// The unit the capacity is in. A capacity without one is a number without a
	// dimension, which the scaling code divides by.
	validVessel.CapacityUnit = BuildFakeValidMeasurementUnit()

	return validVessel
}

// BuildFakeValidVesselsList builds a faked ValidVesselList.
func BuildFakeValidVesselsList() *filtering.QueryFilteredResult[types.ValidVessel] {
	return fake.BuildFakePage(BuildFakeValidVessel)
}

// BuildFakeValidVesselUpdateRequestInput builds a faked ValidVesselUpdateRequestInput from a valid vessel.
func BuildFakeValidVesselUpdateRequestInput() *types.ValidVesselUpdateRequestInput {
	validVessel := BuildFakeValidVessel()

	return converters.ConvertValidVesselToValidVesselUpdateRequestInput(validVessel)
}

// BuildFakeValidVesselCreationRequestInput builds a faked ValidVesselCreationRequestInput.
func BuildFakeValidVesselCreationRequestInput() *types.ValidVesselCreationRequestInput {
	validVessel := BuildFakeValidVessel()

	return converters.ConvertValidVesselToValidVesselCreationRequestInput(validVessel)
}
