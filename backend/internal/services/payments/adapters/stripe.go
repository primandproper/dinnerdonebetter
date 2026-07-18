package adapters

import (
	"context"
	"encoding/json"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v5/encoding"
	"github.com/primandproper/platform-go/v5/observability"
	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"

	"github.com/stripe/stripe-go/v75"
	"github.com/stripe/stripe-go/v75/webhook"
)

// StripeConfig holds configuration for the Stripe adapter.
type StripeConfig struct {
	APIKey        string `json:"apiKey"`
	WebhookSecret string `json:"webhookSecret"`
}

// StripePaymentProcessor implements PaymentProcessor for Stripe (web checkout).
type StripePaymentProcessor struct {
	logger  logging.Logger
	tracer  tracing.Tracer
	encoder encoding.ServerEncoderDecoder
	cfg     *StripeConfig
}

// NewStripePaymentProcessor returns a new Stripe payment processor.
func NewStripePaymentProcessor(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	cfg *StripeConfig,
) *StripePaymentProcessor {
	if cfg == nil {
		cfg = &StripeConfig{}
	}

	return &StripePaymentProcessor{
		encoder: encoding.NewServerEncoderDecoder(logger, tracerProvider, encoding.ContentTypeJSON),
		logger:  logging.NewNamedLogger(logger, "stripe_processor"),
		tracer:  tracing.NewNamedTracer(tracerProvider, "stripe_processor"),
		cfg:     cfg,
	}
}

var _ payments.PaymentProcessor = (*StripePaymentProcessor)(nil)

// VerifyWebhookSignature verifies the Stripe webhook signature.
func (s *StripePaymentProcessor) VerifyWebhookSignature(ctx context.Context, payload []byte, signature, accountID string) bool {
	_, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithSpan(span).WithValue(keys.AccountIDKey, accountID)

	if s.cfg.WebhookSecret == "" {
		logger.Debug("webhook secret is empty")
		return true
	}

	_, err := webhook.ConstructEvent(payload, signature, s.cfg.WebhookSecret)
	return err == nil
}

// ParseWebhookEvent parses a Stripe webhook event and returns subscription-related details.
func (s *StripePaymentProcessor) ParseWebhookEvent(ctx context.Context, payload []byte) (*payments.ParsedWebhookEvent, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	logger := s.logger.WithValue("payload_size", len(payload))

	var event stripe.Event
	if err := s.encoder.DecodeBytes(ctx, payload, &event); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "decoding webhook event")
	}

	result := &payments.ParsedWebhookEvent{
		EventType: string(event.Type),
	}

	switch event.Type {
	case stripe.EventTypeCustomerSubscriptionCreated,
		stripe.EventTypeCustomerSubscriptionUpdated,
		stripe.EventTypeCustomerSubscriptionDeleted:
		var subn stripe.Subscription
		if err := json.Unmarshal(event.Data.Raw, &subn); err != nil {
			return nil, err
		}

		result.SubscriptionID = subn.ID
		result.Status = string(subn.Status)
		if subn.Customer != nil {
			result.AccountID = subn.Customer.ID
		}

		if len(subn.Items.Data) > 0 && subn.Items.Data[0].Price != nil && subn.Items.Data[0].Price.ID != "" {
			result.ProductID = subn.Items.Data[0].Price.ID
		}
	}

	return result, nil
}
