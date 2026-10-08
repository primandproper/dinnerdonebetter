package migrations

import (
	"context"
	"database/sql"
	"embed"
	stderrors "errors"
	"io/fs"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	auditmigrations "github.com/primandproper/platform-go/v15/audit/migrations"
	oauth2clientsmigrations "github.com/primandproper/platform-go/v15/authentication/oauth2clients/migrations"
	oauth2migrations "github.com/primandproper/platform-go/v15/authentication/oauth2serverstore/migrations"
	passkeysmigrations "github.com/primandproper/platform-go/v15/authentication/passkeys/migrations"
	passwordresetmigrations "github.com/primandproper/platform-go/v15/authentication/passwordreset/migrations"
	signindevicesmigrations "github.com/primandproper/platform-go/v15/authentication/signin/devices/migrations"
	refreshtokensmigrations "github.com/primandproper/platform-go/v15/authentication/signin/refreshtokens/migrations"
	webauthndatabase "github.com/primandproper/platform-go/v15/authentication/webauthnsessions"
	webauthnmigrations "github.com/primandproper/platform-go/v15/authentication/webauthnsessions/migrations"
	billingmigrations "github.com/primandproper/platform-go/v15/billing/migrations"
	commentsmigrations "github.com/primandproper/platform-go/v15/comments/migrations"
	dataprivacymigrations "github.com/primandproper/platform-go/v15/dataprivacy/migrations"
	identitymigrations "github.com/primandproper/platform-go/v15/identity/migrations"
	issuereportsmigrations "github.com/primandproper/platform-go/v15/issuereports/migrations"
	linksmigrations "github.com/primandproper/platform-go/v15/links/database/migrations"
	uploadsregistrymigrations "github.com/primandproper/platform-go/v15/mediaregistry/migrations"
	"github.com/primandproper/platform-go/v15/metering"
	meteringmigrations "github.com/primandproper/platform-go/v15/metering/migrations"
	notificationsmigrations "github.com/primandproper/platform-go/v15/notifications/migrations"
	"github.com/primandproper/platform-go/v15/operations"
	operationsmigrations "github.com/primandproper/platform-go/v15/operations/migrations"
	"github.com/primandproper/platform-go/v15/outbox"
	outboxmigrations "github.com/primandproper/platform-go/v15/outbox/migrations"
	authzdatabase "github.com/primandproper/platform-go/v15/rbac"
	authzmigrations "github.com/primandproper/platform-go/v15/rbac/migrations"
	"github.com/primandproper/platform-go/v15/saga"
	sagamigrations "github.com/primandproper/platform-go/v15/saga/migrations"
	settingsmigrations "github.com/primandproper/platform-go/v15/settings/migrations"
	waitlistsmigrations "github.com/primandproper/platform-go/v15/waitlists/migrations"
	"github.com/primandproper/platform-go/v15/webhooks"
	webhooksmigrations "github.com/primandproper/platform-go/v15/webhooks/migrations"
	"github.com/primandproper/platform-go/v15/workqueue"
	workqueuemigrations "github.com/primandproper/platform-go/v15/workqueue/migrations"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/ddl"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	"github.com/primandproper/primitives-go/v2/database/migrate"
	"github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
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
// filename. The sequence is contiguous and ends with this application's own file; see
// docs/migrations.md for why renumbering it is allowed while nothing is deployed, and what it
// costs a local database.
const (
	// identity is first, and has to be: every other table in this schema that names a
	// user or an account has a foreign key into it, and a foreign key cannot reference a
	// table that does not exist yet.
	identityMigrationVersion        = 1
	outboxMigrationVersion          = 2
	sagaMigrationVersion            = 3
	webhooksMigrationVersion        = 4
	auditMigrationVersion           = 5
	dataPrivacyMigrationVersion     = 6
	meteringMigrationVersion        = 7
	operationsMigrationVersion      = 8
	webauthnMigrationVersion        = 9
	oauth2MigrationVersion          = 10
	passwordResetMigrationVersion   = 11
	workQueueMigrationVersion       = 12
	commentsMigrationVersion        = 13
	uploadsRegistryMigrationVersion = 14
	issueReportsMigrationVersion    = 15
	waitlistsMigrationVersion       = 16
	settingsMigrationVersion        = 17
	authorizationMigrationVersion   = 18
	billingMigrationVersion         = 19
	notificationsMigrationVersion   = 20
	oauth2ClientsMigrationVersion   = 21
	refreshTokensMigrationVersion   = 22
	passkeysMigrationVersion        = 23
	actionLinksMigrationVersion     = 24
	signInDevicesMigrationVersion   = 25
	// 26 is migration_files/00026_dinnerdonebetter.sql, and it is last on purpose: its foreign
	// keys name the platform tables above, so it runs once every one of them exists. A platform
	// table adopted later goes before it, and the file moves up a number.
)

// The identity tables other schemas reference.
//
// They are derived from the prefix the identity store is built with rather than spelled,
// because a constraint naming a table the store does not write is a constraint against an
// empty table — which is exactly the failure this application spent a suite discovering.
var (
	identityUsers           = ddl.Qualify(branding.TablePrefix) + "identity_users"
	identityAccounts        = ddl.Qualify(branding.TablePrefix) + "identity_accounts"
	identityUserRoles       = ddl.Qualify(branding.TablePrefix) + "identity_user_roles"
	identityMembershipRoles = ddl.Qualify(branding.TablePrefix) + "identity_membership_roles"
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
	dataPrivacyDDL, err := dataprivacymigrations.SQL(dialect.Postgres, branding.TablePrefix)
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
	oauth2DDL, err := oauth2migrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "rendering oauth2 server migration")
	}

	// The registry, under the same namespace as the four protocol tables above, so one
	// application's oauth2 tables sort together in a database that may hold another's.
	oauth2ClientsDDL, err := renderOAuth2ClientsDDL()
	if err != nil {
		return nil, err
	}

	// Users, accounts, memberships and invitations. platform names its tables
	// identity_*, so these stand beside this application's own while the adoption
	// happens rather than colliding with them — which is what lets the port be built
	// before the switch is thrown.
	identityDDL, err := identitymigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "rendering identity schema")
	}

	// The passkey credentials, under the same namespace. platform names its table
	// webauthn_credentials too, which is why this one carries a prefix and the schema this
	// replaced did not: two tables of one name, and only one of them has a store.
	passkeysDDL, err := renderPasskeysDDL()
	if err != nil {
		return nil, err
	}

	passwordResetDDL, err := renderPasswordResetDDL()
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

	billingDDL, err := renderBillingDDL()
	if err != nil {
		return nil, err
	}

	notificationsDDL, err := renderNotificationsDDL()
	if err != nil {
		return nil, err
	}

	refreshTokensDDL, err := renderRefreshTokensDDL()
	if err != nil {
		return nil, err
	}

	actionLinksDDL, err := renderActionLinksDDL()
	if err != nil {
		return nil, err
	}

	signInDevicesDDL, err := renderSignInDevicesDDL()
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
		migrate.WithGeneratedMigration(workQueueMigrationVersion, "create_work_queue_items_table", workQueueDDL),
		migrate.WithGeneratedMigration(commentsMigrationVersion, "create_comments_table", commentsDDL),
		migrate.WithGeneratedMigration(uploadsRegistryMigrationVersion, "create_uploads_objects_table", uploadsRegistryDDL),
		migrate.WithGeneratedMigration(issueReportsMigrationVersion, "create_issue_reports_table", issueReportsDDL),
		migrate.WithGeneratedMigration(waitlistsMigrationVersion, "create_waitlist_tables", waitlistsDDL),
		migrate.WithGeneratedMigration(settingsMigrationVersion, "create_settings_tables", settingsDDL),
		migrate.WithGeneratedMigration(authorizationMigrationVersion, "create_authorization_tables", authorizationDDL),
		migrate.WithGeneratedMigration(billingMigrationVersion, "create_billing_tables", billingDDL),
		migrate.WithGeneratedMigration(notificationsMigrationVersion, "create_notifications_tables", notificationsDDL),
		migrate.WithGeneratedMigration(oauth2ClientsMigrationVersion, "adopt_oauth2_registered_clients", oauth2ClientsDDL),
		migrate.WithGeneratedMigration(refreshTokensMigrationVersion, "create_signin_refresh_tokens_table", refreshTokensDDL),
		migrate.WithGeneratedMigration(identityMigrationVersion, "create_identity_tables", identityDDL),
		migrate.WithGeneratedMigration(passkeysMigrationVersion, "create_passkey_credentials_table", passkeysDDL),
		migrate.WithGeneratedMigration(actionLinksMigrationVersion, "create_action_links_table", actionLinksDDL),
		migrate.WithGeneratedMigration(signInDevicesMigrationVersion, "create_signin_devices_table", signInDevicesDDL),
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
		// Rollback after a commit reports ErrTxDone and means the commit won, so
		// only anything else is worth saying. There is nothing to return it to
		// from a defer, and the seeding error is the one the caller wants.
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !stderrors.Is(rollbackErr, sql.ErrTxDone) {
			m.logger.Error("rolling back authorization policy transaction", rollbackErr)
		}
	}()

	// Every replica migrates and then seeds at once, and the migrator's lock
	// covers only the schema. Nothing serializes the seeds, and nothing needs to:
	// two seeds of one policy converge on it rather than colliding (see rbac's
	// Resolver.Seed).
	// Built against the transaction so the reads Seed makes to resolve role and
	// permission ids see the rows it has just written.
	resolver, err := authzdatabase.NewResolver(
		&authzdatabase.Config{Dialect: dialect.Postgres, TablePrefix: branding.TablePrefix},
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
// the grants between them, and the inheritance edges — and the foreign keys
// the platform cannot ship.
//
// The tables hold no hand-written policy. The permissions a role holds are
// declared once, as the slices in internal/authorization, and seeded from
// PlatformPolicy() — see Migrator.Migrate. A seed written as INSERT statements
// beside them would be a second declaration, and the two would drift.
//
// The foreign keys target ddb_authz_roles(name) rather than its primary key.
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
// behavior we want.
func renderAuthorizationDDL() (string, error) {
	schema, err := authzmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering authorization migration")
	}

	rolesTable := ddl.Qualify(branding.TablePrefix) + "authz_roles"

	body := &strings.Builder{}
	body.WriteString(schema)
	// The key follows the assignments, which are platform's: a service role is a row in
	// identity_user_roles and a membership role is one in identity_membership_roles. Two
	// constraints rather than one, because platform keeps the two apart — they are granted
	// by different people and answer different questions.
	body.WriteString("\n\nALTER TABLE " + identityUserRoles + "\n\tADD CONSTRAINT " + identityUserRoles + "_role_fk\n\tFOREIGN KEY (role) REFERENCES " + rolesTable + "(name) ON DELETE RESTRICT;\n")
	body.WriteString("\nALTER TABLE " + identityMembershipRoles + "\n\tADD CONSTRAINT " + identityMembershipRoles + "_role_fk\n\tFOREIGN KEY (role) REFERENCES " + rolesTable + "(name) ON DELETE RESTRICT;\n")

	return body.String(), nil
}

// billingTablesOwnedByAccounts are the three billing tables whose rows belong to
// an account. The catalog is the fourth and belongs to nobody.
var billingTablesOwnedByAccounts = []string{
	"billing_subscriptions",
	"billing_purchases",
	"billing_transactions",
}

// renderBillingDDL renders the four billing tables — the catalog, the
// subscriptions, the one-time purchases and the ledger — and the foreign key
// that keeps an erased account's billing from outliving it. The subscription
// status is capitalism's vocabulary, which spells it "canceled".
//
// # The foreign key, and what it decides
//
// belongs_to_account names an account in every row of the three account-owned
// tables, and each gets a key to the accounts table with ON DELETE CASCADE. It is
// what keeps the single identity eraser in internal/build/dataprivacy covering
// billing: an erased user's accounts take their subscriptions, purchases and
// ledger rows with them.
//
// That cascade is preserved behavior rather than a decision. Platform's
// billing/privacy ships a collector and deliberately no eraser, on the grounds
// that financial records carry a statutory retention that outranks a right to
// erasure — and docs/data-privacy.md names payments as the likeliest first domain
// to need retention rather than a cascade. Making that call means dropping
// this key and registering an eraser that anonymizes rather than deletes, which is
// a policy decision this migration does not take on anybody's behalf.
//
// Platform cannot ship the key either way. It does not know that a consumer's
// accounts are rows in a table at all.
func renderBillingDDL() (string, error) {
	schema, err := billingmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering billing migration")
	}

	qualified := ddl.Qualify(branding.TablePrefix)

	body := &strings.Builder{}
	body.WriteString(schema)

	for _, owned := range billingTablesOwnedByAccounts {
		table := qualified + owned
		body.WriteString("\n\nALTER TABLE " + table + "\n\tADD CONSTRAINT " + table + "_account_fk\n\tFOREIGN KEY (belongs_to_account) REFERENCES " + identityAccounts + "(id) ON DELETE CASCADE;\n")
	}

	return body.String(), nil
}

// renderIssueReportsDDL renders the issue report table and the reporter's
// foreign key, which is something platform could not ship: it does not know which
// of a consumer's tables holds a principal.
//
// The key is what keeps the single identity eraser in internal/build/dataprivacy
// covering issue reports — the details are free text somebody typed, so a report
// that outlived its reporter would be personal data no erasure reaches. Without it
// this domain would rely on the eraser platform ships (issuereports/privacy) alone.
//
// The scope column carries no key. Every report is filed under the global scope
// (see ddbissuereports.Scope), whose stored identifier is the empty string and
// names no account, so a key to the accounts table would refuse every report
// anybody filed. A report is the reporter's and the service's administrators',
// not a household's.
func renderIssueReportsDDL() (string, error) {
	schema, err := issuereportsmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering issue reports migration")
	}

	table := ddl.Qualify(branding.TablePrefix) + "issue_reports"

	body := &strings.Builder{}
	body.WriteString(schema)
	body.WriteString("\n\nALTER TABLE " + table + "\n\tADD CONSTRAINT " + table + "_reporter_fk\n\tFOREIGN KEY (reporter) REFERENCES " + identityUsers + "(id) ON DELETE CASCADE;\n")

	return body.String(), nil
}

// renderWaitlistsDDL renders the two waitlist tables.
//
// # No foreign key, and that is not an oversight
//
// A signup's subject_id names a user, and a key to the users table with ON DELETE
// CASCADE would be what kept the single identity eraser covering signups. It
// cannot be added, and the reason is the feature this store was adopted for: a
// withdrawal blanks subject_id to the empty string, which is the column's NOT NULL
// default and names no user. A foreign key there would refuse every withdrawal —
// turning the one write somebody has a right to demand into a constraint violation.
//
// What does the cascade's job is the waitlists eraser, registered in
// internal/build/dataprivacy. See internal/domain/waitlists/privacy for what it
// erases and what it deliberately keeps.
func renderWaitlistsDDL() (string, error) {
	schema, err := waitlistsmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering waitlists migration")
	}

	return schema, nil
}

// renderSettingsDDL renders the three settings tables, the foreign key that keeps
// an erased user's settings from outliving them, and the one setting this
// application ships with.
//
// # The foreign key, and what it is holding
//
// ddb_settings_values.subject_id names a user in every row this application
// writes — see internal/domain/settings for why there is only one subject type —
// so it can carry a key to the users table, and that key is what keeps the
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
	schema, err := settingsmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering settings migration")
	}

	qualified := ddl.Qualify(branding.TablePrefix)
	values := qualified + "settings_values"
	definitions := qualified + "settings_definitions"
	options := qualified + "settings_definition_options"

	body := &strings.Builder{}

	body.WriteString(schema)
	body.WriteString("\n\nALTER TABLE " + values + "\n\tADD CONSTRAINT " + values + "_subject_fk\n\tFOREIGN KEY (subject_id) REFERENCES " + identityUsers + "(id) ON DELETE CASCADE;\n")

	// The one setting this application ships with. Its kind is what it holds, a
	// string, and the two units it may hold are rows in the options table, which
	// is where an enumeration lives.
	body.WriteString("\nINSERT INTO " + definitions + " (id, scope, name, description, kind, default_value, admin_only)\n")
	body.WriteString("VALUES (\n\t'd6me6i4n9qd3gcf5j1p0',\n\t'',\n\t'user_temperature_unit',\n\t'Preferred unit for displaying temperatures (e.g. oven, storage)',\n\t'string',\n\t'fahrenheit',\n\tFALSE\n);\n")
	body.WriteString("\nINSERT INTO " + options + " (definition_id, value)\nVALUES\n\t('d6me6i4n9qd3gcf5j1p0', 'celsius'),\n\t('d6me6i4n9qd3gcf5j1p0', 'fahrenheit');\n")

	return body.String(), nil
}

// renderCommentsDDL renders the comment table.
//
// It carries no foreign keys, by construction rather than by choice: a comment's
// target lives in a table the platform's store has never seen, so there is no
// column it could point at. What erases an author's comments is the comments
// eraser, registered in internal/build/dataprivacy; what removes a deleted
// target's comments is nothing, which is the ruling platform's package
// documentation states plainly.
//
// target_type is text rather than an enum, so adding a target type is a line in
// the catalog that a compiler checks rather than an ALTER TYPE in a migration.
func renderCommentsDDL() (string, error) {
	schema, err := commentsmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering comments migration")
	}

	return schema, nil
}

// userCascade re-creates the foreign key an adopted table loses by being adopted.
//
// Every one of platform's schemas stores a user id as a bare column, and has to: the module
// is multi-engine and does not know what a consumer calls its user table, or whether it has
// one. The consumer does. Two tables here name a user and nothing was pointing them at it —
// which is not merely a missing constraint. This application erases a subject by deleting
// the user row and letting the cascade reach everything that names it, so a table with no
// key is a table erasure does not reach: a deleted user's passkeys and password reset tokens
// outlived them, silently, because a cascade that does not exist raises nothing.
//
// The registry of OAuth2 clients looks like a third and is not; renderOAuth2ClientsDDL says
// why, and what erases one instead.
//
// This is the same statement uploads/registry, issuereports and notifications each write by
// hand; it is a function because there are now five of them and the argument is identical.
func userCascade(table, column string) string {
	return "\n\nALTER TABLE " + table +
		"\n\tADD CONSTRAINT " + table + "_" + column + "_fk" +
		"\n\tFOREIGN KEY (" + column + ") REFERENCES " + identityUsers + "(id) ON DELETE CASCADE;\n"
}

// renderPasskeysDDL renders the passkey credential table, keyed to the user it belongs to.
//
// See userCascade: the key is this repository's to add, and this table is the one that had
// one before the identity adoption took it away.
func renderPasskeysDDL() (string, error) {
	schema, err := passkeysmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering passkey credentials schema")
	}

	table := ddl.Qualify(branding.TablePrefix) + "webauthn_credentials"

	body := &strings.Builder{}

	body.WriteString(schema)
	body.WriteString(userCascade(table, "belongs_to_user"))

	return body.String(), nil
}

// renderOAuth2ClientsDDL renders the administered client registry.
//
// It is the one table here that names a user and does not get a key, and the reason is that
// in this deployment the column is usually not a user. A client is minted by an operator to
// let an application speak for the service on behalf of whoever signs in, so it is
// registered unowned — CreateOAuth2ClientForService passes "" for the owner, and
// oauth2clients.Client.Admits permits any subject for exactly that arrangement. A foreign
// key would refuse every one of them.
//
// A user-owned client is still possible, which is why the erasure of one is a registered
// eraser rather than a cascade; see internal/build/dataprivacy.
//
// The authorization server's own tables are not keyed either, for a different reason.
// oauth2_clients has no user column at all, and the codes and tokens name a subject_id that
// is a claim rather than a row reference — they are short-lived grants that expire on their
// own, and a key onto them would make issuing one depend on the directory.
func renderOAuth2ClientsDDL() (string, error) {
	schema, err := oauth2clientsmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering oauth2 registered clients schema")
	}

	return schema, nil
}

// renderPasswordResetDDL renders the password reset token table, keyed to the
// user each token was issued to.
//
// See userCascade: a reset token is a credential, and one that outlived the
// erasure of the person it resets would be personal data no erasure reaches.
func renderPasswordResetDDL() (string, error) {
	schema, err := passwordresetmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering password reset token migration")
	}

	table := ddl.Qualify(branding.TablePrefix) + "password_reset_tokens"

	body := &strings.Builder{}

	body.WriteString(schema)
	body.WriteString(userCascade(table, "belongs_to_user"))

	return body.String(), nil
}

// renderRefreshTokensDDL renders the table platform's sign-in keeps its rotating refresh
// tokens in, keyed to the user each one was issued to.
//
// See userCascade: a refresh token is a credential, and one that outlived the erasure of
// the person it signs in would be the same leftover the other credential tables had. The
// column is subject_id, which in this deployment is always a user — a refresh token is
// minted only by a password sign-in, and nothing signs in but a person.
//
// The prefix is branding.TablePrefix, so this renders ddb_signin_refresh_tokens, beside the
// other two tables this application's sign-in writes.
func renderRefreshTokensDDL() (string, error) {
	schema, err := refreshtokensmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering refresh token migration")
	}

	table := ddl.Qualify(branding.TablePrefix) + "signin_refresh_tokens"

	body := &strings.Builder{}

	body.WriteString(schema)
	body.WriteString(userCascade(table, "subject_id"))

	return body.String(), nil
}

// renderSignInDevicesDDL renders the table "where you're signed in" reads each login's device
// from: one row per login, recorded by platform's sign-in device hooks when a token is minted
// and deleted when the login is ended.
//
// See userCascade: the rows are an address and a browser somebody signed in from, which is
// personal data, and platform's table names its user by a bare column. The privacy adapter
// registers an eraser for them too — see internal/build/dataprivacy — so the key is the second
// of two things that reach them rather than the only one.
//
// The prefix is branding.TablePrefix, so this renders ddb_signin_devices beside the refresh
// token table whose families it is keyed on.
func renderSignInDevicesDDL() (string, error) {
	schema, err := signindevicesmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering sign-in device migration")
	}

	table := ddl.Qualify(branding.TablePrefix) + "signin_devices"

	body := &strings.Builder{}

	body.WriteString(schema)
	body.WriteString(userCascade(table, "user_id"))

	return body.String(), nil
}

// renderActionLinksDDL renders the action-link table the waitlist confirmation loop mints
// into: a confirmation link and an unsubscribe link for every pending signup.
//
// It renders platform's schema and nothing else. A link's subject is the signup it was minted
// for, and the signup table carries no foreign key a link could follow (see
// renderWaitlistsDDL), so there is no cascade to add: an erased signup's links name a row that
// no longer answers, and a link whose signup is gone is refused by the surface that redeems it
// rather than by the table. What reclaims the rows is the store's sweeper, run in the API
// server — see internal/build/waitlists.
//
// The prefix is waitlists', so this renders ddb_action_links beside the tables whose links it
// holds. A second action that is not a waitlist's would still belong in it: the table is keyed
// by action, and the minter is one registry.
func renderActionLinksDDL() (string, error) {
	schema, err := linksmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering action link migration")
	}

	return schema, nil
}

// renderWebAuthnDDL renders the passkey ceremony session table. It carries no
// key: a row is one ceremony in flight, and a ceremony lasts a minute.
func renderWebAuthnDDL() (string, error) {
	schema, err := webauthnmigrations.SQL(dialect.Postgres, webauthndatabase.DefaultTablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering webauthn session migration")
	}

	return schema, nil
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
	schema, err := auditmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering audit migration")
	}

	appendOnly, err := auditmigrations.AppendOnlyStatements(dialect.Postgres, branding.TablePrefix)
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

// renderUploadsRegistryDDL renders the upload registry table and the owner's
// foreign key.
//
// A content type is text rather than an enum. The set of types this application
// accepts is a rule about what it is willing to store, checked before the bytes
// are written — uploadedmedia.IsValidMimeType, which a compiler checks and a test
// can cover — rather than a column domain that is widened by an ALTER TYPE.
//
// The owner key is the one that matters. Every read of an upload is answered from
// its owner, and this application's owners are all users, so a deleted user whose
// rows outlived them would leave objects nobody can name and nothing will erase.
// The platform ships no such key — it cannot, because it does not know which of a
// consumer's tables holds a principal — and leaves it to the consumer, which is
// here. It is what keeps the single identity eraser in internal/build/dataprivacy
// covering uploads.
//
// This application's own tables that name an upload carry their keys to this one
// in 00026_dinnerdonebetter.sql; see the version constants for why it runs last.
func renderUploadsRegistryDDL() (string, error) {
	schema, err := uploadsregistrymigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering uploads registry migration")
	}

	table := ddl.Qualify(branding.TablePrefix) + "uploads_objects"

	body := &strings.Builder{}

	body.WriteString(schema)
	body.WriteString("\n\nALTER TABLE " + table + "\n\tADD CONSTRAINT " + table + "_owner_fk\n\tFOREIGN KEY (owner_id) REFERENCES " + identityUsers + "(id) ON DELETE CASCADE;\n")

	return body.String(), nil
}

// renderNotificationsDDL renders the inbox and the device registry, and the two
// foreign keys that tie each row to the user it belongs to.
//
// A device has no archived dimension: it is revoked by removal rather than
// retired in place — see the two file comments in platform's notifications/grpc.
//
// # The two foreign keys
//
// Both principal columns name a user in every row this application writes: an
// inbox belongs to a person and so does a handset, and this deployment has no
// other kind of principal. So each can carry a key to the users table, and that
// is what keeps the single identity eraser in internal/build/dataprivacy
// covering both tables.
//
// platform's notifications/privacy erasers are registered as well, and the two
// are not alternatives here. A deployment with non-user principals has only the
// erasers; this one keeps the keys too, because a cascade is exactly what goes
// missing silently in an adoption, and an eraser that deletes rows the cascade
// already took costs one statement that finds nothing.
//
// It is also what stops the subject type widening by accident. A notification
// addressed to an account would be refused by the database rather than
// discouraged by a comment, so widening starts by dropping these keys and
// deciding what erases the rows they were holding.
func renderNotificationsDDL() (string, error) {
	schema, err := notificationsmigrations.SQL(dialect.Postgres, branding.TablePrefix)
	if err != nil {
		return "", errors.Wrap(err, "rendering notifications migration")
	}

	qualified := ddl.Qualify(branding.TablePrefix)
	inbox := qualified + "notifications_inbox"
	devices := qualified + "notifications_devices"

	body := &strings.Builder{}

	body.WriteString(schema)
	body.WriteString("\n\nALTER TABLE " + inbox + "\n\tADD CONSTRAINT " + inbox + "_principal_fk\n\tFOREIGN KEY (principal) REFERENCES " + identityUsers + "(id) ON DELETE CASCADE;\n")
	body.WriteString("\nALTER TABLE " + devices + "\n\tADD CONSTRAINT " + devices + "_principal_fk\n\tFOREIGN KEY (principal) REFERENCES " + identityUsers + "(id) ON DELETE CASCADE;\n")

	return body.String(), nil
}
