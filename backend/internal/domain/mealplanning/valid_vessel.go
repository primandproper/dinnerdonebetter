package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v11/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ValidVesselCreatedServiceEventType indicates a valid vessel was created.
	ValidVesselCreatedServiceEventType = "valid_vessel_created"
	// ValidVesselUpdatedServiceEventType indicates a valid vessel was updated.
	ValidVesselUpdatedServiceEventType = "valid_vessel_updated"
	// ValidVesselArchivedServiceEventType indicates a valid vessel was archived.
	ValidVesselArchivedServiceEventType = "valid_vessel_archived"

	// VesselShapeHemisphere represents a hemisphere-shaped vessel.
	VesselShapeHemisphere = "hemisphere"
	// VesselShapeRectangle represents a rectangular-shaped vessel.
	VesselShapeRectangle = "rectangle"
	// VesselShapeCone represents a cone-shaped vessel.
	VesselShapeCone = "cone"
	// VesselShapePyramid represents a pyramid-shaped vessel.
	VesselShapePyramid = "pyramid"
	// VesselShapeCylinder represents a cylindrical vessel.
	VesselShapeCylinder = "cylinder"
	// VesselShapeSphere represents a spherical vessel.
	VesselShapeSphere = "sphere"
	// VesselShapeCube represents a cubic vessel.
	VesselShapeCube = "cube"
	// VesselShapeOther represents any other vessel shape.
	VesselShapeOther = "other"
)

func init() {
	gob.Register(new(ValidVessel))
	gob.Register(new(ValidVesselCreationRequestInput))
	gob.Register(new(ValidVesselUpdateRequestInput))
}

type (
	// ValidVessel represents a valid vessel.
	ValidVessel struct {
		_ struct{} `json:"-"`

		CreatedAt                      time.Time             `json:"createdAt"`
		ArchivedAt                     *time.Time            `json:"archivedAt"`
		LastUpdatedAt                  *time.Time            `json:"lastUpdatedAt"`
		CapacityUnit                   *ValidMeasurementUnit `json:"capacityUnit"`
		IconPath                       string                `json:"iconPath"`
		PluralName                     string                `json:"pluralName"`
		Description                    string                `json:"description"`
		Name                           string                `json:"name"`
		Slug                           string                `json:"slug"`
		Shape                          string                `json:"shape"`
		ID                             string                `json:"id"`
		WidthInMillimeters             float32               `json:"widthInMillimeters"`
		LengthInMillimeters            float32               `json:"lengthInMillimeters"`
		HeightInMillimeters            float32               `json:"heightInMillimeters"`
		Capacity                       float32               `json:"capacity"`
		IncludeInGeneratedInstructions bool                  `json:"includeInGeneratedInstructions"`
		DisplayInSummaryLists          bool                  `json:"displayInSummaryLists"`
		UsableForStorage               bool                  `json:"usableForStorage"`
	}

	// NullableValidVessel represents a fully nullable valid vessel.
	NullableValidVessel struct {
		_ struct{} `json:"-"`

		ID                             *string                       `json:"id"`
		Name                           *string                       `json:"name"`
		PluralName                     *string                       `json:"pluralName"`
		Description                    *string                       `json:"description"`
		IconPath                       *string                       `json:"iconPath"`
		UsableForStorage               *bool                         `json:"usableForStorage"`
		Slug                           *string                       `json:"slug"`
		DisplayInSummaryLists          *bool                         `json:"displayInSummaryLists"`
		IncludeInGeneratedInstructions *bool                         `json:"includeInGeneratedInstructions"`
		Capacity                       *float32                      `json:"capacity"`
		CapacityUnit                   *NullableValidMeasurementUnit `json:"capacityUnit"`
		WidthInMillimeters             *float32                      `json:"widthInMillimeters"`
		LengthInMillimeters            *float32                      `json:"lengthInMillimeters"`
		HeightInMillimeters            *float32                      `json:"heightInMillimeters"`
		Shape                          *string                       `json:"shape"`
		CreatedAt                      *time.Time                    `json:"createdAt"`
		LastUpdatedAt                  *time.Time                    `json:"lastUpdatedAt"`
		ArchivedAt                     *time.Time                    `json:"archivedAt"`
	}

	// ValidVesselCreationRequestInput represents what a user could set as input for creating valid vessels.
	ValidVesselCreationRequestInput struct {
		_ struct{} `json:"-"`

		CapacityUnitID                 *string `json:"capacityUnitID"`
		Shape                          string  `json:"shape"`
		IconPath                       string  `json:"iconPath"`
		PluralName                     string  `json:"pluralName"`
		Name                           string  `json:"name"`
		Description                    string  `json:"description"`
		Slug                           string  `json:"slug"`
		LengthInMillimeters            float32 `json:"lengthInMillimeters"`
		HeightInMillimeters            float32 `json:"heightInMillimeters"`
		Capacity                       float32 `json:"capacity"`
		WidthInMillimeters             float32 `json:"widthInMillimeters"`
		UsableForStorage               bool    `json:"usableForStorage"`
		IncludeInGeneratedInstructions bool    `json:"includeInGeneratedInstructions"`
		DisplayInSummaryLists          bool    `json:"displayInSummaryLists"`
	}

	// ValidVesselDatabaseCreationInput represents what a user could set as input for creating valid vessels.
	ValidVesselDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		CapacityUnitID                 *string `json:"-"`
		ID                             string  `json:"-"`
		Name                           string  `json:"-"`
		PluralName                     string  `json:"-"`
		Description                    string  `json:"-"`
		IconPath                       string  `json:"-"`
		Shape                          string  `json:"-"`
		Slug                           string  `json:"-"`
		WidthInMillimeters             float32 `json:"-"`
		Capacity                       float32 `json:"-"`
		LengthInMillimeters            float32 `json:"-"`
		HeightInMillimeters            float32 `json:"-"`
		IncludeInGeneratedInstructions bool    `json:"-"`
		DisplayInSummaryLists          bool    `json:"-"`
		UsableForStorage               bool    `json:"-"`
	}

	// ValidVesselUpdateRequestInput represents what a user could set as input for updating valid vessels.
	ValidVesselUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Name                           *string  `json:"name"`
		PluralName                     *string  `json:"pluralName"`
		Description                    *string  `json:"description"`
		IconPath                       *string  `json:"iconPath"`
		UsableForStorage               *bool    `json:"usableForStorage"`
		Slug                           *string  `json:"slug"`
		DisplayInSummaryLists          *bool    `json:"displayInSummaryLists"`
		IncludeInGeneratedInstructions *bool    `json:"includeInGeneratedInstructions"`
		Capacity                       *float32 `json:"capacity"`
		CapacityUnitID                 *string  `json:"capacityUnitID"`
		WidthInMillimeters             *float32 `json:"widthInMillimeters"`
		LengthInMillimeters            *float32 `json:"lengthInMillimeters"`
		HeightInMillimeters            *float32 `json:"heightInMillimeters"`
		Shape                          *string  `json:"shape"`
	}

	// ValidVesselDataManager describes a structure capable of storing valid vessels permanently.
	ValidVesselDataManager interface {
		ValidVesselExists(ctx context.Context, validVesselID string) (bool, error)
		GetValidVessel(ctx context.Context, validVesselID string) (*ValidVessel, error)
		GetRandomValidVessel(ctx context.Context) (*ValidVessel, error)
		GetValidVessels(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidVessel], error)
		SearchForValidVessels(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidVessel], error)
		CreateValidVessel(ctx context.Context, input *ValidVesselDatabaseCreationInput) (*ValidVessel, error)
		UpdateValidVessel(ctx context.Context, updated *ValidVessel) error
		MarkValidVesselsAsIndexed(ctx context.Context, ids []string) error
		ArchiveValidVessel(ctx context.Context, validVesselID string) error
		ScanValidVesselIDsForReindex(ctx context.Context, after string, limit int) ([]string, error)
		GetValidVesselsWithIDs(ctx context.Context, ids []string) ([]*ValidVessel, error)
	}
)

// Update merges a ValidVesselUpdateRequestInput with a valid vessel.
func (x *ValidVessel) Update(input *ValidVesselUpdateRequestInput) {
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

	if input.Capacity != nil && *input.Capacity != x.Capacity {
		x.Capacity = *input.Capacity
	}

	if input.CapacityUnitID != nil && x.CapacityUnit != nil && *input.CapacityUnitID != x.CapacityUnit.ID {
		x.CapacityUnit = &ValidMeasurementUnit{ID: *input.CapacityUnitID}
	}

	if input.WidthInMillimeters != nil && *input.WidthInMillimeters != x.WidthInMillimeters {
		x.WidthInMillimeters = *input.WidthInMillimeters
	}

	if input.LengthInMillimeters != nil && *input.LengthInMillimeters != x.LengthInMillimeters {
		x.LengthInMillimeters = *input.LengthInMillimeters
	}

	if input.HeightInMillimeters != nil && *input.HeightInMillimeters != x.HeightInMillimeters {
		x.HeightInMillimeters = *input.HeightInMillimeters
	}

	if input.Shape != nil && *input.Shape != x.Shape {
		x.Shape = *input.Shape
	}
}

var _ validation.ValidatableWithContext = (*ValidVesselCreationRequestInput)(nil)

// ValidateWithContext validates a ValidVesselCreationRequestInput.
func (x *ValidVesselCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Capacity, validation.When(x.CapacityUnitID != nil, validation.Required)),
		validation.Field(&x.Shape, validation.In(
			VesselShapeHemisphere,
			VesselShapeRectangle,
			VesselShapeCone,
			VesselShapePyramid,
			VesselShapeCylinder,
			VesselShapeSphere,
			VesselShapeCube,
			VesselShapeOther,
		)),
	)
}

var _ validation.ValidatableWithContext = (*ValidVesselDatabaseCreationInput)(nil)

// ValidateWithContext validates a ValidVesselDatabaseCreationInput.
func (x *ValidVesselDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.CapacityUnitID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidVesselUpdateRequestInput)(nil)

// ValidateWithContext validates a ValidVesselUpdateRequestInput.
func (x *ValidVesselUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}
