package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	"github.com/primandproper/platform-go/v11/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ValidIngredientCreatedServiceEventType indicates a valid ingredient was created.
	ValidIngredientCreatedServiceEventType = "valid_ingredient_created"
	// ValidIngredientUpdatedServiceEventType indicates a valid ingredient was updated.
	ValidIngredientUpdatedServiceEventType = "valid_ingredient_updated"
	// ValidIngredientArchivedServiceEventType indicates a valid ingredient was archived.
	ValidIngredientArchivedServiceEventType = "valid_ingredient_archived"
)

func init() {
	gob.Register(new(ValidIngredient))
	gob.Register(new(ValidIngredientCreationRequestInput))
	gob.Register(new(ValidIngredientUpdateRequestInput))
}

type (
	// ValidIngredient represents a valid ingredient.
	ValidIngredient struct {
		_                              struct{}                       `json:"-"`
		CreatedAt                      time.Time                      `json:"createdAt"`
		MinStorageTemperatureInCelsius *float32                       `json:"minStorageTemperatureInCelsius,omitempty"`
		MaxStorageTemperatureInCelsius *float32                       `json:"maxStorageTemperatureInCelsius,omitempty"`
		LastUpdatedAt                  *time.Time                     `json:"lastUpdatedAt"`
		ArchivedAt                     *time.Time                     `json:"archivedAt"`
		Warning                        string                         `json:"warning"`
		IconPath                       string                         `json:"iconPath"`
		PluralName                     string                         `json:"pluralName"`
		StorageInstructions            string                         `json:"storageInstructions"`
		Name                           string                         `json:"name"`
		ID                             string                         `json:"id"`
		Description                    string                         `json:"description"`
		Slug                           string                         `json:"slug"`
		ShoppingSuggestions            string                         `json:"shoppingSuggestions"`
		Media                          []*uploadedmedia.UploadedMedia `json:"media"`
		ContainsEgg                    bool                           `json:"containsEgg"`
		ContainsGluten                 bool                           `json:"containsGluten"`
		ContainsTreeNut                bool                           `json:"containsTreeNut"`
		IsLiquid                       bool                           `json:"isLiquid"`
		ContainsWheat                  bool                           `json:"containsWheat"`
		ContainsSoy                    bool                           `json:"containsSoy"`
		AnimalDerived                  bool                           `json:"animalDerived"`
		RestrictToPreparations         bool                           `json:"restrictToPreparations"`
		ContaminatesEquipment          bool                           `json:"contaminatesEquipment"`
		ContainsSesame                 bool                           `json:"containsSesame"`
		ContainsFish                   bool                           `json:"containsFish"`
		ContainsPeanut                 bool                           `json:"containsPeanut"`
		ContainsDairy                  bool                           `json:"containsDairy"`
		ContainsAlcohol                bool                           `json:"containsAlcohol"`
		AnimalFlesh                    bool                           `json:"animalFlesh"`
		IsStarch                       bool                           `json:"isStarch"`
		IsProtein                      bool                           `json:"isProtein"`
		IsGrain                        bool                           `json:"isGrain"`
		IsFruit                        bool                           `json:"isFruit"`
		IsSalt                         bool                           `json:"isSalt"`
		IsFat                          bool                           `json:"isFat"`
		IsAcid                         bool                           `json:"isAcid"`
		IsHeat                         bool                           `json:"isHeat"`
		ContainsShellfish              bool                           `json:"containsShellfish"`
	}

	// NullableValidIngredient represents a nullable valid ingredient.
	NullableValidIngredient struct {
		_ struct{} `json:"-"`

		CreatedAt                      *time.Time
		LastUpdatedAt                  *time.Time
		ArchivedAt                     *time.Time
		ID                             *string
		Warning                        *string
		Description                    *string
		IconPath                       *string
		PluralName                     *string
		StorageInstructions            *string
		Name                           *string
		MinStorageTemperatureInCelsius *float32
		MaxStorageTemperatureInCelsius *float32
		ContainsShellfish              *bool
		ContainsDairy                  *bool
		AnimalFlesh                    *bool
		IsLiquid                       *bool
		ContainsPeanut                 *bool
		ContainsTreeNut                *bool
		ContainsEgg                    *bool
		ContainsWheat                  *bool
		ContainsSoy                    *bool
		AnimalDerived                  *bool
		RestrictToPreparations         *bool
		ContaminatesEquipment          *bool
		ContainsSesame                 *bool
		ContainsFish                   *bool
		ContainsGluten                 *bool
		Slug                           *string
		ContainsAlcohol                *bool
		ShoppingSuggestions            *string
		IsStarch                       *bool
		IsProtein                      *bool
		IsGrain                        *bool
		IsFruit                        *bool
		IsSalt                         *bool
		IsFat                          *bool
		IsAcid                         *bool
		IsHeat                         *bool
	}

	// ValidIngredientCreationRequestInput represents what a user could set as input for creating valid ingredients.
	ValidIngredientCreationRequestInput struct {
		_ struct{} `json:"-"`

		MinStorageTemperatureInCelsius *float32 `json:"minStorageTemperatureInCelsius,omitempty"`
		MaxStorageTemperatureInCelsius *float32 `json:"maxStorageTemperatureInCelsius,omitempty"`
		Warning                        string   `json:"warning"`
		IconPath                       string   `json:"iconPath"`
		PluralName                     string   `json:"pluralName"`
		StorageInstructions            string   `json:"storageInstructions"`
		Name                           string   `json:"name"`
		Description                    string   `json:"description"`
		Slug                           string   `json:"slug"`
		ShoppingSuggestions            string   `json:"shoppingSuggestions"`
		ContainsFish                   bool     `json:"containsFish"`
		ContainsShellfish              bool     `json:"containsShellfish"`
		AnimalFlesh                    bool     `json:"animalFlesh"`
		ContainsEgg                    bool     `json:"containsEgg"`
		IsLiquid                       bool     `json:"isLiquid"`
		ContainsSoy                    bool     `json:"containsSoy"`
		ContainsPeanut                 bool     `json:"containsPeanut"`
		AnimalDerived                  bool     `json:"animalDerived"`
		RestrictToPreparations         bool     `json:"restrictToPreparations"`
		ContaminatesEquipment          bool     `json:"contaminatesEquipment"`
		ContainsDairy                  bool     `json:"containsDairy"`
		ContainsSesame                 bool     `json:"containsSesame"`
		ContainsTreeNut                bool     `json:"containsTreeNut"`
		ContainsWheat                  bool     `json:"containsWheat"`
		ContainsAlcohol                bool     `json:"containsAlcohol"`
		ContainsGluten                 bool     `json:"containsGluten"`
		IsStarch                       bool     `json:"isStarch"`
		IsProtein                      bool     `json:"isProtein"`
		IsGrain                        bool     `json:"isGrain"`
		IsFruit                        bool     `json:"isFruit"`
		IsSalt                         bool     `json:"isSalt"`
		IsFat                          bool     `json:"isFat"`
		IsAcid                         bool     `json:"isAcid"`
		IsHeat                         bool     `json:"isHeat"`
	}

	// ValidIngredientDatabaseCreationInput represents what a user could set as input for creating valid ingredients.
	ValidIngredientDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		MinStorageTemperatureInCelsius *float32 `json:"-"`
		MaxStorageTemperatureInCelsius *float32 `json:"-"`
		ID                             string   `json:"-"`
		Warning                        string   `json:"-"`
		IconPath                       string   `json:"-"`
		PluralName                     string   `json:"-"`
		StorageInstructions            string   `json:"-"`
		Name                           string   `json:"-"`
		Description                    string   `json:"-"`
		Slug                           string   `json:"-"`
		ShoppingSuggestions            string   `json:"-"`
		ContainsFish                   bool     `json:"-"`
		ContainsShellfish              bool     `json:"-"`
		AnimalFlesh                    bool     `json:"-"`
		ContainsEgg                    bool     `json:"-"`
		IsLiquid                       bool     `json:"-"`
		ContainsSoy                    bool     `json:"-"`
		ContainsPeanut                 bool     `json:"-"`
		AnimalDerived                  bool     `json:"-"`
		RestrictToPreparations         bool     `json:"-"`
		ContaminatesEquipment          bool     `json:"-"`
		ContainsDairy                  bool     `json:"-"`
		ContainsSesame                 bool     `json:"-"`
		ContainsTreeNut                bool     `json:"-"`
		ContainsWheat                  bool     `json:"-"`
		ContainsAlcohol                bool     `json:"-"`
		ContainsGluten                 bool     `json:"-"`
		IsStarch                       bool     `json:"-"`
		IsProtein                      bool     `json:"-"`
		IsGrain                        bool     `json:"-"`
		IsFruit                        bool     `json:"-"`
		IsSalt                         bool     `json:"-"`
		IsFat                          bool     `json:"-"`
		IsAcid                         bool     `json:"-"`
		IsHeat                         bool     `json:"-"`
	}

	// ValidIngredientUpdateRequestInput represents what a user could set as input for updating valid ingredients.
	ValidIngredientUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Name                           *string  `json:"name,omitempty"`
		Description                    *string  `json:"description,omitempty"`
		Warning                        *string  `json:"warning,omitempty"`
		IconPath                       *string  `json:"iconPath,omitempty"`
		ContainsDairy                  *bool    `json:"containsDairy,omitempty"`
		ContainsPeanut                 *bool    `json:"containsPeanut,omitempty"`
		ContainsTreeNut                *bool    `json:"containsTreeNut,omitempty"`
		ContainsEgg                    *bool    `json:"containsEgg,omitempty"`
		ContainsWheat                  *bool    `json:"containsWheat,omitempty"`
		ContainsShellfish              *bool    `json:"containsShellfish,omitempty"`
		ContainsSesame                 *bool    `json:"containsSesame,omitempty"`
		ContainsFish                   *bool    `json:"containsFish,omitempty"`
		ContainsGluten                 *bool    `json:"containsGluten,omitempty"`
		AnimalFlesh                    *bool    `json:"animalFlesh,omitempty"`
		IsLiquid                       *bool    `json:"isLiquid,omitempty"`
		ContainsSoy                    *bool    `json:"containsSoy,omitempty"`
		PluralName                     *string  `json:"pluralName,omitempty"`
		AnimalDerived                  *bool    `json:"animalDerived,omitempty"`
		RestrictToPreparations         *bool    `json:"restrictToPreparations,omitempty"`
		ContaminatesEquipment          *bool    `json:"contaminatesEquipment,omitempty"`
		MinStorageTemperatureInCelsius *float32 `json:"minStorageTemperatureInCelsius,omitempty"`
		MaxStorageTemperatureInCelsius *float32 `json:"maxStorageTemperatureInCelsius,omitempty"`
		StorageInstructions            *string  `json:"storageInstructions,omitempty"`
		Slug                           *string  `json:"slug,omitempty"`
		ContainsAlcohol                *bool    `json:"containsAlcohol,omitempty"`
		ShoppingSuggestions            *string  `json:"shoppingSuggestions,omitempty"`
		IsStarch                       *bool    `json:"isStarch"`
		IsProtein                      *bool    `json:"isProtein"`
		IsGrain                        *bool    `json:"isGrain"`
		IsFruit                        *bool    `json:"isFruit"`
		IsSalt                         *bool    `json:"isSalt"`
		IsFat                          *bool    `json:"isFat"`
		IsAcid                         *bool    `json:"isAcid"`
		IsHeat                         *bool    `json:"isHeat"`
	}

	// ValidIngredientDataManager describes a structure capable of storing valid ingredients permanently.
	ValidIngredientDataManager interface {
		ValidIngredientExists(ctx context.Context, validIngredientID string) (bool, error)
		GetValidIngredient(ctx context.Context, validIngredientID string) (*ValidIngredient, error)
		GetRandomValidIngredient(ctx context.Context) (*ValidIngredient, error)
		GetValidIngredients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidIngredient], error)
		SearchForValidIngredients(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidIngredient], error)
		SearchForValidIngredientsForPreparation(ctx context.Context, preparationID, query string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[ValidIngredient], err error)
		CreateValidIngredient(ctx context.Context, input *ValidIngredientDatabaseCreationInput) (*ValidIngredient, error)
		UpdateValidIngredient(ctx context.Context, updated *ValidIngredient) error
		MarkValidIngredientAsIndexed(ctx context.Context, validIngredientID string) error
		ArchiveValidIngredient(ctx context.Context, validIngredientID string) error
		ScanValidIngredientIDsForReindex(ctx context.Context, after string, limit int) ([]string, error)
		GetValidIngredientsWithIDs(ctx context.Context, ids []string) ([]*ValidIngredient, error)
	}
)

// Update merges a ValidIngredientUpdateRequestInput with a valid ingredient.
func (x *ValidIngredient) Update(input *ValidIngredientUpdateRequestInput) {
	if input.Name != nil && *input.Name != x.Name {
		x.Name = *input.Name
	}

	if input.Description != nil && *input.Description != x.Description {
		x.Description = *input.Description
	}

	if input.Warning != nil && *input.Warning != x.Warning {
		x.Warning = *input.Warning
	}

	if input.ContainsEgg != nil && *input.ContainsEgg != x.ContainsEgg {
		x.ContainsEgg = *input.ContainsEgg
	}

	if input.ContainsDairy != nil && *input.ContainsDairy != x.ContainsDairy {
		x.ContainsDairy = *input.ContainsDairy
	}

	if input.ContainsPeanut != nil && *input.ContainsPeanut != x.ContainsPeanut {
		x.ContainsPeanut = *input.ContainsPeanut
	}

	if input.ContainsTreeNut != nil && *input.ContainsTreeNut != x.ContainsTreeNut {
		x.ContainsTreeNut = *input.ContainsTreeNut
	}

	if input.ContainsSoy != nil && *input.ContainsSoy != x.ContainsSoy {
		x.ContainsSoy = *input.ContainsSoy
	}

	if input.ContainsWheat != nil && *input.ContainsWheat != x.ContainsWheat {
		x.ContainsWheat = *input.ContainsWheat
	}

	if input.ContainsShellfish != nil && *input.ContainsShellfish != x.ContainsShellfish {
		x.ContainsShellfish = *input.ContainsShellfish
	}

	if input.ContainsSesame != nil && *input.ContainsSesame != x.ContainsSesame {
		x.ContainsSesame = *input.ContainsSesame
	}

	if input.ContainsFish != nil && *input.ContainsFish != x.ContainsFish {
		x.ContainsFish = *input.ContainsFish
	}

	if input.ContainsGluten != nil && *input.ContainsGluten != x.ContainsGluten {
		x.ContainsGluten = *input.ContainsGluten
	}

	if input.AnimalFlesh != nil && *input.AnimalFlesh != x.AnimalFlesh {
		x.AnimalFlesh = *input.AnimalFlesh
	}

	if input.IsLiquid != nil && *input.IsLiquid != x.IsLiquid {
		x.IsLiquid = *input.IsLiquid
	}

	if input.IconPath != nil && *input.IconPath != x.IconPath {
		x.IconPath = *input.IconPath
	}

	if input.PluralName != nil && *input.PluralName != x.PluralName {
		x.PluralName = *input.PluralName
	}

	if input.AnimalDerived != nil && *input.AnimalDerived != x.AnimalDerived {
		x.AnimalDerived = *input.AnimalDerived
	}

	if input.RestrictToPreparations != nil && *input.RestrictToPreparations != x.RestrictToPreparations {
		x.RestrictToPreparations = *input.RestrictToPreparations
	}

	if input.ContaminatesEquipment != nil && *input.ContaminatesEquipment != x.ContaminatesEquipment {
		x.ContaminatesEquipment = *input.ContaminatesEquipment
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

	if input.Slug != nil && *input.Slug != x.Slug {
		x.Slug = *input.Slug
	}

	if input.ContainsAlcohol != nil && *input.ContainsAlcohol != x.ContainsAlcohol {
		x.ContainsAlcohol = *input.ContainsAlcohol
	}

	if input.ShoppingSuggestions != nil && *input.ShoppingSuggestions != x.ShoppingSuggestions {
		x.ShoppingSuggestions = *input.ShoppingSuggestions
	}

	if input.IsStarch != nil && *input.IsStarch != x.IsStarch {
		x.IsStarch = *input.IsStarch
	}

	if input.IsProtein != nil && *input.IsProtein != x.IsProtein {
		x.IsProtein = *input.IsProtein
	}

	if input.IsGrain != nil && *input.IsGrain != x.IsGrain {
		x.IsGrain = *input.IsGrain
	}

	if input.IsFruit != nil && *input.IsFruit != x.IsFruit {
		x.IsFruit = *input.IsFruit
	}

	if input.IsSalt != nil && *input.IsSalt != x.IsSalt {
		x.IsSalt = *input.IsSalt
	}

	if input.IsFat != nil && *input.IsFat != x.IsFat {
		x.IsFat = *input.IsFat
	}

	if input.IsAcid != nil && *input.IsAcid != x.IsAcid {
		x.IsAcid = *input.IsAcid
	}

	if input.IsHeat != nil && *input.IsHeat != x.IsHeat {
		x.IsHeat = *input.IsHeat
	}
}

var _ validation.ValidatableWithContext = (*ValidIngredientCreationRequestInput)(nil)

// ValidateWithContext validates a ValidIngredientCreationRequestInput.
func (x *ValidIngredientCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidIngredientDatabaseCreationInput)(nil)

// ValidateWithContext validates a ValidIngredientDatabaseCreationInput.
func (x *ValidIngredientDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Name, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidIngredientUpdateRequestInput)(nil)

// ValidateWithContext validates a ValidIngredientUpdateRequestInput.
func (x *ValidIngredientUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
	)
}
