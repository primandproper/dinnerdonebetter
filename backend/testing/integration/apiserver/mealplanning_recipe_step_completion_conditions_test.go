package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mpconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkRecipeStepCompletionConditionSliceEquality(t *testing.T, stepIndex int, expected, actual []*mealplanning.RecipeStepCompletionCondition) {
	t.Helper()
	require.Len(t, actual, len(expected), "expected recipe step %d completion conditions length", stepIndex)
	for i := range expected {
		checkRecipeStepCompletionConditionEquality(t, stepIndex, i, expected[i], actual[i])
	}
}

func checkRecipeStepCompletionConditionEquality(t *testing.T, stepIndex, condIndex int, expected, actual *mealplanning.RecipeStepCompletionCondition) {
	t.Helper()
	assert.NotEmpty(t, actual.ID, "expected step %d condition %d to have ID", stepIndex, condIndex)
	assert.False(t, actual.CreatedAt.IsZero(), "expected step %d condition %d to have CreatedAt", stepIndex, condIndex)
	assert.NotEmpty(t, actual.BelongsToRecipeStep, "expected step %d condition %d to have BelongsToRecipeStep", stepIndex, condIndex)
	assert.Equal(t, expected.Notes, actual.Notes, "expected step %d condition %d Notes", stepIndex, condIndex)
	assert.Equal(t, expected.Optional, actual.Optional, "expected step %d condition %d Optional", stepIndex, condIndex)
	assert.Equal(t, expected.IngredientState.ID, actual.IngredientState.ID, "expected step %d condition %d IngredientState.ID", stepIndex, condIndex)
	checkRecipeStepCompletionConditionIngredientSliceEquality(t, stepIndex, condIndex, expected.Ingredients, actual.Ingredients)
}

func checkRecipeStepCompletionConditionIngredientSliceEquality(t *testing.T, stepIndex, condIndex int, expected, actual []*mealplanning.RecipeStepCompletionConditionIngredient) {
	t.Helper()
	require.Len(t, actual, len(expected), "expected step %d condition %d ingredients length", stepIndex, condIndex)
	for i := range expected {
		e, a := expected[i], actual[i]
		assert.NotEmpty(t, a.ID, "expected step %d condition %d ingredient %d to have ID", stepIndex, condIndex, i)
		assert.False(t, a.CreatedAt.IsZero(), "expected step %d condition %d ingredient %d to have CreatedAt", stepIndex, condIndex, i)
		assert.NotEmpty(t, a.BelongsToRecipeStepCompletionCondition, "expected step %d condition %d ingredient %d to have BelongsTo...", stepIndex, condIndex, i)
		assert.Equal(t, e.RecipeStepIngredient, a.RecipeStepIngredient, "expected step %d condition %d ingredient %d RecipeStepIngredient", stepIndex, condIndex, i)
	}
}

func TestRecipeStepCompletionConditions_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		createdRecipeStep := createdRecipe.Steps[0]
		require.NotEmpty(t, createdRecipeStep.ID, "created recipe step ID must not be empty")

		// create ingredient state
		createdValidIngredientState := createValidIngredientStateForTest(t)

		input := &mealplanning.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput{
			IngredientStateID:   createdValidIngredientState.ID,
			BelongsToRecipeStep: createdRecipeStep.ID,
			Notes:               t.Name(),
			Optional:            false,
			Ingredients: []*mealplanning.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput{
				{
					RecipeStepIngredient: createdRecipeStep.Ingredients[0].ID,
				},
			},
		}

		createdRecipeStepCompletionConditionRes, err := adminClient.CreateRecipeStepCompletionCondition(ctx, &mealplanninggrpc.CreateRecipeStepCompletionConditionRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStep.ID,
			Input:        converters.ConvertRecipeStepCompletionConditionForExistingRecipeCreationRequestInputToGRPCRecipeStepCompletionConditionForExistingRecipeCreationRequestInput(input),
		})
		require.NoError(t, err)
		require.NotNil(t, createdRecipeStepCompletionConditionRes)

		createdRecipeStepCompletionCondition := converters.ConvertGRPCRecipeStepCompletionConditionToRecipeStepCompletionCondition(createdRecipeStepCompletionConditionRes.Created)

		createdRecipeStepCompletionCondition.Notes = t.Name() + " updated"
		updateInput := mpconverters.ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionUpdateRequestInput(createdRecipeStepCompletionCondition)

		_, err = adminClient.UpdateRecipeStepCompletionCondition(ctx, &mealplanninggrpc.UpdateRecipeStepCompletionConditionRequest{
			RecipeId:                        createdRecipe.ID,
			RecipeStepId:                    createdRecipeStep.ID,
			RecipeStepCompletionConditionId: createdRecipeStepCompletionCondition.ID,
			Input:                           converters.ConvertRecipeStepCompletionConditionUpdateRequestInputToGRPCRecipeStepCompletionConditionUpdateRequestInput(updateInput),
		})
		require.NoError(t, err)

		retrievedRes, err := adminClient.GetRecipeStepCompletionCondition(ctx, &mealplanninggrpc.GetRecipeStepCompletionConditionRequest{
			RecipeId:                        createdRecipe.ID,
			RecipeStepId:                    createdRecipeStep.ID,
			RecipeStepCompletionConditionId: createdRecipeStepCompletionCondition.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, retrievedRes)

		actual := converters.ConvertGRPCRecipeStepCompletionConditionToRecipeStepCompletionCondition(retrievedRes.Result)

		actual.IngredientState = mealplanning.ValidIngredientState{
			ID: createdRecipeStepCompletionCondition.IngredientState.ID,
		}
		actual.CreatedAt = createdRecipeStepCompletionCondition.CreatedAt
		for i := range actual.Ingredients {
			actual.Ingredients[i].CreatedAt = createdRecipeStepCompletionCondition.Ingredients[i].CreatedAt
		}

		// assert recipe step completion condition equality
		checkRecipeStepCompletionConditionEquality(t, -1, -1, createdRecipeStepCompletionCondition, actual)
		assert.NotNil(t, actual.LastUpdatedAt)

		// assert recipe step completion condition list functionality works
		listResponse, err := adminClient.GetRecipeStepCompletionConditions(ctx, &mealplanninggrpc.GetRecipeStepCompletionConditionsRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStep.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.LessOrEqual(
			t,
			1, len(listResponse.Results),
			"expected %d to be <= %d",
			1,
			len(listResponse.Results),
		)

		_, err = adminClient.ArchiveRecipeStepCompletionCondition(ctx, &mealplanninggrpc.ArchiveRecipeStepCompletionConditionRequest{
			RecipeId:                        createdRecipe.ID,
			RecipeStepId:                    createdRecipeStep.ID,
			RecipeStepCompletionConditionId: createdRecipeStepCompletionCondition.ID,
		})
		require.NoError(t, err)

		_, err = adminClient.ArchiveRecipeStep(ctx, &mealplanninggrpc.ArchiveRecipeStepRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStep.ID,
		})
		require.NoError(t, err)

		_, err = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
		assert.NoError(t, err)
	})
}

func TestRecipeStepCompletionConditions_OwnershipDenial(T *testing.T) {
	T.Parallel()

	T.Run("Create", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, otherUserClient := createUserAndClientForTest(t)

		createdRecipeStep := createdRecipe.Steps[0]

		input := &mealplanning.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput{
			IngredientStateID:   "fake-ingredient-state-id",
			BelongsToRecipeStep: createdRecipeStep.ID,
			Notes:               t.Name(),
			Optional:            false,
			Ingredients: []*mealplanning.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput{
				{
					RecipeStepIngredient: createdRecipeStep.Ingredients[0].ID,
				},
			},
		}

		_, err := otherUserClient.CreateRecipeStepCompletionCondition(ctx, &mealplanninggrpc.CreateRecipeStepCompletionConditionRequest{
			RecipeId:     createdRecipe.ID,
			RecipeStepId: createdRecipeStep.ID,
			Input:        converters.ConvertRecipeStepCompletionConditionForExistingRecipeCreationRequestInputToGRPCRecipeStepCompletionConditionForExistingRecipeCreationRequestInput(input),
		})
		require.Error(t, err)
	})

	T.Run("Update", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, otherUserClient := createUserAndClientForTest(t)

		createdRecipeStep := createdRecipe.Steps[0]
		require.NotEmpty(t, createdRecipeStep.CompletionConditions)

		existingCondition := createdRecipeStep.CompletionConditions[0]
		updateInput := &mealplanning.RecipeStepCompletionConditionUpdateRequestInput{
			Notes: new(t.Name()),
		}

		_, err := otherUserClient.UpdateRecipeStepCompletionCondition(ctx, &mealplanninggrpc.UpdateRecipeStepCompletionConditionRequest{
			RecipeId:                        createdRecipe.ID,
			RecipeStepId:                    createdRecipeStep.ID,
			RecipeStepCompletionConditionId: existingCondition.ID,
			Input:                           converters.ConvertRecipeStepCompletionConditionUpdateRequestInputToGRPCRecipeStepCompletionConditionUpdateRequestInput(updateInput),
		})
		require.Error(t, err)
	})

	T.Run("Archive", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, otherUserClient := createUserAndClientForTest(t)

		createdRecipeStep := createdRecipe.Steps[0]
		require.NotEmpty(t, createdRecipeStep.CompletionConditions)

		existingCondition := createdRecipeStep.CompletionConditions[0]

		_, err := otherUserClient.ArchiveRecipeStepCompletionCondition(ctx, &mealplanninggrpc.ArchiveRecipeStepCompletionConditionRequest{
			RecipeId:                        createdRecipe.ID,
			RecipeStepId:                    createdRecipeStep.ID,
			RecipeStepCompletionConditionId: existingCondition.ID,
		})
		require.Error(t, err)
	})
}
