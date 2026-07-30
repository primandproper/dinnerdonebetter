package migrations

import (
	"embed"
	"io/fs"

	"github.com/primandproper/platform-go/v8/database/dialect"
	"github.com/primandproper/platform-go/v8/database/migrate"
	"github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/outbox"
	outboxmigrations "github.com/primandproper/platform-go/v8/outbox/migrations"
)

var (
	//go:embed migration_files/*.sql
	rawMigrations embed.FS
)

// lockKey names the Postgres advisory lock that serializes migrations. Every
// deployment sharing a database derives the same lock ID from it, so one
// replica applies migrations while the rest wait rather than racing.
const lockKey = "dinnerdonebetter"

// outboxMigrationVersion is where the platform's outbox table lands in this repository's
// migration ordering. The platform does not ship a numbered file — numbering is global per
// consumer, so a platform-owned number would collide the moment either side added one — and
// hands us the DDL instead. Keep it above every file in migration_files.
const outboxMigrationVersion = 22

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
	outboxDDL, err := outboxmigrations.SQL(dialect.Postgres, outbox.DefaultTableName)
	if err != nil {
		return nil, errors.Wrap(err, "rendering outbox migration")
	}

	migrator, err := migrate.New(
		dialect.Postgres,
		migrationFiles,
		migrate.WithLogger(logging.EnsureLogger(logger)),
		migrate.WithLockKey(lockKey),
		migrate.WithGeneratedMigration(outboxMigrationVersion, "create_outbox_messages", outboxDDL),
	)
	if err != nil {
		return nil, errors.Wrap(err, "building migrator")
	}

	return migrator, nil
}
