package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidMeasurementUnit_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnit{
			Imperial: true,
		}
		input := &ValidMeasurementUnitUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))
		input.Volumetric = new(true)
		input.Universal = new(true)
		input.Imperial = new(false)
		input.Metric = new(true)

		x.Update(input)
	})
}

func TestValidMeasurementUnitCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitCreationRequestInput{
			Name:   t.Name(),
			Metric: true,
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with both metric and imperial", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitCreationRequestInput{
			Name:     t.Name(),
			Metric:   true,
			Imperial: true,
		}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidMeasurementUnitDatabaseCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitDatabaseCreationInput{
			ID:       t.Name(),
			Name:     t.Name(),
			Imperial: true,
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with both metric and imperial", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitDatabaseCreationInput{
			ID:       t.Name(),
			Name:     t.Name(),
			Imperial: true,
			Metric:   true,
		}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitDatabaseCreationInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidMeasurementUnitUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitUpdateRequestInput{
			Name: new(t.Name()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidMeasurementUnitUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}
