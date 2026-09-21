package migrations

import (
	"testing"

	ddboauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	oauth2database "github.com/primandproper/platform-go/v14/authentication/oauth2serverstore"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server/oauth2servertest"
	"github.com/primandproper/primitives-go/v2/clock"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuerier_Migrate_OAuth2ServerTables is the scenario the oauth2 table prefix
// exists for: two different things in this schema have wanted to be called
// oauth2_clients, and one of them has to carry a namespace or the other's DDL —
// CREATE TABLE IF NOT EXISTS, every time — is a silent no-op against a table with
// entirely different columns, and the authorization server fails on its first
// registration rather than at migration time.
//
// Both of them carry one now. The hand-written oauth2_clients this repository used
// to keep is gone: the administered client registry is platform's
// oauth2_registered_clients, adopted at migration 22, and the authorization server's
// own client table is platform's oauth2_clients. Two tables under one prefix rather
// than a prefixed one beside a bare one — which is the arrangement that still needs
// pinning, because a registry whose rows landed in the server's clients table would
// be a registration endpoint quietly minting credentials the server honors.
func TestQuerier_Migrate_OAuth2ServerTables(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		db, _ := pgtesting.BuildDatabaseContainerForTest(t)

		migrator, err := NewMigrator(loggingnoop.NewLogger())
		require.NoError(t, err)
		require.NoError(t, migrator.Migrate(ctx, db))

		// All four, together. A database holding three of them has an authorization
		// server that fails at whichever step the missing one serves.
		for _, table := range []string{
			"_oauth2_clients",
			"_oauth2_authorization_codes",
			"_oauth2_access_tokens",
			"_oauth2_refresh_tokens",
		} {
			var count int
			require.NoError(t, db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = $1`,
				ddboauth.TablePrefix+table).Scan(&count))
			assert.Equal(t, 1, count, "missing %s", ddboauth.TablePrefix+table)
		}

		// And the registry is its own table beside them. It is the fifth, not one of
		// the four: a listing endpoint, permissions and an audit trail sit behind it,
		// and the server's clients table has none of that.
		var registry int
		require.NoError(t, db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = $1`,
			ddboauth.TablePrefix+"_oauth2_registered_clients").Scan(&registry))
		assert.Equal(t, 1, registry, "the client registry table must exist")

		// belongs_to_user is the column that tells them apart: a registered client has
		// an owner, and the server's clients table has no such notion. Finding it on
		// one and not the other proves the two DDLs did not land on one table.
		for table, expected := range map[string]int{
			ddboauth.TablePrefix + "_oauth2_registered_clients": 1,
			ddboauth.TablePrefix + "_oauth2_clients":            0,
		} {
			var owned int
			require.NoError(t, db.QueryRowContext(ctx,
				`SELECT COUNT(*) FROM information_schema.columns
				 WHERE table_name = $1 AND column_name = 'belongs_to_user'`, table).Scan(&owned))
			assert.Equal(t, expected, owned, "belongs_to_user on %s", table)
		}
	})
}

// TestOAuth2Store_Conformance runs the platform's Store conformance suite against
// the tables this repository's migrator actually creates.
//
// The store and the DDL are both the platform's, so this is not testing platform
// code — it is testing that the two meet where we join them. The table prefix is
// declared in three places that have to agree (the migration, the rendered config,
// and the store), and a prefix that differs between the writer and the reader is
// the misconfiguration that stays invisible until somebody tries to sign in. So is
// a migration registered at a version that never runs.
//
// It also covers the two cases a hand-written test would not think to write, and
// which are the whole reason the durable store exists: a code redeemed twice
// concurrently resolving to exactly one winner, and a record that expires between
// a read and the write that follows it.
func TestOAuth2Store_Conformance(T *testing.T) {
	T.Parallel()

	ctx := T.Context()
	db, dbConfig := pgtesting.BuildDatabaseContainerForTest(T)

	migrator, err := NewMigrator(loggingnoop.NewLogger())
	require.NoError(T, err)
	require.NoError(T, migrator.Migrate(ctx, db))

	client, err := postgres.NewDatabaseClient(ctx, dbConfig,
		postgres.WithLogger(loggingnoop.NewLogger()),
		postgres.WithTracerProvider(tracingnoop.NewTracerProvider()),
	)
	require.NoError(T, err)

	// A store per call rather than one for the suite, because the factory is handed a clock
	// now: the sweep cases build a store on a clock they can advance, and the rest get the
	// wall clock. The database is still shared — the suite gives each record it writes a
	// unique identifier precisely so one database can serve every subtest in parallel.
	//
	// No WithInstanceLocalState here — that deviation is the memory store's, and claiming it
	// would skip the cases that prove this one is shareable across replicas, which is the
	// entire reason we are on it.
	oauth2servertest.Run(T, func(tb testing.TB, c clock.Clock) oauth2server.Store {
		tb.Helper()

		store, storeErr := oauth2database.NewStore(
			&oauth2database.Config{TablePrefix: ddboauth.TablePrefix},
			client,
			oauth2database.WithClock(c),
			oauth2database.WithLogger(loggingnoop.NewLogger()),
			oauth2database.WithTracerProvider(tracingnoop.NewTracerProvider()),
		)
		require.NoError(tb, storeErr)

		return store
	})
}
