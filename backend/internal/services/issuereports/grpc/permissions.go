package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
)

// IssueReportsMethodPermissions is a named type for Wire dependency injection.
type IssueReportsMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the issue reports service's method permissions.
func ProvideMethodPermissions() IssueReportsMethodPermissions {
	return IssueReportsMethodPermissions{
		issuereportssvc.IssueReportsService_CreateIssueReport_FullMethodName: {
			authorization.CreateIssueReportsPermission,
		},
		issuereportssvc.IssueReportsService_GetIssueReport_FullMethodName: {
			authorization.ReadIssueReportsPermission,
		},
		issuereportssvc.IssueReportsService_GetIssueReports_FullMethodName: {
			authorization.ReadIssueReportsPermission,
		},
		issuereportssvc.IssueReportsService_GetIssueReportsByStatus_FullMethodName: {
			authorization.ReadIssueReportsPermission,
		},
		issuereportssvc.IssueReportsService_GetIssueReportsBySubjectType_FullMethodName: {
			authorization.ReadIssueReportsPermission,
		},
		issuereportssvc.IssueReportsService_GetIssueReportsForSubject_FullMethodName: {
			authorization.ReadIssueReportsPermission,
		},
		issuereportssvc.IssueReportsService_UpdateIssueReport_FullMethodName: {
			authorization.UpdateIssueReportsPermission,
		},
		// Triage is a write to the report and is gated by the same capability as a
		// revision. It is not a permission of its own because this application has no
		// role that may resolve a report but not edit one — the account admin who can
		// do either can do both, and a permission nothing distinguishes is a line in
		// the role grid that grants what the one beside it already granted.
		issuereportssvc.IssueReportsService_TransitionIssueReport_FullMethodName: {
			authorization.UpdateIssueReportsPermission,
		},
		issuereportssvc.IssueReportsService_ArchiveIssueReport_FullMethodName: {
			authorization.ArchiveIssueReportsPermission,
		},
		issuereportssvc.IssueReportsService_AddCommentToIssueReport_FullMethodName: {
			authorization.CreateCommentsPermission,
		},
	}
}
