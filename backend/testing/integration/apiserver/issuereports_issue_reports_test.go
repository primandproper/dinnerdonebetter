package integration

import (
	"testing"

	issuereportfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/fakes"
	commentsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/issuereports/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	issuereports "github.com/primandproper/platform-go/v13/issuereports"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func checkIssueReportEquality(t *testing.T, expected, actual *issuereportssvc.IssueReport) {
	t.Helper()

	assert.NotEmpty(t, actual.GetId(), "expected IssueReport to have ID")
	assert.NotNil(t, actual.GetCreatedAt(), "expected IssueReport to have CreatedAt")

	assert.Equal(t, expected.GetKind(), actual.GetKind(), "expected IssueReport Kind")
	assert.Equal(t, expected.GetDetails(), actual.GetDetails(), "expected IssueReport Details")
	assert.Equal(t, expected.GetSubjectType(), actual.GetSubjectType(), "expected IssueReport SubjectType")
	assert.Equal(t, expected.GetSubjectId(), actual.GetSubjectId(), "expected IssueReport SubjectID")
	assert.NotEmpty(t, actual.GetReporter(), "expected IssueReport to have Reporter")
}

// createIssueReportForTest files one report and reads it back.
func createIssueReportForTest(t *testing.T, testClient client.Client) *issuereportssvc.IssueReport {
	t.Helper()
	ctx := t.Context()

	input := grpcconverters.ConvertIssueReportToGRPCIssueReportCreationRequestInput(issuereportfakes.BuildFakeIssueReport())

	created, err := testClient.CreateIssueReport(ctx, &issuereportssvc.CreateIssueReportRequest{Input: input})
	require.NoError(t, err)
	require.NotNil(t, created.GetCreated())

	// A report is born open, whatever the client sent.
	assert.Equal(t, issuereports.StatusOpen.String(), created.GetCreated().GetStatus())
	assert.Empty(t, created.GetCreated().GetResolution())
	assert.Nil(t, created.GetCreated().GetClosedAt())

	retrieved, err := testClient.GetIssueReport(ctx, &issuereportssvc.GetIssueReportRequest{IssueReportId: created.GetCreated().GetId()})
	require.NoError(t, err)
	require.NotNil(t, retrieved.GetResult())
	checkIssueReportEquality(t, created.GetCreated(), retrieved.GetResult())

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

		_, err := c.CreateIssueReport(ctx, &issuereportssvc.CreateIssueReportRequest{})
		require.Error(t, err)
	})

	T.Run("invalid input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.CreateIssueReport(ctx, &issuereportssvc.CreateIssueReportRequest{
			Input: &issuereportssvc.IssueReportCreationRequestInput{Kind: "", Details: ""},
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

		retrieved, err := testClient.GetIssueReport(ctx, &issuereportssvc.GetIssueReportRequest{IssueReportId: created.GetId()})
		require.NoError(t, err)
		assert.NotNil(t, retrieved)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		retrieved, err := testClient.GetIssueReport(ctx, &issuereportssvc.GetIssueReportRequest{IssueReportId: nonexistentID})
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

		_, err := otherClient.GetIssueReport(ctx, &issuereportssvc.GetIssueReportRequest{IssueReportId: created.GetId()})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
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

		expected := []string{}
		for range exampleQuantity {
			expected = append(expected, createIssueReportForTest(t, testClient).GetId())
		}

		results, err := testClient.GetIssueReports(ctx, &issuereportssvc.GetIssueReportsRequest{})
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
		_, err := c.GetIssueReports(ctx, &issuereportssvc.GetIssueReportsRequest{})
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

		open, err := testClient.GetIssueReportsByStatus(ctx, &issuereportssvc.GetIssueReportsByStatusRequest{
			Status: issuereports.StatusOpen.String(),
		})
		require.NoError(t, err)
		require.Len(t, open.GetResults(), 1)
		assert.Equal(t, created.GetId(), open.GetResults()[0].GetId())

		resolved, err := testClient.GetIssueReportsByStatus(ctx, &issuereportssvc.GetIssueReportsByStatusRequest{
			Status: issuereports.StatusResolved.String(),
		})
		require.NoError(t, err)
		assert.Empty(t, resolved.GetResults())
	})

	T.Run("unknown status", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.GetIssueReportsByStatus(ctx, &issuereportssvc.GetIssueReportsByStatusRequest{Status: "closed"})
		require.Error(t, err)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetIssueReportsByStatus(ctx, &issuereportssvc.GetIssueReportsByStatusRequest{})
		assert.Error(t, err)
	})
}

func TestIssueReports_ListingBySubject(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		example := issuereportfakes.BuildFakeIssueReport()
		example.SubjectType = "recipes"

		expected := []string{}
		for range exampleQuantity {
			input := grpcconverters.ConvertIssueReportToGRPCIssueReportCreationRequestInput(example)

			created, err := testClient.CreateIssueReport(ctx, &issuereportssvc.CreateIssueReportRequest{Input: input})
			require.NoError(t, err)
			expected = append(expected, created.GetCreated().GetId())
		}

		byType, err := testClient.GetIssueReportsBySubjectType(ctx, &issuereportssvc.GetIssueReportsBySubjectTypeRequest{
			SubjectType: "recipes",
		})
		require.NoError(t, err)
		assert.Len(t, byType.GetResults(), len(expected))

		// Same index, one column further in: every one of those reports names the
		// same subject.
		forSubject, err := testClient.GetIssueReportsForSubject(ctx, &issuereportssvc.GetIssueReportsForSubjectRequest{
			SubjectType: "recipes",
			SubjectId:   example.SubjectID,
		})
		require.NoError(t, err)
		assert.Len(t, forSubject.GetResults(), len(expected))

		none, err := testClient.GetIssueReportsForSubject(ctx, &issuereportssvc.GetIssueReportsForSubjectRequest{
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

		_, err := c.GetIssueReportsBySubjectType(ctx, &issuereportssvc.GetIssueReportsBySubjectTypeRequest{})
		require.Error(t, err)

		_, err = c.GetIssueReportsForSubject(ctx, &issuereportssvc.GetIssueReportsForSubjectRequest{})
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

		newDetails := "Updated details about the issue"
		updated, err := testClient.UpdateIssueReport(ctx, &issuereportssvc.UpdateIssueReportRequest{
			IssueReportId: created.GetId(),
			Input:         &issuereportssvc.IssueReportUpdateRequestInput{Details: &newDetails},
		})
		require.NoError(t, err)
		assert.Equal(t, newDetails, updated.GetUpdated().GetDetails())

		// Everything the client did not send survives the revision.
		assert.Equal(t, created.GetKind(), updated.GetUpdated().GetKind())
		assert.Equal(t, created.GetStatus(), updated.GetUpdated().GetStatus())

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "issue_reports", RelevantID: created.GetId()},
			{EventType: "updated", ResourceType: "issue_reports", RelevantID: created.GetId()},
		})
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		newDetails := "Updated details"
		_, err := testClient.UpdateIssueReport(ctx, &issuereportssvc.UpdateIssueReportRequest{
			IssueReportId: nonexistentID,
			Input:         &issuereportssvc.IssueReportUpdateRequestInput{Details: &newDetails},
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateIssueReport(ctx, &issuereportssvc.UpdateIssueReportRequest{})
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

		acknowledged, err := testClient.TransitionIssueReport(ctx, &issuereportssvc.TransitionIssueReportRequest{
			IssueReportId: created.GetId(),
			FromStatus:    issuereports.StatusOpen.String(),
			ToStatus:      issuereports.StatusAcknowledged.String(),
		})
		require.NoError(t, err)
		assert.Equal(t, issuereports.StatusAcknowledged.String(), acknowledged.GetResult().GetStatus())
		assert.Nil(t, acknowledged.GetResult().GetClosedAt())

		resolved, err := testClient.TransitionIssueReport(ctx, &issuereportssvc.TransitionIssueReportRequest{
			IssueReportId: created.GetId(),
			FromStatus:    issuereports.StatusAcknowledged.String(),
			ToStatus:      issuereports.StatusResolved.String(),
			Resolution:    "fixed in the next release",
		})
		require.NoError(t, err)
		assert.Equal(t, issuereports.StatusResolved.String(), resolved.GetResult().GetStatus())
		assert.Equal(t, "fixed in the next release", resolved.GetResult().GetResolution())
		assert.NotNil(t, resolved.GetResult().GetClosedAt())

		// Reopening clears the closure, because a reason that no longer holds is
		// worse than none.
		reopened, err := testClient.TransitionIssueReport(ctx, &issuereportssvc.TransitionIssueReportRequest{
			IssueReportId: created.GetId(),
			FromStatus:    issuereports.StatusResolved.String(),
			ToStatus:      issuereports.StatusOpen.String(),
		})
		require.NoError(t, err)
		assert.Equal(t, issuereports.StatusOpen.String(), reopened.GetResult().GetStatus())
		assert.Empty(t, reopened.GetResult().GetResolution())
		assert.Nil(t, reopened.GetResult().GetClosedAt())
	})

	// The guard is what makes this a queue two people can work.
	T.Run("stale from status", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		_, err := testClient.TransitionIssueReport(ctx, &issuereportssvc.TransitionIssueReportRequest{
			IssueReportId: created.GetId(),
			FromStatus:    issuereports.StatusOpen.String(),
			ToStatus:      issuereports.StatusResolved.String(),
			Resolution:    "first",
		})
		require.NoError(t, err)

		_, err = testClient.TransitionIssueReport(ctx, &issuereportssvc.TransitionIssueReportRequest{
			IssueReportId: created.GetId(),
			FromStatus:    issuereports.StatusOpen.String(),
			ToStatus:      issuereports.StatusResolved.String(),
			Resolution:    "second",
		})
		require.Error(t, err)
		assert.Equal(t, codes.Aborted, status.Code(err))

		// The first note stands.
		retrieved, err := testClient.GetIssueReport(ctx, &issuereportssvc.GetIssueReportRequest{IssueReportId: created.GetId()})
		require.NoError(t, err)
		assert.Equal(t, "first", retrieved.GetResult().GetResolution())
	})

	T.Run("a move the lifecycle refuses", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		_, err := testClient.TransitionIssueReport(ctx, &issuereportssvc.TransitionIssueReportRequest{
			IssueReportId: created.GetId(),
			FromStatus:    issuereports.StatusOpen.String(),
			ToStatus:      issuereports.StatusOpen.String(),
		})
		require.Error(t, err)
		assert.Equal(t, codes.FailedPrecondition, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.TransitionIssueReport(ctx, &issuereportssvc.TransitionIssueReportRequest{})
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

		_, err := testClient.ArchiveIssueReport(ctx, &issuereportssvc.ArchiveIssueReportRequest{IssueReportId: created.GetId()})
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

		_, err := testClient.ArchiveIssueReport(ctx, &issuereportssvc.ArchiveIssueReportRequest{IssueReportId: nonexistentID})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveIssueReport(ctx, &issuereportssvc.ArchiveIssueReportRequest{})
		assert.Error(t, err)
	})
}

// TestIssueReports_Commenting pins that the comment path reads the report as the
// caller first: it is both the existence check the comment store's target catalog
// cannot run for this type and the account boundary.
func TestIssueReports_Commenting(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		res, err := testClient.AddCommentToIssueReport(ctx, &issuereportssvc.AddCommentToIssueReportRequest{
			IssueReportId: created.GetId(),
			Input:         &commentsgrpc.CommentCreationRequestInput{Body: "this is a duplicate"},
		})
		require.NoError(t, err)
		require.NotNil(t, res.GetComment())
		assert.Equal(t, created.GetId(), res.GetComment().GetTarget().GetId())
		assert.Equal(t, user.ID, res.GetComment().GetAuthor())
	})

	T.Run("another account's report is not found", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, ownerClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, ownerClient)

		_, otherClient := createUserAndClientForTest(t)

		_, err := otherClient.AddCommentToIssueReport(ctx, &issuereportssvc.AddCommentToIssueReportRequest{
			IssueReportId: created.GetId(),
			Input:         &commentsgrpc.CommentCreationRequestInput{Body: "nope"},
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("nonexistent report", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.AddCommentToIssueReport(ctx, &issuereportssvc.AddCommentToIssueReportRequest{
			IssueReportId: nonexistentID,
			Input:         &commentsgrpc.CommentCreationRequestInput{Body: "nope"},
		})
		require.Error(t, err)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})
}
