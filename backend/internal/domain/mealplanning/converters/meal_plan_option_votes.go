package converters

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v11/identifiers"
)

// ConvertMealPlanOptionVoteToMealPlanOptionVoteUpdateRequestInput creates a DatabaseCreationInput from a CreationInput.
func ConvertMealPlanOptionVoteToMealPlanOptionVoteUpdateRequestInput(input *types.MealPlanOptionVote) *types.MealPlanOptionVoteUpdateRequestInput {
	x := &types.MealPlanOptionVoteUpdateRequestInput{
		Notes:                   &input.Notes,
		BelongsToMealPlanOption: input.BelongsToMealPlanOption,
		Rank:                    &input.Rank,
		Abstain:                 &input.Abstain,
	}

	return x
}

// ConvertMealPlanOptionVoteCreationRequestInputToMealPlanOptionVotesDatabaseCreationInput creates a MealPlanOptionVotesDatabaseCreationInput from a MealPlanOptionVoteCreationRequestInput.
func ConvertMealPlanOptionVoteCreationRequestInputToMealPlanOptionVotesDatabaseCreationInput(input *types.MealPlanOptionVoteCreationRequestInput) *types.MealPlanOptionVotesDatabaseCreationInput {
	var votes []*types.MealPlanOptionVoteDatabaseCreationInput
	for _, vote := range input.Votes {
		votes = append(votes, ConvertMealPlanOptionVoteCreationRequestInputToMealPlanOptionVoteDatabaseCreationInput(vote))
	}

	x := &types.MealPlanOptionVotesDatabaseCreationInput{
		Votes: votes,
	}

	return x
}

func ConvertMealPlanOptionVoteCreationRequestInputToMealPlanOptionVoteDatabaseCreationInput(input *types.MealPlanOptionVoteCreationInput) *types.MealPlanOptionVoteDatabaseCreationInput {
	return &types.MealPlanOptionVoteDatabaseCreationInput{
		ID:                      identifiers.New(),
		Notes:                   input.Notes,
		ByUser:                  input.ByUser,
		BelongsToMealPlanOption: input.BelongsToMealPlanOption,
		Rank:                    input.Rank,
		Abstain:                 input.Abstain,
	}
}
