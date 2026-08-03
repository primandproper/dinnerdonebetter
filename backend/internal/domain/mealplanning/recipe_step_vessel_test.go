package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipeStepVessel_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepVessel{
			MinQuantity: 1234,

			MaxQuantity: new(uint16(1234)),
			Vessel:      &ValidVessel{},
		}
		input := &RecipeStepVesselUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))
		input.UnavailableAfterStep = new(true)
		input.MinQuantity = new(uint16(1))
		input.MaxQuantity = new(uint16(1))
		input.VesselID = new(t.Name())

		x.Update(input)
	})

	T.Run("with scale_factor", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepVessel{ScaleFactor: 1.0}
		input := &RecipeStepVesselUpdateRequestInput{
			ScaleFactor: new(float32(0.5)),
		}
		x.Update(input)
		assert.Equal(t, float32(0.5), x.ScaleFactor) //nolint:testifylint // verbatim assignment; exact equality is the assertion
	})
}

func TestRecipeStepVesselCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepVesselCreationRequestInput{
			ValidPreparationVesselID: new(t.Name()),
			Name:                     t.Name(),
			RecipeStepProductID:      new(t.Name()),
			Notes:                    t.Name(),
			MinQuantity:              fake.Uint16(),

			MaxQuantity: new(fake.Uint16()),
			ScaleFactor: 0.5,
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("recipe step product does not require bridge IDs", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepVesselCreationRequestInput{
			ProductOfRecipeStepIndex:        new(uint64(0)),
			ProductOfRecipeStepProductIndex: new(uint64(0)),
			Name:                            t.Name(),
			Notes:                           t.Name(),
			MinQuantity:                     fake.Uint16(),

			MaxQuantity: new(fake.Uint16()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepVesselCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestRecipeStepVesselDatabaseCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepVesselDatabaseCreationInput{
			ID:                  t.Name(),
			BelongsToRecipeStep: t.Name(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepVesselDatabaseCreationInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestRecipeStepVesselUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepVesselUpdateRequestInput{
			Name:                new(t.Name()),
			BelongsToRecipeStep: new(t.Name()),
			RecipeStepProductID: new(t.Name()),
			Notes:               new(t.Name()),
			MinQuantity:         new(fake.Uint16()),

			MaxQuantity: new(fake.Uint16()),
			ScaleFactor: new(float32(0.5)),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &RecipeStepVesselUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}
