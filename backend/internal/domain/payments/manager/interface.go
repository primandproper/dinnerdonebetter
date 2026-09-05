package manager

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
)

// PaymentsDataManager applies what a payment provider reports to this
// application's own records.
//
// It used to be the seam every payments read and write went through. The stored
// half is platform-go's billing.Store now, and the gRPC service reads and writes
// it directly — see internal/services/payments/grpc. What is left here is the
// half no store can hold: which of a provider's events changes what about a
// subscription, and what that does to the account's billing standing.
type PaymentsDataManager interface {
	// ProcessWebhookEvent applies an already-verified, already-parsed provider event to our own
	// records. Verification and parsing are a payments.PaymentProcessor's job and happen at the
	// transport edge, where the request still exists; by the time an event reaches here it is
	// domain data, and this manager needs to know nothing about HTTP.
	ProcessWebhookEvent(ctx context.Context, provider string, event *payments.ParsedWebhookEvent, accountID string) error
}
