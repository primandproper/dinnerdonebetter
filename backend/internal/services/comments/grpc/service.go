package grpc

import (
	"context"

	commentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/comments/errors"

	comments "github.com/primandproper/platform-go/v14/comments"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
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
		// db supplies the executor every store call now takes; see inTransaction.
		db database.Client
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
	db database.Client,
	commentStore comments.Store,
) commentssvc.CommentsServiceServer {
	return &serviceImpl{
		logger:   logging.NewNamedLogger(logger, o11yName),
		tracer:   tracing.NewNamedTracer(tracerProvider, o11yName),
		db:       db,
		comments: commentStore,
	}
}

// inTransaction runs one store write on a transaction of its own.
//
// As of platform-go v14 a store holds no database handle: a write takes the
// caller's database.Tx. Every write in this service is a single store call, so
// each gets one transaction — which is exactly what the store opened for itself
// before the caller was required to supply it. A handler that ever writes twice
// should take one transaction across both rather than call this twice.
func inTransaction[T any](ctx context.Context, db database.Client, write func(tx database.Tx) (T, error)) (T, error) {
	var out T

	err := db.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		out, writeErr = write(tx)

		return writeErr
	})

	return out, err
}
