package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mpconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkMealPlanOptionVoteEquality(t *testing.T, expected, actual *mealplanning.MealPlanOptionVote) {
	t.Helper()

	assert.NotZero(t, actual.ID)
	assert.Equal(t, expected.Rank, actual.Rank, "expected Rank for meal plan option vote %s to be %v, but it was %v", expected.ID, expected.Rank, actual.Rank)
	assert.Equal(t, expected.Abstain, actual.Abstain, "expected Abstain for meal plan option vote %s to be %v, but it was %v", expected.ID, expected.Abstain, actual.Abstain)
	assert.Equal(t, expected.Notes, actual.Notes, "expected StatusExplanation for meal plan option vote %s to be %v, but it was %v", expected.ID, expected.Notes, actual.Notes)
	assert.NotZero(t, actual.CreatedAt)
}

func TestMealPlanOptionVotes_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("should CRUD", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		createdMealPlan := createMealPlanForTest(t, userClient, nil)

		require.NotEmpty(t, createdMealPlan.Events)
		require.NotEmpty(t, createdMealPlan.Events[0].Options)

		createdMealPlanEvent := createdMealPlan.Events[0]
		createdMealPlanOption := createdMealPlanEvent.Options[0]
		require.NotNil(t, createdMealPlanOption)

		exampleMealPlanOptionVote := fakes.BuildFakeMealPlanOptionVote()
		exampleMealPlanOptionVote.BelongsToMealPlanOption = createdMealPlanOption.ID
		exampleMealPlanOptionVoteInput := mpconverters.ConvertMealPlanOptionVoteToMealPlanOptionVoteCreationRequestInput(exampleMealPlanOptionVote)
		createdMealPlanOptionVotesRes, createErr := userClient.CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
			Input:           converters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(exampleMealPlanOptionVoteInput),
		})
		require.NoError(t, createErr)

		createdMealPlanOptionVotes := []*mealplanning.MealPlanOptionVote{}
		for _, createdMealPlanOptionVote := range createdMealPlanOptionVotesRes.Created {
			createdMealPlanOptionVotes = append(createdMealPlanOptionVotes, converters.ConvertGRPCMealPlanOptionVoteToMealPlanOptionVote(createdMealPlanOptionVote))
		}

		for _, createdMealPlanOptionVote := range createdMealPlanOptionVotes {
			checkMealPlanOptionVoteEquality(t, exampleMealPlanOptionVote, createdMealPlanOptionVote)

			retrievedMealPlanOptionVoteRes, err := userClient.GetMealPlanOptionVote(ctx, &mealplanninggrpc.GetMealPlanOptionVoteRequest{
				MealPlanId:           createdMealPlan.ID,
				MealPlanEventId:      createdMealPlanEvent.ID,
				MealPlanOptionId:     createdMealPlanOption.ID,
				MealPlanOptionVoteId: createdMealPlanOptionVote.ID,
			})
			require.NoError(t, err)
			require.Equal(t, createdMealPlanOption.ID, createdMealPlanOptionVote.BelongsToMealPlanOption)

			createdMealPlanOptionVote = converters.ConvertGRPCMealPlanOptionVoteToMealPlanOptionVote(retrievedMealPlanOptionVoteRes.Result)
			checkMealPlanOptionVoteEquality(t, exampleMealPlanOptionVote, createdMealPlanOptionVote)

			newMealPlanOptionVote := fakes.BuildFakeMealPlanOptionVote()
			updateInput := mpconverters.ConvertMealPlanOptionVoteToMealPlanOptionVoteUpdateRequestInput(newMealPlanOptionVote)
			createdMealPlanOptionVote.Update(updateInput)

			_, err = userClient.UpdateMealPlanOptionVote(ctx, &mealplanninggrpc.UpdateMealPlanOptionVoteRequest{
				MealPlanId:           createdMealPlan.ID,
				MealPlanEventId:      createdMealPlanEvent.ID,
				MealPlanOptionId:     createdMealPlanOption.ID,
				MealPlanOptionVoteId: createdMealPlanOptionVote.ID,
				Input:                converters.ConvertMealPlanOptionVoteUpdateRequestInputToGRPCMealPlanOptionVoteUpdateRequestInput(updateInput),
			})
			assert.NoError(t, err)

			actualRes, err := userClient.GetMealPlanOptionVote(ctx, &mealplanninggrpc.GetMealPlanOptionVoteRequest{
				MealPlanId:           createdMealPlan.ID,
				MealPlanEventId:      createdMealPlanEvent.ID,
				MealPlanOptionId:     createdMealPlanOption.ID,
				MealPlanOptionVoteId: createdMealPlanOptionVote.ID,
			})
			require.NoError(t, err)

			actual := converters.ConvertGRPCMealPlanOptionVoteToMealPlanOptionVote(actualRes.Result)

			// assert meal plan option vote equality
			checkMealPlanOptionVoteEquality(t, newMealPlanOptionVote, actual)
			assert.NotNil(t, actual.LastUpdatedAt)

			_, err = userClient.ArchiveMealPlanOptionVote(ctx, &mealplanninggrpc.ArchiveMealPlanOptionVoteRequest{
				MealPlanId:           createdMealPlan.ID,
				MealPlanEventId:      createdMealPlanEvent.ID,
				MealPlanOptionId:     createdMealPlanOption.ID,
				MealPlanOptionVoteId: createdMealPlanOptionVote.ID,
			})
			assert.NoError(t, err)
		}

		_, err := userClient.ArchiveMealPlanOption(ctx, &mealplanninggrpc.ArchiveMealPlanOptionRequest{
			MealPlanId:       createdMealPlan.ID,
			MealPlanEventId:  createdMealPlanEvent.ID,
			MealPlanOptionId: createdMealPlanOption.ID,
		})
		require.NoError(t, err)

		_, err = userClient.ArchiveMealPlanEvent(ctx, &mealplanninggrpc.ArchiveMealPlanEventRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
		})
		require.NoError(t, err)

		_, err = userClient.ArchiveMealPlan(ctx, &mealplanninggrpc.ArchiveMealPlanRequest{MealPlanId: createdMealPlan.ID})
		require.NoError(t, err)
	})
}

func TestMealPlanOptionVotes_Listing(T *testing.T) {
	T.Parallel()

	T.Run("should be readable in paginated form", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		createdMealPlan := createMealPlanForTest(t, userClient, nil)

		require.NotEmpty(t, createdMealPlan.Events)
		require.NotEmpty(t, createdMealPlan.Events[0].Options)

		createdMealPlanEvent := createdMealPlan.Events[0]
		createdMealPlanOption := createdMealPlanEvent.Options[0]
		require.NotNil(t, createdMealPlanOption)

		exampleMealPlanOptionVote := fakes.BuildFakeMealPlanOptionVote()
		exampleMealPlanOptionVote.BelongsToMealPlanOption = createdMealPlanOption.ID
		exampleMealPlanOptionVoteInput := mpconverters.ConvertMealPlanOptionVoteToMealPlanOptionVoteCreationRequestInput(exampleMealPlanOptionVote)
		createdMealPlanOptionVotesRes, createErr := userClient.CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
			Input:           converters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(exampleMealPlanOptionVoteInput),
		})
		require.NoError(t, createErr)

		createdMealPlanOptionVotes := []*mealplanning.MealPlanOptionVote{}
		for _, createdMealPlanOptionVote := range createdMealPlanOptionVotesRes.Created {
			createdMealPlanOptionVotes = append(createdMealPlanOptionVotes, converters.ConvertGRPCMealPlanOptionVoteToMealPlanOptionVote(createdMealPlanOptionVote))
		}

		for _, createdMealPlanOptionVote := range createdMealPlanOptionVotes {
			checkMealPlanOptionVoteEquality(t, exampleMealPlanOptionVote, createdMealPlanOptionVote)

			retrievedMealPlanOptionVoteRes, err := userClient.GetMealPlanOptionVote(ctx, &mealplanninggrpc.GetMealPlanOptionVoteRequest{
				MealPlanId:           createdMealPlan.ID,
				MealPlanEventId:      createdMealPlanEvent.ID,
				MealPlanOptionId:     createdMealPlanOption.ID,
				MealPlanOptionVoteId: createdMealPlanOptionVote.ID,
			})
			require.NoError(t, err)

			createdMealPlanOptionVote = converters.ConvertGRPCMealPlanOptionVoteToMealPlanOptionVote(retrievedMealPlanOptionVoteRes.Result)
			require.Equal(t, createdMealPlanOption.ID, createdMealPlanOptionVote.BelongsToMealPlanOption)

			checkMealPlanOptionVoteEquality(t, exampleMealPlanOptionVote, createdMealPlanOptionVote)

			// assert meal plan option vote list equality
			actual, err := userClient.GetMealPlanOptionVotes(ctx, &mealplanninggrpc.GetMealPlanOptionVotesRequest{
				Filter:           nil,
				MealPlanId:       createdMealPlan.ID,
				MealPlanEventId:  createdMealPlanEvent.ID,
				MealPlanOptionId: createdMealPlanOption.ID,
			})
			require.NoError(t, err)
			assert.NotEmpty(t, actual.Results)

			_, err = userClient.ArchiveMealPlanOptionVote(ctx, &mealplanninggrpc.ArchiveMealPlanOptionVoteRequest{
				MealPlanId:           createdMealPlan.ID,
				MealPlanEventId:      createdMealPlanEvent.ID,
				MealPlanOptionId:     createdMealPlanOption.ID,
				MealPlanOptionVoteId: createdMealPlanOptionVote.ID,
			})
			assert.NoError(t, err)
		}

		_, err := userClient.ArchiveMealPlanOption(ctx, &mealplanninggrpc.ArchiveMealPlanOptionRequest{
			MealPlanId:       createdMealPlan.ID,
			MealPlanEventId:  createdMealPlanEvent.ID,
			MealPlanOptionId: createdMealPlanOption.ID,
		})
		assert.NoError(t, err)

		_, err = userClient.ArchiveMealPlanEvent(ctx, &mealplanninggrpc.ArchiveMealPlanEventRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
		})
		require.NoError(t, err)

		_, err = userClient.ArchiveMealPlan(ctx, &mealplanninggrpc.ArchiveMealPlanRequest{MealPlanId: createdMealPlan.ID})
		assert.NoError(t, err)
	})
}
