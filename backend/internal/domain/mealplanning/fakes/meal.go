package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
)

// BuildFakeMeal builds a faked meal.
func BuildFakeMeal() *mealplanning.Meal {
	meal := fake.BuildFakeRecord[mealplanning.Meal]()

	// A meal is its components, and one with none is a meal every scaling and grocery
	// calculation returns nothing for.
	components := make([]*mealplanning.MealComponent, 0, exampleQuantity)
	for range exampleQuantity {
		components = append(components, BuildFakeMealComponent())
	}
	meal.Components = components

	// A meal that may be voted on, which is what the meal planning path needs of one.
	meal.EligibleForMealPlans = true

	return meal
}

// BuildFakeMealComponent builds a faked meal component.
func BuildFakeMealComponent() *mealplanning.MealComponent {
	component := fake.BuildFakeRecord[mealplanning.MealComponent]()

	component.Recipe = *BuildFakeRecipe()

	// Unscaled, so that a test that scales a meal is measuring its own factor rather
	// than one the fake applied first.
	component.RecipeScale = 1.0

	// One of the component types the domain enumerates.
	component.ComponentType = mealplanning.MealComponentTypesMain

	return component
}

// BuildFakeMealsList builds a faked MealList.
func BuildFakeMealsList() *filtering.QueryFilteredResult[mealplanning.Meal] {
	return fake.BuildFakePage(BuildFakeMeal)
}

// BuildFakeMealCreationRequestInput builds a faked MealCreationRequestInput.
func BuildFakeMealCreationRequestInput() *mealplanning.MealCreationRequestInput {
	meal := BuildFakeMeal()

	return converters.ConvertMealToMealCreationRequestInput(meal)
}
