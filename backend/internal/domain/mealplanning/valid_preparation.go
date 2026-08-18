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
	// ValidPreparationCreatedServiceEventType indicates a valid preparation was created.
	ValidPreparationCreatedServiceEventType = "valid_preparation_created"
	// ValidPreparationUpdatedServiceEventType indicates a valid preparation was updated.
	ValidPreparationUpdatedServiceEventType = "valid_preparation_updated"
	// ValidPreparationArchivedServiceEventType indicates a valid preparation was archived.
	ValidPreparationArchivedServiceEventType = "valid_preparation_archived"
)

func init() {
	gob.Register(new(ValidPreparation))
	gob.Register(new(ValidPreparationCreationRequestInput))
	gob.Register(new(ValidPreparationUpdateRequestInput))
}

type (
	// ValidPreparation represents a valid preparation.
	ValidPreparation struct {
		_                           struct{}                       `json:"-"`
		CreatedAt                   time.Time                      `json:"createdAt"`
		ArchivedAt                  *time.Time                     `json:"archivedAt"`
		LastUpdatedAt               *time.Time                     `json:"lastUpdatedAt"`
		MaxInstrumentCount          *uint16                        `json:"maxInstrumentCount,omitempty"`
		MaxIngredientCount          *uint16                        `json:"maxIngredientCount,omitempty"`
		MaxVesselCount              *uint16                        `json:"maxVesselCount,omitempty"`
		Name                        string                         `json:"name"`
		PastTense                   string                         `json:"pastTense"`
		ID                          string                         `json:"id"`
		IconPath                    string                         `json:"iconPath"`
		Description                 string                         `json:"description"`
		Slug                        string                         `json:"slug"`
		Media                       []*uploadedmedia.UploadedMedia `json:"media"`
		MinInstrumentCount          uint16                         `json:"minInstrumentCount"`
		MinIngredientCount          uint16                         `json:"minIngredientCount"`
		MinVesselCount              uint16                         `json:"minVesselCount"`
		TemperatureRequired         bool                           `json:"temperatureRequired"`
		TimeEstimateRequired        bool                           `json:"timeEstimateRequired"`
		ConditionExpressionRequired bool                           `json:"conditionExpressionRequired"`
		ConsumesVessel              bool                           `json:"consumesVessel"`
		OnlyForVessels              bool                           `json:"onlyForVessels"`
		YieldsNothing               bool                           `json:"yieldsNothing"`
		RestrictToIngredients       bool                           `json:"restrictToIngredients"`
	}

	// ValidPreparationCreationRequestInput represents what a user could set as input for creating valid preparations.
	ValidPreparationCreationRequestInput struct {
		_ struct{} `json:"-"`

		MaxInstrumentCount          *uint16 `json:"maxInstrumentCount,omitempty"`
		MaxIngredientCount          *uint16 `json:"maxIngredientCount,omitempty"`
		MaxVesselCount              *uint16 `json:"maxVesselCount,omitempty"`
		IconPath                    string  `json:"iconPath"`
		PastTense                   string  `json:"pastTense"`
		Slug                        string  `json:"slug"`
		Name                        string  `json:"name"`
		Description                 string  `json:"description"`
		MinInstrumentCount          uint16  `json:"minInstrumentCount"`
		MinIngredientCount          uint16  `json:"minIngredientCount"`
		MinVesselCount              uint16  `json:"minVesselCount"`
		TemperatureRequired         bool    `json:"temperatureRequired"`
		TimeEstimateRequired        bool    `json:"timeEstimateRequired"`
		ConditionExpressionRequired bool    `json:"conditionExpressionRequired"`
		ConsumesVessel              bool    `json:"consumesVessel"`
		OnlyForVessels              bool    `json:"onlyForVessels"`
		RestrictToIngredients       bool    `json:"restrictToIngredients"`
		YieldsNothing               bool    `json:"yieldsNothing"`
	}

	// ValidPreparationDatabaseCreationInput represents what a user could set as input for creating valid preparations.
	ValidPreparationDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		MaxInstrumentCount          *uint16 `json:"-"`
		MaxIngredientCount          *uint16 `json:"-"`
		MaxVesselCount              *uint16 `json:"-"`
		IconPath                    string  `json:"-"`
		PastTense                   string  `json:"-"`
		Slug                        string  `json:"-"`
		ID                          string  `json:"-"`
		Name                        string  `json:"-"`
		Description                 string  `json:"-"`
		MinInstrumentCount          uint16  `json:"-"`
		MinIngredientCount          uint16  `json:"-"`
		MinVesselCount              uint16  `json:"-"`
		TemperatureRequired         bool    `json:"-"`
		TimeEstimateRequired        bool    `json:"-"`
		ConditionExpressionRequired bool    `json:"-"`
		ConsumesVessel              bool    `json:"-"`
		OnlyForVessels              bool    `json:"-"`
		RestrictToIngredients       bool    `json:"-"`
		YieldsNothing               bool    `json:"-"`
	}

	// ValidPreparationUpdateRequestInput represents what a user could set as input for updating valid preparations.
	ValidPreparationUpdateRequestInput struct {
		_ struct{} `json:"-"`

		MinInstrumentCount          *uint16 `json:"minInstrumentCount,omitempty"`
		MaxInstrumentCount          *uint16 `json:"maxInstrumentCount,omitempty"`
		MinIngredientCount          *uint16 `json:"minIngredientCount,omitempty"`
		MaxIngredientCount          *uint16 `json:"maxIngredientCount,omitempty"`
		MinVesselCount              *uint16 `json:"minVesselCount,omitempty"`
		MaxVesselCount              *uint16 `json:"maxVesselCount,omitempty"`
		Name                        *string `json:"name,omitempty"`
		Description                 *string `json:"description,omitempty"`
		IconPath                    *string `json:"iconPath,omitempty"`
		YieldsNothing               *bool   `json:"yieldsNothing,omitempty"`
		Slug                        *string `json:"slug,omitempty"`
		RestrictToIngredients       *bool   `json:"restrictToIngredients,omitempty"`
		PastTense                   *string `json:"pastTense,omitempty"`
		TemperatureRequired         *bool   `json:"temperatureRequired,omitempty"`
		TimeEstimateRequired        *bool   `json:"timeEstimateRequired,omitempty"`
		ConditionExpressionRequired *bool   `json:"conditionExpressionRequired,omitempty"`
		ConsumesVessel              *bool   `json:"consumesVessel,omitempty"`
		OnlyForVessels              *bool   `json:"onlyForVessels,omitempty"`
	}

	// ValidPreparationDataManager describes a structure capable of storing valid preparations permanently.
	ValidPreparationDataManager interface {
		ValidPreparationExists(ctx context.Context, validPreparationID string) (bool, error)
		GetValidPreparation(ctx context.Context, validPreparationID string) (*ValidPreparation, error)
		GetRandomValidPreparation(ctx context.Context) (*ValidPreparation, error)
		GetValidPreparations(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidPreparation], error)
		SearchForValidPreparations(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidPreparation], error)
		CreateValidPreparation(ctx context.Context, input *ValidPreparationDatabaseCreationInput) (*ValidPreparation, error)
		UpdateValidPreparation(ctx context.Context, updated *ValidPreparation) error
		MarkValidPreparationAsIndexed(ctx context.Context, validPreparationID string) error
		ArchiveValidPreparation(ctx context.Context, validPreparationID string) error
		ScanValidPreparationIDsForReindex(ctx context.Context, after string, limit int) ([]string, error)
		GetValidPreparationsWithIDs(ctx context.Context, ids []string) ([]*ValidPreparation, error)
	}
)

// Update merges a ValidPreparationUpdateRequestInput with a valid preparation.
func (x *ValidPreparation) Update(input *ValidPreparationUpdateRequestInput) {
	if input.Name != nil && *input.Name != x.Name {
		x.Name = *input.Name
	}

	if input.Description != nil && *input.Description != x.Description {
		x.Description = *input.Description
	}

	if input.IconPath != nil && *input.IconPath != x.IconPath {
		x.IconPath = *input.IconPath
	}

	if input.YieldsNothing != nil && *input.YieldsNothing != x.YieldsNothing {
		x.YieldsNothing = *input.YieldsNothing
	}

	if input.RestrictToIngredients != nil && *input.RestrictToIngredients != x.RestrictToIngredients {
		x.RestrictToIngredients = *input.RestrictToIngredients
	}

	if input.MinIngredientCount != nil && *input.MinIngredientCount != x.MinIngredientCount {
		x.MinIngredientCount = *input.MinIngredientCount
	}

	if input.MaxIngredientCount != nil && x.MaxIngredientCount != nil && *input.MaxIngredientCount != *x.MaxIngredientCount {
		x.MaxIngredientCount = input.MaxIngredientCount
	}

	if input.MinInstrumentCount != nil && *input.MinInstrumentCount != x.MinInstrumentCount {
		x.MinInstrumentCount = *input.MinInstrumentCount
	}

	if input.MaxInstrumentCount != nil && x.MaxInstrumentCount != nil && *input.MaxInstrumentCount != *x.MaxInstrumentCount {
		x.MaxInstrumentCount = input.MaxInstrumentCount
	}

	if input.TemperatureRequired != nil && *input.TemperatureRequired != x.TemperatureRequired {
		x.TemperatureRequired = *input.TemperatureRequired
	}

	if input.TimeEstimateRequired != nil && *input.TimeEstimateRequired != x.TimeEstimateRequired {
		x.TimeEstimateRequired = *input.TimeEstimateRequired
	}

	if input.ConditionExpressionRequired != nil && *input.ConditionExpressionRequired != x.ConditionExpressionRequired {
		x.ConditionExpressionRequired = *input.ConditionExpressionRequired
	}

	if input.PastTense != nil && *input.PastTense != x.PastTense {
		x.PastTense = *input.PastTense
	}

	if input.Slug != nil && *input.Slug != x.Slug {
		x.Slug = *input.Slug
	}

	if input.ConsumesVessel != nil && *input.ConsumesVessel != x.ConsumesVessel {
		x.ConsumesVessel = *input.ConsumesVessel
	}

	if input.OnlyForVessels != nil && *input.OnlyForVessels != x.OnlyForVessels {
		x.OnlyForVessels = *input.OnlyForVessels
	}

	if input.MinVesselCount != nil && *input.MinVesselCount != x.MinVesselCount {
		x.MinVesselCount = *input.MinVesselCount
	}

	if input.MaxVesselCount != nil && x.MaxVesselCount != nil && *input.MaxVesselCount != *x.MaxVesselCount {
		x.MaxVesselCount = input.MaxVesselCount
	}
}

var _ validation.ValidatableWithContext = (*ValidPreparationCreationRequestInput)(nil)

// ValidateWithContext validates a ValidPreparationCreationRequestInput.
func (x *ValidPreparationCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidPreparationDatabaseCreationInput)(nil)

// ValidateWithContext validates a ValidPreparationDatabaseCreationInput.
func (x *ValidPreparationDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Name, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidPreparationUpdateRequestInput)(nil)

// ValidateWithContext validates a ValidPreparationUpdateRequestInput.
func (x *ValidPreparationUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}
