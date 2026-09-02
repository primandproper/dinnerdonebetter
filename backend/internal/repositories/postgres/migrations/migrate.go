package migrations

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	ddbaudit "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	ddbcomments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	ddbdataprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	ddboauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	ddbsettings "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	ddbuploadedmedia "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	ddbwaitlists "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"

	auditmigrations "github.com/primandproper/platform-go/v13/audit/migrations"
	oauth2migrations "github.com/primandproper/platform-go/v13/authentication/oauth2server/database/migrations"
	passwordresetmigrations "github.com/primandproper/platform-go/v13/authentication/passwordreset/migrations"
	webauthndatabase "github.com/primandproper/platform-go/v13/authentication/webauthn/database"
	webauthnmigrations "github.com/primandproper/platform-go/v13/authentication/webauthn/database/migrations"
	authzdatabase "github.com/primandproper/platform-go/v13/authorization/database"
	authzmigrations "github.com/primandproper/platform-go/v13/authorization/database/migrations"
	commentsmigrations "github.com/primandproper/platform-go/v13/comments/migrations"
	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/database/ddl"
	"github.com/primandproper/platform-go/v13/database/dialect"
	"github.com/primandproper/platform-go/v13/database/migrate"
	dataprivacymigrations "github.com/primandproper/platform-go/v13/dataprivacy/migrations"
	"github.com/primandproper/platform-go/v13/errors"
	issuereportsmigrations "github.com/primandproper/platform-go/v13/issuereports/migrations"
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
	settingsmigrations "github.com/primandproper/platform-go/v13/settings/migrations"
	uploadsregistrymigrations "github.com/primandproper/platform-go/v13/uploads/registry/migrations"
	waitlistsmigrations "github.com/primandproper/platform-go/v13/waitlists/migrations"
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
	outboxMigrationVersion          = 22
	sagaMigrationVersion            = 24
	webhooksMigrationVersion        = 25
	auditMigrationVersion           = 27
	dataPrivacyMigrationVersion     = 28
	meteringMigrationVersion        = 30
	operationsMigrationVersion      = 31
	webauthnMigrationVersion        = 32
	oauth2MigrationVersion          = 33
	passwordResetMigrationVersion   = 34
	sessionsMigrationVersion        = 35
	workQueueMigrationVersion       = 36
	commentsMigrationVersion        = 37
	uploadsRegistryMigrationVersion = 38
	issueReportsMigrationVersion    = 39
	waitlistsMigrationVersion       = 40
	settingsMigrationVersion        = 41
	authorizationMigrationVersion   = 42
)

// NewMigrator creates a new postgres Migrator over the embedded migration files.
//
// Migrations are ordered by the leading number in their filename, so adding one
// means dropping a numbered .sql file into migration_files — there is no list
// here to keep in sync. Files are read and checked here, so a malformed
// migration fails construction rather than the first Migrate.
//
// The returned Migrator applies the schema and then seeds the authorization
// policy; see its Migrate.
func NewMigrator(logger logging.Logger) (*Migrator, error) {
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

	uploadsRegistryDDL, err := renderUploadsRegistryDDL()
	if err != nil {
		return nil, err
	}

	issueReportsDDL, err := renderIssueReportsDDL()
	if err != nil {
		return nil, err
	}

	waitlistsDDL, err := renderWaitlistsDDL()
	if err != nil {
		return nil, err
	}

	settingsDDL, err := renderSettingsDDL()
	if err != nil {
		return nil, err
	}

	authorizationDDL, err := renderAuthorizationDDL()
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
		migrate.WithGeneratedMigration(uploadsRegistryMigrationVersion, "create_uploads_objects_table", uploadsRegistryDDL),
		migrate.WithGeneratedMigration(issueReportsMigrationVersion, "create_issue_reports_table", issueReportsDDL),
		migrate.WithGeneratedMigration(waitlistsMigrationVersion, "create_waitlist_tables", waitlistsDDL),
		migrate.WithGeneratedMigration(settingsMigrationVersion, "create_settings_tables", settingsDDL),
		migrate.WithGeneratedMigration(authorizationMigrationVersion, "create_authorization_tables", authorizationDDL),
	)
	if err != nil {
		return nil, errors.Wrap(err, "building migrator")
	}

	return &Migrator{schema: migrator, logger: logging.EnsureLogger(logger)}, nil
}

// Migrator applies the schema and then seeds the authorization policy.
//
// The policy is not a migration. It is written by authorization/database's Seed
// from authorization.PlatformPolicy(), which is idempotent, upserts by name, and
// rewrites each named role's grants rather than adding to them — so a permission
// removed in Go is revoked on the next run, which is what makes the Go
// declaration the only one. Rendering it as INSERT statements instead would
// re-create the hand-maintained seed this adoption removed, and would lose the
// revoke.
//
// Seeding lives here, behind the same call every migrating process already
// makes, rather than at a wiring site: an unseeded policy grants nothing, so a
// process that forgot would come up refusing every request, and the container
// test harnesses that migrate a template database would each have to remember.
type Migrator struct {
	schema *migrate.Migrator
	logger logging.Logger
}

var _ database.Migrator = (*Migrator)(nil)

// seedLockKey names the advisory lock that serializes policy seeding.
//
// Migrations run at startup on every replica, and Seed is not safe to run
// concurrently with itself: it clears a role's grants and re-inserts them one
// statement at a time, and the insert carries no ON CONFLICT clause, so two
// replicas seeding the same policy either block on each other's row locks or
// collide on the (role_id, permission_id) primary key. The schema half is
// already serialized by the migrator's own lock; this covers the half that
// follows it.
//
// Filed as platform-go#463 — the fix belongs there, since a consumer holding a
// lock is not something the package can check. This is the local workaround that
// keeps a boot from failing meanwhile, and it goes when that lands.
const seedLockKey = "dinnerdonebetter.authorization.seed"

// Migrate applies every pending migration, then seeds the authorization policy.
func (m *Migrator) Migrate(ctx context.Context, db *sql.DB) error {
	if err := m.schema.Migrate(ctx, db); err != nil {
		return err
	}

	return m.seedPolicy(ctx, db)
}

func (m *Migrator) seedPolicy(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "beginning authorization policy transaction")
	}
	defer func() {
		// Rollback after a commit is a no-op that reports ErrTxDone, which is
		// why the error is dropped rather than joined.
		_ = tx.Rollback()
	}()

	// Held for the transaction, released by the commit below.
	if _, err = tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", seedLockKey); err != nil {
		return errors.Wrap(err, "locking authorization policy")
	}

	// Built against the transaction so the reads Seed makes to resolve role and
	// permission ids see the rows it has just written.
	resolver, err := authzdatabase.NewResolver(
		&authzdatabase.Config{Dialect: dialect.Postgres, TablePrefix: authorization.TablePrefix},
		tx,
		authzdatabase.WithLogger(m.logger),
	)
	if err != nil {
		return errors.Wrap(err, "building authorization resolver")
	}
	// Seed is on the concrete type rather than the PolicyResolver interface, which is
	// why this does not go through authorization.NewDatabaseResolver: reading policy
	// and writing it are different jobs, and only this one writes.

	// ValidateRoles runs inside Seed before anything is written, so a policy with
	// an unknown parent or an inheritance cycle fails the migration rather than
	// landing half-applied.
	if err = resolver.Seed(ctx, tx, authorization.PlatformPolicy()...); err != nil {
		return errors.Wrap(err, "seeding authorization policy")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "committing authorization policy")
	}

	return nil
}

// renderAuthorizationDDL renders the four policy tables — roles, permissions,
// the grants between them, and the inheritance edges — dropping the four that
// 00019_rbac.sql created, and re-creating the one foreign key the platform
// cannot ship.
//
// What this deletes is not a table so much as a second copy of the policy. The
// permissions a role holds were declared twice: as the slices in
// internal/authorization, which the method table and the platform policy are
// written from, and as ~600 lines of INSERT statements across 00019 and 00021,
// which is what authorization actually read. The only thing holding them
// together was a test that string-matched permission names against the
// concatenated migration text, so it could see a name that was never seeded and
// could not see a mapping that was wrong. They had drifted on three of five
// roles. The seed now runs from PlatformPolicy() — see Migrator.Migrate — and
// there is one declaration.
//
// The drop order is the reference order: the two mapping tables and the
// assignment's constraint go before the tables they point at. user_roles takes
// CASCADE because user_role_assignments.role_id referenced it, and that column
// is being replaced by role_name in the same migration sequence.
//
// The foreign key targets ddb_authz_roles(name) rather than its primary key.
// That is legal because the platform indexes name uniquely — deliberately, since
// "reusing the name of an archived role would silently re-grant its authority to
// everyone still assigned it" — and it is what an assignment has to reference,
// because a statement here cannot join a table sqlc's schema does not contain.
// ON DELETE RESTRICT rather than CASCADE: Seed and UpsertRole never delete a
// role row, ArchiveRole soft-deletes, so the only thing this can refuse is a
// hard delete somebody did by hand, and refusing that is right.
//
// The key guarantees the name exists, not that the role is live. An assignment
// naming an archived role resolves to nothing, because the resolution query
// applies the archived predicate at every join — which is fail-closed, and the
// behaviour we want.
//
// user_roles.scope, which constrained a role to 'service' or 'account', has no
// platform counterpart and is not re-created. Nothing read it: no query filtered
// on it, and the one place it should have mattered — ModifyUserPermissions,
// which writes a caller-supplied role name into an account-scoped assignment —
// never consulted it. That is now an allow-list on the input, which is enforced
// where the column was not.
func renderAuthorizationDDL() (string, error) {
	schema, err := authzmigrations.SQL(dialect.Postgres, authorization.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering authorization migration")
	}

	rolesTable := ddl.Qualify(authorization.TablePrefix) + "authz_roles"

	body := &strings.Builder{}
	body.WriteString("DROP TABLE IF EXISTS user_role_permissions;\n")
	body.WriteString("DROP TABLE IF EXISTS user_role_hierarchy;\n")
	body.WriteString("DROP TABLE IF EXISTS permissions;\n")
	body.WriteString("DROP TABLE IF EXISTS user_roles CASCADE;\n\n")
	body.WriteString(schema)
	body.WriteString("\n\nALTER TABLE user_role_assignments\n\tADD CONSTRAINT user_role_assignments_role_fk\n\tFOREIGN KEY (role_name) REFERENCES " + rolesTable + "(name) ON DELETE RESTRICT;\n")

	return body.String(), nil
}

// renderIssueReportsDDL renders the issue report table, dropping the one
// 00009_issue_reports.sql created first, and re-creating the two foreign keys
// that table carried.
//
// Nothing is carried across, and the two tables could not carry it anyway: the
// platform's names the columns differently (reporter, kind, subject_type,
// subject_id), swaps belongs_to_account for the tenancy column every one of its
// reads filters on, and adds the three columns this package was adopted for —
// status, resolution and closed_at, which are what turn a pile of submissions
// into a queue somebody can work.
//
// The old table is dropped rather than left in place because its name is the one
// the platform's default prefix would render, and its DDL says CREATE TABLE IF
// NOT EXISTS — so a deployment that kept it would eventually get a silent no-op
// followed by a store reading columns that are not there. This renders
// ddb_issue_reports; see ddbissuereports.TablePrefix.
//
// Both foreign keys are re-created, and neither is something platform could
// ship: it does not know which of a consumer's tables holds a principal, and it
// does not know that this consumer's tenants are rows in a table at all.
//
// The reporter key is the one that matters. It is what keeps the single identity
// eraser in internal/build/dataprivacy covering issue reports — the details are
// free text somebody typed, so a report that outlived its reporter would be
// personal data no erasure reaches. Without it this domain would need an eraser
// of its own; platform ships one (issuereports/privacy) for consumers whose
// reporters are not rows they own.
//
// The scope key re-creates what belongs_to_account did: a hard-deleted account
// takes its reports with it rather than stranding them in a scope nothing can
// list. It is safe only because this application never files a report in the
// global scope, whose stored identifier is the empty string and would match no
// account — see ddbissuereports.Scope, which maps an empty account to the scope
// the store refuses rather than to the global one.
func renderIssueReportsDDL() (string, error) {
	schema, err := issuereportsmigrations.SQL(dialect.Postgres, ddbissuereports.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering issue reports migration")
	}

	table := ddl.Qualify(ddbissuereports.TablePrefix) + "issue_reports"

	body := &strings.Builder{}
	body.WriteString("DROP TABLE IF EXISTS issue_reports;\n\n")
	body.WriteString(schema)
	body.WriteString("\n\nALTER TABLE " + table + "\n\tADD CONSTRAINT " + table + "_reporter_fk\n\tFOREIGN KEY (reporter) REFERENCES users(id) ON DELETE CASCADE;\n")
	body.WriteString("\nALTER TABLE " + table + "\n\tADD CONSTRAINT " + table + "_scope_fk\n\tFOREIGN KEY (scope) REFERENCES accounts(id) ON DELETE CASCADE;\n")

	return body.String(), nil
}

// renderWaitlistsDDL renders the two waitlist tables, dropping the ones
// 00008_waitlists.sql created first.
//
// Nothing is carried across, and the two schemas could not carry it anyway. The
// platform's list names the closing time closes_at rather than valid_until and
// adds the tenancy column every one of its reads filters on. Its signup replaces
// belongs_to_user/belongs_to_account with a subject pair, and adds the three
// columns this package was adopted for — contact, contact_digest and status,
// which are what turn a pile of opt-ins into a queue somebody can work and a
// withdrawal somebody can rely on.
//
// The old tables are dropped rather than left in place because their names are
// the ones the platform's default prefix would render, and its DDL says CREATE
// TABLE IF NOT EXISTS — so a deployment that kept them would get a silent no-op
// followed by a store reading columns that are not there. This renders
// ddb_waitlists and ddb_waitlist_signups; see ddbwaitlists.TablePrefix.
//
// # No foreign key is re-created, and that is not an oversight
//
// The old waitlist_signups carried belongs_to_user REFERENCES users ON DELETE
// CASCADE, which is what kept the single identity eraser covering signups. The
// new table cannot carry its equivalent, and the reason is the feature this
// adoption was for: a withdrawal blanks subject_id to the empty string, which is
// the column's NOT NULL default and names no user. A foreign key there would
// refuse every withdrawal — turning the one write somebody has a right to demand
// into a constraint violation.
//
// What replaces the cascade is the waitlists eraser, registered in
// internal/build/dataprivacy. See internal/domain/waitlists/privacy for what it
// erases and what it deliberately keeps.
func renderWaitlistsDDL() (string, error) {
	schema, err := waitlistsmigrations.SQL(dialect.Postgres, ddbwaitlists.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering waitlists migration")
	}

	// Signups first: the old signup table references the old list table, and
	// Postgres will not drop a table out from under a foreign key.
	return "DROP TABLE IF EXISTS waitlist_signups;\nDROP TABLE IF EXISTS waitlists;\n\n" + schema, nil
}

// renderSettingsDDL renders the three settings tables, dropping the two
// 00005_settings.sql created first along with the enum it defined, and
// re-creating the foreign key that kept an erased user's settings from
// outliving them.
//
// Nothing is carried across, and the schemas could not carry it anyway. The
// platform's definition names the value's data type `kind` — string, boolean,
// integer, float — where the old `type` named the sort of principal a setting
// was for, and it holds the enumeration as rows in a child table rather than as
// a pipe-delimited string in a column. Its value replaces
// belongs_to_user/belongs_to_account with a subject pair, and adds the tenancy
// column every one of its reads filters on.
//
// The old tables are dropped rather than left in place because nothing reads
// them once the store is platform's, and the seeded setting they held is
// re-seeded below against the new schema. The enum type goes with them: nothing
// else in this schema uses setting_type, and a type left behind by the table
// that defined it is a name the next migration has to work around.
//
// # The foreign key, and what it is holding
//
// ddb_settings_values.subject_id names a user in every row this application
// writes — see internal/domain/settings for why there is only one subject type —
// so the key belongs_to_user carried is re-creatable, and it is what keeps the
// single identity eraser in internal/build/dataprivacy covering settings. A
// preference somebody chose is a fact about them, and a value that outlived its
// user would be personal data no erasure reaches.
//
// It is also the constraint that enforces the one-subject-type decision rather
// than leaving it to convention: a write naming an account would be refused by
// the database, not merely discouraged by a comment. An account-owned setting
// therefore starts with dropping this key and deciding what erases the rows it
// was holding.
//
// Platform cannot ship it. It does not know which of a consumer's tables holds a
// principal, and it does not know that a consumer has narrowed the subject types
// its schema admits to one.
func renderSettingsDDL() (string, error) {
	schema, err := settingsmigrations.SQL(dialect.Postgres, ddbsettings.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering settings migration")
	}

	qualified := ddl.Qualify(ddbsettings.TablePrefix)
	values := qualified + "settings_values"
	definitions := qualified + "settings_definitions"
	options := qualified + "settings_definition_options"

	body := &strings.Builder{}

	// Configurations first: the old configuration table references the old
	// settings table, and Postgres will not drop a table out from under a
	// foreign key. The enum follows both, for the same reason.
	body.WriteString("DROP TABLE IF EXISTS service_setting_configurations;\nDROP TABLE IF EXISTS service_settings;\nDROP TYPE IF EXISTS setting_type;\n\n")
	body.WriteString(schema)
	body.WriteString("\n\nALTER TABLE " + values + "\n\tADD CONSTRAINT " + values + "_subject_fk\n\tFOREIGN KEY (subject_id) REFERENCES users(id) ON DELETE CASCADE;\n")

	// The one setting this application ships with, re-seeded against the new
	// schema. 00021_mealplanning.sql wrote it into the table dropped above, and
	// the id is carried across so that a client holding it still resolves.
	//
	// Its kind is `string` rather than the old `user`: what a setting is for is
	// no longer a property of the definition, and what it holds is. The two
	// units become rows in the options table, which is where an enumeration
	// lives now.
	body.WriteString("\nINSERT INTO " + definitions + " (id, scope, name, description, kind, default_value, admin_only)\n")
	body.WriteString("VALUES (\n\t'd6me6i4n9qd3gcf5j1p0',\n\t'',\n\t'user_temperature_unit',\n\t'Preferred unit for displaying temperatures (e.g. oven, storage)',\n\t'string',\n\t'fahrenheit',\n\tFALSE\n);\n")
	body.WriteString("\nINSERT INTO " + options + " (definition_id, value)\nVALUES\n\t('d6me6i4n9qd3gcf5j1p0', 'celsius'),\n\t('d6me6i4n9qd3gcf5j1p0', 'fahrenheit');\n")

	return body.String(), nil
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

// bridgeTablesReferencingUploadedMedia are this application's tables whose
// uploaded_media_id column names a row in the upload registry.
//
// They are listed rather than derived because nothing can derive them: the
// registry has never heard of them, which is the whole point of a package that
// can be adopted by an application whose schema it does not know.
var bridgeTablesReferencingUploadedMedia = []string{
	"user_avatars",
	"recipe_images",
	"meal_images",
	"recipe_step_images",
	"ingredient_media",
	"preparation_media",
}

// renderUploadsRegistryDDL renders the upload registry table, dropping the
// uploaded_media table 00010_uploaded_media.sql created first, and re-pointing
// this application's foreign keys at what replaced it.
//
// Nothing is carried across, and the two tables could not carry it anyway: the
// platform's names the columns differently (object_key, content_type,
// owner_id), adds the tenancy column every one of its reads filters on, and
// adds the size that was actually stored — a number the old table never held
// and that cannot be recovered from a row, only from the bucket.
//
// The MIME type enum goes with the table because the registry stores a content
// type as text. That is the right shape regardless of who owns the table: the
// set of types this application accepts is a rule about what it is willing to
// store, checked before the bytes are written, and expressing it as a column
// domain meant that widening it was an ALTER TYPE in a migration. It is now
// uploadedmedia.IsValidMimeType, which a compiler checks and a test can cover.
//
// The DROP is CASCADE because six of this application's tables reference the
// old one, and Postgres will not drop a table out from under a foreign key.
// CASCADE drops those constraints — not the tables, and not the
// uploaded_media_id columns, which still name exactly what they named before.
// The ALTERs below then re-point them at the new table, so the referential
// integrity the old schema had survives the swap rather than quietly becoming
// six columns of unchecked text.
//
// The owner cascade is re-created for the same reason and is the more important
// of the two. Every read of an upload is answered from its owner, and this
// application's owners are all users, so a deleted user whose rows outlived them
// would leave objects nobody can name and nothing will erase. The platform ships
// no such key — it cannot, because it does not know which of a consumer's tables
// holds a principal — and leaves it to the consumer, which is here. It is what
// keeps the single identity eraser in internal/build/dataprivacy covering
// uploads; adopting a platform store is exactly where that stops being true by
// default.
// The registry has no update, so update.uploaded_media went with the RPC when
// this table was adopted. Removing it used to need a DELETE here against the
// permissions table, because the seed was hand-written SQL and nothing
// reconciled it with the Go declaration. The policy is seeded from
// PlatformPolicy() now and Seed rewrites each role's grants rather than adding
// to them, so a permission dropped in Go is a grant that disappears on the next
// migration. See renderAuthorizationDDL.
func renderUploadsRegistryDDL() (string, error) {
	schema, err := uploadsregistrymigrations.SQL(dialect.Postgres, ddbuploadedmedia.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering uploads registry migration")
	}

	table := ddl.Qualify(ddbuploadedmedia.TablePrefix) + "uploads_objects"

	body := &strings.Builder{}
	body.WriteString("DROP TABLE IF EXISTS uploaded_media CASCADE;\nDROP TYPE IF EXISTS uploaded_media_mime_type;\n\n")

	body.WriteString(schema)
	body.WriteString("\n\nALTER TABLE " + table + "\n\tADD CONSTRAINT " + table + "_owner_fk\n\tFOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;\n")

	for _, bridge := range bridgeTablesReferencingUploadedMedia {
		body.WriteString("\nALTER TABLE " + bridge + "\n\tADD CONSTRAINT " + bridge + "_uploaded_media_fk\n\tFOREIGN KEY (uploaded_media_id) REFERENCES " + table + "(id) ON DELETE CASCADE;\n")
	}

	return body.String(), nil
}
