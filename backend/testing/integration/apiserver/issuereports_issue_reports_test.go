package integration

import (
	"testing"

	issuereportfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/primandproper/platform-go/v15/identity/identitypb"
	platformissuereports "github.com/primandproper/platform-go/v15/issuereports"
	issuereportspb "github.com/primandproper/platform-go/v15/issuereports/issuereportspb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The issue reports surface's behavior is asserted by platform's issuereports conformance suite,
// run against this deployment in conformance_test.go. What remains here is the audit entries this
// application's store records, and who works the queue — which is this deployment's policy rather
// than anything platform's suite can know.

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

	// Revising a report is a service admin's: the reporter files it, and the queue it lands in is
	// the service's.
	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		_, err := adminClient.UpdateReport(ctx, &issuereportspb.UpdateReportRequest{
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

	// Archiving takes a report out of the service's queue, so it is a service admin's too.
	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, testClient)

		_, err := adminClient.ArchiveReport(ctx, &issuereportspb.ArchiveReportRequest{ReportId: created.GetId()})
		require.NoError(t, err)

		AssertAuditLogContainsFuzzy(t, ctx, testClient, getAccountIDForTest(t, testClient), 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: platformissuereports.ResourceTypeReport, RelevantID: created.GetId()},
			{EventType: "archived", ResourceType: platformissuereports.ResourceTypeReport, RelevantID: created.GetId()},
		})
	})
}

// TestIssueReports_TheQueueIsTheServices pins who works the queue. A report is between the
// person who filed it and the service's administrators — "this user is harassing me" may well be
// about somebody in the reporter's own household — so the household's admin is refused every
// part of it, and a service admin can do all of it.
func TestIssueReports_TheQueueIsTheServices(T *testing.T) {
	T.Parallel()

	T.Run("a household admin can neither page nor transition a fellow member's report", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, householdAdminClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, householdAdminClient)

		memberInput := buildUserRegistrationInputForTest(t)
		member, memberClient := createUserAndClientForTestWithRegistrationInput(t, memberInput)

		invitation := inviteForTest(t, selfIDForTest(t, householdAdminClient), accountID, memberInput.GetUser().GetEmailAddress())
		_, err := memberClient.IdentityService().AcceptInvitation(ctx, &identitypb.AcceptInvitationRequest{
			InvitationId: invitation.ID,
			Token:        invitation.Token,
		})
		require.NoError(t, err)

		// A new token in the household, so the member files as a member of it.
		memberClient, err = buildAuthedGRPCClient(ctx, fetchLoginTokenForUserForTest(t, member))
		require.NoError(t, err)
		_, err = memberClient.IdentityService().SetDefaultAccount(ctx, &identitypb.SetDefaultAccountRequest{AccountId: accountID})
		require.NoError(t, err)

		created := createIssueReportForTest(t, memberClient)

		_, err = householdAdminClient.ListReports(ctx, &issuereportspb.ListReportsRequest{})
		assert.Equal(t, codes.PermissionDenied, status.Code(err), "paging the queue: %v", err)

		_, err = householdAdminClient.ListReportsForSubject(ctx, &issuereportspb.ListReportsForSubjectRequest{
			SubjectType: created.GetSubjectType(),
			SubjectId:   created.GetSubjectId(),
		})
		assert.Equal(t, codes.PermissionDenied, status.Code(err), "paging by subject: %v", err)

		_, err = householdAdminClient.TransitionReport(ctx, &issuereportspb.TransitionReportRequest{
			ReportId:       created.GetId(),
			ExpectedStatus: issuereportspb.ReportStatus_REPORT_STATUS_OPEN,
			TargetStatus:   issuereportspb.ReportStatus_REPORT_STATUS_DECLINED,
			Resolution:     t.Name(),
		})
		assert.Equal(t, codes.PermissionDenied, status.Code(err), "transitioning: %v", err)

		// Nor can they read it, or page what the member filed: those are the reporter's and a
		// service admin's.
		_, err = householdAdminClient.GetReport(ctx, &issuereportspb.GetReportRequest{ReportId: created.GetId()})
		require.Error(t, err)

		_, err = householdAdminClient.ListReportsByReporter(ctx, &issuereportspb.ListReportsByReporterRequest{Reporter: member.ID})
		require.Error(t, err)

		// And the report is still open, untouched by any of it.
		fetched, err := memberClient.GetReport(ctx, &issuereportspb.GetReportRequest{ReportId: created.GetId()})
		require.NoError(t, err)
		assert.Equal(t, issuereportspb.ReportStatus_REPORT_STATUS_OPEN, fetched.GetResult().GetStatus())
	})

	T.Run("a service admin can both page and transition anybody's report", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, reporterClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, reporterClient)

		_, err := adminClient.ListReports(ctx, &issuereportspb.ListReportsRequest{})
		require.NoError(t, err)

		// The whole queue is every test's reports, so the one this test filed is found by what it
		// is about, which nobody else's is.
		page, err := adminClient.ListReportsForSubject(ctx, &issuereportspb.ListReportsForSubjectRequest{
			SubjectType: created.GetSubjectType(),
			SubjectId:   created.GetSubjectId(),
		})
		require.NoError(t, err)
		require.Len(t, page.GetResults(), 1)
		assert.Equal(t, created.GetId(), page.GetResults()[0].GetId())

		fetched, err := adminClient.GetReport(ctx, &issuereportspb.GetReportRequest{ReportId: created.GetId()})
		require.NoError(t, err)
		assert.Equal(t, created.GetId(), fetched.GetResult().GetId())

		transitioned, err := adminClient.TransitionReport(ctx, &issuereportspb.TransitionReportRequest{
			ReportId:       created.GetId(),
			ExpectedStatus: issuereportspb.ReportStatus_REPORT_STATUS_OPEN,
			TargetStatus:   issuereportspb.ReportStatus_REPORT_STATUS_RESOLVED,
			Resolution:     t.Name(),
		})
		require.NoError(t, err)
		assert.Equal(t, issuereportspb.ReportStatus_REPORT_STATUS_RESOLVED, transitioned.GetResult().GetStatus())

		// The reporter reads the outcome back.
		read, err := reporterClient.GetReport(ctx, &issuereportspb.GetReportRequest{ReportId: created.GetId()})
		require.NoError(t, err)
		assert.Equal(t, issuereportspb.ReportStatus_REPORT_STATUS_RESOLVED, read.GetResult().GetStatus())
	})

	T.Run("a reporter cannot work their own report", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, reporterClient := createUserAndClientForTest(t)
		created := createIssueReportForTest(t, reporterClient)

		_, err := reporterClient.TransitionReport(ctx, &issuereportspb.TransitionReportRequest{
			ReportId:       created.GetId(),
			ExpectedStatus: issuereportspb.ReportStatus_REPORT_STATUS_OPEN,
			TargetStatus:   issuereportspb.ReportStatus_REPORT_STATUS_RESOLVED,
		})
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		_, err = reporterClient.ArchiveReport(ctx, &issuereportspb.ArchiveReportRequest{ReportId: created.GetId()})
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})
}

// TestIssueReports_Commenting is gone with the AddCommentToIssueReport RPC.
//
// That RPC named the target from its own name and forwarded to CreateComment;
// the existence check it appeared to add was the comment store's, through the
// target catalog. Commenting on an issue report is now platform's
// CommentsService.CreateComment with an issue-report target, and it is covered
// where every other comment path is.
