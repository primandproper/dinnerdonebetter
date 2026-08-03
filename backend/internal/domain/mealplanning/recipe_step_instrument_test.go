package mealplanning

import (
	"math"
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipeStepInstrument_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepInstrument{
			RecipeStepProductID: new(t.Name()),
			MaxQuantity:         new(uint32(321)),
		}
		input := &RecipeStepInstrumentUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))
		input.Optional = new(true)
		input.RecipeStepProductID = new("whatever")
		input.MaxQuantity = new(uint32(123))

		x.Update(input)
	})

	T.Run("with scale_factor", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepInstrument{ScaleFactor: 1.0}
		input := &RecipeStepInstrumentUpdateRequestInput{
			ScaleFactor: new(float32(0.5)),
		}
		x.Update(input)
		assert.Equal(t, float32(0.5), x.ScaleFactor)
	})
}

func TestRecipeStepInstrumentCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepInstrumentCreationRequestInput{
			ValidPreparationInstrumentID: new(t.Name()),
			Name:                         t.Name(),
			RecipeStepProductID:          new(t.Name()),
			Notes:                        t.Name(),
			PreferenceRank:               uint8(fake.Number(1, math.MaxUint8)),
			Optional:                     fake.Bool(),
			MinQuantity:                  fake.Uint32(),

			MaxQuantity: new(fake.Uint32()),
			ScaleFactor: 0.5,
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("recipe step product does not require bridge IDs", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepInstrumentCreationRequestInput{
			ProductOfRecipeStepIndex:        new(uint64(0)),
			ProductOfRecipeStepProductIndex: new(uint64(0)),
			Name:                            t.Name(),
			Notes:                           t.Name(),
			PreferenceRank:                  uint8(fake.Number(1, math.MaxUint8)),
			Optional:                        fake.Bool(),
			MinQuantity:                     fake.Uint32(),

			MaxQuantity: new(fake.Uint32()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepInstrumentCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestRecipeStepInstrumentDatabaseCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepInstrumentDatabaseCreationInput{
			ID:                  t.Name(),
			InstrumentID:        new(t.Name()),
			BelongsToRecipeStep: t.Name(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepInstrumentDatabaseCreationInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestRecipeStepInstrumentUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepInstrumentUpdateRequestInput{
			InstrumentID:        new(t.Name()),
			Name:                new(t.Name()),
			BelongsToRecipeStep: new(t.Name()),
			RecipeStepProductID: new(t.Name()),
			Notes:               new(t.Name()),
			PreferenceRank:      new(uint8(fake.Number(1, math.MaxUint8))),
			Optional:            new(fake.Bool()),
			MinQuantity:         new(fake.Uint32()),

			MaxQuantity: new(fake.Uint32()),
			ScaleFactor: new(float32(0.5)),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepInstrumentUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}
