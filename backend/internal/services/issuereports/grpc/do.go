package grpc

import (
	commentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/manager"
	issuereportsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/manager"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"

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
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[issuereportsmanager.IssueReportsDataManager](i),
			do.MustInvoke[commentsmanager.CommentsDataManager](i),
		), nil
	})
}
