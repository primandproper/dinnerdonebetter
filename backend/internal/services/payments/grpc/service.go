package grpc

import (
	paymentsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments/manager"
	paymentssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/payments"

	"github.com/primandproper/platform-go/v6/observability/logging"
	"github.com/primandproper/platform-go/v6/observability/tracing"
)

const (
	o11yName = "payments_service"
)

var _ paymentssvc.PaymentsServiceServer = (*serviceImpl)(nil)

type serviceImpl struct {
	paymentssvc.UnimplementedPaymentsServiceServer
	tracer          tracing.Tracer
	logger          logging.Logger
	paymentsManager paymentsmanager.PaymentsDataManager
}

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	paymentsManager paymentsmanager.PaymentsDataManager,
) paymentssvc.PaymentsServiceServer {
	return &serviceImpl{
		logger:          logging.NewNamedLogger(logger, o11yName),
		tracer:          tracing.NewNamedTracer(tracerProvider, o11yName),
		paymentsManager: paymentsManager,
	}
}
