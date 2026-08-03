package mealplanning

import (
	"math"
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserIngredientPreference_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &UserIngredientPreference{}
		input := &UserIngredientPreferenceUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))

		x.Update(input)
	})
}

func TestUserIngredientPreferenceCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		input := &UserIngredientPreferenceCreationRequestInput{
			ValidIngredientID: t.Name(),
			Rating:            1,
		}

		assert.NoError(t, input.ValidateWithContext(ctx))
	})

	T.Run("invalid range", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		input := &UserIngredientPreferenceCreationRequestInput{
			ValidIngredientID: t.Name(),
			Rating:            math.MaxInt8,
		}

		assert.Error(t, input.ValidateWithContext(ctx))
	})

	T.Run("with valid ingredient group ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		input := &UserIngredientPreferenceCreationRequestInput{
			ValidIngredientGroupID: t.Name(),
			Rating:                 1,
		}

		assert.NoError(t, input.ValidateWithContext(ctx))
	})
}

func TestUserIngredientPreferenceDatabaseCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &UserIngredientPreferenceDatabaseCreationInput{
			ValidIngredientID: t.Name(),
			Rating:            minRating,
			CreatedByUser:     t.Name(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &UserIngredientPreferenceDatabaseCreationInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestUserIngredientPreferenceUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &UserIngredientPreferenceUpdateRequestInput{
			IngredientID: new(t.Name()),
			Rating:       new(minRating),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &UserIngredientPreferenceUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}
