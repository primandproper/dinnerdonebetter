package payments

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ProductKindRecurring indicates a subscription/recurring product.
	ProductKindRecurring = "recurring"
	// ProductKindOneTime indicates a one-time purchase product.
	ProductKindOneTime = "one_time"

	// ProductCreatedServiceEventType indicates a product was created.
	ProductCreatedServiceEventType = "product_created"
	// ProductUpdatedServiceEventType indicates a product was updated.
	ProductUpdatedServiceEventType = "product_updated"
	// ProductArchivedServiceEventType indicates a product was archived.
	ProductArchivedServiceEventType = "product_archived"
)

type (
	// Product represents a purchasable item (subscription plan or one-time offering).
	Product struct {
		_                     struct{}   `json:"-"`
		CreatedAt             time.Time  `json:"createdAt"`
		BillingIntervalMonths *int32     `json:"billingIntervalMonths"`
		LastUpdatedAt         *time.Time `json:"lastUpdatedAt"`
		ArchivedAt            *time.Time `json:"archivedAt"`
		ID                    string     `json:"id"`
		Name                  string     `json:"name"`
		Description           string     `json:"description"`
		Kind                  string     `json:"kind"`
		Currency              string     `json:"currency"`
		ExternalProductID     string     `json:"externalProductId"`
		AmountCents           int32      `json:"amountCents"`
	}

	// ProductCreationRequestInput represents input for creating a product.
	ProductCreationRequestInput struct {
		_                     struct{} `json:"-"`
		BillingIntervalMonths *int32   `json:"billingIntervalMonths"`
		Name                  string   `json:"name"`
		Description           string   `json:"description"`
		Kind                  string   `json:"kind"`
		Currency              string   `json:"currency"`
		ExternalProductID     string   `json:"externalProductId"`
		AmountCents           int32    `json:"amountCents"`
	}

	// ProductUpdateRequestInput represents input for updating a product.
	ProductUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Name                  *string `json:"name,omitempty"`
		Description           *string `json:"description,omitempty"`
		Kind                  *string `json:"kind,omitempty"`
		AmountCents           *int32  `json:"amountCents,omitempty"`
		Currency              *string `json:"currency,omitempty"`
		BillingIntervalMonths *int32  `json:"billingIntervalMonths,omitempty"`
		ExternalProductID     *string `json:"externalProductId,omitempty"`
	}

	// ProductDatabaseCreationInput is used for creating a product in the database.
	ProductDatabaseCreationInput struct {
		_                     struct{} `json:"-"`
		BillingIntervalMonths *int32   `json:"-"`
		ID                    string   `json:"-"`
		Name                  string   `json:"-"`
		Description           string   `json:"-"`
		Kind                  string   `json:"-"`
		Currency              string   `json:"-"`
		ExternalProductID     string   `json:"-"`
		AmountCents           int32    `json:"-"`
	}
)

var _ validation.ValidatableWithContext = (*ProductCreationRequestInput)(nil)

// ValidateWithContext validates a ProductCreationRequestInput.
func (x *ProductCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Description, validation.Required),
		validation.Field(&x.Kind, validation.Required, validation.In(ProductKindRecurring, ProductKindOneTime)),
		validation.Field(&x.AmountCents, validation.Min(0)),
		validation.Field(&x.Currency, validation.Required),
		validation.Field(&x.BillingIntervalMonths, validation.When(x.Kind == ProductKindRecurring, validation.Required, validation.Min(1))),
	)
}

var _ validation.ValidatableWithContext = (*ProductUpdateRequestInput)(nil)

// ValidateWithContext validates a ProductUpdateRequestInput.
func (x *ProductUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, x,
		validation.Field(&x.Kind, validation.In(ProductKindRecurring, ProductKindOneTime)),
		validation.Field(&x.AmountCents, validation.Min(0)),
		validation.Field(&x.BillingIntervalMonths, validation.Min(1)),
	)
}
