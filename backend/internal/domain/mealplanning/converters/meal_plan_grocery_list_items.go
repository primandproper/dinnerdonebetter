package converters

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/identifiers"
)

// ConvertMealPlanGroceryListItemToMealPlanGroceryListItemDatabaseCreationInput builds a MealPlanGroceryListItemDatabaseCreationInput from a MealPlanGroceryListItem.
func ConvertMealPlanGroceryListItemToMealPlanGroceryListItemDatabaseCreationInput(input *mealplanning.MealPlanGroceryListItem) *mealplanning.MealPlanGroceryListItemDatabaseCreationInput {
	x := &mealplanning.MealPlanGroceryListItemDatabaseCreationInput{
		ID:                      input.ID,
		BelongsToMealPlan:       input.BelongsToMealPlan,
		ValidIngredientID:       input.Ingredient.ID,
		ValidMeasurementUnitID:  input.MeasurementUnit.ID,
		MinQuantityNeeded:       input.MinQuantityNeeded,
		MaxQuantityNeeded:       input.MaxQuantityNeeded,
		QuantityPurchased:       input.QuantityPurchased,
		PurchasedUPC:            input.PurchasedUPC,
		PurchasePrice:           input.PurchasePrice,
		StatusExplanation:       input.StatusExplanation,
		Status:                  input.Status,
		BelongsToMealPlanOption: input.BelongsToMealPlanOption,
		RecipeID:                input.RecipeID,
		RecipeStepID:            input.RecipeStepID,
		IngredientIndex:         input.IngredientIndex,
		OptionIndex:             input.OptionIndex,
	}

	if input.PurchasedMeasurementUnit != nil {
		x.PurchasedMeasurementUnitID = &input.PurchasedMeasurementUnit.ID
	}

	return x
}

// ConvertMealPlanGroceryListItemToMealPlanGroceryListItemCreationRequestInput builds a MealPlanGroceryListItemCreationRequestInput from a MealPlanGroceryListItem.
func ConvertMealPlanGroceryListItemToMealPlanGroceryListItemCreationRequestInput(input *mealplanning.MealPlanGroceryListItem) *mealplanning.MealPlanGroceryListItemCreationRequestInput {
	x := &mealplanning.MealPlanGroceryListItemCreationRequestInput{
		PurchasedUPC:           input.PurchasedUPC,
		PurchasePrice:          input.PurchasePrice,
		QuantityPurchased:      input.QuantityPurchased,
		StatusExplanation:      input.StatusExplanation,
		Status:                 input.Status,
		BelongsToMealPlan:      input.BelongsToMealPlan,
		ValidIngredientID:      input.Ingredient.ID,
		ValidMeasurementUnitID: input.MeasurementUnit.ID,
		MinQuantityNeeded:      input.MinQuantityNeeded,
		MaxQuantityNeeded:      input.MaxQuantityNeeded,
	}

	if input.PurchasedMeasurementUnit != nil {
		x.PurchasedMeasurementUnitID = &input.PurchasedMeasurementUnit.ID
	}

	return x
}

func ConvertMealPlanGroceryListItemCreationRequestInputToMealPlanGroceryListItemDatabaseCreationInput(input *mealplanning.MealPlanGroceryListItemCreationRequestInput) *mealplanning.MealPlanGroceryListItemDatabaseCreationInput {
	return &mealplanning.MealPlanGroceryListItemDatabaseCreationInput{
		ID:                         identifiers.New(),
		PurchasePrice:              input.PurchasePrice,
		PurchasedUPC:               input.PurchasedUPC,
		PurchasedMeasurementUnitID: input.PurchasedMeasurementUnitID,
		QuantityPurchased:          input.QuantityPurchased,
		Status:                     input.Status,
		StatusExplanation:          input.StatusExplanation,
		ValidMeasurementUnitID:     input.ValidMeasurementUnitID,
		ValidIngredientID:          input.ValidIngredientID,
		BelongsToMealPlan:          input.BelongsToMealPlan,
		MinQuantityNeeded:          input.MinQuantityNeeded,
		MaxQuantityNeeded:          input.MaxQuantityNeeded,
		// Recipe context fields are not included in creation request input
		// They are set separately when creating items with alternatives
	}
}

func ConvertMealPlanGroceryListItemToMealPlanGroceryListItemUpdateRequestInput(input *mealplanning.MealPlanGroceryListItem) *mealplanning.MealPlanGroceryListItemUpdateRequestInput {
	var purchasedMeasurementUnitID *string
	if input.PurchasedMeasurementUnit != nil {
		purchasedMeasurementUnitID = &input.PurchasedMeasurementUnit.ID
	}

	return &mealplanning.MealPlanGroceryListItemUpdateRequestInput{
		BelongsToMealPlan:          &input.BelongsToMealPlan,
		ValidIngredientID:          &input.Ingredient.ID,
		ValidMeasurementUnitID:     &input.MeasurementUnit.ID,
		MinQuantityNeeded:          &input.MinQuantityNeeded,
		MaxQuantityNeeded:          input.MaxQuantityNeeded,
		QuantityPurchased:          input.QuantityPurchased,
		PurchasedMeasurementUnitID: purchasedMeasurementUnitID,
		PurchasedUPC:               input.PurchasedUPC,
		PurchasePrice:              input.PurchasePrice,
		StatusExplanation:          &input.StatusExplanation,
		Status:                     &input.Status,
	}
}
