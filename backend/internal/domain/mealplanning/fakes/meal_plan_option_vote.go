package fakes

import (
	"math"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeMealPlanOptionVote builds a faked meal plan option vote.
func BuildFakeMealPlanOptionVote() *types.MealPlanOptionVote {
	vote := fake.BuildFakeRecord[types.MealPlanOptionVote]()

	// A rank rather than a number: the Schulze count reads it as a preference order, and
	// rank zero is not a preference.
	vote.Rank = uint8(gofakeit.Number(1, math.MaxUint8))

	return vote
}

// BuildFakeMealPlanOptionVotesList builds a faked MealPlanOptionVoteList.
func BuildFakeMealPlanOptionVotesList() *filtering.QueryFilteredResult[types.MealPlanOptionVote] {
	return fake.BuildFakePage(BuildFakeMealPlanOptionVote)
}

// BuildFakeMealPlanOptionVoteUpdateRequestInput builds a faked MealPlanOptionVoteUpdateRequestInput from a meal plan option vote.
func BuildFakeMealPlanOptionVoteUpdateRequestInput() *types.MealPlanOptionVoteUpdateRequestInput {
	mealPlanOptionVote := BuildFakeMealPlanOptionVote()

	return &types.MealPlanOptionVoteUpdateRequestInput{
		Rank:                    &mealPlanOptionVote.Rank,
		Abstain:                 &mealPlanOptionVote.Abstain,
		Notes:                   &mealPlanOptionVote.Notes,
		BelongsToMealPlanOption: mealPlanOptionVote.BelongsToMealPlanOption,
	}
}

// BuildFakeMealPlanOptionVoteCreationRequestInput builds a faked MealPlanOptionVoteCreationRequestInput.
func BuildFakeMealPlanOptionVoteCreationRequestInput() *types.MealPlanOptionVoteCreationRequestInput {
	mealPlanOptionVote := BuildFakeMealPlanOptionVote()

	return converters.ConvertMealPlanOptionVoteToMealPlanOptionVoteCreationRequestInput(mealPlanOptionVote)
}

// BuildFakeMealPlanOptionVoteDatabaseCreationInput builds a faked MealPlanOptionVotesDatabaseCreationInput.
func BuildFakeMealPlanOptionVoteDatabaseCreationInput() *types.MealPlanOptionVotesDatabaseCreationInput {
	mealPlanOptionVote := BuildFakeMealPlanOptionVote()

	return converters.ConvertMealPlanOptionVoteToMealPlanOptionVotesDatabaseCreationInput(mealPlanOptionVote)
}
