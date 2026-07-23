package converters

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v5/identifiers"
)

// ConvertMealCreationRequestInputToMealDatabaseCreationInput creates a MealDatabaseCreationInput from a MealCreationRequestInput.
func ConvertMealCreationRequestInputToMealDatabaseCreationInput(input *mealplanning.MealCreationRequestInput) *mealplanning.MealDatabaseCreationInput {
	convertedComponents := []*mealplanning.MealComponentDatabaseCreationInput{}
	for _, x := range input.Components {
		convertedComponents = append(convertedComponents, ConvertMealComponentCreationRequestInputToMealComponentDatabaseCreationInput(x))
	}

	x := &mealplanning.MealDatabaseCreationInput{
		ID:                   identifiers.New(),
		Name:                 input.Name,
		Description:          input.Description,
		MinEstimatedPortions: input.MinEstimatedPortions,
		MaxEstimatedPortions: input.MaxEstimatedPortions,
		Components:           convertedComponents,
		EligibleForMealPlans: input.EligibleForMealPlans,
	}

	return x
}

// ConvertMealComponentCreationRequestInputToMealComponentDatabaseCreationInput creates a MealComponentDatabaseCreationInput from a MealComponentCreationRequestInput.
func ConvertMealComponentCreationRequestInputToMealComponentDatabaseCreationInput(input *mealplanning.MealComponentCreationRequestInput) *mealplanning.MealComponentDatabaseCreationInput {
	x := &mealplanning.MealComponentDatabaseCreationInput{
		RecipeID:      input.RecipeID,
		ComponentType: input.ComponentType,
		RecipeScale:   input.RecipeScale,
	}

	return x
}

// ConvertMealToMealCreationRequestInput builds a faked MealCreationRequestInput from a Meal.
func ConvertMealToMealCreationRequestInput(meal *mealplanning.Meal) *mealplanning.MealCreationRequestInput {
	convertedComponents := []*mealplanning.MealComponentCreationRequestInput{}
	for _, x := range meal.Components {
		convertedComponents = append(convertedComponents, ConvertMealComponentToMealComponentCreationRequestInput(x))
	}

	return &mealplanning.MealCreationRequestInput{
		Name:                 meal.Name,
		Description:          meal.Description,
		MinEstimatedPortions: meal.MinEstimatedPortions,
		MaxEstimatedPortions: meal.MaxEstimatedPortions,
		Components:           convertedComponents,
		EligibleForMealPlans: meal.EligibleForMealPlans,
	}
}

// ConvertMealComponentToMealComponentCreationRequestInput creates a MealComponentCreationRequestInput from a MealComponent.
func ConvertMealComponentToMealComponentCreationRequestInput(input *mealplanning.MealComponent) *mealplanning.MealComponentCreationRequestInput {
	x := &mealplanning.MealComponentCreationRequestInput{
		RecipeID:      input.Recipe.ID,
		RecipeScale:   input.RecipeScale,
		ComponentType: input.ComponentType,
	}

	return x
}

// ConvertMealToMealDatabaseCreationInput builds a faked MealDatabaseCreationInput from a recipe.
func ConvertMealToMealDatabaseCreationInput(meal *mealplanning.Meal) *mealplanning.MealDatabaseCreationInput {
	convertedComponents := []*mealplanning.MealComponentDatabaseCreationInput{}
	for _, x := range meal.Components {
		convertedComponents = append(convertedComponents, ConvertMealComponentToMealComponentDatabaseCreationInput(x))
	}

	return &mealplanning.MealDatabaseCreationInput{
		ID:                   meal.ID,
		Name:                 meal.Name,
		Description:          meal.Description,
		MinEstimatedPortions: meal.MinEstimatedPortions,
		MaxEstimatedPortions: meal.MaxEstimatedPortions,
		CreatedByUser:        meal.CreatedByUser,
		Components:           convertedComponents,
		EligibleForMealPlans: meal.EligibleForMealPlans,
	}
}

// ConvertMealComponentToMealComponentDatabaseCreationInput creates a MealComponentDatabaseCreationInput from a MealComponent.
func ConvertMealComponentToMealComponentDatabaseCreationInput(input *mealplanning.MealComponent) *mealplanning.MealComponentDatabaseCreationInput {
	x := &mealplanning.MealComponentDatabaseCreationInput{
		RecipeID:      input.Recipe.ID,
		RecipeScale:   input.RecipeScale,
		ComponentType: input.ComponentType,
	}

	return x
}

