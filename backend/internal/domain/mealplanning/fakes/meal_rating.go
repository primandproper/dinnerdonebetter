package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
)

// BuildFakeRecipeRating builds a faked recipe rating.
func BuildFakeRecipeRating() *types.RecipeRating {
	return fake.BuildFakeRecord[types.RecipeRating]()
}

// BuildFakeRecipeRatingsList builds a faked RecipeRatingList.
func BuildFakeRecipeRatingsList() *filtering.QueryFilteredResult[types.RecipeRating] {
	return fake.BuildFakePage(BuildFakeRecipeRating)
}

// BuildFakeRecipeRatingUpdateRequestInput builds a faked RecipeRatingUpdateRequestInput from a recipe rating.
func BuildFakeRecipeRatingUpdateRequestInput() *types.RecipeRatingUpdateRequestInput {
	recipeRating := BuildFakeRecipeRating()

	return converters.ConvertRecipeRatingToRecipeRatingUpdateRequestInput(recipeRating)
}

// BuildFakeRecipeRatingCreationRequestInput builds a faked RecipeRatingCreationRequestInput.
func BuildFakeRecipeRatingCreationRequestInput() *types.RecipeRatingCreationRequestInput {
	recipeRating := BuildFakeRecipeRating()

	return converters.ConvertRecipeRatingToRecipeRatingCreationRequestInput(recipeRating)
}
