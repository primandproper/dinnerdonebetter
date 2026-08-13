package mealplanning

import (
	"context"
	"encoding/gob"
	"errors"
	"time"

	"github.com/primandproper/platform-go/v10/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

const (
	// ValidMeasurementUnitCreatedServiceEventType indicates a valid measurement unit was created.
	ValidMeasurementUnitCreatedServiceEventType = "valid_measurement_unit_created"
	// ValidMeasurementUnitUpdatedServiceEventType indicates a valid measurement unit was updated.
	ValidMeasurementUnitUpdatedServiceEventType = "valid_measurement_unit_updated"
	// ValidMeasurementUnitArchivedServiceEventType indicates a valid measurement unit was archived.
	ValidMeasurementUnitArchivedServiceEventType = "valid_measurement_unit_archived"
)

var (
	errMustBeEitherMetricOrImperial = errors.New("cannot be both metric and imperial")
)

func init() {
	gob.Register(new(ValidMeasurementUnit))
	gob.Register(new(ValidMeasurementUnitCreationRequestInput))
	gob.Register(new(ValidMeasurementUnitUpdateRequestInput))
}

type (
	/*
		ValidMeasurementUnit represents a valid measurement unit.
		Field notes:
			- Universal indicates that the measurement unit is valid for all ingredients. For instance, "grams" is a universal measurement unit.
	*/
	ValidMeasurementUnit struct {
		_ struct{} `json:"-"`

		CreatedAt     time.Time  `json:"createdAt"`
		LastUpdatedAt *time.Time `json:"lastUpdatedAt"`
		ArchivedAt    *time.Time `json:"archivedAt"`
		Name          string     `json:"name"`
		IconPath      string     `json:"iconPath"`
		ID            string     `json:"id"`
		Description   string     `json:"description"`
		PluralName    string     `json:"pluralName"`
		Slug          string     `json:"slug"`
		Volumetric    bool       `json:"volumetric"`
		Universal     bool       `json:"universal"`
		Metric        bool       `json:"metric"`
		Imperial      bool       `json:"imperial"`
	}

	// NullableValidMeasurementUnit represents a nullable valid measurement unit.
	NullableValidMeasurementUnit struct {
		_ struct{} `json:"-"`

		CreatedAt     *time.Time
		LastUpdatedAt *time.Time
		ArchivedAt    *time.Time
		Name          *string
		IconPath      *string
		ID            *string
		Description   *string
		PluralName    *string
		Slug          *string
		Volumetric    *bool
		Universal     *bool
		Metric        *bool
		Imperial      *bool
	}

	// ValidMeasurementUnitCreationRequestInput represents what a user could set as input for creating valid measurement units.
	ValidMeasurementUnitCreationRequestInput struct {
		_ struct{} `json:"-"`

		Name        string `json:"name"`
		Description string `json:"description"`
		IconPath    string `json:"iconPath"`
		PluralName  string `json:"pluralName"`
		Slug        string `json:"slug"`
		Volumetric  bool   `json:"volumetric"`
		Universal   bool   `json:"universal"`
		Metric      bool   `json:"metric"`
		Imperial    bool   `json:"imperial"`
	}

	// ValidMeasurementUnitDatabaseCreationInput represents what a user could set as input for creating valid measurement units.
	ValidMeasurementUnitDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		Name        string `json:"-"`
		Description string `json:"-"`
		ID          string `json:"-"`
		IconPath    string `json:"-"`
		PluralName  string `json:"-"`
		Slug        string `json:"-"`
		Volumetric  bool   `json:"-"`
		Universal   bool   `json:"-"`
		Metric      bool   `json:"-"`
		Imperial    bool   `json:"-"`
	}

	// ValidMeasurementUnitUpdateRequestInput represents what a user could set as input for updating valid measurement units.
	ValidMeasurementUnitUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Name        *string `json:"name,omitempty"`
		Description *string `json:"description,omitempty"`
		IconPath    *string `json:"iconPath,omitempty"`
		Volumetric  *bool   `json:"volumetric,omitempty"`
		Universal   *bool   `json:"universal,omitempty"`
		Metric      *bool   `json:"metric,omitempty"`
		Imperial    *bool   `json:"imperial,omitempty"`
		PluralName  *string `json:"pluralName,omitempty"`
		Slug        *string `json:"slug,omitempty"`
	}

	// ValidMeasurementUnitDataManager describes a structure capable of storing valid measurement units permanently.
	ValidMeasurementUnitDataManager interface {
		ValidMeasurementUnitExists(ctx context.Context, validMeasurementUnitID string) (bool, error)
		GetValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string) (*ValidMeasurementUnit, error)
		GetValidMeasurementUnits(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidMeasurementUnit], error)
		SearchForValidMeasurementUnits(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidMeasurementUnit], error)
		ValidMeasurementUnitsForIngredientID(ctx context.Context, validIngredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidMeasurementUnit], error)
		CreateValidMeasurementUnit(ctx context.Context, input *ValidMeasurementUnitDatabaseCreationInput) (*ValidMeasurementUnit, error)
		UpdateValidMeasurementUnit(ctx context.Context, updated *ValidMeasurementUnit) error
		MarkValidMeasurementUnitAsIndexed(ctx context.Context, validMeasurementUnitID string) error
		ArchiveValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string) error
		ScanValidMeasurementUnitIDsForReindex(ctx context.Context, after string, limit int) ([]string, error)
		GetValidMeasurementUnitsWithIDs(ctx context.Context, ids []string) ([]*ValidMeasurementUnit, error)
	}
)

// Update merges a ValidMeasurementUnitUpdateRequestInput with a valid measurement unit.
func (x *ValidMeasurementUnit) Update(input *ValidMeasurementUnitUpdateRequestInput) {
	if input.Name != nil && *input.Name != x.Name {
		x.Name = *input.Name
	}

	if input.Description != nil && *input.Description != x.Description {
		x.Description = *input.Description
	}

	if input.Volumetric != nil && *input.Volumetric != x.Volumetric {
		x.Volumetric = *input.Volumetric
	}

	if input.IconPath != nil && *input.IconPath != x.IconPath {
		x.IconPath = *input.IconPath
	}

	if input.Universal != nil && *input.Universal != x.Universal {
		x.Universal = *input.Universal
	}

	if input.Metric != nil && *input.Metric != x.Metric {
		x.Metric = *input.Metric
	}

	if input.Imperial != nil && *input.Imperial != x.Imperial {
		x.Imperial = *input.Imperial
	}

	if input.PluralName != nil && *input.PluralName != x.PluralName {
		x.PluralName = *input.PluralName
	}

	if input.Slug != nil && *input.Slug != x.Slug {
		x.Slug = *input.Slug
	}
}

var _ validation.ValidatableWithContext = (*ValidMeasurementUnitCreationRequestInput)(nil)

// ValidateWithContext validates a ValidMeasurementUnitCreationRequestInput.
func (x *ValidMeasurementUnitCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	if (x.Metric && x.Imperial) || (!x.Metric && !x.Imperial) {
		result = multierror.Append(result, errMustBeEitherMetricOrImperial)
	}

	if err := validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*ValidMeasurementUnitDatabaseCreationInput)(nil)

// ValidateWithContext validates a ValidMeasurementUnitDatabaseCreationInput.
func (x *ValidMeasurementUnitDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	if (x.Metric && x.Imperial) || (!x.Metric && !x.Imperial) {
		result = multierror.Append(result, errMustBeEitherMetricOrImperial)
	}

	if err := validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Name, validation.Required),
	); err != nil {
		result = multierror.Append(result, err)
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*ValidMeasurementUnitUpdateRequestInput)(nil)

// ValidateWithContext validates a ValidMeasurementUnitUpdateRequestInput.
func (x *ValidMeasurementUnitUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}
