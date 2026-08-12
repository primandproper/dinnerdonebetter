package issue_reports

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v10/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createIssueReportForTest(t *testing.T, ctx context.Context, exampleIssueReport *types.IssueReport, dbc *repository) *types.IssueReport {
	t.Helper()

	// create
	if exampleIssueReport == nil {
		exampleIssueReport = fakes.BuildFakeIssueReport()
	}
	dbInput := &types.IssueReportDatabaseCreationInput{
		ID:               exampleIssueReport.ID,
		IssueType:        exampleIssueReport.IssueType,
		Details:          exampleIssueReport.Details,
		RelevantTable:    exampleIssueReport.RelevantTable,
		RelevantRecordID: exampleIssueReport.RelevantRecordID,
		CreatedByUser:    exampleIssueReport.CreatedByUser,
		BelongsToAccount: exampleIssueReport.BelongsToAccount,
	}

	created, err := dbc.CreateIssueReport(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)

	exampleIssueReport.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleIssueReport, created)

	issueReport, err := dbc.GetIssueReport(ctx, created.ID)
	exampleIssueReport.CreatedAt = issueReport.CreatedAt

	require.NoError(t, err)
	assert.Equal(t, issueReport, exampleIssueReport)

	return created
}

func TestQuerier_Integration_IssueReports(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	exampleIssueReport := fakes.BuildFakeIssueReport()
	exampleIssueReport.BelongsToAccount = account.ID
	exampleIssueReport.CreatedByUser = user.ID
	createdIssueReports := []*types.IssueReport{}

	// create
	createdIssueReports = append(createdIssueReports, createIssueReportForTest(t, ctx, exampleIssueReport, dbc))

	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeIssueReports, RelevantID: createdIssueReports[0].ID},
	})

	// create more
	for i := range exampleQuantity {
		input := fakes.BuildFakeIssueReport()
		input.IssueType = fmt.Sprintf("%s %d", exampleIssueReport.IssueType, i)
		input.BelongsToAccount = account.ID
		input.CreatedByUser = user.ID
		createdIssueReports = append(createdIssueReports, createIssueReportForTest(t, ctx, input, dbc))
	}

	// fetch as list
	issueReports, err := dbc.GetIssueReports(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, issueReports.Data)
	assert.Len(t, issueReports.Data, len(createdIssueReports))

	// fetch as list for account
	issueReportsByAccount, err := dbc.GetIssueReportsForAccount(ctx, account.ID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, issueReportsByAccount.Data)
	assert.Len(t, issueReportsByAccount.Data, len(createdIssueReports))

	// fetch as list for table
	issueReportsByTable, err := dbc.GetIssueReportsForTable(ctx, createdIssueReports[0].RelevantTable, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, issueReportsByTable.Data)

	// fetch as list for record
	issueReportsByRecord, err := dbc.GetIssueReportsForRecord(ctx, createdIssueReports[0].RelevantTable, createdIssueReports[0].RelevantRecordID, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, issueReportsByRecord.Data)
	assert.Equal(t, createdIssueReports[0].ID, issueReportsByRecord.Data[0].ID)

	// update
	createdIssueReports[0].Details = "Updated details"
	require.NoError(t, dbc.UpdateIssueReport(ctx, createdIssueReports[0]))

	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeIssueReports, RelevantID: createdIssueReports[0].ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeIssueReports, RelevantID: createdIssueReports[0].ID},
	})

	// fetch again to verify update
	updated, err := dbc.GetIssueReport(ctx, createdIssueReports[0].ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated details", updated.Details)

	// delete
	for _, issueReport := range createdIssueReports {
		require.NoError(t, dbc.ArchiveIssueReport(ctx, issueReport.ID))

		pgtesting.AssertAuditLogContains(t, ctx, auditRepo, account.ID, []*audit.AuditLogEntry{
			{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeIssueReports, RelevantID: issueReport.ID},
		})

		var y *types.IssueReport
		y, err = dbc.GetIssueReport(ctx, issueReport.ID)
		assert.Nil(t, y)
		require.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_GetIssueReport(T *testing.T) {
	T.Parallel()

	T.Run("with invalid issue report ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetIssueReport(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetIssueReports(T *testing.T) {
	T.Parallel()

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		dbc, _ := buildDatabaseClientForTest(t)

		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
		account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

		exampleIssueReport := fakes.BuildFakeIssueReport()
		exampleIssueReport.BelongsToAccount = account.ID
		exampleIssueReport.CreatedByUser = user.ID

		created := createIssueReportForTest(t, ctx, exampleIssueReport, dbc)

		// Should work with nil filter (uses default)
		actual, err := dbc.GetIssueReports(ctx, nil)
		require.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotEmpty(t, actual.Data)

		// Cleanup
		assert.NoError(t, dbc.ArchiveIssueReport(ctx, created.ID))
	})
}

func TestQuerier_GetIssueReportsForAccount(T *testing.T) {
	T.Parallel()

	T.Run("with invalid account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		filter := filtering.DefaultQueryFilter()
		c := buildInertClientForTest(t)

		actual, err := c.GetIssueReportsForAccount(ctx, "", filter)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetIssueReportsForTable(T *testing.T) {
	T.Parallel()

	T.Run("with invalid table name", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		filter := filtering.DefaultQueryFilter()
		c := buildInertClientForTest(t)

		actual, err := c.GetIssueReportsForTable(ctx, "", filter)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetIssueReportsForRecord(T *testing.T) {
	T.Parallel()

	T.Run("with invalid table name", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		filter := filtering.DefaultQueryFilter()
		c := buildInertClientForTest(t)

		actual, err := c.GetIssueReportsForRecord(ctx, "", fakes.BuildFakeID(), filter)
		require.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid record ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		filter := filtering.DefaultQueryFilter()
		c := buildInertClientForTest(t)

		actual, err := c.GetIssueReportsForRecord(ctx, "recipes", "", filter)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateIssueReport(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateIssueReport(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateIssueReport(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.UpdateIssueReport(ctx, nil)
		assert.Error(t, err)
	})
}

func TestQuerier_ArchiveIssueReport(T *testing.T) {
	T.Parallel()

	T.Run("with invalid issue report ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveIssueReport(ctx, ""))
	})
}

func TestQuerier_Integration_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.IssueReport]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "issue report",
		CreateItem: func(ctx context.Context, i int) *types.IssueReport {
			issueReport := fakes.BuildFakeIssueReport()
			issueReport.IssueType = fmt.Sprintf("Issue Type %02d", i) // Use zero-padded numbers for consistent sorting
			issueReport.BelongsToAccount = account.ID
			issueReport.CreatedByUser = user.ID
			return createIssueReportForTest(t, ctx, issueReport, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.IssueReport], error) {
			return dbc.GetIssueReportsForAccount(ctx, account.ID, filter)
		},
		GetID: func(issueReport *types.IssueReport) string {
			return issueReport.ID
		},
		CleanupItem: func(ctx context.Context, issueReport *types.IssueReport) error {
			return dbc.ArchiveIssueReport(ctx, issueReport.ID)
		},
	})
}
