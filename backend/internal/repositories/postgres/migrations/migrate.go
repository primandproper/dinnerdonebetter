package migrations

import (
	"embed"
	"io/fs"
	"strings"

	ddbaudit "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	ddbcomments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	ddbdataprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	ddboauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"

	auditmigrations "github.com/primandproper/platform-go/v13/audit/migrations"
	oauth2migrations "github.com/primandproper/platform-go/v13/authentication/oauth2server/database/migrations"
	passwordresetmigrations "github.com/primandproper/platform-go/v13/authentication/passwordreset/migrations"
	webauthndatabase "github.com/primandproper/platform-go/v13/authentication/webauthn/database"
	webauthnmigrations "github.com/primandproper/platform-go/v13/authentication/webauthn/database/migrations"
	commentsmigrations "github.com/primandproper/platform-go/v13/comments/migrations"
	"github.com/primandproper/platform-go/v13/database/dialect"
	"github.com/primandproper/platform-go/v13/database/migrate"
	dataprivacymigrations "github.com/primandproper/platform-go/v13/dataprivacy/migrations"
	"github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/metering"
	meteringmigrations "github.com/primandproper/platform-go/v13/metering/migrations"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/operations"
	operationsmigrations "github.com/primandproper/platform-go/v13/operations/migrations"
	"github.com/primandproper/platform-go/v13/outbox"
	outboxmigrations "github.com/primandproper/platform-go/v13/outbox/migrations"
	"github.com/primandproper/platform-go/v13/saga"
	sagamigrations "github.com/primandproper/platform-go/v13/saga/migrations"
	sessionsmigrations "github.com/primandproper/platform-go/v13/sessions/database/migrations"
	"github.com/primandproper/platform-go/v13/webhooks"
	webhooksmigrations "github.com/primandproper/platform-go/v13/webhooks/migrations"
	"github.com/primandproper/platform-go/v13/workqueue"
	workqueuemigrations "github.com/primandproper/platform-go/v13/workqueue/migrations"
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
	outboxMigrationVersion        = 22
	sagaMigrationVersion          = 24
	webhooksMigrationVersion      = 25
	auditMigrationVersion         = 27
	dataPrivacyMigrationVersion   = 28
	meteringMigrationVersion      = 30
	operationsMigrationVersion    = 31
	webauthnMigrationVersion      = 32
	oauth2MigrationVersion        = 33
	passwordResetMigrationVersion = 34
	sessionsMigrationVersion      = 35
	workQueueMigrationVersion     = 36
	commentsMigrationVersion      = 37
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

	// And the operations table. v10 fulfills data privacy requests as operations, so the
	// tier that used to be internal to the dataprivacy service is now a durable record of
	// its own: one row per unit of tracked work, claimed by a worker and polled by clients.
	operationsDDL, err := operationsmigrations.SQL(dialect.Postgres, operations.DefaultTablePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "rendering operations migration")
	}

	webauthnDDL, err := renderWebAuthnDDL()
	if err != nil {
		return nil, err
	}

	// And the authorization server's four tables — registered clients, authorization codes,
	// access tokens, and refresh tokens. They are created together because the store that
	// reads them is one interface, and a deployment holding three of the four has a server
	// that fails at whichever step the missing one serves.
	oauth2DDL, err := oauth2migrations.SQL(dialect.Postgres, ddboauth.TablePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "rendering oauth2 server migration")
	}

	passwordResetDDL, err := renderPasswordResetDDL()
	if err != nil {
		return nil, err
	}

	sessionsDDL, err := renderSessionsDDL()
	if err != nil {
		return nil, err
	}

	// And the leased work queue's one table, which meal plan task notifications are claimed
	// from. Its two partial indexes are the claim and reap predicates: copied by hand, they
	// are how a claim that should touch the ready backlog starts scanning every item the
	// queue has ever held.
	//
	// One table serves every logical queue — the queue's name is the leading column of its
	// primary key — so a second queue is a second Config, not a second migration. That is
	// not hypothetical here: the operations tier is built on this same package and has been
	// claiming from this table since #1367, against a table nothing created. Its migration
	// creates the operations rows, not the queue rows the dispatch runs on, and the platform
	// leaves the queue's DDL to the consumer precisely because migration numbers are ours.
	// So this creates a table two queues share rather than one.
	workQueueDDL, err := workqueuemigrations.SQL(dialect.Postgres, workqueue.DefaultTablePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "rendering work queue migration")
	}

	commentsDDL, err := renderCommentsDDL()
	if err != nil {
		return nil, err
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
		migrate.WithGeneratedMigration(operationsMigrationVersion, "create_operations_table", operationsDDL),
		migrate.WithGeneratedMigration(webauthnMigrationVersion, "create_webauthn_sessions_table", webauthnDDL),
		migrate.WithGeneratedMigration(oauth2MigrationVersion, "create_oauth2_server_tables", oauth2DDL),
		migrate.WithGeneratedMigration(passwordResetMigrationVersion, "create_password_reset_tokens_table", passwordResetDDL),
		migrate.WithGeneratedMigration(sessionsMigrationVersion, "create_sessions_table", sessionsDDL),
		migrate.WithGeneratedMigration(workQueueMigrationVersion, "create_work_queue_items_table", workQueueDDL),
		migrate.WithGeneratedMigration(commentsMigrationVersion, "create_comments_table", commentsDDL),
	)
	if err != nil {
		return nil, errors.Wrap(err, "building migrator")
	}

	return migrator, nil
}

// renderCommentsDDL renders the comment table, dropping the one 00012_comments.sql
// created first, along with the enum that migration and 00021_mealplanning.sql
// between them defined.
//
// Nothing is carried across, and the two tables could not carry it anyway: the
// platform's names the columns differently (body, author, target_id, parent_id),
// adds the tenancy column every one of its reads filters on, and drops both
// foreign keys. The foreign keys are the substantive loss and they are lost by
// construction rather than by choice — a comment's target lives in a table the
// platform's store has never seen, so there is no column it could point at. What
// replaced the belongs_to_user cascade is the comments eraser, registered in
// internal/build/dataprivacy; what replaced the target cascade is nothing, which
// is the ruling platform's package documentation states plainly.
//
// The enum goes with the table because the platform stores target_type as text.
// That is the right shape here regardless: adding a target type was an ALTER TYPE
// in a migration, and it is now a line in the catalog that a compiler checks.
//
// The old table is dropped rather than left in place because its name is the one
// the platform's default prefix would render, and its DDL says CREATE TABLE IF
// NOT EXISTS — so a deployment that kept it would eventually get a silent no-op
// followed by a store reading columns that are not there. This renders
// ddb_comments; see ddbcomments.TablePrefix.
func renderCommentsDDL() (string, error) {
	schema, err := commentsmigrations.SQL(dialect.Postgres, ddbcomments.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering comments migration")
	}

	return "DROP TABLE IF EXISTS comments;\nDROP TYPE IF EXISTS comment_target_type;\n\n" + schema, nil
}

// renderPasswordResetDDL renders the password reset token table, dropping the one
// 00003_auth.sql created first.
//
// The drop is not a migration of the old table, and nothing is carried across. A row is
// one outstanding reset link, the links last thirty minutes, and the column the old table
// stored the token in held the token itself — so there is nothing to translate into the
// new digest column that would not amount to writing the secrets back down. The worst
// case at deploy is a handful of people clicking "email me a link" again.
//
// The old table is dropped rather than left in place because it is dead weight holding
// live credentials: a table of plaintext reset tokens that nothing reads, and that no
// sweeper empties, is a backup waiting to leak. Its name is also the platform schema's,
// which is why the new table carries auth.TablePrefix — see that constant.
func renderPasswordResetDDL() (string, error) {
	schema, err := passwordresetmigrations.SQL(dialect.Postgres, ddbauth.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering password reset token migration")
	}

	return "DROP TABLE IF EXISTS password_reset_tokens;\n\n" + schema, nil
}

// renderSessionsDDL renders the session table, dropping the one 00018_user_sessions.sql
// created first.
//
// Nothing is carried across, and the two tables are not the same shape anyway. The old one
// keyed a session by the JTI of the token issued alongside it and ended one by stamping a
// revoked_at column that only the reads knew to filter on; the platform's keys a session by
// an identifier of its own and ends one by removing the row, so there is no state a
// revocation can be read past. The worst case at deploy is that everybody signs in again,
// which is what a session store's rows are for.
//
// The old table is dropped rather than left behind because a second table recording which
// sessions are live is the one thing sessions/database exists to prevent: the moment the
// two disagree a revocation has not taken, and nothing says which of them was right.
//
// The prefix is auth.TablePrefix, so this renders ddb_sessions. The platform's own name is
// "sessions", generic enough that a database shared with anything else would eventually
// collide — and its DDL says CREATE TABLE IF NOT EXISTS, so the collision would be a silent
// no-op followed by a store reading columns that are not there.
func renderSessionsDDL() (string, error) {
	schema, err := sessionsmigrations.SQL(dialect.Postgres, ddbauth.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering sessions migration")
	}

	return "DROP TABLE IF EXISTS user_sessions;\n\n" + schema, nil
}

// renderWebAuthnDDL renders the passkey ceremony session table, dropping the one this
// repository used to hand-write first.
//
// The drop is not a migration of the old table: 00017 created it with a JSONB session_data
// column and the platform's schema stores BYTEA, so a CREATE TABLE IF NOT EXISTS over the
// old one would silently keep a table the store cannot write to. Nothing is lost by dropping
// it. A row is one passkey ceremony in flight, a ceremony lasts a minute, and the worst case
// at deploy is a handful of users pressing the passkey prompt again.
func renderWebAuthnDDL() (string, error) {
	schema, err := webauthnmigrations.SQL(dialect.Postgres, webauthndatabase.DefaultTablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering webauthn session migration")
	}

	return "DROP TABLE IF EXISTS webauthn_sessions;\n\n" + schema, nil
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
