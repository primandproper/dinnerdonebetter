package grpcconverters

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	fakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertUserDataCollectionToGRPCDataCollection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		input := &mealplanning.UserDataCollection{
			Recipes:                     []mealplanning.Recipe{*fakes.BuildFakeRecipe()},
			MealPlans:                   []mealplanning.MealPlan{*fakes.BuildFakeMealPlan()},
			Meals:                       []mealplanning.Meal{*fakes.BuildFakeMeal()},
			UserIngredientPreferences:   []mealplanning.UserIngredientPreference{*fakes.BuildFakeUserIngredientPreference()},
			AccountInstrumentOwnerships: []mealplanning.AccountInstrumentOwnership{*fakes.BuildFakeAccountInstrumentOwnership()},
			RecipeRatings:               []mealplanning.RecipeRating{*fakes.BuildFakeRecipeRating()},
		}

		result := ConvertUserDataCollectionToGRPCDataCollection(input)

		require.NotNil(t, result)
		assert.Len(t, result.Recipes, 1)
		assert.Len(t, result.MealPlans, 1)
		assert.Len(t, result.Meals, 1)
		assert.Len(t, result.UserIngredientPreferences, 1)
		assert.Len(t, result.AccountInstrumentOwnerships, 1)
		assert.Len(t, result.RecipeRatings, 1)
	})

	T.Run("empty", func(t *testing.T) {
		t.Parallel()

		result := ConvertUserDataCollectionToGRPCDataCollection(&mealplanning.UserDataCollection{})

		require.NotNil(t, result)
		assert.Empty(t, result.Recipes)
		assert.Empty(t, result.MealPlans)
	})
}
