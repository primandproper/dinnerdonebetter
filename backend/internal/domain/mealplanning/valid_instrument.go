package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v13/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ValidInstrumentCreatedServiceEventType indicates a valid instrument was created.
	ValidInstrumentCreatedServiceEventType = "valid_instrument_created"
	// ValidInstrumentUpdatedServiceEventType indicates a valid instrument was updated.
	ValidInstrumentUpdatedServiceEventType = "valid_instrument_updated"
	// ValidInstrumentArchivedServiceEventType indicates a valid instrument was archived.
	ValidInstrumentArchivedServiceEventType = "valid_instrument_archived"
)

func init() {
	gob.Register(new(ValidInstrument))
	gob.Register(new(ValidInstrumentCreationRequestInput))
	gob.Register(new(ValidInstrumentUpdateRequestInput))
}

type (
	// ValidInstrument represents a valid instrument.
	ValidInstrument struct {
		_ struct{} `json:"-"`

		CreatedAt                      time.Time  `json:"createdAt"`
		LastUpdatedAt                  *time.Time `json:"lastUpdatedAt"`
		ArchivedAt                     *time.Time `json:"archivedAt"`
		IconPath                       string     `json:"iconPath"`
		ID                             string     `json:"id"`
		Name                           string     `json:"name"`
		PluralName                     string     `json:"pluralName"`
		Description                    string     `json:"description"`
		Slug                           string     `json:"slug"`
		DisplayInSummaryLists          bool       `json:"displayInSummaryLists"`
		IncludeInGeneratedInstructions bool       `json:"includeInGeneratedInstructions"`
		UsableForStorage               bool       `json:"usableForStorage"`
	}

	// NullableValidInstrument represents a fully nullable valid instrument.
	NullableValidInstrument struct {
		_ struct{} `json:"-"`

		LastUpdatedAt                  *time.Time
		ArchivedAt                     *time.Time
		Description                    *string
		IconPath                       *string
		ID                             *string
		Name                           *string
		Slug                           *string
		DisplayInSummaryLists          *bool
		IncludeInGeneratedInstructions *bool
		PluralName                     *string
		UsableForStorage               *bool
		CreatedAt                      *time.Time
	}

	// ValidInstrumentCreationRequestInput represents what a user could set as input for creating valid instruments.
	ValidInstrumentCreationRequestInput struct {
		_ struct{} `json:"-"`

		Name                           string `json:"name"`
		PluralName                     string `json:"pluralName"`
		Description                    string `json:"description"`
		IconPath                       string `json:"iconPath"`
		Slug                           string `json:"slug"`
		DisplayInSummaryLists          bool   `json:"displayInSummaryLists"`
		IncludeInGeneratedInstructions bool   `json:"includeInGeneratedInstructions"`
		UsableForStorage               bool   `json:"usableForStorage"`
	}

	// ValidInstrumentDatabaseCreationInput represents what a user could set as input for creating valid instruments.
	ValidInstrumentDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                             string `json:"-"`
		Name                           string `json:"-"`
		PluralName                     string `json:"-"`
		Description                    string `json:"-"`
		IconPath                       string `json:"-"`
		Slug                           string `json:"-"`
		DisplayInSummaryLists          bool   `json:"-"`
		UsableForStorage               bool   `json:"-"`
		IncludeInGeneratedInstructions bool   `json:"-"`
	}

	// ValidInstrumentUpdateRequestInput represents what a user could set as input for updating valid instruments.
	ValidInstrumentUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Name                           *string `json:"name,omitempty"`
		PluralName                     *string `json:"pluralName,omitempty"`
		Description                    *string `json:"description,omitempty"`
		IconPath                       *string `json:"iconPath,omitempty"`
		Slug                           *string `json:"slug,omitempty"`
		UsableForStorage               *bool   `json:"usableForStorage,omitempty"`
		DisplayInSummaryLists          *bool   `json:"displayInSummaryLists,omitempty"`
		IncludeInGeneratedInstructions *bool   `json:"includeInGeneratedInstructions,omitempty"`
	}

	// ValidInstrumentDataManager describes a structure capable of storing valid instruments permanently.
	ValidInstrumentDataManager interface {
		ValidInstrumentExists(ctx context.Context, validInstrumentID string) (bool, error)
		GetValidInstrument(ctx context.Context, validInstrumentID string) (*ValidInstrument, error)
		GetRandomValidInstrument(ctx context.Context) (*ValidInstrument, error)
		GetValidInstruments(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidInstrument], error)
		SearchForValidInstruments(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidInstrument], error)
		SearchForValidInstrumentsNotOwnedByAccount(ctx context.Context, accountID, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidInstrument], error)
		CreateValidInstrument(ctx context.Context, input *ValidInstrumentDatabaseCreationInput) (*ValidInstrument, error)
		UpdateValidInstrument(ctx context.Context, updated *ValidInstrument) error
		MarkValidInstrumentsAsIndexed(ctx context.Context, ids []string) error
		ArchiveValidInstrument(ctx context.Context, validInstrumentID string) error
		ScanValidInstrumentIDsForReindex(ctx context.Context, after string, limit int) ([]string, error)
		GetValidInstrumentsWithIDs(ctx context.Context, ids []string) ([]*ValidInstrument, error)
	}
)

// Update merges a ValidInstrumentUpdateRequestInput with a valid instrument.
func (x *ValidInstrument) Update(input *ValidInstrumentUpdateRequestInput) {
	if input.Name != nil && *input.Name != x.Name {
		x.Name = *input.Name
	}

	if input.PluralName != nil && *input.PluralName != x.PluralName {
		x.PluralName = *input.PluralName
	}

	if input.Description != nil && *input.Description != x.Description {
		x.Description = *input.Description
	}

	if input.IconPath != nil && *input.IconPath != x.IconPath {
		x.IconPath = *input.IconPath
	}

	if input.UsableForStorage != nil && *input.UsableForStorage != x.UsableForStorage {
		x.UsableForStorage = *input.UsableForStorage
	}

	if input.Slug != nil && *input.Slug != x.Slug {
		x.Slug = *input.Slug
	}

	if input.DisplayInSummaryLists != nil && *input.DisplayInSummaryLists != x.DisplayInSummaryLists {
		x.DisplayInSummaryLists = *input.DisplayInSummaryLists
	}

	if input.IncludeInGeneratedInstructions != nil && *input.IncludeInGeneratedInstructions != x.IncludeInGeneratedInstructions {
		x.IncludeInGeneratedInstructions = *input.IncludeInGeneratedInstructions
	}
}

var _ validation.ValidatableWithContext = (*ValidInstrumentCreationRequestInput)(nil)

// ValidateWithContext validates a ValidInstrumentCreationRequestInput.
func (x *ValidInstrumentCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidInstrumentDatabaseCreationInput)(nil)

// ValidateWithContext validates a ValidInstrumentDatabaseCreationInput.
func (x *ValidInstrumentDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Name, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidInstrumentUpdateRequestInput)(nil)

// ValidateWithContext validates a ValidInstrumentUpdateRequestInput.
func (x *ValidInstrumentUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}
