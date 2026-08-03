package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidIngredient_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual := &ValidIngredient{}

		input := &ValidIngredientUpdateRequestInput{}
		require.NoError(t, fake.Struct(&input))
		input.ContainsEgg = new(true)
		input.ContainsDairy = new(true)
		input.ContainsPeanut = new(true)
		input.ContainsTreeNut = new(true)
		input.ContainsSoy = new(true)
		input.ContainsWheat = new(true)
		input.ContainsShellfish = new(true)
		input.ContainsSesame = new(true)
		input.ContainsFish = new(true)
		input.ContainsGluten = new(true)
		input.AnimalFlesh = new(true)
		input.AnimalDerived = new(true)
		input.RestrictToPreparations = new(true)
		input.ContainsAlcohol = new(true)
		input.IsLiquid = new(true)
		input.IsProtein = new(true)
		input.IsStarch = new(true)
		input.IsGrain = new(true)
		input.IsFruit = new(true)
		input.IsSalt = new(true)
		input.IsFat = new(true)
		input.IsAcid = new(true)
		input.IsHeat = new(true)

		actual.Update(input)
	})
}

func TestValidIngredientCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ValidIngredientCreationRequestInput{
			Name: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestValidIngredientDatabaseCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ValidIngredientDatabaseCreationInput{
			ID:   t.Name(),
			Name: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestValidIngredientUpdateRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ValidIngredientUpdateRequestInput{
			Name: new(t.Name()),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
