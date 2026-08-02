package adapters

import (
	"io"
	"net/http"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v9/encoding"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	bearerPrefix = "Bearer "

	// revenueCatAuthHeader is the header RevenueCat sends its webhook auth token in.
	revenueCatAuthHeader = "Authorization"
)

// RevenueCatConfig holds configuration for the RevenueCat adapter.
type RevenueCatConfig struct {
	APIKey            string // Secret API key (sk_xxx) for REST API
	WebhookAuthHeader string // Expected value in Authorization header for webhooks (set in RevenueCat dashboard)
}

// RevenueCatPaymentProcessor implements PaymentProcessor for RevenueCat (mobile in-app purchases).
type RevenueCatPaymentProcessor struct {
	logger  logging.Logger
	tracer  tracing.Tracer
	encoder encoding.ServerEncoderDecoder
	cfg     *RevenueCatConfig
}

// NewRevenueCatPaymentProcessor returns a new RevenueCat payment processor.
func NewRevenueCatPaymentProcessor(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	cfg *RevenueCatConfig,
) *RevenueCatPaymentProcessor {
	if cfg == nil {
		cfg = &RevenueCatConfig{}
	}
	return &RevenueCatPaymentProcessor{
		encoder: encoding.NewServerEncoderDecoder(encoding.ContentTypeJSON, encoding.WithLogger(logger), encoding.WithTracerProvider(tracerProvider)),
		logger:  logging.NewNamedLogger(logger, "revenuecat_processor"),
		tracer:  tracing.NewNamedTracer(tracerProvider, "revenuecat_processor"),
		cfg:     cfg,
	}
}

var _ payments.PaymentProcessor = (*RevenueCatPaymentProcessor)(nil)

// RevenueCat event types that imply subscription status.
const (
	rcEventInitialPurchase = "INITIAL_PURCHASE"
	rcEventRenewal         = "RENEWAL"
	rcEventExpiration      = "EXPIRATION"
	rcEventCancellation    = "CANCELLATION"
)

// revenueCatWebhookPayload represents the structure of a RevenueCat webhook event.
type revenueCatWebhookPayload struct {
	ExpirationAtMs   *int64   `json:"expiration_at_ms"`
	PurchasedAtMs    *int64   `json:"purchased_at_ms"`
	ProductID        string   `json:"product_id"`
	Type             string   `json:"type"`
	TransactionID    string   `json:"transaction_id"`
	OriginalTxnID    string   `json:"original_transaction_id"`
	AppUserID        string   `json:"app_user_id"`
	ID               string   `json:"id"`
	Store            string   `json:"store"`
	Environment      string   `json:"environment"`
	CancelReason     string   `json:"cancel_reason"`
	ExpirationReason string   `json:"expiration_reason"`
	EntitlementIDs   []string `json:"entitlement_ids"`
}

// verifyWebhookSignature validates the Authorization header (RevenueCat uses Bearer or custom header).
func (r *RevenueCatPaymentProcessor) verifyWebhookSignature(signature string) bool {
	if r.cfg.WebhookAuthHeader == "" {
		return true // no config = accept (for dev)
	}

	// RevenueCat sends Authorization header; signature param may contain "Bearer <token>" or the raw value
	expected := strings.TrimSpace(r.cfg.WebhookAuthHeader)
	got := strings.TrimSpace(signature)

	if strings.HasPrefix(expected, bearerPrefix) && strings.HasPrefix(got, bearerPrefix) {
		return expected == got
	}
	return expected == got
}

// HandleWebhook validates a RevenueCat webhook's Authorization header and parses its JSON body.
func (r *RevenueCatPaymentProcessor) HandleWebhook(req *http.Request) (*payments.ParsedWebhookEvent, error) {
	ctx, span := r.tracer.StartSpan(req.Context())
	defer span.End()

	logger := r.logger.WithSpan(span)

	if !r.verifyWebhookSignature(req.Header.Get(revenueCatAuthHeader)) {
		return nil, observability.PrepareAndLogError(ErrInvalidWebhookSignature, logger, span, "verifying webhook signature")
	}

	// Bound the body of this public, unauthenticated endpoint so a hostile client can't force an
	// unbounded allocation. RevenueCat's events are well under this.
	payload, err := io.ReadAll(http.MaxBytesReader(nil, req.Body, maxWebhookBodyBytes))
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "reading webhook body")
	}

	logger = logger.WithValue("payload_size", len(payload))

	var p revenueCatWebhookPayload
	if err = r.encoder.DecodeBytes(ctx, payload, &p); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "decoding webhook payload")
	}

	status := ""
	switch p.Type {
	case rcEventInitialPurchase, rcEventRenewal:
		status = payments.SubscriptionStatusActive
	case rcEventExpiration, rcEventCancellation:
		status = payments.SubscriptionStatusCancelled
	}

	return &payments.ParsedWebhookEvent{
		EventType:      p.Type,
		AccountID:      p.AppUserID,
		SubscriptionID: p.TransactionID,
		ProductID:      p.ProductID,
		Status:         status,
	}, nil
}
