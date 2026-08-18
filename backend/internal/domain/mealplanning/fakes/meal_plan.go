package fakes

import (
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
)

// BuildFakeMealPlan builds a faked meal plan.
func BuildFakeMealPlan() *types.MealPlan {
	mealPlan := fake.BuildFakeRecord[types.MealPlan]()

	// A plan that is open for votes, which is where every other state is reached from.
	mealPlan.Status = string(types.MealPlanStatusAwaitingVotes)
	mealPlan.ElectionMethod = types.MealPlanElectionMethodSchulze
	mealPlan.TasksCreated = false
	mealPlan.GroceryListInitialized = false

	// The voting deadline must be in the future but before every event's start time
	// (events start in ten minutes, see BuildFakeMealPlanEvent), so the meal plan passes
	// MealPlanCreationRequestInput validation.
	mealPlan.VotingDeadline = time.Now().Add(5 * time.Minute).Truncate(time.Second).UTC()

	// Events of this plan rather than of three unrelated ones.
	events := make([]*types.MealPlanEvent, 0, exampleQuantity)
	for range exampleQuantity {
		event := BuildFakeMealPlanEvent()
		event.BelongsToMealPlan = mealPlan.ID
		events = append(events, event)
	}
	mealPlan.Events = events

	return mealPlan
}

// BuildFakeMealPlansList builds a faked MealPlanList.
func BuildFakeMealPlansList() *filtering.QueryFilteredResult[types.MealPlan] {
	return fake.BuildFakePage(BuildFakeMealPlan)
}

// BuildFakeMealPlanUpdateRequestInput builds a faked MealPlanUpdateRequestInput from a meal plan.
func BuildFakeMealPlanUpdateRequestInput() *types.MealPlanUpdateRequestInput {
	mealPlan := BuildFakeMealPlan()

	return converters.ConvertMealPlanToMealPlanUpdateRequestInput(mealPlan)
}

// BuildFakeMealPlanCreationRequestInput builds a faked MealPlanCreationRequestInput.
func BuildFakeMealPlanCreationRequestInput() *types.MealPlanCreationRequestInput {
	mealPlan := BuildFakeMealPlan()

	return converters.ConvertMealPlanToMealPlanCreationRequestInput(mealPlan)
}
