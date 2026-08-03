package grpc

import (
	commentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/manager"
	commentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	o11yName = "comments_service"
)

var _ commentssvc.CommentsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		commentssvc.UnimplementedCommentsServiceServer
		tracer          tracing.Tracer
		logger          logging.Logger
		commentsManager commentsmanager.CommentsDataManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	commentsManager commentsmanager.CommentsDataManager,
) commentssvc.CommentsServiceServer {
	return &serviceImpl{
		logger:          logging.NewNamedLogger(logger, o11yName),
		tracer:          tracing.NewNamedTracer(tracerProvider, o11yName),
		commentsManager: commentsManager,
	}
}
