package managers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMealPlanningManager_ListMealPlanEvents(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanEventsList()
		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			GetMealPlanEventsFunc: func(_ context.Context, mealPlanID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanEvent], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlanEvents(ctx, exampleMealPlanID, exampleOwnerID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanEventsCalls(), 1)
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

		actual, err := mpm.ListMealPlanEvents(ctx, exampleMealPlanID, exampleOwnerID, nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetMealPlanEventsCalls())
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
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, expected.BelongsToMealPlan, exampleOwnerID, true),
			CreateMealPlanEventFunc: func(_ context.Context, input *types.MealPlanEventDatabaseCreationInput) (*types.MealPlanEvent, error) {
				assert.Equal(t, expected.BelongsToMealPlan, input.BelongsToMealPlan)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanEvent(ctx, expected.BelongsToMealPlan, exampleOwnerID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanEventCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeMealPlanEventCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanEvent(ctx, exampleMealPlanID, exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.CreateMealPlanEventCalls())
	})
}

func TestMealPlanningManager_ReadMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanEvent()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			GetMealPlanEventFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string) (*types.MealPlanEvent, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, expected.ID, mealPlanEventID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanEvent(ctx, exampleMealPlanID, expected.ID, exampleOwnerID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanEventCalls(), 1)
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

		actual, err := mpm.ReadMealPlanEvent(ctx, exampleMealPlanID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetMealPlanEventCalls())
	})
}

func TestMealPlanningManager_UpdateMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanEvent := fakes.BuildFakeMealPlanEvent()
		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanEventUpdateRequestInput()
		exampleInput.StartsAt = &exampleMealPlanEvent.StartsAt

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
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

		require.NoError(t, mpm.UpdateMealPlanEvent(ctx, exampleMealPlanID, exampleMealPlanEvent.ID, exampleOwnerID, exampleInput))

		assert.Len(t, db.GetMealPlanEventCalls(), 1)
		assert.Len(t, db.UpdateMealPlanEventCalls(), 1)
	})

	T.Run("when start time changes clears notification sent for event", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanEvent := fakes.BuildFakeMealPlanEvent()
		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanEventUpdateRequestInput()
		newStartsAt := exampleMealPlanEvent.StartsAt.Add(time.Hour)
		exampleInput.StartsAt = &newStartsAt

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
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

		require.NoError(t, mpm.UpdateMealPlanEvent(ctx, exampleMealPlanID, exampleMealPlanEvent.ID, exampleOwnerID, exampleInput))

		assert.Len(t, db.GetMealPlanEventCalls(), 1)
		assert.Len(t, db.UpdateMealPlanEventCalls(), 1)
		assert.Len(t, db.ClearMealPlanTaskNotificationSentForEventCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanEventUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.UpdateMealPlanEvent(ctx, exampleMealPlanID, fake.BuildFakeID(), exampleOwnerID, exampleInput)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.GetMealPlanEventCalls())
		assert.Empty(t, db.UpdateMealPlanEventCalls())
	})
}

func TestMealPlanningManager_SwapMealPlanEvents(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fake.BuildFakeID()
		eventIDA := fake.BuildFakeID()
		eventIDB := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, mealPlanID, exampleOwnerID, true),
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

		err := mpm.SwapMealPlanEvents(ctx, mealPlanID, eventIDA, eventIDB, exampleOwnerID)
		require.NoError(t, err)

		assert.Len(t, db.SwapMealPlanEventsCalls(), 1)
		assert.Len(t, db.ClearMealPlanTaskNotificationSentForEventCalls(), 2)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, mealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.SwapMealPlanEvents(ctx, mealPlanID, fake.BuildFakeID(), fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.SwapMealPlanEventsCalls())
		assert.Empty(t, db.ClearMealPlanTaskNotificationSentForEventCalls())
	})
}

func TestMealPlanningManager_ArchiveMealPlanEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanEvent()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, mealPlanID, exampleOwnerID, true),
			ArchiveMealPlanEventFunc: func(_ context.Context, actualMealPlanID string, mealPlanEventID string) error {
				assert.Equal(t, mealPlanID, actualMealPlanID)
				assert.Equal(t, expected.ID, mealPlanEventID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanEvent(ctx, mealPlanID, expected.ID, exampleOwnerID)
		require.NoError(t, err)

		assert.Len(t, db.ArchiveMealPlanEventCalls(), 1)
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, mealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanEvent(ctx, mealPlanID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveMealPlanEventCalls())
	})
}
