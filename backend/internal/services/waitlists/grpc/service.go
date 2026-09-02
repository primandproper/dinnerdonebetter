package grpc

import (
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/waitlists/errors"

	"github.com/primandproper/platform-go/v13/clock"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	waitlists "github.com/primandproper/platform-go/v13/waitlists"
)

const (
	o11yName = "waitlists_service"
)

var _ waitlistssvc.WaitlistsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		waitlistssvc.UnimplementedWaitlistsServiceServer
		tracer    tracing.Tracer
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
	waitlistStore waitlists.Store,
) waitlistssvc.WaitlistsServiceServer {
	return &serviceImpl{
		logger:    logging.NewNamedLogger(logger, o11yName),
		tracer:    tracing.NewNamedTracer(tracerProvider, o11yName),
		waitlists: waitlistStore,
		clock:     clock.NewClock(),
	}
}
