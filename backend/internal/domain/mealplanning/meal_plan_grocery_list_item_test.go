package mealplanning

import (
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMealPlanGroceryListItemCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &MealPlanGroceryListItemCreationRequestInput{
			BelongsToMealPlan:      t.Name(),
			ValidIngredientID:      t.Name(),
			ValidMeasurementUnitID: t.Name(),
			Status:                 MealPlanGroceryListItemStatusUnknown,
			MinQuantityNeeded:      1.23,
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestMealPlanGroceryListItemDatabaseCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &MealPlanGroceryListItemDatabaseCreationInput{
			ID:                     t.Name(),
			BelongsToMealPlan:      t.Name(),
			ValidIngredientID:      t.Name(),
			ValidMeasurementUnitID: t.Name(),
			MinQuantityNeeded:      1.23,

			MaxQuantityNeeded: new(float32(1.23)),
			Status:            MealPlanGroceryListItemStatusUnknown,
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestMealPlanGroceryListItemUpdateRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &MealPlanGroceryListItemUpdateRequestInput{
			BelongsToMealPlan:      new(t.Name()),
			ValidIngredientID:      new(t.Name()),
			ValidMeasurementUnitID: new(t.Name()),
			MinQuantityNeeded:      new(float32(1.23)),

			MaxQuantityNeeded: new(float32(1.23)),
			Status:            new(MealPlanGroceryListItemStatusUnknown),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestMealPlanGroceryListItem_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanGroceryListItem{
			PurchasedMeasurementUnit: &ValidMeasurementUnit{},
			MaxQuantityNeeded:        new(float32(1.23)),
		}
		input := &MealPlanGroceryListItemUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))
		input.PurchasedMeasurementUnitID = new(t.Name())
		input.MaxQuantityNeeded = new(float32(3.21))

		x.Update(input)
	})
}
