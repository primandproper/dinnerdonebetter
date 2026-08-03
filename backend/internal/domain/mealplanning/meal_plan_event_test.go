package mealplanning

import (
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMealPlanEventCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &MealPlanEventCreationRequestInput{
			MealName: SecondBreakfastMealName,
			StartsAt: time.Now(),
			EndsAt:   time.Now().Add(time.Hour),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})

	T.Run("with invalid time", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		tt := time.Now()
		x := &MealPlanEventCreationRequestInput{
			MealName: SecondBreakfastMealName,
			StartsAt: tt,
			EndsAt:   tt,
		}

		assert.Error(t, x.ValidateWithContext(ctx))
	})
}

func TestMealPlanEventDatabaseCreationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &MealPlanEventDatabaseCreationInput{
			ID:                t.Name(),
			BelongsToMealPlan: t.Name(),
			MealName:          SecondBreakfastMealName,
			StartsAt:          time.Now(),
			EndsAt:            time.Now(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestMealPlanEventUpdateRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &MealPlanEventUpdateRequestInput{
			MealName: new(SecondBreakfastMealName),
			StartsAt: new(time.Now()),
			EndsAt:   new(time.Now()),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestMealPlanEvent_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanEvent{}
		input := &MealPlanEventUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))
		input.StartsAt = new(time.Now())
		input.EndsAt = new(time.Now())

		x.Update(input)
	})
}
