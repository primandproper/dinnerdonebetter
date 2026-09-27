package integration

import (
	"testing"

	issuereportfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	issuereportspb "github.com/primandproper/platform-go/v14/issuereports/issuereportspb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The issue reports surface's behavior is asserted by platform's issuereports conformance suite,
// run against this deployment in conformance_test.go. What remains here is the audit entries this
// application's store records, and two update behaviors that suite does not assert yet.

func checkIssueReportEquality(t *testing.T, expected, actual *issuereportspb.IssueReport) {
	t.Helper()

	assert.NotEmpty(t, actual.GetId(), "expected IssueReport to have ID")
	assert.NotNil(t, actual.GetCreatedAt(), "expected IssueReport to have CreatedAt")

	assert.Equal(t, expected.GetKind(), actual.GetKind(), "expected IssueReport Kind")
	assert.Equal(t, expected.GetDetails(), actual.GetDetails(), "expected IssueReport Details")
	assert.Equal(t, expected.GetSubjectType(), actual.GetSubjectType(), "expected IssueReport SubjectType")
	assert.Equal(t, expected.GetSubjectId(), actual.GetSubjectId(), "expected IssueReport SubjectID")
	assert.NotEmpty(t, actual.GetReporter(), "expected IssueReport to have Reporter")
}

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

// createIssueReportForTest files one report and reads it back.
func createIssueReportForTest(t *testing.T, testClient client.Client) *issuereportspb.IssueReport {
	t.Helper()
	ctx := t.Context()

	created, err := testClient.CreateReport(ctx, &issuereportspb.CreateReportRequest{Input: creationInputForTest()})
	require.NoError(t, err)
	require.NotNil(t, created.GetResult())

	// A report is born open, whatever the client sent.
	assert.Equal(t, issuereportspb.ReportStatus_REPORT_STATUS_OPEN, created.GetResult().GetStatus())
	assert.Empty(t, created.GetResult().GetResolution())
	assert.Nil(t, created.GetResult().GetClosedAt())

	retrieved, err := testClient.GetReport(ctx, &issuereportspb.GetReportRequest{ReportId: created.GetResult().GetId()})
	require.NoError(t, err)
	require.NotNil(t, retrieved.GetResult())
	checkIssueReportEquality(t, created.GetResult(), retrieved.GetResult())

	return retrieved.GetResult()
}

func TestIssueReports_Creating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "issue_reports", RelevantID: created.GetId()},
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
			{EventType: "created", ResourceType: "issue_reports", RelevantID: created.GetId()},
			{EventType: "updated", ResourceType: "issue_reports", RelevantID: created.GetId()},
		})
	})

	// The behavior a client has to know about. The local RPC took *string fields and
	// merged, so a partial update left the rest alone; platform's takes plain strings and
	// writes all four, so an omitted field is cleared rather than kept — a read before the
	// write is the client's job now.
	//
	// With one guard underneath it. Two of the four are what make a report reachable at
	// all, so clearing them is refused rather than performed: a report with no kind is one
	// nobody has decided who should look at, and one with no details records that somebody
	// was unhappy and nothing anyone can act on. The two optional fields are the ones a
	// careless update really does drop.
	//
	// This stays until platform's issuereports conformance suite asserts it.
	T.Run("an omitted optional field is cleared rather than kept", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)
		require.NotEmpty(t, created.GetSubjectType())

		updated, err := testClient.UpdateReport(ctx, &issuereportspb.UpdateReportRequest{
			ReportId: created.GetId(),
			Input: &issuereportspb.IssueReportUpdateInput{
				Kind:    created.GetKind(),
				Details: "only the details",
			},
		})
		require.NoError(t, err)

		assert.Equal(t, "only the details", updated.GetResult().GetDetails())
		assert.Equal(t, created.GetKind(), updated.GetResult().GetKind())
		assert.Empty(t, updated.GetResult().GetSubjectType(), "an omitted subject type survived a replacement")
		assert.Empty(t, updated.GetResult().GetSubjectId())
	})

	// This stays until platform's issuereports conformance suite asserts it.
	T.Run("an update that would empty the kind is refused", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)
		require.NotEmpty(t, created.GetKind())

		_, err := testClient.UpdateReport(ctx, &issuereportspb.UpdateReportRequest{
			ReportId: created.GetId(),
			Input:    &issuereportspb.IssueReportUpdateInput{Details: "only the details"},
		})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))

		// Refused before anything was written, so the report is as it was.
		retrieved, err := testClient.GetReport(ctx, &issuereportspb.GetReportRequest{ReportId: created.GetId()})
		require.NoError(t, err)
		assert.Equal(t, created.GetKind(), retrieved.GetResult().GetKind())
		assert.Equal(t, created.GetDetails(), retrieved.GetResult().GetDetails())
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
			{EventType: "created", ResourceType: "issue_reports", RelevantID: created.GetId()},
			{EventType: "archived", ResourceType: "issue_reports", RelevantID: created.GetId()},
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
