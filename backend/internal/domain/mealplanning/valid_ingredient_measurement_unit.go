package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v7/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ValidIngredientMeasurementUnitCreatedServiceEventType indicates a valid ingredient measurement unit was created.
	ValidIngredientMeasurementUnitCreatedServiceEventType = "valid_ingredient_measurement_unit_created"
	// ValidIngredientMeasurementUnitUpdatedServiceEventType indicates a valid ingredient measurement unit was updated.
	ValidIngredientMeasurementUnitUpdatedServiceEventType = "valid_ingredient_measurement_unit_updated"
	// ValidIngredientMeasurementUnitArchivedServiceEventType indicates a valid ingredient measurement unit was archived.
	ValidIngredientMeasurementUnitArchivedServiceEventType = "valid_ingredient_measurement_unit_archived"
)

func init() {
	gob.Register(new(ValidIngredientMeasurementUnit))
	gob.Register(new(ValidIngredientMeasurementUnitCreationRequestInput))
	gob.Register(new(ValidIngredientMeasurementUnitUpdateRequestInput))
}

type (
	// ValidIngredientMeasurementUnit represents a valid ingredient measurement unit.
	ValidIngredientMeasurementUnit struct {
		_                    struct{}             `json:"-"`
		CreatedAt            time.Time            `json:"createdAt"`
		LastUpdatedAt        *time.Time           `json:"lastUpdatedAt"`
		ArchivedAt           *time.Time           `json:"archivedAt"`
		MaxAllowableQuantity *float32             `json:"maxAllowableQuantity,omitempty"`
		Notes                string               `json:"notes"`
		ID                   string               `json:"id"`
		MeasurementUnit      ValidMeasurementUnit `json:"measurementUnit"`
		Ingredient           ValidIngredient      `json:"ingredient"`
		MinAllowableQuantity float32              `json:"minAllowableQuantity"`
	}

	// ValidIngredientMeasurementUnitCreationRequestInput represents what a user could set as input for creating valid ingredient measurement units.
	ValidIngredientMeasurementUnitCreationRequestInput struct {
		_                      struct{} `json:"-"`
		MaxAllowableQuantity   *float32 `json:"maxAllowableQuantity,omitempty"`
		Notes                  string   `json:"notes"`
		ValidMeasurementUnitID string   `json:"validMeasurementUnitID"`
		ValidIngredientID      string   `json:"validIngredientID"`
		MinAllowableQuantity   float32  `json:"minAllowableQuantity"`
	}

	// ValidIngredientMeasurementUnitDatabaseCreationInput represents what a user could set as input for creating valid ingredient measurement units.
	ValidIngredientMeasurementUnitDatabaseCreationInput struct {
		_                      struct{} `json:"-"`
		MaxAllowableQuantity   *float32 `json:"-"`
		ID                     string   `json:"-"`
		Notes                  string   `json:"-"`
		ValidMeasurementUnitID string   `json:"-"`
		ValidIngredientID      string   `json:"-"`
		MinAllowableQuantity   float32  `json:"-"`
	}

	// ValidIngredientMeasurementUnitUpdateRequestInput represents what a user could set as input for updating valid ingredient measurement units.
	ValidIngredientMeasurementUnitUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Notes                  *string  `json:"notes,omitempty"`
		ValidMeasurementUnitID *string  `json:"validMeasurementUnitID,omitempty"`
		ValidIngredientID      *string  `json:"validIngredientID,omitempty"`
		MinAllowableQuantity   *float32 `json:"minAllowableQuantity,omitempty"`
		MaxAllowableQuantity   *float32 `json:"maxAllowableQuantity,omitempty"`
	}

	// ValidIngredientMeasurementUnitDataManager describes a structure capable of storing valid ingredient measurement units permanently.
	ValidIngredientMeasurementUnitDataManager interface {
		ValidIngredientMeasurementUnitExists(ctx context.Context, validIngredientMeasurementUnitID string) (bool, error)
		GetValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string) (*ValidIngredientMeasurementUnit, error)
		GetValidIngredientMeasurementUnits(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidIngredientMeasurementUnit], error)
		GetValidIngredientMeasurementUnitsForIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidIngredientMeasurementUnit], error)
		GetValidIngredientMeasurementUnitsForMeasurementUnit(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidIngredientMeasurementUnit], error)
		GetValidIngredientMeasurementUnitsByIDs(ctx context.Context, ids []string) (map[string]*ValidIngredientMeasurementUnit, error)
		CreateValidIngredientMeasurementUnit(ctx context.Context, input *ValidIngredientMeasurementUnitDatabaseCreationInput) (*ValidIngredientMeasurementUnit, error)
		UpdateValidIngredientMeasurementUnit(ctx context.Context, updated *ValidIngredientMeasurementUnit) error
		ArchiveValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string) error
	}
)

// Update merges a ValidIngredientMeasurementUnitUpdateRequestInput with a valid ingredient measurement unit.
func (x *ValidIngredientMeasurementUnit) Update(input *ValidIngredientMeasurementUnitUpdateRequestInput) {
	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}

	if input.ValidMeasurementUnitID != nil && *input.ValidMeasurementUnitID != x.MeasurementUnit.ID {
		x.MeasurementUnit.ID = *input.ValidMeasurementUnitID
	}

	if input.ValidIngredientID != nil && *input.ValidIngredientID != x.Ingredient.ID {
		x.Ingredient.ID = *input.ValidIngredientID
	}

	if input.MinAllowableQuantity != nil && *input.MinAllowableQuantity != x.MinAllowableQuantity {
		x.MinAllowableQuantity = *input.MinAllowableQuantity
	}

	if input.MaxAllowableQuantity != nil && (x.MaxAllowableQuantity == nil || *input.MaxAllowableQuantity != *x.MaxAllowableQuantity) {
		x.MaxAllowableQuantity = input.MaxAllowableQuantity
	}
}

var _ validation.ValidatableWithContext = (*ValidIngredientMeasurementUnitCreationRequestInput)(nil)

// ValidateWithContext validates a ValidIngredientMeasurementUnitCreationRequestInput.
func (x *ValidIngredientMeasurementUnitCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ValidMeasurementUnitID, validation.Required),
		validation.Field(&x.ValidIngredientID, validation.Required),
		validation.Field(&x.MinAllowableQuantity, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidIngredientMeasurementUnitDatabaseCreationInput)(nil)

// ValidateWithContext validates a ValidIngredientMeasurementUnitDatabaseCreationInput.
func (x *ValidIngredientMeasurementUnitDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.ValidMeasurementUnitID, validation.Required),
		validation.Field(&x.ValidIngredientID, validation.Required),
		validation.Field(&x.MinAllowableQuantity, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidIngredientMeasurementUnitUpdateRequestInput)(nil)

// ValidateWithContext validates a ValidIngredientMeasurementUnitUpdateRequestInput.
func (x *ValidIngredientMeasurementUnitUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ValidMeasurementUnitID, validation.Required),
		validation.Field(&x.ValidIngredientID, validation.Required),
	)
}
