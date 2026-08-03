package migrations

import (
	"strings"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	auditmigrations "github.com/primandproper/platform-go/v9/audit/migrations"
	"github.com/primandproper/platform-go/v9/database/dialect"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"

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
}

func TestRenderAuditAppendOnly(T *testing.T) {
	T.Parallel()

	// The fence is the whole point of the function. goose splits a migration into
	// statements on semicolons, and the trigger function's dollar-quoted body is
	// full of them, so an unfenced body reaches the database as fragments rather
	// than as a function — and does so silently.
	T.Run("fences every statement", func(t *testing.T) {
		t.Parallel()

		body, err := renderAuditAppendOnly()
		require.NoError(t, err)

		statements, err := auditmigrations.AppendOnlyStatements(dialect.Postgres, audit.TablePrefix)
		require.NoError(t, err)
		require.NotEmpty(t, statements)

		assert.Equal(t, len(statements), strings.Count(body, "-- +goose StatementBegin"))
		assert.Equal(t, len(statements), strings.Count(body, "-- +goose StatementEnd"))
		assert.Contains(t, body, "$$")
	})

	T.Run("targets the namespaced table", func(t *testing.T) {
		t.Parallel()

		body, err := renderAuditAppendOnly()
		require.NoError(t, err)

		assert.Contains(t, body, audit.TablePrefix+"_audit_log_entries")
	})
}
