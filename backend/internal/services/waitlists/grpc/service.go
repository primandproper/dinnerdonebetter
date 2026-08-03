package grpc

import (
	waitlistsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists/manager"
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	o11yName = "waitlists_service"
)

var _ waitlistssvc.WaitlistsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		waitlistssvc.UnimplementedWaitlistsServiceServer
		tracer           tracing.Tracer
		logger           logging.Logger
		waitlistsManager waitlistsmanager.WaitlistsDataManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	waitlistsManager waitlistsmanager.WaitlistsDataManager,
) waitlistssvc.WaitlistsServiceServer {
	return &serviceImpl{
		logger:           logging.NewNamedLogger(logger, o11yName),
		tracer:           tracing.NewNamedTracer(tracerProvider, o11yName),
		waitlistsManager: waitlistsManager,
	}
}
