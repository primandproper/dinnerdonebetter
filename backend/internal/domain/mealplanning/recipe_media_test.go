package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipeMedia_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipeMedia{}
		input := &RecipeMediaUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))

		x.Update(input)
	})
}

func TestRecipeMediaCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &RecipeMediaCreationRequestInput{}
		fake.Struct(&x)

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestRecipeMediaDatabaseCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &RecipeMediaDatabaseCreationInput{}
		fake.Struct(&x)

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
