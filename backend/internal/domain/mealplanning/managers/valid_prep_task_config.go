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

func (m *mealPlanningManager) ListValidPrepTaskConfigs(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPrepTaskConfig], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidPrepTaskConfigs(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid prep task configs")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateValidPrepTaskConfig(ctx context.Context, input *types.ValidPrepTaskConfigCreationRequestInput) (*types.ValidPrepTaskConfig, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidPrepTaskConfigCreationRequestInputToValidPrepTaskConfigDatabaseCreationInput(input)
	created, err := m.db.CreateValidPrepTaskConfig(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid prep task config")
	}

	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return created, nil
}

func (m *mealPlanningManager) ReadValidPrepTaskConfig(ctx context.Context, validPrepTaskConfigID string) (*types.ValidPrepTaskConfig, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPrepTaskConfigIDKey, validPrepTaskConfigID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPrepTaskConfigIDKey, validPrepTaskConfigID)

	result, err := m.db.GetValidPrepTaskConfig(ctx, validPrepTaskConfigID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid prep task config")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateValidPrepTaskConfig(ctx context.Context, validPrepTaskConfigID string, input *types.ValidPrepTaskConfigUpdateRequestInput) (*types.ValidPrepTaskConfig, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPrepTaskConfigIDKey, validPrepTaskConfigID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPrepTaskConfigIDKey, validPrepTaskConfigID)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	existingValidPrepTaskConfig, err := m.db.GetValidPrepTaskConfig(ctx, validPrepTaskConfigID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid prep task config")
	}

	existingValidPrepTaskConfig.Update(input)
	if err = m.db.UpdateValidPrepTaskConfig(ctx, existingValidPrepTaskConfig); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid prep task config")
	}

	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	existingValidPrepTaskConfig, err = m.db.GetValidPrepTaskConfig(ctx, validPrepTaskConfigID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid prep task config")
	}

	return existingValidPrepTaskConfig, nil
}

func (m *mealPlanningManager) ArchiveValidPrepTaskConfig(ctx context.Context, validPrepTaskConfigID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPrepTaskConfigIDKey, validPrepTaskConfigID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPrepTaskConfigIDKey, validPrepTaskConfigID)

	if err := m.db.ArchiveValidPrepTaskConfig(ctx, validPrepTaskConfigID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid prep task config")
	}

	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return nil
}

func (m *mealPlanningManager) SearchValidPrepTaskConfigsByIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPrepTaskConfig], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, ingredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, ingredientID)

	results, err := m.db.GetValidPrepTaskConfigsForIngredient(ctx, ingredientID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid prep task configs for ingredient")
	}

	return results, nil
}

func (m *mealPlanningManager) SearchValidPrepTaskConfigsByPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPrepTaskConfig], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, preparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, preparationID)

	results, err := m.db.GetValidPrepTaskConfigsForPreparation(ctx, preparationID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid prep task configs for preparation")
	}

	return results, nil
}

func (m *mealPlanningManager) SearchValidPrepTaskConfigsByIngredientAndPreparation(ctx context.Context, ingredientID, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPrepTaskConfig], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).
		WithValue(mealplanningkeys.ValidIngredientIDKey, ingredientID).
		WithValue(mealplanningkeys.ValidPreparationIDKey, preparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, ingredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, preparationID)

	results, err := m.db.GetValidPrepTaskConfigsForIngredientAndPreparation(ctx, ingredientID, preparationID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid prep task configs for ingredient and preparation")
	}

	return results, nil
}
