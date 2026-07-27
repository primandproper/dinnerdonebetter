package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v7/filtering"
	"github.com/primandproper/platform-go/v7/numbers"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ValidMeasurementUnitConversionCreatedServiceEventType indicates a valid measurement conversion was created.
	ValidMeasurementUnitConversionCreatedServiceEventType = "valid_measurement_unit_conversion_created"
	// ValidMeasurementUnitConversionUpdatedServiceEventType indicates a valid measurement conversion was updated.
	ValidMeasurementUnitConversionUpdatedServiceEventType = "valid_measurement_unit_conversion_updated"
	// ValidMeasurementUnitConversionArchivedServiceEventType indicates a valid measurement conversion was archived.
	ValidMeasurementUnitConversionArchivedServiceEventType = "valid_measurement_unit_conversion_archived"
)

func init() {
	gob.Register(new(ValidMeasurementUnitConversion))
	gob.Register(new(ValidMeasurementUnitConversionCreationRequestInput))
	gob.Register(new(ValidMeasurementUnitConversionUpdateRequestInput))
}

type (
	// MeasurementUnitConversionMismatch represents an ingredient that has multiple valid measurement units
	// but lacks a conversion between them in the database.
	MeasurementUnitConversionMismatch struct {
		_          struct{}             `json:"-"`
		FromUnit   ValidMeasurementUnit `json:"fromUnit"`
		ToUnit     ValidMeasurementUnit `json:"toUnit"`
		Ingredient ValidIngredient      `json:"ingredient"`
	}

	// ValidMeasurementUnitConversion represents a valid measurement conversion.
	ValidMeasurementUnitConversion struct {
		_ struct{} `json:"-"`

		CreatedAt         time.Time            `json:"createdAt"`
		LastUpdatedAt     *time.Time           `json:"lastUpdatedAt"`
		ArchivedAt        *time.Time           `json:"archivedAt"`
		OnlyForIngredient *ValidIngredient     `json:"onlyForIngredient"`
		Notes             string               `json:"notes"`
		ID                string               `json:"id"`
		From              ValidMeasurementUnit `json:"from"`
		To                ValidMeasurementUnit `json:"to"`
		Modifier          float32              `json:"modifier"`
	}

	// ValidMeasurementUnitConversionCreationRequestInput represents what a user could set as input for creating valid measurement conversions.
	ValidMeasurementUnitConversionCreationRequestInput struct {
		_ struct{} `json:"-"`

		OnlyForIngredient *string `json:"onlyForIngredient"`
		From              string  `json:"from"`
		To                string  `json:"to"`
		Notes             string  `json:"notes"`
		Modifier          float32 `json:"modifier"`
	}

	// ValidMeasurementUnitConversionDatabaseCreationInput represents what a user could set as input for creating valid measurement conversions.
	ValidMeasurementUnitConversionDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		OnlyForIngredient *string `json:"-"`
		ID                string  `json:"-"`
		From              string  `json:"-"`
		To                string  `json:"-"`
		Notes             string  `json:"-"`
		Modifier          float32 `json:"-"`
	}

	// ValidMeasurementUnitConversionUpdateRequestInput represents what a user could set as input for updating valid measurement conversions.
	ValidMeasurementUnitConversionUpdateRequestInput struct {
		_ struct{} `json:"-"`

		From              *string  `json:"from,omitempty"`
		To                *string  `json:"to,omitempty"`
		OnlyForIngredient *string  `json:"onlyForIngredient,omitempty"`
		Modifier          *float32 `json:"modifier,omitempty"`
		Notes             *string  `json:"notes,omitempty"`
	}

	// ValidMeasurementUnitConversionDataManager describes a structure capable of storing valid measurement conversions permanently.
	ValidMeasurementUnitConversionDataManager interface {
		ValidMeasurementUnitConversionExists(ctx context.Context, validMeasurementUnitConversionID string) (bool, error)
		GetValidMeasurementUnitConversion(ctx context.Context, validMeasurementUnitConversionID string) (*ValidMeasurementUnitConversion, error)
		CreateValidMeasurementUnitConversion(ctx context.Context, input *ValidMeasurementUnitConversionDatabaseCreationInput) (*ValidMeasurementUnitConversion, error)
		UpdateValidMeasurementUnitConversion(ctx context.Context, updated *ValidMeasurementUnitConversion) error
		ArchiveValidMeasurementUnitConversion(ctx context.Context, validMeasurementUnitConversionID string) error
		GetValidMeasurementUnitConversionsForUnit(ctx context.Context, validMeasurementUnitID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidMeasurementUnitConversion], error)
		GetValidMeasurementUnitConversionsForIngredients(ctx context.Context, validIngredientIDs []string) ([]*ValidMeasurementUnitConversion, error)
		GetMeasurementUnitConversionMismatches(ctx context.Context) ([]*MeasurementUnitConversionMismatch, error)
	}
)

// Update merges a ValidMeasurementUnitConversionUpdateRequestInput with a valid measurement conversion.
func (x *ValidMeasurementUnitConversion) Update(input *ValidMeasurementUnitConversionUpdateRequestInput) {
	if input.From != nil && *input.From != x.From.ID {
		x.From.ID = *input.From
	}

	if input.To != nil && *input.To != x.To.ID {
		x.To.ID = *input.To
	}

	if input.OnlyForIngredient != nil && x.OnlyForIngredient != nil && *input.OnlyForIngredient != x.OnlyForIngredient.ID {
		x.OnlyForIngredient.ID = *input.OnlyForIngredient
	}

	if input.Modifier != nil && *input.Modifier != x.Modifier {
		x.Modifier = *input.Modifier
	}

	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}
}

// ConvertFromToTo converts a value from the "From" unit to the "To" unit.
// The optional precision parameter specifies the number of decimal places to round to (default: 2).
// For example, if this conversion is from cups to milliliters with a modifier of 236.588,
// calling ConvertFromToTo(2) would return 473.18 (2 cups = 473.18 milliliters, rounded to 2 decimal places).
func (x *ValidMeasurementUnitConversion) ConvertFromToTo(value float32, precision ...uint8) float32 {
	result := value * x.Modifier

	p := uint8(2)
	if len(precision) > 0 {
		p = precision[0]
	}

	return numbers.RoundToDecimalPlaces(result, p)
}

// ConvertToToFrom converts a value from the "To" unit back to the "From" unit.
// The optional precision parameter specifies the number of decimal places to round to (default: 2).
// For example, if this conversion is from cups to milliliters with a modifier of 236.588,
// calling ConvertToToFrom(473.176) would return 2.0 (473.176 milliliters = 2.0 cups, rounded to 2 decimal places).
func (x *ValidMeasurementUnitConversion) ConvertToToFrom(value float32, precision ...uint8) float32 {
	result := value / x.Modifier

	p := uint8(2)
	if len(precision) > 0 {
		p = precision[0]
	}

	return numbers.RoundToDecimalPlaces(result, p)
}

var _ validation.ValidatableWithContext = (*ValidMeasurementUnitConversionCreationRequestInput)(nil)

// ValidateWithContext validates a ValidMeasurementUnitConversionCreationRequestInput.
func (x *ValidMeasurementUnitConversionCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	if x.From == x.To {
		return validation.NewError("from", "cannot be the same as to")
	}

	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.From, validation.Required),
		validation.Field(&x.To, validation.Required),
		validation.Field(&x.Modifier, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidMeasurementUnitConversionDatabaseCreationInput)(nil)

// ValidateWithContext validates a ValidMeasurementUnitConversionDatabaseCreationInput.
func (x *ValidMeasurementUnitConversionDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.From, validation.Required),
		validation.Field(&x.To, validation.Required),
		validation.Field(&x.Modifier, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidMeasurementUnitConversionUpdateRequestInput)(nil)

// ValidateWithContext validates a ValidMeasurementUnitConversionUpdateRequestInput.
func (x *ValidMeasurementUnitConversionUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.From, validation.Required),
		validation.Field(&x.To, validation.Required),
		validation.Field(&x.Modifier, validation.Required),
	)
}
