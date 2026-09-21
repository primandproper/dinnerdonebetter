package keys

const (
	idSuffix = ".id"

	// ProductIDKey is the standard key for referring to a product's ID.
	ProductIDKey = "product" + idSuffix
	// SubscriptionIDKey is the standard key for referring to a subscription's ID.
	SubscriptionIDKey = "subscription" + idSuffix
	// SubscriptionStatusKey names the standing a processor reported for a subscription.
	// It is logged when that standing is a word this deployment cannot place, which is
	// the one case where the value itself is the thing somebody needs to read.
	SubscriptionStatusKey = "subscription.status"
	// PurchaseIDKey is the standard key for referring to a purchase's ID.
	PurchaseIDKey = "purchase" + idSuffix
	// PaymentTransactionIDKey is the standard key for referring to a payment transaction's ID.
	PaymentTransactionIDKey = "payment_transaction" + idSuffix
)
