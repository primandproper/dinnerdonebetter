package payments

import "net/http"

// ParsedWebhookEvent holds the result of parsing a provider webhook payload.
type ParsedWebhookEvent struct {
	EventType      string // e.g. "subscription.updated", "INITIAL_PURCHASE"
	AccountID      string // app_user_id, customer ID, etc.
	SubscriptionID string // external subscription or transaction ID
	ProductID      string // external product ID (e.g. StoreKit product_id for RevenueCat)
	Status         string // subscription status when known from payload (e.g. "active", "cancelled")
}

// PaymentProcessor defines the interface for payment provider webhook handling.
// Implementations (e.g., Stripe, RevenueCat) verify and parse webhooks; the manager writes to the database.
//
// It takes the whole request rather than a payload plus a signature because verification is
// the provider's business, not ours: a provider decides which headers it signs and how much
// of a body it will read, and pulling those apart here would mean deciding them on its
// behalf. platform-go's capitalism.PaymentManager draws the same line, and the Stripe
// implementation behind this interface is now that one.
type PaymentProcessor interface {
	HandleWebhook(req *http.Request) (*ParsedWebhookEvent, error)
}
