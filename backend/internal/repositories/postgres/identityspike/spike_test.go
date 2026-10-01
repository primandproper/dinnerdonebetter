// Package identityspike is a pass-2 spike, not production code.
//
// It answers one question about adopting platform-go v14's identity package:
// whether this application's audit entry and data change event can be written
// inside the transaction an identity operation runs in.
//
// It matters because identity adopts through a different seam than comments
// did. A comment write goes through comments.Store, so this repo decorates the
// store and its two extra statements land in the server's transaction — see
// internal/repositories/postgres/comments/server_transaction_test.go. An
// identity write goes through identity.Service, which orchestrates several
// store calls in one transaction of its own, so decorating the store would
// record one row of an operation that writes three. platform's answer is
// identity.Hooks: one method per operation, each handed the operation's
// database.Tx.
//
// This package builds that seam against a real database and proves it commits
// and rolls back as one. It deliberately does not port anything: the identity
// repository here is 24,000 lines and the adoption is a separate piece of work.
package identityspike

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v14/identity"
	identitymigrations "github.com/primandproper/platform-go/v14/identity/migrations"
	"github.com/primandproper/platform-go/v14/outbox"
	outboxmigrations "github.com/primandproper/platform-go/v14/outbox/migrations"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tablePrefix is what this application would namespace platform's identity
// tables with, matching every other adopted store here.
const tablePrefix = "ddb"

// TestMain migrates a template database carrying platform's identity DDL
// alongside this application's own.
//
// Both, because the spike's question is whether the two can share a
// transaction, and that is only a real question when they are in one database.
func TestMain(m *testing.M) {
	os.Exit(pgtesting.RunTestsWithSharedDatabase(m, func(ctx context.Context, db *sql.DB) error {
		migrator, err := migrations.NewMigrator(loggingnoop.NewLogger())
		if err != nil {
			return err
		}

		if err = migrator.Migrate(ctx, db); err != nil {
			return err
		}

		// platform's identity tables, which this application's migrations do not
		// create today: adopting identity is what would add them, and the DDL is
		// rendered from the platform rather than copied.
		identityDDL, err := identitymigrations.SQL(dialect.Postgres, tablePrefix)
		if err != nil {
			return err
		}

		if _, err = db.ExecContext(ctx, identityDDL); err != nil {
			return err
		}

		// The outbox is already in this application's migrations, but rendering it
		// here as well keeps the spike readable about what it depends on.
		_ = outboxmigrations.SQL

		return nil
	}))
}

// recordingHooks is the shape this application's audit layer would take.
//
// It embeds identity.NoopHooks, which is the adoption platform documents: only
// the operations that need recording are overridden, and an operation added to
// the interface later does not break this type.
//
// The two statements it writes stand in for the real ones — an audit entry and
// an outbox row. What is under test is not their content but where they land,
// so they are the simplest writes that a rollback would be visible in.
type recordingHooks struct {
	identity.NoopHooks

	writer *outbox.Writer

	// failWith, when set, is returned instead of recording. It is how the test
	// induces the failure whose blast radius is the whole point.
	failWith error
}

func (h *recordingHooks) AfterRegister(
	ctx context.Context,
	tx database.Tx,
	scope tenancy.Scope,
	registration *identity.Registration,
) error {
	if h.failWith != nil {
		return h.failWith
	}

	// A row of this application's own, in the operation's transaction, keyed to
	// the user the operation just wrote. This is the audit entry's stand-in.
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO spike_identity_records (id, user_id, event) VALUES ($1, $2, $3)`,
		identifiers.New(), registration.User.ID, "registered"); err != nil {
		return err
	}

	// And the event, through platform's outbox writer on the same executor,
	// which is the real machinery rather than a stand-in.
	return h.writer.Enqueue(ctx, tx, outbox.Message{
		Topic:   "data_changes",
		Payload: map[string]string{"event": "user_registered", "userID": registration.User.ID},
	})
}

type fixture struct {
	service *identity.Service
	store   identity.Store
	hooks   *recordingHooks
	db      database.Client
}

func buildFixture(t *testing.T) *fixture {
	t.Helper()

	ctx := t.Context()

	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	db, err := postgres.NewDatabaseClient(ctx, config,
		postgres.WithLogger(loggingnoop.NewLogger()),
		postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NoError(t, err)

	// The side table this application's recording writes into. A real adoption
	// would write the audit log and the outbox; this is the same shape with the
	// columns the assertions need.
	_, err = db.Writer().ExecContext(ctx,
		`CREATE TABLE IF NOT EXISTS spike_identity_records (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			event TEXT NOT NULL
		)`)
	require.NoError(t, err)

	store, err := identity.NewSQLStore(db, identity.WithTablePrefix(tablePrefix))
	require.NoError(t, err)

	writer, err := outbox.NewWriter(dialect.Postgres,
		outbox.WithWriterLogger(loggingnoop.NewLogger()))
	require.NoError(t, err)

	hooks := &recordingHooks{writer: writer}

	service, err := identity.NewService(db, store,
		identity.WithHooks(hooks),
		identity.WithServiceLogger(loggingnoop.NewLogger()))
	require.NoError(t, err)

	return &fixture{service: service, store: store, hooks: hooks, db: db}
}

func (f *fixture) count(t *testing.T, ctx context.Context, query string, args ...any) int {
	t.Helper()

	var n int
	require.NoError(t, f.db.Reader().QueryRowContext(ctx, query, args...).Scan(&n))

	return n
}

func (f *fixture) userRows(t *testing.T, ctx context.Context, username string) int {
	t.Helper()

	return f.count(t, ctx, `SELECT COUNT(*) FROM `+tablePrefix+`_identity_users WHERE username = $1`, username)
}

func (f *fixture) recordRows(t *testing.T, ctx context.Context) int {
	t.Helper()

	return f.count(t, ctx, `SELECT COUNT(*) FROM spike_identity_records`)
}

func (f *fixture) outboxRows(t *testing.T, ctx context.Context) int {
	t.Helper()

	return f.count(t, ctx, `SELECT COUNT(*) FROM outbox_messages`)
}

func newRegistration() (*identity.User, *identity.Account) {
	username := "spike_" + identifiers.New()

	return &identity.User{
		ID:            identifiers.New(),
		Username:      username,
		EmailAddress:  username + "@example.com",
		AccountStatus: identity.StatusUnverified,
	}, &identity.Account{
		ID:   identifiers.New(),
		Name: "the " + username + " household",
	}
}

// TestHooks_CommitsWithTheOperation is the happy half: a registration writes a
// user, an account and a membership, and this application's two statements land
// in the same transaction.
func TestHooks_CommitsWithTheOperation(T *testing.T) {
	T.Parallel()

	T.Run("the registration and the recording land together", func(t *testing.T) {
		t.Parallel()

		fixture := buildFixture(t)
		ctx := t.Context()

		require.Zero(t, fixture.recordRows(t, ctx))
		require.Zero(t, fixture.outboxRows(t, ctx))

		user, account := newRegistration()

		registration, err := fixture.service.Register(ctx, tenancy.Global(), user, account, []string{"account_admin"})
		require.NoError(t, err)
		require.NotNil(t, registration)

		assert.Equal(t, 1, fixture.userRows(t, ctx, user.Username))
		assert.Equal(t, 1, fixture.recordRows(t, ctx), "the hook's row should have committed with the registration")
		assert.Equal(t, 1, fixture.outboxRows(t, ctx), "the hook's event should have committed with the registration")
	})
}

// TestHooks_RollBackWithTheOperation is the load-bearing half, and the identity
// analog of the comments test.
//
// The hook fails. The registration's own writes — the user, the account and the
// owner membership — have already run on the transaction by then, and all of
// them must be gone afterwards. If they are not, then adopting identity means
// an application that can register a user it has no record of registering.
func TestHooks_RollBackWithTheOperation(T *testing.T) {
	T.Parallel()

	T.Run("a failed hook takes the user, the account and the membership with it", func(t *testing.T) {
		t.Parallel()

		fixture := buildFixture(t)
		ctx := t.Context()

		errRecordingUnavailable := platformerrors.New("the audit log is unavailable")
		fixture.hooks.failWith = errRecordingUnavailable

		user, account := newRegistration()

		registration, err := fixture.service.Register(ctx, tenancy.Global(), user, account, []string{"account_admin"})
		require.Error(t, err)
		assert.Nil(t, registration)

		assert.Zero(t, fixture.userRows(t, ctx, user.Username),
			"the user should have rolled back with the failed recording")
		assert.Zero(t, fixture.count(t, ctx,
			`SELECT COUNT(*) FROM `+tablePrefix+`_identity_accounts WHERE id = $1`, account.ID),
			"the account should have rolled back with the failed recording")
		assert.Zero(t, fixture.count(t, ctx,
			`SELECT COUNT(*) FROM `+tablePrefix+`_identity_memberships WHERE belongs_to_user = $1`, user.ID),
			"the membership should have rolled back with the failed recording")
		assert.Zero(t, fixture.recordRows(t, ctx))
		assert.Zero(t, fixture.outboxRows(t, ctx))
	})
}
