package http

import (
	"io"
	"net/http"

	paymentsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments/manager"

	"github.com/primandproper/platform-go/v7/observability"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/tracing"

	"github.com/go-chi/chi/v5"
)

const (
	// StripeSignatureHeader is the header Stripe uses for webhook signatures.
	StripeSignatureHeader = "Stripe-Signature"
	// RevenueCatAuthHeader is the header RevenueCat uses for webhook auth (Authorization).
	RevenueCatAuthHeader = "Authorization"
	// DefaultSignatureHeader is a fallback header name for provider-agnostic use.
	DefaultSignatureHeader = "X-Webhook-Signature"
)

// WebhookHandler handles HTTP POST requests from payment providers for webhook events.
type WebhookHandler struct {
	tracer          tracing.Tracer
	logger          logging.Logger
	paymentsManager paymentsmanager.PaymentsDataManager
	signatureHeader string
}

// WebhookSignatureHeader is the name of the HTTP header containing the webhook signature.
// Inject via wire.Value(paymentshttp.WebhookSignatureHeader("Stripe-Signature")) or similar.
type WebhookSignatureHeader string

// NewWebhookHandler returns a new WebhookHandler.
func NewWebhookHandler(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	paymentsManager paymentsmanager.PaymentsDataManager,
	signatureHeader WebhookSignatureHeader,
) *WebhookHandler {
	sig := string(signatureHeader)
	if sig == "" {
		sig = DefaultSignatureHeader
	}
	return &WebhookHandler{
		tracer:          tracing.NewNamedTracer(tracerProvider, "payments_webhook"),
		logger:          logging.NewNamedLogger(logger, "payments_webhook"),
		paymentsManager: paymentsManager,
		signatureHeader: sig,
	}
}

// Handle processes an incoming webhook POST request.
func (h *WebhookHandler) Handle(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.StartSpan(r.Context())
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("reading webhook body", err)
		http.Error(w, "failed to read body", http.StatusBadRequest)
		return
	}
	if err = r.Body.Close(); err != nil {
		h.logger.Error("closing webhook body", err)
	}

	provider := chi.URLParam(r, "provider")
	signature := r.Header.Get(h.signatureHeader)
	if provider == "revenuecat" {
		signature = r.Header.Get(RevenueCatAuthHeader)
	}

	// accountID from URL or query - for stub we pass empty; manager will parse from payload
	accountID := ""
	if id := r.URL.Query().Get("account_id"); id != "" {
		accountID = id
	}

	if err = h.paymentsManager.ProcessWebhookEvent(ctx, provider, payload, signature, accountID); err != nil {
		observability.AcknowledgeError(err, h.logger, span, "processing webhook event")
		http.Error(w, "webhook processing failed", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
