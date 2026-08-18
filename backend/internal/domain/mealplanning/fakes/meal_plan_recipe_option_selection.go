package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/pointer"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeMealPlanRecipeOptionSelection builds a faked meal plan recipe option selection.
func BuildFakeMealPlanRecipeOptionSelection() *types.MealPlanRecipeOptionSelection {
	selection := fake.BuildFakeRecord[types.MealPlanRecipeOptionSelection]()

	// One of the two things a selection can be about, which the type validates.
	selection.SelectionType = types.MealPlanRecipeOptionSelectionTypeIngredient

	return selection
}

// BuildFakeMealPlanRecipeOptionSelectionsList builds a faked MealPlanRecipeOptionSelectionsList.
func BuildFakeMealPlanRecipeOptionSelectionsList() *filtering.QueryFilteredResult[types.MealPlanRecipeOptionSelection] {
	return fake.BuildFakePage(BuildFakeMealPlanRecipeOptionSelection)
}

// BuildFakeMealPlanRecipeOptionSelectionDatabaseCreationInput builds a faked MealPlanRecipeOptionSelectionDatabaseCreationInput.
func BuildFakeMealPlanRecipeOptionSelectionDatabaseCreationInput() *types.MealPlanRecipeOptionSelectionDatabaseCreationInput {
	selection := BuildFakeMealPlanRecipeOptionSelection()

	return converters.ConvertMealPlanRecipeOptionSelectionToMealPlanRecipeOptionSelectionDatabaseCreationInput(selection)
}

// BuildFakeMealPlanRecipeOptionSelectionUpdateRequestInput builds a faked MealPlanRecipeOptionSelectionUpdateRequestInput.
//
// Its one field is optional, and an update that changes nothing is one no assertion can
// see, so it is filled here rather than left to BuildFakeRecord.
func BuildFakeMealPlanRecipeOptionSelectionUpdateRequestInput() *types.MealPlanRecipeOptionSelectionUpdateRequestInput {
	return &types.MealPlanRecipeOptionSelectionUpdateRequestInput{
		SelectedOptionIndex: pointer.To(gofakeit.Uint16()),
	}
}

// BuildFakeMealPlanRecipeOptionSelectionCreationRequestInput builds a faked MealPlanRecipeOptionSelectionCreationRequestInput.
func BuildFakeMealPlanRecipeOptionSelectionCreationRequestInput() *types.MealPlanRecipeOptionSelectionCreationRequestInput {
	input := fake.BuildFakeRecord[types.MealPlanRecipeOptionSelectionCreationRequestInput]()

	input.SelectionType = types.MealPlanRecipeOptionSelectionTypeIngredient

	// The first ingredient and its first option, which is what a recipe with one option
	// group per ingredient offers.
	input.IngredientIndex = 0
	input.SelectedOptionIndex = 0

	return input
}
