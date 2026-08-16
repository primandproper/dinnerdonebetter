package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
)

// BuildFakeMealPlanTaskDatabaseCreationInputs builds a faked MealPlanTaskList.
func BuildFakeMealPlanTaskDatabaseCreationInputs() []*types.MealPlanTaskDatabaseCreationInput {
	var examples []*types.MealPlanTaskDatabaseCreationInput
	for range exampleQuantity {
		examples = append(examples, &types.MealPlanTaskDatabaseCreationInput{
			MealPlanOptionID:    "",
			ID:                  BuildFakeID(),
			StatusExplanation:   buildUniqueString(),
			CreationExplanation: buildUniqueString(),
		})
	}

	return examples
}
