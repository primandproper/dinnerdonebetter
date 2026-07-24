package managers

import (
	"testing"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v6/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMealPlanningManager_GetMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fakes.BuildFakeID()
		recipeStepID := fakes.BuildFakeID()
		ingredientIndex := uint16(0)
		selectionType := types.MealPlanRecipeOptionSelectionTypeIngredient
		expected := fakes.BuildFakeMealPlanRecipeOptionSelection()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanRecipeOptionSelection), testutils.ContextMatcher, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType).Return(expected, nil)
			},
		)

		actual, err := mpm.GetMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_GetMealPlanRecipeOptionSelectionsForMealPlanOption(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		expected := fakes.BuildFakeMealPlanRecipeOptionSelectionsList()
		mealPlanOptionID := fakes.BuildFakeID()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetSelectionsForMealPlanOption), testutils.ContextMatcher, mealPlanOptionID, testutils.QueryFilterMatcher).Return(expected, nil)
			},
		)

		actual, err := mpm.GetMealPlanRecipeOptionSelectionsForMealPlanOption(ctx, mealPlanOptionID, nil)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_CreateMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fakes.BuildFakeID()
		expected := fakes.BuildFakeMealPlanRecipeOptionSelection()
		fakeInput := fakes.BuildFakeMealPlanRecipeOptionSelectionCreationRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.CreateMealPlanRecipeOptionSelection), testutils.ContextMatcher, testutils.MatchType[*types.MealPlanRecipeOptionSelectionDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				types.MealPlanRecipeOptionSelectionCreatedServiceEventType: {"meal_plan_recipe_option_selection_id", mealplanningkeys.MealPlanOptionIDKey},
			},
		)

		actual, err := mpm.CreateMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, fakeInput)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_UpdateMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		existing := fakes.BuildFakeMealPlanRecipeOptionSelection()
		mealPlanOptionID := fakes.BuildFakeID()
		recipeStepID := fakes.BuildFakeID()
		ingredientIndex := uint16(0)
		selectionType := types.MealPlanRecipeOptionSelectionTypeIngredient
		fakeInput := fakes.BuildFakeMealPlanRecipeOptionSelectionUpdateRequestInput()

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.GetMealPlanRecipeOptionSelection), testutils.ContextMatcher, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType).Return(existing, nil)
				db.On(reflection.GetMethodName(mpm.db.UpdateMealPlanRecipeOptionSelection), testutils.ContextMatcher, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType, testutils.MatchType[*types.MealPlanRecipeOptionSelectionUpdateRequestInput]()).Return(nil)
			},
			map[string][]string{
				types.MealPlanRecipeOptionSelectionUpdatedServiceEventType: {"meal_plan_recipe_option_selection_id", mealplanningkeys.MealPlanOptionIDKey},
			},
		)

		err := mpm.UpdateMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType, fakeInput)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestMealPlanningManager_ArchiveMealPlanRecipeOptionSelection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		mpm := buildMealPlanManagerForTest(t)

		mealPlanOptionID := fakes.BuildFakeID()
		recipeStepID := fakes.BuildFakeID()
		ingredientIndex := uint16(0)
		selectionType := types.MealPlanRecipeOptionSelectionTypeIngredient

		expectations := setupExpectationsForMealPlanningManager(
			mpm,
			func(db *mealplanningmock.Repository) {
				db.On(reflection.GetMethodName(mpm.db.ArchiveMealPlanRecipeOptionSelection), testutils.ContextMatcher, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType).Return(nil)
			},
			map[string][]string{
				types.MealPlanRecipeOptionSelectionArchivedServiceEventType: {
					mealplanningkeys.MealPlanOptionIDKey,
					"recipe_step_id",
					"ingredient_index",
					"selection_type",
				},
			},
		)

		err := mpm.ArchiveMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
