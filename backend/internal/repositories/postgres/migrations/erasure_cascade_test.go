package migrations

import (
	"context"
	"database/sql"
	"testing"

	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuerier_Migrate_ErasingAUserTakesTheirCredentials pins the keys that make this
// application's single identity eraser cover three adopted tables.
//
// Erasure here is one delete: the user row goes, and every table naming that user goes with
// it through a foreign key. That is a design decision rather than an accident — see
// internal/build/dataprivacy, which registers a separate eraser only for the two domains
// where a key is impossible — and it means a table with no key is a table erasure silently
// does not reach. Nothing raises. The delete succeeds, the request reports Deleted, and the
// rows are still there.
//
// All three of these had no key until this test was written. platform stores a user id as a
// bare column and has to: the module is multi-engine and does not know what a consumer calls
// its user table. Re-creating the key is the consumer's job, which uploads/registry,
// issuereports and notifications each did at adoption and these three did not.
//
// webauthn_credentials is the sharpest case, because it is a regression rather than an
// omission: the hand-written table carried REFERENCES users(id) ON DELETE CASCADE, and
// adopting platform's passkeys schema dropped the key with the table. A passkey is a
// credential that still authenticates somebody the service has erased.
//
// The registry of OAuth2 clients is the third table that names a user and is deliberately
// not here: most of its rows name nobody, so it can have no key, and a registered eraser
// covers the ones that do. See renderOAuth2ClientsDDL and EraserKeyOAuth2Clients.
func TestQuerier_Migrate_ErasingAUserTakesTheirCredentials(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		db, _ := pgtesting.BuildDatabaseContainerForTest(t)

		migrator, err := NewMigrator(loggingnoop.NewLogger())
		require.NoError(t, err)
		require.NoError(t, migrator.Migrate(ctx, db))

		const userID = "erasure_subject"

		_, err = db.ExecContext(ctx,
			"INSERT INTO ddb_identity_users (id, scope, username, display_name, email_address, hashed_password, two_factor_secret, account_status) "+
				"VALUES ($1, $2, $3, $3, $4, '', '', 'good')",
			userID, tenancy.Global().Owner(), "erasable", "erasable@example.com")
		require.NoError(t, err)

		// One row per table, in the shape each store writes. The scope is the owner
		// spelling rather than the prose one — see insertWebAuthnCredentialForTest.
		scope := tenancy.Global().Owner()

		writes := []struct {
			table     string
			statement string
			args      []any
		}{
			{
				table: "ddb_webauthn_credentials",
				statement: `INSERT INTO ddb_webauthn_credentials
					(id, scope, belongs_to_user, credential_id, public_key, sign_count, transports, friendly_name)
					VALUES ($1, $2, $3, $4, $5, 0, '[]', 'a passkey')`,
				args: []any{"cred_1", scope, userID, []byte("credential-id"), []byte("public-key")},
			},
			{
				table: "ddb_password_reset_tokens",
				statement: `INSERT INTO ddb_password_reset_tokens
					(id, scope, belongs_to_user, token_digest, expires_at, created_at)
					VALUES ($1, $2, $3, $4, NOW() + INTERVAL '1 hour', NOW())`,
				args: []any{"reset_1", scope, userID, "digest"},
			},
		}

		for _, write := range writes {
			_, err = db.ExecContext(ctx, write.statement, write.args...)
			require.NoError(t, err, "seeding %s", write.table)
			require.Equal(t, 1, countFor(ctx, t, db, write.table, userID), "seeding %s", write.table)
		}

		_, err = db.ExecContext(ctx, "DELETE FROM ddb_identity_users WHERE id = $1", userID)
		require.NoError(t, err)

		for _, write := range writes {
			assert.Zero(t, countFor(ctx, t, db, write.table, userID),
				"%s outlived the user it belongs to", write.table)
		}
	})
}

// countFor counts the rows in table belonging to userID.
//
// The table name is interpolated because a table name cannot be a bind parameter and the
// three values it takes are constants above, not input.
func countFor(ctx context.Context, t *testing.T, db *sql.DB, table, userID string) int {
	t.Helper()

	var count int
	// #nosec G201 -- the table name is one of three constants declared in this file.
	require.NoError(t, db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM "+table+" WHERE belongs_to_user = $1", userID).Scan(&count))

	return count
}
