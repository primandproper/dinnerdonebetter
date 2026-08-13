package adapters

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v10/capitalism"
	capstripe "github.com/primandproper/platform-go/v10/capitalism/stripe"
	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/observability"
	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/tracing"

	"github.com/stripe/stripe-go/v81"
)

const stripeO11yName = "stripe_processor"

// ErrNoStripeEventCaptured indicates the platform manager accepted a webhook without handing
// us an event. It should be unreachable — capitalism invokes the handler for every event it
// verifies — and exists so that a change there surfaces as an error rather than as a webhook
// we quietly treat as a no-op.
var ErrNoStripeEventCaptured = platformerrors.New("stripe webhook produced no event")

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

// stripeEventSink is where the capitalism event handler leaves the event it was given.
//
// capitalism.PaymentManager reports whether a webhook verified, not what was in it — the
// event goes to a handler registered once at construction, and this processor serves every
// request. A sink carried on the request's own context is therefore how one manager hands
// each concurrent caller its own event.
type stripeEventSink struct {
	event *capstripe.Event
}

type stripeEventSinkContextKey struct{}

// captureStripeEvent is the capitalism.EventHandler this processor registers. Acting on an
// event is the payments manager's job, so all this does is hand it to the caller waiting on it.
func captureStripeEvent(ctx context.Context, event *capstripe.Event) error {
	if sink, ok := ctx.Value(stripeEventSinkContextKey{}).(*stripeEventSink); ok {
		sink.event = event
	}

	return nil
}

// NewStripePaymentProcessor returns a new Stripe payment processor backed by the platform's
// Stripe payment manager.
func NewStripePaymentProcessor(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	cfg *capstripe.Config,
) (*StripePaymentProcessor, error) {
	manager, err := capstripe.NewPaymentManager(
		cfg,
		captureStripeEvent,
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

	sink := &stripeEventSink{}
	req = req.WithContext(context.WithValue(ctx, stripeEventSinkContextKey{}, sink))

	if err := s.manager.HandleEventWebhook(req); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "handling stripe webhook")
	}

	if sink.event == nil {
		return nil, observability.PrepareAndLogError(ErrNoStripeEventCaptured, logger, span, "handling stripe webhook")
	}

	parsed, err := parseStripeEvent(sink.event)
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
func parseStripeEvent(event *capstripe.Event) (*payments.ParsedWebhookEvent, error) {
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
