package grpc

import (
	"context"

	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/waitlists/errors"

	waitlists "github.com/primandproper/platform-go/v14/waitlists"
	"github.com/primandproper/primitives-go/v2/clock"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "waitlists_service"
)

var _ waitlistssvc.WaitlistsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		waitlistssvc.UnimplementedWaitlistsServiceServer
		tracer    tracing.Tracer
		db        database.Client
		logger    logging.Logger
		waitlists waitlists.Store

		// clock decides whether a list is still open, for WaitlistIsOpen.
		//
		// It is the real clock, which is also the store's default — the two have to
		// agree, or a list this service calls open is one the store refuses a signup
		// for. A deployment that gives the store a clock of its own has to give this
		// one the same clock.
		clock clock.Clock
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	db database.Client,
	waitlistStore waitlists.Store,
) waitlistssvc.WaitlistsServiceServer {
	return &serviceImpl{
		logger:    logging.NewNamedLogger(logger, o11yName),
		tracer:    tracing.NewNamedTracer(tracerProvider, o11yName),
		db:        db,
		waitlists: waitlistStore,
		clock:     clock.NewClock(),
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
