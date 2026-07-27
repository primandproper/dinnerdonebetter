package managers

import (
	"context"
	"testing"
	"time"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v7/filtering"

	"github.com/stretchr/testify/assert"
)

func TestMealPlanningManager_ListMealPlanEvents(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanEventsList()
		exampleMealPlanID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanEventsFunc: func(_ context.Context, mealPlanID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanEvent], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlanEvents(ctx, exampleMealPlanID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanEventsCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			CreateMealPlanEventFunc: func(_ context.Context, _ *types.MealPlanEventDatabaseCreationInput) (*types.MealPlanEvent, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanEvent(ctx, expected.BelongsToMealPlan, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanEventCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanEventFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string) (*types.MealPlanEvent, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, expected.ID, mealPlanEventID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanEvent(ctx, exampleMealPlanID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanEventCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanEventFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string) (*types.MealPlanEvent, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEvent.ID, mealPlanEventID)

				return exampleMealPlanEvent, nil
			},
			UpdateMealPlanEventFunc: func(_ context.Context, _ *types.MealPlanEvent) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		assert.NoError(t, mpm.UpdateMealPlanEvent(ctx, exampleMealPlanID, exampleMealPlanEvent.ID, exampleInput))

		assert.Len(t, db.GetMealPlanEventCalls(), 1)
		assert.Len(t, db.UpdateMealPlanEventCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanEventFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string) (*types.MealPlanEvent, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEvent.ID, mealPlanEventID)

				return exampleMealPlanEvent, nil
			},
			UpdateMealPlanEventFunc: func(_ context.Context, _ *types.MealPlanEvent) error {
				return nil
			},
			ClearMealPlanTaskNotificationSentForEventFunc: func(_ context.Context, mealPlanEventID string) error {
				assert.Equal(t, exampleMealPlanEvent.ID, mealPlanEventID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		assert.NoError(t, mpm.UpdateMealPlanEvent(ctx, exampleMealPlanID, exampleMealPlanEvent.ID, exampleInput))

		assert.Len(t, db.GetMealPlanEventCalls(), 1)
		assert.Len(t, db.UpdateMealPlanEventCalls(), 1)
		assert.Len(t, db.ClearMealPlanTaskNotificationSentForEventCalls(), 1)
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

		db := &mealplanningmock.RepositoryMock{
			SwapMealPlanEventsFunc: func(_ context.Context, actualMealPlanID string, mealPlanEventIDA string, mealPlanEventIDB string) error {
				assert.Equal(t, mealPlanID, actualMealPlanID)
				assert.Equal(t, eventIDA, mealPlanEventIDA)
				assert.Equal(t, eventIDB, mealPlanEventIDB)

				return nil
			},
			// both swapped events have their notification flag cleared.
			ClearMealPlanTaskNotificationSentForEventFunc: func(_ context.Context, mealPlanEventID string) error {
				assert.Contains(t, []string{eventIDA, eventIDB}, mealPlanEventID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.SwapMealPlanEvents(ctx, mealPlanID, eventIDA, eventIDB)
		assert.NoError(t, err)

		assert.Len(t, db.SwapMealPlanEventsCalls(), 1)
		assert.Len(t, db.ClearMealPlanTaskNotificationSentForEventCalls(), 2)
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

		db := &mealplanningmock.RepositoryMock{
			ArchiveMealPlanEventFunc: func(_ context.Context, actualMealPlanID string, mealPlanEventID string) error {
				assert.Equal(t, mealPlanID, actualMealPlanID)
				assert.Equal(t, expected.ID, mealPlanEventID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanEvent(ctx, mealPlanID, expected.ID)
		assert.NoError(t, err)

		assert.Len(t, db.ArchiveMealPlanEventCalls(), 1)
	})
}
