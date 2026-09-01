package grpc

import (
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"

	comments "github.com/primandproper/platform-go/v13/comments"
	issuereports "github.com/primandproper/platform-go/v13/issuereports"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterIssueReportsService registers the issue reports gRPC service with the injector.
func RegisterIssueReportsService(i do.Injector) {
	do.Provide[IssueReportsMethodPermissions](i, func(i do.Injector) (IssueReportsMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[issuereportssvc.IssueReportsServiceServer](i, func(i do.Injector) (issuereportssvc.IssueReportsServiceServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[issuereports.Store](i),
			do.MustInvoke[comments.Store](i),
		), nil
	})
}
