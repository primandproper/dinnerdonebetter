package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidPreparationCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidPreparationCreationRequestInput{
			Name:                  t.Name(),
			Description:           t.Name(),
			IconPath:              t.Name(),
			PastTense:             t.Name(),
			YieldsNothing:         fake.Bool(),
			RestrictToIngredients: fake.Bool(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidPreparationCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidPreparationUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidPreparationUpdateRequestInput{
			Name:                  new(t.Name()),
			Description:           new(t.Name()),
			IconPath:              new(t.Name()),
			PastTense:             new(t.Name()),
			YieldsNothing:         new(fake.Bool()),
			RestrictToIngredients: new(fake.Bool()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidPreparationUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidPreparation_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidPreparation{
			MaxInstrumentCount: new(uint16(0)),
			MaxIngredientCount: new(uint16(0)),
			MaxVesselCount:     new(uint16(0)),
		}
		input := &ValidPreparationUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))
		input.YieldsNothing = new(true)
		input.RestrictToIngredients = new(true)
		input.MaxIngredientCount = new(uint16(1))
		input.MaxInstrumentCount = new(uint16(1))
		input.MaxVesselCount = new(uint16(1))
		input.TemperatureRequired = new(true)
		input.TimeEstimateRequired = new(true)
		input.OnlyForVessels = new(true)
		input.ConsumesVessel = new(true)
		input.ConditionExpressionRequired = new(true)

		x.Update(input)
	})
}

func TestValidPreparationCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ValidPreparationCreationRequestInput{
			Name: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestValidPreparationDatabaseCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ValidPreparationDatabaseCreationInput{
			ID:   t.Name(),
			Name: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestValidPreparationUpdateRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ValidPreparationUpdateRequestInput{
			Name: new(t.Name()),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
