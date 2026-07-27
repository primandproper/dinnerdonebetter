package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v7/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ValidPrepTaskConfigCreatedServiceEventType indicates a valid ingredient preparation storage config was created.
	ValidPrepTaskConfigCreatedServiceEventType = "valid_prep_task_config_created"
	// ValidPrepTaskConfigUpdatedServiceEventType indicates a valid ingredient preparation storage config was updated.
	ValidPrepTaskConfigUpdatedServiceEventType = "valid_prep_task_config_updated"
	// ValidPrepTaskConfigArchivedServiceEventType indicates a valid ingredient preparation storage config was archived.
	ValidPrepTaskConfigArchivedServiceEventType = "valid_prep_task_config_archived"
)

func init() {
	gob.Register(new(ValidPrepTaskConfig))
	gob.Register(new(ValidPrepTaskConfigCreationRequestInput))
	gob.Register(new(ValidPrepTaskConfigUpdateRequestInput))
}

type (
	// ValidPrepTaskConfig represents reusable knowledge about how long
	// a prepared ingredient can be stored under specific conditions.
	// Example: "diced onion" can be stored for 72 hours in an airtight container at 0-4°C.
	ValidPrepTaskConfig struct {
		_                              struct{}         `json:"-"`
		CreatedAt                      time.Time        `json:"createdAt"`
		MaxStorageTemperatureInCelsius *float32         `json:"maxStorageTemperatureInCelsius,omitempty"`
		ArchivedAt                     *time.Time       `json:"archivedAt"`
		MaxStorageDurationInSeconds    *uint32          `json:"maxStorageDurationInSeconds,omitempty"`
		MinStorageTemperatureInCelsius *float32         `json:"minStorageTemperatureInCelsius,omitempty"`
		LastUpdatedAt                  *time.Time       `json:"lastUpdatedAt"`
		ID                             string           `json:"id"`
		StorageType                    string           `json:"storageType"`
		StorageInstructions            string           `json:"storageInstructions"`
		Notes                          string           `json:"notes"`
		Source                         string           `json:"source"`
		Preparation                    ValidPreparation `json:"preparation"`
		Ingredient                     ValidIngredient  `json:"ingredient"`
		MinStorageDurationInSeconds    uint32           `json:"minStorageDurationInSeconds"`
	}

	// ValidPrepTaskConfigCreationRequestInput represents what a user could set as input for creating valid ingredient preparation storage configs.
	ValidPrepTaskConfigCreationRequestInput struct {
		_                              struct{} `json:"-"`
		MaxStorageDurationInSeconds    *uint32  `json:"maxStorageDurationInSeconds,omitempty"`
		MinStorageTemperatureInCelsius *float32 `json:"minStorageTemperatureInCelsius,omitempty"`
		MaxStorageTemperatureInCelsius *float32 `json:"maxStorageTemperatureInCelsius,omitempty"`
		StorageType                    string   `json:"storageType"`
		StorageInstructions            string   `json:"storageInstructions"`
		Notes                          string   `json:"notes"`
		Source                         string   `json:"source"`
		ValidPreparationID             string   `json:"validPreparationID"`
		ValidIngredientID              string   `json:"validIngredientID"`
		MinStorageDurationInSeconds    uint32   `json:"minStorageDurationInSeconds"`
	}

	// ValidPrepTaskConfigDatabaseCreationInput represents what a user could set as input for creating valid ingredient preparation storage configs.
	ValidPrepTaskConfigDatabaseCreationInput struct {
		_                              struct{} `json:"-"`
		MaxStorageDurationInSeconds    *uint32  `json:"-"`
		MinStorageTemperatureInCelsius *float32 `json:"-"`
		MaxStorageTemperatureInCelsius *float32 `json:"-"`
		ID                             string   `json:"-"`
		StorageType                    string   `json:"-"`
		StorageInstructions            string   `json:"-"`
		Notes                          string   `json:"-"`
		Source                         string   `json:"-"`
		ValidPreparationID             string   `json:"-"`
		ValidIngredientID              string   `json:"-"`
		MinStorageDurationInSeconds    uint32   `json:"-"`
	}

	// ValidPrepTaskConfigUpdateRequestInput represents what a user could set as input for updating valid ingredient preparation storage configs.
	ValidPrepTaskConfigUpdateRequestInput struct {
		_ struct{} `json:"-"`

		MinStorageDurationInSeconds    *uint32  `json:"minStorageDurationInSeconds,omitempty"`
		MaxStorageDurationInSeconds    *uint32  `json:"maxStorageDurationInSeconds,omitempty"`
		MinStorageTemperatureInCelsius *float32 `json:"minStorageTemperatureInCelsius,omitempty"`
		MaxStorageTemperatureInCelsius *float32 `json:"maxStorageTemperatureInCelsius,omitempty"`
		StorageType                    *string  `json:"storageType,omitempty"`
		StorageInstructions            *string  `json:"storageInstructions,omitempty"`
		Notes                          *string  `json:"notes,omitempty"`
		Source                         *string  `json:"source,omitempty"`
		ValidPreparationID             *string  `json:"validPreparationID,omitempty"`
		ValidIngredientID              *string  `json:"validIngredientID,omitempty"`
	}

	// ValidPrepTaskConfigDataManager describes a structure capable of storing valid ingredient preparation storage configs permanently.
	ValidPrepTaskConfigDataManager interface {
		ValidPrepTaskConfigExists(ctx context.Context, validIngredientPreparationStorageConfigID string) (bool, error)
		GetValidPrepTaskConfig(ctx context.Context, validIngredientPreparationStorageConfigID string) (*ValidPrepTaskConfig, error)
		GetValidPrepTaskConfigs(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidPrepTaskConfig], error)
		GetValidPrepTaskConfigsForIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidPrepTaskConfig], error)
		GetValidPrepTaskConfigsForPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidPrepTaskConfig], error)
		GetValidPrepTaskConfigsForIngredientAndPreparation(ctx context.Context, ingredientID, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidPrepTaskConfig], error)
		CreateValidPrepTaskConfig(ctx context.Context, input *ValidPrepTaskConfigDatabaseCreationInput) (*ValidPrepTaskConfig, error)
		UpdateValidPrepTaskConfig(ctx context.Context, updated *ValidPrepTaskConfig) error
		ArchiveValidPrepTaskConfig(ctx context.Context, validIngredientPreparationStorageConfigID string) error
	}
)

// Update merges a ValidPrepTaskConfigUpdateRequestInput with a valid ingredient preparation storage config.
func (x *ValidPrepTaskConfig) Update(input *ValidPrepTaskConfigUpdateRequestInput) {
	if input.StorageType != nil && *input.StorageType != x.StorageType {
		x.StorageType = *input.StorageType
	}

	if input.StorageInstructions != nil && *input.StorageInstructions != x.StorageInstructions {
		x.StorageInstructions = *input.StorageInstructions
	}

	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}

	if input.Source != nil && *input.Source != x.Source {
		x.Source = *input.Source
	}

	if input.ValidPreparationID != nil && *input.ValidPreparationID != x.Preparation.ID {
		x.Preparation.ID = *input.ValidPreparationID
	}

	if input.ValidIngredientID != nil && *input.ValidIngredientID != x.Ingredient.ID {
		x.Ingredient.ID = *input.ValidIngredientID
	}

	if input.MinStorageDurationInSeconds != nil && *input.MinStorageDurationInSeconds != x.MinStorageDurationInSeconds {
		x.MinStorageDurationInSeconds = *input.MinStorageDurationInSeconds
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
}

var _ validation.ValidatableWithContext = (*ValidPrepTaskConfigCreationRequestInput)(nil)

// ValidateWithContext validates a ValidPrepTaskConfigCreationRequestInput.
func (x *ValidPrepTaskConfigCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ValidPreparationID, validation.Required),
		validation.Field(&x.ValidIngredientID, validation.Required),
		validation.Field(&x.StorageType, validation.Required),
		validation.Field(&x.MinStorageDurationInSeconds, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidPrepTaskConfigDatabaseCreationInput)(nil)

// ValidateWithContext validates a ValidPrepTaskConfigDatabaseCreationInput.
func (x *ValidPrepTaskConfigDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.ValidPreparationID, validation.Required),
		validation.Field(&x.ValidIngredientID, validation.Required),
		validation.Field(&x.StorageType, validation.Required),
		validation.Field(&x.MinStorageDurationInSeconds, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidPrepTaskConfigUpdateRequestInput)(nil)

// ValidateWithContext validates a ValidPrepTaskConfigUpdateRequestInput.
func (x *ValidPrepTaskConfigUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ValidPreparationID, validation.Required),
		validation.Field(&x.ValidIngredientID, validation.Required),
	)
}
