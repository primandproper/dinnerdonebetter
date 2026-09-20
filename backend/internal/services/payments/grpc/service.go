package grpc

import (
	"context"

	paymentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/errors"

	"github.com/primandproper/platform-go/v14/billing"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
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
	db      database.Client
	logger  logging.Logger
	billing billing.Store
}

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	db database.Client,
	billingStore billing.Store,
) paymentssvc.PaymentsServiceServer {
	return &serviceImpl{
		logger:  logging.NewNamedLogger(logger, o11yName),
		tracer:  tracing.NewNamedTracer(tracerProvider, o11yName),
		db:      db,
		billing: billingStore,
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
