package managers

import (
	"context"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

func (m *mealPlanningManager) ListValidPreparationVessels(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationVessel], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidPreparationVessels(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing valid preparation vessels")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateValidPreparationVessel(ctx context.Context, input *types.ValidPreparationVesselCreationRequestInput) (*types.ValidPreparationVessel, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidPreparationVesselCreationRequestInputToValidPreparationVesselDatabaseCreationInput(input)
	created, err := m.db.CreateValidPreparationVessel(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid preparation vessel")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadValidPreparationVessel(ctx context.Context, validPreparationVesselID string) (*types.ValidPreparationVessel, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationVesselIDKey, validPreparationVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationVesselIDKey, validPreparationVesselID)

	result, err := m.db.GetValidPreparationVessel(ctx, validPreparationVesselID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation vessel")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateValidPreparationVessel(ctx context.Context, validPreparationVesselID string, input *types.ValidPreparationVesselUpdateRequestInput) (*types.ValidPreparationVessel, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationVesselIDKey, validPreparationVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationVesselIDKey, validPreparationVesselID)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	existingValidPreparationVessel, err := m.db.GetValidPreparationVessel(ctx, validPreparationVesselID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation vessel")
	}

	existingValidPreparationVessel.Update(input)
	if err = m.db.UpdateValidPreparationVessel(ctx, existingValidPreparationVessel); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid preparation vessel")
	}

	existingValidPreparationVessel, err = m.db.GetValidPreparationVessel(ctx, validPreparationVesselID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid preparation vessel")
	}

	return existingValidPreparationVessel, nil
}

func (m *mealPlanningManager) ArchiveValidPreparationVessel(ctx context.Context, validPreparationVesselID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationVesselIDKey, validPreparationVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationVesselIDKey, validPreparationVesselID)

	if err := m.db.ArchiveValidPreparationVessel(ctx, validPreparationVesselID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid preparation vessel")
	}

	return nil
}

func (m *mealPlanningManager) SearchValidPreparationVesselsByPreparation(ctx context.Context, validPreparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationVessel], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, validPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, validPreparationID)

	validPreparationVessels, err := m.db.GetValidPreparationVesselsForPreparation(ctx, validPreparationID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "searching valid preparation vessels by preparation")
	}

	return validPreparationVessels, nil
}

func (m *mealPlanningManager) SearchValidPreparationVesselsByVessel(ctx context.Context, validVesselID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationVessel], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidVesselIDKey, validVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, validVesselID)

	validPreparationVessels, err := m.db.GetValidPreparationVesselsForVessel(ctx, validVesselID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "searching valid preparation vessels by vessel")
	}

	return validPreparationVessels, nil
}
