package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipeRating_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipeRating{}
		input := &RecipeRatingUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))

		x.Update(input)
	})
}

func TestRecipeRatingCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &RecipeRatingCreationRequestInput{
			BelongsToRecipe: t.Name(),
			Difficulty:      1.0,
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &RecipeRatingCreationRequestInput{}

		assert.Error(t, x.ValidateWithContext(ctx))
	})
}

func TestRecipeRatingDatabaseCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &RecipeRatingDatabaseCreationInput{
			ID:              t.Name(),
			BelongsToRecipe: t.Name(),
			CreatedByUser:   t.Name(),
			Difficulty:      1.0,
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &RecipeRatingDatabaseCreationInput{}

		assert.Error(t, x.ValidateWithContext(ctx))
	})
}

func TestRecipeRatingUpdateRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &RecipeRatingUpdateRequestInput{
			BelongsToRecipe: new(t.Name()),
			Difficulty:      new(float32(1.0)),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &RecipeRatingUpdateRequestInput{}

		assert.Error(t, x.ValidateWithContext(ctx))
	})
}
