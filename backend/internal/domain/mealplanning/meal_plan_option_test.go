package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
)

func TestMealPlanOption_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanOption{}
		input := &MealPlanOptionUpdateRequestInput{}

		assert.NoError(t, fake.Struct(&input))

		x.Update(input)
	})
}

func TestMealPlanOptionCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanOptionCreationRequestInput{
			AssignedCook:       new(t.Name()),
			AssignedDishwasher: new(t.Name()),
			MealID:             t.Name(),
			Notes:              t.Name(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanOptionCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestMealPlanOptionDatabaseCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanOptionDatabaseCreationInput{
			ID:                     t.Name(),
			BelongsToMealPlanEvent: t.Name(),
			MealID:                 t.Name(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanOptionDatabaseCreationInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestMealPlanOptionUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanOptionUpdateRequestInput{
			AssignedCook:           new(t.Name()),
			AssignedDishwasher:     new(t.Name()),
			BelongsToMealPlanEvent: new(t.Name()),
			MealID:                 new(t.Name()),
			Notes:                  new(t.Name()),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("empty input is valid", func(t *testing.T) {
		t.Parallel()

		// Every field on an update is optional, and BelongsToMealPlanEvent is server-set (json:"-"),
		// so an empty partial update is legitimately valid and must not be rejected.
		x := &MealPlanOptionUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})
}
