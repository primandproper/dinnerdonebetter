package identity

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"

	"github.com/primandproper/platform-go/v11/database"
	searchsync "github.com/primandproper/platform-go/v11/search/sync"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fetchIndexEvents reads every index event enqueued for a topic, oldest first, along with the
// partition key the row was written under.
func fetchIndexEvents(ctx context.Context, t *testing.T, db database.SQLQueryExecutor, topic string) (events []searchsync.Event, keys []string) {
	t.Helper()

	rows, err := db.QueryContext(ctx, `SELECT partition_key, payload FROM outbox_messages WHERE topic = $1 ORDER BY created_at, id`, topic)
	require.NoError(t, err)
	defer func() { assert.NoError(t, rows.Close()) }()

	for rows.Next() {
		var (
			key     string
			payload []byte
		)
		require.NoError(t, rows.Scan(&key, &payload))

		var event searchsync.Event
		require.NoError(t, json.Unmarshal(payload, &event))

		events = append(events, event)
		keys = append(keys, key)
	}
	require.NoError(t, rows.Err())

	return events, keys
}

func TestQuerier_Integration_UserIndexEventsCommitWithTheirWrite(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := createUserForTest(t, ctx, nil, dbc)

	// Signup is the only write that puts a user in the index for the first time, so it is
	// also the only one whose absence would leave a user unsearchable until the next reindex.
	events, keys := fetchIndexEvents(ctx, t, dbc.writeDB, identityindexing.IndexTypeUsers)
	require.Len(t, events, 1)
	assert.Equal(t, user.ID, events[0].DocumentID)
	assert.Equal(t, searchsync.OpUpsert, events[0].Op)
	// The document ID is the partition key: two edits to one user can never be applied out
	// of order, however many relays are running.
	assert.Equal(t, user.ID, keys[0])

	// Every field the index holds — username, email address, and the name details — is
	// changed by a write of its own, and each has to say so.
	require.NoError(t, dbc.UpdateUserUsername(ctx, user.ID, "new_"+user.Username))
	require.NoError(t, dbc.UpdateUserEmailAddress(ctx, user.ID, "new_"+user.EmailAddress))
	require.NoError(t, dbc.UpdateUserDetails(ctx, user.ID, &identity.UserDetailsDatabaseUpdateInput{
		FirstName: "First",
		LastName:  "Last",
	}))

	events, _ = fetchIndexEvents(ctx, t, dbc.writeDB, identityindexing.IndexTypeUsers)
	require.Len(t, events, 4)
	for _, event := range events {
		assert.Equal(t, user.ID, event.DocumentID)
		assert.Equal(t, searchsync.OpUpsert, event.Op)
	}

	// An archived user is one search must stop returning, so archival is a delete rather
	// than an upsert of a row that is no longer there.
	require.NoError(t, dbc.ArchiveUser(ctx, user.ID))

	events, _ = fetchIndexEvents(ctx, t, dbc.writeDB, identityindexing.IndexTypeUsers)
	require.Len(t, events, 5)
	assert.Equal(t, user.ID, events[4].DocumentID)
	assert.Equal(t, searchsync.OpDelete, events[4].Op)
}

func TestQuerier_Integration_UserIndexEventRollsBackWithItsWrite(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	// Archiving a user that does not exist affects no rows, so the transaction rolls back and
	// the index event rolls back with it. Published downstream of the commit instead, this is
	// the case that could announce a change the database never made.
	require.Error(t, dbc.ArchiveUser(ctx, "nonexistent"))

	events, _ := fetchIndexEvents(ctx, t, dbc.writeDB, identityindexing.IndexTypeUsers)
	assert.Empty(t, events)
}
