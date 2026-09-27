package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	auditgrpc "github.com/primandproper/platform-go/v14/audit/auditpb"
	"github.com/primandproper/platform-go/v14/identity/identitypb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The audit surface's behavior — reads by id, confinement to the caller's chain, verification —
// is asserted by platform's audit conformance suite, run against this deployment in
// conformance_test.go. What remains here is this application's own: that registering a user and
// creating an account record the entries this application's identity hooks write.
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
