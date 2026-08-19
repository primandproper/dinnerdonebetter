package payments

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	SubscriptionStatusActive     = "active"
	SubscriptionStatusCancelled  = "cancelled"
	SubscriptionStatusPastDue    = "past_due"
	SubscriptionStatusTrialing   = "trialing"
	SubscriptionStatusIncomplete = "incomplete"

	// SubscriptionCreatedServiceEventType indicates a subscription was created.
	SubscriptionCreatedServiceEventType = "subscription_created"
	// SubscriptionUpdatedServiceEventType indicates a subscription was updated.
	SubscriptionUpdatedServiceEventType = "subscription_updated"
	// SubscriptionArchivedServiceEventType indicates a subscription was archived.
	SubscriptionArchivedServiceEventType = "subscription_archived"
)

type (
	// Subscription represents a recurring agreement for an account.
	//
	// It is not a duplicate of capitalism.SubscriptionCreationInput. That type is a provider
	// call's arguments — CustomerID, PriceID, IdempotencyKey — naming things that exist in
	// Stripe and nowhere here. This is the row: it belongs to an account, points at one of our
	// products, has a status this application transitions and archival timestamps this
	// application stamps, and holds the provider's own identifier in ExternalSubscriptionID
	// precisely because the two identities are separate. The platform deliberately has no
	// counterpart to it — how a subscription is stored, scoped, and expired is an
	// application's decision.
	Subscription struct {
		_                      struct{}   `json:"-"`
		CurrentPeriodStart     time.Time  `json:"currentPeriodStart"`
		CurrentPeriodEnd       time.Time  `json:"currentPeriodEnd"`
		CreatedAt              time.Time  `json:"createdAt"`
		LastUpdatedAt          *time.Time `json:"lastUpdatedAt"`
		ArchivedAt             *time.Time `json:"archivedAt"`
		ID                     string     `json:"id"`
		BelongsToAccount       string     `json:"belongsToAccount"`
		ProductID              string     `json:"productId"`
		ExternalSubscriptionID string     `json:"externalSubscriptionId"`
		Status                 string     `json:"status"`
	}

	// SubscriptionCreationRequestInput represents input for creating a subscription.
	SubscriptionCreationRequestInput struct {
		_                      struct{}  `json:"-"`
		CurrentPeriodStart     time.Time `json:"currentPeriodStart"`
		CurrentPeriodEnd       time.Time `json:"currentPeriodEnd"`
		BelongsToAccount       string    `json:"belongsToAccount"`
		ProductID              string    `json:"productId"`
		ExternalSubscriptionID string    `json:"externalSubscriptionId"`
		Status                 string    `json:"status"`
	}

	// SubscriptionUpdateRequestInput represents input for updating a subscription.
	SubscriptionUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Status             *string    `json:"status,omitempty"`
		CurrentPeriodStart *time.Time `json:"currentPeriodStart,omitempty"`
		CurrentPeriodEnd   *time.Time `json:"currentPeriodEnd,omitempty"`
	}

	// SubscriptionDatabaseCreationInput is used for creating a subscription in the database.
	SubscriptionDatabaseCreationInput struct {
		_                      struct{}  `json:"-"`
		CurrentPeriodStart     time.Time `json:"-"`
		CurrentPeriodEnd       time.Time `json:"-"`
		ID                     string    `json:"-"`
		BelongsToAccount       string    `json:"-"`
		ProductID              string    `json:"-"`
		ExternalSubscriptionID string    `json:"-"`
		Status                 string    `json:"-"`
	}
)

var _ validation.ValidatableWithContext = (*SubscriptionCreationRequestInput)(nil)

// ValidateWithContext validates a SubscriptionCreationRequestInput.
func (x *SubscriptionCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.BelongsToAccount, validation.Required),
		validation.Field(&x.ProductID, validation.Required),
		validation.Field(&x.Status, validation.Required, validation.In(
			SubscriptionStatusActive, SubscriptionStatusCancelled,
			SubscriptionStatusPastDue, SubscriptionStatusTrialing, SubscriptionStatusIncomplete,
		)),
	)
}

var _ validation.ValidatableWithContext = (*SubscriptionUpdateRequestInput)(nil)

// ValidateWithContext validates a SubscriptionUpdateRequestInput.
func (x *SubscriptionUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Status, validation.In(
			SubscriptionStatusActive, SubscriptionStatusCancelled,
			SubscriptionStatusPastDue, SubscriptionStatusTrialing, SubscriptionStatusIncomplete,
		)),
	)
}
