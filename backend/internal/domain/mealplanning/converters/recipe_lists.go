package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v10/identifiers"
)

func ConvertRecipeListItemCreationRequestInputToRecipeListItemDatabaseCreationInput(x *mealplanning.RecipeListItemCreationRequestInput, recipeListID string) *mealplanning.RecipeListItemDatabaseCreationInput {
	return &mealplanning.RecipeListItemDatabaseCreationInput{
		ID:                  identifiers.New(),
		RecipeID:            x.RecipeID,
		Notes:               x.Notes,
		BelongsToRecipeList: recipeListID,
	}
}
