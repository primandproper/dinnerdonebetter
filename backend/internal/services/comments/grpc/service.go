package grpc

import (
	commentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/comments/errors"

	comments "github.com/primandproper/platform-go/v13/comments"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	o11yName = "comments_service"
)

var _ commentssvc.CommentsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		commentssvc.UnimplementedCommentsServiceServer
		tracer tracing.Tracer
		logger logging.Logger
		// comments is the store directly, with no manager between it and this
		// service. There is nothing left for one to do: validation, the target
		// catalog, the thread depth and the scope are all the store's, and a tier
		// that only forwarded would be a tier whose only effect was to make the
		// errors harder to attribute.
		comments comments.Store
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	commentStore comments.Store,
) commentssvc.CommentsServiceServer {
	return &serviceImpl{
		logger:   logging.NewNamedLogger(logger, o11yName),
		tracer:   tracing.NewNamedTracer(tracerProvider, o11yName),
		comments: commentStore,
	}
}
