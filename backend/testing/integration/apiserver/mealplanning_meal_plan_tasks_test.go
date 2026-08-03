package integration

import (
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mpconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	authgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	identitygrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkMealPlanTaskEquality(t *testing.T, expected, actual *mealplanning.MealPlanTask) {
	t.Helper()

	assert.NotEmpty(t, actual.ID)
	assert.Equal(t, expected.CreationExplanation, actual.CreationExplanation, "expected CreationExplanation for meal plan %s to be %v, but it was %v", expected.CreationExplanation, expected.CreationExplanation, actual.CreationExplanation)
	assert.Equal(t, expected.Status, actual.Status, "expected Status for meal plan %s to be %v, but it was %v", expected.Status, expected.Status, actual.Status)
	assert.Equal(t, expected.StatusExplanation, actual.StatusExplanation, "expected StatusExplanation for meal plan %s to be %v, but it was %v", expected.StatusExplanation, expected.StatusExplanation, actual.StatusExplanation)
	assert.Equal(t, expected.AssignedToUser, actual.AssignedToUser, "expected AssignedToUser for meal plan %s to be %v, but it was %v", expected.AssignedToUser, expected.AssignedToUser, actual.AssignedToUser)
	assert.Equal(t, expected.CompletedAt, actual.CompletedAt, "expected CompletedAt for meal plan %s to be %v, but it was %v", expected.CompletedAt, expected.CompletedAt, actual.CompletedAt)

	assert.NotZero(t, actual.CreatedAt)
}

func TestMealPlanTasks_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("should CRUD", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		createdMealPlan := createMealPlanForTest(t, adminClient, nil)

		exampleMealPlanTask := fakes.BuildFakeMealPlanTask()
		exampleMealPlanTaskInput := mpconverters.ConvertMealPlanTaskToMealPlanTaskCreationRequestInput(exampleMealPlanTask)

		exampleMealPlanTaskInput.MealPlanOptionID = createdMealPlan.Events[0].Options[0].ID
		exampleMealPlanTaskInput.RecipePrepTaskID = createdMealPlan.Events[0].Options[0].Meal.Components[0].Recipe.PrepTasks[0].ID

		createdMealPlanTaskRes, err := adminClient.CreateMealPlanTask(ctx, &mealplanninggrpc.CreateMealPlanTaskRequest{
			MealPlanId: createdMealPlan.ID,
			Input:      converters.ConvertMealPlanTaskCreationRequestInputToGRPCMealPlanTaskCreationRequestInput(exampleMealPlanTaskInput),
		})
		require.NoError(t, err)

		createdMealPlanTask := converters.ConvertGRPCMealPlanTaskToMealPlanTask(createdMealPlanTaskRes.Created)
		checkMealPlanTaskEquality(t, exampleMealPlanTask, createdMealPlanTask)

		actualRes, err := adminClient.GetMealPlanTask(ctx, &mealplanninggrpc.GetMealPlanTaskRequest{
			MealPlanId:     createdMealPlan.ID,
			MealPlanTaskId: createdMealPlanTask.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, actualRes)

		actual := converters.ConvertGRPCMealPlanTaskToMealPlanTask(actualRes.Result)
		// assert meal plan task equality
		checkMealPlanTaskEquality(t, exampleMealPlanTask, actual)
	})
}

// createFinalizedMealPlanWithTasks creates a meal plan that has been finalized and has tasks created.
// It sets up household members, votes, forces the deadline, runs finalization, and returns the
// meal plan ID and the user client that owns the meal plan.
func createFinalizedMealPlanWithTasks(t *testing.T) (string, client.Client) {
	t.Helper()
	ctx := t.Context()

	_, accountAdminUserClient := createUserAndClientForTest(t)

	currentStatus, statusErr := accountAdminUserClient.GetAuthStatus(ctx, &authgrpc.GetAuthStatusRequest{})
	require.NotNil(t, currentStatus)
	require.NoError(t, statusErr)
	relevantAccountID := currentStatus.ActiveAccount

	// Create 2 additional household members
	createdClients := []client.Client{}
	for range 2 {
		u, c := createUserAndClientForTest(t)

		invitation, err := accountAdminUserClient.CreateAccountInvitation(ctx, &identitygrpc.CreateAccountInvitationRequest{
			Input: &identitygrpc.AccountInvitationCreationRequestInput{
				Note:    t.Name(),
				ToName:  t.Name(),
				ToEmail: u.EmailAddress,
			},
		})
		require.NoError(t, err)

		_, err = c.AcceptAccountInvitation(ctx, &identitygrpc.AcceptAccountInvitationRequest{
			AccountInvitationId: invitation.Created.Id,
			Input: &identitygrpc.AccountInvitationUpdateRequestInput{
				Token: invitation.Created.Token,
				Note:  t.Name(),
			},
		})
		require.NoError(t, err)

		_, err = c.SetDefaultAccount(ctx, &identitygrpc.SetDefaultAccountRequest{AccountId: relevantAccountID})
		require.NoError(t, err)

		tokenResponse, err := c.LoginForToken(ctx, &authgrpc.LoginForTokenRequest{Input: &authgrpc.UserLoginInput{
			Username:  u.Username,
			Password:  u.HashedPassword,
			TotpToken: generateTOTPCodeForUserForTest(t, u),
		}})
		require.NoError(t, err)
		assert.NotNil(t, tokenResponse)

		createdClients = append(createdClients, c)
	}

	// Create meals
	createdMeals := []*mealplanning.Meal{}
	for range 3 {
		createdMeal := createMealForTest(t, accountAdminUserClient, nil)
		createdMeals = append(createdMeals, createdMeal)
	}

	now := time.Now()
	exampleMealPlan := &mealplanning.MealPlan{
		Notes:          t.Name(),
		Status:         string(mealplanning.MealPlanStatusAwaitingVotes),
		VotingDeadline: now.Add(10 * time.Second),
		ElectionMethod: mealplanning.MealPlanElectionMethodSchulze,
		Events: []*mealplanning.MealPlanEvent{
			{
				StartsAt: now.Add(24 * time.Hour),
				EndsAt:   now.Add(72 * time.Hour),
				MealName: mealplanning.BreakfastMealName,
				Options: []*mealplanning.MealPlanOption{
					{Meal: mealplanning.Meal{ID: createdMeals[0].ID}, Notes: "option A"},
					{Meal: mealplanning.Meal{ID: createdMeals[1].ID}, Notes: "option B"},
					{Meal: mealplanning.Meal{ID: createdMeals[2].ID}, Notes: "option C"},
				},
			},
		},
	}

	createdMealPlan := createMealPlanForTest(t, accountAdminUserClient, exampleMealPlan)
	createdMealPlanEvent := createdMealPlan.Events[0]

	// All 3 users vote
	for i, userClient := range append(createdClients, accountAdminUserClient) {
		votes := &mealplanning.MealPlanOptionVoteCreationRequestInput{
			Votes: []*mealplanning.MealPlanOptionVoteCreationInput{
				{BelongsToMealPlanOption: createdMealPlanEvent.Options[0].ID, Rank: uint8(i % 3)},
				{BelongsToMealPlanOption: createdMealPlanEvent.Options[1].ID, Rank: uint8((i + 1) % 3)},
				{BelongsToMealPlanOption: createdMealPlanEvent.Options[2].ID, Rank: uint8((i + 2) % 3)},
			},
		}
		_, err := userClient.CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
			Input:           converters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(votes),
		})
		require.NoError(t, err)
	}

	// Force voting deadline to the past
	q := generated.New()
	rowsAffected, err := q.UpdateMealPlan(ctx, databaseClient.Writer(), &generated.UpdateMealPlanParams{
		Notes:            createdMealPlan.Notes,
		Status:           generated.MealPlanStatus(createdMealPlan.Status),
		VotingDeadline:   time.Now().Add(-time.Minute),
		BelongsToAccount: relevantAccountID,
		ID:               createdMealPlan.ID,
	})
	require.NoError(t, err)
	require.NotZero(t, rowsAffected)

	// Finalize the meal plan directly (per-plan, not global worker, to avoid conflicts with parallel tests)
	finalizeRes, err := accountAdminUserClient.FinalizeMealPlan(ctx, &mealplanninggrpc.FinalizeMealPlanRequest{MealPlanId: createdMealPlan.ID})
	require.NoError(t, err)
	require.NotNil(t, finalizeRes)

	return createdMealPlan.ID, accountAdminUserClient
}

func TestMealPlanTasks_RunMealPlanTaskCreatorWorker(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		mealPlanID, userClient := createFinalizedMealPlanWithTasks(t)

		// The admin endpoint starts a finalization saga for every plan that needs one; the saga
		// worker is what creates the tasks, a moment later.
		runRes, err := adminClient.RunMealPlanTaskCreatorWorker(ctx, &mealplanninggrpc.RunMealPlanTaskCreatorWorkerRequest{})
		require.NoError(t, err)
		require.NotNil(t, runRes)

		assert.NotEmpty(t, awaitMealPlanTasks(t, ctx, userClient, mealPlanID))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		res, err := c.RunMealPlanTaskCreatorWorker(ctx, &mealplanninggrpc.RunMealPlanTaskCreatorWorkerRequest{})
		require.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestMealPlanTasks_UpdateMealPlanTaskStatus(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		mealPlanID, userClient := createFinalizedMealPlanWithTasks(t)

		// Nudge the starter, then wait for the saga to get to the task step.
		runRes, err := adminClient.RunMealPlanTaskCreatorWorker(ctx, &mealplanninggrpc.RunMealPlanTaskCreatorWorkerRequest{})
		require.NoError(t, err)
		require.NotNil(t, runRes)

		taskToUpdate := awaitMealPlanTasks(t, ctx, userClient, mealPlanID)[0]
		newStatus := mealplanninggrpc.MealPlanTaskStatus_MEAL_PLAN_TASK_STATUS_FINISHED

		// Update the task status
		updateRes, err := userClient.UpdateMealPlanTaskStatus(ctx, &mealplanninggrpc.UpdateMealPlanTaskStatusRequest{
			MealPlanId:     mealPlanID,
			MealPlanTaskId: taskToUpdate.Id,
			Input: &mealplanninggrpc.MealPlanTaskStatusChangeRequestInput{
				Id:                taskToUpdate.Id,
				Status:            &newStatus,
				StatusExplanation: "task completed during integration test",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, updateRes)
		require.NotNil(t, updateRes.Updated)

		updatedTask := converters.ConvertGRPCMealPlanTaskToMealPlanTask(updateRes.Updated)
		assert.Equal(t, mealplanning.MealPlanTaskStatusFinished, updatedTask.Status)
		assert.Equal(t, "task completed during integration test", updatedTask.StatusExplanation)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		newStatus := mealplanninggrpc.MealPlanTaskStatus_MEAL_PLAN_TASK_STATUS_FINISHED
		c := buildUnauthenticatedGRPCClientForTest(t)
		res, err := c.UpdateMealPlanTaskStatus(ctx, &mealplanninggrpc.UpdateMealPlanTaskStatusRequest{
			MealPlanId:     nonexistentID,
			MealPlanTaskId: nonexistentID,
			Input: &mealplanninggrpc.MealPlanTaskStatusChangeRequestInput{
				Id:                nonexistentID,
				Status:            &newStatus,
				StatusExplanation: "should fail",
			},
		})
		require.Error(t, err)
		assert.Nil(t, res)
	})

	T.Run("nonexistent meal plan task", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		newStatus := mealplanninggrpc.MealPlanTaskStatus_MEAL_PLAN_TASK_STATUS_FINISHED
		res, err := userClient.UpdateMealPlanTaskStatus(ctx, &mealplanninggrpc.UpdateMealPlanTaskStatusRequest{
			MealPlanId:     nonexistentID,
			MealPlanTaskId: nonexistentID,
			Input: &mealplanninggrpc.MealPlanTaskStatusChangeRequestInput{
				Id:                nonexistentID,
				Status:            &newStatus,
				StatusExplanation: "should fail",
			},
		})
		require.Error(t, err)
		assert.Nil(t, res)
	})
}
