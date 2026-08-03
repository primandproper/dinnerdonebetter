package auditlogentries

import (
	"context"
	"testing"
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformaudit "github.com/primandproper/platform-go/v9/audit"
	"github.com/primandproper/platform-go/v9/database"
	mockdatabase "github.com/primandproper/platform-go/v9/database/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	resourceTypeForTest = "example"
)

// recordForTest appends one entry to an account's chain and hands back what was
// written, chain fields included — Record writes those back into the entry it was
// given, which is how a caller notarizes what it just recorded.
func recordForTest(t *testing.T, ctx context.Context, dbc *repository, account *identity.Account, user *identity.User, eventType types.EventType) *types.Entry {
	t.Helper()

	entry := &types.Entry{
		Scope:        account.ID,
		ResourceType: resourceTypeForTest,
		ResourceID:   identityfakes.BuildFakeID(),
		EventType:    eventType,
		Actor:        types.UserActor(user.ID),
	}

	require.NoError(t, dbc.Record(ctx, dbc.Writer(), entry))
	require.NotEmpty(t, entry.ID, "Record assigns an ID")
	require.NotEmpty(t, entry.Hash, "Record assigns a hash")

	return entry
}

func TestQuerier_Integration_AuditLogEntries(t *testing.T) {
	ctx := t.Context()
	dbc := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.Writer())
	exampleAccount := identityfakes.BuildFakeAccount()
	exampleAccount.BelongsToUser = user.ID
	account := pgtesting.CreateAccountForTest(t, exampleAccount, user.ID, dbc.Writer())

	recorded := []*types.Entry{}
	for range exampleQuantity {
		recorded = append(recorded, recordForTest(t, ctx, dbc, account, user, types.EventCreated))
	}

	// Positions are assigned in order and are contiguous, which is what makes a
	// removed row a hole rather than an absence nobody can see.
	for i, entry := range recorded {
		assert.Equal(t, int64(i), entry.Seq)
	}

	t.Run("get by ID", func(t *testing.T) {
		fetched, err := dbc.GetAuditLogEntry(ctx, recorded[0].ID)
		require.NoError(t, err)
		require.NotNil(t, fetched)

		assert.Equal(t, recorded[0].ID, fetched.ID)
		assert.Equal(t, resourceTypeForTest, fetched.ResourceType)
		assert.Equal(t, recorded[0].ResourceID, fetched.RelevantID)
		assert.Equal(t, string(types.EventCreated), fetched.EventType)
		assert.Equal(t, user.ID, fetched.BelongsToUser)
		require.NotNil(t, fetched.BelongsToAccount)
		assert.Equal(t, account.ID, *fetched.BelongsToAccount)
	})

	t.Run("list for user", func(t *testing.T) {
		entries, err := dbc.GetAuditLogEntriesForUser(ctx, user.ID, nil)
		require.NoError(t, err)
		assert.Len(t, entries.Data, len(recorded))
	})

	t.Run("list for account", func(t *testing.T) {
		entries, err := dbc.GetAuditLogEntriesForAccount(ctx, account.ID, nil)
		require.NoError(t, err)
		assert.Len(t, entries.Data, len(recorded))
	})

	t.Run("list for user and resource types", func(t *testing.T) {
		entries, err := dbc.GetAuditLogEntriesForUserAndResourceTypes(ctx, user.ID, []string{resourceTypeForTest}, nil)
		require.NoError(t, err)
		assert.Len(t, entries.Data, len(recorded))

		entries, err = dbc.GetAuditLogEntriesForUserAndResourceTypes(ctx, user.ID, []string{"something_else"}, nil)
		require.NoError(t, err)
		assert.Empty(t, entries.Data)
	})

	t.Run("list for account and resource types", func(t *testing.T) {
		entries, err := dbc.GetAuditLogEntriesForAccountAndResourceTypes(ctx, account.ID, []string{resourceTypeForTest}, nil)
		require.NoError(t, err)
		assert.Len(t, entries.Data, len(recorded))
	})
}

// TestQuerier_Integration_AuditLogChain is the test the migration exists for. The
// hand-rolled log this replaced could not have passed it: nothing chained its
// entries, so an edit or a removal left no trace at all.
func TestQuerier_Integration_AuditLogChain(t *testing.T) {
	ctx := t.Context()
	dbc := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.Writer())
	exampleAccount := identityfakes.BuildFakeAccount()
	exampleAccount.BelongsToUser = user.ID
	account := pgtesting.CreateAccountForTest(t, exampleAccount, user.ID, dbc.Writer())

	recorded := []*types.Entry{}
	for range 3 {
		recorded = append(recorded, recordForTest(t, ctx, dbc, account, user, types.EventUpdated))
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
		// The append-only trigger, installed by migration 25. This is stronger
		// than the chain: an edit is not something to be discovered afterwards,
		// it is something the database will not do.
		_, err := dbc.Writer().ExecContext(ctx,
			"UPDATE ddb_audit_log_entries SET resource_type = $1 WHERE id = $2",
			"tampered", recorded[1].ID,
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "append-only")
	})

	t.Run("names the break when an entry is removed", func(t *testing.T) {
		// DELETE is deliberately permitted — retention has to prune, and no
		// trigger can tell that sweep apart from an attacker. The chain is what
		// covers it, so removing the middle entry must be detectable and must be
		// attributed to the right position.
		_, err := dbc.Writer().ExecContext(ctx,
			"DELETE FROM ddb_audit_log_entries WHERE id = $1", recorded[1].ID)
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
// would leave the secret durable, which is the whole distinction.
func TestQuerier_Integration_AuditLogRedaction(t *testing.T) {
	ctx := t.Context()
	dbc := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.Writer())
	exampleAccount := identityfakes.BuildFakeAccount()
	exampleAccount.BelongsToUser = user.ID
	account := pgtesting.CreateAccountForTest(t, exampleAccount, user.ID, dbc.Writer())

	entry := &types.Entry{
		Scope:        account.ID,
		ResourceType: "users",
		ResourceID:   user.ID,
		EventType:    types.EventUpdated,
		Actor:        types.UserActor(user.ID),
		Changes: map[string]types.Change{
			"password":  {Old: "hunter2", New: "correct-horse-battery-staple"},
			"firstName": {Old: "before", New: "after"},
		},
	}

	require.NoError(t, dbc.Record(ctx, dbc.Writer(), entry))

	// Record writes back what was actually stored, redaction included, so the value
	// a caller goes on to log cannot disagree with the value in the table.
	assert.NotContains(t, entry.Changes, "password")
	assert.Contains(t, entry.Changes, "firstName")

	fetched, err := dbc.GetAuditLogEntry(ctx, entry.ID)
	require.NoError(t, err)
	assert.NotContains(t, fetched.Changes, "password")
	require.Contains(t, fetched.Changes, "firstName")
	assert.Equal(t, "before", fetched.Changes["firstName"].OldValue)
	assert.Equal(t, "after", fetched.Changes["firstName"].NewValue)

	var raw []byte
	require.NoError(t, dbc.Reader().QueryRowContext(ctx,
		"SELECT change_set FROM ddb_audit_log_entries WHERE id = $1", entry.ID).Scan(&raw))
	assert.NotContains(t, string(raw), "hunter2")
}

func TestQuerier_GetAuditLogEntry(T *testing.T) {
	T.Parallel()

	T.Run("with invalid audit log entry ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetAuditLogEntry(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_Record(T *testing.T) {
	T.Parallel()

	T.Run("with no entries", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		// Nothing to record is not an error, so a caller assembling entries
		// conditionally does not have to guard the call.
		assert.NoError(t, c.Record(ctx, nil))
	})

	T.Run("with an entry missing an actor", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.Record(ctx, noopExecutor(), &types.Entry{
			ResourceType: resourceTypeForTest,
			EventType:    types.EventCreated,
		})
		assert.Error(t, err)
	})
}

// noopExecutor is a query executor that is never reached: validation rejects the
// entry before any statement is built.
func noopExecutor() database.SQLQueryExecutor {
	return &mockdatabase.SQLQueryExecutorMock{}
}
