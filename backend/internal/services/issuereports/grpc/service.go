package grpc

import (
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/issuereports/errors"

	comments "github.com/primandproper/platform-go/v13/comments"
	issuereports "github.com/primandproper/platform-go/v13/issuereports"
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
		tracer       tracing.Tracer
		logger       logging.Logger
		issueReports issuereports.Store
		comments     comments.Store
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	issueReports issuereports.Store,
	commentStore comments.Store,
) issuereportssvc.IssueReportsServiceServer {
	return &serviceImpl{
		logger:       logging.NewNamedLogger(logger, o11yName),
		tracer:       tracing.NewNamedTracer(tracerProvider, o11yName),
		issueReports: issueReports,
		comments:     commentStore,
	}
}
