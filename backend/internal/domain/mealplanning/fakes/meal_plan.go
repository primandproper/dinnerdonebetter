package fakes

import (
	"time"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v8/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeMealPlan builds a faked meal plan.
func BuildFakeMealPlan() *types.MealPlan {
	mealPlanID := BuildFakeID()

	var events []*types.MealPlanEvent
	for range exampleQuantity {
		event := BuildFakeMealPlanEvent()
		event.BelongsToMealPlan = mealPlanID
		events = append(events, event)
	}

	// The voting deadline must be in the future but before every event's start time (events start in
	// ten minutes, see BuildFakeMealPlanEvent), so the meal plan passes MealPlanCreationRequestInput validation.
	votingDeadline := time.Now().Add(5 * time.Minute).Truncate(time.Second).UTC()
	return &types.MealPlan{
		ID:                     mealPlanID,
		Notes:                  buildUniqueString(),
		Status:                 string(types.MealPlanStatusAwaitingVotes),
		VotingDeadline:         votingDeadline,
		CreatedAt:              BuildFakeTime(),
		BelongsToAccount:       fake.UUID(),
		TasksCreated:           false,
		GroceryListInitialized: false,
		ElectionMethod:         types.MealPlanElectionMethodSchulze,
		Events:                 events,
		CreatedByUser:          BuildFakeID(),
	}
}

// BuildFakeMealPlansList builds a faked MealPlanList.
func BuildFakeMealPlansList() *filtering.QueryFilteredResult[types.MealPlan] {
	var examples []*types.MealPlan
	for range exampleQuantity {
		examples = append(examples, BuildFakeMealPlan())
	}

	return &filtering.QueryFilteredResult[types.MealPlan]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
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
