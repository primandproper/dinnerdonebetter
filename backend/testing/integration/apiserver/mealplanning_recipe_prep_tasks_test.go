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

func checkRecipePrepTaskStepSliceEquality(t *testing.T, taskIndex int, expected, actual []*mealplanning.RecipePrepTaskStep) {
	t.Helper()
	require.Len(t, actual, len(expected), "expected prep task %d steps length", taskIndex)
	for i := range expected {
		checkRecipePrepTaskStepEquality(t, taskIndex, i, expected[i], actual[i])
	}
}

func checkRecipePrepTaskStepEquality(t *testing.T, taskIndex, stepIndex int, expected, actual *mealplanning.RecipePrepTaskStep) {
	t.Helper()
	assert.NotEmpty(t, actual.ID, "expected prep task %d step %d to have ID", taskIndex, stepIndex)
	assert.NotEmpty(t, actual.BelongsToRecipeStep, "expected prep task %d step %d to have BelongsToRecipeStep", taskIndex, stepIndex)
	assert.NotEmpty(t, actual.BelongsToRecipePrepTask, "expected prep task %d step %d to have BelongsToRecipePrepTask", taskIndex, stepIndex)
	assert.Equal(t, expected.SatisfiesRecipeStep, actual.SatisfiesRecipeStep, "expected prep task %d step %d SatisfiesRecipeStep", taskIndex, stepIndex)
}

func checkRecipePrepTaskSliceEquality(t *testing.T, expected, actual []*mealplanning.RecipePrepTask) {
	t.Helper()

	require.Len(t, actual, len(expected), "expected recipe prep tasks length")
	for i := range expected {
		checkRecipePrepTaskEquality(t, i, expected[i], actual[i])
	}
}

func checkRecipePrepTaskEquality(t *testing.T, taskIndex int, expected, actual *mealplanning.RecipePrepTask) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected prep task %d to have ID", taskIndex)
	assert.False(t, actual.CreatedAt.IsZero(), "expected prep task %d to have CreatedAt", taskIndex)
	assert.NotEmpty(t, actual.BelongsToRecipe, "expected prep task %d to have BelongsToRecipe", taskIndex)
	assert.Equal(t, expected.Name, actual.Name, "expected prep task %d Name", taskIndex)
	assert.Equal(t, expected.Description, actual.Description, "expected prep task %d Description", taskIndex)
	assert.Equal(t, expected.StorageType, actual.StorageType, "expected prep task %d StorageType", taskIndex)
	assert.Equal(t, expected.ExplicitStorageInstructions, actual.ExplicitStorageInstructions, "expected prep task %d ExplicitStorageInstructions", taskIndex)
	assert.Equal(t, expected.Notes, actual.Notes, "expected prep task %d Notes", taskIndex)
	assert.Equal(t, expected.Optional, actual.Optional, "expected prep task %d Optional", taskIndex)
	assert.Equal(t, expected.MinStorageTemperatureInCelsius, actual.MinStorageTemperatureInCelsius, "expected prep task %d MinStorageTemperatureInCelsius", taskIndex)
	assert.Equal(t, expected.MaxStorageTemperatureInCelsius, actual.MaxStorageTemperatureInCelsius, "expected prep task %d MaxStorageTemperatureInCelsius", taskIndex)
	assert.Equal(t, expected.MinTimeBufferBeforeRecipeInSeconds, actual.MinTimeBufferBeforeRecipeInSeconds, "expected prep task %d MinTimeBufferBeforeRecipeInSeconds", taskIndex)
	assert.Equal(t, expected.MaxTimeBufferBeforeRecipeInSeconds, actual.MaxTimeBufferBeforeRecipeInSeconds, "expected prep task %d MaxTimeBufferBeforeRecipeInSeconds", taskIndex)
	checkRecipePrepTaskStepSliceEquality(t, taskIndex, expected.TaskSteps, actual.TaskSteps)
}

func createRecipePrepTaskForTest(t *testing.T) (*mealplanning.Recipe, *mealplanning.RecipePrepTask) {
	t.Helper()
	ctx := t.Context()

	_, _, createdRecipe := createRecipeForTest(t, nil)

	var createdRecipeStep *mealplanning.RecipeStep
	for _, step := range createdRecipe.Steps {
		createdRecipeStep = step
		break
	}
	require.NotNil(t, createdRecipeStep)

	exampleRecipePrepTask := fakes.BuildFakeRecipePrepTask()
	exampleRecipePrepTask.BelongsToRecipe = createdRecipe.ID
	exampleRecipePrepTask.TaskSteps = []*mealplanning.RecipePrepTaskStep{
		{
			BelongsToRecipeStep:     createdRecipeStep.ID,
			BelongsToRecipePrepTask: exampleRecipePrepTask.ID,
			SatisfiesRecipeStep:     true,
		},
	}

	exampleInput := mpconverters.ConvertRecipePrepTaskToRecipePrepTaskCreationRequestInput(exampleRecipePrepTask)

	createdRecipePrepTaskRes, err := adminClient.CreateRecipePrepTask(ctx, &mealplanninggrpc.CreateRecipePrepTaskRequest{
		RecipeId: createdRecipe.ID,
		Input:    converters.ConvertRecipePrepTaskCreationRequestInputToGRPCRecipePrepTaskCreationRequestInput(exampleInput),
	})
	require.NoError(t, err)
	createdRecipePrepTask := converters.ConvertGRPCRecipePrepTaskToRecipePrepTask(createdRecipePrepTaskRes.Created)

	retrievedRes, err := adminClient.GetRecipePrepTask(ctx, &mealplanninggrpc.GetRecipePrepTaskRequest{
		RecipeId:         createdRecipe.ID,
		RecipePrepTaskId: createdRecipePrepTask.ID,
	})
	require.NoError(t, err)

	actual := converters.ConvertGRPCRecipePrepTaskToRecipePrepTask(retrievedRes.Result)
	checkRecipePrepTaskEquality(t, -1, exampleRecipePrepTask, actual)

	return createdRecipe, actual
}

func TestRecipePrepTasks_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("should CRUD", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		createdRecipe, actual := createRecipePrepTaskForTest(t)

		newRecipePrepTask := fakes.BuildFakeRecipePrepTask()
		newRecipePrepTask.ID = actual.ID
		newRecipePrepTask.BelongsToRecipe = createdRecipe.ID
		newRecipePrepTask.TaskSteps = actual.TaskSteps

		updateInput := mpconverters.ConvertRecipePrepTaskToRecipePrepTaskUpdateRequestInput(newRecipePrepTask)
		actual.Update(updateInput)

		_, err := adminClient.UpdateRecipePrepTask(ctx, &mealplanninggrpc.UpdateRecipePrepTaskRequest{
			RecipeId:         createdRecipe.ID,
			RecipePrepTaskId: actual.ID,
			Input:            converters.ConvertRecipePrepTaskUpdateRequestInputToGRPCRecipePrepTaskUpdateRequestInput(updateInput),
		})
		require.NoError(t, err)

		retrievedRes, err := adminClient.GetRecipePrepTask(ctx, &mealplanninggrpc.GetRecipePrepTaskRequest{
			RecipeId:         createdRecipe.ID,
			RecipePrepTaskId: actual.ID,
		})
		require.NoError(t, err)

		retrieved := converters.ConvertGRPCRecipePrepTaskToRecipePrepTask(retrievedRes.Result)

		// assert recipe prep task equality
		checkRecipePrepTaskEquality(t, -1, actual, retrieved)
		assert.NotNil(t, retrieved.LastUpdatedAt)

		_, err = adminClient.ArchiveRecipePrepTask(ctx, &mealplanninggrpc.ArchiveRecipePrepTaskRequest{
			RecipeId:         createdRecipe.ID,
			RecipePrepTaskId: actual.ID,
		})
		assert.NoError(t, err)

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipePrepTasks_Listing(T *testing.T) {
	T.Parallel()

	T.Run("should CRUD", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		var createdRecipeStep *mealplanning.RecipeStep
		for _, step := range createdRecipe.Steps {
			createdRecipeStep = step
			break
		}
		require.NotNil(t, createdRecipeStep)

		var expected []*mealplanning.RecipePrepTask
		for range 5 {
			exampleRecipePrepTask := fakes.BuildFakeRecipePrepTask()
			exampleRecipePrepTask.BelongsToRecipe = createdRecipe.ID
			exampleRecipePrepTask.TaskSteps = []*mealplanning.RecipePrepTaskStep{
				{
					BelongsToRecipeStep:     createdRecipeStep.ID,
					BelongsToRecipePrepTask: exampleRecipePrepTask.ID,
					SatisfiesRecipeStep:     true,
				},
			}

			exampleInput := mpconverters.ConvertRecipePrepTaskToRecipePrepTaskCreationRequestInput(exampleRecipePrepTask)

			createdRecipePrepTaskRes, err := adminClient.CreateRecipePrepTask(ctx, &mealplanninggrpc.CreateRecipePrepTaskRequest{
				RecipeId: createdRecipe.ID,
				Input:    converters.ConvertRecipePrepTaskCreationRequestInputToGRPCRecipePrepTaskCreationRequestInput(exampleInput),
			})
			require.NoError(t, err)

			createdRecipePrepTask := converters.ConvertGRPCRecipePrepTaskToRecipePrepTask(createdRecipePrepTaskRes.Created)

			for j := range createdRecipePrepTask.TaskSteps {
				exampleRecipePrepTask.TaskSteps[j].ID = createdRecipePrepTask.TaskSteps[j].ID
				exampleRecipePrepTask.TaskSteps[j].BelongsToRecipePrepTask = createdRecipePrepTask.ID
			}

			checkRecipePrepTaskEquality(t, -1, exampleRecipePrepTask, createdRecipePrepTask)

			retrievedRecipePrepTaskRes, err := adminClient.GetRecipePrepTask(ctx, &mealplanninggrpc.GetRecipePrepTaskRequest{
				RecipeId:         createdRecipe.ID,
				RecipePrepTaskId: createdRecipePrepTask.ID,
			})
			require.NoError(t, err)

			createdRecipePrepTask = converters.ConvertGRPCRecipePrepTaskToRecipePrepTask(retrievedRecipePrepTaskRes.Result)
			require.Equal(t, createdRecipe.ID, createdRecipePrepTask.BelongsToRecipe)

			expected = append(expected, createdRecipePrepTask)
		}

		// assert recipe prep task list equality
		actual, err := adminClient.GetRecipePrepTasks(ctx, &mealplanninggrpc.GetRecipePrepTasksRequest{RecipeId: createdRecipe.ID})
		require.NoError(t, err)
		assert.LessOrEqual(
			t,
			len(expected), len(actual.Results),
			"expected %d to be <= %d",
			len(expected),
			len(actual.Results),
		)

		for _, createdRecipePrepTask := range expected {
			_, err = adminClient.ArchiveRecipePrepTask(ctx, &mealplanninggrpc.ArchiveRecipePrepTaskRequest{
				RecipeId:         createdRecipe.ID,
				RecipePrepTaskId: createdRecipePrepTask.ID,
			})
			assert.NoError(t, err)
		}

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipePrepTasks_OwnershipDenial(T *testing.T) {
	T.Parallel()

	T.Run("Create", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, otherUserClient := createUserAndClientForTest(t)

		var createdRecipeStep *mealplanning.RecipeStep
		for _, step := range createdRecipe.Steps {
			createdRecipeStep = step
			break
		}
		require.NotNil(t, createdRecipeStep)

		exampleRecipePrepTask := fakes.BuildFakeRecipePrepTask()
		exampleRecipePrepTask.BelongsToRecipe = createdRecipe.ID
		exampleRecipePrepTask.TaskSteps = []*mealplanning.RecipePrepTaskStep{
			{
				BelongsToRecipeStep:     createdRecipeStep.ID,
				BelongsToRecipePrepTask: exampleRecipePrepTask.ID,
				SatisfiesRecipeStep:     true,
			},
		}

		exampleInput := mpconverters.ConvertRecipePrepTaskToRecipePrepTaskCreationRequestInput(exampleRecipePrepTask)

		_, err := otherUserClient.CreateRecipePrepTask(ctx, &mealplanninggrpc.CreateRecipePrepTaskRequest{
			RecipeId: createdRecipe.ID,
			Input:    converters.ConvertRecipePrepTaskCreationRequestInputToGRPCRecipePrepTaskCreationRequestInput(exampleInput),
		})
		require.Error(t, err)
	})

	T.Run("Update", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		createdRecipe, actual := createRecipePrepTaskForTest(t)
		_, otherUserClient := createUserAndClientForTest(t)

		updateInput := mpconverters.ConvertRecipePrepTaskToRecipePrepTaskUpdateRequestInput(actual)

		_, err := otherUserClient.UpdateRecipePrepTask(ctx, &mealplanninggrpc.UpdateRecipePrepTaskRequest{
			RecipeId:         createdRecipe.ID,
			RecipePrepTaskId: actual.ID,
			Input:            converters.ConvertRecipePrepTaskUpdateRequestInputToGRPCRecipePrepTaskUpdateRequestInput(updateInput),
		})
		require.Error(t, err)
	})

	T.Run("Archive", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		createdRecipe, actual := createRecipePrepTaskForTest(t)
		_, otherUserClient := createUserAndClientForTest(t)

		_, err := otherUserClient.ArchiveRecipePrepTask(ctx, &mealplanninggrpc.ArchiveRecipePrepTaskRequest{
			RecipeId:         createdRecipe.ID,
			RecipePrepTaskId: actual.ID,
		})
		require.Error(t, err)
	})
}
