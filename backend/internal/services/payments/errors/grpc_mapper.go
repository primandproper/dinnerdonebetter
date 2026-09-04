// Package errors maps the billing store's sentinels onto gRPC codes.
//
// Without it every one of them reaches a client as the handler's default, which
// is Internal — and "the server broke" is the wrong thing to tell somebody whose
// real problem is that they priced a product in a currency with four letters.
package errors

import (
	"errors"

	"github.com/primandproper/platform-go/v13/billing"
	"github.com/primandproper/platform-go/v13/errors/grpc"

	"google.golang.org/grpc/codes"
)

func init() {
	grpc.RegisterGRPCErrorMapper(billingGRPCMapper{})
}

type billingGRPCMapper struct{}

func (billingGRPCMapper) Map(err error) (code codes.Code, ok bool) {
	if err == nil {
		return codes.Unknown, false
	}

	switch {
	// No such live row. An archived row and one that never existed are the same
	// answer, deliberately.
	case errors.Is(err, billing.ErrProductNotFound),
		errors.Is(err, billing.ErrSubscriptionNotFound),
		errors.Is(err, billing.ErrPurchaseNotFound),
		errors.Is(err, billing.ErrTransactionNotFound):
		return codes.NotFound, true

	// Malformed input — the request could not have succeeded as written,
	// whatever the state of the database.
	case errors.Is(err, billing.ErrNilProduct),
		errors.Is(err, billing.ErrNilSubscription),
		errors.Is(err, billing.ErrNilPurchase),
		errors.Is(err, billing.ErrNilTransaction),
		errors.Is(err, billing.ErrEmptyProductName),
		errors.Is(err, billing.ErrEmptyAccount),
		errors.Is(err, billing.ErrEmptyProduct),
		errors.Is(err, billing.ErrEmptyExternalID),
		errors.Is(err, billing.ErrEmptyPeriod),
		errors.Is(err, billing.ErrBackwardsPeriod),
		errors.Is(err, billing.ErrInvalidKind),
		errors.Is(err, billing.ErrInvalidStatus),
		errors.Is(err, billing.ErrInvalidCurrency),
		errors.Is(err, billing.ErrNegativeAmount),
		errors.Is(err, billing.ErrEmptyBillingInterval),
		errors.Is(err, billing.ErrUnexpectedBillingInterval),
		errors.Is(err, billing.ErrAmbiguousTransaction):
		return codes.InvalidArgument, true

	// The provider-side id is already claimed in this scope. Over the API that is
	// an administrator entering an id twice; from a webhook it is the redelivery
	// the store exists to refuse, and the manager reads it before it gets here.
	case errors.Is(err, billing.ErrProductExists),
		errors.Is(err, billing.ErrSubscriptionExists),
		errors.Is(err, billing.ErrPurchaseExists),
		errors.Is(err, billing.ErrTransactionExists),
		errors.Is(err, billing.ErrIDTaken):
		return codes.AlreadyExists, true

	// Well-formed, and refused by what is already stored rather than by the
	// request: the row already holds that status, or the purchase already
	// completed. Nothing changed and nothing needs to; a caller replaying an event
	// acknowledges it on this.
	case errors.Is(err, billing.ErrStatusUnchanged),
		errors.Is(err, billing.ErrAlreadyCompleted):
		return codes.FailedPrecondition, true

	default:
		return codes.Unknown, false
	}
}
