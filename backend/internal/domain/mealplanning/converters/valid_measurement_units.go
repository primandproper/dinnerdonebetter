package converters

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v11/identifiers"
)

// ConvertValidMeasurementUnitToValidMeasurementUnitUpdateRequestInput creates a ValidMeasurementUnitUpdateRequestInput from a MeasurementUnit.
func ConvertValidMeasurementUnitToValidMeasurementUnitUpdateRequestInput(input *types.ValidMeasurementUnit) *types.ValidMeasurementUnitUpdateRequestInput {
	x := &types.ValidMeasurementUnitUpdateRequestInput{
		Name:        &input.Name,
		Description: &input.Description,
		IconPath:    &input.IconPath,
		Volumetric:  &input.Volumetric,
		Universal:   &input.Universal,
		Metric:      &input.Metric,
		Imperial:    &input.Imperial,
		Slug:        &input.Slug,
		PluralName:  &input.PluralName,
	}

	return x
}

// ConvertValidMeasurementUnitCreationRequestInputToValidMeasurementUnitDatabaseCreationInput creates a ValidMeasurementUnitDatabaseCreationInput from a ValidMeasurementUnitCreationRequestInput.
func ConvertValidMeasurementUnitCreationRequestInputToValidMeasurementUnitDatabaseCreationInput(input *types.ValidMeasurementUnitCreationRequestInput) *types.ValidMeasurementUnitDatabaseCreationInput {
	x := &types.ValidMeasurementUnitDatabaseCreationInput{
		ID:          identifiers.New(),
		Name:        input.Name,
		Description: input.Description,
		Volumetric:  input.Volumetric,
		IconPath:    input.IconPath,
		Universal:   input.Universal,
		Metric:      input.Metric,
		Imperial:    input.Imperial,
		Slug:        input.Slug,
		PluralName:  input.PluralName,
	}

	return x
}

// ConvertValidMeasurementUnitToValidMeasurementUnitCreationRequestInput builds a ValidMeasurementUnitCreationRequestInput from a MeasurementUnit.
func ConvertValidMeasurementUnitToValidMeasurementUnitCreationRequestInput(validMeasurementUnit *types.ValidMeasurementUnit) *types.ValidMeasurementUnitCreationRequestInput {
	return &types.ValidMeasurementUnitCreationRequestInput{
		Name:        validMeasurementUnit.Name,
		Description: validMeasurementUnit.Description,
		Volumetric:  validMeasurementUnit.Volumetric,
		IconPath:    validMeasurementUnit.IconPath,
		Universal:   validMeasurementUnit.Universal,
		Metric:      validMeasurementUnit.Metric,
		Imperial:    validMeasurementUnit.Imperial,
		PluralName:  validMeasurementUnit.PluralName,
		Slug:        validMeasurementUnit.Slug,
	}
}

// ConvertValidMeasurementUnitToValidMeasurementUnitDatabaseCreationInput builds a ValidMeasurementUnitDatabaseCreationInput from a MeasurementUnit.
func ConvertValidMeasurementUnitToValidMeasurementUnitDatabaseCreationInput(validMeasurementUnit *types.ValidMeasurementUnit) *types.ValidMeasurementUnitDatabaseCreationInput {
	return &types.ValidMeasurementUnitDatabaseCreationInput{
		ID:          validMeasurementUnit.ID,
		Name:        validMeasurementUnit.Name,
		Description: validMeasurementUnit.Description,
		Volumetric:  validMeasurementUnit.Volumetric,
		IconPath:    validMeasurementUnit.IconPath,
		Universal:   validMeasurementUnit.Universal,
		Metric:      validMeasurementUnit.Metric,
		Imperial:    validMeasurementUnit.Imperial,
		PluralName:  validMeasurementUnit.PluralName,
		Slug:        validMeasurementUnit.Slug,
	}
}

// ConvertNullableValidMeasurementUnitToValidMeasurementUnit produces a ValidMeasurementUnit from a NullableValidMeasurementUnit.
func ConvertNullableValidMeasurementUnitToValidMeasurementUnit(x *types.NullableValidMeasurementUnit) *types.ValidMeasurementUnit {
	if x != nil && x.ID != nil {
		return &types.ValidMeasurementUnit{
			CreatedAt:     *x.CreatedAt,
			LastUpdatedAt: x.LastUpdatedAt,
			ArchivedAt:    x.ArchivedAt,
			Name:          *x.Name,
			IconPath:      *x.IconPath,
			ID:            *x.ID,
			Description:   *x.Description,
			PluralName:    *x.PluralName,
			Slug:          *x.Slug,
			Volumetric:    *x.Volumetric,
			Universal:     *x.Universal,
			Metric:        *x.Metric,
			Imperial:      *x.Imperial,
		}
	}
	return nil
}

// ConvertValidMeasurementUnitToNullableValidMeasurementUnit converts a NullableValidMeasurementUnit to a ValidMeasurementUnit.
func ConvertValidMeasurementUnitToNullableValidMeasurementUnit(input *types.ValidMeasurementUnit) *types.NullableValidMeasurementUnit {
	return &types.NullableValidMeasurementUnit{
		CreatedAt:     &input.CreatedAt,
		LastUpdatedAt: input.LastUpdatedAt,
		ArchivedAt:    input.ArchivedAt,
		Name:          &input.Name,
		IconPath:      &input.IconPath,
		ID:            &input.ID,
		Description:   &input.Description,
		PluralName:    &input.PluralName,
		Slug:          &input.Slug,
		Volumetric:    &input.Volumetric,
		Universal:     &input.Universal,
		Metric:        &input.Metric,
		Imperial:      &input.Imperial,
	}
}
