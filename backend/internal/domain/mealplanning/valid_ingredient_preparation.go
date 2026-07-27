package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v7/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ValidIngredientPreparationCreatedServiceEventType indicates a valid ingredient preparation was created.
	ValidIngredientPreparationCreatedServiceEventType = "valid_ingredient_preparation_created"
	// ValidIngredientPreparationUpdatedServiceEventType indicates a valid ingredient preparation was updated.
	ValidIngredientPreparationUpdatedServiceEventType = "valid_ingredient_preparation_updated"
	// ValidIngredientPreparationArchivedServiceEventType indicates a valid ingredient preparation was archived.
	ValidIngredientPreparationArchivedServiceEventType = "valid_ingredient_preparation_archived"
)

func init() {
	gob.Register(new(ValidIngredientPreparation))
	gob.Register(new(ValidIngredientPreparationCreationRequestInput))
	gob.Register(new(ValidIngredientPreparationUpdateRequestInput))
}

type (
	// ValidIngredientPreparation represents a valid ingredient preparation.
	ValidIngredientPreparation struct {
		_ struct{} `json:"-"`

		CreatedAt     time.Time        `json:"createdAt"`
		LastUpdatedAt *time.Time       `json:"lastUpdatedAt"`
		ArchivedAt    *time.Time       `json:"archivedAt"`
		Notes         string           `json:"notes"`
		ID            string           `json:"id"`
		Preparation   ValidPreparation `json:"preparation"`
		Ingredient    ValidIngredient  `json:"ingredient"`
	}

	// ValidIngredientPreparationCreationRequestInput represents what a user could set as input for creating valid ingredient preparations.
	ValidIngredientPreparationCreationRequestInput struct {
		_ struct{} `json:"-"`

		Notes              string `json:"notes"`
		ValidPreparationID string `json:"validPreparationID"`
		ValidIngredientID  string `json:"validIngredientID"`
	}

	// ValidIngredientPreparationDatabaseCreationInput represents what a user could set as input for creating valid ingredient preparations.
	ValidIngredientPreparationDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                 string `json:"-"`
		Notes              string `json:"-"`
		ValidPreparationID string `json:"-"`
		ValidIngredientID  string `json:"-"`
	}

	// ValidIngredientPreparationUpdateRequestInput represents what a user could set as input for updating valid ingredient preparations.
	ValidIngredientPreparationUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Notes              *string `json:"notes,omitempty"`
		ValidPreparationID *string `json:"validPreparationID,omitempty"`
		ValidIngredientID  *string `json:"validIngredientID,omitempty"`
	}

	// ValidIngredientPreparationDataManager describes a structure capable of storing valid ingredient preparations permanently.
	ValidIngredientPreparationDataManager interface {
		ValidIngredientPreparationExists(ctx context.Context, validIngredientPreparationID string) (bool, error)
		GetValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string) (*ValidIngredientPreparation, error)
		GetValidIngredientPreparations(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidIngredientPreparation], error)
		GetValidIngredientPreparationsForIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidIngredientPreparation], error)
		GetValidIngredientPreparationsForPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidIngredientPreparation], error)
		GetValidIngredientPreparationsByIDs(ctx context.Context, ids []string) (map[string]*ValidIngredientPreparation, error)
		CreateValidIngredientPreparation(ctx context.Context, input *ValidIngredientPreparationDatabaseCreationInput) (*ValidIngredientPreparation, error)
		UpdateValidIngredientPreparation(ctx context.Context, updated *ValidIngredientPreparation) error
		ArchiveValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string) error
	}
)

// Update merges a ValidIngredientPreparationUpdateRequestInput with a valid ingredient preparation.
func (x *ValidIngredientPreparation) Update(input *ValidIngredientPreparationUpdateRequestInput) {
	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}

	if input.ValidPreparationID != nil && *input.ValidPreparationID != x.Preparation.ID {
		x.Preparation.ID = *input.ValidPreparationID
	}

	if input.ValidIngredientID != nil && *input.ValidIngredientID != x.Ingredient.ID {
		x.Ingredient.ID = *input.ValidIngredientID
	}
}

var _ validation.ValidatableWithContext = (*ValidIngredientPreparationCreationRequestInput)(nil)

// ValidateWithContext validates a ValidIngredientPreparationCreationRequestInput.
func (x *ValidIngredientPreparationCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ValidPreparationID, validation.Required),
		validation.Field(&x.ValidIngredientID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidIngredientPreparationDatabaseCreationInput)(nil)

// ValidateWithContext validates a ValidIngredientPreparationDatabaseCreationInput.
func (x *ValidIngredientPreparationDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.ValidPreparationID, validation.Required),
		validation.Field(&x.ValidIngredientID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidIngredientPreparationUpdateRequestInput)(nil)

// ValidateWithContext validates a ValidIngredientPreparationUpdateRequestInput.
func (x *ValidIngredientPreparationUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ValidPreparationID, validation.Required),
		validation.Field(&x.ValidIngredientID, validation.Required),
	)
}
