package migrations

import (
	"strings"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v11/identifiers"
	"github.com/primandproper/platform-go/v11/metering"
	loggingnoop "github.com/primandproper/platform-go/v11/observability/logging/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuerier_Migrate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		db, _ := pgtesting.BuildDatabaseContainerForTest(t)
		migrator, err := NewMigrator(loggingnoop.NewLogger())
		require.NoError(t, err)
		require.NoError(t, migrator.Migrate(ctx, db))
	})

	// The audit tables are append-only at the database, not merely by convention.
	// This is the assertion that the trigger survived being rendered, fenced, and
	// applied — an UPDATE that quietly succeeded would mean the chain is the only
	// thing standing between a tampered row and a clean read.
	T.Run("audit entries reject updates", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		db, _ := pgtesting.BuildDatabaseContainerForTest(t)
		migrator, err := NewMigrator(loggingnoop.NewLogger())
		require.NoError(t, err)
		require.NoError(t, migrator.Migrate(ctx, db))

		entryID, scope, actorID := identifiers.New(), identifiers.New(), identifiers.New()
		entries := audit.TablePrefix + "_audit_log_entries"

		_, err = db.ExecContext(ctx,
			`INSERT INTO `+entries+`
			 (id, seq, scope, recorded_at, event_type, resource_type, resource_id, actor_id, hash)
			 VALUES ($1, 0, $2, NOW(), 'created', 'recipes', $3, $4, 'deadbeef')`,
			entryID, scope, identifiers.New(), actorID)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx,
			`UPDATE `+entries+` SET resource_id = $1 WHERE id = $2`, identifiers.New(), entryID)
		require.Error(t, err, "the database must refuse to edit a recorded entry")
		assert.Contains(t, err.Error(), "append-only")

		// DELETE stays permitted: retention has to remove aged entries, and no trigger
		// can tell that sweep apart from an attacker. The chain covers deletion instead.
		_, err = db.ExecContext(ctx, `DELETE FROM `+entries+` WHERE id = $1`, entryID)
		require.NoError(t, err, "retention has to be able to delete")
	})

	// The metering event ledger dedupes on (meter, idempotency_key) rather than on the key
	// alone, and the difference is money. Callers are told to key usage by the identifier of
	// the thing that caused it, and one such thing routinely feeds more than one meter — an
	// upload that bills both a byte count and a request count. Keyed on the key alone, the
	// second meter's insert is silently deduped against the first and that meter is
	// under-billed forever. This asserts the fix landed in the DDL we actually apply.
	T.Run("metering events dedupe per meter", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		db, _ := pgtesting.BuildDatabaseContainerForTest(t)
		migrator, err := NewMigrator(loggingnoop.NewLogger())
		require.NoError(t, err)
		require.NoError(t, migrator.Migrate(ctx, db))

		events := metering.DefaultTablePrefix + "metering_events"
		key, subject := identifiers.New(), identifiers.New()

		insert := func(meter string) error {
			_, execErr := db.ExecContext(ctx,
				`INSERT INTO `+events+`
				 (idempotency_key, subject, meter, quantity, occurred_at, recorded_at, period_start)
				 VALUES ($1, $2, $3, 1, NOW(), NOW(), NOW())`,
				key, subject, meter)

			return execErr
		}

		require.NoError(t, insert("uploaded_media_bytes"))
		require.NoError(t, insert("api_requests"), "one key feeding two meters must record twice")
		require.Error(t, insert("api_requests"), "the same key on the same meter must record once")
	})
}

func TestRenderAuditDDL(T *testing.T) {
	T.Parallel()

	T.Run("renders both tables", func(t *testing.T) {
		t.Parallel()

		body, err := renderAuditDDL()
		require.NoError(t, err)

		// Prefixed, so that the DDL cannot land on the name the hand-rolled log used.
		assert.Contains(t, body, audit.TablePrefix+"_audit_log_entries")
		assert.Contains(t, body, audit.TablePrefix+"_audit_log_chains")

		// The uniqueness constraint is the guarantee rather than an index for speed:
		// it is what makes a forked chain something the table cannot hold, instead of
		// something a verifier has to detect after the fact.
		assert.Contains(t, body, "CREATE UNIQUE INDEX")
	})

	// The trigger function body contains semicolons, and goose splits a migration on
	// semicolons unless the statement is fenced. Unfenced, the body reaches the
	// database as fragments — so this is checked here rather than discovered on a
	// deploy.
	T.Run("fences every append-only statement", func(t *testing.T) {
		t.Parallel()

		body, err := renderAuditDDL()
		require.NoError(t, err)

		assert.Contains(t, body, "audit_log_entries_reject_update")
		assert.Contains(t, body, "CREATE TRIGGER")

		begins := strings.Count(body, "-- +goose StatementBegin")
		ends := strings.Count(body, "-- +goose StatementEnd")

		assert.Positive(t, begins)
		assert.Equal(t, begins, ends, "every fence must be closed")

		// The dollar-quoted function body has to sit inside a fence, not beside one.
		fnAt := strings.Index(body, "$$ LANGUAGE plpgsql")
		require.Positive(t, fnAt)
		assert.Positive(t, strings.LastIndex(body[:fnAt], "-- +goose StatementBegin"),
			"the function body must open a fence before it")
		assert.Positive(t, strings.Index(body[fnAt:], "-- +goose StatementEnd"),
			"the function body must close its fence after it")
	})

	T.Run("rejects updates but not deletes", func(t *testing.T) {
		t.Parallel()

		body, err := renderAuditDDL()
		require.NoError(t, err)

		assert.Contains(t, body, "BEFORE UPDATE ON")
		assert.NotContains(t, body, "BEFORE DELETE ON")
	})
}

// TestQuerier_Migrate_OverLegacyAuditTable is the scenario the table prefix exists
// for: a database that already holds the hand-rolled log's table.
//
// The platform's DDL says CREATE TABLE IF NOT EXISTS, so without the prefix this
// migration would be a silent no-op against that table and every audit write
// afterwards would target the wrong columns. With it there is nothing to collide
// with, and the stale table is left alone rather than dropped — a migration that
// destroys an audit log is the wrong default even when we are confident it holds
// nothing worth keeping.
func TestQuerier_Migrate_OverLegacyAuditTable(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		db, _ := pgtesting.BuildDatabaseContainerForTest(t)

		_, err := db.ExecContext(ctx, `
			CREATE TYPE audit_log_event_type AS ENUM ('other', 'created', 'updated', 'archived');
			CREATE TABLE audit_log_entries (
				id TEXT NOT NULL PRIMARY KEY,
				resource_type TEXT NOT NULL,
				relevant_id TEXT NOT NULL DEFAULT '',
				event_type audit_log_event_type NOT NULL DEFAULT 'other',
				changes JSONB NOT NULL,
				created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL
			)`)
		require.NoError(t, err)

		migrator, err := NewMigrator(loggingnoop.NewLogger())
		require.NoError(t, err)
		require.NoError(t, migrator.Migrate(ctx, db))

		// The prefixed tables exist and are the platform's, not the legacy shape.
		var seqColumns int
		require.NoError(t, db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM information_schema.columns WHERE table_name = $1 AND column_name = 'seq'`,
			audit.TablePrefix+"_audit_log_entries").Scan(&seqColumns))
		assert.Equal(t, 1, seqColumns, "the chained entries table must be the one that got created")

		// And the legacy table is untouched rather than dropped.
		var legacy int
		require.NoError(t, db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'audit_log_entries'`).Scan(&legacy))
		assert.Equal(t, 1, legacy)
	})
}
