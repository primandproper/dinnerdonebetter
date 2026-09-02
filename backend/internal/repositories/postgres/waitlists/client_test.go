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

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/database/postgres"
	"github.com/primandproper/platform-go/v13/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	waitlists "github.com/primandproper/platform-go/v13/waitlists"

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
func buildDatabaseClientForTest(t *testing.T) (waitlists.Store, audit.Repository, database.SQLQueryExecutor) {
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

	return c, auditLogEntryRepo, pgc.Writer()
}

// signatoryForTest creates a user and an account for them, and returns the user.
//
// The signup table has no foreign key to either — a withdrawal blanks the
// subject reference, so it cannot have one — but the audit entry a signup write
// records names the user, and the audit chain does.
func signatoryForTest(t *testing.T, writer database.SQLQueryExecutor) string {
	t.Helper()

	user := pgtesting.CreateUserForTest(t, nil, writer)
	pgtesting.CreateAccountForTest(t, nil, user.ID, writer)

	return user.ID
}

// openListForTest opens one list that is still taking signups.
func openListForTest(t *testing.T, ctx context.Context, dbc waitlists.Store) *waitlists.List {
	t.Helper()

	list, err := dbc.CreateList(ctx, ddbwaitlists.Scope(), fakes.BuildFakeWaitlist())
	require.NoError(t, err)

	return list
}

func TestRepository_Integration_Waitlists(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, _ := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	example := fakes.BuildFakeWaitlist()

	created, err := dbc.CreateList(ctx, scope, example)
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

	fetched, err := dbc.GetList(ctx, scope, created.ID)
	require.NoError(t, err)
	assert.Equal(t, example.Name, fetched.Name)
	assert.Equal(t, example.Description, fetched.Description)

	page, err := dbc.ListLists(ctx, scope, nil)
	require.NoError(t, err)
	require.Len(t, page.Data, 1)
	assert.Equal(t, created.ID, page.Data[0].ID)

	// An open list is on the open page, which is the read a signup form offers.
	open, err := dbc.ListOpenLists(ctx, scope, nil)
	require.NoError(t, err)
	require.Len(t, open.Data, 1)

	fetched.Name = "renamed"
	require.NoError(t, dbc.UpdateList(ctx, scope, fetched))

	updated, err := dbc.GetList(ctx, scope, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Name)
	assert.NotNil(t, updated.LastUpdatedAt)

	require.NoError(t, dbc.ArchiveList(ctx, scope, created.ID))

	afterArchive, err := dbc.GetList(ctx, scope, created.ID)
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
	dbc, _, writer := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, writer)
	list := openListForTest(t, ctx, dbc)

	require.NoError(t, dbc.ArchiveList(ctx, scope, list.ID))

	_, err := dbc.Join(ctx, scope, list.ID, fakes.BuildFakeWaitlistSignupForUser(userID))
	require.Error(t, err)
	// The list is gone as far as the signup path is concerned: the read that
	// decides whether it is open cannot find it.
	require.ErrorIs(t, err, waitlists.ErrListNotFound)
}

// TestRepository_Integration_ClosedListTakesNoSignups pins the other half: a live
// list past its closing time refuses a signup and says why.
func TestRepository_Integration_ClosedListTakesNoSignups(t *testing.T) {
	ctx := t.Context()
	dbc, _, writer := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, writer)

	closed := fakes.BuildFakeWaitlist()
	closed.ClosesAt = time.Now().Add(-time.Hour).UTC()

	list, err := dbc.CreateList(ctx, scope, closed)
	require.NoError(t, err)

	_, err = dbc.Join(ctx, scope, list.ID, fakes.BuildFakeWaitlistSignupForUser(userID))
	require.Error(t, err)
	require.ErrorIs(t, err, waitlists.ErrListClosed)

	// And it is off the open page while still being in the catalog.
	open, err := dbc.ListOpenLists(ctx, scope, nil)
	require.NoError(t, err)
	assert.Empty(t, open.Data)

	all, err := dbc.ListLists(ctx, scope, nil)
	require.NoError(t, err)
	assert.Len(t, all.Data, 1)
}

func TestRepository_Integration_WaitlistSignups(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, writer := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, writer)
	list := openListForTest(t, ctx, dbc)

	example := fakes.BuildFakeWaitlistSignupForUser(userID)

	joined, err := dbc.Join(ctx, scope, list.ID, example)
	require.NoError(t, err)
	assert.Equal(t, waitlists.StatusWaiting, joined.Status)
	assert.Equal(t, list.ID, joined.ListID)
	assert.NotEmpty(t, joined.ContactDigest)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, userID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeWaitlistSignups, RelevantID: joined.ID},
	})

	fetched, err := dbc.GetSignup(ctx, scope, list.ID, joined.ID)
	require.NoError(t, err)
	assert.Equal(t, example.Contact, fetched.Contact)
	assert.Equal(t, ddbwaitlists.SubjectFor(userID), fetched.Subject)

	// The address finds the row whichever capitalization the caller has.
	byContact, err := dbc.GetSignupByContact(ctx, scope, list.ID, strings.ToUpper(example.Contact))
	require.NoError(t, err)
	assert.Equal(t, joined.ID, byContact.ID)

	forList, err := dbc.ListSignups(ctx, scope, list.ID, nil)
	require.NoError(t, err)
	require.Len(t, forList.Data, 1)

	forSubject, err := dbc.ListSignupsForSubject(ctx, scope, ddbwaitlists.SubjectFor(userID), nil)
	require.NoError(t, err)
	require.Len(t, forSubject.Data, 1)

	require.NoError(t, dbc.UpdateSignupNotes(ctx, scope, list.ID, joined.ID, "moved up the queue"))

	noted, err := dbc.GetSignup(ctx, scope, list.ID, joined.ID)
	require.NoError(t, err)
	assert.Equal(t, "moved up the queue", noted.Notes)
	// A note moves nobody, which is the whole reason the two stamps are separate.
	assert.Nil(t, noted.StatusChangedAt)

	require.NoError(t, dbc.ArchiveSignup(ctx, scope, list.ID, joined.ID))

	afterArchive, err := dbc.GetSignup(ctx, scope, list.ID, joined.ID)
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
	dbc, auditRepo, writer := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, writer)
	list := openListForTest(t, ctx, dbc)

	joined, err := dbc.Join(ctx, scope, list.ID, fakes.BuildFakeWaitlistSignupForUser(userID))
	require.NoError(t, err)

	require.NoError(t, dbc.Invite(ctx, scope, list.ID, joined.ID))

	invited, err := dbc.GetSignup(ctx, scope, list.ID, joined.ID)
	require.NoError(t, err)
	assert.Equal(t, waitlists.StatusInvited, invited.Status)
	require.NotNil(t, invited.StatusChangedAt)

	// The second invitation is refused rather than sending a second email.
	err = dbc.Invite(ctx, scope, list.ID, joined.ID)
	require.Error(t, err)
	require.ErrorIs(t, err, waitlists.ErrWrongStatus)

	require.NoError(t, dbc.Convert(ctx, scope, list.ID, joined.ID))

	converted, err := dbc.GetSignup(ctx, scope, list.ID, joined.ID)
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
	dbc, auditRepo, writer := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, writer)
	list := openListForTest(t, ctx, dbc)

	example := fakes.BuildFakeWaitlistSignupForUser(userID)

	joined, err := dbc.Join(ctx, scope, list.ID, example)
	require.NoError(t, err)

	require.NoError(t, dbc.Withdraw(ctx, scope, list.ID, joined.ID))

	// The row is still live, and it no longer says who it was about.
	withdrawn, err := dbc.GetSignup(ctx, scope, list.ID, joined.ID)
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

	_, err = dbc.Join(ctx, scope, list.ID, again)
	require.Error(t, err)
	require.ErrorIs(t, err, waitlists.ErrContactWithdrawn)

	// A second withdrawal reports itself rather than restamping the moment they
	// left.
	err = dbc.Withdraw(ctx, scope, list.ID, joined.ID)
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
	dbc, _, writer := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, writer)
	list := openListForTest(t, ctx, dbc)

	example := fakes.BuildFakeWaitlistSignupForUser(userID)

	joined, err := dbc.Join(ctx, scope, list.ID, example)
	require.NoError(t, err)
	require.NoError(t, dbc.ArchiveSignup(ctx, scope, list.ID, joined.ID))

	again := fakes.BuildFakeWaitlistSignupForUser(userID)
	again.Contact = example.Contact

	_, err = dbc.Join(ctx, scope, list.ID, again)
	require.Error(t, err)
	require.ErrorIs(t, err, waitlists.ErrAlreadySignedUp)
	assert.NotErrorIs(t, err, waitlists.ErrContactWithdrawn)
}

// TestRepository_Integration_MissingRowsRecordNothing pins that a write aimed at
// a row that is not there is an error before anything is written down about it.
func TestRepository_Integration_MissingRowsRecordNothing(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, writer := buildDatabaseClientForTest(t)
	scope := ddbwaitlists.Scope()

	userID := signatoryForTest(t, writer)
	list := openListForTest(t, ctx, dbc)

	require.ErrorIs(t, dbc.ArchiveList(ctx, scope, identifiers.New()), waitlists.ErrListNotFound)
	require.ErrorIs(t, dbc.ArchiveSignup(ctx, scope, list.ID, identifiers.New()), waitlists.ErrSignupNotFound)
	require.ErrorIs(t, dbc.Withdraw(ctx, scope, list.ID, identifiers.New()), waitlists.ErrSignupNotFound)
	require.ErrorIs(t, dbc.Invite(ctx, scope, list.ID, identifiers.New()), waitlists.ErrSignupNotFound)

	entries, err := auditRepo.GetAuditLogEntriesForUser(ctx, userID, nil)
	require.NoError(t, err)
	assert.Empty(t, entries.Data)
}
