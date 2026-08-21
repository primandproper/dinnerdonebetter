package internalops

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/primandproper/platform-go/v12/database"
	"github.com/primandproper/platform-go/v12/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// gooseVersionTable is where the migrator records what it has applied. It is the one
// table in the public schema DestroyAllData deliberately leaves alone: truncating it
// would leave a fully migrated database claiming it had never been migrated.
const gooseVersionTable = "goose_db_version"

// TestQuerier_Integration_DestroyAllData pins the property the query is for — that it
// empties every table the schema has, not merely the ones somebody wrote a query builder
// for.
//
// The old statement was a literal list assembled at generation time from
// registerTableName calls in cmd/tools/codegen/queries, so a table with no generated
// queries was silently absent from it: eleven were, including sessions and
// webauthn_credentials, which is auth state surviving from one integration test into the
// next. The statement now reads pg_tables when it runs, and this test is what says so.
//
// It also covers the platform-owned tables — the webhook, audit, saga, metering, outbox,
// operations and authorization server ones — which never passed through this repository's
// generator at all and so were never in the list either.
func TestQuerier_Integration_DestroyAllData(t *testing.T) {
	ctx := t.Context()
	dbc := buildDatabaseClientForTest(t)

	// sessions is one of the tables the registry-derived list missed, and it takes rows
	// without needing a user to belong to. queue_test_messages was in that list, so
	// seeding both says the fix added tables rather than swapped one set for another.
	_, err := dbc.writeDB.ExecContext(ctx,
		"INSERT INTO sessions (token, data, expiry) VALUES ($1, $2, NOW() + INTERVAL '1 hour')",
		identifiers.New(), []byte("session data"),
	)
	require.NoError(t, err)
	require.NoError(t, dbc.CreateQueueTestMessage(ctx, identifiers.New(), "destroy-all-data-"+identifiers.New()[:8]))

	require.NotZero(t, countRows(ctx, t, dbc.readDB, "sessions"))
	require.NotZero(t, countRows(ctx, t, dbc.readDB, "queue_test_messages"))

	appliedMigrations := countRows(ctx, t, dbc.readDB, gooseVersionTable)
	require.NotZero(t, appliedMigrations)

	before := relfilenodes(ctx, t, dbc.readDB)
	require.NotEmpty(t, before)
	// The tables the registry missed, named so a regression that drops them again fails
	// here by name rather than as a missing key in the sweep below. Two of the eleven —
	// user_data_disclosures and webhook_trigger_events — are not here because 00029 and
	// 00026 have since dropped them for their platform equivalents.
	for _, table := range []string{
		"ingredient_media", "meal_images", "preparation_media", "recipe_images",
		"recipe_step_images", "sessions", "user_device_tokens", "webauthn_credentials",
		"webauthn_sessions",
	} {
		require.Contains(t, before, table)
	}

	require.NoError(t, dbc.generatedQuerier.DestroyAllData(ctx, dbc.writeDB))

	assert.Zero(t, countRows(ctx, t, dbc.readDB, "sessions"))
	assert.Zero(t, countRows(ctx, t, dbc.readDB, "queue_test_messages"))

	// Emptiness is not enough to prove coverage, because a table nobody seeded is empty
	// whether or not the statement named it. TRUNCATE rewrites a table's file, though, so
	// a table whose filenode did not move is one the statement skipped.
	after := relfilenodes(ctx, t, dbc.readDB)
	for table, filenode := range before {
		if table == gooseVersionTable {
			continue
		}

		assert.NotEqual(t, filenode, after[table], "%s was not truncated", table)
	}

	assert.Equal(t, before[gooseVersionTable], after[gooseVersionTable], "the migration record was truncated")
	assert.Equal(t, appliedMigrations, countRows(ctx, t, dbc.readDB, gooseVersionTable))
}

// relfilenodes maps every ordinary table in the public schema to the file backing it.
// Ordinary is relkind 'r': a partitioned table has no file of its own, and neither does a
// view or a sequence.
func relfilenodes(ctx context.Context, t *testing.T, db database.SQLQueryExecutor) map[string]int64 {
	t.Helper()

	rows, err := db.QueryContext(ctx, `SELECT c.relname, pg_relation_filenode(c.oid)
		FROM pg_class AS c
			JOIN pg_namespace AS n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' AND c.relkind = 'r'`)
	require.NoError(t, err)

	defer func() { assert.NoError(t, rows.Close()) }()

	filenodes := map[string]int64{}
	for rows.Next() {
		var (
			table    string
			filenode sql.NullInt64
		)

		require.NoError(t, rows.Scan(&table, &filenode))

		if filenode.Valid {
			filenodes[table] = filenode.Int64
		}
	}

	require.NoError(t, rows.Err())

	return filenodes
}

func countRows(ctx context.Context, t *testing.T, db database.SQLQueryExecutor, table string) int {
	t.Helper()

	var count int
	// The table names this is called with are constants in this file, not input.
	require.NoError(t, db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count))

	return count
}
