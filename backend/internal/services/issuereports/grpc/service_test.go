package grpc

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"

	commentsmock "github.com/primandproper/platform-go/v13/comments/mock"
	"github.com/primandproper/platform-go/v13/fake"
	issuereportsmock "github.com/primandproper/platform-go/v13/issuereports/mock"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
)

// buildTestService builds a service backed by the given store mocks. A nil mock gets an
// unconfigured one, which panics if any of its methods are called — so a test that
// reaches a store it did not arrange fails loudly rather than on a nil result.
func buildTestService(t *testing.T, reports *issuereportsmock.StoreMock, comments *commentsmock.StoreMock) *serviceImpl {
	t.Helper()

	if reports == nil {
		reports = &issuereportsmock.StoreMock{}
	}

	if comments == nil {
		comments = &commentsmock.StoreMock{}
	}

	return &serviceImpl{
		tracer:       tracing.NewTracerForTest(t.Name()),
		logger:       loggingnoop.NewLogger(),
		issueReports: reports,
		comments:     comments,
	}
}

// sessionContextFor returns a context carrying session data for the given user and account.
func sessionContextFor(t *testing.T, userID, accountID string) context.Context {
	t.Helper()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		ActiveAccountID: accountID,
		Requester:       sessions.RequesterInfo{UserID: userID},
	})
}

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service := NewService(
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			&issuereportsmock.StoreMock{},
			&commentsmock.StoreMock{},
		)

		assert.NotNil(t, service)
		assert.Implements(t, (*issuereportssvc.IssueReportsServiceServer)(nil), service)
	})
}

func TestProvideMethodPermissions(t *testing.T) {
	t.Parallel()

	t.Run("covers every method the service serves", func(t *testing.T) {
		t.Parallel()

		permissions := ProvideMethodPermissions()

		// A method with no entry is refused by the interceptor, so an RPC added
		// without one is an RPC nobody can call. Listing them here is what makes that
		// a test failure rather than a support ticket.
		for _, method := range []string{
			issuereportssvc.IssueReportsService_CreateIssueReport_FullMethodName,
			issuereportssvc.IssueReportsService_GetIssueReport_FullMethodName,
			issuereportssvc.IssueReportsService_GetIssueReports_FullMethodName,
			issuereportssvc.IssueReportsService_GetIssueReportsByStatus_FullMethodName,
			issuereportssvc.IssueReportsService_GetIssueReportsBySubjectType_FullMethodName,
			issuereportssvc.IssueReportsService_GetIssueReportsForSubject_FullMethodName,
			issuereportssvc.IssueReportsService_UpdateIssueReport_FullMethodName,
			issuereportssvc.IssueReportsService_TransitionIssueReport_FullMethodName,
			issuereportssvc.IssueReportsService_ArchiveIssueReport_FullMethodName,
			issuereportssvc.IssueReportsService_AddCommentToIssueReport_FullMethodName,
		} {
			assert.NotEmpty(t, permissions[method], method)
		}
	})
}

// buildSessionContextForTest returns a context carrying session data for an arbitrary
// user and account.
func buildSessionContextForTest(t *testing.T) (ctx context.Context, userID, accountID string) {
	t.Helper()

	userID, accountID = fake.BuildFakeID(), fake.BuildFakeID()

	return sessionContextFor(t, userID, accountID), userID, accountID
}
