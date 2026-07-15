package managers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v4/errors"
	"github.com/primandproper/platform-go/v4/filtering"
	"github.com/primandproper/platform-go/v4/observability"
	"github.com/primandproper/platform-go/v4/observability/tracing"
)

func (m *mealPlanningManager) ValidMeasurementUnitConversionsForMeasurementUnit(ctx context.Context, validMeasurementUnitID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnitConversion], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := m.db.GetValidMeasurementUnitConversionsForUnit(ctx, validMeasurementUnitID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid measurement unit conversions from unit")
	}

	return results, nil
}

func (m *mealPlanningManager) GetValidMeasurementUnitConversionsForIngredients(ctx context.Context, validIngredientIDs []string) ([]*types.ValidMeasurementUnitConversion, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidMeasurementUnitConversionsForIngredients(ctx, validIngredientIDs)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid measurement unit conversions for ingredients")
	}

	return results, nil
}

func (m *mealPlanningManager) GetMeasurementUnitConversionMismatches(ctx context.Context) ([]*types.MeasurementUnitConversionMismatch, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetMeasurementUnitConversionMismatches(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching measurement unit conversion mismatches")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateValidMeasurementUnitConversion(ctx context.Context, input *types.ValidMeasurementUnitConversionCreationRequestInput) (*types.ValidMeasurementUnitConversion, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidMeasurementUnitConversionCreationRequestInputToValidMeasurementUnitConversionDatabaseCreationInput(input)
	created, err := m.db.CreateValidMeasurementUnitConversion(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid measurement unit conversion")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidMeasurementUnitConversionCreatedServiceEventType, map[string]any{
		mealplanningkeys.ValidMeasurementUnitConversionIDKey: created.ID,
	}))

	return created, nil
}

func (m *mealPlanningManager) ReadValidMeasurementUnitConversion(ctx context.Context, validMeasurementUnitConversionID string) (*types.ValidMeasurementUnitConversion, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitConversionIDKey, validMeasurementUnitConversionID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitConversionIDKey, validMeasurementUnitConversionID)

	result, err := m.db.GetValidMeasurementUnitConversion(ctx, validMeasurementUnitConversionID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid measurement unit conversion")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateValidMeasurementUnitConversion(ctx context.Context, validMeasurementUnitConversionID string, input *types.ValidMeasurementUnitConversionUpdateRequestInput) (*types.ValidMeasurementUnitConversion, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitConversionIDKey, validMeasurementUnitConversionID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitConversionIDKey, validMeasurementUnitConversionID)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	existingValidMeasurementUnitConversion, err := m.db.GetValidMeasurementUnitConversion(ctx, validMeasurementUnitConversionID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid measurement unit conversion")
	}

	existingValidMeasurementUnitConversion.Update(input)
	if err = m.db.UpdateValidMeasurementUnitConversion(ctx, existingValidMeasurementUnitConversion); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid measurement unit conversion")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidMeasurementUnitConversionUpdatedServiceEventType, map[string]any{
		mealplanningkeys.ValidMeasurementUnitConversionIDKey: existingValidMeasurementUnitConversion.ID,
	}))

	existingValidMeasurementUnitConversion, err = m.db.GetValidMeasurementUnitConversion(ctx, validMeasurementUnitConversionID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid measurement unit conversion")
	}

	return existingValidMeasurementUnitConversion, nil
}

func (m *mealPlanningManager) ArchiveValidMeasurementUnitConversion(ctx context.Context, validMeasurementUnitConversionID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitConversionIDKey, validMeasurementUnitConversionID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitConversionIDKey, validMeasurementUnitConversionID)

	if err := m.db.ArchiveValidMeasurementUnitConversion(ctx, validMeasurementUnitConversionID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid measurement unit conversion")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidMeasurementUnitConversionArchivedServiceEventType, map[string]any{
		mealplanningkeys.ValidMeasurementUnitConversionIDKey: validMeasurementUnitConversionID,
	}))

	return nil
}
