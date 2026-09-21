package waitlists

import (
	"context"
	"database/sql"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbwaitlists "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	waitlists "github.com/primandproper/platform-go/v14/waitlists"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain starts the one postgres container this package's tests share and migrates
// the template database each of them is cloned from, so that a test costs a database
// clone rather than a container start plus a migration replay. See
// pgtesting.RunTestsWithSharedDatabase.
func TestMain(m *testing.M) {
	os.Exit(pgtesting.RunTestsWithSharedDatabase(m, func(ctx context.Context, db *sql.DB) error {
		migrator, err := migrations.NewMigrator(loggingnoop.NewLogger())
		if err != nil {
			return err
		}

		return migrator.Migrate(ctx, db)
	}))
}

// buildDatabaseClientForTest builds the store over a real database.
func buildDatabaseClientForTest(t *testing.T) (waitlists.Store, audit.Repository, database.Client) {
	t.Helper()

	ctx := t.Context()

	// Already migrated: the template this was cloned from was migrated once in TestMain.
	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), pgc)
	require.NoError(t, err)

	c, err := ProvideWaitlistsRepository(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		auditLogEntryRepo,
		pgc,
		nil,
	)
	require.NoError(t, err)

	return c, auditLogEntryRepo, pgc
}

// signatoryForTest creates a user and an account for them, and returns the user.
//
// The signup table has no foreign key to either — a withdrawal blanks the
// subject reference, so it cannot have one — but the audit entry a signup write
// records names the user, and the audit chain does.
func signatoryForTest(t *testing.T, db database.Client) string {
	t.Helper()

	user := pgtesting.CreateUserForTest(t, nil, db.Writer())
	pgtesting.CreateAccountForTest(t, nil, user.ID, db.Writer())

	return user.ID
}

// openListForTest opens one list that is still taking signups.
func openListForTest(t *testing.T, ctx context.Context, dbc waitlists.Store, db database.Client) *waitlists.List {
	t.Helper()

	list, err := writeT(ctx, db, func(tx database.Tx) (*waitlists.List, error) {
		return dbc.CreateList(ctx, tx, ddbwaitlists.Scope(), fakes.BuildFakeWaitlist())
	})
	require.NoError(t, err)

	return list
}

func TestRepository_Integration_Waitlists(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	example := fakes.BuildFakeWaitlist()

	created, err := writeT(ctx, db, func(tx database.Tx) (*waitlists.List, error) {
		return dbc.CreateList(ctx, tx, scope, example)
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.False(t, created.CreatedAt.IsZero())

	// A list belongs to nobody, so its entries are recorded under the
	// unattributed actor — the same shape the table this replaced recorded under,
	// and the reason internal/domain/audit names that actor rather than leaving
	// it blank.
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, audit.UnattributedActorID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeWaitlists, RelevantID: created.ID},
	})

	fetched, err := dbc.GetList(ctx, db.Reader(), scope, created.ID)
	require.NoError(t, err)
	assert.Equal(t, example.Name, fetched.Name)
	assert.Equal(t, example.Description, fetched.Description)

	page, err := dbc.ListLists(ctx, db.Reader(), scope, nil)
	require.NoError(t, err)
	require.Len(t, page.Data, 1)
	assert.Equal(t, created.ID, page.Data[0].ID)

	// An open list is on the open page, which is the read a signup form offers.
	open, err := dbc.ListOpenLists(ctx, db.Reader(), scope, nil)
	require.NoError(t, err)
	require.Len(t, open.Data, 1)

	fetched.Name = "renamed"

	// v14's write methods answer with the row as stored, so the assertions below could read
	// it from here. They re-read instead: what a later request sees is what these tests are
	// about, and a returned struct cannot tell a committed write from an uncommitted one.
	_, err = writeT(ctx, db, func(tx database.Tx) (*waitlists.List, error) {
		return dbc.UpdateList(ctx, tx, scope, fetched)
	})
	require.NoError(t, err)

	updated, err := dbc.GetList(ctx, db.Reader(), scope, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Name)
	assert.NotNil(t, updated.LastUpdatedAt)

	_, err = writeT(ctx, db, func(tx database.Tx) (*waitlists.List, error) {
		return dbc.ArchiveList(ctx, tx, scope, created.ID)
	})
	require.NoError(t, err)

	afterArchive, err := dbc.GetList(ctx, db.Reader(), scope, created.ID)
	require.Error(t, err)
	assert.Nil(t, afterArchive)
	require.ErrorIs(t, err, waitlists.ErrListNotFound)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, audit.UnattributedActorID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeWaitlists, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeWaitlists, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeWaitlists, RelevantID: created.ID},
	})
}

// TestRepository_Integration_ArchivedListTakesNoSignups pins that archiving a
// list closes it immediately, whatever its closing time says.
func TestRepository_Integration_ArchivedListTakesNoSignups(t *testing.T) {
	ctx := t.Context()
	dbc, _, db := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, db)
	list := openListForTest(t, ctx, dbc, db)

	_, err := writeT(ctx, db, func(tx database.Tx) (*waitlists.List, error) {
		return dbc.ArchiveList(ctx, tx, scope, list.ID)
	})
	require.NoError(t, err)

	_, err = joinT(ctx, db, dbc, scope, list.ID, fakes.BuildFakeWaitlistSignupForUser(userID))
	require.Error(t, err)
	// The list is gone as far as the signup path is concerned: the read that
	// decides whether it is open cannot find it.
	require.ErrorIs(t, err, waitlists.ErrListNotFound)
}

// TestRepository_Integration_ClosedListTakesNoSignups pins the other half: a live
// list past its closing time refuses a signup and says why.
func TestRepository_Integration_ClosedListTakesNoSignups(t *testing.T) {
	ctx := t.Context()
	dbc, _, db := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, db)

	closed := fakes.BuildFakeWaitlist()
	closed.ClosesAt = time.Now().Add(-time.Hour).UTC()

	list, err := writeT(ctx, db, func(tx database.Tx) (*waitlists.List, error) {
		return dbc.CreateList(ctx, tx, scope, closed)
	})
	require.NoError(t, err)

	_, err = joinT(ctx, db, dbc, scope, list.ID, fakes.BuildFakeWaitlistSignupForUser(userID))
	require.Error(t, err)
	require.ErrorIs(t, err, waitlists.ErrListClosed)

	// And it is off the open page while still being in the catalog.
	open, err := dbc.ListOpenLists(ctx, db.Reader(), scope, nil)
	require.NoError(t, err)
	assert.Empty(t, open.Data)

	all, err := dbc.ListLists(ctx, db.Reader(), scope, nil)
	require.NoError(t, err)
	assert.Len(t, all.Data, 1)
}

func TestRepository_Integration_WaitlistSignups(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, db)
	list := openListForTest(t, ctx, dbc, db)

	example := fakes.BuildFakeWaitlistSignupForUser(userID)

	joined, err := joinT(ctx, db, dbc, scope, list.ID, example)
	require.NoError(t, err)
	assert.Equal(t, waitlists.StatusWaiting, joined.Status)
	assert.Equal(t, list.ID, joined.ListID)
	assert.NotEmpty(t, joined.ContactDigest)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, userID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeWaitlistSignups, RelevantID: joined.ID},
	})

	fetched, err := dbc.GetSignup(ctx, db.Reader(), scope, list.ID, joined.ID)
	require.NoError(t, err)
	assert.Equal(t, example.Contact, fetched.Contact)
	assert.Equal(t, ddbwaitlists.SubjectFor(userID), fetched.Subject)

	// The address finds the row whichever capitalization the caller has.
	byContact, err := dbc.GetSignupByContact(ctx, db.Reader(), scope, list.ID, strings.ToUpper(example.Contact))
	require.NoError(t, err)
	assert.Equal(t, joined.ID, byContact.ID)

	forList, err := dbc.ListSignups(ctx, db.Reader(), scope, list.ID, nil)
	require.NoError(t, err)
	require.Len(t, forList.Data, 1)

	forSubject, err := dbc.ListSignupsForSubject(ctx, db.Reader(), scope, ddbwaitlists.SubjectFor(userID), nil)
	require.NoError(t, err)
	require.Len(t, forSubject.Data, 1)

	_, err = writeT(ctx, db, func(tx database.Tx) (*waitlists.Signup, error) {
		return dbc.UpdateSignupNotes(ctx, tx, scope, list.ID, joined.ID, "moved up the queue")
	})
	require.NoError(t, err)

	noted, err := dbc.GetSignup(ctx, db.Reader(), scope, list.ID, joined.ID)
	require.NoError(t, err)
	assert.Equal(t, "moved up the queue", noted.Notes)
	// A note moves nobody, which is the whole reason the two stamps are separate.
	assert.Nil(t, noted.StatusChangedAt)

	_, err = writeT(ctx, db, func(tx database.Tx) (*waitlists.Signup, error) {
		return dbc.ArchiveSignup(ctx, tx, scope, list.ID, joined.ID)
	})
	require.NoError(t, err)

	afterArchive, err := dbc.GetSignup(ctx, db.Reader(), scope, list.ID, joined.ID)
	require.Error(t, err)
	assert.Nil(t, afterArchive)
	require.ErrorIs(t, err, waitlists.ErrSignupNotFound)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, userID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeWaitlistSignups, RelevantID: joined.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeWaitlistSignups, RelevantID: joined.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeWaitlistSignups, RelevantID: joined.ID},
	})
}

// TestRepository_Integration_SignupLifecycle walks the queue this package was
// adopted for, and pins that a lost guard writes nothing and records nothing.
// Two operators inviting the same person is exactly what that looks like from
// the second one's side.
func TestRepository_Integration_SignupLifecycle(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, db)
	list := openListForTest(t, ctx, dbc, db)

	joined, err := joinT(ctx, db, dbc, scope, list.ID, fakes.BuildFakeWaitlistSignupForUser(userID))
	require.NoError(t, err)

	_, err = writeT(ctx, db, func(tx database.Tx) (*waitlists.Signup, error) {
		return dbc.Invite(ctx, tx, scope, list.ID, joined.ID)
	})
	require.NoError(t, err)

	invited, err := dbc.GetSignup(ctx, db.Reader(), scope, list.ID, joined.ID)
	require.NoError(t, err)
	assert.Equal(t, waitlists.StatusInvited, invited.Status)
	require.NotNil(t, invited.StatusChangedAt)

	// The second invitation is refused rather than sending a second email.
	_, err = writeT(ctx, db, func(tx database.Tx) (*waitlists.Signup, error) {
		return dbc.Invite(ctx, tx, scope, list.ID, joined.ID)
	})
	require.Error(t, err)
	require.ErrorIs(t, err, waitlists.ErrWrongStatus)

	_, err = writeT(ctx, db, func(tx database.Tx) (*waitlists.Signup, error) {
		return dbc.Convert(ctx, tx, scope, list.ID, joined.ID)
	})
	require.NoError(t, err)

	converted, err := dbc.GetSignup(ctx, db.Reader(), scope, list.ID, joined.ID)
	require.NoError(t, err)
	assert.Equal(t, waitlists.StatusConverted, converted.Status)

	// Three entries: the join, and the two moves that took. The refused
	// invitation is not among them.
	entries, err := auditRepo.GetAuditLogEntriesForUser(ctx, userID, nil)
	require.NoError(t, err)
	assert.Len(t, entries.Data, 3)
}

// TestRepository_Integration_WithdrawalOutlivesTheAddress is the obligation this
// adoption was for.
//
// The local table had no way to express it: a signup was a row that could be
// archived, and archiving frees nothing and suppresses nothing, so the next
// signup from the same person simply succeeded.
func TestRepository_Integration_WithdrawalOutlivesTheAddress(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, db)
	list := openListForTest(t, ctx, dbc, db)

	example := fakes.BuildFakeWaitlistSignupForUser(userID)

	joined, err := joinT(ctx, db, dbc, scope, list.ID, example)
	require.NoError(t, err)

	_, err = writeT(ctx, db, func(tx database.Tx) (*waitlists.Signup, error) {
		return dbc.Withdraw(ctx, tx, scope, list.ID, joined.ID)
	})
	require.NoError(t, err)

	// The row is still live, and it no longer says who it was about.
	withdrawn, err := dbc.GetSignup(ctx, db.Reader(), scope, list.ID, joined.ID)
	require.NoError(t, err)
	assert.Equal(t, waitlists.StatusWithdrawn, withdrawn.Status)
	assert.Empty(t, withdrawn.Contact)
	assert.Empty(t, withdrawn.Notes)
	assert.True(t, withdrawn.Subject.Anonymous())

	// Filling the form in again does not put them back on the list. The
	// suppression is on the address rather than on the person, which is what makes
	// it work for a list somebody joined with no account at all — and in this
	// application the address is the session's, so it is the same address either
	// way. See internal/services/waitlists/grpc.
	again := fakes.BuildFakeWaitlistSignupForUser(userID)
	again.Contact = example.Contact

	_, err = joinT(ctx, db, dbc, scope, list.ID, again)
	require.Error(t, err)
	require.ErrorIs(t, err, waitlists.ErrContactWithdrawn)

	// A second withdrawal reports itself rather than restamping the moment they
	// left.
	_, err = writeT(ctx, db, func(tx database.Tx) (*waitlists.Signup, error) {
		return dbc.Withdraw(ctx, tx, scope, list.ID, joined.ID)
	})
	require.Error(t, err)
	require.ErrorIs(t, err, waitlists.ErrAlreadyWithdrawn)

	// The audit entry still names them, which is the point of reading the signup
	// before the store blanks it.
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, userID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeWaitlistSignups, RelevantID: joined.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeWaitlistSignups, RelevantID: joined.ID},
	})
}

// TestRepository_Integration_ArchivingIsNotWithdrawing pins the distinction the
// store offers two methods for. An archived signup keeps its address, so the
// next attempt from it is a duplicate rather than an honored opt-out.
func TestRepository_Integration_ArchivingIsNotWithdrawing(t *testing.T) {
	ctx := t.Context()
	dbc, _, db := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, db)
	list := openListForTest(t, ctx, dbc, db)

	example := fakes.BuildFakeWaitlistSignupForUser(userID)

	joined, err := joinT(ctx, db, dbc, scope, list.ID, example)
	require.NoError(t, err)

	_, err = writeT(ctx, db, func(tx database.Tx) (*waitlists.Signup, error) {
		return dbc.ArchiveSignup(ctx, tx, scope, list.ID, joined.ID)
	})
	require.NoError(t, err)

	again := fakes.BuildFakeWaitlistSignupForUser(userID)
	again.Contact = example.Contact

	_, err = joinT(ctx, db, dbc, scope, list.ID, again)
	require.Error(t, err)
	require.ErrorIs(t, err, waitlists.ErrAlreadySignedUp)
	assert.NotErrorIs(t, err, waitlists.ErrContactWithdrawn)
}

// TestRepository_Integration_MissingRowsRecordNothing pins that a write aimed at
// a row that is not there is an error before anything is written down about it.
func TestRepository_Integration_MissingRowsRecordNothing(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, db)
	list := openListForTest(t, ctx, dbc, db)

	_, err := writeT(ctx, db, func(tx database.Tx) (*waitlists.List, error) {
		return dbc.ArchiveList(ctx, tx, scope, identifiers.New())
	})
	require.ErrorIs(t, err, waitlists.ErrListNotFound)

	for _, write := range map[string]func(tx database.Tx) (*waitlists.Signup, error){
		"archive": func(tx database.Tx) (*waitlists.Signup, error) {
			return dbc.ArchiveSignup(ctx, tx, scope, list.ID, identifiers.New())
		},
		"withdraw": func(tx database.Tx) (*waitlists.Signup, error) {
			return dbc.Withdraw(ctx, tx, scope, list.ID, identifiers.New())
		},
		"invite": func(tx database.Tx) (*waitlists.Signup, error) {
			return dbc.Invite(ctx, tx, scope, list.ID, identifiers.New())
		},
	} {
		_, err = writeT(ctx, db, write)
		require.ErrorIs(t, err, waitlists.ErrSignupNotFound)
	}

	entries, err := auditRepo.GetAuditLogEntriesForUser(ctx, userID, nil)
	require.NoError(t, err)
	assert.Empty(t, entries.Data)
}

// writeT runs one store write on a transaction of its own.
//
// As of platform-go v14 a store write takes the caller's database.Tx, so a test that wants one
// row written supplies the transaction the production caller would. It answers with the error
// rather than asserting on it, so the assertions above read as they did — and because several of
// them are about a write that is *supposed* to fail.
func writeT[T any](ctx context.Context, db database.Client, write func(tx database.Tx) (T, error)) (T, error) {
	var out T

	err := db.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		out, writeErr = write(tx)

		return writeErr
	})

	return out, err
}

// joinT signs one person up, which happens often enough here to be worth naming.
func joinT(ctx context.Context, db database.Client, dbc waitlists.Store, scope tenancy.Scope, listID string, signup *waitlists.Signup) (*waitlists.Signup, error) {
	return writeT(ctx, db, func(tx database.Tx) (*waitlists.Signup, error) {
		return dbc.Join(ctx, tx, scope, listID, signup)
	})
}
