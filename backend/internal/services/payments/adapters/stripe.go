package adapters

import (
	"encoding/json"
	"net/http"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v13/capitalism"
	capstripe "github.com/primandproper/platform-go/v13/capitalism/stripe"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"github.com/stripe/stripe-go/v81"
)

const stripeO11yName = "stripe_processor"

// StripePaymentProcessor implements PaymentProcessor for Stripe (web checkout).
//
// Signature verification and event decoding belong to platform-go's
// capitalism.PaymentManager; this type's whole job is turning the platform-owned event it
// hands back into one of our domain's ParsedWebhookEvents.
type StripePaymentProcessor struct {
	logger  logging.Logger
	tracer  tracing.Tracer
	manager capitalism.PaymentManager
}

var _ payments.PaymentProcessor = (*StripePaymentProcessor)(nil)

// NewStripePaymentProcessor returns a new Stripe payment processor backed by the platform's
// Stripe payment manager.
func NewStripePaymentProcessor(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	cfg *capstripe.Config,
) (*StripePaymentProcessor, error) {
	manager, err := capstripe.NewPaymentManager(
		cfg,
		capstripe.WithLogger(logger),
		capstripe.WithTracerProvider(tracerProvider),
	)
	if err != nil {
		return nil, err
	}

	return &StripePaymentProcessor{
		logger:  logging.NewNamedLogger(logger, stripeO11yName),
		tracer:  tracing.NewNamedTracer(tracerProvider, stripeO11yName),
		manager: manager,
	}, nil
}

// HandleWebhook verifies the request's Stripe signature and returns the parsed event.
func (s *StripePaymentProcessor) HandleWebhook(req *http.Request) (*payments.ParsedWebhookEvent, error) {
	ctx, span := s.tracer.StartSpan(req.Context())
	defer span.End()

	logger := s.logger.WithSpan(span)

	event, err := s.manager.HandleEventWebhook(req.WithContext(ctx))
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "handling stripe webhook")
	}

	parsed, err := parseStripeEvent(event)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "parsing stripe event")
	}

	return parsed, nil
}

// parseStripeEvent renders a verified, platform-owned event as one of our domain events.
//
// capitalism hands over the event's raw JSON rather than a stripe-go struct, so the SDK
// version used to decode it is ours to pick and ours to bump: a Stripe major becomes a change
// to this function rather than a breaking change for every consumer of platform-go.
func parseStripeEvent(event *capitalism.Event) (*payments.ParsedWebhookEvent, error) {
	result := &payments.ParsedWebhookEvent{
		EventType: event.Type,
	}

	switch stripe.EventType(event.Type) {
	case stripe.EventTypeCustomerSubscriptionCreated,
		stripe.EventTypeCustomerSubscriptionUpdated,
		stripe.EventTypeCustomerSubscriptionDeleted:
		var subn stripe.Subscription
		if err := json.Unmarshal(event.Payload, &subn); err != nil {
			return nil, err
		}

		result.SubscriptionID = subn.ID
		result.Status = string(subn.Status)
		if subn.Customer != nil {
			result.AccountID = subn.Customer.ID
		}

		if subn.Items != nil && len(subn.Items.Data) > 0 && subn.Items.Data[0].Price != nil && subn.Items.Data[0].Price.ID != "" {
			result.ProductID = subn.Items.Data[0].Price.ID
		}
	}

	return result, nil
}
