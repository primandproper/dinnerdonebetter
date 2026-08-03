package mealplanning

import (
	"math"
	"testing"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMealPlanOptionVote_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanOptionVote{}
		input := &MealPlanOptionVoteUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))
		input.Abstain = new(true)

		x.Update(input)
	})
}

func TestMealPlanOptionVoteCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		// every vote must name the meal plan option it belongs to.
		x := &MealPlanOptionVoteCreationRequestInput{
			Votes: []*MealPlanOptionVoteCreationInput{
				{BelongsToMealPlanOption: t.Name()},
			},
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanOptionVoteCreationRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestMealPlanOptionVotesDatabaseCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanOptionVotesDatabaseCreationInput{
			Votes: []*MealPlanOptionVoteDatabaseCreationInput{
				{},
			},
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanOptionVotesDatabaseCreationInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}

func TestMealPlanOptionVoteUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanOptionVoteUpdateRequestInput{
			Rank:                    new(uint8(fake.Number(1, math.MaxUint8))),
			Abstain:                 new(fake.Bool()),
			Notes:                   new(t.Name()),
			BelongsToMealPlanOption: t.Name(),
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanOptionVoteUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.Error(t, actual)
	})
}
