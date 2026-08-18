package fakes

import (
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeMealPlanEvent builds a faked meal plan event.
func BuildFakeMealPlanEvent() *types.MealPlanEvent {
	event := fake.BuildFakeRecord[types.MealPlanEvent]()

	// An event that has not happened yet and lasts a week. Ten minutes out because a
	// meal plan's voting deadline has to fall between now and the first event's start,
	// and five minutes is what BuildFakeMealPlan picks.
	now := time.Now().Truncate(time.Second).UTC()
	event.StartsAt = now.Add(10 * time.Minute)
	event.EndsAt = now.Add(7 * 24 * time.Hour)

	// One of the meals of the day the domain names.
	event.MealName = gofakeit.RandomString([]string{
		types.BreakfastMealName,
		types.SecondBreakfastMealName,
		types.BrunchMealName,
		types.LunchMealName,
		types.SupperMealName,
		types.DinnerMealName,
	})

	// Options of this event rather than of three unrelated ones — an event's options are
	// what its votes are cast against.
	options := []*types.MealPlanOption{}
	for _, opt := range BuildFakeMealPlanOptionsList().Data {
		opt.BelongsToMealPlanEvent = event.ID
		options = append(options, opt)
	}
	event.Options = options

	return event
}

// BuildFakeMealPlanEventsList builds a faked MealPlanEventList.
func BuildFakeMealPlanEventsList() *filtering.QueryFilteredResult[types.MealPlanEvent] {
	return fake.BuildFakePage(BuildFakeMealPlanEvent)
}

// BuildFakeMealPlanEventUpdateRequestInput builds a faked MealPlanEventUpdateRequestInput from a meal plan event.
func BuildFakeMealPlanEventUpdateRequestInput() *types.MealPlanEventUpdateRequestInput {
	mealPlanEvent := BuildFakeMealPlanEvent()

	return &types.MealPlanEventUpdateRequestInput{
		Notes:             &mealPlanEvent.Notes,
		StartsAt:          &mealPlanEvent.StartsAt,
		EndsAt:            &mealPlanEvent.EndsAt,
		MealName:          &mealPlanEvent.MealName,
		BelongsToMealPlan: mealPlanEvent.BelongsToMealPlan,
	}
}

// BuildFakeMealPlanEventCreationRequestInput builds a faked MealPlanEventCreationRequestInput.
func BuildFakeMealPlanEventCreationRequestInput() *types.MealPlanEventCreationRequestInput {
	mealPlanEvent := BuildFakeMealPlanEvent()

	return converters.ConvertMealPlanEventToMealPlanEventCreationRequestInput(mealPlanEvent)
}
