package mealplanning

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidIngredientGroupCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientGroupCreationRequestInput{
			Name:        t.Name(),
			Description: t.Name(),
			Slug:        t.Name(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientGroupCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidIngredientGroupUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientGroupUpdateRequestInput{
			Name:        new(t.Name()),
			Description: new(t.Name()),
			Slug:        new(t.Name()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &ValidIngredientGroupUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestValidIngredientGroupCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ValidIngredientGroupCreationRequestInput{
			Name: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestValidIngredientGroupDatabaseCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ValidIngredientGroupDatabaseCreationInput{
			ID:   t.Name(),
			Name: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestValidIngredientGroupUpdateRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ValidIngredientGroupUpdateRequestInput{
			Name: new(t.Name()),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestValidIngredientGroup_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual := &ValidIngredientGroup{
			Name:        "",
			Description: "",
			Slug:        "",
		}

		input := &ValidIngredientGroupUpdateRequestInput{
			Name:        new(t.Name()),
			Description: new(t.Name()),
			Slug:        new(t.Name()),
		}

		expected := &ValidIngredientGroup{
			Name:        *input.Name,
			Description: *input.Description,
			Slug:        *input.Slug,
		}

		actual.Update(input)

		assert.Equal(t, expected, actual)
	})
}
