package integration

import (
	"testing"

	auditgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/audit"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"

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
		forUser, err := userClient.GetAuditLogEntriesForUser(ctx, &auditgrpc.GetAuditLogEntriesForUserRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(forUser.Results), 3, "expected at least 3 audit log entries (user, account, membership)")
	})

	T.Run("create user and fetch audit logs for that user returns entries with expected structure", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, userClient := createUserAndClientForTest(t)

		forUser, err := userClient.GetAuditLogEntriesForUser(ctx, &auditgrpc.GetAuditLogEntriesForUserRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)
		require.NotEmpty(t, forUser.Results)

		// Verify we have at least one entry with the user's ID
		foundUserEntry := false
		for _, entry := range forUser.Results {
			if entry.BelongsToUser == user.ID {
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
		forUser, err := userClient.GetAuditLogEntriesForUser(ctx, &auditgrpc.GetAuditLogEntriesForUserRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)
		require.NotEmpty(t, forUser.Results)

		// pick the first entry and get it by ID
		entryID := forUser.Results[0].Id

		result, err := userClient.GetAuditLogEntryByID(ctx, &auditgrpc.GetAuditLogEntryByIDRequest{
			AuditLogEntryId: entryID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, entryID, result.Result.Id)
	})

	T.Run("nonexistent entry", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		result, err := userClient.GetAuditLogEntryByID(ctx, &auditgrpc.GetAuditLogEntryByIDRequest{
			AuditLogEntryId: nonexistentID,
		})
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		result, err := c.GetAuditLogEntryByID(ctx, &auditgrpc.GetAuditLogEntryByIDRequest{
			AuditLogEntryId: nonexistentID,
		})
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestAuditLogEntries_Listing_ForAccount(T *testing.T) {
	T.Parallel()

	T.Run("create user and fetch audit logs for default account", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		accountsRes, err := userClient.GetAccounts(ctx, &identitysvc.GetAccountsRequest{})
		require.NoError(t, err)
		require.NotEmpty(t, accountsRes.Results, "user registration should create a default account")
		accountID := accountsRes.Results[0].Id

		// Account creation creates: account, account_user_membership - each generates an audit log entry
		forAccount, err := userClient.GetAuditLogEntriesForAccount(ctx, &auditgrpc.GetAuditLogEntriesForAccountRequest{
			AccountId: accountID,
		})
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(forAccount.Results), 2, "expected at least 2 audit log entries (account, membership)")
	})

	T.Run("create account and fetch audit logs for that account", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		lat, lng := float32(39.15643684457086), float32(-83.09156328830157)
		createRes, err := userClient.CreateAccount(ctx, &identitysvc.CreateAccountRequest{
			Input: &identitysvc.AccountCreationRequestInput{
				Name:      "integration test account",
				Latitude:  &lat,
				Longitude: &lng,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, createRes.Created)
		accountID := createRes.Created.Id

		forAccount, err := userClient.GetAuditLogEntriesForAccount(ctx, &auditgrpc.GetAuditLogEntriesForAccountRequest{
			AccountId: accountID,
		})
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(forAccount.Results), 2, "expected at least 2 audit log entries (account, membership)")

		// Verify entries belong to the correct account
		for _, entry := range forAccount.Results {
			assert.Equal(t, accountID, entry.BelongsToAccount, "audit log entry should belong to the created account")
		}
	})
}
