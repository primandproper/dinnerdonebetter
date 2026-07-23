package managers

import (
	"database/sql"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanOptionVotes), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.ListMealPlanOptionVotes(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.MealPlanEventIsEligibleForVoting), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID).Return(true, nil)
				for _, vote := range fakeInput.Votes {
					db.On(reflection.GetMethodName(mpm.db.GetMealPlanOption), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, vote.BelongsToMealPlanOption).Return(fakes.BuildFakeMealPlanOption(), nil)
				}
				db.On(reflection.GetMethodName(mpm.db.CreateMealPlanOptionVote), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanOptionVotesDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.MealPlanOptionVoteCreatedServiceEventType: {"vote_count", "created"},
			},
		)

		actual, err := mpm.CreateMealPlanOptionVotes(ctx, exampleMealPlanID, exampleMealPlanEventID, creatorID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})

	T.Run("with event not eligible for voting", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		creatorID := fakes.BuildFakeID()
		fakeInput := fakes.BuildFakeMealPlanOptionVoteCreationRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.MealPlanEventIsEligibleForVoting), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID).Return(false, nil)
			},
		)

		actual, err := mpm.CreateMealPlanOptionVotes(ctx, exampleMealPlanID, exampleMealPlanEventID, creatorID, fakeInput)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, types.ErrMealPlanEventNotEligibleForVoting)

		mock.AssertExpectationsForObjects(t, expectations...)
	})

	T.Run("with option not belonging to event", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		creatorID := fakes.BuildFakeID()
		fakeInput := fakes.BuildFakeMealPlanOptionVoteCreationRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.MealPlanEventIsEligibleForVoting), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID).Return(true, nil)
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanOption), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, mock.AnythingOfType("string")).Return((*types.MealPlanOption)(nil), sql.ErrNoRows)
			},
		)

		actual, err := mpm.CreateMealPlanOptionVotes(ctx, exampleMealPlanID, exampleMealPlanEventID, creatorID, fakeInput)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, types.ErrMealPlanOptionNotFoundForEvent)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanOptionVote), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, expected.ID).Return(expected, nil)
			},
		)

		actual, err := mpm.ReadMealPlanOptionVote(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanOptionVote), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, exampleMealPlanOptionVote.ID).Return(exampleMealPlanOptionVote, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateMealPlanOptionVote), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanOptionVote]()).Return(nil)
			},
			map[string][]string{
				types.MealPlanOptionVoteUpdatedServiceEventType: {
					mealplanningkeys.MealPlanIDKey,
					mealplanningkeys.MealPlanEventIDKey,
					mealplanningkeys.MealPlanOptionIDKey,
					mealplanningkeys.MealPlanOptionVoteIDKey,
				},
			},
		)

		assert.NoError(t, mpm.UpdateMealPlanOptionVote(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOptionID, exampleMealPlanOptionVote.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.ArchiveMealPlanOptionVote), testutils.ContextMatcher, mealPlanID, mealPlanEventID, mealPlanOptionID, expected.ID).Return(nil)
			},
			map[string][]string{
				types.MealPlanOptionVoteArchivedServiceEventType: {
					mealplanningkeys.MealPlanIDKey,
					mealplanningkeys.MealPlanEventIDKey,
					mealplanningkeys.MealPlanOptionIDKey,
					mealplanningkeys.MealPlanOptionVoteIDKey,
				},
			},
		)

		err := mpm.ArchiveMealPlanOptionVote(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, expected.ID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
