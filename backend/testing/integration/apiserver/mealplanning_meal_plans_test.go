package integration

import (
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mpconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	authgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	identitygrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func checkMealPlanEquality(t *testing.T, expected, actual *mealplanning.MealPlan) {
	t.Helper()

	assert.NotEmpty(t, actual.ID)
	assert.Equal(t, expected.Notes, actual.Notes, "expected Notes for meal plan %s to be %v, but it was %v", expected.ID, expected.Notes, actual.Notes)
	assert.Equal(t, expected.Status, actual.Status, "expected Status for meal plan %s to be %v, but it was %v", expected.ID, expected.Status, actual.Status)
	assert.WithinDuration(t, expected.VotingDeadline, actual.VotingDeadline, time.Nanosecond*1000, "expected VotingDeadline for meal plan %s to be %v, but it was %v", expected.ID, expected.VotingDeadline, actual.VotingDeadline)
	assert.Equal(t, expected.TasksCreated, actual.TasksCreated, "expected TasksCreated for meal plan %s to be %v, but it was %v", expected.ID, expected.TasksCreated, actual.TasksCreated)
	assert.Equal(t, expected.ElectionMethod, actual.ElectionMethod, "expected ElectionMethod for meal plan %s to be %v, but it was %v", expected.ID, expected.ElectionMethod, actual.ElectionMethod)
	assert.Equal(t, expected.GroceryListInitialized, actual.GroceryListInitialized, "expected GroceryListInitialized for meal plan %s to be %v, but it was %v", expected.ID, expected.GroceryListInitialized, actual.GroceryListInitialized)
	assert.NotZero(t, actual.CreatedAt)
}

func createMealPlanForTest(t *testing.T, clientToUse client.Client, mealPlan *mealplanning.MealPlan) *mealplanning.MealPlan {
	t.Helper()
	ctx := t.Context()

	if mealPlan == nil {
		mealPlan = fakes.BuildFakeMealPlan()
		for i, evt := range mealPlan.Events {
			for j := range evt.Options {
				createdMeal := createMealForTest(t, clientToUse, nil)
				mealPlan.Events[i].Options[j].Meal.ID = createdMeal.ID
				mealPlan.Events[i].Options[j].AssignedCook = nil
			}
		}
	}

	exampleMealPlanInput := mpconverters.ConvertMealPlanToMealPlanCreationRequestInput(mealPlan)
	createdMealPlanRes, err := clientToUse.CreateMealPlan(ctx, &mealplanninggrpc.CreateMealPlanRequest{
		Input: converters.ConvertMealPlanCreationRequestInputToGRPCMealPlanCreationRequestInput(exampleMealPlanInput),
	})
	require.NoError(t, err)
	require.NotEmpty(t, createdMealPlanRes.Created.Id)

	mealPlanRes, err := clientToUse.GetMealPlan(ctx, &mealplanninggrpc.GetMealPlanRequest{MealPlanId: createdMealPlanRes.Created.Id})
	require.NoError(t, err)

	actual := converters.ConvertGRPCMealPlanToMealPlan(mealPlanRes.Result)
	checkMealPlanEquality(t, mealPlan, actual)

	return actual
}

func TestMealPlans_AutoFinalizedImmediateGroceryAndTasks(T *testing.T) {
	T.Parallel()

	T.Run("single-option meal plan is finalized at creation and has grocery items and tasks immediately", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		meal := createMealForTest(t, userClient, nil)

		now := time.Now().Truncate(time.Second).UTC()
		inFiveMinutes := now.Add(5 * time.Minute)
		inTenMinutes := now.Add(10 * time.Minute)
		inOneWeek := now.Add(7 * 24 * time.Hour)

		mealPlan := &mealplanning.MealPlan{
			Notes: "auto-finalized test",
			// voting deadline must be before every event's start time; see MealPlanCreationRequestInput.ValidateWithContext.
			VotingDeadline: inFiveMinutes,
			ElectionMethod: mealplanning.MealPlanElectionMethodSchulze,
			Status:         string(mealplanning.MealPlanStatusFinalized),
			// Both flags are false at creation. Creating a finalized plan starts its
			// finalization saga rather than running the pipeline inline, so the read-back
			// below happens before the steps that set them; the awaits further down are what
			// assert they get set.
			Events: []*mealplanning.MealPlanEvent{
				{
					Notes:    "dinner",
					StartsAt: inTenMinutes,
					EndsAt:   inOneWeek,
					MealName: mealplanning.DinnerMealName,
					Options: []*mealplanning.MealPlanOption{
						{
							Meal: mealplanning.Meal{ID: meal.ID},
						},
					},
				},
			},
		}

		created := createMealPlanForTest(t, userClient, mealPlan)
		mealPlanID := created.ID

		// Creating a plan already finalized starts its finalization saga, which builds both of
		// these. It no longer happens before CreateMealPlan returns — that used to mean a user's
		// request waited on the grocery list of every finalized plan in the database.
		assert.NotEmpty(t, awaitMealPlanGroceryListItems(t, ctx, userClient, mealPlanID))
		assert.NotEmpty(t, awaitMealPlanTasks(t, ctx, userClient, mealPlanID))

		// And the flags the steps write with them, which are what stop a second saga — or a
		// second pass of this one — from doing the work again.
		finishedRes, err := userClient.GetMealPlan(ctx, &mealplanninggrpc.GetMealPlanRequest{MealPlanId: mealPlanID})
		require.NoError(t, err)

		finished := converters.ConvertGRPCMealPlanToMealPlan(finishedRes.Result)
		assert.True(t, finished.TasksCreated)
		assert.True(t, finished.GroceryListInitialized)
	})
}

func TestMealPlans_Listing(T *testing.T) {
	T.Parallel()

	T.Run("should be readable in paginated form", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		// Capture timestamp before creating meal plans so we can filter to only our meal plans.
		// Without this filter, other tests' meal plans can fill the first page when running the full suite.
		createdAfter := time.Now().UTC().Add(-5 * time.Minute)
		createdAfterProto := timestamppb.New(createdAfter)

		var expected []*mealplanning.MealPlan
		for range 5 {
			createdMealPlan := createMealPlanForTest(t, userClient, nil)
			expected = append(expected, createdMealPlan)
		}

		// assert meal plan list equality - filter to our meal plans only
		actual, err := userClient.GetMealPlansForAccount(ctx, &mealplanninggrpc.GetMealPlansForAccountRequest{
			Filter: &filtering.QueryFilter{CreatedAfter: createdAfterProto},
		})
		require.NoError(t, err)
		assert.LessOrEqual(
			t,
			len(expected), len(actual.Results),
			"expected %d to be <= %d",
			len(expected),
			len(actual.Results),
		)

		// assert audit log entries for created meal plans (pre-cleanup)
		accountID := getAccountIDForTest(t, userClient)
		expectedAudit := make([]*ExpectedAuditEntry, 0, len(expected))
		for _, mp := range expected {
			expectedAudit = append(expectedAudit, &ExpectedAuditEntry{
				EventType:    "created",
				ResourceType: "meal_plans",
				RelevantID:   mp.ID,
			})
		}
		AssertAuditLogContainsFuzzy(t, ctx, userClient, accountID, 15, expectedAudit)

		for _, createdMealPlan := range expected {
			_, err = userClient.ArchiveMealPlan(ctx, &mealplanninggrpc.ArchiveMealPlanRequest{MealPlanId: createdMealPlan.ID})
			assert.NoError(t, err)
		}
	})
}

func TestMealPlans_CompleteLifecycleForAllVotesReceived(T *testing.T) {
	T.Parallel()

	T.Run("should resolve the meal plan status upon receiving all votes", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create a userClient for the meal plan account
		_, accountAdminUserClient := createUserAndClientForTest(t)

		// create account members
		currentStatus, statusErr := accountAdminUserClient.GetAuthStatus(ctx, &authgrpc.GetAuthStatusRequest{})
		require.NotNil(t, currentStatus)
		require.NoError(t, statusErr)
		relevantAccountID := currentStatus.ActiveAccount

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

			sentInvitations, err := accountAdminUserClient.GetSentAccountInvitations(ctx, &identitygrpc.GetSentAccountInvitationsRequest{})
			require.NotNil(t, sentInvitations)
			require.NoError(t, err)
			assert.NotEmpty(t, sentInvitations.Results)

			invitations, err := c.GetReceivedAccountInvitations(ctx, &identitygrpc.GetReceivedAccountInvitationsRequest{})
			require.NotNil(t, invitations)
			require.NoError(t, err)
			assert.NotEmpty(t, invitations.Results)

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

		// create recipes for meal plan
		createdMeals := []*mealplanning.Meal{}
		for range 3 {
			createdMeal := createMealForTest(t, accountAdminUserClient, nil)
			createdMeals = append(createdMeals, createdMeal)
		}

		const baseDeadline = 10 * time.Second
		now := time.Now()

		exampleMealPlan := &mealplanning.MealPlan{
			Notes:          t.Name(),
			Status:         string(mealplanning.MealPlanStatusAwaitingVotes),
			VotingDeadline: now.Add(baseDeadline),
			ElectionMethod: mealplanning.MealPlanElectionMethodSchulze,
			Events: []*mealplanning.MealPlanEvent{
				{
					StartsAt: now.Add(24 * time.Hour),
					EndsAt:   now.Add(72 * time.Hour),
					MealName: mealplanning.BreakfastMealName,
					Options: []*mealplanning.MealPlanOption{
						{
							Meal:  mealplanning.Meal{ID: createdMeals[0].ID},
							Notes: "option A",
						},
						{
							Meal:  mealplanning.Meal{ID: createdMeals[1].ID},
							Notes: "option B",
						},
						{
							Meal:  mealplanning.Meal{ID: createdMeals[2].ID},
							Notes: "option C",
						},
					},
				},
			},
		}

		createdMealPlan := createMealPlanForTest(t, accountAdminUserClient, exampleMealPlan)
		createdMealPlanEvent := createdMealPlan.Events[0]
		require.NotNil(t, createdMealPlanEvent)

		userAVotes := &mealplanning.MealPlanOptionVoteCreationRequestInput{
			Votes: []*mealplanning.MealPlanOptionVoteCreationInput{
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[0].ID,
					Rank:                    0,
				},
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[1].ID,
					Rank:                    2,
				},
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[2].ID,
					Rank:                    1,
				},
			},
		}

		userBVotes := &mealplanning.MealPlanOptionVoteCreationRequestInput{
			Votes: []*mealplanning.MealPlanOptionVoteCreationInput{
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[0].ID,
					Rank:                    0,
				},
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[1].ID,
					Rank:                    1,
				},
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[2].ID,
					Rank:                    2,
				},
			},
		}

		userCVotes := &mealplanning.MealPlanOptionVoteCreationRequestInput{
			Votes: []*mealplanning.MealPlanOptionVoteCreationInput{
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[0].ID,
					Rank:                    1,
				},
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[1].ID,
					Rank:                    0,
				},
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[2].ID,
					Rank:                    2,
				},
			},
		}

		createdMealPlanOptionVotesA, err := createdClients[0].CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
			Input:           converters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(userAVotes),
		})
		require.NoError(t, err)
		require.NotNil(t, createdMealPlanOptionVotesA)

		createdMealPlanOptionVotesB, err := createdClients[1].CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
			Input:           converters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(userBVotes),
		})
		require.NoError(t, err)
		require.NotNil(t, createdMealPlanOptionVotesB)

		createdMealPlanOptionVotesC, err := accountAdminUserClient.CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
			Input:           converters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(userCVotes),
		})
		require.NoError(t, err)
		require.NotNil(t, createdMealPlanOptionVotesC)

		createdMealPlan.VotingDeadline = time.Now().Add(-time.Minute)

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

		// The admin endpoint starts a saga per finalizable plan; its first step does the tally.
		runRes, err := adminClient.RunFinalizeMealPlanWorker(ctx, &mealplanninggrpc.RunFinalizeMealPlanWorkerRequest{})
		require.NoError(t, err)
		require.NotNil(t, runRes)

		actual := awaitMealPlanFinalized(t, ctx, accountAdminUserClient, createdMealPlan.ID)

		for _, event := range actual.Events {
			selectionMade := false
			for _, opt := range event.Options {
				if opt.Chosen {
					selectionMade = true
					break
				}
			}
			require.True(t, selectionMade)
		}
	})
}

func TestMealPlans_UpdateMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		createdMealPlan := createMealPlanForTest(t, userClient, nil)

		updatedNotes := "updated notes for meal plan"

		_, err := userClient.UpdateMealPlan(ctx, &mealplanninggrpc.UpdateMealPlanRequest{
			MealPlanId: createdMealPlan.ID,
			Input: &mealplanninggrpc.MealPlanUpdateRequestInput{
				Notes: new(updatedNotes),
			},
		})
		require.NoError(t, err)

		// Verify via a fresh GET
		fetchRes, err := userClient.GetMealPlan(ctx, &mealplanninggrpc.GetMealPlanRequest{MealPlanId: createdMealPlan.ID})
		require.NoError(t, err)
		fetched := converters.ConvertGRPCMealPlanToMealPlan(fetchRes.Result)
		assert.Equal(t, updatedNotes, fetched.Notes)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		updatedNotes := "should fail"
		res, err := c.UpdateMealPlan(ctx, &mealplanninggrpc.UpdateMealPlanRequest{
			MealPlanId: nonexistentID,
			Input: &mealplanninggrpc.MealPlanUpdateRequestInput{
				Notes: new(updatedNotes),
			},
		})
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	T.Run("nonexistent meal plan", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		updatedNotes := "should fail"
		res, err := userClient.UpdateMealPlan(ctx, &mealplanninggrpc.UpdateMealPlanRequest{
			MealPlanId: nonexistentID,
			Input: &mealplanninggrpc.MealPlanUpdateRequestInput{
				Notes: new(updatedNotes),
			},
		})
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestMealPlans_FinalizeMealPlan(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Create a user and their client
		_, accountAdminUserClient := createUserAndClientForTest(t)

		// Get account ID
		currentStatus, statusErr := accountAdminUserClient.GetAuthStatus(ctx, &authgrpc.GetAuthStatusRequest{})
		require.NotNil(t, currentStatus)
		require.NoError(t, statusErr)
		relevantAccountID := currentStatus.ActiveAccount

		// Create 2 additional household members (total 3 including admin)
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

		// Create meals for meal plan options
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
						{
							Meal:  mealplanning.Meal{ID: createdMeals[0].ID},
							Notes: "option A",
						},
						{
							Meal:  mealplanning.Meal{ID: createdMeals[1].ID},
							Notes: "option B",
						},
						{
							Meal:  mealplanning.Meal{ID: createdMeals[2].ID},
							Notes: "option C",
						},
					},
				},
			},
		}

		createdMealPlan := createMealPlanForTest(t, accountAdminUserClient, exampleMealPlan)
		createdMealPlanEvent := createdMealPlan.Events[0]
		require.NotNil(t, createdMealPlanEvent)

		// All 3 users vote
		allVotes := []*mealplanning.MealPlanOptionVoteCreationRequestInput{
			{
				Votes: []*mealplanning.MealPlanOptionVoteCreationInput{
					{BelongsToMealPlanOption: createdMealPlanEvent.Options[0].ID, Rank: 0},
					{BelongsToMealPlanOption: createdMealPlanEvent.Options[1].ID, Rank: 2},
					{BelongsToMealPlanOption: createdMealPlanEvent.Options[2].ID, Rank: 1},
				},
			},
			{
				Votes: []*mealplanning.MealPlanOptionVoteCreationInput{
					{BelongsToMealPlanOption: createdMealPlanEvent.Options[0].ID, Rank: 0},
					{BelongsToMealPlanOption: createdMealPlanEvent.Options[1].ID, Rank: 1},
					{BelongsToMealPlanOption: createdMealPlanEvent.Options[2].ID, Rank: 2},
				},
			},
			{
				Votes: []*mealplanning.MealPlanOptionVoteCreationInput{
					{BelongsToMealPlanOption: createdMealPlanEvent.Options[0].ID, Rank: 1},
					{BelongsToMealPlanOption: createdMealPlanEvent.Options[1].ID, Rank: 0},
					{BelongsToMealPlanOption: createdMealPlanEvent.Options[2].ID, Rank: 2},
				},
			},
		}

		// Client A votes
		_, err := createdClients[0].CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
			Input:           converters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(allVotes[0]),
		})
		require.NoError(t, err)

		// Client B votes
		_, err = createdClients[1].CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
			Input:           converters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(allVotes[1]),
		})
		require.NoError(t, err)

		// Admin votes
		_, err = accountAdminUserClient.CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
			Input:           converters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(allVotes[2]),
		})
		require.NoError(t, err)

		// Set voting deadline to the past so FinalizeMealPlan can finalize
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

		// Call FinalizeMealPlan directly (not the worker)
		finalizeRes, err := accountAdminUserClient.FinalizeMealPlan(ctx, &mealplanninggrpc.FinalizeMealPlanRequest{
			MealPlanId: createdMealPlan.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, finalizeRes)
		assert.True(t, finalizeRes.Finalized)

		// Verify the meal plan is now finalized
		fetchRes, err := accountAdminUserClient.GetMealPlan(ctx, &mealplanninggrpc.GetMealPlanRequest{MealPlanId: createdMealPlan.ID})
		require.NoError(t, err)
		actual := converters.ConvertGRPCMealPlanToMealPlan(fetchRes.Result)
		assert.Equal(t, string(mealplanning.MealPlanStatusFinalized), actual.Status)

		// Verify a selection was made
		for _, event := range actual.Events {
			selectionMade := false
			for _, opt := range event.Options {
				if opt.Chosen {
					selectionMade = true
					break
				}
			}
			require.True(t, selectionMade)
		}
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		res, err := c.FinalizeMealPlan(ctx, &mealplanninggrpc.FinalizeMealPlanRequest{
			MealPlanId: nonexistentID,
		})
		assert.Error(t, err)
		assert.Nil(t, res)
	})

	T.Run("nonexistent meal plan", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		res, err := userClient.FinalizeMealPlan(ctx, &mealplanninggrpc.FinalizeMealPlanRequest{
			MealPlanId: nonexistentID,
		})
		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestMealPlans_CompleteLifecycleForSomeVotesReceived(T *testing.T) {
	T.Parallel()

	T.Run("should resolve the meal plan status upon receiving some votes", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// create a userClient for the meal plan account
		_, accountAdminUserClient := createUserAndClientForTest(t)

		// create account members
		currentStatus, statusErr := accountAdminUserClient.GetAuthStatus(ctx, &authgrpc.GetAuthStatusRequest{})
		require.NotNil(t, currentStatus)
		require.NoError(t, statusErr)
		relevantAccountID := currentStatus.ActiveAccount

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

			sentInvitations, err := accountAdminUserClient.GetSentAccountInvitations(ctx, &identitygrpc.GetSentAccountInvitationsRequest{})
			require.NotNil(t, sentInvitations)
			require.NoError(t, err)
			assert.NotEmpty(t, sentInvitations.Results)

			invitations, err := c.GetReceivedAccountInvitations(ctx, &identitygrpc.GetReceivedAccountInvitationsRequest{})
			require.NotNil(t, invitations)
			require.NoError(t, err)
			assert.NotEmpty(t, invitations.Results)

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

		// create recipes for meal plan
		createdMeals := []*mealplanning.Meal{}
		for range 3 {
			createdMeal := createMealForTest(t, accountAdminUserClient, nil)
			createdMeals = append(createdMeals, createdMeal)
		}

		const baseDeadline = time.Minute
		now := time.Now()

		exampleMealPlan := &mealplanning.MealPlan{
			Notes:          t.Name(),
			Status:         string(mealplanning.MealPlanStatusAwaitingVotes),
			VotingDeadline: now.Add(baseDeadline),
			ElectionMethod: mealplanning.MealPlanElectionMethodSchulze,
			Events: []*mealplanning.MealPlanEvent{
				{
					StartsAt: now.Add(24 * time.Hour),
					EndsAt:   now.Add(72 * time.Hour),
					MealName: mealplanning.BreakfastMealName,
					Options: []*mealplanning.MealPlanOption{
						{
							Meal:  mealplanning.Meal{ID: createdMeals[0].ID},
							Notes: "option A",
						},
						{
							Meal:  mealplanning.Meal{ID: createdMeals[1].ID},
							Notes: "option B",
						},
						{
							Meal:  mealplanning.Meal{ID: createdMeals[2].ID},
							Notes: "option C",
						},
					},
				},
			},
		}

		exampleMealPlanInput := mpconverters.ConvertMealPlanToMealPlanCreationRequestInput(exampleMealPlan)
		createMealPlanRes, err := accountAdminUserClient.CreateMealPlan(ctx, &mealplanninggrpc.CreateMealPlanRequest{Input: converters.ConvertMealPlanCreationRequestInputToGRPCMealPlanCreationRequestInput(exampleMealPlanInput)})
		require.NotEmpty(t, createMealPlanRes.Created.Id)
		require.NoError(t, err)

		createdMealPlanRes, err := accountAdminUserClient.GetMealPlan(ctx, &mealplanninggrpc.GetMealPlanRequest{MealPlanId: createMealPlanRes.Created.Id})
		require.NotNil(t, createdMealPlanRes)
		require.NoError(t, err)

		createdMealPlan := converters.ConvertGRPCMealPlanToMealPlan(createdMealPlanRes.Result)

		checkMealPlanEquality(t, exampleMealPlan, createdMealPlan)

		createdMealPlanEvent := createdMealPlan.Events[0]

		userAVotes := &mealplanning.MealPlanOptionVoteCreationRequestInput{
			Votes: []*mealplanning.MealPlanOptionVoteCreationInput{
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[0].ID,
					Rank:                    0,
				},
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[1].ID,
					Rank:                    2,
				},
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[2].ID,
					Rank:                    1,
				},
			},
		}

		userBVotes := &mealplanning.MealPlanOptionVoteCreationRequestInput{
			Votes: []*mealplanning.MealPlanOptionVoteCreationInput{
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[0].ID,
					Rank:                    0,
				},
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[1].ID,
					Rank:                    1,
				},
				{
					BelongsToMealPlanOption: createdMealPlanEvent.Options[2].ID,
					Rank:                    2,
				},
			},
		}

		createdMealPlanOptionVotesA, err := createdClients[0].CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
			Input:           converters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(userAVotes),
		})
		require.NoError(t, err)
		require.NotNil(t, createdMealPlanOptionVotesA)

		createdMealPlanOptionVotesB, err := createdClients[1].CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      createdMealPlan.ID,
			MealPlanEventId: createdMealPlanEvent.ID,
			Input:           converters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(userBVotes),
		})
		require.NoError(t, err)
		require.NotNil(t, createdMealPlanOptionVotesB)

		createdMealPlanRes, err = accountAdminUserClient.GetMealPlan(ctx, &mealplanninggrpc.GetMealPlanRequest{MealPlanId: createdMealPlan.ID})
		require.NotNil(t, createdMealPlanRes)
		require.NoError(t, err)

		createdMealPlan = converters.ConvertGRPCMealPlanToMealPlan(createdMealPlanRes.Result)
		assert.Equal(t, string(mealplanning.MealPlanStatusAwaitingVotes), createdMealPlan.Status)

		q := generated.New()
		rowsAffected, err := q.UpdateMealPlan(ctx, databaseClient.Writer(), &generated.UpdateMealPlanParams{
			Notes:            createdMealPlan.Notes,
			Status:           generated.MealPlanStatus(mealplanning.MealPlanStatusAwaitingVotes),
			VotingDeadline:   time.Now().Add(-10 * time.Hour),
			BelongsToAccount: relevantAccountID,
			ID:               createdMealPlan.ID,
		})
		require.NoError(t, err)
		require.NotZero(t, rowsAffected)

		time.Sleep(10 * time.Second)

		runRes, err := adminClient.RunFinalizeMealPlanWorker(ctx, &mealplanninggrpc.RunFinalizeMealPlanWorkerRequest{})
		require.NoError(t, err)
		require.NotNil(t, runRes)

		createdMealPlan = awaitMealPlanFinalized(t, ctx, accountAdminUserClient, createdMealPlan.ID)

		for _, event := range createdMealPlan.Events {
			selectionMade := false
			for _, opt := range event.Options {
				if opt.Chosen {
					selectionMade = true
					break
				}
			}
			require.True(t, selectionMade)
		}
	})
}
