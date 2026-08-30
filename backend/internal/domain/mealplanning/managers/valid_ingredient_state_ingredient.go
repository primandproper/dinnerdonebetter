package managers

import (
	"context"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

func (m *mealPlanningManager) ListValidIngredientStateIngredients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientStateIngredient], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidIngredientStateIngredients(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing valid ingredient state ingredients")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateValidIngredientStateIngredient(ctx context.Context, input *types.ValidIngredientStateIngredientCreationRequestInput) (*types.ValidIngredientStateIngredient, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidIngredientStateIngredientCreationRequestInputToValidIngredientStateIngredientDatabaseCreationInput(input)
	created, err := m.db.CreateValidIngredientStateIngredient(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid ingredient state ingredient")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadValidIngredientStateIngredient(ctx context.Context, validIngredientStateIngredientID string) (*types.ValidIngredientStateIngredient, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIngredientIDKey, validIngredientStateIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, validIngredientStateIngredientID)

	result, err := m.db.GetValidIngredientStateIngredient(ctx, validIngredientStateIngredientID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient state ingredient")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateValidIngredientStateIngredient(ctx context.Context, validIngredientStateIngredientID string, input *types.ValidIngredientStateIngredientUpdateRequestInput) (*types.ValidIngredientStateIngredient, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIngredientIDKey, validIngredientStateIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, validIngredientStateIngredientID)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	existingValidIngredientStateIngredient, err := m.db.GetValidIngredientStateIngredient(ctx, validIngredientStateIngredientID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient state ingredient")
	}

	existingValidIngredientStateIngredient.Update(input)
	if err = m.db.UpdateValidIngredientStateIngredient(ctx, existingValidIngredientStateIngredient); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid ingredient state ingredient")
	}

	existingValidIngredientStateIngredient, err = m.db.GetValidIngredientStateIngredient(ctx, validIngredientStateIngredientID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid ingredient state ingredient")
	}

	return existingValidIngredientStateIngredient, nil
}

func (m *mealPlanningManager) ArchiveValidIngredientStateIngredient(ctx context.Context, validIngredientStateIngredientID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIngredientIDKey, validIngredientStateIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, validIngredientStateIngredientID)

	if err := m.db.ArchiveValidIngredientStateIngredient(ctx, validIngredientStateIngredientID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid ingredient state ingredient")
	}

	return nil
}

func (m *mealPlanningManager) SearchValidIngredientStateIngredientsByIngredient(ctx context.Context, validIngredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientStateIngredient], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	results, err := m.db.GetValidIngredientStateIngredientsForIngredient(ctx, validIngredientID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient state ingredients for ingredient")
	}

	return results, nil
}

func (m *mealPlanningManager) SearchValidIngredientStateIngredientsByIngredientState(ctx context.Context, validIngredientStateID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientStateIngredient], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)

	results, err := m.db.GetValidIngredientStateIngredientsForIngredientState(ctx, validIngredientStateID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient state ingredients for ingredient state")
	}

	return results, nil
}
