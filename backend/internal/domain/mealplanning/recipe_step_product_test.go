package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipeStepProduct_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepProduct{
			MeasurementUnit:        &ValidMeasurementUnit{},
			ContainedInVesselIndex: new(uint16(3)),
		}
		input := &RecipeStepProductUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))
		input.Compostable = new(true)
		input.ContainedInVesselIndex = new(uint16(1))
		input.IsLiquid = new(true)
		input.IsWaste = new(true)

		x.Update(input)
	})
}

func TestRecipeStepProductCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepProductCreationRequestInput{
			Name:                           t.Name(),
			Type:                           RecipeStepProductIngredientType,
			MeasurementUnitID:              new(t.Name()),
			MinMeasurementQuantity:         new(fake.Float32()),
			QuantityNotes:                  t.Name(),
			Compostable:                    fake.Bool(),
			MaxStorageDurationInSeconds:    new(fake.Uint32()),
			MinStorageTemperatureInCelsius: new(fake.Float32()),
			MaxStorageTemperatureInCelsius: new(fake.Float32()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepProductCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestRecipeStepProductUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepProductUpdateRequestInput{
			Name:                           new(t.Name()),
			Type:                           new(RecipeStepProductIngredientType),
			MeasurementUnitID:              new(t.Name()),
			MinMeasurementQuantity:         new(fake.Float32()),
			MaxMeasurementQuantity:         new(fake.Float32()),
			QuantityNotes:                  new(t.Name()),
			Compostable:                    new(fake.Bool()),
			MinStorageTemperatureInCelsius: new(fake.Float32()),
			MaxStorageTemperatureInCelsius: new(fake.Float32()),
			MinStorageDurationInSeconds:    new(fake.Uint32()),
			MaxStorageDurationInSeconds:    new(fake.Uint32()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepProductUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}
