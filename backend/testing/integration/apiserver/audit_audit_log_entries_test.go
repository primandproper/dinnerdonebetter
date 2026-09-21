package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"
	"github.com/primandproper/platform-go/v14/identity/identitypb"

	auditgrpc "github.com/primandproper/platform-go/v14/audit/auditpb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuditLogEntries_Listing_ForUser(T *testing.T) {
	T.Parallel()

	T.Run("create user and fetch audit logs for that user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, userClient := createUserAndClientForTest(t)

		// User creation creates: user, account, account_user_membership - each generates an audit log entry
		forUser, err := userClient.ListEntries(ctx, &auditgrpc.ListEntriesRequest{
			Query: &auditgrpc.EntryQuery{ActorId: user.ID},
		})
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(forUser.GetResults()), 3, "expected at least 3 audit log entries (user, account, membership)")
	})

	T.Run("create user and fetch audit logs for that user returns entries with expected structure", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, userClient := createUserAndClientForTest(t)

		forUser, err := userClient.ListEntries(ctx, &auditgrpc.ListEntriesRequest{
			Query: &auditgrpc.EntryQuery{ActorId: user.ID},
		})
		require.NoError(t, err)
		require.NotEmpty(t, forUser.GetResults())

		// Verify we have at least one entry with the user's ID
		foundUserEntry := false
		for _, entry := range forUser.GetResults() {
			if entry.GetActor().GetId() == user.ID {
				foundUserEntry = true
				break
			}
		}
		assert.True(t, foundUserEntry, "expected at least one audit log entry belonging to the created user")
	})
}

func TestAuditLogEntries_GetByID(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, userClient := createUserAndClientForTest(t)

		// User creation generates audit log entries; fetch them for the user
		forUser, err := userClient.ListEntries(ctx, &auditgrpc.ListEntriesRequest{
			Query: &auditgrpc.EntryQuery{ActorId: user.ID},
		})
		require.NoError(t, err)
		require.NotEmpty(t, forUser.GetResults())

		// pick the first entry and get it by ID
		entryID := forUser.GetResults()[0].Id

		result, err := userClient.GetEntry(ctx, &auditgrpc.GetEntryRequest{EntryId: entryID})
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, entryID, result.GetEntry().GetId())
	})

	T.Run("nonexistent entry", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		result, err := userClient.GetEntry(ctx, &auditgrpc.GetEntryRequest{EntryId: nonexistentID})
		require.Error(t, err)
		assert.Nil(t, result)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		result, err := c.GetEntry(ctx, &auditgrpc.GetEntryRequest{EntryId: nonexistentID})
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestAuditLogEntries_Listing_ForAccount(T *testing.T) {
	T.Parallel()

	T.Run("create user and fetch audit logs for default account", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, userClient := createUserAndClientForTest(t)

		accountsRes, err := userClient.IdentityService().ListAccountsForUser(ctx, &identitypb.ListAccountsForUserRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)
		require.NotEmpty(t, accountsRes.GetResults(), "user registration should create a default account")

		// The default account is the session's active one, and the active one is the scope
		// this read resolves — so the account is not named here.
		//
		// Account creation creates: account, account_user_membership - each generates an
		// audit log entry.
		forAccount, err := userClient.ListEntries(ctx, &auditgrpc.ListEntriesRequest{})
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(forAccount.GetResults()), 2, "expected at least 2 audit log entries (account, membership)")
	})

	T.Run("create account and fetch audit logs for that account", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, userClient := createUserAndClientForTest(t)

		createRes, err := userClient.IdentityService().CreateAccount(ctx, &identitypb.CreateAccountRequest{
			Name:       "integration test account",
			OwnerRoles: []string{authorization.AccountAdminRoleName},
		})
		require.NoError(t, err)
		require.NotNil(t, createRes.GetAccount())
		accountID := createRes.GetAccount().GetId()

		// Read as somebody whose active account is the new one, which takes
		// impersonation: a chain is read by being in it, and the account a session is
		// in is not a request field. The RPC this replaced took an account id and
		// checked membership afterwards; that check is the scope now.
		impersonated := client.ImpersonateUseAndAccountContext(ctx, user.ID, accountID)

		forAccount, err := adminClient.ListEntries(impersonated, &auditgrpc.ListEntriesRequest{})
		require.NoError(t, err)

		// Account creation creates: account, account_user_membership.
		assert.GreaterOrEqual(t, len(forAccount.GetResults()), 2, "expected at least 2 audit log entries (account, membership)")

		// There is no per-entry account to check, and its absence is the guarantee
		// rather than a loss: every entry in this response is already the impersonated
		// account's, because that is the chain the read was bound to. An entry naming
		// its own account would be one a reader had to re-check.
	})
}
