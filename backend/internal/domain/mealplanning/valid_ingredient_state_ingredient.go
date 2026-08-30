package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v13/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ValidIngredientStateIngredientCreatedServiceEventType indicates a valid ingredient state ingredient was created.
	ValidIngredientStateIngredientCreatedServiceEventType = "valid_ingredient_state_ingredient_created"
	// ValidIngredientStateIngredientUpdatedServiceEventType indicates a valid ingredient state ingredient was updated.
	ValidIngredientStateIngredientUpdatedServiceEventType = "valid_ingredient_state_ingredient_updated"
	// ValidIngredientStateIngredientArchivedServiceEventType indicates a valid ingredient state ingredient was archived.
	ValidIngredientStateIngredientArchivedServiceEventType = "valid_ingredient_state_ingredient_archived"
)

func init() {
	gob.Register(new(ValidIngredientStateIngredient))
	gob.Register(new(ValidIngredientStateIngredientCreationRequestInput))
	gob.Register(new(ValidIngredientStateIngredientUpdateRequestInput))
}

type (
	// ValidIngredientStateIngredient represents a valid ingredient state ingredient.
	ValidIngredientStateIngredient struct {
		_ struct{} `json:"-"`

		CreatedAt       time.Time            `json:"createdAt"`
		LastUpdatedAt   *time.Time           `json:"lastUpdatedAt"`
		ArchivedAt      *time.Time           `json:"archivedAt"`
		Notes           string               `json:"notes"`
		ID              string               `json:"id"`
		IngredientState ValidIngredientState `json:"ingredientState"`
		Ingredient      ValidIngredient      `json:"ingredient"`
	}

	// ValidIngredientStateIngredientCreationRequestInput represents what a user could set as input for creating valid ingredient state ingredients.
	ValidIngredientStateIngredientCreationRequestInput struct {
		_ struct{} `json:"-"`

		Notes                  string `json:"notes"`
		ValidIngredientStateID string `json:"validIngredientStateID"`
		ValidIngredientID      string `json:"validIngredientID"`
	}

	// ValidIngredientStateIngredientDatabaseCreationInput represents what a user could set as input for creating valid ingredient state ingredients.
	ValidIngredientStateIngredientDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                     string `json:"-"`
		Notes                  string `json:"-"`
		ValidIngredientStateID string `json:"-"`
		ValidIngredientID      string `json:"-"`
	}

	// ValidIngredientStateIngredientUpdateRequestInput represents what a user could set as input for updating valid ingredient state ingredients.
	ValidIngredientStateIngredientUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Notes                  *string `json:"notes,omitempty"`
		ValidIngredientStateID *string `json:"validIngredientStateID,omitempty"`
		ValidIngredientID      *string `json:"validIngredientID,omitempty"`
	}

	// ValidIngredientStateIngredientDataManager describes a structure capable of storing valid ingredient state ingredients permanently.
	ValidIngredientStateIngredientDataManager interface {
		ValidIngredientStateIngredientExists(ctx context.Context, validIngredientPreparationID string) (bool, error)
		GetValidIngredientStateIngredient(ctx context.Context, validIngredientPreparationID string) (*ValidIngredientStateIngredient, error)
		GetValidIngredientStateIngredients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidIngredientStateIngredient], error)
		GetValidIngredientStateIngredientsForIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidIngredientStateIngredient], error)
		GetValidIngredientStateIngredientsForIngredientState(ctx context.Context, ingredientStateID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidIngredientStateIngredient], error)
		CreateValidIngredientStateIngredient(ctx context.Context, input *ValidIngredientStateIngredientDatabaseCreationInput) (*ValidIngredientStateIngredient, error)
		UpdateValidIngredientStateIngredient(ctx context.Context, updated *ValidIngredientStateIngredient) error
		ArchiveValidIngredientStateIngredient(ctx context.Context, validIngredientPreparationID string) error
	}
)

// Update merges a ValidIngredientStateIngredientUpdateRequestInput with a valid ingredient state ingredient.
func (x *ValidIngredientStateIngredient) Update(input *ValidIngredientStateIngredientUpdateRequestInput) {
	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}

	if input.ValidIngredientStateID != nil && *input.ValidIngredientStateID != x.IngredientState.ID {
		x.IngredientState.ID = *input.ValidIngredientStateID
	}

	if input.ValidIngredientID != nil && *input.ValidIngredientID != x.Ingredient.ID {
		x.Ingredient.ID = *input.ValidIngredientID
	}
}

var _ validation.ValidatableWithContext = (*ValidIngredientStateIngredientCreationRequestInput)(nil)

// ValidateWithContext validates a ValidIngredientStateIngredientCreationRequestInput.
func (x *ValidIngredientStateIngredientCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ValidIngredientStateID, validation.Required),
		validation.Field(&x.ValidIngredientID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidIngredientStateIngredientDatabaseCreationInput)(nil)

// ValidateWithContext validates a ValidIngredientStateIngredientDatabaseCreationInput.
func (x *ValidIngredientStateIngredientDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.ValidIngredientStateID, validation.Required),
		validation.Field(&x.ValidIngredientID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidIngredientStateIngredientUpdateRequestInput)(nil)

// ValidateWithContext validates a ValidIngredientStateIngredientUpdateRequestInput.
func (x *ValidIngredientStateIngredientUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ValidIngredientStateID, validation.Required),
		validation.Field(&x.ValidIngredientID, validation.Required),
	)
}
