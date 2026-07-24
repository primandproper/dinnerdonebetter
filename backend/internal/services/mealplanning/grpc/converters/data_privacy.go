package grpcconverters

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
)

// ConvertUserDataCollectionToGRPCDataCollection converts a domain mealplanning UserDataCollection to a proto DataCollection.
func ConvertUserDataCollectionToGRPCDataCollection(input *mealplanning.UserDataCollection) *mealplanningsvc.DataCollection {
	result := &mealplanningsvc.DataCollection{}

	for i := range input.AccountInstrumentOwnerships {
		result.AccountInstrumentOwnerships = append(result.AccountInstrumentOwnerships, ConvertAccountInstrumentOwnershipToGRPCAccountInstrumentOwnership(&input.AccountInstrumentOwnerships[i]))
	}

	for i := range input.MealPlans {
		result.MealPlans = append(result.MealPlans, ConvertMealPlanToGRPCMealPlan(&input.MealPlans[i]))
	}

	for i := range input.RecipeRatings {
		result.RecipeRatings = append(result.RecipeRatings, ConvertRecipeRatingToGRPCRecipeRating(&input.RecipeRatings[i]))
	}

	for i := range input.Recipes {
		result.Recipes = append(result.Recipes, ConvertRecipeToGRPCRecipe(&input.Recipes[i]))
	}

	for i := range input.Meals {
		result.Meals = append(result.Meals, ConvertMealToGRPCMeal(&input.Meals[i]))
	}

	for i := range input.UserIngredientPreferences {
		result.UserIngredientPreferences = append(result.UserIngredientPreferences, ConvertUserIngredientPreferenceToGRPCUserIngredientPreference(&input.UserIngredientPreferences[i]))
	}

	return result
}
