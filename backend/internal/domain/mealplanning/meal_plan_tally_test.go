package mealplanning_test

import (
	"context"
	"errors"
	"testing"
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/primitives-go/v2/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// firstTiebreak always picks the first of the tied options.
func firstTiebreak(int) int { return 0 }

// buildMealPlanForTally builds a plan awaiting votes with the given number of events, each
// offering two options and carrying no ballots yet.
func buildMealPlanForTally(eventCount int) *types.MealPlan {
	mealPlan := fakes.BuildFakeMealPlan()

	mealPlan.Events = []*types.MealPlanEvent{}
	for range eventCount {
		event := fakes.BuildFakeMealPlanEvent()
		event.BelongsToMealPlan = mealPlan.ID
		event.Options = event.Options[:2]
		for _, option := range event.Options {
			option.Votes = []*types.MealPlanOptionVote{}
		}
		mealPlan.Events = append(mealPlan.Events, event)
	}

	return mealPlan
}

// castBallot records voter ranking an event's options in the order given, first preferred.
func castBallot(event *types.MealPlanEvent, voter string, preference ...*types.MealPlanOption) {
	for i, option := range preference {
		vote := fakes.BuildFakeMealPlanOptionVote()
		vote.BelongsToMealPlanOption = option.ID
		vote.ByUser = voter
		vote.Abstain = false
		vote.Rank = uint8(i + 1)
		option.Votes = append(option.Votes, vote)
	}
}

func TestInitialMealPlanStatus(T *testing.T) {
	T.Parallel()

	T.Run("with an event offering a choice", func(t *testing.T) {
		t.Parallel()

		events := []*types.MealPlanEventDatabaseCreationInput{
			{Options: []*types.MealPlanOptionDatabaseCreationInput{{ID: fake.BuildFakeID()}}},
			{Options: []*types.MealPlanOptionDatabaseCreationInput{{ID: fake.BuildFakeID()}, {ID: fake.BuildFakeID()}}},
		}

		assert.Equal(t, types.MealPlanStatusAwaitingVotes, types.InitialMealPlanStatus(events))
	})

	T.Run("with nothing to vote on", func(t *testing.T) {
		t.Parallel()

		events := []*types.MealPlanEventDatabaseCreationInput{
			{Options: []*types.MealPlanOptionDatabaseCreationInput{{ID: fake.BuildFakeID()}}},
			{Options: []*types.MealPlanOptionDatabaseCreationInput{}},
		}

		assert.Equal(t, types.MealPlanStatusFinalized, types.InitialMealPlanStatus(events))
	})
}

func TestTallyMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("decides every event once every member has voted", func(t *testing.T) {
		t.Parallel()

		alice, bob := fake.BuildFakeID(), fake.BuildFakeID()
		mealPlan := buildMealPlanForTally(2)
		for _, event := range mealPlan.Events {
			castBallot(event, alice, event.Options[1], event.Options[0])
			castBallot(event, bob, event.Options[1], event.Options[0])
		}

		tally, err := types.TallyMealPlan(mealPlan, []string{alice, bob}, time.Now(), firstTiebreak)
		require.NoError(t, err)

		assert.True(t, tally.Finalized)
		assert.Empty(t, tally.AwaitingVotesFrom)
		require.Len(t, tally.Decisions, 2)
		for i, decision := range tally.Decisions {
			assert.Equal(t, mealPlan.Events[i].ID, decision.MealPlanEventID)
			assert.Equal(t, mealPlan.Events[i].Options[1].ID, decision.MealPlanOptionID)
			assert.False(t, decision.Tiebroken)
		}
	})

	T.Run("holds the plan open before the deadline while a member has not voted", func(t *testing.T) {
		t.Parallel()

		voter, abstainer := fake.BuildFakeID(), fake.BuildFakeID()
		mealPlan := buildMealPlanForTally(1)
		event := mealPlan.Events[0]
		castBallot(event, voter, event.Options...)

		tally, err := types.TallyMealPlan(mealPlan, []string{voter, abstainer}, mealPlan.VotingDeadline.Add(-time.Minute), firstTiebreak)
		require.NoError(t, err)

		assert.False(t, tally.Finalized)
		assert.Empty(t, tally.Decisions)
		assert.Equal(t, []string{abstainer}, tally.AwaitingVotesFrom)
	})

	T.Run("decides no event after the first one still waiting", func(t *testing.T) {
		t.Parallel()

		voter, latecomer := fake.BuildFakeID(), fake.BuildFakeID()
		mealPlan := buildMealPlanForTally(3)
		for _, event := range mealPlan.Events {
			castBallot(event, voter, event.Options...)
		}
		// Every member voted on the first and last events; the latecomer skipped the middle.
		for _, i := range []int{0, 2} {
			castBallot(mealPlan.Events[i], latecomer, mealPlan.Events[i].Options...)
		}

		tally, err := types.TallyMealPlan(mealPlan, []string{voter, latecomer}, mealPlan.VotingDeadline.Add(-time.Minute), firstTiebreak)
		require.NoError(t, err)

		assert.False(t, tally.Finalized)
		require.Len(t, tally.Decisions, 1)
		assert.Equal(t, mealPlan.Events[0].ID, tally.Decisions[0].MealPlanEventID)
		assert.Equal(t, []string{latecomer}, tally.AwaitingVotesFrom)
	})

	T.Run("decides on the ballots it has once the deadline has passed", func(t *testing.T) {
		t.Parallel()

		voter, abstainer := fake.BuildFakeID(), fake.BuildFakeID()
		mealPlan := buildMealPlanForTally(1)
		event := mealPlan.Events[0]
		castBallot(event, voter, event.Options[1], event.Options[0])

		tally, err := types.TallyMealPlan(mealPlan, []string{voter, abstainer}, mealPlan.VotingDeadline.Add(time.Minute), firstTiebreak)
		require.NoError(t, err)

		assert.True(t, tally.Finalized)
		require.Len(t, tally.Decisions, 1)
		assert.Equal(t, event.Options[1].ID, tally.Decisions[0].MealPlanOptionID)
		assert.Equal(t, []string{abstainer}, tally.AwaitingVotesFrom)
	})

	T.Run("finalizes around an event nobody voted on once the deadline has passed", func(t *testing.T) {
		t.Parallel()

		mealPlan := buildMealPlanForTally(1)

		tally, err := types.TallyMealPlan(mealPlan, []string{fake.BuildFakeID()}, mealPlan.VotingDeadline.Add(time.Minute), firstTiebreak)
		require.NoError(t, err)

		assert.True(t, tally.Finalized)
		assert.Empty(t, tally.Decisions)
	})

	T.Run("leaves an event voting already settled alone", func(t *testing.T) {
		t.Parallel()

		mealPlan := buildMealPlanForTally(1)
		mealPlan.Events[0].Options[0].Chosen = true

		tally, err := types.TallyMealPlan(mealPlan, []string{fake.BuildFakeID()}, mealPlan.VotingDeadline.Add(-time.Minute), firstTiebreak)
		require.NoError(t, err)

		assert.True(t, tally.Finalized)
		assert.Empty(t, tally.Decisions)
		assert.Empty(t, tally.AwaitingVotesFrom)
	})

	T.Run("breaks a tie with the tiebreak it was given", func(t *testing.T) {
		t.Parallel()

		alice, bob := fake.BuildFakeID(), fake.BuildFakeID()
		mealPlan := buildMealPlanForTally(1)
		event := mealPlan.Events[0]
		castBallot(event, alice, event.Options[0], event.Options[1])
		castBallot(event, bob, event.Options[1], event.Options[0])

		chosen := map[string]bool{}
		for pick := range 2 {
			var offered int
			tally, err := types.TallyMealPlan(mealPlan, []string{alice, bob}, time.Now(), func(n int) int {
				offered = n
				return pick
			})
			require.NoError(t, err)

			assert.Equal(t, 2, offered)
			require.Len(t, tally.Decisions, 1)
			assert.True(t, tally.Decisions[0].Tiebroken)
			chosen[tally.Decisions[0].MealPlanOptionID] = true
		}

		assert.Equal(t, map[string]bool{event.Options[0].ID: true, event.Options[1].ID: true}, chosen)
	})

	T.Run("with a plan that is already finalized", func(t *testing.T) {
		t.Parallel()

		mealPlan := buildMealPlanForTally(1)
		mealPlan.Status = string(types.MealPlanStatusFinalized)

		tally, err := types.TallyMealPlan(mealPlan, nil, time.Now(), firstTiebreak)
		assert.Nil(t, tally)
		require.ErrorIs(t, err, types.ErrAlreadyFinalized)
	})

	T.Run("with nil meal plan", func(t *testing.T) {
		t.Parallel()

		tally, err := types.TallyMealPlan(nil, nil, time.Now(), firstTiebreak)
		assert.Nil(t, tally)
		require.Error(t, err)
	})
}

type electorateForTest struct {
	err     error
	members []string
}

func (e *electorateForTest) MembersOfAccount(context.Context, string) ([]string, error) {
	return e.members, e.err
}

func TestFinalizeMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("records what the tally decided", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		voter := fake.BuildFakeID()
		accountID := fake.BuildFakeID()
		mealPlan := buildMealPlanForTally(1)
		event := mealPlan.Events[0]
		castBallot(event, voter, event.Options[1], event.Options[0])

		repo := &mealplanningmock.RepositoryMock{
			GetMealPlanFunc: func(_ context.Context, mealPlanID, gotAccountID string) (*types.MealPlan, error) {
				assert.Equal(t, mealPlan.ID, mealPlanID)
				assert.Equal(t, accountID, gotAccountID)

				return mealPlan, nil
			},
			RecordMealPlanTallyFunc: func(_ context.Context, recorded *types.MealPlan, tally *types.MealPlanTally) error {
				assert.Equal(t, mealPlan, recorded)
				assert.True(t, tally.Finalized)

				return nil
			},
		}

		tally, err := types.FinalizeMealPlan(ctx, repo, &electorateForTest{members: []string{voter}}, mealPlan.ID, accountID, time.Now(), firstTiebreak)
		require.NoError(t, err)

		require.Len(t, tally.Decisions, 1)
		assert.Equal(t, event.Options[1].ID, tally.Decisions[0].MealPlanOptionID)
		assert.Len(t, repo.RecordMealPlanTallyCalls(), 1)
	})

	T.Run("with an electorate that cannot be read", func(t *testing.T) {
		t.Parallel()

		repo := &mealplanningmock.RepositoryMock{}
		expectedErr := errors.New(fake.BuildFakeID())

		tally, err := types.FinalizeMealPlan(t.Context(), repo, &electorateForTest{err: expectedErr}, fake.BuildFakeID(), fake.BuildFakeID(), time.Now(), firstTiebreak)
		assert.Nil(t, tally)
		require.ErrorIs(t, err, expectedErr)

		assert.Empty(t, repo.GetMealPlanCalls())
	})

	T.Run("with a plan that is already finalized", func(t *testing.T) {
		t.Parallel()

		mealPlan := buildMealPlanForTally(1)
		mealPlan.Status = string(types.MealPlanStatusFinalized)

		repo := &mealplanningmock.RepositoryMock{
			GetMealPlanFunc: func(context.Context, string, string) (*types.MealPlan, error) {
				return mealPlan, nil
			},
		}

		tally, err := types.FinalizeMealPlan(t.Context(), repo, &electorateForTest{}, mealPlan.ID, fake.BuildFakeID(), time.Now(), firstTiebreak)
		assert.Nil(t, tally)
		require.ErrorIs(t, err, types.ErrAlreadyFinalized)

		assert.Empty(t, repo.RecordMealPlanTallyCalls())
	})

	T.Run("with missing IDs", func(t *testing.T) {
		t.Parallel()

		repo := &mealplanningmock.RepositoryMock{}

		tally, err := types.FinalizeMealPlan(t.Context(), repo, &electorateForTest{}, "", fake.BuildFakeID(), time.Now(), firstTiebreak)
		assert.Nil(t, tally)
		require.Error(t, err)
	})
}
