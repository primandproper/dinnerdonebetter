package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v5/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeRecipeStepCompletionCondition builds a faked recipe step ingredient.
// NOTE: this currently represents a typical recipe step ingredient with a valid ingredient and not a product.
func BuildFakeRecipeStepCompletionCondition() *types.RecipeStepCompletionCondition {
	id := BuildFakeID()
	var ingredients []*types.RecipeStepCompletionConditionIngredient
	for range exampleQuantity {
		ingredient := BuildFakeRecipeStepCompletionConditionIngredient()
		ingredient.BelongsToRecipeStepCompletionCondition = id
		ingredients = append(ingredients, ingredient)
	}

	return &types.RecipeStepCompletionCondition{
		Optional:            fake.Bool(),
		IngredientState:     *BuildFakeValidIngredientState(),
		ID:                  id,
		BelongsToRecipeStep: BuildFakeID(),
		Notes:               buildUniqueString(),
		Ingredients:         ingredients,
	}
}

// BuildFakeRecipeStepCompletionConditionIngredient builds a faked recipe step ingredient.
// NOTE: this currently represents a typical recipe step ingredient with a valid ingredient and not a product.
func BuildFakeRecipeStepCompletionConditionIngredient() *types.RecipeStepCompletionConditionIngredient {
	return &types.RecipeStepCompletionConditionIngredient{
		ID:                                     BuildFakeID(),
		BelongsToRecipeStepCompletionCondition: BuildFakeID(),
		RecipeStepIngredient:                   BuildFakeID(),
	}
}

// BuildFakeRecipeStepCompletionConditionsList builds a faked RecipeStepCompletionConditionList.
func BuildFakeRecipeStepCompletionConditionsList() *filtering.QueryFilteredResult[types.RecipeStepCompletionCondition] {
	var examples []*types.RecipeStepCompletionCondition
	for range exampleQuantity {
		examples = append(examples, BuildFakeRecipeStepCompletionCondition())
	}

	return &filtering.QueryFilteredResult[types.RecipeStepCompletionCondition]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeRecipeStepCompletionConditionUpdateRequestInput builds a faked RecipeStepCompletionConditionUpdateRequestInput from a recipe step ingredient.
func BuildFakeRecipeStepCompletionConditionUpdateRequestInput() *types.RecipeStepCompletionConditionUpdateRequestInput {
	recipeStepIngredient := BuildFakeRecipeStepCompletionCondition()
	return &types.RecipeStepCompletionConditionUpdateRequestInput{
		Optional:            &recipeStepIngredient.Optional,
		BelongsToRecipeStep: &recipeStepIngredient.BelongsToRecipeStep,
		IngredientStateID:   &recipeStepIngredient.IngredientState.ID,
		Notes:               &recipeStepIngredient.Notes,
	}
}

// BuildFakeRecipeStepCompletionConditionForExistingRecipeCreationRequestInput builds a faked BuildFakeRecipeStepCompletionConditionForExistingRecipeCreationRequestInput.
func BuildFakeRecipeStepCompletionConditionForExistingRecipeCreationRequestInput() *types.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput {
	recipeStepIngredient := BuildFakeRecipeStepCompletionCondition()
	return converters.ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionForExistingRecipeCreationRequestInput(recipeStepIngredient)
}
