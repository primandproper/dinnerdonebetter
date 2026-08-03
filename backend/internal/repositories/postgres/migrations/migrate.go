package migrations

import (
	"embed"
	"io/fs"
	"strings"

	ddbaudit "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbdataprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"

	auditmigrations "github.com/primandproper/platform-go/v9/audit/migrations"
	"github.com/primandproper/platform-go/v9/database/dialect"
	"github.com/primandproper/platform-go/v9/database/migrate"
	dataprivacymigrations "github.com/primandproper/platform-go/v9/dataprivacy/migrations"
	"github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/metering"
	meteringmigrations "github.com/primandproper/platform-go/v9/metering/migrations"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/outbox"
	outboxmigrations "github.com/primandproper/platform-go/v9/outbox/migrations"
	"github.com/primandproper/platform-go/v9/saga"
	sagamigrations "github.com/primandproper/platform-go/v9/saga/migrations"
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

// Where the platform's own tables land in this repository's migration ordering. The platform
// ships no numbered files — numbering is global per consumer, so a platform-owned number would
// collide the moment either side added one — and hands us the DDL instead.
//
// The numbering is one sequence shared with migration_files, so these must not collide with a
// filename and must never be renumbered once applied. Adding another means taking the next
// free number, whichever side it comes from.
const (
	outboxMigrationVersion      = 22
	sagaMigrationVersion        = 24
	webhooksMigrationVersion    = 25
	auditMigrationVersion       = 27
	dataPrivacyMigrationVersion = 28
	meteringMigrationVersion    = 30
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

	// Likewise the metering event ledger and totals tables. The library owns that schema
	// because its counting logic is inseparable from it — the ingest dedupe is a primary key,
	// the concurrent fold is an UPDATE expression, and Consume's atomicity is a row lock.
	meteringDDL, err := meteringmigrations.SQL(dialect.Postgres, metering.DefaultTablePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "rendering metering migration")
	}

	auditDDL, err := renderAuditDDL()
	if err != nil {
		return nil, err
	}

	// Likewise for the saga instance table, which durable meal plan finalization runs on.
	sagaDDL, err := sagamigrations.SQL(dialect.Postgres, saga.DefaultTablePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "rendering saga migration")
	}

	// Likewise the five webhook tables — endpoints, subscriptions, deliveries, dispatches, and
	// attempts — together with the partial indexes the claim predicate depends on. Copying
	// those by hand is how a claim quietly starts scanning history instead of backlog.
	webhooksDDL, err := webhooksmigrations.SQL(dialect.Postgres, webhooks.DefaultTablePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "rendering webhooks migration")
	}

	// And the data privacy request table, which is one row per export or erasure —
	// replacing user_data_disclosures, whose partial indexes it also carries. The
	// claim, expiry, and overdue predicates all depend on those being partial; copied
	// by hand, they are how a sweep that should touch the backlog starts scanning
	// every request the system has ever served.
	dataPrivacyDDL, err := dataprivacymigrations.SQL(dialect.Postgres, ddbdataprivacy.TablePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "rendering data privacy migration")
	}

	migrator, err := migrate.New(
		dialect.Postgres,
		migrationFiles,
		migrate.WithLogger(logging.EnsureLogger(logger)),
		migrate.WithLockKey(lockKey),
		migrate.WithGeneratedMigration(outboxMigrationVersion, "create_outbox_messages", outboxDDL),
		migrate.WithGeneratedMigration(sagaMigrationVersion, "create_saga_instances", sagaDDL),
		migrate.WithGeneratedMigration(webhooksMigrationVersion, "create_webhooks_tables", webhooksDDL),
		migrate.WithGeneratedMigration(auditMigrationVersion, "create_audit_tables", auditDDL),
		migrate.WithGeneratedMigration(dataPrivacyMigrationVersion, "create_dataprivacy_requests", dataPrivacyDDL),
		migrate.WithGeneratedMigration(meteringMigrationVersion, "create_metering_tables", meteringDDL),
	)
	if err != nil {
		return nil, errors.Wrap(err, "building migrator")
	}

	return migrator, nil
}

// renderAuditDDL renders the audit tables and the triggers that make them
// append-only, as one migration body.
//
// The triggers are the half worth explaining. Without them, editing a recorded
// entry is something the hash chain reveals after the fact; with them it is
// something the database refuses outright, and a guarantee enforced at write time
// is worth more than one enforced at audit time. DELETE is deliberately still
// permitted — retention has to remove aged entries and no trigger can tell that
// sweep apart from an attacker — and the chain covers deletion instead.
//
// The platform hands the triggers back pre-split and refuses to join them,
// because goose splits a migration into statements on semicolons and the
// Postgres trigger function has semicolons inside its body — joined naively, the
// migrator would be handed two halves of a trigger. Each statement is therefore
// fenced individually with StatementBegin/StatementEnd, which is what tells goose
// to execute it whole. Getting this wrong fails at construction rather than
// mid-deploy: the annotator refuses a dollar-quoted body with no fence.
func renderAuditDDL() (string, error) {
	schema, err := auditmigrations.SQL(dialect.Postgres, ddbaudit.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering audit migration")
	}

	appendOnly, err := auditmigrations.AppendOnlyStatements(dialect.Postgres, ddbaudit.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering audit append-only triggers")
	}

	body := &strings.Builder{}
	body.WriteString(schema)

	for _, statement := range appendOnly {
		body.WriteString("\n-- +goose StatementBegin\n")
		body.WriteString(statement)
		body.WriteString(";\n-- +goose StatementEnd\n")
	}

	return body.String(), nil
}
