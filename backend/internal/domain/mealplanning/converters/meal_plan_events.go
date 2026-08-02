package converters

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/identifiers"
)

// ConvertMealPlanEventToMealPlanEventUpdateRequestInput creates a MealPlanEventUpdateRequestInput from a MealPlanEvent.
func ConvertMealPlanEventToMealPlanEventUpdateRequestInput(input *types.MealPlanEvent) *types.MealPlanEventUpdateRequestInput {
	x := &types.MealPlanEventUpdateRequestInput{
		BelongsToMealPlan: input.BelongsToMealPlan,
		Notes:             &input.Notes,
		StartsAt:          &input.StartsAt,
		EndsAt:            &input.EndsAt,
		MealName:          &input.MealName,
	}

	return x
}

// ConvertMealPlanEventCreationRequestInputToMealPlanEventDatabaseCreationInput creates a MealPlanEventDatabaseCreationInput from a MealPlanEventCreationRequestInput.
func ConvertMealPlanEventCreationRequestInputToMealPlanEventDatabaseCreationInput(input *types.MealPlanEventCreationRequestInput) *types.MealPlanEventDatabaseCreationInput {
	options := []*types.MealPlanOptionDatabaseCreationInput{}
	for _, option := range input.Options {
		options = append(options, ConvertMealPlanOptionCreationRequestInputToMealPlanOptionDatabaseCreationInput(option))
	}

	x := &types.MealPlanEventDatabaseCreationInput{
		ID:       identifiers.New(),
		Notes:    input.Notes,
		StartsAt: input.StartsAt,
		EndsAt:   input.EndsAt,
		MealName: input.MealName,
		Options:  options,
	}

	return x
}

// ConvertMealPlanEventToMealPlanEventCreationRequestInput builds a MealPlanEventCreationRequestInput from a meal plan.
func ConvertMealPlanEventToMealPlanEventCreationRequestInput(mealPlanEvent *types.MealPlanEvent) *types.MealPlanEventCreationRequestInput {
	options := []*types.MealPlanOptionCreationRequestInput{}
	for _, opt := range mealPlanEvent.Options {
		opt.BelongsToMealPlanEvent = mealPlanEvent.ID
		options = append(options, ConvertMealPlanOptionToMealPlanOptionCreationRequestInput(opt))
	}

	return &types.MealPlanEventCreationRequestInput{
		Notes:    mealPlanEvent.Notes,
		StartsAt: mealPlanEvent.StartsAt,
		EndsAt:   mealPlanEvent.EndsAt,
		MealName: mealPlanEvent.MealName,
		Options:  options,
	}
}

// ConvertMealPlanEventToMealPlanEventDatabaseCreationInput builds a MealPlanEventDatabaseCreationInput from a meal plan.
func ConvertMealPlanEventToMealPlanEventDatabaseCreationInput(mealPlanEvent *types.MealPlanEvent) *types.MealPlanEventDatabaseCreationInput {
	options := []*types.MealPlanOptionDatabaseCreationInput{}
	for _, option := range mealPlanEvent.Options {
		options = append(options, ConvertMealPlanOptionToMealPlanOptionDatabaseCreationInput(option))
	}

	return &types.MealPlanEventDatabaseCreationInput{
		ID:                mealPlanEvent.ID,
		Notes:             mealPlanEvent.Notes,
		StartsAt:          mealPlanEvent.StartsAt,
		EndsAt:            mealPlanEvent.EndsAt,
		MealName:          mealPlanEvent.MealName,
		Options:           options,
		BelongsToMealPlan: mealPlanEvent.BelongsToMealPlan,
	}
}
