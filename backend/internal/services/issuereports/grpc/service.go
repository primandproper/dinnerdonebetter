package grpc

import (
	issuereportsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/manager"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"

	comments "github.com/primandproper/platform-go/v13/comments"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	o11yName = "issue_reports_service"
)

var _ issuereportssvc.IssueReportsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		issuereportssvc.UnimplementedIssueReportsServiceServer
		tracer              tracing.Tracer
		logger              logging.Logger
		issueReportsManager issuereportsmanager.IssueReportsDataManager
		comments            comments.Store
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	issueReportsManager issuereportsmanager.IssueReportsDataManager,
	commentStore comments.Store,
) issuereportssvc.IssueReportsServiceServer {
	return &serviceImpl{
		logger:              logging.NewNamedLogger(logger, o11yName),
		tracer:              tracing.NewNamedTracer(tracerProvider, o11yName),
		issueReportsManager: issueReportsManager,
		comments:            commentStore,
	}
}
