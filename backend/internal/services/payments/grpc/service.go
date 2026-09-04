package grpc

import (
	paymentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/errors"

	"github.com/primandproper/platform-go/v13/billing"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	o11yName = "payments_service"
)

var _ paymentssvc.PaymentsServiceServer = (*serviceImpl)(nil)

// serviceImpl serves the stored half of payments — the catalog, an account's
// subscriptions, purchases and ledger — straight off the billing store.
//
// There is no manager between them. The store owns every rule about what a row
// may hold, and this application has nothing to add on the way in beyond which
// scope the request is for and whose account it may read. What the payments
// manager still does — turning a provider's event into a subscription's standing
// — has no RPC, because a provider is not a client.
type serviceImpl struct {
	paymentssvc.UnimplementedPaymentsServiceServer
	tracer  tracing.Tracer
	logger  logging.Logger
	billing billing.Store
}

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	billingStore billing.Store,
) paymentssvc.PaymentsServiceServer {
	return &serviceImpl{
		logger:  logging.NewNamedLogger(logger, o11yName),
		tracer:  tracing.NewNamedTracer(tracerProvider, o11yName),
		billing: billingStore,
	}
}
