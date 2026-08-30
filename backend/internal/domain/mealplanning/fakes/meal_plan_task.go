package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/pointer"
)

// mealPlanTaskUnfinished is the status a task is created in.
const mealPlanTaskUnfinished = "unfinished"

// BuildFakeMealPlanTask builds a faked meal plan task.
func BuildFakeMealPlanTask() *types.MealPlanTask {
	task := fake.BuildFakeRecord[types.MealPlanTask]()

	task.Status = mealPlanTaskUnfinished

	// What the task is for and what it is prepping: both are whole records, and the
	// worker that runs the task reads through them.
	task.MealPlanOption = *BuildFakeMealPlanOption()
	task.RecipePrepTask = *BuildFakeRecipePrepTask()

	return task
}

// BuildFakeMealPlanTaskCreationRequestInput builds a faked MealPlanTaskCreationRequestInput.
func BuildFakeMealPlanTaskCreationRequestInput() *types.MealPlanTaskCreationRequestInput {
	x := BuildFakeMealPlanTask()

	return converters.ConvertMealPlanTaskToMealPlanTaskCreationRequestInput(x)
}

// BuildFakeMealPlanTasksList builds a faked MealPlanTaskList.
func BuildFakeMealPlanTasksList() *filtering.QueryFilteredResult[types.MealPlanTask] {
	return fake.BuildFakePage(BuildFakeMealPlanTask)
}

// BuildFakeMealPlanTaskDatabaseCreationInputs builds faked MealPlanTaskDatabaseCreationInputs.
func BuildFakeMealPlanTaskDatabaseCreationInputs() []*types.MealPlanTaskDatabaseCreationInput {
	examples := make([]*types.MealPlanTaskDatabaseCreationInput, 0, exampleQuantity)
	for range exampleQuantity {
		input := fake.BuildFakeRecord[types.MealPlanTaskDatabaseCreationInput]()

		// Empty because the caller fills it with the option it is writing tasks for.
		input.MealPlanOptionID = ""

		examples = append(examples, input)
	}

	return examples
}

// BuildFakeMealPlanTaskStatusChangeRequestInput builds a faked MealPlanTaskStatusChangeRequestInput.
func BuildFakeMealPlanTaskStatusChangeRequestInput() *types.MealPlanTaskStatusChangeRequestInput {
	input := fake.BuildFakeRecord[types.MealPlanTaskStatusChangeRequestInput]()
	input.Status = pointer.To(mealPlanTaskUnfinished)

	return input
}
