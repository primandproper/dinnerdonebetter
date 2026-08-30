package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
)

// BuildFakeRecipeStepCompletionCondition builds a faked recipe step completion condition.
func BuildFakeRecipeStepCompletionCondition() *types.RecipeStepCompletionCondition {
	condition := fake.BuildFakeRecord[types.RecipeStepCompletionCondition]()

	// The state the ingredients have to reach for the step to be done.
	condition.IngredientState = *BuildFakeValidIngredientState()

	// Ingredients of this condition rather than of three unrelated ones.
	ingredients := make([]*types.RecipeStepCompletionConditionIngredient, 0, exampleQuantity)
	for range exampleQuantity {
		ingredient := BuildFakeRecipeStepCompletionConditionIngredient()
		ingredient.BelongsToRecipeStepCompletionCondition = condition.ID
		ingredients = append(ingredients, ingredient)
	}
	condition.Ingredients = ingredients

	return condition
}

// BuildFakeRecipeStepCompletionConditionIngredient builds a faked recipe step completion condition ingredient.
func BuildFakeRecipeStepCompletionConditionIngredient() *types.RecipeStepCompletionConditionIngredient {
	return fake.BuildFakeRecord[types.RecipeStepCompletionConditionIngredient]()
}

// BuildFakeRecipeStepCompletionConditionsList builds a faked RecipeStepCompletionConditionList.
func BuildFakeRecipeStepCompletionConditionsList() *filtering.QueryFilteredResult[types.RecipeStepCompletionCondition] {
	return fake.BuildFakePage(BuildFakeRecipeStepCompletionCondition)
}

// BuildFakeRecipeStepCompletionConditionUpdateRequestInput builds a faked RecipeStepCompletionConditionUpdateRequestInput from a completion condition.
func BuildFakeRecipeStepCompletionConditionUpdateRequestInput() *types.RecipeStepCompletionConditionUpdateRequestInput {
	condition := BuildFakeRecipeStepCompletionCondition()

	return &types.RecipeStepCompletionConditionUpdateRequestInput{
		Optional:            &condition.Optional,
		BelongsToRecipeStep: &condition.BelongsToRecipeStep,
		IngredientStateID:   &condition.IngredientState.ID,
		Notes:               &condition.Notes,
	}
}

// BuildFakeRecipeStepCompletionConditionForExistingRecipeCreationRequestInput builds a faked RecipeStepCompletionConditionForExistingRecipeCreationRequestInput.
func BuildFakeRecipeStepCompletionConditionForExistingRecipeCreationRequestInput() *types.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput {
	condition := BuildFakeRecipeStepCompletionCondition()

	return converters.ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionForExistingRecipeCreationRequestInput(condition)
}
