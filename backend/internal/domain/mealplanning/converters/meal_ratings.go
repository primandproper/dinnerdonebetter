package converters

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v8/identifiers"
)

// ConvertRecipeRatingToRecipeRatingUpdateRequestInput creates a DatabaseCreationInput from a CreationInput.
func ConvertRecipeRatingToRecipeRatingUpdateRequestInput(x *types.RecipeRating) *types.RecipeRatingUpdateRequestInput {
	out := &types.RecipeRatingUpdateRequestInput{
		BelongsToRecipe: &x.BelongsToRecipe,
		Taste:           &x.Taste,
		Difficulty:      &x.Difficulty,
		Cleanup:         &x.Cleanup,
		Instructions:    &x.Instructions,
		Overall:         &x.Overall,
		Notes:           &x.Notes,
	}

	return out
}

// ConvertRecipeRatingCreationRequestInputToRecipeRatingDatabaseCreationInput creates a DatabaseCreationInput from a CreationInput.
func ConvertRecipeRatingCreationRequestInputToRecipeRatingDatabaseCreationInput(x *types.RecipeRatingCreationRequestInput) *types.RecipeRatingDatabaseCreationInput {
	out := &types.RecipeRatingDatabaseCreationInput{
		ID:              identifiers.New(),
		BelongsToRecipe: x.BelongsToRecipe,
		Notes:           x.Notes,
		CreatedByUser:   x.CreatedByUser,
		Taste:           x.Taste,
		Difficulty:      x.Difficulty,
		Cleanup:         x.Cleanup,
		Instructions:    x.Instructions,
		Overall:         x.Overall,
	}

	return out
}

// ConvertRecipeRatingToRecipeRatingCreationRequestInput builds a RecipeRatingCreationRequestInput from a Ingredient.
func ConvertRecipeRatingToRecipeRatingCreationRequestInput(x *types.RecipeRating) *types.RecipeRatingCreationRequestInput {
	return &types.RecipeRatingCreationRequestInput{
		BelongsToRecipe: x.BelongsToRecipe,
		Notes:           x.Notes,
		CreatedByUser:   x.CreatedByUser,
		Taste:           x.Taste,
		Difficulty:      x.Difficulty,
		Cleanup:         x.Cleanup,
		Instructions:    x.Instructions,
		Overall:         x.Overall,
	}
}

// ConvertRecipeRatingToRecipeRatingDatabaseCreationInput builds a RecipeRatingDatabaseCreationInput from a RecipeRating.
func ConvertRecipeRatingToRecipeRatingDatabaseCreationInput(x *types.RecipeRating) *types.RecipeRatingDatabaseCreationInput {
	return &types.RecipeRatingDatabaseCreationInput{
		ID:              x.ID,
		BelongsToRecipe: x.BelongsToRecipe,
		Notes:           x.Notes,
		CreatedByUser:   x.CreatedByUser,
		Taste:           x.Taste,
		Difficulty:      x.Difficulty,
		Cleanup:         x.Cleanup,
		Instructions:    x.Instructions,
		Overall:         x.Overall,
	}
}
