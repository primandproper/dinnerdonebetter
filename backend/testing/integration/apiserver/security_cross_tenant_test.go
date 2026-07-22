package integration

import (
	"testing"

	auditsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/audit"
	authsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	identitysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	mealplanninggrpc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	webhookssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"

	mpconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mpfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mpgrpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

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

// TestCrossTenant_AuditLogForAccount_Denied asserts that a user cannot read the account-scoped audit
// log of an account they are not a member of. audit/grpc/service.go GetAuditLogEntriesForAccount
// returns codes.PermissionDenied for non-members.
func TestCrossTenant_AuditLogForAccount_Denied(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, clientA := createUserAndClientForTest(t)
		_, clientB := createUserAndClientForTest(t)

		activeA, err := clientA.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		accountAID := getActiveAccountIDForClientForTest(t, activeA)

		// positive control: A can read its own account's audit log.
		ownResp, err := clientA.GetAuditLogEntriesForAccount(ctx, &auditsvc.GetAuditLogEntriesForAccountRequest{
			AccountId: accountAID,
		})
		require.NoError(t, err)
		require.NotNil(t, ownResp)

		// cross-tenant: B is not a member of A's account.
		_, err = clientB.GetAuditLogEntriesForAccount(ctx, &auditsvc.GetAuditLogEntriesForAccountRequest{
			AccountId: accountAID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})
}

// TestCrossTenant_AuditLogForUser_Denied asserts that a user cannot read another user's user-scoped
// audit log. audit/grpc/service.go GetAuditLogEntriesForUser returns codes.PermissionDenied unless
// the requester is the target user (or a service admin).
func TestCrossTenant_AuditLogForUser_Denied(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		userA, clientA := createUserAndClientForTest(t)
		_, clientB := createUserAndClientForTest(t)

		// positive control: A can read its own user audit log.
		ownResp, err := clientA.GetAuditLogEntriesForUser(ctx, &auditsvc.GetAuditLogEntriesForUserRequest{
			UserId: userA.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, ownResp)

		// cross-tenant: B may not read A's user audit log.
		_, err = clientB.GetAuditLogEntriesForUser(ctx, &auditsvc.GetAuditLogEntriesForUserRequest{
			UserId: userA.ID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})
}

// TestCrossTenant_AuditLogEntryByID_Denied asserts that a user cannot read an individual audit log
// entry belonging to another user/account by its ID. audit/grpc/service.go GetAuditLogEntryByID
// returns codes.PermissionDenied when the entry belongs to neither the requester nor their active
// account.
func TestCrossTenant_AuditLogEntryByID_Denied(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		userA, clientA := createUserAndClientForTest(t)
		_, clientB := createUserAndClientForTest(t)

		// A performs an auditable op so that a fresh, A-owned entry exists.
		createWebhookForTest(t, clientA)

		// A lists its OWN entries (allowed — self) and picks one to target.
		forUser, err := clientA.GetAuditLogEntriesForUser(ctx, &auditsvc.GetAuditLogEntriesForUserRequest{
			UserId: userA.ID,
		})
		require.NoError(t, err)
		require.NotEmpty(t, forUser.Results)
		entryID := forUser.Results[0].Id
		require.NotEmpty(t, entryID)

		// positive control: A can fetch its own entry by ID.
		ownEntry, err := clientA.GetAuditLogEntryByID(ctx, &auditsvc.GetAuditLogEntryByIDRequest{
			AuditLogEntryId: entryID,
		})
		require.NoError(t, err)
		require.NotNil(t, ownEntry)
		assert.Equal(t, entryID, ownEntry.Result.Id)

		// cross-tenant: B may not fetch A's entry by ID.
		_, err = clientB.GetAuditLogEntryByID(ctx, &auditsvc.GetAuditLogEntryByIDRequest{
			AuditLogEntryId: entryID,
		})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
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

		// A creates a webhook and adds a trigger config to it.
		webhookA := createWebhookForTest(t, clientA)
		catalogEvent := createWebhookTriggerEventCatalogForTest(t, ctx, clientA, "cross_tenant_archived", "for cross-tenant test")

		addedConfig, err := clientA.AddWebhookTriggerConfig(ctx, &webhookssvc.AddWebhookTriggerConfigRequest{
			WebhookId: webhookA.ID,
			Input: &webhookssvc.WebhookTriggerConfigCreationRequestInput{
				BelongsToWebhook: webhookA.ID,
				TriggerEventId:   catalogEvent.Id,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, addedConfig.Created)
		configID := addedConfig.Created.Id
		require.NotEmpty(t, configID)

		// cross-tenant: B attempts to archive A's trigger config. Ownership is enforced via the
		// account-scoped GetWebhook lookup, which cannot see A's webhook from B's session, so the
		// handler surfaces codes.Internal rather than PermissionDenied.
		_, err = clientB.ArchiveWebhookTriggerConfig(ctx, &webhookssvc.ArchiveWebhookTriggerConfigRequest{
			WebhookId:              webhookA.ID,
			WebhookTriggerConfigId: configID,
		})
		require.Error(t, err)

		// security property: A's trigger config must still exist after B's attempt.
		refreshed, err := clientA.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{WebhookId: webhookA.ID})
		require.NoError(t, err)
		require.NotNil(t, refreshed.Result)

		var stillPresent bool
		for _, cfg := range refreshed.Result.TriggerConfigs {
			if cfg.Id == configID {
				stillPresent = true
				break
			}
		}
		assert.True(t, stillPresent, "expected A's webhook trigger config %q to survive B's cross-tenant archive attempt", configID)

		// positive control: A can archive its own trigger config.
		_, err = clientA.ArchiveWebhookTriggerConfig(ctx, &webhookssvc.ArchiveWebhookTriggerConfigRequest{
			WebhookId:              webhookA.ID,
			WebhookTriggerConfigId: configID,
		})
		require.NoError(t, err)
	})
}
