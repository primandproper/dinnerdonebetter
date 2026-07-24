package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v6/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// AccountInstrumentOwnershipCreatedServiceEventType indicates an account instrument ownership was created.
	AccountInstrumentOwnershipCreatedServiceEventType = "account_instrument_ownership_created"
	// AccountInstrumentOwnershipUpdatedServiceEventType indicates an account instrument ownership was updated.
	AccountInstrumentOwnershipUpdatedServiceEventType = "account_instrument_ownership_updated"
	// AccountInstrumentOwnershipArchivedServiceEventType indicates an account instrument ownership was archived.
	AccountInstrumentOwnershipArchivedServiceEventType = "account_instrument_ownership_archived"
)

func init() {
	gob.Register(new(AccountInstrumentOwnership))
	gob.Register(new(AccountInstrumentOwnershipCreationRequestInput))
	gob.Register(new(AccountInstrumentOwnershipUpdateRequestInput))
}

type (
	// AccountInstrumentOwnership represents an account instrument ownership.
	AccountInstrumentOwnership struct {
		_ struct{} `json:"-"`

		CreatedAt        time.Time       `json:"createdAt"`
		ArchivedAt       *time.Time      `json:"archivedAt"`
		LastUpdatedAt    *time.Time      `json:"lastUpdatedAt"`
		ID               string          `json:"id"`
		Notes            string          `json:"notes"`
		BelongsToAccount string          `json:"belongsToAccount"`
		Instrument       ValidInstrument `json:"instrument"`
		Quantity         uint16          `json:"quantity"`
	}

	// AccountInstrumentOwnershipCreationRequestInput represents what a user could set as input for creating account instrument ownerships.
	AccountInstrumentOwnershipCreationRequestInput struct {
		_ struct{} `json:"-"`

		Notes             string `json:"notes"`
		ValidInstrumentID string `json:"validInstrumentID"`
		Quantity          uint16 `json:"quantity"`
	}

	// AccountInstrumentOwnershipDatabaseCreationInput represents what a user could set as input for creating account instrument ownerships.
	AccountInstrumentOwnershipDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                string `json:"-"`
		Notes             string `json:"-"`
		ValidInstrumentID string `json:"-"`
		BelongsToAccount  string `json:"-"`
		Quantity          uint16 `json:"-"`
	}

	// AccountInstrumentOwnershipUpdateRequestInput represents what a user could set as input for updating account instrument ownerships.
	AccountInstrumentOwnershipUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Notes             *string `json:"notes"`
		Quantity          *uint16 `json:"quantity"`
		ValidInstrumentID *string `json:"validInstrumentID"`
	}

	// AccountInstrumentOwnershipDataManager describes a structure capable of storing account instrument ownerships permanently.
	AccountInstrumentOwnershipDataManager interface {
		AccountInstrumentOwnershipExists(ctx context.Context, accountInstrumentOwnershipID, accountID string) (bool, error)
		GetAccountInstrumentOwnership(ctx context.Context, accountInstrumentOwnershipID, accountID string) (*AccountInstrumentOwnership, error)
		GetAccountInstrumentOwnerships(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AccountInstrumentOwnership], error)
		CreateAccountInstrumentOwnership(ctx context.Context, input *AccountInstrumentOwnershipDatabaseCreationInput) (*AccountInstrumentOwnership, error)
		UpdateAccountInstrumentOwnership(ctx context.Context, updated *AccountInstrumentOwnership) error
		ArchiveAccountInstrumentOwnership(ctx context.Context, accountInstrumentOwnershipID, accountID string) error
	}
)

// Update merges an AccountInstrumentOwnershipUpdateRequestInput with an account instrument ownership.
func (x *AccountInstrumentOwnership) Update(input *AccountInstrumentOwnershipUpdateRequestInput) {
	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}

	if input.Quantity != nil && *input.Quantity != x.Quantity {
		x.Quantity = *input.Quantity
	}

	if input.ValidInstrumentID != nil && *input.ValidInstrumentID != x.Instrument.ID {
		x.Instrument = ValidInstrument{ID: *input.ValidInstrumentID}
	}
}

var _ validation.ValidatableWithContext = (*AccountInstrumentOwnershipCreationRequestInput)(nil)

// ValidateWithContext validates a AccountInstrumentOwnershipCreationRequestInput.
func (x *AccountInstrumentOwnershipCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Quantity, validation.Required),
		validation.Field(&x.ValidInstrumentID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*AccountInstrumentOwnershipDatabaseCreationInput)(nil)

// ValidateWithContext validates a AccountInstrumentOwnershipDatabaseCreationInput.
func (x *AccountInstrumentOwnershipDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Quantity, validation.Required),
		validation.Field(&x.ValidInstrumentID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*AccountInstrumentOwnershipUpdateRequestInput)(nil)

// ValidateWithContext validates a AccountInstrumentOwnershipUpdateRequestInput.
func (x *AccountInstrumentOwnershipUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Quantity, validation.Required),
		validation.Field(&x.ValidInstrumentID, validation.Required),
	)
}
