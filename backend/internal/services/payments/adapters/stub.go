package adapters

import (
	"net/http"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v13/observability/logging"
)

// StubPaymentProcessor is a no-op implementation of PaymentProcessor for local development and testing.
type StubPaymentProcessor struct {
	logger logging.Logger
}

var _ payments.PaymentProcessor = (*StubPaymentProcessor)(nil)

// NewStubPaymentProcessor returns a new stub payment processor.
func NewStubPaymentProcessor(logger logging.Logger) *StubPaymentProcessor {
	return &StubPaymentProcessor{
		logger: logging.NewNamedLogger(logger, "stub_payment_processor"),
	}
}

// HandleWebhook accepts any request and returns placeholder values. Real adapters verify
// signatures and parse provider-specific payloads.
func (s *StubPaymentProcessor) HandleWebhook(_ *http.Request) (*payments.ParsedWebhookEvent, error) {
	s.logger.Info("StubPaymentProcessor HandleWebhook invoked")
	return &payments.ParsedWebhookEvent{
		EventType:      "subscription.updated",
		AccountID:      "",
		SubscriptionID: "",
		ProductID:      "",
	}, nil
}
