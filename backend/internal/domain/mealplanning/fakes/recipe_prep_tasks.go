package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/pointer"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeRecipePrepTask builds a faked recipe prep task.
func BuildFakeRecipePrepTask() *types.RecipePrepTask {
	prepTask := fake.BuildFakeRecord[types.RecipePrepTask]()

	// One of the four ways this domain knows to store something between steps.
	prepTask.StorageType = gofakeit.RandomString([]string{
		types.RecipePrepTaskStorageTypeUncovered,
		types.RecipePrepTaskStorageTypeCovered,
		types.RecipePrepTaskStorageTypeAirtightContainer,
		types.RecipePrepTaskStorageTypeWireRack,
	})

	// Two ranges rather than four independent numbers.
	prepTask.MinStorageTemperatureInCelsius, prepTask.MaxStorageTemperatureInCelsius = BuildFakeOptionalFloat32MinMax()
	prepTask.MinTimeBufferBeforeRecipeInSeconds, prepTask.MaxTimeBufferBeforeRecipeInSeconds = BuildFakeUint32WithOptionalMax()

	// A task with no steps is a task about nothing.
	taskSteps := make([]*types.RecipePrepTaskStep, 0, exampleQuantity)
	for range exampleQuantity {
		taskSteps = append(taskSteps, BuildFakeRecipePrepTaskStep())
	}
	prepTask.TaskSteps = taskSteps

	return prepTask
}

// BuildFakeRecipePrepTasksList builds a faked list of recipe prep tasks.
//
// Hand-written rather than fake.BuildFakePage because it returns the data without
// a page around it: prep tasks are read as a recipe's children rather than as a page,
// and the only caller reads Data.
func BuildFakeRecipePrepTasksList() *filtering.QueryFilteredResult[types.RecipePrepTask] {
	recipePrepTasks := &filtering.QueryFilteredResult[types.RecipePrepTask]{}
	for range exampleQuantity {
		recipePrepTasks.Data = append(recipePrepTasks.Data, BuildFakeRecipePrepTask())
	}

	return recipePrepTasks
}

// BuildFakeRecipePrepTaskStep builds a faked recipe prep task step.
func BuildFakeRecipePrepTaskStep() *types.RecipePrepTaskStep {
	return fake.BuildFakeRecord[types.RecipePrepTaskStep]()
}

func BuildFakeRecipePrepTaskStepCreationRequestInput() *types.RecipePrepTaskStepCreationRequestInput {
	return &types.RecipePrepTaskStepCreationRequestInput{
		BelongsToRecipeStep: fake.BuildFakeID(),
		SatisfiesRecipeStep: gofakeit.Bool(),
	}
}

func BuildFakeRecipePrepTaskStepUpdateRequestInput() *types.RecipePrepTaskStepUpdateRequestInput {
	return &types.RecipePrepTaskStepUpdateRequestInput{
		BelongsToRecipeStep:     pointer.To(fake.BuildFakeID()),
		BelongsToRecipePrepTask: pointer.To(fake.BuildFakeID()),
		SatisfiesRecipeStep:     pointer.To(gofakeit.Bool()),
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
		Notes:                              fake.BuildFakeString(),
		ExplicitStorageInstructions:        fake.BuildFakeString(),
		Name:                               fake.BuildFakeString(),
		Optional:                           gofakeit.Bool(),
		Description:                        fake.BuildFakeString(),
		StorageType:                        types.RecipePrepTaskStorageTypeUncovered,
		BelongsToRecipe:                    fake.BuildFakeID(),
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
		Notes:                              pointer.To(fake.BuildFakeString()),
		ExplicitStorageInstructions:        pointer.To(fake.BuildFakeString()),
		Name:                               pointer.To(fake.BuildFakeString()),
		Description:                        pointer.To(fake.BuildFakeString()),
		Optional:                           pointer.To(gofakeit.Bool()),
		StorageType:                        pointer.To(types.RecipePrepTaskStorageTypeUncovered),
		BelongsToRecipe:                    pointer.To(fake.BuildFakeID()),
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
		Notes:                              pointer.To(fake.BuildFakeString()),
		ExplicitStorageInstructions:        pointer.To(fake.BuildFakeString()),
		Name:                               pointer.To(fake.BuildFakeString()),
		Description:                        pointer.To(fake.BuildFakeString()),
		Optional:                           pointer.To(gofakeit.Bool()),
		StorageType:                        pointer.To(types.RecipePrepTaskStorageTypeUncovered),
		BelongsToRecipe:                    pointer.To(fake.BuildFakeID()),
		TaskSteps:                          taskSteps,
		MinTimeBufferBeforeRecipeInSeconds: minBuf,
		MaxTimeBufferBeforeRecipeInSeconds: maxBuf,
		MinStorageTemperatureInCelsius:     minTemp,
		MaxStorageTemperatureInCelsius:     maxTemp,
	}
}
