package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidVessel_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidVessel{
			CapacityUnit: &ValidMeasurementUnit{},
		}
		input := &ValidVesselUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))
		input.CapacityUnitID = new(t.Name())
		input.UsableForStorage = new(true)
		input.DisplayInSummaryLists = new(true)
		input.IncludeInGeneratedInstructions = new(true)

		x.Update(input)
	})
}

func TestValidVesselCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidVesselCreationRequestInput{
			Name:           t.Name(),
			Description:    t.Name(),
			IconPath:       t.Name(),
			Capacity:       exampleQuantity,
			CapacityUnitID: new(t.Name()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidVesselCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidVesselDatabaseCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidVesselDatabaseCreationInput{
			ID:             t.Name(),
			Name:           t.Name(),
			CapacityUnitID: new(t.Name()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidVesselDatabaseCreationInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidVesselUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidVesselUpdateRequestInput{
			Name:        new(t.Name()),
			Description: new(t.Name()),
			IconPath:    new(t.Name()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidVesselUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}
