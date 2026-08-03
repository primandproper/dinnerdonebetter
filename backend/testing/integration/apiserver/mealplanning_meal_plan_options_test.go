package integration

import (
	"testing"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mpconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func checkMealPlanOptionEquality(t *testing.T, expected, actual *types.MealPlanOption) {
	t.Helper()

	assert.NotEmpty(t, actual.ID)
	assert.Equal(t, expected.Meal.ID, actual.Meal.ID, "expected RecipeID for meal plan option %s to be %v, but it was %v", expected.ID, expected.Meal.ID, actual.Meal.ID)
	assert.Equal(t, expected.Notes, actual.Notes, "expected StatusExplanation for meal plan option %s to be %v, but it was %v", expected.ID, expected.Notes, actual.Notes)
	assert.Equal(t, expected.AssignedCook, actual.AssignedCook, "expected AssignedCook for meal plan option %s to be %v, but it was %v", expected.ID, expected.AssignedCook, actual.AssignedCook)
	assert.Equal(t, expected.AssignedDishwasher, actual.AssignedDishwasher, "expected AssignedDishwasher for meal plan option %s to be %v, but it was %v", expected.ID, expected.AssignedDishwasher, actual.AssignedDishwasher)
	assert.NotZero(t, actual.CreatedAt)
}

func TestMealPlanOptions_CompleteLifecycle(T *testing.T) {
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

		newMealPlanOption := fakes.BuildFakeMealPlanOption()
		newMealPlanOption.Meal.ID = createdMealPlanOption.Meal.ID
		newMealPlanOption.BelongsToMealPlanEvent = createdMealPlanEvent.ID
		newMealPlanOption.AssignedCook = nil

		updateInput := mpconverters.ConvertMealPlanOptionToMealPlanOptionUpdateRequestInput(newMealPlanOption)
		createdMealPlanOption.Update(updateInput)
		_, err := userClient.UpdateMealPlanOption(ctx, &mealplanninggrpc.UpdateMealPlanOptionRequest{
			MealPlanId:       createdMealPlan.ID,
			MealPlanOptionId: createdMealPlanOption.ID,
			MealPlanEventId:  createdMealPlanEvent.ID,
			Input:            converters.ConvertMealPlanOptionUpdateRequestInputToGRPCMealPlanOptionUpdateRequestInput(updateInput),
		})
		require.NoError(t, err)

		actualRes, err := userClient.GetMealPlanOption(ctx, &mealplanninggrpc.GetMealPlanOptionRequest{
			MealPlanId:       createdMealPlan.ID,
			MealPlanEventId:  createdMealPlanEvent.ID,
			MealPlanOptionId: createdMealPlanOption.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, actualRes)

		actual := converters.ConvertGRPCMealPlanOptionToMealPlanOption(actualRes.Result)

		// assert meal plan option equality
		checkMealPlanOptionEquality(t, newMealPlanOption, actual)
		assert.NotNil(t, actual.LastUpdatedAt)

		_, err = userClient.ArchiveMealPlanOption(ctx, &mealplanninggrpc.ArchiveMealPlanOptionRequest{
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

func TestMealPlanOptions_Listing(T *testing.T) {
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

		var expected []*types.MealPlanOption
		for range 5 {
			exampleMealPlanOption := fakes.BuildFakeMealPlanOption()
			exampleMealPlanOption.Meal.ID = createdMealPlanOption.Meal.ID
			exampleMealPlanOption.BelongsToMealPlanEvent = createdMealPlanEvent.ID
			exampleMealPlanOption.AssignedCook = nil

			createdMeal := createMealForTest(t, userClient, nil)
			exampleMealPlanOption.Meal.ID = createdMeal.ID

			exampleMealPlanOptionInput := mpconverters.ConvertMealPlanOptionToMealPlanOptionCreationRequestInput(exampleMealPlanOption)
			createdMealPlanOptionRes, err := userClient.CreateMealPlanOption(ctx, &mealplanninggrpc.CreateMealPlanOptionRequest{
				MealPlanId:      createdMealPlan.ID,
				MealPlanEventId: createdMealPlanEvent.ID,
				Input:           converters.ConvertMealPlanOptionCreationRequestInputToGRPCMealPlanOptionCreationRequestInput(exampleMealPlanOptionInput),
			})
			require.NoError(t, err)

			newlyCreatedMealPlanOption := converters.ConvertGRPCMealPlanOptionToMealPlanOption(createdMealPlanOptionRes.Created)
			checkMealPlanOptionEquality(t, exampleMealPlanOption, newlyCreatedMealPlanOption)

			retrievedMealPlanOptionRes, err := userClient.GetMealPlanOption(ctx, &mealplanninggrpc.GetMealPlanOptionRequest{
				MealPlanId:       createdMealPlan.ID,
				MealPlanEventId:  createdMealPlanEvent.ID,
				MealPlanOptionId: newlyCreatedMealPlanOption.ID,
			})
			require.NotNil(t, retrievedMealPlanOptionRes)
			require.NoError(t, err)
			require.Equal(t, createdMealPlanEvent.ID, retrievedMealPlanOptionRes.Result.BelongsToMealPlanEvent)

			expected = append(expected, converters.ConvertGRPCMealPlanOptionToMealPlanOption(retrievedMealPlanOptionRes.Result))
		}

		// assert meal plan option list equality
		actual, err := userClient.GetMealPlanOptions(ctx, &mealplanninggrpc.GetMealPlanOptionsRequest{
			Filter:          nil,
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
		})
		require.NotNil(t, actual)
		require.NoError(t, err)
		assert.LessOrEqual(
			t,
			len(expected), len(actual.Results),
			"expected %d to be <= %d",
			len(expected),
			len(actual.Results),
		)

		_, err = userClient.ArchiveMealPlan(ctx, &mealplanninggrpc.ArchiveMealPlanRequest{MealPlanId: createdMealPlan.ID})
		assert.NoError(t, err)
	})
}

func TestMealPlanOptions_DuplicatePrevention(T *testing.T) {
	T.Parallel()

	T.Run("rejects adding same meal twice to same event", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		createdMealPlan := createMealPlanForTest(t, userClient, nil)

		require.NotEmpty(t, createdMealPlan.Events)
		createdMealPlanEvent := createdMealPlan.Events[0]
		createdMealPlanOption := createdMealPlanEvent.Options[0]
		require.NotNil(t, createdMealPlanOption)

		exampleMealPlanOption := fakes.BuildFakeMealPlanOption()
		exampleMealPlanOption.Meal.ID = createdMealPlanOption.Meal.ID
		exampleMealPlanOption.BelongsToMealPlanEvent = createdMealPlanEvent.ID
		exampleMealPlanOption.AssignedCook = nil

		exampleMealPlanOptionInput := mpconverters.ConvertMealPlanOptionToMealPlanOptionCreationRequestInput(exampleMealPlanOption)

		_, err := userClient.CreateMealPlanOption(ctx, &mealplanninggrpc.CreateMealPlanOptionRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
			Input:           converters.ConvertMealPlanOptionCreationRequestInputToGRPCMealPlanOptionCreationRequestInput(exampleMealPlanOptionInput),
		})
		assert.Error(t, err)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})

	T.Run("allows same meal in different events", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		createdMealPlan := createMealPlanForTest(t, userClient, nil)

		require.NotEmpty(t, createdMealPlan.Events)
		require.GreaterOrEqual(t, len(createdMealPlan.Events), 2, "need at least 2 events")
		event1 := createdMealPlan.Events[0]
		event2 := createdMealPlan.Events[1]
		mealID := event1.Options[0].Meal.ID

		exampleOption := fakes.BuildFakeMealPlanOption()
		exampleOption.Meal.ID = mealID
		exampleOption.BelongsToMealPlanEvent = event2.ID
		exampleOption.AssignedCook = nil
		exampleOptionInput := mpconverters.ConvertMealPlanOptionToMealPlanOptionCreationRequestInput(exampleOption)

		_, err := userClient.CreateMealPlanOption(ctx, &mealplanninggrpc.CreateMealPlanOptionRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: event2.ID,
			Input:           converters.ConvertMealPlanOptionCreationRequestInputToGRPCMealPlanOptionCreationRequestInput(exampleOptionInput),
		})
		require.NoError(t, err)
	})

	T.Run("allows different meals in same event", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		createdMealPlan := createMealPlanForTest(t, userClient, nil)

		require.NotEmpty(t, createdMealPlan.Events)
		createdMealPlanEvent := createdMealPlan.Events[0]
		existingOption := createdMealPlanEvent.Options[0]

		createdMeal := createMealForTest(t, userClient, nil)
		exampleOption := fakes.BuildFakeMealPlanOption()
		exampleOption.Meal.ID = createdMeal.ID
		exampleOption.BelongsToMealPlanEvent = createdMealPlanEvent.ID
		exampleOption.AssignedCook = nil
		exampleOptionInput := mpconverters.ConvertMealPlanOptionToMealPlanOptionCreationRequestInput(exampleOption)

		_, err := userClient.CreateMealPlanOption(ctx, &mealplanninggrpc.CreateMealPlanOptionRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
			Input:           converters.ConvertMealPlanOptionCreationRequestInputToGRPCMealPlanOptionCreationRequestInput(exampleOptionInput),
		})
		require.NoError(t, err)

		_ = existingOption
	})
}
