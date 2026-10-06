package http

import (
	"errors"
	"net/http"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/manager"

	"github.com/primandproper/platform-go/v15/billing"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// ErrUnknownPaymentProvider indicates a webhook arrived for a provider that has no registered processor.
var ErrUnknownPaymentProvider = platformerrors.New("unknown payment provider")

// WebhookHandler handles HTTP POST requests from payment providers for webhook events.
//
// This whole package is deletable once platform-go's billing/http can replace it. That handler
// already does what the fixes below do by hand — the account only from the signed payload, 500
// for anything a redelivery could fix, one endpoint per provider — and reconciles through
// billing/sync. Two gaps keep it out of reach, and both are upstream:
//
//   - primitives-go#63: capitalism.SubscriptionState carries no price or
//     product identifier, so billing/sync's Place has nothing to answer Placement.ProductID with,
//     and every delivery that opens a subscription would be refused with ErrNoPlacement.
//   - platform-go#1146: identity has no read of an account by its payment
//     processor customer ID, so a Stripe Place cannot find the account a delivery is for.
//
// Until then this is the local handler, holding the line on the two defects billing/http's own
// documentation names as the ones this handler had.
type WebhookHandler struct {
	tracer            tracing.Tracer
	logger            logging.Logger
	paymentsManager   paymentsmanager.PaymentsDataManager
	processorRegistry payments.PaymentProcessorRegistry
}

// NewWebhookHandler returns a new WebhookHandler.
func NewWebhookHandler(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	paymentsManager paymentsmanager.PaymentsDataManager,
	processorRegistry payments.PaymentProcessorRegistry,
) *WebhookHandler {
	return &WebhookHandler{
		tracer:            tracing.NewNamedTracer(tracerProvider, "payments_webhook"),
		logger:            logging.NewNamedLogger(logger, "payments_webhook"),
		paymentsManager:   paymentsManager,
		processorRegistry: processorRegistry,
	}
}

// For returns the endpoint for one provider.
//
// One route per provider rather than a /{provider} pattern, so the set of endpoints is the set
// of providers this service takes deliveries from, written down at the router: a path naming
// any other provider is a 404 from the router rather than a request this handler has to turn
// away, and the provider a handler serves is fixed when it is mounted rather than read off
// whatever path a caller chose.
func (h *WebhookHandler) For(provider string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.handle(w, r, provider)
	})
}

// handle processes an incoming webhook POST request for the given provider.
//
// It hands the whole request to the provider's processor rather than reading the body here:
// each provider signs a different part of a request, and only its own processor knows which.
// What comes back is domain data, which is all the payments manager is given.
//
// That is also why this stays a raw handler rather than becoming a typed route. routing.RawBody
// would preserve the bytes — it reads the body whole and parses nothing — but bytes are not what
// this handler passes on. PaymentProcessor.HandleWebhook takes the request, and Stripe's verifier
// reads the signature header and the body off it together; a typed route would have to hand that
// verifier a request rebuilt from the parts the router bound, which is not the one the client sent.
// The bound a typed route would have brought with it is applied at registration instead.
//
// The status code is the only thing a provider reads, so it is the whole of the answer. 400 is
// for a delivery sending the same bytes again cannot fix: one that failed verification or could
// not be parsed, and one naming a subscription or product this service has never heard of.
// Everything else is 500, so the provider redelivers. This used to answer 400 to all of it,
// which told the provider that a database blip was a bad delivery and not to send it again.
//
// Nothing about the request but its signed body decides which account a delivery is for. This
// used to read an account_id query parameter that overrode the one in the signed payload; it is
// gone, and a query string is now ignored entirely.
func (h *WebhookHandler) handle(w http.ResponseWriter, r *http.Request, provider string) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Carry this span on the request the processor is handed, so its own spans nest under it.
	r = r.WithContext(ctx)

	logger := h.logger.WithSpan(span).WithValue("provider", provider)

	// The route exists, so a provider without a processor is this service's wiring being wrong
	// rather than the delivery, and the provider should keep trying until somebody fixes it.
	processor, ok := h.processorRegistry.GetProcessor(provider)
	if !ok {
		observability.AcknowledgeError(ErrUnknownPaymentProvider, logger, span, "resolving payment processor")
		http.Error(w, "webhook processing failed", http.StatusInternalServerError)
		return
	}

	event, err := processor.HandleWebhook(r)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "handling webhook")
		http.Error(w, "webhook verification failed", http.StatusBadRequest)
		return
	}

	if err = h.paymentsManager.ProcessWebhookEvent(ctx, provider, event); err != nil {
		observability.AcknowledgeError(err, logger, span, "processing webhook event")
		http.Error(w, "webhook processing failed", processingFailureStatus(err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// processingFailureStatus is the status a verified delivery the manager could not apply is
// answered with.
//
// A delivery naming something this service has never heard of is the one failure a redelivery
// cannot fix, so it alone is a 400. Every other failure — a store or transaction error, a
// timeout — is a 500, which is what makes the provider try again.
func processingFailureStatus(err error) int {
	if errors.Is(err, billing.ErrSubscriptionNotFound) || errors.Is(err, billing.ErrProductNotFound) {
		return http.StatusBadRequest
	}

	return http.StatusInternalServerError
}
