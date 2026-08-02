package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v9/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeMealPlanRecipeOptionSelection builds a faked meal plan recipe option selection.
func BuildFakeMealPlanRecipeOptionSelection() *types.MealPlanRecipeOptionSelection {
	return &types.MealPlanRecipeOptionSelection{
		ID:                      BuildFakeID(),
		BelongsToMealPlanOption: BuildFakeID(),
		RecipeID:                BuildFakeID(),
		RecipeStepID:            BuildFakeID(),
		IngredientIndex:         fake.Uint16(),
		SelectedOptionIndex:     fake.Uint16(),
		SelectionType:           types.MealPlanRecipeOptionSelectionTypeIngredient,
		CreatedAt:               BuildFakeTime(),
		LastUpdatedAt:           nil,
	}
}

// BuildFakeMealPlanRecipeOptionSelectionsList builds a faked MealPlanRecipeOptionSelectionsList.
func BuildFakeMealPlanRecipeOptionSelectionsList() *filtering.QueryFilteredResult[types.MealPlanRecipeOptionSelection] {
	var examples []*types.MealPlanRecipeOptionSelection
	for range exampleQuantity {
		examples = append(examples, BuildFakeMealPlanRecipeOptionSelection())
	}

	return &filtering.QueryFilteredResult[types.MealPlanRecipeOptionSelection]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeMealPlanRecipeOptionSelectionDatabaseCreationInput builds a faked MealPlanRecipeOptionSelectionDatabaseCreationInput.
func BuildFakeMealPlanRecipeOptionSelectionDatabaseCreationInput() *types.MealPlanRecipeOptionSelectionDatabaseCreationInput {
	selection := BuildFakeMealPlanRecipeOptionSelection()
	return converters.ConvertMealPlanRecipeOptionSelectionToMealPlanRecipeOptionSelectionDatabaseCreationInput(selection)
}

// BuildFakeMealPlanRecipeOptionSelectionUpdateRequestInput builds a faked MealPlanRecipeOptionSelectionUpdateRequestInput.
func BuildFakeMealPlanRecipeOptionSelectionUpdateRequestInput() *types.MealPlanRecipeOptionSelectionUpdateRequestInput {
	selectedOptionIndex := fake.Uint16()
	return &types.MealPlanRecipeOptionSelectionUpdateRequestInput{
		SelectedOptionIndex: &selectedOptionIndex,
	}
}

func BuildFakeMealPlanRecipeOptionSelectionCreationRequestInput() *types.MealPlanRecipeOptionSelectionCreationRequestInput {
	return &types.MealPlanRecipeOptionSelectionCreationRequestInput{
		RecipeID:            BuildFakeID(),
		RecipeStepID:        BuildFakeID(),
		SelectionType:       types.MealPlanRecipeOptionSelectionTypeIngredient,
		IngredientIndex:     0,
		SelectedOptionIndex: 0,
	}
}
