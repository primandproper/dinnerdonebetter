package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v5/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// RecipeStepProductIngredientType represents one of the valid recipe step product type values.
	RecipeStepProductIngredientType = "ingredient"
	// RecipeStepProductInstrumentType represents one of the valid recipe step product type values.
	RecipeStepProductInstrumentType = "instrument"
	// RecipeStepProductVesselType represents one of the valid recipe step product type values.
	RecipeStepProductVesselType = "vessel"

	// RecipeStepProductCreatedServiceEventType indicates a recipe step product was created.
	RecipeStepProductCreatedServiceEventType = "recipe_step_product_created"
	// RecipeStepProductUpdatedServiceEventType indicates a recipe step product was updated.
	RecipeStepProductUpdatedServiceEventType = "recipe_step_product_updated"
	// RecipeStepProductArchivedServiceEventType indicates a recipe step product was archived.
	RecipeStepProductArchivedServiceEventType = "recipe_step_product_archived"
)

func init() {
	gob.Register(new(RecipeStepProduct))
	gob.Register(new(RecipeStepProductCreationRequestInput))
	gob.Register(new(RecipeStepProductUpdateRequestInput))
}

type (
	// RecipeStepProduct represents a recipe step product.
	RecipeStepProduct struct {
		_ struct{} `json:"-"`

		CreatedAt                      time.Time             `json:"createdAt"`
		MinStorageTemperatureInCelsius *float32              `json:"minStorageTemperatureInCelsius,omitempty"`
		MaxStorageTemperatureInCelsius *float32              `json:"maxStorageTemperatureInCelsius,omitempty"`
		MinStorageDurationInSeconds    *uint32               `json:"minStorageDurationInSeconds,omitempty"`
		MaxStorageDurationInSeconds    *uint32               `json:"maxStorageDurationInSeconds,omitempty"`
		MinMeasurementQuantity         *float32              `json:"minMeasurementQuantity,omitempty"`
		MaxMeasurementQuantity         *float32              `json:"maxMeasurementQuantity,omitempty"`
		MinItemQuantity                *float32              `json:"minItemQuantity,omitempty"`
		MaxItemQuantity                *float32              `json:"maxItemQuantity,omitempty"`
		ArchivedAt                     *time.Time            `json:"archivedAt"`
		LastUpdatedAt                  *time.Time            `json:"lastUpdatedAt"`
		MeasurementUnit                *ValidMeasurementUnit `json:"measurementUnit"`
		ContainedInVesselIndex         *uint16               `json:"containedInVesselIndex"`
		Name                           string                `json:"name"`
		BelongsToRecipeStep            string                `json:"belongsToRecipeStep"`
		Type                           string                `json:"type"`
		ID                             string                `json:"id"`
		StorageInstructions            string                `json:"storageInstructions"`
		QuantityNotes                  string                `json:"quantityNotes"`
		Index                          uint16                `json:"index"`
		IsWaste                        bool                  `json:"isWaste"`
		IsLiquid                       bool                  `json:"isLiquid"`
		Compostable                    bool                  `json:"compostable"`
	}

	// RecipeStepProductCreationRequestInput represents what a user could set as input for creating recipe step products.
	RecipeStepProductCreationRequestInput struct {
		_ struct{} `json:"-"`

		MinStorageTemperatureInCelsius *float32 `json:"minStorageTemperatureInCelsius,omitempty"`
		MaxStorageTemperatureInCelsius *float32 `json:"maxStorageTemperatureInCelsius,omitempty"`
		MinStorageDurationInSeconds    *uint32  `json:"minStorageDurationInSeconds,omitempty"`
		MaxStorageDurationInSeconds    *uint32  `json:"maxStorageDurationInSeconds,omitempty"`
		MinMeasurementQuantity         *float32 `json:"minMeasurementQuantity,omitempty"`
		MaxMeasurementQuantity         *float32 `json:"maxMeasurementQuantity,omitempty"`
		MinItemQuantity                *float32 `json:"minItemQuantity,omitempty"`
		MaxItemQuantity                *float32 `json:"maxItemQuantity,omitempty"`
		MeasurementUnitID              *string  `json:"measurementUnitID"`
		ContainedInVesselIndex         *uint16  `json:"containedInVesselIndex"`
		QuantityNotes                  string   `json:"quantityNotes"`
		Name                           string   `json:"name"`
		StorageInstructions            string   `json:"storageInstructions"`
		Type                           string   `json:"type"`
		Index                          uint16   `json:"index"`
		Compostable                    bool     `json:"compostable"`
		IsLiquid                       bool     `json:"isLiquid"`
		IsWaste                        bool     `json:"isWaste"`
	}

	// RecipeStepProductDatabaseCreationInput represents what a user could set as input for creating recipe step products.
	RecipeStepProductDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		MinStorageTemperatureInCelsius *float32 `json:"-"`
		MaxStorageTemperatureInCelsius *float32 `json:"-"`
		MinStorageDurationInSeconds    *uint32  `json:"-"`
		MaxStorageDurationInSeconds    *uint32  `json:"-"`
		MinMeasurementQuantity         *float32 `json:"-"`
		MaxMeasurementQuantity         *float32 `json:"-"`
		MinItemQuantity                *float32 `json:"-"`
		MaxItemQuantity                *float32 `json:"-"`
		MeasurementUnitID              *string  `json:"-"`
		ContainedInVesselIndex         *uint16  `json:"-"`
		Name                           string   `json:"-"`
		BelongsToRecipeStep            string   `json:"-"`
		StorageInstructions            string   `json:"-"`
		QuantityNotes                  string   `json:"-"`
		ID                             string   `json:"-"`
		Type                           string   `json:"-"`
		Index                          uint16   `json:"-"`
		Compostable                    bool     `json:"-"`
		IsLiquid                       bool     `json:"-"`
		IsWaste                        bool     `json:"-"`
	}

	// RecipeStepProductUpdateRequestInput represents what a user could set as input for updating recipe step products.
	RecipeStepProductUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Name                           *string  `json:"name,omitempty"`
		Type                           *string  `json:"type,omitempty"`
		MeasurementUnitID              *string  `json:"measurementUnitID,omitempty"`
		QuantityNotes                  *string  `json:"quantityNotes,omitempty"`
		BelongsToRecipeStep            *string  `json:"belongsToRecipeStep,omitempty"`
		MinStorageTemperatureInCelsius *float32 `json:"minStorageTemperatureInCelsius,omitempty"`
		MaxStorageTemperatureInCelsius *float32 `json:"maxStorageTemperatureInCelsius,omitempty"`
		MinStorageDurationInSeconds    *uint32  `json:"minStorageDurationInSeconds,omitempty"`
		MaxStorageDurationInSeconds    *uint32  `json:"maxStorageDurationInSeconds,omitempty"`
		MinMeasurementQuantity         *float32 `json:"minMeasurementQuantity,omitempty"`
		MaxMeasurementQuantity         *float32 `json:"maxMeasurementQuantity,omitempty"`
		MinItemQuantity                *float32 `json:"minItemQuantity,omitempty"`
		MaxItemQuantity                *float32 `json:"maxItemQuantity,omitempty"`
		Compostable                    *bool    `json:"compostable,omitempty"`
		StorageInstructions            *string  `json:"storageInstructions,omitempty"`
		IsLiquid                       *bool    `json:"isLiquid,omitempty"`
		IsWaste                        *bool    `json:"isWaste,omitempty"`
		Index                          *uint16  `json:"index,omitempty"`
		ContainedInVesselIndex         *uint16  `json:"containedInVesselIndex,omitempty"`
	}

	// RecipeStepProductDataManager describes a structure capable of storing recipe step products permanently.
	RecipeStepProductDataManager interface {
		RecipeStepProductExists(ctx context.Context, recipeID, recipeStepID, recipeStepProductID string) (bool, error)
		GetRecipeStepProduct(ctx context.Context, recipeID, recipeStepID, recipeStepProductID string) (*RecipeStepProduct, error)
		GetRecipeStepProducts(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[RecipeStepProduct], error)
		CreateRecipeStepProduct(ctx context.Context, input *RecipeStepProductDatabaseCreationInput) (*RecipeStepProduct, error)
		UpdateRecipeStepProduct(ctx context.Context, updated *RecipeStepProduct) error
		ArchiveRecipeStepProduct(ctx context.Context, recipeStepID, recipeStepProductID string) error
	}
)

// Update merges an RecipeStepProductUpdateRequestInput with a recipe step product.
func (x *RecipeStepProduct) Update(input *RecipeStepProductUpdateRequestInput) {
	if input.Name != nil && *input.Name != x.Name {
		x.Name = *input.Name
	}

	if input.Type != nil && *input.Type != x.Type {
		x.Type = *input.Type
	}

	if input.MeasurementUnitID != nil && x.MeasurementUnit != nil && *input.MeasurementUnitID != x.MeasurementUnit.ID {
		x.MeasurementUnit = &ValidMeasurementUnit{ID: *input.MeasurementUnitID}
	}

	if input.MinMeasurementQuantity != nil && input.MinMeasurementQuantity != x.MinMeasurementQuantity {
		x.MinMeasurementQuantity = input.MinMeasurementQuantity
	}

	if input.MaxMeasurementQuantity != nil && input.MaxMeasurementQuantity != x.MaxMeasurementQuantity {
		x.MaxMeasurementQuantity = input.MaxMeasurementQuantity
	}

	if input.MinItemQuantity != nil && input.MinItemQuantity != x.MinItemQuantity {
		x.MinItemQuantity = input.MinItemQuantity
	}

	if input.MaxItemQuantity != nil && input.MaxItemQuantity != x.MaxItemQuantity {
		x.MaxItemQuantity = input.MaxItemQuantity
	}

	if input.QuantityNotes != nil && *input.QuantityNotes != x.QuantityNotes {
		x.QuantityNotes = *input.QuantityNotes
	}

	if input.Compostable != nil && *input.Compostable != x.Compostable {
		x.Compostable = *input.Compostable
	}

	if input.MinStorageDurationInSeconds != nil && input.MinStorageDurationInSeconds != x.MinStorageDurationInSeconds {
		x.MinStorageDurationInSeconds = input.MinStorageDurationInSeconds
	}

	if input.MaxStorageDurationInSeconds != nil && input.MaxStorageDurationInSeconds != x.MaxStorageDurationInSeconds {
		x.MaxStorageDurationInSeconds = input.MaxStorageDurationInSeconds
	}

	if input.MinStorageTemperatureInCelsius != nil && input.MinStorageTemperatureInCelsius != x.MinStorageTemperatureInCelsius {
		x.MinStorageTemperatureInCelsius = input.MinStorageTemperatureInCelsius
	}

	if input.MaxStorageTemperatureInCelsius != nil && input.MaxStorageTemperatureInCelsius != x.MaxStorageTemperatureInCelsius {
		x.MaxStorageTemperatureInCelsius = input.MaxStorageTemperatureInCelsius
	}

	if input.StorageInstructions != nil && *input.StorageInstructions != x.StorageInstructions {
		x.StorageInstructions = *input.StorageInstructions
	}

	if input.IsLiquid != nil && *input.IsLiquid != x.IsLiquid {
		x.IsLiquid = *input.IsLiquid
	}

	if input.IsWaste != nil && *input.IsWaste != x.IsWaste {
		x.IsWaste = *input.IsWaste
	}

	if input.Index != nil && *input.Index != x.Index {
		x.Index = *input.Index
	}

	if input.ContainedInVesselIndex != nil && x.ContainedInVesselIndex != nil && *input.ContainedInVesselIndex != *x.ContainedInVesselIndex {
		x.ContainedInVesselIndex = input.ContainedInVesselIndex
	}
}

var _ validation.ValidatableWithContext = (*RecipeStepProductCreationRequestInput)(nil)

// ValidateWithContext validates a RecipeStepProductCreationRequestInput.
func (x *RecipeStepProductCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Type, validation.In(RecipeStepProductIngredientType, RecipeStepProductInstrumentType, RecipeStepProductVesselType)),
		validation.Field(&x.MinMeasurementQuantity, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*RecipeStepProductDatabaseCreationInput)(nil)

// ValidateWithContext validates a RecipeStepProductDatabaseCreationInput.
func (x *RecipeStepProductDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.BelongsToRecipeStep, validation.Required),
		validation.Field(&x.Type, validation.In(RecipeStepProductIngredientType, RecipeStepProductInstrumentType, RecipeStepProductVesselType)),
	)
}

var _ validation.ValidatableWithContext = (*RecipeStepProductUpdateRequestInput)(nil)

// ValidateWithContext validates a RecipeStepProductUpdateRequestInput.
func (x *RecipeStepProductUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Type, validation.In(RecipeStepProductIngredientType, RecipeStepProductInstrumentType, RecipeStepProductVesselType)),
		validation.Field(&x.MeasurementUnitID, validation.Required),
		validation.Field(&x.MinMeasurementQuantity, validation.Required),
	)
}
