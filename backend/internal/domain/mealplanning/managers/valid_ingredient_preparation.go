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

func (m *mealPlanningManager) ListValidIngredientPreparations(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientPreparation], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidIngredientPreparations(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient preparations")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateValidIngredientPreparation(ctx context.Context, input *types.ValidIngredientPreparationCreationRequestInput) (*types.ValidIngredientPreparation, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidIngredientPreparationCreationRequestInputToValidIngredientPreparationDatabaseCreationInput(input)
	created, err := m.db.CreateValidIngredientPreparation(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid ingredient preparation")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidIngredientPreparationCreatedServiceEventType, map[string]any{
		mealplanningkeys.ValidIngredientPreparationIDKey: created.ID,
	}))

	return created, nil
}

func (m *mealPlanningManager) ReadValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string) (*types.ValidIngredientPreparation, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientPreparationIDKey, validIngredientPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, validIngredientPreparationID)

	result, err := m.db.GetValidIngredientPreparation(ctx, validIngredientPreparationID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient preparation")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string, input *types.ValidIngredientPreparationUpdateRequestInput) (*types.ValidIngredientPreparation, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientPreparationIDKey, validIngredientPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, validIngredientPreparationID)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	existingValidIngredientPreparation, err := m.db.GetValidIngredientPreparation(ctx, validIngredientPreparationID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient preparation")
	}

	existingValidIngredientPreparation.Update(input)
	if err = m.db.UpdateValidIngredientPreparation(ctx, existingValidIngredientPreparation); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid ingredient preparation")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidIngredientPreparationUpdatedServiceEventType, map[string]any{
		mealplanningkeys.ValidIngredientPreparationIDKey: existingValidIngredientPreparation.ID,
	}))

	existingValidIngredientPreparation, err = m.db.GetValidIngredientPreparation(ctx, validIngredientPreparationID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid ingredient preparation")
	}

	return existingValidIngredientPreparation, nil
}

func (m *mealPlanningManager) ArchiveValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientPreparationIDKey, validIngredientPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, validIngredientPreparationID)

	if err := m.db.ArchiveValidIngredientPreparation(ctx, validIngredientPreparationID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid ingredient preparation")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidIngredientPreparationArchivedServiceEventType, map[string]any{
		mealplanningkeys.ValidIngredientPreparationIDKey: validIngredientPreparationID,
	}))

	return nil
}

func (m *mealPlanningManager) SearchValidIngredientPreparationsByIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientPreparation], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, ingredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, ingredientID)

	results, err := m.db.GetValidIngredientPreparationsForIngredient(ctx, ingredientID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient preparations for ingredient")
	}

	return results, nil
}

func (m *mealPlanningManager) SearchValidIngredientPreparationsByPreparation(ctx context.Context, validPreparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientPreparation], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, validPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, validPreparationID)

	results, err := m.db.GetValidIngredientPreparationsForPreparation(ctx, validPreparationID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient preparations for ingredient")
	}

	return results, nil
}
