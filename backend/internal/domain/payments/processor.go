package payments

import (
	"net/http"

	"github.com/primandproper/platform-go/v13/capitalism"
)

// ParsedWebhookEvent holds the result of parsing a provider webhook payload.
type ParsedWebhookEvent struct {
	EventType      string // e.g. "subscription.updated", "INITIAL_PURCHASE"
	AccountID      string // app_user_id, customer ID, etc.
	SubscriptionID string // external subscription or transaction ID
	ProductID      string // external product ID (e.g. StoreKit product_id for RevenueCat)

	// Status is where the provider says the subscription stands, in
	// capitalism's vocabulary, or SubscriptionStatusUnknown when the event does
	// not say. It is capitalism's type rather than a string because that is what
	// the billing store writes: an adapter that could hand over a word the store
	// refuses would be an adapter whose mistake surfaced as a database error.
	Status capitalism.SubscriptionStatus
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
