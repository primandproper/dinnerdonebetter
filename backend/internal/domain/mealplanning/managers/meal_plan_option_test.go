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

func TestMealPlanningManager_ListMealPlanOptions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanOptionsList()
		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			GetMealPlanOptionsFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanOption], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlanOptions(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleOwnerID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanOptionsCalls(), 1)
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

		actual, err := mpm.ListMealPlanOptions(ctx, exampleMealPlanID, fake.BuildFakeID(), exampleOwnerID, nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetMealPlanOptionsCalls())
	})
}

func TestMealPlanningManager_CreateMealPlanOptionWithEventID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOption()
		fakeInput := fakes.BuildFakeMealPlanOptionCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:      mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanEventExistsFunc: mealPlanEventExistsStub(t, exampleMealPlanID, exampleMealPlanEventID, true),
			MealExistsAsOptionInEventFunc: func(_ context.Context, mealPlanEventID string, mealID string) (bool, error) {
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, fakeInput.MealID, mealID)

				return false, nil
			},
			CreateMealPlanOptionFunc: func(_ context.Context, input *types.MealPlanOptionDatabaseCreationInput) (*types.MealPlanOption, error) {
				assert.Equal(t, exampleMealPlanEventID, input.BelongsToMealPlanEvent)
				assert.Equal(t, fakeInput.MealID, input.MealID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanOptionWithEventID(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleOwnerID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanOptionCalls(), 1)
	})

	T.Run("with inline selections", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOption()
		fakeInput := fakes.BuildFakeMealPlanOptionCreationRequestInput()
		fakeInput.Selections = []*types.MealPlanRecipeOptionSelectionCreationRequestInput{
			fakes.BuildFakeMealPlanRecipeOptionSelectionCreationRequestInput(),
			fakes.BuildFakeMealPlanRecipeOptionSelectionCreationRequestInput(),
		}

		createdSelection := fakes.BuildFakeMealPlanRecipeOptionSelection()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:      mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanEventExistsFunc: mealPlanEventExistsStub(t, exampleMealPlanID, exampleMealPlanEventID, true),
			MealExistsAsOptionInEventFunc: func(context.Context, string, string) (bool, error) {
				return false, nil
			},
			CreateMealPlanOptionFunc: func(_ context.Context, _ *types.MealPlanOptionDatabaseCreationInput) (*types.MealPlanOption, error) {
				return expected, nil
			},
			CreateMealPlanRecipeOptionSelectionFunc: func(_ context.Context, in *types.MealPlanRecipeOptionSelectionDatabaseCreationInput) (*types.MealPlanRecipeOptionSelection, error) {
				assert.Equal(t, expected.ID, in.BelongsToMealPlanOption)

				return createdSelection, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanOptionWithEventID(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleOwnerID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanOptionCalls(), 1)
		// one selection is persisted per selection on the input.
		assert.Len(t, db.CreateMealPlanRecipeOptionSelectionCalls(), len(fakeInput.Selections))
	})

	T.Run("returns ErrDuplicateMealPlanOption when MealExistsAsOptionInEvent returns true", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		eventID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeMealPlanOptionCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:      mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanEventExistsFunc: mealPlanEventExistsStub(t, exampleMealPlanID, eventID, true),
			MealExistsAsOptionInEventFunc: func(_ context.Context, mealPlanEventID string, mealID string) (bool, error) {
				assert.Equal(t, eventID, mealPlanEventID)
				assert.Equal(t, fakeInput.MealID, mealID)

				return true, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanOptionWithEventID(ctx, exampleMealPlanID, eventID, exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, types.ErrDuplicateMealPlanOption)
		assert.Nil(t, actual)

		assert.Len(t, db.MealExistsAsOptionInEventCalls(), 1)
		assert.Empty(t, db.CreateMealPlanOptionCalls())
	})

	T.Run("with a caller outside the meal plan's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeMealPlanOptionCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanOptionWithEventID(ctx, exampleMealPlanID, fake.BuildFakeID(), exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.MealPlanEventExistsCalls())
		assert.Empty(t, db.CreateMealPlanOptionCalls())
	})

	T.Run("with an event from another meal plan", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		fakeInput := fakes.BuildFakeMealPlanOptionCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:      mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanEventExistsFunc: mealPlanEventExistsStub(t, exampleMealPlanID, exampleMealPlanEventID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanOptionWithEventID(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleOwnerID, fakeInput)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.MealExistsAsOptionInEventCalls())
		assert.Empty(t, db.CreateMealPlanOptionCalls())
	})
}

func TestMealPlanningManager_ReadMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOption()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			GetMealPlanOptionFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string) (*types.MealPlanOption, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, expected.ID, mealPlanOptionID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanOption(ctx, exampleMealPlanID, exampleMealPlanEventID, expected.ID, exampleOwnerID)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

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

		actual, err := mpm.ReadMealPlanOption(ctx, exampleMealPlanID, fake.BuildFakeID(), fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetMealPlanOptionCalls())
	})
}

func TestMealPlanningManager_UpdateMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanOption := fakes.BuildFakeMealPlanOption()
		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanOptionUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc: mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			GetMealPlanOptionFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string) (*types.MealPlanOption, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, exampleMealPlanOption.ID, mealPlanOptionID)

				return exampleMealPlanOption, nil
			},
			UpdateMealPlanOptionFunc: func(_ context.Context, _ *types.MealPlanOption) error {
				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		require.NoError(t, mpm.UpdateMealPlanOption(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOption.ID, exampleOwnerID, exampleInput))

		assert.Len(t, db.GetMealPlanOptionCalls(), 1)
		assert.Len(t, db.UpdateMealPlanOptionCalls(), 1)
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

		err := mpm.UpdateMealPlanOption(ctx, exampleMealPlanID, fake.BuildFakeID(), fake.BuildFakeID(), exampleOwnerID, fakes.BuildFakeMealPlanOptionUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.GetMealPlanOptionCalls())
		assert.Empty(t, db.UpdateMealPlanOptionCalls())
	})
}

func TestMealPlanningManager_ArchiveMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fake.BuildFakeID()
		mealPlanEventID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOption()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:      mealPlanExistsStub(t, mealPlanID, exampleOwnerID, true),
			MealPlanEventExistsFunc: mealPlanEventExistsStub(t, mealPlanID, mealPlanEventID, true),
			ArchiveMealPlanOptionFunc: func(_ context.Context, actualMealPlanID string, actualMealPlanEventID string, mealPlanOptionID string) error {
				assert.Equal(t, mealPlanID, actualMealPlanID)
				assert.Equal(t, mealPlanEventID, actualMealPlanEventID)
				assert.Equal(t, expected.ID, mealPlanOptionID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanOption(ctx, mealPlanID, mealPlanEventID, expected.ID, exampleOwnerID)
		require.NoError(t, err)

		assert.Len(t, db.ArchiveMealPlanOptionCalls(), 1)
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

		err := mpm.ArchiveMealPlanOption(ctx, exampleMealPlanID, fake.BuildFakeID(), fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveMealPlanOptionCalls())
	})

	// The archive is keyed by the option and its event, so the event has to be shown to be the plan's.
	T.Run("with an event from another meal plan", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fake.BuildFakeID()
		exampleMealPlanEventID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanExistsFunc:      mealPlanExistsStub(t, exampleMealPlanID, exampleOwnerID, true),
			MealPlanEventExistsFunc: mealPlanEventExistsStub(t, exampleMealPlanID, exampleMealPlanEventID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanOption(ctx, exampleMealPlanID, exampleMealPlanEventID, fake.BuildFakeID(), exampleOwnerID)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveMealPlanOptionCalls())
	})
}
