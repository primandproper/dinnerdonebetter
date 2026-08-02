package http

import (
	"net/http"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/manager"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/go-chi/chi/v5"
)

// ErrUnknownPaymentProvider indicates a webhook arrived for a provider that has no registered processor.
var ErrUnknownPaymentProvider = platformerrors.New("unknown payment provider")

// WebhookHandler handles HTTP POST requests from payment providers for webhook events.
type WebhookHandler struct {
	tracer            tracing.Tracer
	logger            logging.Logger
	paymentsManager   paymentsmanager.PaymentsDataManager
	processorRegistry payments.PaymentProcessorRegistry
}

// NewWebhookHandler returns a new WebhookHandler.
func NewWebhookHandler(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
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

// Handle processes an incoming webhook POST request.
//
// It hands the whole request to the provider's processor rather than reading the body here:
// each provider signs a different part of a request, and only its own processor knows which.
// What comes back is domain data, which is all the payments manager is given.
func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Carry this span on the request the processor is handed, so its own spans nest under it.
	r = r.WithContext(ctx)

	provider := chi.URLParam(r, "provider")
	logger := h.logger.WithSpan(span).WithValue("provider", provider)

	processor, ok := h.processorRegistry.GetProcessor(provider)
	if !ok {
		observability.AcknowledgeError(ErrUnknownPaymentProvider, logger, span, "resolving payment processor")
		http.Error(w, "unknown payment provider", http.StatusBadRequest)
		return
	}

	event, err := processor.HandleWebhook(r)
	if err != nil {
		observability.AcknowledgeError(err, logger, span, "handling webhook")
		http.Error(w, "webhook processing failed", http.StatusBadRequest)
		return
	}

	// accountID from the URL when the provider's payload doesn't carry one; the manager falls
	// back to the event's own (e.g. RevenueCat's app_user_id) when this is empty.
	accountID := r.URL.Query().Get("account_id")

	if err = h.paymentsManager.ProcessWebhookEvent(ctx, provider, event, accountID); err != nil {
		observability.AcknowledgeError(err, logger, span, "processing webhook event")
		http.Error(w, "webhook processing failed", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
