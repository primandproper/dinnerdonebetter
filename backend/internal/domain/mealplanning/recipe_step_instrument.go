package mealplanning

import (
	"context"
	"encoding/gob"
	"errors"
	"time"

	"github.com/primandproper/platform-go/v8/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

const (
	// RecipeStepInstrumentCreatedServiceEventType indicates a recipe step instrument was created.
	RecipeStepInstrumentCreatedServiceEventType = "recipe_step_instrument_created"
	// RecipeStepInstrumentUpdatedServiceEventType indicates a recipe step instrument was updated.
	RecipeStepInstrumentUpdatedServiceEventType = "recipe_step_instrument_updated"
	// RecipeStepInstrumentArchivedServiceEventType indicates a recipe step instrument was archived.
	RecipeStepInstrumentArchivedServiceEventType = "recipe_step_instrument_archived"
)

var (
	errValidPreparationInstrumentIDRequired = errors.New("validPreparationInstrumentID is required when not referencing a recipe step product")
)

func init() {
	gob.Register(new(RecipeStepInstrument))
	gob.Register(new(RecipeStepInstrumentCreationRequestInput))
	gob.Register(new(RecipeStepInstrumentUpdateRequestInput))
}

type (
	// RecipeStepInstrument represents a recipe step instrument.
	RecipeStepInstrument struct {
		_ struct{} `json:"-"`

		CreatedAt           time.Time        `json:"createdAt"`
		MaxQuantity         *uint32          `json:"maxQuantity,omitempty"`
		Instrument          *ValidInstrument `json:"instrument"`
		LastUpdatedAt       *time.Time       `json:"lastUpdatedAt"`
		RecipeStepProductID *string          `json:"recipeStepProductID"`
		ArchivedAt          *time.Time       `json:"archivedAt"`
		Notes               string           `json:"notes"`
		BelongsToRecipeStep string           `json:"belongsToRecipeStep"`
		ID                  string           `json:"id"`
		Name                string           `json:"name"`
		MinQuantity         uint32           `json:"minQuantity"`
		Index               uint16           `json:"index"`
		OptionIndex         uint16           `json:"optionIndex"`
		PreferenceRank      uint8            `json:"preferenceRank"`
		Optional            bool             `json:"optional"`
		ScaleFactor         float32          `json:"scaleFactor"`
	}

	// RecipeStepInstrumentCreationRequestInput represents what a user could set as input for creating recipe step instruments.
	RecipeStepInstrumentCreationRequestInput struct {
		_ struct{} `json:"-"`

		MaxQuantity                     *uint32 `json:"maxQuantity,omitempty"`
		ValidPreparationInstrumentID    *string `json:"validPreparationInstrumentID"`
		RecipeStepProductID             *string `json:"recipeStepProductID"`
		ProductOfRecipeStepIndex        *uint64 `json:"productOfRecipeStepIndex"`
		ProductOfRecipeStepProductIndex *uint64 `json:"productOfRecipeStepProductIndex"`
		Index                           *uint16 `json:"index,omitempty"`
		Notes                           string  `json:"notes"`
		Name                            string  `json:"name"`
		MinQuantity                     uint32  `json:"minQuantity"`
		OptionIndex                     uint16  `json:"optionIndex"`
		Optional                        bool    `json:"optional"`
		PreferenceRank                  uint8   `json:"preferenceRank"`
		ScaleFactor                     float32 `json:"scaleFactor"`
	}

	// RecipeStepInstrumentDatabaseCreationInput represents what a user could set as input for creating recipe step instruments.
	RecipeStepInstrumentDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		MaxQuantity                     *uint32 `json:"-"`
		ProductOfRecipeStepIndex        *uint64 `json:"-"`
		RecipeStepProductID             *string `json:"-"`
		ProductOfRecipeStepProductIndex *uint64 `json:"-"`
		InstrumentID                    *string `json:"-"`
		ValidPreparationInstrumentID    *string `json:"-"`
		BelongsToRecipeStep             string  `json:"-"`
		Name                            string  `json:"-"`
		ID                              string  `json:"-"`
		Notes                           string  `json:"-"`
		MinQuantity                     uint32  `json:"-"`
		Index                           uint16  `json:"-"`
		OptionIndex                     uint16  `json:"-"`
		Optional                        bool    `json:"-"`
		PreferenceRank                  uint8   `json:"-"`
		ScaleFactor                     float32 `json:"-"`
	}

	// RecipeStepInstrumentUpdateRequestInput represents what a user could set as input for updating recipe step instruments.
	RecipeStepInstrumentUpdateRequestInput struct {
		_ struct{} `json:"-"`

		InstrumentID        *string  `json:"instrumentID,omitempty"`
		RecipeStepProductID *string  `json:"recipeStepProductID,omitempty"`
		Notes               *string  `json:"notes,omitempty"`
		PreferenceRank      *uint8   `json:"preferenceRank,omitempty"`
		BelongsToRecipeStep *string  `json:"belongsToRecipeStep,omitempty"`
		Name                *string  `json:"name,omitempty"`
		Optional            *bool    `json:"optional,omitempty"`
		Index               *uint16  `json:"index,omitempty"`
		OptionIndex         *uint16  `json:"optionIndex,omitempty"`
		MinQuantity         *uint32  `json:"minQuantity,omitempty"`
		MaxQuantity         *uint32  `json:"maxQuantity,omitempty"`
		ScaleFactor         *float32 `json:"scaleFactor,omitempty"`
	}

	// RecipeStepInstrumentDataManager describes a structure capable of storing recipe step instruments permanently.
	RecipeStepInstrumentDataManager interface {
		RecipeStepInstrumentExists(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string) (bool, error)
		GetRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string) (*RecipeStepInstrument, error)
		GetRecipeStepInstruments(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[RecipeStepInstrument], error)
		CreateRecipeStepInstrument(ctx context.Context, input *RecipeStepInstrumentDatabaseCreationInput) (*RecipeStepInstrument, error)
		UpdateRecipeStepInstrument(ctx context.Context, updated *RecipeStepInstrument) error
		ArchiveRecipeStepInstrument(ctx context.Context, recipeStepID, recipeStepInstrumentID string) error
	}
)

// Update merges an RecipeStepInstrumentUpdateRequestInput with a recipe step instrument.
func (x *RecipeStepInstrument) Update(input *RecipeStepInstrumentUpdateRequestInput) {
	if input.InstrumentID != nil && (x.Instrument == nil || (*input.InstrumentID != "" && *input.InstrumentID != x.Instrument.ID)) {
		x.Instrument = &ValidInstrument{ID: *input.InstrumentID}
	}

	if input.RecipeStepProductID != nil && x.RecipeStepProductID != nil && *input.RecipeStepProductID != *x.RecipeStepProductID {
		x.RecipeStepProductID = input.RecipeStepProductID
	}

	if input.Name != nil && *input.Name != x.Name {
		x.Name = *input.Name
	}

	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}

	if input.PreferenceRank != nil && *input.PreferenceRank != x.PreferenceRank {
		x.PreferenceRank = *input.PreferenceRank
	}

	if input.Optional != nil && *input.Optional != x.Optional {
		x.Optional = *input.Optional
	}

	if input.MinQuantity != nil && *input.MinQuantity != x.MinQuantity {
		x.MinQuantity = *input.MinQuantity
	}

	if input.MaxQuantity != nil && (x.MaxQuantity == nil || *input.MaxQuantity != *x.MaxQuantity) {
		x.MaxQuantity = input.MaxQuantity
	}

	if input.Index != nil && *input.Index != x.Index {
		x.Index = *input.Index
	}

	if input.OptionIndex != nil && *input.OptionIndex != x.OptionIndex {
		x.OptionIndex = *input.OptionIndex
	}

	if input.ScaleFactor != nil && *input.ScaleFactor != x.ScaleFactor {
		x.ScaleFactor = *input.ScaleFactor
	}
}

var _ validation.ValidatableWithContext = (*RecipeStepInstrumentCreationRequestInput)(nil)

// ValidateWithContext validates a RecipeStepInstrumentCreationRequestInput.
func (x *RecipeStepInstrumentCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	err := &multierror.Error{}

	// When not referencing a recipe step product, bridge table ID is required
	if isRecipeStepProduct := x.ProductOfRecipeStepIndex != nil; !isRecipeStepProduct {
		if x.ValidPreparationInstrumentID == nil || *x.ValidPreparationInstrumentID == "" {
			err = multierror.Append(err, errValidPreparationInstrumentIDRequired)
		}
	}

	validationErr := validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
	if validationErr != nil {
		err = multierror.Append(err, validationErr)
	}

	return err.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*RecipeStepInstrumentDatabaseCreationInput)(nil)

// ValidateWithContext validates a RecipeStepInstrumentDatabaseCreationInput.
func (x *RecipeStepInstrumentDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.InstrumentID, validation.Required),
		validation.Field(&x.BelongsToRecipeStep, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*RecipeStepInstrumentUpdateRequestInput)(nil)

// ValidateWithContext validates a RecipeStepInstrumentUpdateRequestInput.
func (x *RecipeStepInstrumentUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.BelongsToRecipeStep, validation.Required),
	)
}
