package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidIngredientStateIngredient_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientStateIngredient{}
		input := &ValidIngredientStateIngredientUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))

		x.Update(input)
	})
}

func TestValidIngredientStateIngredientCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientStateIngredientCreationRequestInput{
			Notes:                  t.Name(),
			ValidIngredientStateID: t.Name(),
			ValidIngredientID:      t.Name(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientStateIngredientCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidIngredientStateIngredientDatabaseCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientStateIngredientDatabaseCreationInput{
			ID:                     t.Name(),
			ValidIngredientStateID: t.Name(),
			ValidIngredientID:      t.Name(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientStateIngredientDatabaseCreationInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidIngredientStateIngredientUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientStateIngredientUpdateRequestInput{
			Notes:                  new(t.Name()),
			ValidIngredientStateID: new(t.Name()),
			ValidIngredientID:      new(t.Name()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientStateIngredientUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}
