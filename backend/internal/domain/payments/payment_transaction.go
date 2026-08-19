package payments

import (
	"time"
)

const (
	PaymentTransactionStatusSucceeded = "succeeded"
	PaymentTransactionStatusFailed    = "failed"
	PaymentTransactionStatusPending   = "pending"
	PaymentTransactionStatusRefunded  = "refunded"
)

type (
	// PaymentTransaction represents an audit record of a payment attempt.
	//
	// It is not a duplicate of capitalism.PaymentIntentCreationInput, despite the overlap in
	// vocabulary. That type is an argument to a provider call — CustomerID, IdempotencyKey, a
	// currency and an amount, all of which are what Stripe needs to be asked for a payment
	// intent — and it is gone once the call returns. This is the stored resource the attempt
	// left behind: it belongs to an account, references a subscription or purchase of ours,
	// carries the provider's ID as a foreign key rather than as its identity, and is read back
	// by the API. Different layers, and nothing to fold together; see Subscription for the same
	// distinction on the recurring side.
	PaymentTransaction struct {
		_                     struct{}  `json:"-"`
		CreatedAt             time.Time `json:"createdAt"`
		SubscriptionID        *string   `json:"subscriptionId"`
		PurchaseID            *string   `json:"purchaseId"`
		ID                    string    `json:"id"`
		BelongsToAccount      string    `json:"belongsToAccount"`
		ExternalTransactionID string    `json:"externalTransactionId"`
		Currency              string    `json:"currency"`
		Status                string    `json:"status"`
		AmountCents           int32     `json:"amountCents"`
	}

	// PaymentTransactionDatabaseCreationInput is used for creating a payment transaction in the database.
	PaymentTransactionDatabaseCreationInput struct {
		_                     struct{} `json:"-"`
		SubscriptionID        *string  `json:"-"`
		PurchaseID            *string  `json:"-"`
		ID                    string   `json:"-"`
		BelongsToAccount      string   `json:"-"`
		ExternalTransactionID string   `json:"-"`
		Currency              string   `json:"-"`
		Status                string   `json:"-"`
		AmountCents           int32    `json:"-"`
	}
)
