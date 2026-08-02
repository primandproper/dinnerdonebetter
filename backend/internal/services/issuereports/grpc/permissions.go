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
		issuereportssvc.IssueReportsService_GetIssueReportsForAccount_FullMethodName: {
			authorization.ReadIssueReportsPermission,
		},
		issuereportssvc.IssueReportsService_GetIssueReportsForTable_FullMethodName: {
			authorization.ReadIssueReportsPermission,
		},
		issuereportssvc.IssueReportsService_GetIssueReportsForRecord_FullMethodName: {
			authorization.ReadIssueReportsPermission,
		},
		issuereportssvc.IssueReportsService_UpdateIssueReport_FullMethodName: {
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
