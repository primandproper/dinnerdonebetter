package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidIngredientMeasurementUnit_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientMeasurementUnit{
			MaxAllowableQuantity: new(float32(3.21)),
		}
		input := &ValidIngredientMeasurementUnitUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))
		input.MaxAllowableQuantity = new(float32(1.23))

		x.Update(input)
	})
}

func TestValidIngredientMeasurementUnitCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientMeasurementUnitCreationRequestInput{
			Notes:                  t.Name(),
			ValidMeasurementUnitID: t.Name(),
			ValidIngredientID:      t.Name(),
			MinAllowableQuantity:   fake.Float32(),

			MaxAllowableQuantity: new(fake.Float32()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientMeasurementUnitCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidIngredientMeasurementUnitDatabaseCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientMeasurementUnitDatabaseCreationInput{
			ID:                     t.Name(),
			ValidMeasurementUnitID: t.Name(),
			ValidIngredientID:      t.Name(),
			MinAllowableQuantity:   fake.Float32(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientMeasurementUnitDatabaseCreationInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidIngredientMeasurementUnitUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientMeasurementUnitUpdateRequestInput{
			Notes:                  new(t.Name()),
			ValidMeasurementUnitID: new(t.Name()),
			ValidIngredientID:      new(t.Name()),
			MinAllowableQuantity:   new(fake.Float32()),

			MaxAllowableQuantity: new(fake.Float32()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientMeasurementUnitUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}
