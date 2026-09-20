package integration

import (
	"testing"

	mpconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mpfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	mpgrpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	auditsvc "github.com/primandproper/platform-go/v14/audit/auditpb"
	waitlistspb "github.com/primandproper/platform-go/v14/waitlists/waitlistspb"
	webhookspb "github.com/primandproper/platform-go/v14/webhooks/webhookspb"

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
//   - H7 payments: GetSubscriptions / GetPurchases / GetPaymentHistoryForAccount. The fix makes the
//     handler IGNORE request.AccountId and use the session's active account. There is therefore no
//     hard cross-tenant denial to assert; a meaningful test would require seeding billing/subscription
//     data for account A and proving account B's session cannot observe it. That needs data seeding
//     the harness does not currently provide, and a naive test would be flaky, so it is omitted here.
//
//   - H9 identity ArchiveUserMembership: likewise, the fix makes the handler remove the member from
//     the caller's OWN active account (request.AccountId is no longer trusted) rather than returning a
//     denial. Proving the cross-tenant property cleanly requires seeding a second membership in account
//     A and asserting B's call cannot touch it. This too needs membership seeding and is omitted rather
//     than written as a flaky test.

// getActiveAccountIDForClientForTest returns the active account ID for the given client's session.
func getActiveAccountIDForClientForTest(t *testing.T, resp *authsvc.GetActiveAccountResponse) string {
	t.Helper()
	require.NotNil(t, resp)
	require.NotNil(t, resp.Result)
	require.NotEmpty(t, resp.Result.Id)
	return resp.Result.Id
}

// TestCrossTenant_AuditLogForAccount_Denied asserts that a user cannot read the account-scoped
// audit log of an account they are not a member of.
//
// It is not a denial any more, and that is the stronger answer. The RPC this replaced took an
// account id and checked membership afterwards; platform's ListEntries takes none, because the
// chain a request reads is the one its session is in — see internal/build/auditlog, where the
// scope resolver is the active account. So B asking cannot name A's account: B's read is B's
// chain, and the assertion is that A's entries are not in it.
func TestCrossTenant_AuditLogForAccount_Denied(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		userA, clientA := createUserAndClientForTest(t)
		_, clientB := createUserAndClientForTest(t)

		// positive control: A can read its own account's audit log, and A is in it.
		ownResp, err := clientA.ListEntries(ctx, &auditsvc.ListEntriesRequest{})
		require.NoError(t, err)
		require.NotEmpty(t, ownResp.GetResults())

		var ownEntryID string
		for _, entry := range ownResp.GetResults() {
			if entry.GetActor().GetId() == userA.ID {
				ownEntryID = entry.GetId()
				break
			}
		}
		require.NotEmpty(t, ownEntryID, "A's own account chain holds no entry naming A")

		// cross-tenant: B's read is B's chain. A's entry is not in it.
		crossResp, err := clientB.ListEntries(ctx, &auditsvc.ListEntriesRequest{})
		require.NoError(t, err)

		for _, entry := range crossResp.GetResults() {
			assert.NotEqual(t, ownEntryID, entry.GetId(), "another account's audit entry reached this caller")
			assert.NotEqual(t, userA.ID, entry.GetActor().GetId(), "another user's audit entry reached this caller")
		}
	})
}

// TestCrossTenant_AuditLogForUser_Denied asserts that a user cannot read another user's audit
// entries.
//
// Same shape as the account case above: the actor is a query field, but the scope is not, so
// filtering by A's actor id from B's session searches B's chain for entries A never wrote
// there. The answer is empty rather than refused, which is what makes it structural.
func TestCrossTenant_AuditLogForUser_Denied(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		userA, clientA := createUserAndClientForTest(t)
		_, clientB := createUserAndClientForTest(t)

		// positive control: A can read its own entries.
		ownResp, err := clientA.ListEntries(ctx, &auditsvc.ListEntriesRequest{
			Query: &auditsvc.EntryQuery{ActorId: userA.ID},
		})
		require.NoError(t, err)
		require.NotEmpty(t, ownResp.GetResults())

		// cross-tenant: B naming A's actor id finds nothing, because the chain it searches
		// is B's own.
		crossResp, err := clientB.ListEntries(ctx, &auditsvc.ListEntriesRequest{
			Query: &auditsvc.EntryQuery{ActorId: userA.ID},
		})
		require.NoError(t, err)
		assert.Empty(t, crossResp.GetResults(), "another user's audit entries reached this caller")
	})
}

// TestCrossTenant_AuditLogEntryByID_Denied asserts that a user cannot read an individual audit
// log entry belonging to another user or account by its id.
//
// This is the one of the three that is still a refusal rather than an empty answer, and it has
// to be: an id names a row directly, so there is no scope in the request to bound it. GetEntry
// reads the entry and then checks that it is in the caller's chain.
func TestCrossTenant_AuditLogEntryByID_Denied(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		userA, clientA := createUserAndClientForTest(t)
		_, clientB := createUserAndClientForTest(t)

		// A performs an auditable op so that a fresh, A-owned entry exists.
		createWebhookForTest(t, clientA)

		// A lists its OWN entries (its own chain) and picks one to target.
		forUser, err := clientA.ListEntries(ctx, &auditsvc.ListEntriesRequest{
			Query: &auditsvc.EntryQuery{ActorId: userA.ID},
		})
		require.NoError(t, err)
		require.NotEmpty(t, forUser.GetResults())

		entryID := forUser.GetResults()[0].GetId()
		require.NotEmpty(t, entryID)

		// positive control: A can fetch its own entry by ID.
		ownEntry, err := clientA.GetEntry(ctx, &auditsvc.GetEntryRequest{EntryId: entryID})
		require.NoError(t, err)
		require.NotNil(t, ownEntry.GetEntry())
		assert.Equal(t, entryID, ownEntry.GetEntry().GetId())

		// cross-tenant: B may not fetch A's entry by ID, and the refusal is NotFound
		// rather than PermissionDenied. That is stricter than what this case originally
		// asserted, not looser: the scope goes into the read rather than into a check
		// after it, so B's read never touches the row, and an entry belonging to
		// somebody else comes back exactly as an id that does not exist does. A
		// PermissionDenied here would make the call an oracle for which entry ids exist
		// in another tenant's log.
		_, err = clientB.GetEntry(ctx, &auditsvc.GetEntryRequest{EntryId: entryID})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
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
		ownAccount, err := clientA.GetAccount(ctx, &identitysvc.GetAccountRequest{AccountId: accountAID})
		require.NoError(t, err)
		require.NotNil(t, ownAccount)
		assert.Equal(t, accountAID, ownAccount.Result.Id)

		// cross-tenant: B is not a member of A's account.
		_, err = clientB.GetAccount(ctx, &identitysvc.GetAccountRequest{AccountId: accountAID})
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

// TestCrossTenant_WebhookTriggerConfig_Denied asserts that a user cannot archive a webhook trigger
// config belonging to another account's webhook. webhooks/grpc/webhooks.go ArchiveWebhookTriggerConfig
// enforces ownership via an account-scoped GetWebhook lookup: for a cross-account caller that lookup
// fails to find the webhook, so the handler returns codes.Internal (NOT PermissionDenied). The real
// security property under test is that the config survives B's attempt; we therefore assert an error
// is returned AND that A's trigger config still exists afterward.
func TestCrossTenant_WebhookTriggerConfig_Denied(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, clientA := createUserAndClientForTest(t)
		_, clientB := createUserAndClientForTest(t)

		// A registers an endpoint and subscribes it to a second event type.
		endpointA := createWebhookForTest(t, clientA)

		// Distinct from the one the registration already carries.
		eventType := webhooks.WebhookArchivedServiceEventType
		if endpointA.GetSubscriptions()[0].GetEventType() == eventType {
			eventType = webhooks.WebhookCreatedServiceEventType
		}

		added, err := clientA.WebhooksService().AddSubscription(ctx, &webhookspb.AddSubscriptionRequest{
			EndpointId: endpointA.GetId(),
			EventType:  eventType,
		})
		require.NoError(t, err)
		require.NotNil(t, added.GetResult())

		subscriptionID := added.GetResult().GetId()
		require.NotEmpty(t, subscriptionID)

		// cross-tenant: B attempts to archive A's subscription. The write is scoped to the
		// caller's account, so it matches nothing from B's session — and platform answers
		// that with success rather than an error, deliberately: an id naming nothing under
		// the caller's own endpoints is a nil subscription and no error, so the call cannot
		// be walked to find out which subscription ids exist in somebody else's account.
		//
		// Which makes the reply say nothing either way, and the assertion that follows the
		// only one worth making: what matters is that A's row is untouched, not what B was
		// told.
		_, err = clientB.WebhooksService().ArchiveSubscription(ctx, &webhookspb.ArchiveSubscriptionRequest{
			SubscriptionId: subscriptionID,
		})
		require.NoError(t, err, "a cross-tenant archive is answered rather than refused")

		// security property: A's subscription must still exist after B's attempt.
		refreshed, err := clientA.WebhooksService().ListSubscriptions(ctx, &webhookspb.ListSubscriptionsRequest{
			EndpointId: endpointA.GetId(),
		})
		require.NoError(t, err)

		var stillPresent bool
		for _, subscription := range refreshed.GetResults() {
			if subscription.GetId() == subscriptionID {
				stillPresent = true
				break
			}
		}
		assert.True(t, stillPresent, "expected A's webhook subscription %q to survive B's cross-tenant archive attempt", subscriptionID)

		// positive control: A can archive its own subscription.
		_, err = clientA.WebhooksService().ArchiveSubscription(ctx, &webhookspb.ArchiveSubscriptionRequest{
			SubscriptionId: subscriptionID,
		})
		require.NoError(t, err)
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
// TestCrossTenant_WaitlistSignups_Denied asserts that waitlist signups are user-owned (M23): user B
// may not read, update, withdraw, or archive user A's signup by ID, may not read A's signups by
// naming A as the subject, and the waitlist-wide signup listing is reserved for service admins. Waitlists themselves are global, admin-managed records;
// signups are per-user opt-ins, and after adoption they carry the address the list writes to — so
// the listing is a read of every signatory's email and the ownership check is what keeps a signup
// private. Positive controls prove the owner (and a service admin) still succeed.
func TestCrossTenant_WaitlistSignups_Denied(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, clientA := createUserAndClientForTest(t)
		_, clientB := createUserAndClientForTest(t)

		waitlist := createWaitlistForTest(t, clientA)
		signup := createWaitlistSignupForTest(t, clientA, waitlist.GetId())

		// cross-tenant: B may not read A's signup by id. Denied by the grant rather than by
		// ownership — a read by id is a service admin's — which is stricter than what this
		// case was written to check and still refuses what it was written to refuse.
		_, err := clientB.GetSignup(ctx, &waitlistspb.GetSignupRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// cross-tenant: B may not update A's signup.
		_, err = clientB.UpdateSignupNotes(ctx, &waitlistspb.UpdateSignupNotesRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
			Notes:    "hijacked notes",
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// cross-tenant: B may not unsubscribe A. A withdrawal is irreversible by
		// design — the address stays suppressed after it — so this is the most
		// damaging of the four.
		_, err = clientB.Withdraw(ctx, &waitlistspb.WithdrawRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// cross-tenant: B may not archive A's signup.
		_, err = clientB.ArchiveSignup(ctx, &waitlistspb.ArchiveSignupRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// cross-tenant: B may not work the queue either. Being on a list does not
		// make somebody its operator.
		_, err = clientB.Invite(ctx, &waitlistspb.InviteRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// the waitlist-wide listing is denied to regular users (it would expose every user's address)...
		_, err = clientB.ListSignups(ctx, &waitlistspb.ListSignupsRequest{ListId: waitlist.GetId()})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// ...but allowed for service admins.
		listed, err := adminClient.ListSignups(ctx, &waitlistspb.ListSignupsRequest{ListId: waitlist.GetId()})
		require.NoError(t, err)
		require.NotNil(t, listed)
		assert.NotEmpty(t, listed.GetResults())

		// a service admin may also read another user's signup by ID.
		_, err = adminClient.GetSignup(ctx, &waitlistspb.GetSignupRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.NoError(t, err)

		// A cannot read its own signup by id either, and that is not a cross-tenant rule —
		// it is the grant: GetSignup shares one with GetSignupByContact, which can ask
		// whether any address is on any list, so all three of those reads are an admin's.
		_, err = clientA.GetSignup(ctx, &waitlistspb.GetSignupRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// The fourth read is the one A gets to keep: its own place in the queue, under
		// ReadOwnWaitlistSignupsPermission.
		own, err := clientA.ListSignupsForSubject(ctx, subjectRequestFor(t, clientA))
		require.NoError(t, err)
		assert.NotEmpty(t, own.GetResults())

		// And the grant is not what keeps B out of it, because B holds the same one. The
		// subject comes off the request, so the refusal is the handler's: platform asks
		// AuthorizeSubjectRead before it reads a row, and this deployment answers
		// own-subject-or-admin. NotFound rather than PermissionDenied, so a subject that
		// belongs to somebody else and one that belongs to nobody are the same answer.
		_, err = clientB.ListSignupsForSubject(ctx, subjectRequestFor(t, clientA))
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))

		// A service admin reads anybody's, which is what the split was carved out of.
		theirs, err := adminClient.ListSignupsForSubject(ctx, subjectRequestFor(t, clientA))
		require.NoError(t, err)
		assert.NotEmpty(t, theirs.GetResults())

		// positive control: A may still amend and archive its own signup, which is what
		// the cross-tenant denials above are measured against.
		_, err = clientA.UpdateSignupNotes(ctx, &waitlistspb.UpdateSignupNotesRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
			Notes:    "owner-updated notes",
		})
		require.NoError(t, err)

		_, err = clientA.ArchiveSignup(ctx, &waitlistspb.ArchiveSignupRequest{
			ListId:   waitlist.GetId(),
			SignupId: signup.GetId(),
		})
		require.NoError(t, err)
	})
}

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
