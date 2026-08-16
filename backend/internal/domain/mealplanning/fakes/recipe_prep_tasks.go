package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/pointer"

	fake "github.com/brianvoe/gofakeit/v7"
)

func BuildFakeRecipePrepTasksList() *filtering.QueryFilteredResult[types.RecipePrepTask] {
	recipePrepTasks := &filtering.QueryFilteredResult[types.RecipePrepTask]{}
	for range exampleQuantity {
		recipePrepTasks.Data = append(recipePrepTasks.Data, BuildFakeRecipePrepTask())
	}

	return recipePrepTasks
}

func BuildFakeRecipePrepTaskUpdateRequestInputFromRecipePrepTask(input *types.RecipePrepTask) *types.RecipePrepTaskUpdateRequestInput {
	taskSteps := []*types.RecipePrepTaskStepUpdateRequestInput{}
	for _, x := range input.TaskSteps {
		taskSteps = append(taskSteps, converters.ConvertRecipePrepTaskStepToRecipePrepTaskStepUpdateRequestInput(x))
	}

	minTemp, maxTemp := BuildFakeOptionalFloat32MinMax()
	minBuf, maxBuf := BuildFakeOptionalUint32MinMax()
	return &types.RecipePrepTaskUpdateRequestInput{
		Notes:                              new(buildUniqueString()),
		ExplicitStorageInstructions:        new(buildUniqueString()),
		Name:                               new(buildUniqueString()),
		Description:                        new(buildUniqueString()),
		Optional:                           new(fake.Bool()),
		StorageType:                        pointer.To(types.RecipePrepTaskStorageTypeUncovered),
		BelongsToRecipe:                    new(BuildFakeID()),
		TaskSteps:                          taskSteps,
		MinTimeBufferBeforeRecipeInSeconds: minBuf,
		MaxTimeBufferBeforeRecipeInSeconds: maxBuf,
		MinStorageTemperatureInCelsius:     minTemp,
		MaxStorageTemperatureInCelsius:     maxTemp,
	}
}
