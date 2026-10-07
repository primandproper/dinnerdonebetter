package integration

import (
	"testing"

	issuereportfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	platformissuereports "github.com/primandproper/platform-go/v15/issuereports"
	issuereportspb "github.com/primandproper/platform-go/v15/issuereports/issuereportspb"

	"github.com/stretchr/testify/require"
)

// The issue reports surface's behavior is asserted by platform's issuereports conformance suite,
// run against this deployment in conformance_test.go. What remains here is the audit entries this
// application's store records.

// creationInputForTest builds what a client sends to file a report.
func creationInputForTest() *issuereportspb.IssueReportCreationInput {
	example := issuereportfakes.BuildFakeIssueReport()

	return &issuereportspb.IssueReportCreationInput{
		Kind:        example.Kind,
		Details:     example.Details,
		SubjectType: example.SubjectType,
		SubjectId:   example.SubjectID,
	}
}

// createIssueReportForTest files one report. What the filing answers, and that the filer reads
// it back, is conformance/issuereports'.
func createIssueReportForTest(t *testing.T, testClient client.Client) *issuereportspb.IssueReport {
	t.Helper()

	created, err := testClient.CreateReport(t.Context(), &issuereportspb.CreateReportRequest{Input: creationInputForTest()})
	require.NoError(t, err)
	require.NotNil(t, created.GetResult())

	return created.GetResult()
}

func TestIssueReports_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: platformissuereports.ResourceTypeReport, RelevantID: created.GetId()},
		})
	})
}

func TestIssueReports_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		_, err := testClient.UpdateReport(ctx, &issuereportspb.UpdateReportRequest{
			ReportId: created.GetId(),
			Input: &issuereportspb.IssueReportUpdateInput{
				Kind:        created.GetKind(),
				Details:     "Updated details about the issue",
				SubjectType: created.GetSubjectType(),
				SubjectId:   created.GetSubjectId(),
			},
		})
		require.NoError(t, err)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: platformissuereports.ResourceTypeReport, RelevantID: created.GetId()},
			{EventType: "updated", ResourceType: platformissuereports.ResourceTypeReport, RelevantID: created.GetId()},
		})
	})
}

func TestIssueReports_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		_, err := testClient.ArchiveReport(ctx, &issuereportspb.ArchiveReportRequest{ReportId: created.GetId()})
		require.NoError(t, err)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: platformissuereports.ResourceTypeReport, RelevantID: created.GetId()},
			{EventType: "archived", ResourceType: platformissuereports.ResourceTypeReport, RelevantID: created.GetId()},
		})
	})
}

// TestIssueReports_Commenting is gone with the AddCommentToIssueReport RPC.
//
// That RPC named the target from its own name and forwarded to CreateComment;
// the existence check it appeared to add was the comment store's, through the
// target catalog. Commenting on an issue report is now platform's
// CommentsService.CreateComment with an issue-report target, and it is covered
// where every other comment path is.
