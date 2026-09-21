package issuereports

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	issuereports "github.com/primandproper/platform-go/v14/issuereports"
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
func buildDatabaseClientForTest(t *testing.T) (issuereports.Store, audit.Repository, database.Client) {
	t.Helper()

	ctx := t.Context()

	// Already migrated: the template this was cloned from was migrated once in TestMain.
	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), pgc)
	require.NoError(t, err)

	c, err := ProvideIssueReportsRepository(
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

// reporterForTest creates a user and an account for them, and returns both. Both
// rows have to exist: the rendered table re-creates the reporter and scope foreign
// keys the local table carried.
func reporterForTest(t *testing.T, db database.Client) (userID, accountID string) {
	t.Helper()

	user := pgtesting.CreateUserForTest(t, nil, db.Writer())
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, db.Writer())

	return user.ID, account.ID
}

func TestRepository_Integration_IssueReports(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)

	userID, accountID := reporterForTest(t, db)

	report := fakes.BuildFakeIssueReportForScope(accountID)
	report.Reporter = userID

	// create
	created, err := createT(ctx, db, dbc, report)
	require.NoError(t, err)
	assert.Equal(t, issuereports.StatusOpen, created.Status)
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, userID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeIssueReports, RelevantID: report.ID},
	})

	fetched, err := dbc.GetReport(ctx, db.Reader(), ddbissuereports.Scope(accountID), report.ID)
	require.NoError(t, err)
	assert.Equal(t, report.Kind, fetched.Kind)
	assert.Equal(t, report.Details, fetched.Details)
	assert.Equal(t, userID, fetched.Reporter)
	assert.Nil(t, fetched.ClosedAt)

	// read as the account's list, and as the open queue
	page, err := dbc.ListReports(ctx, db.Reader(), ddbissuereports.Scope(accountID), nil)
	require.NoError(t, err)
	require.Len(t, page.Data, 1)
	assert.Equal(t, report.ID, page.Data[0].ID)

	open, err := dbc.ListReportsByStatus(ctx, db.Reader(), ddbissuereports.Scope(accountID), issuereports.StatusOpen, nil)
	require.NoError(t, err)
	require.Len(t, open.Data, 1)

	// update
	fetched.Details = "updated details"
	_, err = writeT(ctx, db, func(tx database.Tx) (*issuereports.Report, error) {
		return dbc.UpdateReport(ctx, tx, ddbissuereports.Scope(accountID), fetched)
	})
	require.NoError(t, err)
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, userID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeIssueReports, RelevantID: report.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeIssueReports, RelevantID: report.ID},
	})

	updated, err := dbc.GetReport(ctx, db.Reader(), ddbissuereports.Scope(accountID), report.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated details", updated.Details)
	assert.NotNil(t, updated.LastUpdatedAt)

	// archive
	_, err = writeT(ctx, db, func(tx database.Tx) (*issuereports.Report, error) {
		return dbc.ArchiveReport(ctx, tx, ddbissuereports.Scope(accountID), report.ID)
	})
	require.NoError(t, err)
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, userID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeIssueReports, RelevantID: report.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeIssueReports, RelevantID: report.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeIssueReports, RelevantID: report.ID},
	})

	afterArchive, err := dbc.GetReport(ctx, db.Reader(), ddbissuereports.Scope(accountID), report.ID)
	require.Error(t, err)
	assert.Nil(t, afterArchive)
	assert.ErrorIs(t, err, issuereports.ErrReportNotFound)
}

// TestRepository_Integration_TriageLifecycle walks the lifecycle this package was
// adopted for, and pins the two facts a queue rests on: a terminal status stamps
// closed_at and stores the note, and reopening clears both.
func TestRepository_Integration_TriageLifecycle(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)

	userID, accountID := reporterForTest(t, db)
	scope := ddbissuereports.Scope(accountID)

	report := fakes.BuildFakeIssueReportForScope(accountID)
	report.Reporter = userID
	_, err := createT(ctx, db, dbc, report)
	require.NoError(t, err)

	acknowledged, err := transitionT(ctx, db, dbc, scope, report.ID, issuereports.StatusOpen, issuereports.StatusAcknowledged, "")
	require.NoError(t, err)
	assert.Equal(t, issuereports.StatusAcknowledged, acknowledged.Status)
	assert.Nil(t, acknowledged.ClosedAt)

	resolved, err := transitionT(ctx, db, dbc, scope, report.ID, issuereports.StatusAcknowledged, issuereports.StatusResolved, "fixed")
	require.NoError(t, err)
	assert.Equal(t, issuereports.StatusResolved, resolved.Status)
	assert.Equal(t, "fixed", resolved.Resolution)
	require.NotNil(t, resolved.ClosedAt)

	// A reopen clears the closure, because a reason that no longer holds is worse
	// than none.
	reopened, err := transitionT(ctx, db, dbc, scope, report.ID, issuereports.StatusResolved, issuereports.StatusOpen, "")
	require.NoError(t, err)
	assert.Equal(t, issuereports.StatusOpen, reopened.Status)
	assert.Empty(t, reopened.Resolution)
	assert.Nil(t, reopened.ClosedAt)

	// Every move is recorded, so "who resolved this and when" is answerable from
	// the audit log rather than from the one row the last write left behind.
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, userID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeIssueReports, RelevantID: report.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeIssueReports, RelevantID: report.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeIssueReports, RelevantID: report.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeIssueReports, RelevantID: report.ID},
	})
}

// TestRepository_Integration_TransitionGuardRecordsNothing pins that a lost guard
// writes nothing and records nothing. Two triagers resolving the same report is
// exactly what this looks like from the second one's side.
func TestRepository_Integration_TransitionGuardRecordsNothing(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)

	userID, accountID := reporterForTest(t, db)
	scope := ddbissuereports.Scope(accountID)

	report := fakes.BuildFakeIssueReportForScope(accountID)
	report.Reporter = userID
	_, err := createT(ctx, db, dbc, report)
	require.NoError(t, err)

	_, err = transitionT(ctx, db, dbc, scope, report.ID, issuereports.StatusOpen, issuereports.StatusResolved, "first")
	require.NoError(t, err)

	second, err := transitionT(ctx, db, dbc, scope, report.ID, issuereports.StatusOpen, issuereports.StatusResolved, "second")
	require.Error(t, err)
	assert.Nil(t, second)
	require.ErrorIs(t, err, issuereports.ErrStatusConflict)

	// The first note stands. The whole point of the guard is that the second write
	// does not overwrite it.
	stored, err := dbc.GetReport(ctx, db.Reader(), scope, report.ID)
	require.NoError(t, err)
	assert.Equal(t, "first", stored.Resolution)

	entries, err := auditRepo.GetAuditLogEntriesForUser(ctx, userID, nil)
	require.NoError(t, err)
	assert.Len(t, entries.Data, 2)
}

// TestRepository_Integration_ScopeIsTheAccountBoundary pins that a report filed in
// one account is not readable from another. This is what replaced the
// belongs_to_account check the service used to run after the read.
func TestRepository_Integration_ScopeIsTheAccountBoundary(t *testing.T) {
	ctx := t.Context()
	dbc, _, db := buildDatabaseClientForTest(t)

	userID, accountID := reporterForTest(t, db)
	_, otherAccountID := reporterForTest(t, db)

	report := fakes.BuildFakeIssueReportForScope(accountID)
	report.Reporter = userID
	_, err := createT(ctx, db, dbc, report)
	require.NoError(t, err)

	fetched, err := dbc.GetReport(ctx, db.Reader(), ddbissuereports.Scope(otherAccountID), report.ID)
	require.Error(t, err)
	assert.Nil(t, fetched)
	require.ErrorIs(t, err, issuereports.ErrReportNotFound)

	page, err := dbc.ListReports(ctx, db.Reader(), ddbissuereports.Scope(otherAccountID), nil)
	require.NoError(t, err)
	assert.Empty(t, page.Data)
}

// TestRepository_Integration_ErasureFollowsTheReporter pins the cascade this
// adoption re-created by hand. The details are free text somebody typed, so a
// report that outlived its reporter would be personal data the single identity
// eraser never reaches.
func TestRepository_Integration_ErasureFollowsTheReporter(t *testing.T) {
	ctx := t.Context()
	dbc, _, db := buildDatabaseClientForTest(t)

	userID, accountID := reporterForTest(t, db)

	report := fakes.BuildFakeIssueReportForScope(accountID)
	report.Reporter = userID
	_, err := createT(ctx, db, dbc, report)
	require.NoError(t, err)

	_, err = db.Writer().ExecContext(ctx, "DELETE FROM ddb_identity_users WHERE id = $1", userID)
	require.NoError(t, err)

	fetched, err := dbc.GetReport(ctx, db.Reader(), ddbissuereports.Scope(accountID), report.ID)
	require.Error(t, err)
	assert.Nil(t, fetched)
	assert.ErrorIs(t, err, issuereports.ErrReportNotFound)
}

// TestRepository_Integration_ArchiveMissingRecordsNothing pins that a failed
// archive records nothing. The read that finds the reporter is also what makes an
// absent report an error before anything is written down about it.
func TestRepository_Integration_ArchiveMissingRecordsNothing(t *testing.T) {
	ctx := t.Context()
	dbc, _, db := buildDatabaseClientForTest(t)

	_, accountID := reporterForTest(t, db)

	_, err := writeT(ctx, db, func(tx database.Tx) (*issuereports.Report, error) {
		return dbc.ArchiveReport(ctx, tx, ddbissuereports.Scope(accountID), identifiers.New())
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, issuereports.ErrReportNotFound)
}

// writeT runs one store write on a transaction of its own.
//
// As of platform-go v14 a store write takes the caller's database.Tx, so a test that wants one
// row written supplies the transaction the production caller would. It answers with the error
// rather than asserting on it, because several of the writes here are supposed to fail.
func writeT[T any](ctx context.Context, db database.Client, write func(tx database.Tx) (T, error)) (T, error) {
	var out T

	err := db.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		out, writeErr = write(tx)

		return writeErr
	})

	return out, err
}

// createT files one report, in the scope the report itself names.
func createT(ctx context.Context, db database.Client, dbc issuereports.Store, report *issuereports.Report) (*issuereports.Report, error) {
	return writeT(ctx, db, func(tx database.Tx) (*issuereports.Report, error) {
		return dbc.CreateReport(ctx, tx, ddbissuereports.Scope(report.Scope.Owner()), report)
	})
}

// transitionT moves one report between statuses.
func transitionT(ctx context.Context, db database.Client, dbc issuereports.Store, scope tenancy.Scope, reportID string, from, to issuereports.Status, resolution string) (*issuereports.Report, error) {
	return writeT(ctx, db, func(tx database.Tx) (*issuereports.Report, error) {
		return dbc.TransitionReport(ctx, tx, scope, reportID, from, to, resolution)
	})
}
