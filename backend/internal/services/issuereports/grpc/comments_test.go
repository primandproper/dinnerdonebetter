package grpc

import (
	"context"
	"testing"

	ddbcomments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/fakes"
	commentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"

	comments "github.com/primandproper/platform-go/v13/comments"
	commentsmock "github.com/primandproper/platform-go/v13/comments/mock"
	"github.com/primandproper/platform-go/v13/fake"
	issuereports "github.com/primandproper/platform-go/v13/issuereports"
	issuereportsmock "github.com/primandproper/platform-go/v13/issuereports/mock"
	"github.com/primandproper/platform-go/v13/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
)

func TestServiceImpl_AddCommentToIssueReport(t *testing.T) {
	t.Parallel()

	t.Run("files the comment against the report", func(t *testing.T) {
		t.Parallel()

		ctx, userID, accountID := buildSessionContextForTest(t)
		report := issuereportfakes.BuildFakeIssueReportForScope(accountID)
		body := "this one is a duplicate of the one filed yesterday"

		reports := &issuereportsmock.StoreMock{
			GetReportFunc: func(context.Context, tenancy.Scope, string) (*issuereports.Report, error) {
				return report, nil
			},
		}

		var written *comments.Comment
		store := &commentsmock.StoreMock{
			CreateCommentFunc: func(_ context.Context, comment *comments.Comment) error {
				comment.ID = fake.BuildFakeID()
				written = comment

				return nil
			},
		}

		res, err := buildTestService(t, reports, store).AddCommentToIssueReport(ctx, &issuereportssvc.AddCommentToIssueReportRequest{
			IssueReportId: report.ID,
			Input:         &commentssvc.CommentCreationRequestInput{Body: body},
		})
		require.NoError(t, err)
		require.NotNil(t, res.GetComment())

		// The target comes from the request path, never from the body: the URL says
		// which report this is about, and a body that could say otherwise is a body
		// that could file a comment against something else entirely.
		assert.Equal(t, ddbissuereports.CommentTargetTypeIssueReports, written.Target.Type)
		assert.Equal(t, report.ID, written.Target.ID)
		assert.Equal(t, userID, written.Author)
		assert.Equal(t, body, written.Body)

		// Comments are filed globally while issue reports are filed per account —
		// which is exactly why the report is read here rather than by the comment
		// store's target catalog.
		assert.Equal(t, ddbcomments.Scope(), written.Scope)
	})

	// The read is the authorization check as well as the existence check. A report
	// in another account is not there, so no comment can be filed against it.
	t.Run("refuses a report the caller cannot see", func(t *testing.T) {
		t.Parallel()

		ctx, _, _ := buildSessionContextForTest(t)

		reports := &issuereportsmock.StoreMock{
			GetReportFunc: func(context.Context, tenancy.Scope, string) (*issuereports.Report, error) {
				return nil, issuereports.ErrReportNotFound
			},
		}

		// The comment store is left unconfigured on purpose: reaching it would panic,
		// which is how this test proves the write never happened.
		_, err := buildTestService(t, reports, nil).AddCommentToIssueReport(ctx, &issuereportssvc.AddCommentToIssueReportRequest{
			IssueReportId: fake.BuildFakeID(),
			Input:         &commentssvc.CommentCreationRequestInput{Body: "hello"},
		})
		assertCode(t, err, codes.NotFound)
	})

	t.Run("requires a session", func(t *testing.T) {
		t.Parallel()

		_, err := buildTestService(t, nil, nil).AddCommentToIssueReport(t.Context(), &issuereportssvc.AddCommentToIssueReportRequest{})
		assertCode(t, err, codes.Unauthenticated)
	})
}
