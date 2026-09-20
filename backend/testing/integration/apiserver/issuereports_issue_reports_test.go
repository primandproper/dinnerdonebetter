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

// The issue reports surface is platform's now, and two things about it differ from the
// local one these tests were written against.
//
// A status is an enum rather than a string, so the case that sent "closed" and expected
// InvalidArgument is gone: an unknown status is now unrepresentable on the wire, and what
// remains expressible is UNSPECIFIED, which is what the replacement case sends.
//
// UpdateReport replaces rather than merges. The local RPC took *string fields and left
// anything omitted alone; platform's takes plain strings and writes all four, so a client
// sending only Details blanks the kind and the subject. That is a real change for a client,
// and it is pinned below rather than worked around.

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

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.CreateReport(ctx, &issuereportspb.CreateReportRequest{})
		require.Error(t, err)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.CreateReport(ctx, &issuereportspb.CreateReportRequest{
			Input: &issuereportspb.IssueReportCreationInput{Kind: "", Details: ""},
		})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestIssueReports_Reading(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		retrieved, err := testClient.GetReport(ctx, &issuereportspb.GetReportRequest{ReportId: created.GetId()})
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrieved, err := testClient.GetReport(ctx, &issuereportspb.GetReportRequest{ReportId: nonexistentID})
		require.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	// A report filed by another account is absent rather than forbidden, which is
	// the account boundary the tenancy column now enforces. The old service read the
	// row and then compared belongs_to_account, which told the caller which report
	// IDs existed elsewhere.
	T.Run("another account's report is not found", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, ownerClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, ownerClient)

		_, otherClient := createUserAndClientForTest(t)

		_, err := otherClient.GetReport(ctx, &issuereportspb.GetReportRequest{ReportId: created.GetId()})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetReport(ctx, &issuereportspb.GetReportRequest{})
		assert.Error(t, err)
	})
}

func TestIssueReports_Listing(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		expected := []string{}
		for range exampleQuantity {
			expected = append(expected, createIssueReportForTest(t, testClient).GetId())
		}

		results, err := testClient.ListReports(ctx, &issuereportspb.ListReportsRequest{})
		require.NoError(t, err)
		require.NotNil(t, results)

		// The list is the account's, not everybody's: it holds exactly what this
		// account filed.
		actual := []string{}
		for _, report := range results.GetResults() {
			actual = append(actual, report.GetId())
		}
		assert.ElementsMatch(t, expected, actual)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ListReports(ctx, &issuereportspb.ListReportsRequest{})
		assert.Error(t, err)
	})
}

func TestIssueReports_ListingByStatus(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		open, err := testClient.ListReportsByStatus(ctx, &issuereportspb.ListReportsByStatusRequest{
			Status: issuereportspb.ReportStatus_REPORT_STATUS_OPEN,
		})
		require.NoError(t, err)
		require.Len(t, open.GetResults(), 1)
		assert.Equal(t, created.GetId(), open.GetResults()[0].GetId())

		resolved, err := testClient.ListReportsByStatus(ctx, &issuereportspb.ListReportsByStatusRequest{
			Status: issuereportspb.ReportStatus_REPORT_STATUS_RESOLVED,
		})
		require.NoError(t, err)
		assert.Empty(t, resolved.GetResults())
	})

	// An unknown status is unrepresentable now that the field is an enum, so what is left
	// to refuse is the zero value — a request that named no status at all.
	T.Run("unspecified status", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.ListReportsByStatus(ctx, &issuereportspb.ListReportsByStatusRequest{
			Status: issuereportspb.ReportStatus_REPORT_STATUS_UNSPECIFIED,
		})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ListReportsByStatus(ctx, &issuereportspb.ListReportsByStatusRequest{})
		assert.Error(t, err)
	})
}

func TestIssueReports_ListingBySubject(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		input := creationInputForTest()
		input.SubjectType = "recipes"

		expected := []string{}
		for range exampleQuantity {
			created, err := testClient.CreateReport(ctx, &issuereportspb.CreateReportRequest{Input: input})
			require.NoError(t, err)
			expected = append(expected, created.GetResult().GetId())
		}

		byType, err := testClient.ListReportsBySubjectType(ctx, &issuereportspb.ListReportsBySubjectTypeRequest{
			SubjectType: "recipes",
		})
		require.NoError(t, err)
		assert.Len(t, byType.GetResults(), len(expected))

		// Same index, one column further in: every one of those reports names the
		// same subject.
		forSubject, err := testClient.ListReportsForSubject(ctx, &issuereportspb.ListReportsForSubjectRequest{
			SubjectType: "recipes",
			SubjectId:   input.GetSubjectId(),
		})
		require.NoError(t, err)
		assert.Len(t, forSubject.GetResults(), len(expected))

		none, err := testClient.ListReportsForSubject(ctx, &issuereportspb.ListReportsForSubjectRequest{
			SubjectType: "recipes",
			SubjectId:   nonexistentID,
		})
		require.NoError(t, err)
		assert.Empty(t, none.GetResults())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ListReportsBySubjectType(ctx, &issuereportspb.ListReportsBySubjectTypeRequest{})
		require.Error(t, err)

		_, err = c.ListReportsForSubject(ctx, &issuereportspb.ListReportsForSubjectRequest{})
		assert.Error(t, err)
	})
}

func TestIssueReports_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		const newDetails = "Updated details about the issue"

		updated, err := testClient.UpdateReport(ctx, &issuereportspb.UpdateReportRequest{
			ReportId: created.GetId(),
			Input: &issuereportspb.IssueReportUpdateInput{
				Kind:        created.GetKind(),
				Details:     newDetails,
				SubjectType: created.GetSubjectType(),
				SubjectId:   created.GetSubjectId(),
			},
		})
		require.NoError(t, err)
		assert.Equal(t, newDetails, updated.GetResult().GetDetails())

		// Restated rather than omitted, because the input is a replacement. The fields
		// the client sent back unchanged are unchanged; see the case below for what an
		// omission does.
		assert.Equal(t, created.GetKind(), updated.GetResult().GetKind())
		assert.Equal(t, created.GetSubjectType(), updated.GetResult().GetSubjectType())

		// The status is not on the update input at all, which is what keeps a revision
		// from being a triage decision: moving a report is TransitionReport's job.
		assert.Equal(t, created.GetStatus(), updated.GetResult().GetStatus())

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "issue_reports", RelevantID: created.GetId()},
			{EventType: "updated", ResourceType: "issue_reports", RelevantID: created.GetId()},
		})
	})

	// The behaviour a client has to know about. The local RPC took *string fields and
	// merged, so a partial update left the rest alone; platform's takes plain strings and
	// writes all four. A client that sends one field clears the other three.
	T.Run("an omitted field is cleared rather than kept", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)
		require.NotEmpty(t, created.GetKind())

		updated, err := testClient.UpdateReport(ctx, &issuereportspb.UpdateReportRequest{
			ReportId: created.GetId(),
			Input:    &issuereportspb.IssueReportUpdateInput{Details: "only the details"},
		})
		require.NoError(t, err)

		assert.Equal(t, "only the details", updated.GetResult().GetDetails())
		assert.Empty(t, updated.GetResult().GetKind(), "an omitted kind survived a replacement")
		assert.Empty(t, updated.GetResult().GetSubjectType())
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.UpdateReport(ctx, &issuereportspb.UpdateReportRequest{
			ReportId: nonexistentID,
			Input:    &issuereportspb.IssueReportUpdateInput{Details: "Updated details"},
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateReport(ctx, &issuereportspb.UpdateReportRequest{})
		assert.Error(t, err)
	})
}

// TestIssueReports_Triaging walks the lifecycle end to end over the wire.
func TestIssueReports_Triaging(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		acknowledged, err := testClient.TransitionReport(ctx, &issuereportspb.TransitionReportRequest{
			ReportId:       created.GetId(),
			ExpectedStatus: issuereportspb.ReportStatus_REPORT_STATUS_OPEN,
			TargetStatus:   issuereportspb.ReportStatus_REPORT_STATUS_ACKNOWLEDGED,
		})
		require.NoError(t, err)
		assert.Equal(t, issuereportspb.ReportStatus_REPORT_STATUS_ACKNOWLEDGED, acknowledged.GetResult().GetStatus())
		assert.Nil(t, acknowledged.GetResult().GetClosedAt())

		resolved, err := testClient.TransitionReport(ctx, &issuereportspb.TransitionReportRequest{
			ReportId:       created.GetId(),
			ExpectedStatus: issuereportspb.ReportStatus_REPORT_STATUS_ACKNOWLEDGED,
			TargetStatus:   issuereportspb.ReportStatus_REPORT_STATUS_RESOLVED,
			Resolution:     "fixed in the next release",
		})
		require.NoError(t, err)
		assert.Equal(t, issuereportspb.ReportStatus_REPORT_STATUS_RESOLVED, resolved.GetResult().GetStatus())
		assert.Equal(t, "fixed in the next release", resolved.GetResult().GetResolution())
		assert.NotNil(t, resolved.GetResult().GetClosedAt())

		// Reopening clears the closure, because a reason that no longer holds is
		// worse than none.
		reopened, err := testClient.TransitionReport(ctx, &issuereportspb.TransitionReportRequest{
			ReportId:       created.GetId(),
			ExpectedStatus: issuereportspb.ReportStatus_REPORT_STATUS_RESOLVED,
			TargetStatus:   issuereportspb.ReportStatus_REPORT_STATUS_OPEN,
		})
		require.NoError(t, err)
		assert.Equal(t, issuereportspb.ReportStatus_REPORT_STATUS_OPEN, reopened.GetResult().GetStatus())
		assert.Empty(t, reopened.GetResult().GetResolution())
		assert.Nil(t, reopened.GetResult().GetClosedAt())
	})

	// The guard is what makes this a queue two people can work.
	T.Run("stale expected status", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		_, err := testClient.TransitionReport(ctx, &issuereportspb.TransitionReportRequest{
			ReportId:       created.GetId(),
			ExpectedStatus: issuereportspb.ReportStatus_REPORT_STATUS_OPEN,
			TargetStatus:   issuereportspb.ReportStatus_REPORT_STATUS_RESOLVED,
			Resolution:     "first",
		})
		require.NoError(t, err)

		_, err = testClient.TransitionReport(ctx, &issuereportspb.TransitionReportRequest{
			ReportId:       created.GetId(),
			ExpectedStatus: issuereportspb.ReportStatus_REPORT_STATUS_OPEN,
			TargetStatus:   issuereportspb.ReportStatus_REPORT_STATUS_RESOLVED,
			Resolution:     "second",
		})
		require.Error(t, err)
		assert.Equal(t, codes.Aborted, status.Code(err))

		// The first note stands.
		retrieved, err := testClient.GetReport(ctx, &issuereportspb.GetReportRequest{ReportId: created.GetId()})
		require.NoError(t, err)
		assert.Equal(t, "first", retrieved.GetResult().GetResolution())
	})

	T.Run("a move the lifecycle refuses", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		_, err := testClient.TransitionReport(ctx, &issuereportspb.TransitionReportRequest{
			ReportId:       created.GetId(),
			ExpectedStatus: issuereportspb.ReportStatus_REPORT_STATUS_OPEN,
			TargetStatus:   issuereportspb.ReportStatus_REPORT_STATUS_OPEN,
		})
		require.Error(t, err)
		assert.Equal(t, codes.FailedPrecondition, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.TransitionReport(ctx, &issuereportspb.TransitionReportRequest{})
		assert.Error(t, err)
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

	T.Run("nonexistentID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.ArchiveReport(ctx, &issuereportspb.ArchiveReportRequest{ReportId: nonexistentID})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveReport(ctx, &issuereportspb.ArchiveReportRequest{})
		assert.Error(t, err)
	})
}

// TestIssueReports_Commenting is gone with the AddCommentToIssueReport RPC.
//
// That RPC named the target from its own name and forwarded to CreateComment;
// the existence check it appeared to add was the comment store's, through the
// target catalog. Commenting on an issue report is now platform's
// CommentsService.CreateComment with an issue-report target, and it is covered
// where every other comment path is.
