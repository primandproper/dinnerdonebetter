package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v4/filtering"
)

func BuildFakeMealPlanGroceryListItem() *types.MealPlanGroceryListItem {
	minQty, maxQty := BuildFakeFloat32WithOptionalMax()
	return &types.MealPlanGroceryListItem{
		ID:                       BuildFakeID(),
		BelongsToMealPlan:        BuildFakeID(),
		Ingredient:               *BuildFakeValidIngredient(),
		MeasurementUnit:          *BuildFakeValidMeasurementUnit(),
		MinQuantityNeeded:        minQty,
		MaxQuantityNeeded:        maxQty,
		QuantityPurchased:        nil,
		PurchasedMeasurementUnit: nil,
		PurchasedUPC:             nil,
		PurchasePrice:            nil,
		StatusExplanation:        buildUniqueString(),
		Status:                   types.MealPlanGroceryListItemStatusUnknown,
		CreatedAt:                BuildFakeTime(),
		// Recipe context fields (optional - only set when item is part of a choice group)
		BelongsToMealPlanOption: nil,
		RecipeID:                nil,
		RecipeStepID:            nil,
		IngredientIndex:         nil,
		OptionIndex:             nil,
	}
}

func BuildFakeMealPlanGroceryListItemsList() *filtering.QueryFilteredResult[types.MealPlanGroceryListItem] {
	var examples []*types.MealPlanGroceryListItem
	for range exampleQuantity {
		examples = append(examples, BuildFakeMealPlanGroceryListItem())
	}

	return &filtering.QueryFilteredResult[types.MealPlanGroceryListItem]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

func BuildFakeMealPlanGroceryListItemCreationRequestInput() *types.MealPlanGroceryListItemCreationRequestInput {
	mealPlanGroceryListItem := BuildFakeMealPlanGroceryListItem()
	return converters.ConvertMealPlanGroceryListItemToMealPlanGroceryListItemCreationRequestInput(mealPlanGroceryListItem)
}

func BuildFakeMealPlanGroceryListItemUpdateRequestInput() *types.MealPlanGroceryListItemUpdateRequestInput {
	mealPlanGroceryListItem := BuildFakeMealPlanGroceryListItem()
	return converters.ConvertMealPlanGroceryListItemToMealPlanGroceryListItemUpdateRequestInput(mealPlanGroceryListItem)
}
