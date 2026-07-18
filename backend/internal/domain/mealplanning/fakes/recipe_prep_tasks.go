package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v5/filtering"
	"github.com/primandproper/platform-go/v5/pointer"

	fake "github.com/brianvoe/gofakeit/v7"
)

func BuildFakeRecipePrepTask() *types.RecipePrepTask {
	recipePrepTaskSteps := []*types.RecipePrepTaskStep{}
	for range exampleQuantity {
		recipePrepTaskSteps = append(recipePrepTaskSteps, BuildFakeRecipePrepTaskStep())
	}

	minTemp, maxTemp := BuildFakeOptionalFloat32MinMax()
	minBuf, maxBuf := BuildFakeUint32WithOptionalMax()
	return &types.RecipePrepTask{
		ID:                          BuildFakeID(),
		Notes:                       buildUniqueString(),
		Name:                        buildUniqueString(),
		Description:                 buildUniqueString(),
		Optional:                    fake.Bool(),
		ExplicitStorageInstructions: buildUniqueString(),
		StorageType: fake.RandomString([]string{
			types.RecipePrepTaskStorageTypeUncovered,
			types.RecipePrepTaskStorageTypeCovered,
			types.RecipePrepTaskStorageTypeAirtightContainer,
			types.RecipePrepTaskStorageTypeWireRack,
		}),
		TaskSteps:                          recipePrepTaskSteps,
		MinStorageTemperatureInCelsius:     minTemp,
		MaxStorageTemperatureInCelsius:     maxTemp,
		MinTimeBufferBeforeRecipeInSeconds: minBuf,
		MaxTimeBufferBeforeRecipeInSeconds: maxBuf,
		BelongsToRecipe:                    BuildFakeID(),
		CreatedAt:                          BuildFakeTime(),
		LastUpdatedAt:                      nil,
		ArchivedAt:                         nil,
	}
}

func BuildFakeRecipePrepTasksList() *filtering.QueryFilteredResult[types.RecipePrepTask] {
	recipePrepTasks := &filtering.QueryFilteredResult[types.RecipePrepTask]{}
	for range exampleQuantity {
		recipePrepTasks.Data = append(recipePrepTasks.Data, BuildFakeRecipePrepTask())
	}

	return recipePrepTasks
}

func BuildFakeRecipePrepTaskStep() *types.RecipePrepTaskStep {
	return &types.RecipePrepTaskStep{
		ID:                      BuildFakeID(),
		BelongsToRecipeStep:     BuildFakeID(),
		BelongsToRecipePrepTask: BuildFakeID(),
		SatisfiesRecipeStep:     fake.Bool(),
	}
}

func BuildFakeRecipePrepTaskStepCreationRequestInput() *types.RecipePrepTaskStepCreationRequestInput {
	return &types.RecipePrepTaskStepCreationRequestInput{
		BelongsToRecipeStep: BuildFakeID(),
		SatisfiesRecipeStep: fake.Bool(),
	}
}

func BuildFakeRecipePrepTaskStepUpdateRequestInput() *types.RecipePrepTaskStepUpdateRequestInput {
	return &types.RecipePrepTaskStepUpdateRequestInput{
		BelongsToRecipeStep:     new(BuildFakeID()),
		BelongsToRecipePrepTask: new(BuildFakeID()),
		SatisfiesRecipeStep:     new(fake.Bool()),
	}
}

func BuildFakeRecipePrepTaskCreationRequestInput() *types.RecipePrepTaskCreationRequestInput {
	taskSteps := []*types.RecipePrepTaskStepCreationRequestInput{}
	for range exampleQuantity {
		taskSteps = append(taskSteps, BuildFakeRecipePrepTaskStepCreationRequestInput())
	}

	minTemp, maxTemp := BuildFakeOptionalFloat32MinMax()
	minBuf, maxBuf := BuildFakeUint32WithOptionalMax()
	return &types.RecipePrepTaskCreationRequestInput{
		Notes:                              buildUniqueString(),
		ExplicitStorageInstructions:        buildUniqueString(),
		Name:                               buildUniqueString(),
		Optional:                           fake.Bool(),
		Description:                        buildUniqueString(),
		StorageType:                        types.RecipePrepTaskStorageTypeUncovered,
		BelongsToRecipe:                    BuildFakeID(),
		RecipeSteps:                        taskSteps,
		MinTimeBufferBeforeRecipeInSeconds: minBuf,
		MaxTimeBufferBeforeRecipeInSeconds: maxBuf,
		MinStorageTemperatureInCelsius:     minTemp,
		MaxStorageTemperatureInCelsius:     maxTemp,
	}
}

func BuildFakeRecipePrepTaskUpdateRequestInput() *types.RecipePrepTaskUpdateRequestInput {
	taskSteps := []*types.RecipePrepTaskStepUpdateRequestInput{}
	for range exampleQuantity {
		taskSteps = append(taskSteps, BuildFakeRecipePrepTaskStepUpdateRequestInput())
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
		MinTimeBufferBeforeRecipeInSeconds: minBuf,
		MaxTimeBufferBeforeRecipeInSeconds: maxBuf,
		MinStorageTemperatureInCelsius:     minTemp,
		MaxStorageTemperatureInCelsius:     maxTemp,
		TaskSteps:                          taskSteps,
	}
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
