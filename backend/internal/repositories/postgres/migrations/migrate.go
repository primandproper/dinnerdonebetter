package migrations

import (
	"embed"
	"io/fs"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	auditmigrations "github.com/primandproper/platform-go/v9/audit/migrations"
	"github.com/primandproper/platform-go/v9/database/dialect"
	"github.com/primandproper/platform-go/v9/database/migrate"
	"github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/outbox"
	outboxmigrations "github.com/primandproper/platform-go/v9/outbox/migrations"
)

var (
	//go:embed migration_files/*.sql
	rawMigrations embed.FS
)

// lockKey names the Postgres advisory lock that serializes migrations. Every
// deployment sharing a database derives the same lock ID from it, so one
// replica applies migrations while the rest wait rather than racing.
const lockKey = "dinnerdonebetter"

// Where the platform's own tables land in this repository's migration ordering. The platform
// does not ship numbered files — numbering is global per consumer, so a platform-owned number
// would collide the moment either side added one — and hands us the DDL instead. Keep these
// above every file in migration_files.
const (
	outboxMigrationVersion = 22

	// The audit tables come after 00023, which drops the hand-rolled log they
	// replace. The order does not strictly matter, since the prefix keeps the two
	// schemas from colliding, but a reader following the sequence should see the
	// old thing go and the new thing arrive in that order.
	auditMigrationVersion = 24

	// The append-only triggers are their own migration rather than part of the
	// schema above, because they are separately privileged: the Postgres variant
	// creates a function. A deployment that would rather revoke UPDATE and DELETE
	// from the application role has a strictly stronger guarantee without them,
	// and can skip this one.
	auditAppendOnlyMigrationVersion = 25
)

// NewMigrator creates a new postgres Migrator over the embedded migration files.
//
// Migrations are ordered by the leading number in their filename, so adding one
// means dropping a numbered .sql file into migration_files — there is no list
// here to keep in sync. Files are read and checked here, so a malformed
// migration fails construction rather than the first Migrate.
func NewMigrator(logger logging.Logger) (*migrate.Migrator, error) {
	migrationFiles, err := fs.Sub(rawMigrations, "migration_files")
	if err != nil {
		return nil, errors.Wrap(err, "opening migration files")
	}

	// The outbox table's DDL is rendered from the platform rather than copied into
	// migration_files, so it stays in sync as that package evolves.
	outboxDDL, err := outboxmigrations.SQL(dialect.Postgres, outbox.DefaultTablePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "rendering outbox migration")
	}

	// Likewise the audit tables. The prefix is the domain's constant rather than a
	// literal, so the tables the migration creates are by construction the ones the
	// Recorder writes to and the Sweeper prunes.
	auditDDL, err := auditmigrations.SQL(dialect.Postgres, audit.TablePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "rendering audit migration")
	}

	auditAppendOnlyDDL, err := renderAuditAppendOnly()
	if err != nil {
		return nil, err
	}

	migrator, err := migrate.New(
		dialect.Postgres,
		migrationFiles,
		migrate.WithLogger(logging.EnsureLogger(logger)),
		migrate.WithLockKey(lockKey),
		migrate.WithGeneratedMigration(outboxMigrationVersion, "create_outbox_messages", outboxDDL),
		migrate.WithGeneratedMigration(auditMigrationVersion, "create_audit_tables", auditDDL),
		migrate.WithGeneratedMigration(auditAppendOnlyMigrationVersion, "enforce_audit_append_only", auditAppendOnlyDDL),
	)
	if err != nil {
		return nil, errors.Wrap(err, "building migrator")
	}

	return migrator, nil
}

// renderAuditAppendOnly assembles the triggers that make the entries table
// refuse an UPDATE.
//
// The statements are fenced individually rather than joined, and that is the
// whole reason this is not a one-line SQL call. goose splits a migration into
// statements on semicolons, and the Postgres variant's trigger function is a
// dollar-quoted body full of them — joined, the next tool to split on a
// semicolon gets two halves of a function and no way to notice.
//
// DELETE is deliberately left permitted. Retention has to remove aged entries
// and no trigger can tell that sweep apart from an attacker, so deletion is
// covered by the chain instead: entries carry contiguous positions within a
// scope, so a removed row leaves a hole Verify reports, and the sweep records
// where it pruned to so its own holes are distinguishable from everyone else's.
func renderAuditAppendOnly() (string, error) {
	statements, err := auditmigrations.AppendOnlyStatements(dialect.Postgres, audit.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering audit append-only statements")
	}

	var body strings.Builder
	for _, statement := range statements {
		body.WriteString("-- +goose StatementBegin\n")
		body.WriteString(statement)
		body.WriteString(";\n-- +goose StatementEnd\n\n")
	}

	return body.String(), nil
}
