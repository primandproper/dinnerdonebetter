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

func TestMealPlanningManager_GetMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		recipeStepID := fake.BuildFakeID()
		ingredientIndex := uint16(0)
		selectionType := types.MealPlanRecipeOptionSelectionTypeIngredient
		expected := fakes.BuildFakeMealPlanRecipeOptionSelection()

		db := &mealplanningmock.RepositoryMock{
			MealPlanOptionBelongsToAccountFunc: mealPlanOptionBelongsToAccountStub(t, mealPlanOptionID, exampleOwnerID, true),
			GetMealPlanRecipeOptionSelectionFunc: func(_ context.Context, actualMealPlanOptionID string, actualRecipeStepID string, actualIngredientIndex uint16, actualSelectionType string) (*types.MealPlanRecipeOptionSelection, error) {
				assert.Equal(t, mealPlanOptionID, actualMealPlanOptionID)
				assert.Equal(t, recipeStepID, actualRecipeStepID)
				assert.Equal(t, ingredientIndex, actualIngredientIndex)
				assert.Equal(t, selectionType, actualSelectionType)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.GetMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, exampleOwnerID, ingredientIndex, selectionType)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetMealPlanRecipeOptionSelectionCalls(), 1)
	})

	T.Run("with a caller outside the meal plan option's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanOptionBelongsToAccountFunc: mealPlanOptionBelongsToAccountStub(t, mealPlanOptionID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.GetMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, fake.BuildFakeID(), exampleOwnerID, 0, types.MealPlanRecipeOptionSelectionTypeIngredient)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetMealPlanRecipeOptionSelectionCalls())
	})
}

func TestMealPlanningManager_GetMealPlanRecipeOptionSelectionsForMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanRecipeOptionSelectionsList()
		mealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanOptionBelongsToAccountFunc: mealPlanOptionBelongsToAccountStub(t, mealPlanOptionID, exampleOwnerID, true),
			GetSelectionsForMealPlanOptionFunc: func(_ context.Context, actualMealPlanOptionID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanRecipeOptionSelection], error) {
				assert.Equal(t, mealPlanOptionID, actualMealPlanOptionID)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.GetMealPlanRecipeOptionSelectionsForMealPlanOption(ctx, mealPlanOptionID, exampleOwnerID, nil)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.GetSelectionsForMealPlanOptionCalls(), 1)
	})

	T.Run("with a caller outside the meal plan option's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanOptionBelongsToAccountFunc: mealPlanOptionBelongsToAccountStub(t, mealPlanOptionID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.GetMealPlanRecipeOptionSelectionsForMealPlanOption(ctx, mealPlanOptionID, exampleOwnerID, nil)
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.GetSelectionsForMealPlanOptionCalls())
	})
}

func TestMealPlanningManager_CreateMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		expected := fakes.BuildFakeMealPlanRecipeOptionSelection()
		fakeInput := fakes.BuildFakeMealPlanRecipeOptionSelectionCreationRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanOptionBelongsToAccountFunc: mealPlanOptionBelongsToAccountStub(t, mealPlanOptionID, exampleOwnerID, true),
			CreateMealPlanRecipeOptionSelectionFunc: func(_ context.Context, input *types.MealPlanRecipeOptionSelectionDatabaseCreationInput) (*types.MealPlanRecipeOptionSelection, error) {
				assert.Equal(t, mealPlanOptionID, input.BelongsToMealPlanOption)

				return expected, nil
			},
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, exampleOwnerID, fakeInput)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, db.CreateMealPlanRecipeOptionSelectionCalls(), 1)
	})

	T.Run("with a caller outside the meal plan option's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanOptionBelongsToAccountFunc: mealPlanOptionBelongsToAccountStub(t, mealPlanOptionID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		actual, err := mpm.CreateMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, exampleOwnerID, fakes.BuildFakeMealPlanRecipeOptionSelectionCreationRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, actual)

		assert.Empty(t, db.CreateMealPlanRecipeOptionSelectionCalls())
	})
}

func TestMealPlanningManager_UpdateMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		existing := fakes.BuildFakeMealPlanRecipeOptionSelection()
		mealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		recipeStepID := fake.BuildFakeID()
		ingredientIndex := uint16(0)
		selectionType := types.MealPlanRecipeOptionSelectionTypeIngredient
		fakeInput := fakes.BuildFakeMealPlanRecipeOptionSelectionUpdateRequestInput()

		db := &mealplanningmock.RepositoryMock{
			MealPlanOptionBelongsToAccountFunc: mealPlanOptionBelongsToAccountStub(t, mealPlanOptionID, exampleOwnerID, true),
			GetMealPlanRecipeOptionSelectionFunc: func(_ context.Context, actualMealPlanOptionID string, actualRecipeStepID string, actualIngredientIndex uint16, actualSelectionType string) (*types.MealPlanRecipeOptionSelection, error) {
				assert.Equal(t, mealPlanOptionID, actualMealPlanOptionID)
				assert.Equal(t, recipeStepID, actualRecipeStepID)
				assert.Equal(t, ingredientIndex, actualIngredientIndex)
				assert.Equal(t, selectionType, actualSelectionType)

				return existing, nil
			},
			UpdateMealPlanRecipeOptionSelectionFunc: func(_ context.Context, actualMealPlanOptionID string, actualRecipeStepID string, actualIngredientIndex uint16, actualSelectionType string, _ *types.MealPlanRecipeOptionSelectionUpdateRequestInput) error {
				assert.Equal(t, mealPlanOptionID, actualMealPlanOptionID)
				assert.Equal(t, recipeStepID, actualRecipeStepID)
				assert.Equal(t, ingredientIndex, actualIngredientIndex)
				assert.Equal(t, selectionType, actualSelectionType)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.UpdateMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, exampleOwnerID, ingredientIndex, selectionType, fakeInput)
		require.NoError(t, err)

		assert.Len(t, db.GetMealPlanRecipeOptionSelectionCalls(), 1)
		assert.Len(t, db.UpdateMealPlanRecipeOptionSelectionCalls(), 1)
	})

	T.Run("with a caller outside the meal plan option's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanOptionBelongsToAccountFunc: mealPlanOptionBelongsToAccountStub(t, mealPlanOptionID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.UpdateMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, fake.BuildFakeID(), exampleOwnerID, 0, types.MealPlanRecipeOptionSelectionTypeIngredient, fakes.BuildFakeMealPlanRecipeOptionSelectionUpdateRequestInput())
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.GetMealPlanRecipeOptionSelectionCalls())
		assert.Empty(t, db.UpdateMealPlanRecipeOptionSelectionCalls())
	})
}

func TestMealPlanningManager_ArchiveMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()
		recipeStepID := fake.BuildFakeID()
		ingredientIndex := uint16(0)
		selectionType := types.MealPlanRecipeOptionSelectionTypeIngredient

		db := &mealplanningmock.RepositoryMock{
			MealPlanOptionBelongsToAccountFunc: mealPlanOptionBelongsToAccountStub(t, mealPlanOptionID, exampleOwnerID, true),
			ArchiveMealPlanRecipeOptionSelectionFunc: func(_ context.Context, actualMealPlanOptionID string, actualRecipeStepID string, actualIngredientIndex uint16, actualSelectionType string) error {
				assert.Equal(t, mealPlanOptionID, actualMealPlanOptionID)
				assert.Equal(t, recipeStepID, actualRecipeStepID)
				assert.Equal(t, ingredientIndex, actualIngredientIndex)
				assert.Equal(t, selectionType, actualSelectionType)

				return nil
			},
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, exampleOwnerID, ingredientIndex, selectionType)
		require.NoError(t, err)

		assert.Len(t, db.ArchiveMealPlanRecipeOptionSelectionCalls(), 1)
	})

	T.Run("with a caller outside the meal plan option's account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fake.BuildFakeID()
		exampleOwnerID := fake.BuildFakeID()

		db := &mealplanningmock.RepositoryMock{
			MealPlanOptionBelongsToAccountFunc: mealPlanOptionBelongsToAccountStub(t, mealPlanOptionID, exampleOwnerID, false),
		}
		attachRepositoryToManager(mpm, db)

		err := mpm.ArchiveMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, fake.BuildFakeID(), exampleOwnerID, 0, types.MealPlanRecipeOptionSelectionTypeIngredient)
		require.ErrorIs(t, err, sql.ErrNoRows)

		assert.Empty(t, db.ArchiveMealPlanRecipeOptionSelectionCalls())
	})
}
