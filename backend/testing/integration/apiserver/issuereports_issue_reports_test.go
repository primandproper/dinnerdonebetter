package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/converters"
	issuereportfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/fakes"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/issuereports/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkIssueReportEquality(t *testing.T, expected, actual *issuereports.IssueReport) {
	t.Helper()

	assert.NotEmpty(t, actual.ID, "expected IssueReport to have ID")
	assert.NotZero(t, actual.CreatedAt, "expected IssueReport to have CreatedAt")

	assert.Equal(t, expected.IssueType, actual.IssueType, "expected IssueReport IssueType")
	assert.Equal(t, expected.Details, actual.Details, "expected IssueReport Details")
	assert.Equal(t, expected.RelevantTable, actual.RelevantTable, "expected IssueReport RelevantTable")
	assert.Equal(t, expected.RelevantRecordID, actual.RelevantRecordID, "expected IssueReport RelevantRecordID")
	assert.NotEmpty(t, actual.CreatedByUser, "expected IssueReport to have CreatedByUser")
	assert.NotEmpty(t, actual.BelongsToAccount, "expected IssueReport to have BelongsToAccount")
}

func createIssueReportForTest(t *testing.T, testClient client.Client) *issuereports.IssueReport {
	t.Helper()
	ctx := t.Context()

	exampleIssueReport := issuereportfakes.BuildFakeIssueReport()
	exampleIssueReportInput := converters.ConvertIssueReportToIssueReportCreationRequestInput(exampleIssueReport)

	input := grpcconverters.ConvertIssueReportCreationRequestInputToGRPCIssueReportCreationRequestInput(exampleIssueReportInput)

	createdIssueReport, err := testClient.CreateIssueReport(ctx, &issuereportssvc.CreateIssueReportRequest{Input: input})
	require.NoError(t, err)
	converted := grpcconverters.ConvertGRPCIssueReportToIssueReport(createdIssueReport.Created)
	checkIssueReportEquality(t, exampleIssueReport, converted)

	retrievedIssueReport, err := testClient.GetIssueReport(ctx, &issuereportssvc.GetIssueReportRequest{IssueReportId: createdIssueReport.Created.Id})
	require.NoError(t, err)
	require.NotNil(t, retrievedIssueReport)

	issueReport := grpcconverters.ConvertGRPCIssueReportToIssueReport(retrievedIssueReport.Result)
	checkIssueReportEquality(t, converted, issueReport)

	return issueReport
}

func TestIssueReports_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "issue_reports", RelevantID: created.ID},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateIssueReport(ctx, &issuereportssvc.CreateIssueReportRequest{})
		require.Error(t, err)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		exampleIssueReportInput := &issuereports.IssueReportCreationRequestInput{
			IssueType: "", // empty issue type should fail validation
			Details:   "",
		}

		input := grpcconverters.ConvertIssueReportCreationRequestInputToGRPCIssueReportCreationRequestInput(exampleIssueReportInput)

		_, err := testClient.CreateIssueReport(ctx, &issuereportssvc.CreateIssueReportRequest{Input: input})
		assert.Error(t, err)
	})
}

func TestIssueReports_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdIssueReport := createIssueReportForTest(t, testClient)

		retrieved, err := testClient.GetIssueReport(ctx, &issuereportssvc.GetIssueReportRequest{IssueReportId: createdIssueReport.ID})
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrieved, err := testClient.GetIssueReport(ctx, &issuereportssvc.GetIssueReportRequest{IssueReportId: nonexistentID})
		assert.Error(t, err)
		assert.Nil(t, retrieved)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetIssueReport(ctx, &issuereportssvc.GetIssueReportRequest{})
		assert.Error(t, err)
	})
}

func TestIssueReports_Listing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		createdIssueReports := []*issuereports.IssueReport{}
		for range exampleQuantity {
			createdIssueReports = append(createdIssueReports, createIssueReportForTest(t, testClient))
		}

		results, err := testClient.GetIssueReports(ctx, &issuereportssvc.GetIssueReportsRequest{})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdIssueReports))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetIssueReports(ctx, &issuereportssvc.GetIssueReportsRequest{})
		assert.Error(t, err)
	})
}

func TestIssueReports_ListingForAccount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		activeAccount, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)

		createdIssueReports := []*issuereports.IssueReport{}
		for range exampleQuantity {
			createdIssueReports = append(createdIssueReports, createIssueReportForTest(t, testClient))
		}

		results, err := testClient.GetIssueReportsForAccount(ctx, &issuereportssvc.GetIssueReportsForAccountRequest{
			AccountId: activeAccount.Result.Id,
		})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), len(createdIssueReports))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetIssueReportsForAccount(ctx, &issuereportssvc.GetIssueReportsForAccountRequest{})
		assert.Error(t, err)
	})
}

func TestIssueReports_ListingForTable(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		// Create some issue reports with specific table names
		tableName := "recipes"
		for range exampleQuantity {
			exampleIssueReport := issuereportfakes.BuildFakeIssueReport()
			exampleIssueReport.RelevantTable = tableName
			exampleIssueReportInput := converters.ConvertIssueReportToIssueReportCreationRequestInput(exampleIssueReport)
			input := grpcconverters.ConvertIssueReportCreationRequestInputToGRPCIssueReportCreationRequestInput(exampleIssueReportInput)

			_, err := testClient.CreateIssueReport(ctx, &issuereportssvc.CreateIssueReportRequest{Input: input})
			require.NoError(t, err)
		}

		results, err := testClient.GetIssueReportsForTable(ctx, &issuereportssvc.GetIssueReportsForTableRequest{
			TableName: tableName,
		})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), exampleQuantity)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetIssueReportsForTable(ctx, &issuereportssvc.GetIssueReportsForTableRequest{})
		assert.Error(t, err)
	})
}

func TestIssueReports_ListingForRecord(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		// Create some issue reports with specific table name and record ID
		tableName := "recipes"
		recordID := "test-record-123"
		for range exampleQuantity {
			exampleIssueReport := issuereportfakes.BuildFakeIssueReport()
			exampleIssueReport.RelevantTable = tableName
			exampleIssueReport.RelevantRecordID = recordID
			exampleIssueReportInput := converters.ConvertIssueReportToIssueReportCreationRequestInput(exampleIssueReport)
			input := grpcconverters.ConvertIssueReportCreationRequestInputToGRPCIssueReportCreationRequestInput(exampleIssueReportInput)

			_, err := testClient.CreateIssueReport(ctx, &issuereportssvc.CreateIssueReportRequest{Input: input})
			require.NoError(t, err)
		}

		results, err := testClient.GetIssueReportsForRecord(ctx, &issuereportssvc.GetIssueReportsForRecordRequest{
			TableName: tableName,
			RecordId:  recordID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.GreaterOrEqual(t, len(results.Results), exampleQuantity)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetIssueReportsForRecord(ctx, &issuereportssvc.GetIssueReportsForRecordRequest{})
		assert.Error(t, err)
	})
}

func TestIssueReports_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdIssueReport := createIssueReportForTest(t, testClient)

		newDetails := "Updated details about the issue"
		updated, err := testClient.UpdateIssueReport(ctx, &issuereportssvc.UpdateIssueReportRequest{
			IssueReportId: createdIssueReport.ID,
			Input: &issuereportssvc.IssueReportUpdateRequestInput{
				Details: &newDetails,
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, updated)
		assert.Equal(t, newDetails, updated.Updated.Details)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "issue_reports", RelevantID: createdIssueReport.ID},
			{EventType: "updated", ResourceType: "issue_reports", RelevantID: createdIssueReport.ID},
		})
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		newDetails := "Updated details"
		_, err := testClient.UpdateIssueReport(ctx, &issuereportssvc.UpdateIssueReportRequest{
			IssueReportId: nonexistentID,
			Input: &issuereportssvc.IssueReportUpdateRequestInput{
				Details: &newDetails,
			},
		})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateIssueReport(ctx, &issuereportssvc.UpdateIssueReportRequest{})
		assert.Error(t, err)
	})
}

func TestIssueReports_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createdIssueReport := createIssueReportForTest(t, testClient)

		_, err := testClient.ArchiveIssueReport(ctx, &issuereportssvc.ArchiveIssueReportRequest{IssueReportId: createdIssueReport.ID})
		assert.NoError(t, err)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "issue_reports", RelevantID: createdIssueReport.ID},
			{EventType: "archived", ResourceType: "issue_reports", RelevantID: createdIssueReport.ID},
		})
	})

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		createIssueReportForTest(t, testClient)

		_, err := testClient.ArchiveIssueReport(ctx, &issuereportssvc.ArchiveIssueReportRequest{IssueReportId: nonexistentID})
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveIssueReport(ctx, &issuereportssvc.ArchiveIssueReportRequest{})
		assert.Error(t, err)
	})
}
