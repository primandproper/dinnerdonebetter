package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v7/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ValidPreparationInstrumentCreatedServiceEventType indicates a valid preparation instrument was created.
	ValidPreparationInstrumentCreatedServiceEventType = "valid_preparation_instrument_created"
	// ValidPreparationInstrumentUpdatedServiceEventType indicates a valid preparation instrument was updated.
	ValidPreparationInstrumentUpdatedServiceEventType = "valid_preparation_instrument_updated"
	// ValidPreparationInstrumentArchivedServiceEventType indicates a valid preparation instrument was archived.
	ValidPreparationInstrumentArchivedServiceEventType = "valid_preparation_instrument_archived"
)

func init() {
	gob.Register(new(ValidPreparationInstrument))
	gob.Register(new(ValidPreparationInstrumentCreationRequestInput))
	gob.Register(new(ValidPreparationInstrumentUpdateRequestInput))
}

type (
	// ValidPreparationInstrument represents a valid preparation instrument.
	ValidPreparationInstrument struct {
		_ struct{} `json:"-"`

		CreatedAt     time.Time        `json:"createdAt"`
		LastUpdatedAt *time.Time       `json:"lastUpdatedAt"`
		ArchivedAt    *time.Time       `json:"archivedAt"`
		ID            string           `json:"id"`
		Notes         string           `json:"notes"`
		Instrument    ValidInstrument  `json:"instrument"`
		Preparation   ValidPreparation `json:"preparation"`
	}

	// ValidPreparationInstrumentCreationRequestInput represents what a user could set as input for creating valid preparation instruments.
	ValidPreparationInstrumentCreationRequestInput struct {
		_ struct{} `json:"-"`

		Notes              string `json:"notes"`
		ValidPreparationID string `json:"validPreparationID"`
		ValidInstrumentID  string `json:"validInstrumentID"`
	}

	// ValidPreparationInstrumentDatabaseCreationInput represents what a user could set as input for creating valid preparation instruments.
	ValidPreparationInstrumentDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                 string `json:"-"`
		Notes              string `json:"-"`
		ValidPreparationID string `json:"-"`
		ValidInstrumentID  string `json:"-"`
	}

	// ValidPreparationInstrumentUpdateRequestInput represents what a user could set as input for updating valid preparation instruments.
	ValidPreparationInstrumentUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Notes              *string `json:"notes,omitempty"`
		ValidPreparationID *string `json:"validPreparationID,omitempty"`
		ValidInstrumentID  *string `json:"validInstrumentID,omitempty"`
	}

	// ValidPreparationInstrumentDataManager describes a structure capable of storing valid preparation instruments permanently.
	ValidPreparationInstrumentDataManager interface {
		ValidPreparationInstrumentExists(ctx context.Context, validPreparationInstrumentID string) (bool, error)
		GetValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string) (*ValidPreparationInstrument, error)
		GetValidPreparationInstruments(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidPreparationInstrument], error)
		GetValidPreparationInstrumentsForPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidPreparationInstrument], error)
		GetValidPreparationInstrumentsForInstrument(ctx context.Context, instrumentID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidPreparationInstrument], error)
		GetValidPreparationInstrumentsByIDs(ctx context.Context, ids []string) (map[string]*ValidPreparationInstrument, error)
		CreateValidPreparationInstrument(ctx context.Context, input *ValidPreparationInstrumentDatabaseCreationInput) (*ValidPreparationInstrument, error)
		UpdateValidPreparationInstrument(ctx context.Context, updated *ValidPreparationInstrument) error
		ArchiveValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string) error
	}
)

// Update merges a ValidPreparationInstrumentUpdateRequestInput with a valid preparation instrument.
func (x *ValidPreparationInstrument) Update(input *ValidPreparationInstrumentUpdateRequestInput) {
	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}

	if input.ValidPreparationID != nil && *input.ValidPreparationID != x.Preparation.ID {
		x.Preparation.ID = *input.ValidPreparationID
	}

	if input.ValidInstrumentID != nil && *input.ValidInstrumentID != x.Instrument.ID {
		x.Instrument.ID = *input.ValidInstrumentID
	}
}

var _ validation.ValidatableWithContext = (*ValidPreparationInstrumentCreationRequestInput)(nil)

// ValidateWithContext validates a ValidPreparationInstrumentCreationRequestInput.
func (x *ValidPreparationInstrumentCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ValidPreparationID, validation.Required),
		validation.Field(&x.ValidInstrumentID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidPreparationInstrumentDatabaseCreationInput)(nil)

// ValidateWithContext validates a ValidPreparationInstrumentDatabaseCreationInput.
func (x *ValidPreparationInstrumentDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.ValidPreparationID, validation.Required),
		validation.Field(&x.ValidInstrumentID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidPreparationInstrumentUpdateRequestInput)(nil)

// ValidateWithContext validates a ValidPreparationInstrumentUpdateRequestInput.
func (x *ValidPreparationInstrumentUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ValidPreparationID, validation.Required),
		validation.Field(&x.ValidInstrumentID, validation.Required),
	)
}
