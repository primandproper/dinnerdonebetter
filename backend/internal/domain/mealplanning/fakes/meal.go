package fakes

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v7/filtering"
)

// BuildFakeMeal builds a faked meal.
func BuildFakeMeal() *mealplanning.Meal {
	recipes := []*mealplanning.MealComponent{}
	for range exampleQuantity {
		recipes = append(recipes, BuildFakeMealComponent())
	}

	return &mealplanning.Meal{
		ID:                   BuildFakeID(),
		Name:                 buildUniqueString(),
		Description:          buildUniqueString(),
		MinEstimatedPortions: float32(buildFakeNumber()),
		MaxEstimatedPortions: nil,
		CreatedAt:            BuildFakeTime(),
		CreatedByUser:        BuildFakeID(),
		Components:           recipes,
		EligibleForMealPlans: true,
	}
}

// BuildFakeMealComponent builds a faked meal component.
func BuildFakeMealComponent() *mealplanning.MealComponent {
	return &mealplanning.MealComponent{
		Recipe:        *BuildFakeRecipe(),
		RecipeScale:   float32(1.0),
		ComponentType: mealplanning.MealComponentTypesMain,
	}
}

// BuildFakeMealsList builds a faked MealList.
func BuildFakeMealsList() *filtering.QueryFilteredResult[mealplanning.Meal] {
	var examples []*mealplanning.Meal
	for range exampleQuantity {
		examples = append(examples, BuildFakeMeal())
	}

	return &filtering.QueryFilteredResult[mealplanning.Meal]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeMealCreationRequestInput builds a faked MealCreationRequestInput.
func BuildFakeMealCreationRequestInput() *mealplanning.MealCreationRequestInput {
	recipe := BuildFakeMeal()
	return converters.ConvertMealToMealCreationRequestInput(recipe)
}
