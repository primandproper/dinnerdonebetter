package managers

import (
	"context"
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	"github.com/primandproper/platform-go/v9/filtering"

	"github.com/stretchr/testify/assert"
)

func TestMealPlanningManager_ListMealPlanOptions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanOptionsList()
		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanOptionsFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanOption], error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ListMealPlanOptions(ctx, exampleMealPlanID, exampleMealPlanEventID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanOptionsCalls(), 1)
	})
}

func TestMealPlanningManager_CreateMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanOption()
		fakeInput := fakes.BuildFakeMealPlanOptionCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			CreateMealPlanOptionFunc: func(_ context.Context, _ *types.MealPlanOptionDatabaseCreationInput) (*types.MealPlanOption, error) {
				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanOption(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanOptionCalls(), 1)
	})

	T.Run("with inline selections", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanOption()
		fakeInput := fakes.BuildFakeMealPlanOptionCreationRequestInput()
		fakeInput.Selections = []*types.MealPlanRecipeOptionSelectionCreationRequestInput{
			fakes.BuildFakeMealPlanRecipeOptionSelectionCreationRequestInput(),
			fakes.BuildFakeMealPlanRecipeOptionSelectionCreationRequestInput(),
		}

		createdSelection := fakes.BuildFakeMealPlanRecipeOptionSelection()

		db := &mealplanningmock.RepositoryMock{
			CreateMealPlanOptionFunc: func(_ context.Context, _ *types.MealPlanOptionDatabaseCreationInput) (*types.MealPlanOption, error) {
				return expected, nil
			},
			CreateMealPlanRecipeOptionSelectionFunc: func(_ context.Context, in *types.MealPlanRecipeOptionSelectionDatabaseCreationInput) (*types.MealPlanRecipeOptionSelection, error) {
				assert.Equal(t, expected.ID, in.BelongsToMealPlanOption)

				return createdSelection, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanOption(ctx, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanOptionCalls(), 1)
		// one selection is persisted per selection on the input.
		assert.Len(t, db.CreateMealPlanRecipeOptionSelectionCalls(), len(fakeInput.Selections))
	})
}

func TestMealPlanningManager_CreateMealPlanOptionWithEventID(T *testing.T) {
	T.Parallel()

	T.Run("returns ErrDuplicateMealPlanOption when MealExistsAsOptionInEvent returns true", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		eventID := fakes.BuildFakeID()
		fakeInput := fakes.BuildFakeMealPlanOptionCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealExistsAsOptionInEventFunc: func(_ context.Context, mealPlanEventID string, mealID string) (bool, error) {
				assert.Equal(t, eventID, mealPlanEventID)
				assert.Equal(t, fakeInput.MealID, mealID)

				return true, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanOptionWithEventID(ctx, eventID, fakeInput)
		assert.ErrorIs(t, err, types.ErrDuplicateMealPlanOption)
		assert.Nil(t, actual)

		assert.Len(t, db.MealExistsAsOptionInEventCalls(), 1)
	})
}

func TestMealPlanningManager_ReadMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOption()

		db := &mealplanningmock.RepositoryMock{
			GetMealPlanOptionFunc: func(_ context.Context, mealPlanID string, mealPlanEventID string, mealPlanOptionID string) (*types.MealPlanOption, error) {
				assert.Equal(t, exampleMealPlanID, mealPlanID)
				assert.Equal(t, exampleMealPlanEventID, mealPlanEventID)
				assert.Equal(t, expected.ID, mealPlanOptionID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.ReadMealPlanOption(ctx, exampleMealPlanID, exampleMealPlanEventID, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanOptionCalls(), 1)
	})
}

func TestMealPlanningManager_MealPlanOptionBelongsToAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanOptionID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanOptionBelongsToAccountFunc: func(_ context.Context, mealPlanOptionID string, accountID string) (bool, error) {
				assert.Equal(t, exampleMealPlanOptionID, mealPlanOptionID)
				assert.Equal(t, exampleAccountID, accountID)

				return true, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		belongs, err := mpm.MealPlanOptionBelongsToAccount(ctx, exampleMealPlanOptionID, exampleAccountID)
		assert.NoError(t, err)
		assert.True(t, belongs)

		assert.Len(t, db.MealPlanOptionBelongsToAccountCalls(), 1)
	})

	T.Run("with option belonging to another account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanOptionID := fakes.BuildFakeID()
		exampleAccountID := fakes.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanOptionBelongsToAccountFunc: func(_ context.Context, mealPlanOptionID string, accountID string) (bool, error) {
				assert.Equal(t, exampleMealPlanOptionID, mealPlanOptionID)
				assert.Equal(t, exampleAccountID, accountID)

				return false, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		belongs, err := mpm.MealPlanOptionBelongsToAccount(ctx, exampleMealPlanOptionID, exampleAccountID)
		assert.NoError(t, err)
		assert.False(t, belongs)

		assert.Len(t, db.MealPlanOptionBelongsToAccountCalls(), 1)
	})
}

func TestMealPlanningManager_UpdateMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		exampleMealPlanOption := fakes.BuildFakeMealPlanOption()
		exampleMealPlanID := fakes.BuildFakeID()
		exampleMealPlanEventID := fakes.BuildFakeID()
		exampleInput := fakes.BuildFakeMealPlanOptionUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
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

		assert.NoError(t, mpm.UpdateMealPlanOption(ctx, exampleMealPlanID, exampleMealPlanEventID, exampleMealPlanOption.ID, exampleInput))

		assert.Len(t, db.GetMealPlanOptionCalls(), 1)
		assert.Len(t, db.UpdateMealPlanOptionCalls(), 1)
	})
}

func TestMealPlanningManager_ArchiveMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanID := fakes.BuildFakeID()
		mealPlanEventID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanOption()

		db := &mealplanningmock.RepositoryMock{
			ArchiveMealPlanOptionFunc: func(_ context.Context, actualMealPlanID string, actualMealPlanEventID string, mealPlanOptionID string) error {
				assert.Equal(t, mealPlanID, actualMealPlanID)
				assert.Equal(t, mealPlanEventID, actualMealPlanEventID)
				assert.Equal(t, expected.ID, mealPlanOptionID)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanOption(ctx, mealPlanID, mealPlanEventID, expected.ID)
		assert.NoError(t, err)

		assert.Len(t, db.ArchiveMealPlanOptionCalls(), 1)
	})
}
