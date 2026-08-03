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
	"github.com/primandproper/platform-go/v9/webhooks"
	webhooksmigrations "github.com/primandproper/platform-go/v9/webhooks/migrations"
)

var (
	//go:embed migration_files/*.sql
	rawMigrations embed.FS
)

// lockKey names the Postgres advisory lock that serializes migrations. Every
// deployment sharing a database derives the same lock ID from it, so one
// replica applies migrations while the rest wait rather than racing.
const lockKey = "dinnerdonebetter"

// Versions for the platform-supplied schemas. The platform ships no numbered files — numbering
// is global per consumer, so a platform-owned number would collide the moment either side added
// one — and hands us the DDL instead.
//
// These interleave with migration_files rather than sitting above all of them, because a
// numbered file may need to run after one of them: 00023 backfills the webhook endpoints table,
// which cannot exist before 23 creates it.
const (
	// outboxMigrationVersion is where the platform's outbox table lands.
	outboxMigrationVersion = 22

	// webhooksMigrationVersion is where the platform's five webhook tables land. It must stay
	// below 00023_webhooks_platform.sql, which backfills endpoints for existing webhooks.
	webhooksMigrationVersion = 23
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

	// Likewise the five webhook tables — endpoints, subscriptions, deliveries, dispatches, and
	// attempts — together with the partial indexes the claim predicate depends on. Copying
	// those by hand is how a claim quietly starts scanning history instead of backlog.
	webhooksDDL, err := webhooksmigrations.SQL(dialect.Postgres, webhooks.DefaultTablePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "rendering webhooks migration")
	}

	migrator, err := migrate.New(
		dialect.Postgres,
		migrationFiles,
		migrate.WithLogger(logging.EnsureLogger(logger)),
		migrate.WithLockKey(lockKey),
		migrate.WithGeneratedMigration(outboxMigrationVersion, "create_outbox_messages", outboxDDL),
		migrate.WithGeneratedMigration(webhooksMigrationVersion, "create_webhooks_tables", webhooksDDL),
	)
	if err != nil {
		return nil, errors.Wrap(err, "building migrator")
	}

	return migrator, nil
}
