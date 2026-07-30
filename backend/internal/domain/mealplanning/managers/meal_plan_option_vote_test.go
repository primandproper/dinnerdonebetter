package managers

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v8/filtering"

	"github.com/stretchr/testify/assert"
)

func TestMealPlanningManager_ListMealPlanOptionVotes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanOptionVotesList()
		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleMealPlanOptionID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanOptionVotesFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanOptionVote], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, exampleMealPlanOptionID, mealPlanOptionID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlanOptionVotes(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanOptionVotesCalls(), 1)
	})
}

func TestMealPlanningManager_CreateMealPlanOptionVotes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		creatorID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOptionVotesList().Data
		fakeInput := fakes.BuildFakeMealPlanOptionVoteCreationRequestInput()

		// every vote on the input resolves its own meal plan option.
		votedOptionIDs := map[string]bool{}
		for _, vote := range fakeInput.Votes {
			votedOptionIDs[vote.BelongsToMealPlanOption] = true
		}

		db := &mealplanningmock.RepositoryMock{
			MealPlanEventIsEligibleForVotingFunc: func(_ context.Context, mealPlanID, mealPlanEventID string) (bool, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)

				return true, nil
			},
			GetMealPlanOptionFunc: func(_ context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) (*types.MealPlanOption, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.True(t, votedOptionIDs[mealPlanOptionID], "unexpected meal plan option fetched: %s", mealPlanOptionID)

				return fakes.BuildFakeMealPlanOption(), nil
			},
			CreateMealPlanOptionVoteFunc: func(_ context.Context, _ *types.MealPlanOptionVotesDatabaseCreationInput) ([]*types.MealPlanOptionVote, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanOptionVotes(ctx, exampleMealPlanID, exampleMealPlanEventID, creatorID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.MealPlanEventIsEligibleForVotingCalls(), 1)
		assert.Len(t, db.GetMealPlanOptionCalls(), len(fakeInput.Votes))
		assert.Len(t, db.CreateMealPlanOptionVoteCalls(), 1)
	})

	T.Run("with event not eligible for voting", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		creatorID := fakes.BuildFakeID()
		fakeInput := fakes.BuildFakeMealPlanOptionVoteCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanEventIsEligibleForVotingFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string) (bool, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)

				return false, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanOptionVotes(ctx, exampleMealPlanID, exampleMealPlanEventID, creatorID, fakeInput)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, types.ErrMealPlanEventNotEligibleForVoting)

		assert.Len(t, db.MealPlanEventIsEligibleForVotingCalls(), 1)
	})

	T.Run("with option not belonging to event", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		creatorID := fakes.BuildFakeID()
		fakeInput := fakes.BuildFakeMealPlanOptionVoteCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanEventIsEligibleForVotingFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string) (bool, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)

				return true, nil
			},
			GetMealPlanOptionFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, _ string) (*types.MealPlanOption, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)

				return nil, sql.ErrNoRows
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanOptionVotes(ctx, exampleMealPlanID, exampleMealPlanEventID, creatorID, fakeInput)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, types.ErrMealPlanOptionNotFoundForEvent)

		assert.Len(t, db.MealPlanEventIsEligibleForVotingCalls(), 1)
		assert.Len(t, db.GetMealPlanOptionCalls(), 1)
	})
}

func TestMealPlanningManager_ReadMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleMealPlanOptionID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOptionVote()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanOptionVoteFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string, mealPlanOptionVoteID string) (*types.MealPlanOptionVote, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, exampleMealPlanOptionID, mealPlanOptionID)
				assert.Equal(t, expected.ID, mealPlanOptionVoteID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanOptionVote(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanOptionVoteCalls(), 1)
	})
}

func TestMealPlanningManager_UpdateMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanOptionVote := fakes.BuildFakeMealPlanOptionVote()
		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanOptionID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanOptionVoteUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanOptionVoteFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string, mealPlanOptionVoteID string) (*types.MealPlanOptionVote, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, exampleMealPlanOptionID, mealPlanOptionID)
				assert.Equal(t, exampleMealPlanOptionVote.ID, mealPlanOptionVoteID)

				return exampleMealPlanOptionVote, nil
			},
			UpdateMealPlanOptionVoteFunc: func(_ context.Context, _ *types.MealPlanOptionVote) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		assert.NoError(t, mpm.UpdateMealPlanOptionVote(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, exampleMealPlanOptionVote.ID, exampleInput))

		assert.Len(t, db.GetMealPlanOptionVoteCalls(), 1)
		assert.Len(t, db.UpdateMealPlanOptionVoteCalls(), 1)
	})
}

func TestMealPlanningManager_ArchiveMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fakes.BuildFakeID()
		mealPlanEventID := fakes.BuildFakeID()
		mealPlanOptionID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOptionVote()

		db := &mealplanningmock.RepositoryMock{
			ArchiveMealPlanOptionVoteFunc: func(_ context.Context, actualMealPlanID string, actualMealPlanEventID string, actualMealPlanOptionID string, mealPlanOptionVoteID string) error {
				assert.Equal(t, mealPlanID, actualMealPlanID)
				assert.Equal(t, mealPlanEventID, actualMealPlanEventID)
				assert.Equal(t, mealPlanOptionID, actualMealPlanOptionID)
				assert.Equal(t, expected.ID, mealPlanOptionVoteID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanOptionVote(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, expected.ID)
		assert.NoError(t, err)

		assert.Len(t, db.ArchiveMealPlanOptionVoteCalls(), 1)
	})
}
