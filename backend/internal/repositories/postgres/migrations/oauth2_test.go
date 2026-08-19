package migrations

import (
	"testing"

	ddboauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v11/authentication/oauth2server"
	oauth2database "github.com/primandproper/platform-go/v11/authentication/oauth2server/database"
	"github.com/primandproper/platform-go/v11/authentication/oauth2server/oauth2servertest"
	"github.com/primandproper/platform-go/v11/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v11/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v11/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuerier_Migrate_OAuth2ServerTables is the scenario the oauth2 table prefix
// exists for, and unlike the audit one it is not a database this repository merely
// might encounter: 00004_oauth.sql creates a table called oauth2_clients in every
// database we have, and the platform's schema names its first table the same thing.
//
// The platform's DDL says CREATE TABLE IF NOT EXISTS, so without the prefix this
// migration would be a silent no-op against a table with entirely different columns,
// and the authorization server would fail on its first registration rather than at
// migration time. With the prefix there is nothing to collide with, and the
// go-oauth2 server keeps running on its own tables until #1288 retires it.
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

		// And the go-oauth2 server's table is still its own. `client_secret` is a
		// column only the legacy shape has — the platform's clients table stores a
		// `secret_hash` — so finding it here proves the two did not merge.
		var legacySecret int
		require.NoError(t, db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM information_schema.columns
			 WHERE table_name = 'oauth2_clients' AND column_name = 'client_secret'`).Scan(&legacySecret))
		assert.Equal(t, 1, legacySecret, "the go-oauth2 clients table must be untouched")

		var platformSecret int
		require.NoError(t, db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM information_schema.columns
			 WHERE table_name = $1 AND column_name = 'secret_hash'`,
			ddboauth.TablePrefix+"_oauth2_clients").Scan(&platformSecret))
		assert.Equal(t, 1, platformSecret, "the platform's clients table must be the platform's shape")
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

	store, err := oauth2database.NewStore(
		&oauth2database.Config{TablePrefix: ddboauth.TablePrefix},
		client,
		oauth2database.WithLogger(loggingnoop.NewLogger()),
		oauth2database.WithTracerProvider(tracingnoop.NewTracerProvider()),
	)
	require.NoError(T, err)

	// One store for every subtest: the suite gives each record it writes a unique
	// identifier precisely so a single database can serve all of them in parallel.
	// No WithInstanceLocalState here — that deviation is the memory store's, and
	// claiming it would skip the cases that prove this one is shareable across
	// replicas, which is the entire reason we are on it.
	oauth2servertest.Run(T, func(testing.TB) oauth2server.Store { return store })
}
