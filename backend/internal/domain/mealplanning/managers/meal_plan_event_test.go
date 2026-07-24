package managers

import (
	"testing"
	"time"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v6/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMealPlanningManager_ListMealPlanEvents(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanEventsList()
		exampleMealPlanID := fakes.BuildFakeID()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanEvents), testutils.ContextMatcher, exampleMealPlanID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.ListMealPlanEvents(ctx, exampleMealPlanID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_CreateMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanEvent()
		fakeInput := fakes.BuildFakeMealPlanEventCreationRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.CreateMealPlanEvent), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanEventDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.MealPlanEventCreatedServiceEventType: {mealplanningkeys.MealPlanEventIDKey},
			},
		)

		actual, err := mpm.CreateMealPlanEvent(ctx, expected.BelongsToMealPlan, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ReadMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanEvent()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanEvent), testutils.ContextMatcher, exampleMealPlanID, expected.ID).Return(expected, nil)
			},
		)

		actual, err := mpm.ReadMealPlanEvent(ctx, exampleMealPlanID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_UpdateMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanEvent := fakes.BuildFakeMealPlanEvent()
		exampleMealPlanID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanEventUpdateRequestInput()
		exampleInput.StartsAt = &exampleMealPlanEvent.StartsAt

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanEvent), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEvent.ID).Return(exampleMealPlanEvent, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateMealPlanEvent), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanEvent]()).Return(nil)
			},
			map[string][]string{
				types.MealPlanEventUpdatedServiceEventType: {
					mealplanningkeys.MealPlanIDKey,
					mealplanningkeys.MealPlanEventIDKey,
				},
			},
		)

		assert.NoError(t, mpm.UpdateMealPlanEvent(ctx, exampleMealPlanID, exampleMealPlanEvent.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
	})

	T.Run("when start time changes clears notification sent for event", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanEvent := fakes.BuildFakeMealPlanEvent()
		exampleMealPlanID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanEventUpdateRequestInput()
		newStartsAt := exampleMealPlanEvent.StartsAt.Add(time.Hour)
		exampleInput.StartsAt = &newStartsAt

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanEvent), testutils.ContextMatcher, exampleMealPlanID, exampleMealPlanEvent.ID).Return(exampleMealPlanEvent, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateMealPlanEvent), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanEvent]()).Return(nil)
				db.On(reflection.GetMethodName(mpm.db.ClearMealPlanTaskNotificationSentForEvent), testutils.ContextMatcher, exampleMealPlanEvent.ID).Return(nil)
			},
			map[string][]string{
				types.MealPlanEventUpdatedServiceEventType: {
					mealplanningkeys.MealPlanIDKey,
					mealplanningkeys.MealPlanEventIDKey,
				},
			},
		)

		assert.NoError(t, mpm.UpdateMealPlanEvent(ctx, exampleMealPlanID, exampleMealPlanEvent.ID, exampleInput))

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_SwapMealPlanEvents(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fakes.BuildFakeID()
		eventIDA := fakes.BuildFakeID()
		eventIDB := fakes.BuildFakeID()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.SwapMealPlanEvents), testutils.ContextMatcher, mealPlanID, eventIDA, eventIDB).Return(nil)
				db.On(reflection.GetMethodName(mpm.db.ClearMealPlanTaskNotificationSentForEvent), testutils.ContextMatcher, eventIDA).Return(nil)
				db.On(reflection.GetMethodName(mpm.db.ClearMealPlanTaskNotificationSentForEvent), testutils.ContextMatcher, eventIDB).Return(nil)
			},
			map[string][]string{
				types.MealPlanEventUpdatedServiceEventType: {
					mealplanningkeys.MealPlanIDKey,
				},
			},
		)

		err := mpm.SwapMealPlanEvents(ctx, mealPlanID, eventIDA, eventIDB)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ArchiveMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanEvent()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.ArchiveMealPlanEvent), testutils.ContextMatcher, mealPlanID, expected.ID).Return(nil)
			},
			map[string][]string{
				types.MealPlanEventArchivedServiceEventType: {
					mealplanningkeys.MealPlanIDKey,
					mealplanningkeys.MealPlanEventIDKey,
				},
			},
		)

		err := mpm.ArchiveMealPlanEvent(ctx, mealPlanID, expected.ID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
