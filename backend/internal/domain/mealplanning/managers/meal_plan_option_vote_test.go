package managers

import (
	"context"
	"database/sql"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMealPlanningManager_ListMealPlanOptionVotes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanOptionVotesList()
		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleMealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			GetMealPlanOptionVotesFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanOptionVote], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, exampleMealPlanOptionID, mealPlanOptionID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlanOptionVotes(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, exampleOwnerID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanOptionVotesCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlanOptionVotes(ctx, exampleMealPlanID, fake.BuildFakeID(), fake.BuildFakeID(), exampleOwnerID, nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetMealPlanOptionVotesCalls())
	})
}

func TestMealPlanningManager_CreateMealPlanOptionVotes(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		creatorID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOptionVotesList().Data
		fakeInput := fakes.BuildFakeMealPlanOptionVoteCreationRequestInput()

		// every vote on the input resolves its own meal plan option.
		votedOptionIDs := map[string]bool{}
		for _, vote := range fakeInput.Votes {
			votedOptionIDs[vote.BelongsToMealPlanOption] = true
		}

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
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

		actual, err := mpm.CreateMealPlanOptionVotes(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleOwnerID, creatorID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.MealPlanEventIsEligibleForVotingCalls(), 1)
		assert.Len(t, db.GetMealPlanOptionCalls(), len(fakeInput.Votes))
		assert.Len(t, db.CreateMealPlanOptionVoteCalls(), 1)
	})

	T.Run("with event not eligible for voting", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		creatorID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeMealPlanOptionVoteCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanEventIsEligibleForVotingFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string) (bool, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)

				return false, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanOptionVotes(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleOwnerID, creatorID, fakeInput)
		assert.Nil(t, actual)
		require.ErrorIs(t, err, types.ErrMealPlanEventNotEligibleForVoting)

		assert.Len(t, db.MealPlanEventIsEligibleForVotingCalls(), 1)
	})

	T.Run("with option not belonging to event", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		creatorID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeMealPlanOptionVoteCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
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

		actual, err := mpm.CreateMealPlanOptionVotes(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleOwnerID, creatorID, fakeInput)
		assert.Nil(t, actual)
		require.ErrorIs(t, err, types.ErrMealPlanOptionNotFoundForEvent)

		assert.Len(t, db.MealPlanEventIsEligibleForVotingCalls(), 1)
		assert.Len(t, db.GetMealPlanOptionCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanOptionVotes(ctx, exampleMealPlanID, fake.BuildFakeID(), exampleOwnerID, fake.BuildFakeID(), fakes.BuildFakeMealPlanOptionVoteCreationRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.MealPlanEventIsEligibleForVotingCalls())
		assert.Empty(t, db.CreateMealPlanOptionVoteCalls())
	})
}

func TestMealPlanningManager_ReadMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleMealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOptionVote()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			GetMealPlanOptionVoteFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string, mealPlanOptionVoteID string) (*types.MealPlanOptionVote, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, exampleMealPlanOptionID, mealPlanOptionID)
				assert.Equal(t, expected.ID, mealPlanOptionVoteID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanOptionVote(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, expected.ID, exampleOwnerID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanOptionVoteCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanOptionVote(ctx, exampleMealPlanID, fake.BuildFakeID(), fake.BuildFakeID(), fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetMealPlanOptionVoteCalls())
	})
}

func TestMealPlanningManager_UpdateMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanOptionVote := fakes.BuildFakeMealPlanOptionVote()
		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanOptionID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanOptionVoteUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
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

		require.NoError(t, mpm.UpdateMealPlanOptionVote(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, exampleMealPlanOptionVote.ID, exampleOwnerID, exampleInput))

		assert.Len(t, db.GetMealPlanOptionVoteCalls(), 1)
		assert.Len(t, db.UpdateMealPlanOptionVoteCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.UpdateMealPlanOptionVote(ctx, exampleMealPlanID, fake.BuildFakeID(), fake.BuildFakeID(), fake.BuildFakeID(), exampleOwnerID, fakes.BuildFakeMealPlanOptionVoteUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.GetMealPlanOptionVoteCalls())
		assert.Empty(t, db.UpdateMealPlanOptionVoteCalls())
	})
}

func TestMealPlanningManager_ArchiveMealPlanOptionVote(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fake.BuildFakeID()
		mealPlanEventID := fake.BuildFakeID()
		mealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOptionVote()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:       mealPlanExistsStub(t, mealPlanID, exampleOwnerID, true),
			MealPlanOptionExistsFunc: mealPlanOptionExistsStub(t, mealPlanID, mealPlanEventID, mealPlanOptionID, true),
			ArchiveMealPlanOptionVoteFunc: func(_ context.Context, actualMealPlanID string, actualMealPlanEventID string, actualMealPlanOptionID string, mealPlanOptionVoteID string) error {
				assert.Equal(t, mealPlanID, actualMealPlanID)
				assert.Equal(t, mealPlanEventID, actualMealPlanEventID)
				assert.Equal(t, mealPlanOptionID, actualMealPlanOptionID)
				assert.Equal(t, expected.ID, mealPlanOptionVoteID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanOptionVote(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, expected.ID, exampleOwnerID)
		require.NoError(t, err)

		assert.Len(t, db.ArchiveMealPlanOptionVoteCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanOptionVote(ctx, exampleMealPlanID, fake.BuildFakeID(), fake.BuildFakeID(), fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveMealPlanOptionVoteCalls())
	})

	// The archive is keyed by the vote and its option, so the option has to be shown to be the plan's.
	T.Run("with an option from another meal plan", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleMealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:       mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanOptionExistsFunc: mealPlanOptionExistsStub(t, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanOptionVote(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveMealPlanOptionVoteCalls())
	})
}
