package authorization

import (
	issuereportsgrpc "github.com/primandproper/platform-go/v14/issuereports/grpc"
)

// The issue report permissions are platform's, re-exported under the names this
// application's policy already spells. See comments_permissions.go.
const (
	// CreateIssueReportsPermission is a permission.
	CreateIssueReportsPermission = issuereportsgrpc.PermissionFileReports
	// ReadIssueReportsPermission is a permission.
	ReadIssueReportsPermission = issuereportsgrpc.PermissionReadReports
	// TriageIssueReportsPermission gates the administrative reads.
	TriageIssueReportsPermission = issuereportsgrpc.PermissionTriageReports
	// UpdateIssueReportsPermission is a permission.
	UpdateIssueReportsPermission = issuereportsgrpc.PermissionUpdateReports
	// TransitionIssueReportsPermission gates moving a report through triage.
	TransitionIssueReportsPermission = issuereportsgrpc.PermissionTransitionReports
	// ArchiveIssueReportsPermission is a permission.
	ArchiveIssueReportsPermission = issuereportsgrpc.PermissionArchiveReports
)
