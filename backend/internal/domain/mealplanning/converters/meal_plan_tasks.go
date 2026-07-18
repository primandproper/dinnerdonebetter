package converters

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v5/identifiers"
)

// ConvertMealPlanTaskCreationRequestInputToMealPlanTaskDatabaseCreationInput creates a MealPlanTaskDatabaseCreationInput from a MealPlanTaskCreationRequestInput.
func ConvertMealPlanTaskCreationRequestInputToMealPlanTaskDatabaseCreationInput(input *types.MealPlanTaskCreationRequestInput) *types.MealPlanTaskDatabaseCreationInput {
	x := &types.MealPlanTaskDatabaseCreationInput{
		ID:                  identifiers.New(),
		AssignedToUser:      input.AssignedToUser,
		CreationExplanation: input.CreationExplanation,
		StatusExplanation:   input.StatusExplanation,
		MealPlanOptionID:    input.MealPlanOptionID,
		RecipePrepTaskID:    input.RecipePrepTaskID,
	}

	return x
}

// ConvertMealPlanTaskToMealPlanTaskCreationRequestInput builds a MealPlanTaskCreationRequestInput.
func ConvertMealPlanTaskToMealPlanTaskCreationRequestInput(x *types.MealPlanTask) *types.MealPlanTaskCreationRequestInput {
	return &types.MealPlanTaskCreationRequestInput{
		AssignedToUser:      x.AssignedToUser,
		Status:              x.Status,
		CreationExplanation: x.CreationExplanation,
		StatusExplanation:   x.StatusExplanation,
		MealPlanOptionID:    x.MealPlanOption.ID,
		RecipePrepTaskID:    x.RecipePrepTask.ID,
	}
}

// ConvertMealPlanTaskToMealPlanTaskDatabaseCreationInput builds a MealPlanTaskDatabaseCreationInput.
func ConvertMealPlanTaskToMealPlanTaskDatabaseCreationInput(x *types.MealPlanTask) *types.MealPlanTaskDatabaseCreationInput {
	return &types.MealPlanTaskDatabaseCreationInput{
		ID:                  x.ID,
		AssignedToUser:      x.AssignedToUser,
		CreationExplanation: x.CreationExplanation,
		StatusExplanation:   x.StatusExplanation,
		MealPlanOptionID:    x.MealPlanOption.ID,
		RecipePrepTaskID:    x.RecipePrepTask.ID,
	}
}
