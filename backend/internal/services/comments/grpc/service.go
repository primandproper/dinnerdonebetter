package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	commentsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/manager"
	commentssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/comments"

	"github.com/primandproper/platform-go/v3/observability/logging"
	"github.com/primandproper/platform-go/v3/observability/tracing"
)

const (
	o11yName = "comments_service"
)

var _ commentssvc.CommentsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		commentssvc.UnimplementedCommentsServiceServer
		tracer                    tracing.Tracer
		logger                    logging.Logger
		sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error)
		commentsManager           commentsmanager.CommentsDataManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	commentsManager commentsmanager.CommentsDataManager,
) commentssvc.CommentsServiceServer {
	return &serviceImpl{
		logger:                    logging.NewNamedLogger(logger, o11yName),
		tracer:                    tracing.NewNamedTracer(tracerProvider, o11yName),
		sessionContextDataFetcher: sessions.FetchContextDataFromContext,
		commentsManager:           commentsManager,
	}
}
