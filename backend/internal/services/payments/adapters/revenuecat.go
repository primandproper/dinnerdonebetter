package adapters

import (
	"encoding/json"
	"net/http"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v13/capitalism"
	caprevenuecat "github.com/primandproper/platform-go/v13/capitalism/revenuecat"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const revenueCatO11yName = "revenuecat_processor"

// RevenueCatPaymentProcessor implements PaymentProcessor for RevenueCat (mobile in-app purchases).
//
// Like the Stripe one, it owns no verification and no vendor knowledge: signature checking,
// envelope decoding and the event-type-to-status table belong to platform-go's
// capitalism.PaymentManager, and this type's whole job is turning the platform-owned event it
// hands back into one of our domain's ParsedWebhookEvents.
type RevenueCatPaymentProcessor struct {
	logger  logging.Logger
	tracer  tracing.Tracer
	manager capitalism.PaymentManager
}

var _ payments.PaymentProcessor = (*RevenueCatPaymentProcessor)(nil)

// NewRevenueCatPaymentProcessor returns a new RevenueCat payment processor backed by the
// platform's RevenueCat payment manager.
//
// It returns an error where the hand-rolled adapter it replaces could not: a manager without a
// signing secret is refused at construction rather than at the first delivery, because
// RevenueCat is inbound-only and a secretless manager could do nothing at all.
func NewRevenueCatPaymentProcessor(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	cfg *caprevenuecat.Config,
) (*RevenueCatPaymentProcessor, error) {
	manager, err := caprevenuecat.NewPaymentManager(
		cfg,
		caprevenuecat.WithLogger(logger),
		caprevenuecat.WithTracerProvider(tracerProvider),
	)
	if err != nil {
		return nil, err
	}

	return &RevenueCatPaymentProcessor{
		logger:  logging.NewNamedLogger(logger, revenueCatO11yName),
		tracer:  tracing.NewNamedTracer(tracerProvider, revenueCatO11yName),
		manager: manager,
	}, nil
}

// HandleWebhook verifies the request's RevenueCat signature and returns the parsed event.
func (r *RevenueCatPaymentProcessor) HandleWebhook(req *http.Request) (*payments.ParsedWebhookEvent, error) {
	ctx, span := r.tracer.StartSpan(req.Context())
	defer span.End()

	logger := r.logger.WithSpan(span)

	event, err := r.manager.HandleEventWebhook(req.WithContext(ctx))
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "handling revenuecat webhook")
	}

	parsed, err := parseRevenueCatEvent(event)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "parsing revenuecat event")
	}

	return parsed, nil
}

// parseRevenueCatEvent renders a verified, platform-owned event as one of our domain events.
//
// EventType stays RevenueCat's own word for what happened, because that is what the payments
// manager switches on — INITIAL_PURCHASE, EXPIRATION, and the rest.
func parseRevenueCatEvent(event *capitalism.Event) (*payments.ParsedWebhookEvent, error) {
	result := &payments.ParsedWebhookEvent{
		EventType: event.Type,
	}

	// product_id is the one thing the payments manager needs that capitalism's inbound
	// vocabulary does not name: SubscriptionState carries the identifiers and the status, and
	// everything else stays in the raw payload for whoever wants it. Decoding a single field
	// out of it here is what that arrangement is for — a struct mirroring RevenueCat's event
	// would make every field RevenueCat adds a change to this file.
	if len(event.Payload) > 0 {
		var payload struct {
			ProductID string `json:"product_id"`
		}

		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			return nil, err
		}

		result.ProductID = payload.ProductID
	}

	// Nil for the events RevenueCat documents as carrying no subscription standing — a test
	// event, a transfer — which leave the account and subscription fields empty and fall
	// through the manager's switch.
	if event.Subscription != nil {
		result.AccountID = event.Subscription.CustomerID
		result.SubscriptionID = event.Subscription.ID
		result.Status = revenueCatSubscriptionStatus(event.Subscription.Status)
	}

	return result, nil
}

// revenueCatSubscriptionStatus maps capitalism's subscription vocabulary onto ours.
//
// The two are near-identical because both read like Stripe's, which is the set the industry
// copied; the mapping exists for the three values our domain has no constant for and for
// "cancelled", which we spell with two Ls.
func revenueCatSubscriptionStatus(status capitalism.SubscriptionStatus) string {
	switch status {
	case capitalism.SubscriptionStatusActive:
		return payments.SubscriptionStatusActive
	case capitalism.SubscriptionStatusTrialing:
		return payments.SubscriptionStatusTrialing
	case capitalism.SubscriptionStatusPastDue:
		return payments.SubscriptionStatusPastDue
	case capitalism.SubscriptionStatusCanceled:
		return payments.SubscriptionStatusCancelled
	case capitalism.SubscriptionStatusIncomplete, capitalism.SubscriptionStatusIncompleteExpired:
		return payments.SubscriptionStatusIncomplete
	case capitalism.SubscriptionStatusUnpaid:
		// Not ended, but not collecting either: the processor has stopped retrying and left
		// invoices outstanding, which is the same standing our past-due value describes.
		return payments.SubscriptionStatusPastDue
	case capitalism.SubscriptionStatusPaused:
		// Deliberately suspended at the processor and expected to resume. Our vocabulary has
		// no value for it, and cancelled is the one that gates entitlement correctly: nothing
		// is being collected, so nothing is paid for.
		return payments.SubscriptionStatusCancelled
	default:
		// capitalism.SubscriptionStatusUnknown, and anything a later version adds. Left empty
		// rather than guessed onto a neighboring value: the manager's own default arm treats
		// an unrecognized event as a no-op, and inventing a status here is how a status
		// RevenueCat added last week becomes a wrongly locked-out account.
		return ""
	}
}
