package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v5/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeMealPlanOption builds a faked meal plan option.
func BuildFakeMealPlanOption() *types.MealPlanOption {
	var examples []*types.MealPlanOptionVote
	for range exampleQuantity {
		examples = append(examples, BuildFakeMealPlanOptionVote())
	}

	meal := BuildFakeMeal()
	meal.Components = nil

	return &types.MealPlanOption{
		ID:                     BuildFakeID(),
		AssignedCook:           func(s string) *string { return &s }(BuildFakeID()),
		Meal:                   *meal,
		Votes:                  examples,
		Chosen:                 false,
		Notes:                  buildUniqueString(),
		CreatedAt:              BuildFakeTime(),
		BelongsToMealPlanEvent: fake.UUID(),
	}
}

// BuildFakeMealPlanOptionsList builds a faked MealPlanOptionList.
func BuildFakeMealPlanOptionsList() *filtering.QueryFilteredResult[types.MealPlanOption] {
	var examples []*types.MealPlanOption
	for range exampleQuantity {
		examples = append(examples, BuildFakeMealPlanOption())
	}

	return &filtering.QueryFilteredResult[types.MealPlanOption]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeMealPlanOptionUpdateRequestInput builds a faked MealPlanOptionUpdateRequestInput from a meal plan option.
func BuildFakeMealPlanOptionUpdateRequestInput() *types.MealPlanOptionUpdateRequestInput {
	mealPlanOption := BuildFakeMealPlanOption()
	return converters.ConvertMealPlanOptionToMealPlanOptionUpdateRequestInput(mealPlanOption)
}

// BuildFakeMealPlanOptionCreationRequestInput builds a faked MealPlanOptionCreationRequestInput.
func BuildFakeMealPlanOptionCreationRequestInput() *types.MealPlanOptionCreationRequestInput {
	mealPlanOption := BuildFakeMealPlanOption()
	return converters.ConvertMealPlanOptionToMealPlanOptionCreationRequestInput(mealPlanOption)
}
