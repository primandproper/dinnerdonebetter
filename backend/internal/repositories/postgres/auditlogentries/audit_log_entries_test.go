package auditlogentries

import (
	"context"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformaudit "github.com/primandproper/platform-go/v9/audit"
	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const resourceTypeForTest = "example"

// recordForTest appends one entry to an account's chain and hands back what was
// written, chain fields included — Record writes those back into the entry it was
// given, which is how a caller notarizes what it just recorded.
//
// It records inside a transaction rather than against the pool, because that is
// the only way a repository is allowed to record: Record holds the scope's chain
// row for the length of the caller's transaction, and against the pool that lock
// lapses before the INSERT it exists to protect.
func recordForTest(t *testing.T, ctx context.Context, dbc *repository, client database.Client, account *identity.Account, user *identity.User) *audit.AuditLogEntry {
	t.Helper()

	entry := &audit.AuditLogEntry{
		BelongsToAccount: &account.ID,
		BelongsToUser:    user.ID,
		ResourceType:     resourceTypeForTest,
		RelevantID:       identifiers.New(),
		EventType:        audit.AuditLogEventTypeUpdated,
	}

	require.NoError(t, client.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		return dbc.Record(ctx, tx, entry)
	}))

	require.NotEmpty(t, entry.ID, "Record assigns an ID")
	require.NotEmpty(t, entry.Hash, "Record assigns a hash")

	return entry
}

// accountForTest creates a user and an account they own, which is what an entry
// needs to have somewhere to be scoped to.
func accountForTest(t *testing.T, client database.Client) (*identity.User, *identity.Account) {
	t.Helper()

	user := pgtesting.CreateUserForTest(t, nil, client.Writer())

	exampleAccount := identityfakes.BuildFakeAccount()
	exampleAccount.BelongsToUser = user.ID

	return user, pgtesting.CreateAccountForTest(t, exampleAccount, user.ID, client.Writer())
}

// TestQuerier_Integration_AuditLogChain is the test that asks whether any of this
// is worth having. The chain either detects a removal and names where, or the
// tamper evidence is decoration.
func TestQuerier_Integration_AuditLogChain(t *testing.T) {
	ctx := t.Context()
	dbc, client := buildDatabaseClientForTest(t)

	user, account := accountForTest(t, client)

	var recorded []*audit.AuditLogEntry
	for range 3 {
		recorded = append(recorded, recordForTest(t, ctx, dbc, client, account, user))
	}

	t.Run("verifies clean", func(t *testing.T) {
		result, err := dbc.VerifyChain(ctx, account.ID, time.Time{}, time.Time{})
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.True(t, result.Intact())
		assert.Nil(t, result.FirstBreak)
		assert.Equal(t, len(recorded), result.Checked)
	})

	t.Run("refuses an UPDATE outright", func(t *testing.T) {
		// The append-only trigger. This is stronger than the chain: an edit is not
		// something to be discovered afterwards, it is something the database will
		// not do.
		_, err := client.Writer().ExecContext(ctx,
			"UPDATE "+audit.TablePrefix+"_audit_log_entries SET resource_type = $1 WHERE id = $2",
			"tampered", recorded[1].ID)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "append-only")
	})

	t.Run("names the break when an entry is removed", func(t *testing.T) {
		// DELETE is deliberately permitted — retention has to prune, and no trigger
		// can tell that sweep apart from an attacker. The chain is what covers it, so
		// removing the middle entry must be detectable and must be attributed to the
		// right position rather than merely somewhere.
		_, err := client.Writer().ExecContext(ctx,
			"DELETE FROM "+audit.TablePrefix+"_audit_log_entries WHERE id = $1", recorded[1].ID)
		require.NoError(t, err)

		result, err := dbc.VerifyChain(ctx, account.ID, time.Time{}, time.Time{})
		require.NoError(t, err)
		require.NotNil(t, result)

		require.False(t, result.Intact())
		require.NotNil(t, result.FirstBreak)
		assert.Equal(t, platformaudit.BreakMissingEntry, result.FirstBreak.Reason)
		assert.Equal(t, recorded[1].Seq, result.FirstBreak.Seq)
	})
}

// TestQuerier_Integration_AuditLogRedaction proves a declared Redaction keeps a
// value out of the table rather than out of the response. Filtering at read time
// would leave the secret durable, which is the whole distinction — so this asserts
// against the stored bytes, not against what the reader hands back.
func TestQuerier_Integration_AuditLogRedaction(t *testing.T) {
	ctx := t.Context()
	dbc, client := buildDatabaseClientForTest(t)

	user, account := accountForTest(t, client)

	entry := &audit.AuditLogEntry{
		BelongsToAccount: &account.ID,
		BelongsToUser:    user.ID,
		ResourceType:     "users",
		RelevantID:       user.ID,
		EventType:        audit.AuditLogEventTypeUpdated,
		Changes: map[string]audit.Change{
			"password":  {Old: "hunter2", New: "correct-horse-battery-staple"},
			"firstName": {Old: "before", New: "after"},
		},
	}

	require.NoError(t, client.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		return dbc.Record(ctx, tx, entry)
	}))

	// Record writes back what was actually stored, redaction included, so the value
	// a caller goes on to log cannot disagree with the value in the table.
	assert.NotContains(t, entry.Changes, "password")
	assert.Contains(t, entry.Changes, "firstName")

	fetched, err := dbc.GetAuditLogEntry(ctx, entry.ID)
	require.NoError(t, err)
	assert.NotContains(t, fetched.Changes, "password")
	require.Contains(t, fetched.Changes, "firstName")
	assert.Equal(t, "before", fetched.Changes["firstName"].Old)
	assert.Equal(t, "after", fetched.Changes["firstName"].New)

	var raw []byte
	require.NoError(t, client.Reader().QueryRowContext(ctx,
		"SELECT change_set FROM "+audit.TablePrefix+"_audit_log_entries WHERE id = $1", entry.ID).Scan(&raw))
	assert.NotContains(t, string(raw), "hunter2")
}

func TestQuerier_GetAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("with empty ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetAuditLogEntry(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}
