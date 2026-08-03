package mealplanning

import (
	"testing"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMealPlan_Update(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &MealPlan{}
		input := &MealPlanUpdateRequestInput{}

		require.NoError(t, fake.Struct(&input))

		x.Update(input)
	})
}

func TestMealPlanDatabaseCreationInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanDatabaseCreationInput{
			ID:               t.Name(),
			VotingDeadline:   time.Now().Add(24 * time.Hour),
			BelongsToAccount: t.Name(),
			CreatedByUser:    t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanDatabaseCreationInput{}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})
}

func TestMealPlanCreationRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("with voting deadline before all events", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		event1Start := now.Add(72 * time.Hour)    // Event 1 starts 3 days from now
		event2Start := now.Add(96 * time.Hour)    // Event 2 starts 4 days from now
		votingDeadline := now.Add(24 * time.Hour) // Voting deadline is 1 day from now

		x := &MealPlanCreationRequestInput{
			VotingDeadline: votingDeadline,
			ElectionMethod: MealPlanElectionMethodSchulze,
			Events: []*MealPlanEventCreationRequestInput{
				{
					MealName: BreakfastMealName,
					Notes:    t.Name(),
					StartsAt: event1Start,
					EndsAt:   event1Start.Add(2 * time.Hour),
				},
				{
					MealName: DinnerMealName,
					Notes:    t.Name(),
					StartsAt: event2Start,
					EndsAt:   event2Start.Add(2 * time.Hour),
				},
			},
		}

		assert.NoError(t, x.ValidateWithContext(t.Context()))
	})

	T.Run("with voting deadline after event start", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		eventStart := now.Add(24 * time.Hour)     // Event starts 1 day from now
		votingDeadline := now.Add(48 * time.Hour) // Voting deadline is 2 days from now (AFTER event start)

		x := &MealPlanCreationRequestInput{
			VotingDeadline: votingDeadline,
			ElectionMethod: MealPlanElectionMethodSchulze,
			Events: []*MealPlanEventCreationRequestInput{
				{
					MealName: BreakfastMealName,
					Notes:    t.Name(),
					StartsAt: eventStart,
					EndsAt:   eventStart.Add(2 * time.Hour),
				},
			},
		}

		err := x.ValidateWithContext(t.Context())
		require.Error(t, err)
		assert.Equal(t, errVotingDeadlineAfterStart, err)
	})

	T.Run("with voting deadline equal to event start", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		eventStart := now.Add(24 * time.Hour) // Event starts 1 day from now
		votingDeadline := eventStart          // Voting deadline equals event start (should fail)

		x := &MealPlanCreationRequestInput{
			VotingDeadline: votingDeadline,
			ElectionMethod: MealPlanElectionMethodSchulze,
			Events: []*MealPlanEventCreationRequestInput{
				{
					MealName: BreakfastMealName,
					Notes:    t.Name(),
					StartsAt: eventStart,
					EndsAt:   eventStart.Add(2 * time.Hour),
				},
			},
		}

		err := x.ValidateWithContext(t.Context())
		require.Error(t, err)
		assert.Equal(t, errVotingDeadlineAfterStart, err)
	})

	T.Run("with voting deadline in the past", func(t *testing.T) {
		t.Parallel()

		now := time.Now()
		eventStart := now.Add(48 * time.Hour)      // Event starts 2 days from now
		votingDeadline := now.Add(-24 * time.Hour) // Voting deadline is 1 day ago (in the past)

		x := &MealPlanCreationRequestInput{
			VotingDeadline: votingDeadline,
			ElectionMethod: MealPlanElectionMethodSchulze,
			Events: []*MealPlanEventCreationRequestInput{
				{
					MealName: BreakfastMealName,
					Notes:    t.Name(),
					StartsAt: eventStart,
					EndsAt:   eventStart.Add(2 * time.Hour),
				},
			},
		}

		err := x.ValidateWithContext(t.Context())
		require.Error(t, err)
		assert.Equal(t, errInvalidVotingDeadline, err)
	})

	T.Run("with invalid structure", func(t *testing.T) {
		t.Parallel()

		x := &MealPlanCreationRequestInput{}

		assert.Error(t, x.ValidateWithContext(t.Context()))
	})
}

func TestMealPlanUpdateRequestInput_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleTime := time.Now()

		x := &MealPlanUpdateRequestInput{
			VotingDeadline: &exampleTime,
		}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})

	T.Run("empty input is valid", func(t *testing.T) {
		t.Parallel()

		// Every field on an update is optional (see MealPlan.Update, which only applies non-nil fields),
		// so an empty partial update is legitimately valid and must not be rejected. Mirrors
		// MealPlanOptionUpdateRequestInput.
		x := &MealPlanUpdateRequestInput{}

		actual := x.ValidateWithContext(t.Context())
		assert.NoError(t, actual)
	})
}
