package adapters

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v8/observability/logging"
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

// VerifyWebhookSignature always returns true for the stub (accepts all webhooks in dev).
func (s *StubPaymentProcessor) VerifyWebhookSignature(_ context.Context, _ []byte, _, _ string) bool {
	s.logger.Info("StubPaymentProcessor VerifyWebhookSignature invoked")
	return true
}

// ParseWebhookEvent returns placeholder values. Real adapters would parse provider-specific payloads.
func (s *StubPaymentProcessor) ParseWebhookEvent(_ context.Context, _ []byte) (*payments.ParsedWebhookEvent, error) {
	s.logger.Info("StubPaymentProcessor ParseWebhookEvent invoked")
	return &payments.ParsedWebhookEvent{
		EventType:      "subscription.updated",
		AccountID:      "",
		SubscriptionID: "",
		ProductID:      "",
	}, nil
}
