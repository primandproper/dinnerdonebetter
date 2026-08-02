package managers

import (
	"context"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

func (m *mealPlanningManager) ListValidIngredientMeasurementUnits(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientMeasurementUnit], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidIngredientMeasurementUnits(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient measurement units")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateValidIngredientMeasurementUnit(ctx context.Context, input *types.ValidIngredientMeasurementUnitCreationRequestInput) (*types.ValidIngredientMeasurementUnit, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidIngredientMeasurementUnitCreationRequestInputToValidIngredientMeasurementUnitDatabaseCreationInput(input)
	created, err := m.db.CreateValidIngredientMeasurementUnit(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid ingredient measurement unit")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string) (*types.ValidIngredientMeasurementUnit, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientMeasurementUnitIDKey, validIngredientMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientMeasurementUnitIDKey, validIngredientMeasurementUnitID)

	result, err := m.db.GetValidIngredientMeasurementUnit(ctx, validIngredientMeasurementUnitID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient measurement unit")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string, input *types.ValidIngredientMeasurementUnitUpdateRequestInput) (*types.ValidIngredientMeasurementUnit, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientMeasurementUnitIDKey, validIngredientMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientMeasurementUnitIDKey, validIngredientMeasurementUnitID)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	existingValidIngredientMeasurementUnit, err := m.db.GetValidIngredientMeasurementUnit(ctx, validIngredientMeasurementUnitID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient measurement unit")
	}

	existingValidIngredientMeasurementUnit.Update(input)
	if err = m.db.UpdateValidIngredientMeasurementUnit(ctx, existingValidIngredientMeasurementUnit); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid ingredient measurement unit")
	}

	existingValidIngredientMeasurementUnit, err = m.db.GetValidIngredientMeasurementUnit(ctx, validIngredientMeasurementUnitID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid ingredient measurement unit")
	}

	return existingValidIngredientMeasurementUnit, nil
}

func (m *mealPlanningManager) ArchiveValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientMeasurementUnitIDKey, validIngredientMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientMeasurementUnitIDKey, validIngredientMeasurementUnitID)

	if err := m.db.ArchiveValidIngredientMeasurementUnit(ctx, validIngredientMeasurementUnitID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid ingredient measurement unit")
	}

	return nil
}

func (m *mealPlanningManager) SearchValidIngredientMeasurementUnitsByIngredient(ctx context.Context, validIngredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientMeasurementUnit], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	results, err := m.db.GetValidIngredientMeasurementUnitsForIngredient(ctx, validIngredientID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient measurement units for ingredient")
	}

	return results, nil
}

func (m *mealPlanningManager) SearchValidIngredientMeasurementUnitsByMeasurementUnit(ctx context.Context, validMeasurementUnitID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientMeasurementUnit], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)

	results, err := m.db.GetValidIngredientMeasurementUnitsForMeasurementUnit(ctx, validMeasurementUnitID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient measurement units for measurement unit")
	}

	return results, nil
}
