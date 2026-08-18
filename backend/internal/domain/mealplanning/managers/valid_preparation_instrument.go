package managers

import (
	"context"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/observability"
	"github.com/primandproper/platform-go/v11/observability/tracing"
)

func (m *mealPlanningManager) ListValidPreparationInstruments(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationInstrument], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidPreparationInstruments(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing valid preparation instruments")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateValidPreparationInstrument(ctx context.Context, input *types.ValidPreparationInstrumentCreationRequestInput) (*types.ValidPreparationInstrument, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidPreparationInstrumentCreationRequestInputToValidPreparationInstrumentDatabaseCreationInput(input)
	created, err := m.db.CreateValidPreparationInstrument(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid preparation instrument")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string) (*types.ValidPreparationInstrument, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationInstrumentIDKey, validPreparationInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, validPreparationInstrumentID)

	result, err := m.db.GetValidPreparationInstrument(ctx, validPreparationInstrumentID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation instrument")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string, input *types.ValidPreparationInstrumentUpdateRequestInput) (*types.ValidPreparationInstrument, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationInstrumentIDKey, validPreparationInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, validPreparationInstrumentID)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	existingValidPreparationInstrument, err := m.db.GetValidPreparationInstrument(ctx, validPreparationInstrumentID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation instrument")
	}

	existingValidPreparationInstrument.Update(input)
	if err = m.db.UpdateValidPreparationInstrument(ctx, existingValidPreparationInstrument); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid preparation instrument")
	}

	existingValidPreparationInstrument, err = m.db.GetValidPreparationInstrument(ctx, validPreparationInstrumentID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid preparation instrument")
	}

	return existingValidPreparationInstrument, nil
}

func (m *mealPlanningManager) ArchiveValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationInstrumentIDKey, validPreparationInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, validPreparationInstrumentID)

	if err := m.db.ArchiveValidPreparationInstrument(ctx, validPreparationInstrumentID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid preparation instrument")
	}

	return nil
}

func (m *mealPlanningManager) SearchValidPreparationInstrumentsByPreparation(ctx context.Context, validPreparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationInstrument], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, validPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, validPreparationID)

	validPreparationInstruments, err := m.db.GetValidPreparationInstrumentsForPreparation(ctx, validPreparationID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "searching valid preparation instruments by preparation")
	}

	return validPreparationInstruments, nil
}

func (m *mealPlanningManager) SearchValidPreparationInstrumentsByInstrument(ctx context.Context, validInstrumentID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationInstrument], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)

	validPreparationInstruments, err := m.db.GetValidPreparationInstrumentsForInstrument(ctx, validInstrumentID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "searching valid preparation instruments by instrument")
	}

	return validPreparationInstruments, nil
}
