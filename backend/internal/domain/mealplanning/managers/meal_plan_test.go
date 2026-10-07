package managers

import (
	"context"
	"testing"
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	clockmock "github.com/primandproper/primitives-go/v2/clock/mock"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMealPlanningManager_ListMealPlans(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlansList()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlansForAccountFunc: func(_ context.Context, accountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlan], error) {
				assert.Equal(t, exampleOwnerID, accountID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlans(ctx, exampleOwnerID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlansForAccountCalls(), 1)
	})
}

func TestMealPlanningManager_AnnotateMealPlanSummaries(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		votedOn := fakes.BuildFakeMealPlan()
		notVotedOn := fakes.BuildFakeMealPlan()
		exampleUserID := fake.BuildFakeID()
		exampleChosenMealName := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetChosenMealNamesForMealPlansFunc: func(_ context.Context, mealPlanIDs []string) (map[string]string, error) {
				assert.Equal(t, []string{votedOn.ID, notVotedOn.ID}, mealPlanIDs)

				return map[string]string{votedOn.Events[0].ID: exampleChosenMealName}, nil
			},
			GetMealPlanIDsVotedOnByUserFunc: func(_ context.Context, userID string, mealPlanIDs []string) ([]string, error) {
				assert.Equal(t, exampleUserID, userID)
				assert.Equal(t, []string{votedOn.ID, notVotedOn.ID}, mealPlanIDs)

				return []string{votedOn.ID}, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.AnnotateMealPlanSummaries(ctx, exampleUserID, []*types.MealPlan{votedOn, notVotedOn})
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, exampleChosenMealName, actual.ChosenMealNamesByEventID[votedOn.Events[0].ID])
		assert.True(t, actual.VotedOnMealPlanIDs[votedOn.ID])
		assert.False(t, actual.VotedOnMealPlanIDs[notVotedOn.ID])

		assert.Len(t, db.GetChosenMealNamesForMealPlansCalls(), 1)
		assert.Len(t, db.GetMealPlanIDsVotedOnByUserCalls(), 1)
	})

	T.Run("with empty page", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		db := &mealplanningmock.RepositoryMock{}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.AnnotateMealPlanSummaries(ctx, fake.BuildFakeID(), nil)
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Empty(t, actual.ChosenMealNamesByEventID)
		assert.Empty(t, actual.VotedOnMealPlanIDs)

		// An empty page asks the database nothing.
		assert.Empty(t, db.GetChosenMealNamesForMealPlansCalls())
		assert.Empty(t, db.GetMealPlanIDsVotedOnByUserCalls())
	})

	T.Run("with empty user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		actual, err := mpm.AnnotateMealPlanSummaries(ctx, "", []*types.MealPlan{fakes.BuildFakeMealPlan()})
		assert.Nil(t, actual)
		require.Error(t, err)
	})
}

func TestMealPlanningManager_CreateMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		ownerID := fake.BuildFakeID()
		creatorID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlan()
		fakeInput := fakes.BuildFakeMealPlanCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateMealPlanFunc: func(_ context.Context, input *types.MealPlanDatabaseCreationInput) (*types.MealPlan, error) {
				assert.Equal(t, ownerID, input.BelongsToAccount)
				assert.Equal(t, creatorID, input.CreatedByUser)
				assert.Equal(t, string(types.InitialMealPlanStatus(input.Events)), input.Status)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlan(ctx, ownerID, creatorID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanCalls(), 1)
	})

	T.Run("with a voting deadline already behind the manager's clock", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		fakeInput := fakes.BuildFakeMealPlanCreationRequestInput()
		mpm := newManagerForTestWithClock(t, nil, &clockmock.ClockMock{
			NowFunc: func() time.Time { return fakeInput.VotingDeadline.Add(time.Minute) },
		})

		db := &mealplanningmock.RepositoryMock{}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlan(ctx, fake.BuildFakeID(), fake.BuildFakeID(), fakeInput)
		require.Error(t, err)
		assert.Nil(t, actual)

		assert.Empty(t, db.CreateMealPlanCalls())
	})

	T.Run("starts the finalization saga when meal plan is created finalized", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		starter := &fakeFinalizationStarter{}

		mpm := buildMealPlanManagerForTestWithStarter(t, starter)

		ownerID := fake.BuildFakeID()
		creatorID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlan()
		expected.Status = string(types.MealPlanStatusFinalized)
		fakeInput := fakes.BuildFakeMealPlanCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateMealPlanFunc: func(_ context.Context, _ *types.MealPlanDatabaseCreationInput) (*types.MealPlan, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlan(ctx, ownerID, creatorID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanCalls(), 1)
		assert.Equal(t, []string{expected.ID}, starter.calls)
	})
}

func TestMealPlanningManager_ReadMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlan()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanFunc: func(_ context.Context, mealPlanID, accountID string) (*types.MealPlan, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, expected.ID, accountID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlan(ctx, exampleMealPlanID, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanCalls(), 1)
	})
}

func TestMealPlanningManager_UpdateMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlan := fakes.BuildFakeMealPlan()
		ownerID := fake.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanFunc: func(_ context.Context, mealPlanID, accountID string) (*types.MealPlan, error) {
				assert.Equal(t, exampleMealPlan.ID, mealPlanID)
				assert.Equal(t, ownerID, accountID)

				return exampleMealPlan, nil
			},
			UpdateMealPlanFunc: func(_ context.Context, _ *types.MealPlan) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		require.NoError(t, mpm.UpdateMealPlan(ctx, exampleMealPlan.ID, ownerID, exampleInput))

		assert.Len(t, db.GetMealPlanCalls(), 1)
		assert.Len(t, db.UpdateMealPlanCalls(), 1)
	})
}

func TestMealPlanningManager_ArchiveMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlan()

		db := &mealplanningmock.RepositoryMock{
			ArchiveMealPlanFunc: func(_ context.Context, mealPlanID, accountID string) error {
				assert.Equal(t, expected.ID, mealPlanID)
				assert.Equal(t, expected.CreatedByUser, accountID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlan(ctx, expected.ID, expected.CreatedByUser)
		require.NoError(t, err)

		assert.Len(t, db.ArchiveMealPlanCalls(), 1)
	})
}

// buildVotedMealPlanForTest builds a plan awaiting votes whose one event offers two options,
// with a ballot from every member on each, so a tally of it finalizes.
func buildVotedMealPlanForTest(members []string) *types.MealPlan {
	mealPlan := fakes.BuildFakeMealPlan()
	event := mealPlan.Events[0]
	event.Options = event.Options[:2]
	mealPlan.Events = []*types.MealPlanEvent{event}

	for i, option := range event.Options {
		option.Votes = []*types.MealPlanOptionVote{}
		for _, member := range members {
			vote := fakes.BuildFakeMealPlanOptionVote()
			vote.BelongsToMealPlanOption = option.ID
			vote.ByUser = member
			vote.Abstain = false
			vote.Rank = uint8(i + 1)
			option.Votes = append(option.Votes, vote)
		}
	}

	return mealPlan
}

func TestMealPlanningManager_FinalizeMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		members := []string{fake.BuildFakeID(), fake.BuildFakeID()}
		mpm.electorate = &fakeElectorate{members: members}

		expected := buildVotedMealPlanForTest(members)
		exampleAccountID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanFunc: func(_ context.Context, mealPlanID, accountID string) (*types.MealPlan, error) {
				assert.Equal(t, expected.ID, mealPlanID)
				assert.Equal(t, exampleAccountID, accountID)

				return expected, nil
			},
			RecordMealPlanTallyFunc: func(_ context.Context, mealPlan *types.MealPlan, tally *types.MealPlanTally) error {
				assert.Equal(t, expected, mealPlan)
				assert.True(t, tally.Finalized)
				require.Len(t, tally.Decisions, 1)
				assert.Equal(t, expected.Events[0].Options[0].ID, tally.Decisions[0].MealPlanOptionID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		finalized, err := mpm.FinalizeMealPlan(ctx, expected.ID, exampleAccountID)
		assert.True(t, finalized)
		require.NoError(t, err)

		assert.Len(t, db.RecordMealPlanTallyCalls(), 1)
	})

	T.Run("starts the finalization saga when finalized", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		starter := &fakeFinalizationStarter{}

		mpm := buildMealPlanManagerForTestWithStarter(t, starter)

		members := []string{fake.BuildFakeID()}
		mpm.electorate = &fakeElectorate{members: members}

		expected := buildVotedMealPlanForTest(members)

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanFunc: func(context.Context, string, string) (*types.MealPlan, error) {
				return expected, nil
			},
			RecordMealPlanTallyFunc: func(context.Context, *types.MealPlan, *types.MealPlanTally) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		finalized, err := mpm.FinalizeMealPlan(ctx, expected.ID, fake.BuildFakeID())
		assert.True(t, finalized)
		require.NoError(t, err)

		assert.Equal(t, []string{expected.ID}, starter.calls)
	})

	T.Run("leaves the saga alone while a member has yet to vote", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		starter := &fakeFinalizationStarter{}

		mpm := buildMealPlanManagerForTestWithStarter(t, starter)

		voter := fake.BuildFakeID()
		mpm.electorate = &fakeElectorate{members: []string{voter, fake.BuildFakeID()}}

		expected := buildVotedMealPlanForTest([]string{voter})

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanFunc: func(context.Context, string, string) (*types.MealPlan, error) {
				return expected, nil
			},
			RecordMealPlanTallyFunc: func(_ context.Context, _ *types.MealPlan, tally *types.MealPlanTally) error {
				assert.False(t, tally.Finalized)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		finalized, err := mpm.FinalizeMealPlan(ctx, expected.ID, fake.BuildFakeID())
		assert.False(t, finalized)
		require.NoError(t, err)

		assert.Empty(t, starter.calls)
	})

	T.Run("with a plan that is already finalized", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlan()
		expected.Status = string(types.MealPlanStatusFinalized)

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanFunc: func(context.Context, string, string) (*types.MealPlan, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		finalized, err := mpm.FinalizeMealPlan(ctx, expected.ID, fake.BuildFakeID())
		assert.False(t, finalized)
		require.ErrorIs(t, err, types.ErrAlreadyFinalized)

		assert.Empty(t, db.RecordMealPlanTallyCalls())
	})
}
