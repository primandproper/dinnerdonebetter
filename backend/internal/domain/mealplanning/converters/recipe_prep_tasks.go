package converters

import (
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v6/identifiers"
)

// ConvertRecipePrepTaskToRecipePrepTaskUpdateRequestInput creates a RecipePrepTaskUpdateRequestInput from a RecipePrepTask.
func ConvertRecipePrepTaskToRecipePrepTaskUpdateRequestInput(input *mealplanning.RecipePrepTask) *mealplanning.RecipePrepTaskUpdateRequestInput {
	taskSteps := []*mealplanning.RecipePrepTaskStepUpdateRequestInput{}
	for _, x := range input.TaskSteps {
		y := x
		taskSteps = append(taskSteps, &mealplanning.RecipePrepTaskStepUpdateRequestInput{
			BelongsToRecipeStep:     &y.BelongsToRecipeStep,
			BelongsToRecipePrepTask: &y.BelongsToRecipePrepTask,
			SatisfiesRecipeStep:     &y.SatisfiesRecipeStep,
		})
	}
	x := &mealplanning.RecipePrepTaskUpdateRequestInput{
		Name:                               &input.Name,
		Description:                        &input.Description,
		Notes:                              &input.Notes,
		ExplicitStorageInstructions:        &input.ExplicitStorageInstructions,
		Optional:                           &input.Optional,
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: &input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
		StorageType:                        &input.StorageType,
		BelongsToRecipe:                    &input.BelongsToRecipe,
		TaskSteps:                          taskSteps,
	}

	return x
}

// ConvertRecipePrepTaskCreationRequestInputToRecipePrepTaskDatabaseCreationInput creates a DatabaseCreationInput from a CreationInput.
func ConvertRecipePrepTaskCreationRequestInputToRecipePrepTaskDatabaseCreationInput(input *mealplanning.RecipePrepTaskCreationRequestInput) *mealplanning.RecipePrepTaskDatabaseCreationInput {
	taskSteps := []*mealplanning.RecipePrepTaskStepDatabaseCreationInput{}
	for _, x := range input.RecipeSteps {
		taskSteps = append(taskSteps, &mealplanning.RecipePrepTaskStepDatabaseCreationInput{
			ID:                  identifiers.New(),
			BelongsToRecipeStep: x.BelongsToRecipeStep,
			SatisfiesRecipeStep: x.SatisfiesRecipeStep,
		})
	}

	x := &mealplanning.RecipePrepTaskDatabaseCreationInput{
		ID:                                 identifiers.New(),
		Name:                               input.Name,
		Description:                        input.Description,
		Notes:                              input.Notes,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Optional:                           input.Optional,
		StorageType:                        input.StorageType,
		BelongsToRecipe:                    input.BelongsToRecipe,
		TaskSteps:                          taskSteps,
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
	}

	return x
}

// ConvertRecipePrepTaskWithinRecipeCreationRequestInputToRecipePrepTaskDatabaseCreationInput creates a DatabaseCreationInput from a CreationInput.
func ConvertRecipePrepTaskWithinRecipeCreationRequestInputToRecipePrepTaskDatabaseCreationInput(recipe *mealplanning.RecipeDatabaseCreationInput, input *mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput) (*mealplanning.RecipePrepTaskDatabaseCreationInput, error) {
	x := &mealplanning.RecipePrepTaskDatabaseCreationInput{
		ID:                                 identifiers.New(),
		Name:                               input.Name,
		Description:                        input.Description,
		Notes:                              input.Notes,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Optional:                           input.Optional,
		StorageType:                        input.StorageType,
		BelongsToRecipe:                    input.BelongsToRecipe,
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
	}

	x.TaskSteps = []*mealplanning.RecipePrepTaskStepDatabaseCreationInput{}
	for i, ts := range input.RecipeSteps {
		if rs := recipe.FindStepByIndex(ts.BelongsToRecipeStepIndex); rs != nil {
			x.TaskSteps = append(x.TaskSteps, &mealplanning.RecipePrepTaskStepDatabaseCreationInput{
				ID:                      identifiers.New(),
				BelongsToRecipeStep:     rs.ID,
				BelongsToRecipePrepTask: x.ID,
				SatisfiesRecipeStep:     ts.SatisfiesRecipeStep,
			})
		} else {
			return nil, fmt.Errorf("task step #%d has an invalid recipe step index", i+1)
		}
	}

	return x, nil
}

// ConvertRecipePrepTaskToRecipePrepTaskDatabaseCreationInput builds a RecipePrepTaskDatabaseCreationInput from a recipe prep task.
func ConvertRecipePrepTaskToRecipePrepTaskDatabaseCreationInput(input *mealplanning.RecipePrepTask) *mealplanning.RecipePrepTaskDatabaseCreationInput {
	taskSteps := []*mealplanning.RecipePrepTaskStepDatabaseCreationInput{}
	for _, step := range input.TaskSteps {
		taskSteps = append(taskSteps, ConvertRecipePrepTaskStepToRecipePrepTaskStepDatabaseCreationInput(step))
	}

	return &mealplanning.RecipePrepTaskDatabaseCreationInput{
		ID:                                 input.ID,
		Name:                               input.Name,
		Description:                        input.Description,
		Notes:                              input.Notes,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Optional:                           input.Optional,
		StorageType:                        input.StorageType,
		TaskSteps:                          taskSteps,
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
		BelongsToRecipe:                    input.BelongsToRecipe,
	}
}

func ConvertRecipePrepTaskStepToRecipePrepTaskStepCreationRequestInput(input *mealplanning.RecipePrepTaskStep) *mealplanning.RecipePrepTaskStepCreationRequestInput {
	return &mealplanning.RecipePrepTaskStepCreationRequestInput{
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		SatisfiesRecipeStep: input.SatisfiesRecipeStep,
	}
}

func ConvertRecipePrepTaskToRecipePrepTaskCreationRequestInput(input *mealplanning.RecipePrepTask) *mealplanning.RecipePrepTaskCreationRequestInput {
	taskSteps := []*mealplanning.RecipePrepTaskStepCreationRequestInput{}
	for _, x := range input.TaskSteps {
		taskSteps = append(taskSteps, ConvertRecipePrepTaskStepToRecipePrepTaskStepCreationRequestInput(x))
	}

	return &mealplanning.RecipePrepTaskCreationRequestInput{
		Name:                               input.Name,
		Description:                        input.Description,
		Notes:                              input.Notes,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Optional:                           input.Optional,
		StorageType:                        input.StorageType,
		BelongsToRecipe:                    input.BelongsToRecipe,
		RecipeSteps:                        taskSteps,
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
	}
}

func ConvertRecipePrepTaskToRecipePrepTaskWithinRecipeCreationRequestInput(recipe *mealplanning.Recipe, input *mealplanning.RecipePrepTask) *mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput {
	taskSteps := []*mealplanning.RecipePrepTaskStepWithinRecipeCreationRequestInput{}
	for _, x := range input.TaskSteps {
		taskSteps = append(taskSteps, ConvertRecipePrepTaskStepToRecipePrepTaskStepWithinRecipeCreationRequestInput(recipe, x))
	}

	return &mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput{
		Name:                               input.Name,
		Description:                        input.Description,
		Notes:                              input.Notes,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Optional:                           input.Optional,
		StorageType:                        input.StorageType,
		BelongsToRecipe:                    input.BelongsToRecipe,
		RecipeSteps:                        taskSteps,
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
	}
}

func ConvertRecipePrepTaskStepToRecipePrepTaskStepWithinRecipeCreationRequestInput(recipe *mealplanning.Recipe, input *mealplanning.RecipePrepTaskStep) *mealplanning.RecipePrepTaskStepWithinRecipeCreationRequestInput {
	var belongsToIndex uint32
	if x := recipe.FindStepByID(input.BelongsToRecipeStep); x != nil {
		belongsToIndex = x.Index
	}

	return &mealplanning.RecipePrepTaskStepWithinRecipeCreationRequestInput{
		BelongsToRecipeStepIndex: belongsToIndex,
		SatisfiesRecipeStep:      input.SatisfiesRecipeStep,
	}
}

func ConvertRecipePrepTaskStepToRecipePrepTaskStepUpdateRequestInput(input *mealplanning.RecipePrepTaskStep) *mealplanning.RecipePrepTaskStepUpdateRequestInput {
	return &mealplanning.RecipePrepTaskStepUpdateRequestInput{
		BelongsToRecipeStep:     new(input.BelongsToRecipeStep),
		BelongsToRecipePrepTask: new(input.BelongsToRecipePrepTask),
		SatisfiesRecipeStep:     new(input.SatisfiesRecipeStep),
	}
}

func ConvertRecipePrepTaskStepToRecipePrepTaskStepDatabaseCreationInput(input *mealplanning.RecipePrepTaskStep) *mealplanning.RecipePrepTaskStepDatabaseCreationInput {
	return &mealplanning.RecipePrepTaskStepDatabaseCreationInput{
		ID:                      input.ID,
		BelongsToRecipeStep:     input.BelongsToRecipeStep,
		BelongsToRecipePrepTask: input.BelongsToRecipePrepTask,
		SatisfiesRecipeStep:     input.SatisfiesRecipeStep,
	}
}
