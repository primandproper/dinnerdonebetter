package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
)

// BuildFakeRecipeCreationRequestInput builds a faked RecipeCreationRequestInput.
func BuildFakeRecipeCreationRequestInput() *mealplanning.RecipeCreationRequestInput {
	exampleRecipe := BuildFakeRecipe()
	exampleCreationInput := converters.ConvertRecipeToRecipeCreationRequestInput(exampleRecipe)
	examplePrepTask := BuildFakeRecipePrepTask()
	examplePrepTaskInput := converters.ConvertRecipePrepTaskToRecipePrepTaskWithinRecipeCreationRequestInput(exampleRecipe, examplePrepTask)
	examplePrepTaskInput.RecipeSteps = []*mealplanning.RecipePrepTaskStepWithinRecipeCreationRequestInput{
		{
			BelongsToRecipeStepIndex: exampleCreationInput.Steps[0].Index,
			SatisfiesRecipeStep:      false,
		},
	}
	exampleCreationInput.PrepTasks = []*mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput{
		examplePrepTaskInput,
	}

	return exampleCreationInput
}
