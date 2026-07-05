package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v3/database/filtering"
)

// BuildFakeRecipeRating builds a faked valid ingredient.
func BuildFakeRecipeRating() *types.RecipeRating {
	return &types.RecipeRating{
		CreatedAt:       BuildFakeTime(),
		Notes:           buildUniqueString(),
		ID:              BuildFakeID(),
		BelongsToRecipe: BuildFakeID(),
		CreatedByUser:   BuildFakeID(),
		Taste:           float32(buildFakeNumber()),
		Instructions:    float32(buildFakeNumber()),
		Overall:         float32(buildFakeNumber()),
		Cleanup:         float32(buildFakeNumber()),
		Difficulty:      float32(buildFakeNumber()),
	}
}

// BuildFakeRecipeRatingsList builds a faked RecipeRatingList.
func BuildFakeRecipeRatingsList() *filtering.QueryFilteredResult[types.RecipeRating] {
	var examples []*types.RecipeRating
	for range exampleQuantity {
		examples = append(examples, BuildFakeRecipeRating())
	}

	return &filtering.QueryFilteredResult[types.RecipeRating]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeRecipeRatingUpdateRequestInput builds a faked RecipeRatingUpdateRequestInput from a valid ingredient.
func BuildFakeRecipeRatingUpdateRequestInput() *types.RecipeRatingUpdateRequestInput {
	recipeRating := BuildFakeRecipeRating()
	return converters.ConvertRecipeRatingToRecipeRatingUpdateRequestInput(recipeRating)
}

// BuildFakeRecipeRatingCreationRequestInput builds a faked RecipeRatingCreationRequestInput.
func BuildFakeRecipeRatingCreationRequestInput() *types.RecipeRatingCreationRequestInput {
	recipeRating := BuildFakeRecipeRating()
	return converters.ConvertRecipeRatingToRecipeRatingCreationRequestInput(recipeRating)
}
