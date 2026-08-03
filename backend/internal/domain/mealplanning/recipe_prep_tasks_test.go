package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipePrepTask_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipePrepTask{}
		input := &RecipePrepTaskUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))
		input.Optional = new(true)
		input.BelongsToRecipe = new(t.Name())

		x.Update(input)
	})
}

func TestRecipePrepTaskCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipePrepTaskCreationRequestInput{
			BelongsToRecipe:                    t.Name(),
			Name:                               t.Name(),
			StorageType:                        t.Name(),
			MinStorageTemperatureInCelsius:     new(fake.Float32()),
			MaxStorageTemperatureInCelsius:     new(fake.Float32()),
			MinTimeBufferBeforeRecipeInSeconds: fake.Uint32(),
			MaxTimeBufferBeforeRecipeInSeconds: new(fake.Uint32()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &RecipePrepTaskCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestRecipePrepTaskDatabaseCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipePrepTaskDatabaseCreationInput{
			ID:              t.Name(),
			BelongsToRecipe: t.Name(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &RecipePrepTaskDatabaseCreationInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestRecipePrepTaskUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &RecipePrepTaskUpdateRequestInput{
			BelongsToRecipe: new(t.Name()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &RecipePrepTaskUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}
