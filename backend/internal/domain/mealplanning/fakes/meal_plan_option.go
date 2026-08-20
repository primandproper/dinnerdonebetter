package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/pointer"
)

// BuildFakeMealPlanOption builds a faked meal plan option.
func BuildFakeMealPlanOption() *types.MealPlanOption {
	option := fake.BuildFakeRecord[types.MealPlanOption]()

	// An option nobody has chosen yet, which is what voting decides. TieBroken is
	// pinned for a second reason as well: there is no tie_broken column anywhere in the
	// schema, so the field cannot survive a write and reads back false regardless. A
	// random value here fails every round-trip assertion about half the time.
	option.Chosen = false
	option.TieBroken = false
	option.AssignedCook = pointer.To(fake.BuildFakeID())

	// The meal on offer, without its components: an option is read alongside every other
	// option of its event, and each component drags a whole recipe graph behind it.
	meal := BuildFakeMeal()
	meal.Components = nil
	option.Meal = *meal

	votes := make([]*types.MealPlanOptionVote, 0, exampleQuantity)
	for range exampleQuantity {
		votes = append(votes, BuildFakeMealPlanOptionVote())
	}
	option.Votes = votes

	return option
}

// BuildFakeMealPlanOptionsList builds a faked MealPlanOptionList.
func BuildFakeMealPlanOptionsList() *filtering.QueryFilteredResult[types.MealPlanOption] {
	return fake.BuildFakePage(BuildFakeMealPlanOption)
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
