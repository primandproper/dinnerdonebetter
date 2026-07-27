package adapters

import (
	"context"
	"strings"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v7/encoding"
	"github.com/primandproper/platform-go/v7/observability"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/tracing"
)

const bearerPrefix = "Bearer "

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
		encoder: encoding.NewServerEncoderDecoder(logger, tracerProvider, encoding.ContentTypeJSON),
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

// VerifyWebhookSignature validates the Authorization header (RevenueCat uses Bearer or custom header).
func (r *RevenueCatPaymentProcessor) VerifyWebhookSignature(ctx context.Context, _ []byte, signature, _ string) bool {
	_, span := r.tracer.StartSpan(ctx)
	defer span.End()

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

// ParseWebhookEvent parses RevenueCat webhook JSON and returns event details.
func (r *RevenueCatPaymentProcessor) ParseWebhookEvent(ctx context.Context, payload []byte) (*payments.ParsedWebhookEvent, error) {
	_, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithValue("payload_size", len(payload))

	var p revenueCatWebhookPayload
	if err := r.encoder.DecodeBytes(ctx, payload, &p); err != nil {
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
