package converters

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v3/identifiers"
)

// ConvertMealPlanOptionToMealPlanOptionUpdateRequestInput creates a MealPlanOptionUpdateRequestInput from a BelongsToMealPlan.
func ConvertMealPlanOptionToMealPlanOptionUpdateRequestInput(input *types.MealPlanOption) *types.MealPlanOptionUpdateRequestInput {
	x := &types.MealPlanOptionUpdateRequestInput{
		MealID:                 &input.Meal.ID,
		Notes:                  &input.Notes,
		MealScale:              &input.MealScale,
		BelongsToMealPlanEvent: &input.BelongsToMealPlanEvent,
		AssignedCook:           input.AssignedCook,
		AssignedDishwasher:     input.AssignedDishwasher,
	}

	return x
}

// ConvertMealPlanOptionCreationRequestInputToMealPlanOptionDatabaseCreationInput creates a MealPlanOptionDatabaseCreationInput from a MealPlanOptionCreationRequestInput.
func ConvertMealPlanOptionCreationRequestInputToMealPlanOptionDatabaseCreationInput(input *types.MealPlanOptionCreationRequestInput) *types.MealPlanOptionDatabaseCreationInput {
	x := &types.MealPlanOptionDatabaseCreationInput{
		ID:                 identifiers.New(),
		MealID:             input.MealID,
		MealScale:          input.MealScale,
		Notes:              input.Notes,
		AssignedCook:       input.AssignedCook,
		AssignedDishwasher: input.AssignedDishwasher,
	}

	return x
}

// ConvertMealPlanOptionVoteToMealPlanOptionVoteCreationRequestInput builds a MealPlanOptionVoteCreationRequestInput from a meal plan option vote.
func ConvertMealPlanOptionVoteToMealPlanOptionVoteCreationRequestInput(mealPlanOptionVote *types.MealPlanOptionVote) *types.MealPlanOptionVoteCreationRequestInput {
	return &types.MealPlanOptionVoteCreationRequestInput{
		Votes: []*types.MealPlanOptionVoteCreationInput{
			{
				ID:                      mealPlanOptionVote.ID,
				Rank:                    mealPlanOptionVote.Rank,
				Abstain:                 mealPlanOptionVote.Abstain,
				Notes:                   mealPlanOptionVote.Notes,
				BelongsToMealPlanOption: mealPlanOptionVote.BelongsToMealPlanOption,
				ByUser:                  mealPlanOptionVote.ByUser,
			},
		},
	}
}

// ConvertMealPlanOptionVoteToMealPlanOptionVotesDatabaseCreationInput builds a MealPlanOptionVotesDatabaseCreationInput from a meal plan option vote.
func ConvertMealPlanOptionVoteToMealPlanOptionVotesDatabaseCreationInput(mealPlanOptionVote *types.MealPlanOptionVote) *types.MealPlanOptionVotesDatabaseCreationInput {
	return &types.MealPlanOptionVotesDatabaseCreationInput{
		Votes: []*types.MealPlanOptionVoteDatabaseCreationInput{
			{
				ID:                      mealPlanOptionVote.ID,
				Rank:                    mealPlanOptionVote.Rank,
				Abstain:                 mealPlanOptionVote.Abstain,
				Notes:                   mealPlanOptionVote.Notes,
				BelongsToMealPlanOption: mealPlanOptionVote.BelongsToMealPlanOption,
				ByUser:                  mealPlanOptionVote.ByUser,
			},
		},
	}
}

// ConvertMealPlanOptionToMealPlanOptionCreationRequestInput builds a MealPlanOptionCreationRequestInput from a meal plan option.
func ConvertMealPlanOptionToMealPlanOptionCreationRequestInput(mealPlanOption *types.MealPlanOption) *types.MealPlanOptionCreationRequestInput {
	return &types.MealPlanOptionCreationRequestInput{
		MealID:             mealPlanOption.Meal.ID,
		Notes:              mealPlanOption.Notes,
		AssignedCook:       mealPlanOption.AssignedCook,
		AssignedDishwasher: mealPlanOption.AssignedDishwasher,
		MealScale:          mealPlanOption.MealScale,
	}
}

// ConvertMealPlanOptionToMealPlanOptionDatabaseCreationInput builds a MealPlanOptionDatabaseCreationInput from a meal plan option.
func ConvertMealPlanOptionToMealPlanOptionDatabaseCreationInput(mealPlanOption *types.MealPlanOption) *types.MealPlanOptionDatabaseCreationInput {
	return &types.MealPlanOptionDatabaseCreationInput{
		ID:                     mealPlanOption.ID,
		MealID:                 mealPlanOption.Meal.ID,
		Notes:                  mealPlanOption.Notes,
		AssignedCook:           mealPlanOption.AssignedCook,
		AssignedDishwasher:     mealPlanOption.AssignedDishwasher,
		BelongsToMealPlanEvent: mealPlanOption.BelongsToMealPlanEvent,
		MealScale:              mealPlanOption.MealScale,
	}
}
