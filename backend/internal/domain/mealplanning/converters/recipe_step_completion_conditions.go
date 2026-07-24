package converters

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v6/identifiers"

	"github.com/ccoveille/go-safecast"
)

// ConvertRecipeStepCompletionConditionCreationRequestInputToRecipeStepCompletionConditionDatabaseCreationInput creates a RecipeStepCompletionConditionDatabaseCreationInput from a RecipeStepCompletionConditionCreationRequestInput.
func ConvertRecipeStepCompletionConditionCreationRequestInputToRecipeStepCompletionConditionDatabaseCreationInput(recipeStep *types.RecipeStepDatabaseCreationInput, input *types.RecipeStepCompletionConditionCreationRequestInput) *types.RecipeStepCompletionConditionDatabaseCreationInput {
	recipeStepCompletionConditionID := identifiers.New()

	var ingredients []*types.RecipeStepCompletionConditionIngredientDatabaseCreationInput
	for _, i := range input.Ingredients {
		x := &types.RecipeStepCompletionConditionIngredientDatabaseCreationInput{
			ID:                                     identifiers.New(),
			RecipeStepIngredient:                   recipeStep.Ingredients[i].ID,
			BelongsToRecipeStepCompletionCondition: recipeStepCompletionConditionID,
		}

		ingredients = append(ingredients, x)
	}

	x := &types.RecipeStepCompletionConditionDatabaseCreationInput{
		ID:                  recipeStepCompletionConditionID,
		IngredientStateID:   input.IngredientStateID,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Notes:               input.Notes,
		Ingredients:         ingredients,
		Optional:            input.Optional,
	}

	return x
}

// ConvertRecipeStepCompletionConditionForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionDatabaseCreationInput creates a RecipeStepCompletionConditionDatabaseCreationInput from a RecipeStepCompletionConditionForExitingRecipeCreationRequestInput.
func ConvertRecipeStepCompletionConditionForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionDatabaseCreationInput(input *types.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput) *types.RecipeStepCompletionConditionDatabaseCreationInput {
	id := identifiers.New()

	var ingredients []*types.RecipeStepCompletionConditionIngredientDatabaseCreationInput
	for _, i := range input.Ingredients {
		x := ConvertRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionIngredientDatabaseCreationInput(i)
		x.BelongsToRecipeStepCompletionCondition = id
		ingredients = append(ingredients, x)
	}

	x := &types.RecipeStepCompletionConditionDatabaseCreationInput{
		ID:                  id,
		IngredientStateID:   input.IngredientStateID,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Notes:               input.Notes,
		Ingredients:         ingredients,
		Optional:            input.Optional,
	}

	return x
}

// ConvertRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionIngredientDatabaseCreationInput creates a RecipeStepCompletionConditionIngredientDatabaseCreationInput from a RecipeStepCompletionConditionCreationRequestInput.
func ConvertRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionIngredientDatabaseCreationInput(input *types.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput) *types.RecipeStepCompletionConditionIngredientDatabaseCreationInput {
	x := &types.RecipeStepCompletionConditionIngredientDatabaseCreationInput{
		ID:                   identifiers.New(),
		RecipeStepIngredient: input.RecipeStepIngredient,
	}

	return x
}

// ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionUpdateRequestInput creates a RecipeStepCompletionConditionUpdateRequestInput from a RecipeStepCompletionCondition.
func ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionUpdateRequestInput(input *types.RecipeStepCompletionCondition) *types.RecipeStepCompletionConditionUpdateRequestInput {
	x := &types.RecipeStepCompletionConditionUpdateRequestInput{
		IngredientStateID:   &input.IngredientState.ID,
		BelongsToRecipeStep: &input.BelongsToRecipeStep,
		Notes:               &input.Notes,
		Optional:            &input.Optional,
	}

	return x
}

// ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionCreationRequestInput builds a RecipeStepCompletionConditionCreationRequestInput from a RecipeStepCompletionCondition.
func ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionCreationRequestInput(recipeStep *types.RecipeStep, recipeStepCompletionCondition *types.RecipeStepCompletionCondition) *types.RecipeStepCompletionConditionCreationRequestInput {
	var ingredients []uint64
	for _, ingredientIndex := range recipeStepCompletionCondition.Ingredients {
		for i, ingredient := range recipeStep.Ingredients {
			if ingredient.ID == ingredientIndex.RecipeStepIngredient {
				x, err := safecast.Convert[uint64](i)
				mustnt(err)
				ingredients = append(ingredients, x)
			}
		}
	}

	return &types.RecipeStepCompletionConditionCreationRequestInput{
		IngredientStateID:   recipeStepCompletionCondition.IngredientState.ID,
		BelongsToRecipeStep: recipeStepCompletionCondition.BelongsToRecipeStep,
		Notes:               recipeStepCompletionCondition.Notes,
		Ingredients:         ingredients,
		Optional:            recipeStepCompletionCondition.Optional,
	}
}

// ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionForExistingRecipeCreationRequestInput builds a RecipeStepCompletionConditionCreationRequestInput from a RecipeStepCompletionCondition.
func ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionForExistingRecipeCreationRequestInput(recipeStepCompletionCondition *types.RecipeStepCompletionCondition) *types.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput {
	var ingredients []*types.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput
	for _, i := range recipeStepCompletionCondition.Ingredients {
		x := ConvertRecipeStepCompletionConditionIngredientToRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput(i)
		ingredients = append(ingredients, x)
	}

	return &types.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput{
		IngredientStateID:   recipeStepCompletionCondition.IngredientState.ID,
		BelongsToRecipeStep: recipeStepCompletionCondition.BelongsToRecipeStep,
		Notes:               recipeStepCompletionCondition.Notes,
		Ingredients:         ingredients,
		Optional:            recipeStepCompletionCondition.Optional,
	}
}

// ConvertRecipeStepCompletionConditionIngredientToRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput builds a RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput from a RecipeStepCompletionCondition.
func ConvertRecipeStepCompletionConditionIngredientToRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput(recipeStepCompletionConditionIngredient *types.RecipeStepCompletionConditionIngredient) *types.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput {
	return &types.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput{
		RecipeStepIngredient: recipeStepCompletionConditionIngredient.RecipeStepIngredient,
	}
}

// ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionDatabaseCreationInput builds a RecipeStepCompletionConditionDatabaseCreationInput from a RecipeStepCompletionCondition.
func ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionDatabaseCreationInput(recipeStepCompletionCondition *types.RecipeStepCompletionCondition) *types.RecipeStepCompletionConditionDatabaseCreationInput {
	ingredients := []*types.RecipeStepCompletionConditionIngredientDatabaseCreationInput{}
	for _, ingredient := range recipeStepCompletionCondition.Ingredients {
		ingredients = append(ingredients, &types.RecipeStepCompletionConditionIngredientDatabaseCreationInput{
			ID:                                     ingredient.ID,
			BelongsToRecipeStepCompletionCondition: ingredient.BelongsToRecipeStepCompletionCondition,
			RecipeStepIngredient:                   ingredient.RecipeStepIngredient,
		})
	}

	return &types.RecipeStepCompletionConditionDatabaseCreationInput{
		ID:                  recipeStepCompletionCondition.ID,
		Optional:            recipeStepCompletionCondition.Optional,
		Notes:               recipeStepCompletionCondition.Notes,
		IngredientStateID:   recipeStepCompletionCondition.IngredientState.ID,
		BelongsToRecipeStep: recipeStepCompletionCondition.BelongsToRecipeStep,
		Ingredients:         ingredients,
	}
}
