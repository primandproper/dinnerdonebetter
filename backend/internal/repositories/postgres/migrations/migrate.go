package migrations

import (
	"embed"
	"io/fs"

	"github.com/primandproper/platform-go/v9/database/dialect"
	"github.com/primandproper/platform-go/v9/database/migrate"
	"github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/outbox"
	outboxmigrations "github.com/primandproper/platform-go/v9/outbox/migrations"
	"github.com/primandproper/platform-go/v9/saga"
	sagamigrations "github.com/primandproper/platform-go/v9/saga/migrations"
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
// ships no numbered files — numbering is global per consumer, so a platform-owned number would
// collide the moment either side added one — and hands us the DDL instead.
//
// The numbering is one sequence shared with migration_files, so these must not collide with a
// filename and must never be renumbered once applied. Adding another means taking the next
// free number, whichever side it comes from.
const (
	outboxMigrationVersion = 22
	sagaMigrationVersion   = 24
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

	// Likewise for the saga instance table, which durable meal plan finalization runs on.
	sagaDDL, err := sagamigrations.SQL(dialect.Postgres, saga.DefaultTablePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "rendering saga migration")
	}

	migrator, err := migrate.New(
		dialect.Postgres,
		migrationFiles,
		migrate.WithLogger(logging.EnsureLogger(logger)),
		migrate.WithLockKey(lockKey),
		migrate.WithGeneratedMigration(outboxMigrationVersion, "create_outbox_messages", outboxDDL),
		migrate.WithGeneratedMigration(sagaMigrationVersion, "create_saga_instances", sagaDDL),
	)
	if err != nil {
		return nil, errors.Wrap(err, "building migrator")
	}

	return migrator, nil
}
