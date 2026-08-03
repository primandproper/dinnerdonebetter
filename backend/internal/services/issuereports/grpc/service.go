package grpc

import (
	commentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/manager"
	issuereportsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/manager"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
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
		commentsManager     commentsmanager.CommentsDataManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	issueReportsManager issuereportsmanager.IssueReportsDataManager,
	commentsManager commentsmanager.CommentsDataManager,
) issuereportssvc.IssueReportsServiceServer {
	return &serviceImpl{
		logger:              logging.NewNamedLogger(logger, o11yName),
		tracer:              tracing.NewNamedTracer(tracerProvider, o11yName),
		issueReportsManager: issueReportsManager,
		commentsManager:     commentsManager,
	}
}
