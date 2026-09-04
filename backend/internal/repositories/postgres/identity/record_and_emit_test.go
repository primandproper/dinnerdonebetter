package identity

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v13/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fetchDataChangeEventTypes reads the event type off every data change message enqueued so far,
// oldest first. It reads the outbox directly rather than through a relay because the point of
// these tests is what the repository's transaction wrote, not what a broker later carried.
func fetchDataChangeEventTypes(ctx context.Context, t *testing.T, db database.SQLQueryExecutor) []string {
	t.Helper()

	rows, err := db.QueryContext(ctx, `SELECT payload FROM outbox_messages WHERE topic = $1 ORDER BY created_at, id`, testDataChangesTopic)
	require.NoError(t, err)
	defer func() { assert.NoError(t, rows.Close()) }()

	var eventTypes []string
	for rows.Next() {
		var payload []byte
		require.NoError(t, rows.Scan(&payload))

		var msg audit.DataChangeMessage
		require.NoError(t, json.Unmarshal(payload, &msg))

		eventTypes = append(eventTypes, msg.EventType)
	}
	require.NoError(t, rows.Err())

	return eventTypes
}

// TestQuerier_Integration_WritesRecordAndEmitTogether pins what RecordAndEmit exists to
// guarantee: a write that owes both an audit log entry and a data change event produces both,
// as further statements in the transaction that made the change.
//
// Each half used to be its own block at every write, which made the easiest mistake to make the
// one nothing catches. A missing entry leaves a row with no provenance and the chain does not
// notice, because a chain records what it was given; a missing event leaves the search index
// stale and no webhook fired. This asserts both halves for every user write that owes them.
func TestQuerier_Integration_WritesRecordAndEmitTogether(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := createUserForTest(t, ctx, nil, dbc)

	require.NoError(t, dbc.UpdateUserUsername(ctx, user.ID, "new_"+user.Username))
	require.NoError(t, dbc.UpdateUserEmailAddress(ctx, user.ID, "new_"+user.EmailAddress))
	require.NoError(t, dbc.UpdateUserDetails(ctx, user.ID, &identity.UserDetailsDatabaseUpdateInput{
		FirstName: "First",
		LastName:  "Last",
	}))
	require.NoError(t, dbc.ArchiveUser(ctx, user.ID))

	// The audit half: every one of those writes is answerable for.
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeUsers, RelevantID: user.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeUsers, RelevantID: user.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeUsers, RelevantID: user.ID},
	})

	// The event half: each write announced the specific thing it changed. Asserting the
	// distinct types rather than a count is what would catch one write emitting another's
	// event, which a count cannot see.
	eventTypes := fetchDataChangeEventTypes(ctx, t, dbc.writeDB)
	assert.Contains(t, eventTypes, identity.UsernameChangedEventType)
	assert.Contains(t, eventTypes, identity.EmailAddressChangedEventType)
	assert.Contains(t, eventTypes, identity.UserDetailsChangedEventType)
	assert.Contains(t, eventTypes, identity.UserArchivedServiceEventType)
}

// TestQuerier_Integration_RecordAndEmitRollsBackWithItsWrite pins the other direction: neither
// half outlives a write that did not happen.
//
// Archiving a user that does not exist affects no rows, so the transaction rolls back and both
// statements roll back with it. An audit entry for a change the database never made is worse
// than no entry at all, because the chain would carry it as fact.
func TestQuerier_Integration_RecordAndEmitRollsBackWithItsWrite(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	require.Error(t, dbc.ArchiveUser(ctx, "nonexistent"))

	entries, err := auditRepo.GetAuditLogEntriesForUser(ctx, "nonexistent", nil)
	require.NoError(t, err)
	assert.Empty(t, entries.Data)

	assert.Empty(t, fetchDataChangeEventTypes(ctx, t, dbc.writeDB))
}
