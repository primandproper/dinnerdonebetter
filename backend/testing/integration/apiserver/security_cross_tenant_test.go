package integration

import (
	"testing"

	mpconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mpfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	mpgrpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	"github.com/primandproper/platform-go/v14/identity/identitypb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// This file holds NEW negative / cross-tenant integration tests that positively assert the
// IDOR / authorization denials introduced on branch `code-review-fixes-2026-07`. Each test
// provisions TWO users in SEPARATE accounts (A and B) and asserts that user B is DENIED access
// to user A's resource, alongside a positive control proving the check is scoped (the legitimate
// owner still succeeds) rather than globally broken. These tests are designed to FAIL against the
// pre-fix behavior and PASS now.
//
// INTENTIONALLY OMITTED (documented so the gap is visible):
//
//   - H9 identity ArchiveUserMembership: the fix makes the handler remove the member from
//     the caller's OWN active account (request.AccountId is no longer trusted) rather than returning a
//     denial. Proving the cross-tenant property cleanly requires seeding a second membership in account
//     A and asserting B's call cannot touch it. This too needs membership seeding and is omitted rather
//     than written as a flaky test.
//
// Webhook subscriptions, audit entries and billing reads have no case here: platform's
// conformance suites, run against this deployment in conformance_test.go, assert them. The webhooks
// suite that a neighbor's archive touches neither an endpoint nor its subscriptions; the audit
// suite that a listing, an actor query and a read by id are all confined to the caller's own
// chain, a neighbor's entry reading as an id nobody wrote; and the billing suite that naming
// another account's subscriptions, purchases or transactions is refused.

// getActiveAccountIDForClientForTest returns the active account ID for the given client's session.
func getActiveAccountIDForClientForTest(t *testing.T, resp *authsvc.GetActiveAccountResponse) string {
	t.Helper()
	require.NotNil(t, resp)
	require.NotNil(t, resp.Result)
	require.NotEmpty(t, resp.Result.Id)
	return resp.Result.Id
}

// TestCrossTenant_GetAccount_Denied asserts that a user cannot read an account they are not a member
// of. identity/grpc/accounts.go GetAccount returns codes.PermissionDenied (errNotAuthorizedForAccount)
// for non-members.
func TestCrossTenant_GetAccount_Denied(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, clientA := createUserAndClientForTest(t)
		_, clientB := createUserAndClientForTest(t)

		activeA, err := clientA.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		accountAID := getActiveAccountIDForClientForTest(t, activeA)

		// positive control: A can read its own account.
		ownAccount, err := clientA.IdentityService().GetAccount(ctx, &identitypb.GetAccountRequest{AccountId: accountAID})
		require.NoError(t, err)
		require.NotNil(t, ownAccount)
		assert.Equal(t, accountAID, ownAccount.GetAccount().GetId())

		// cross-tenant: B is not a member of A's account.
		_, err = clientB.IdentityService().GetAccount(ctx, &identitypb.GetAccountRequest{AccountId: accountAID})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})
}

// TestCrossTenant_RecipeRating_Denied asserts that a user cannot mutate a recipe rating authored by
// another user. mealplanning/grpc/recipes.go verifyRecipeRatingOwnership returns codes.PermissionDenied
// when rating.CreatedByUser != requester.
func TestCrossTenant_RecipeRating_Denied(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)

		_, clientA := createUserAndClientForTest(t)
		_, clientB := createUserAndClientForTest(t)

		// A authors the rating.
		rating := createRecipeRatingForTest(t, createdRecipe.ID, clientA)

		// cross-tenant: B may not update A's rating.
		newRating := mpfakes.BuildFakeRecipeRating()
		newRating.BelongsToRecipe = createdRecipe.ID
		updateInput := mpconverters.ConvertRecipeRatingToRecipeRatingUpdateRequestInput(newRating)

		_, err := clientB.UpdateRecipeRating(ctx, &mealplanninggrpc.UpdateRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: rating.ID,
			Input:          mpgrpcconverters.ConvertRecipeRatingUpdateRequestInputToGRPCRecipeRatingUpdateRequestInput(updateInput),
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// cross-tenant: B may not archive A's rating.
		_, err = clientB.ArchiveRecipeRating(ctx, &mealplanninggrpc.ArchiveRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: rating.ID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// positive control: A can update its own rating.
		ownUpdate := mpfakes.BuildFakeRecipeRating()
		ownUpdate.BelongsToRecipe = createdRecipe.ID
		ownUpdate.CreatedByUser = rating.CreatedByUser
		ownUpdateInput := mpconverters.ConvertRecipeRatingToRecipeRatingUpdateRequestInput(ownUpdate)

		_, err = clientA.UpdateRecipeRating(ctx, &mealplanninggrpc.UpdateRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: rating.ID,
			Input:          mpgrpcconverters.ConvertRecipeRatingUpdateRequestInputToGRPCRecipeRatingUpdateRequestInput(ownUpdateInput),
		})
		require.NoError(t, err)

		// positive control: A can archive its own rating.
		_, err = clientA.ArchiveRecipeRating(ctx, &mealplanninggrpc.ArchiveRecipeRatingRequest{
			RecipeId:       createdRecipe.ID,
			RecipeRatingId: rating.ID,
		})
		require.NoError(t, err)
	})
}

// TestCrossTenant_MealLists_NotLeaked asserts that meal lists are user-scoped: user B's GetMealLists
// never returns user A's meal lists. GetMealLists filters by the session user's ID (belongs_to_user),
// so this is a "no leak" property rather than a hard denial. Meal list items are returned nested inside
// each meal list (there is no separately-registered GetMealListItems RPC), so scoping GetMealLists also
// prevents leaking A's items: since B never sees A's list, it never sees A's items either.
func TestCrossTenant_MealLists_NotLeaked(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, clientA := createUserAndClientForTest(t)
		_, clientB := createUserAndClientForTest(t)

		// A creates a meal list.
		createRes, err := clientA.CreateMealList(ctx, &mealplanninggrpc.CreateMealListRequest{
			Input: &mealplanninggrpc.MealListCreationRequestInput{Name: t.Name(), Description: "desc"},
		})
		require.NoError(t, err)
		listAID := createRes.Created.Id
		require.NotEmpty(t, listAID)

		// positive control: A sees its own list.
		ownLists, err := clientA.GetMealLists(ctx, &mealplanninggrpc.GetMealListsRequest{})
		require.NoError(t, err)
		var ownFound bool
		for _, l := range ownLists.Results {
			if l.Id == listAID {
				ownFound = true
				break
			}
		}
		assert.True(t, ownFound, "A should see its own meal list")

		// cross-tenant: B must not see A's list.
		bLists, err := clientB.GetMealLists(ctx, &mealplanninggrpc.GetMealListsRequest{})
		require.NoError(t, err)
		for _, l := range bLists.Results {
			assert.NotEqual(t, listAID, l.Id, "B must not see A's meal list %q", listAID)
		}
	})
}

// TestCrossTenant_MealPlanRecipeOptionSelections_Denied asserts that the recipe-option-selection
// handlers cannot be used against another account's meal plan option. The requests carry only a
// MealPlanOptionId, so the service resolves the option through its event and meal plan to an
// account (verifyMealPlanOptionAccess) and returns codes.NotFound when it does not belong to the
// caller's active account.
func TestCrossTenant_MealPlanRecipeOptionSelections_Denied(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// A owns a meal plan whose single option supports selections.
		setup := createMealPlanWithAlternativeIngredientsForSelectionTests(t)
		clientA := setup.userClient

		// A creates a selection on its own option (positive control for create).
		createRes, err := clientA.CreateMealPlanRecipeOptionSelection(ctx, &mealplanninggrpc.CreateMealPlanRecipeOptionSelectionRequest{
			MealPlanOptionId: setup.mealPlanOptionID,
			Input: &mealplanninggrpc.MealPlanRecipeOptionSelectionCreationRequestInput{
				RecipeId:            setup.recipe.ID,
				RecipeStepId:        setup.recipe.Steps[0].ID,
				IngredientIndex:     0,
				SelectedOptionIndex: 1,
				SelectionType:       mealplanninggrpc.MealPlanRecipeOptionSelectionType_MEAL_PLAN_RECIPE_OPTION_SELECTION_TYPE_INGREDIENT,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, createRes.Created)

		_, clientB := createUserAndClientForTest(t)

		// cross-tenant: B cannot read A's selection.
		_, err = clientB.GetMealPlanRecipeOptionSelection(ctx, &mealplanninggrpc.GetMealPlanRecipeOptionSelectionRequest{
			MealPlanOptionId: setup.mealPlanOptionID,
			RecipeStepId:     setup.recipe.Steps[0].ID,
			IngredientIndex:  0,
			SelectionType:    mealplanninggrpc.MealPlanRecipeOptionSelectionType_MEAL_PLAN_RECIPE_OPTION_SELECTION_TYPE_INGREDIENT,
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))

		// cross-tenant: B cannot list selections for A's option.
		_, err = clientB.GetMealPlanRecipeOptionSelectionsForMealPlanOption(ctx, &mealplanninggrpc.GetMealPlanRecipeOptionSelectionsForMealPlanOptionRequest{
			MealPlanOptionId: setup.mealPlanOptionID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))

		// cross-tenant: B cannot create a selection on A's option.
		_, err = clientB.CreateMealPlanRecipeOptionSelection(ctx, &mealplanninggrpc.CreateMealPlanRecipeOptionSelectionRequest{
			MealPlanOptionId: setup.mealPlanOptionID,
			Input: &mealplanninggrpc.MealPlanRecipeOptionSelectionCreationRequestInput{
				RecipeId:            setup.recipe.ID,
				RecipeStepId:        setup.recipe.Steps[0].ID,
				IngredientIndex:     1,
				SelectedOptionIndex: 0,
				SelectionType:       mealplanninggrpc.MealPlanRecipeOptionSelectionType_MEAL_PLAN_RECIPE_OPTION_SELECTION_TYPE_INGREDIENT,
			},
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))

		// cross-tenant: B cannot update A's selection.
		_, err = clientB.UpdateMealPlanRecipeOptionSelection(ctx, &mealplanninggrpc.UpdateMealPlanRecipeOptionSelectionRequest{
			MealPlanOptionId: setup.mealPlanOptionID,
			RecipeStepId:     setup.recipe.Steps[0].ID,
			IngredientIndex:  0,
			SelectionType:    mealplanninggrpc.MealPlanRecipeOptionSelectionType_MEAL_PLAN_RECIPE_OPTION_SELECTION_TYPE_INGREDIENT,
			Input: &mealplanninggrpc.MealPlanRecipeOptionSelectionUpdateRequestInput{
				SelectedOptionIndex: 0,
			},
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))

		// cross-tenant: B cannot archive A's selection.
		_, err = clientB.ArchiveMealPlanRecipeOptionSelection(ctx, &mealplanninggrpc.ArchiveMealPlanRecipeOptionSelectionRequest{
			MealPlanOptionId: setup.mealPlanOptionID,
			RecipeStepId:     setup.recipe.Steps[0].ID,
			IngredientIndex:  0,
			SelectionType:    mealplanninggrpc.MealPlanRecipeOptionSelectionType_MEAL_PLAN_RECIPE_OPTION_SELECTION_TYPE_INGREDIENT,
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))

		// security property + positive control: A's selection survives and remains readable.
		getRes, err := clientA.GetMealPlanRecipeOptionSelection(ctx, &mealplanninggrpc.GetMealPlanRecipeOptionSelectionRequest{
			MealPlanOptionId: setup.mealPlanOptionID,
			RecipeStepId:     setup.recipe.Steps[0].ID,
			IngredientIndex:  0,
			SelectionType:    mealplanninggrpc.MealPlanRecipeOptionSelectionType_MEAL_PLAN_RECIPE_OPTION_SELECTION_TYPE_INGREDIENT,
		})
		require.NoError(t, err)
		require.NotNil(t, getRes.Result)
		assert.Equal(t, createRes.Created.Id, getRes.Result.Id)
		assert.Equal(t, uint32(1), getRes.Result.SelectedOptionIndex, "B's cross-tenant update attempt must not have modified A's selection")
	})
}

// TestCrossTenant_MealPlanOptionVotes_Denied asserts the H18 vote-scoping fix: a vote's target
// option must belong to the meal plan event named in the request, and the meal plan must belong to
// the caller's account. User B cannot vote on user A's option — neither by naming A's plan (denied
// at the plan-access check) nor by smuggling A's option ID under B's own plan (denied at the
// option-resolution check). NOTE: the companion eligibility gate (votes rejected once the plan
// leaves 'awaiting_votes') is covered by manager unit tests; driving a plan out of awaiting_votes
// deterministically requires the finalization worker, which this harness does not run.
func TestCrossTenant_MealPlanOptionVotes_Denied(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// A owns a meal plan with at least one votable option.
		_, clientA := createUserAndClientForTest(t)
		mealPlanA := createMealPlanForTest(t, clientA, nil)
		require.NotEmpty(t, mealPlanA.Events)
		require.NotEmpty(t, mealPlanA.Events[0].Options)
		optionA := mealPlanA.Events[0].Options[0]

		// B owns an unrelated meal plan of their own.
		_, clientB := createUserAndClientForTest(t)
		mealPlanB := createMealPlanForTest(t, clientB, nil)
		require.NotEmpty(t, mealPlanB.Events)
		require.NotEmpty(t, mealPlanB.Events[0].Options)
		eventB := mealPlanB.Events[0]

		// cross-tenant: B cannot vote by naming A's plan and event directly.
		voteOnA := mpfakes.BuildFakeMealPlanOptionVote()
		voteOnA.BelongsToMealPlanOption = optionA.ID
		voteOnAInput := mpconverters.ConvertMealPlanOptionVoteToMealPlanOptionVoteCreationRequestInput(voteOnA)
		_, err := clientB.CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      mealPlanA.ID,
			MealPlanEventId: mealPlanA.Events[0].ID,
			Input:           mpgrpcconverters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(voteOnAInput),
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))

		// cross-tenant: B cannot smuggle A's option ID under B's own (accessible) plan and event.
		_, err = clientB.CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      mealPlanB.ID,
			MealPlanEventId: eventB.ID,
			Input:           mpgrpcconverters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(voteOnAInput),
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))

		// positive control: B voting on B's own option succeeds.
		voteOnB := mpfakes.BuildFakeMealPlanOptionVote()
		voteOnB.BelongsToMealPlanOption = eventB.Options[0].ID
		voteOnBInput := mpconverters.ConvertMealPlanOptionVoteToMealPlanOptionVoteCreationRequestInput(voteOnB)
		createRes, err := clientB.CreateMealPlanOptionVote(ctx, &mealplanninggrpc.CreateMealPlanOptionVoteRequest{
			MealPlanId:      mealPlanB.ID,
			MealPlanEventId: eventB.ID,
			Input:           mpgrpcconverters.ConvertMealPlanOptionVoteCreationRequestInputToGRPCMealPlanOptionVoteCreationRequestInput(voteOnBInput),
		})
		require.NoError(t, err)
		require.NotEmpty(t, createRes.Created)

		// security property: no vote from B landed on A's option.
		votesOnA, err := clientA.GetMealPlanOptionVotes(ctx, &mealplanninggrpc.GetMealPlanOptionVotesRequest{
			MealPlanId:       mealPlanA.ID,
			MealPlanEventId:  mealPlanA.Events[0].ID,
			MealPlanOptionId: optionA.ID,
		})
		require.NoError(t, err)
		assert.Empty(t, votesOnA.Results)
	})
}
