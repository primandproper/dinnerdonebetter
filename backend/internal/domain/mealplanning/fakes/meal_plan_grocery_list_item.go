package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
)

// BuildFakeMealPlanGroceryListItem builds a faked meal plan grocery list item.
//
// Everything the type leaves optional stays absent: an item nobody has bought yet, and
// one that belongs to the plan rather than to a particular option's choice group.
func BuildFakeMealPlanGroceryListItem() *types.MealPlanGroceryListItem {
	item := fake.BuildFakeRecord[types.MealPlanGroceryListItem]()

	// What to buy and what it is measured in.
	item.Ingredient = *BuildFakeValidIngredient()
	item.MeasurementUnit = *BuildFakeValidMeasurementUnit()
	item.MinQuantityNeeded, item.MaxQuantityNeeded = BuildFakeFloat32WithOptionalMax()

	// One of the statuses the type validates against.
	item.Status = types.MealPlanGroceryListItemStatusUnknown

	return item
}

// BuildFakeMealPlanGroceryListItemsList builds a faked MealPlanGroceryListItemList.
func BuildFakeMealPlanGroceryListItemsList() *filtering.QueryFilteredResult[types.MealPlanGroceryListItem] {
	return fake.BuildFakePage(BuildFakeMealPlanGroceryListItem)
}

// BuildFakeMealPlanGroceryListItemCreationRequestInput builds a faked MealPlanGroceryListItemCreationRequestInput.
func BuildFakeMealPlanGroceryListItemCreationRequestInput() *types.MealPlanGroceryListItemCreationRequestInput {
	mealPlanGroceryListItem := BuildFakeMealPlanGroceryListItem()

	return converters.ConvertMealPlanGroceryListItemToMealPlanGroceryListItemCreationRequestInput(mealPlanGroceryListItem)
}

// BuildFakeMealPlanGroceryListItemUpdateRequestInput builds a faked MealPlanGroceryListItemUpdateRequestInput.
func BuildFakeMealPlanGroceryListItemUpdateRequestInput() *types.MealPlanGroceryListItemUpdateRequestInput {
	mealPlanGroceryListItem := BuildFakeMealPlanGroceryListItem()

	return converters.ConvertMealPlanGroceryListItemToMealPlanGroceryListItemUpdateRequestInput(mealPlanGroceryListItem)
}
