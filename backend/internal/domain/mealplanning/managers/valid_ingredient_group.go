package managers

import (
	"context"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/observability"
	platformkeys "github.com/primandproper/platform-go/v10/observability/keys"
	"github.com/primandproper/platform-go/v10/observability/tracing"
)

// Interface, struct, and constructor are defined in manager.go.
func (m *mealPlanningManager) SearchValidIngredientGroups(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientGroup], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(platformkeys.SearchQueryKey, query).WithValue(platformkeys.UseDatabaseKey, !useSearchService)
	tracing.AttachToSpan(span, platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, platformkeys.UseDatabaseKey, !useSearchService)

	results, err := m.db.SearchForValidIngredientGroups(ctx, query, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "searching for valid ingredient groups failed")
	}

	return results, nil
}

func (m *mealPlanningManager) ListValidIngredientGroups(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientGroup], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidIngredientGroups(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing valid ingredient groups")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateValidIngredientGroup(ctx context.Context, input *types.ValidIngredientGroupCreationRequestInput) (*types.ValidIngredientGroup, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidIngredientGroupCreationRequestInputToValidIngredientGroupDatabaseCreationInput(input)
	created, err := m.db.CreateValidIngredientGroup(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid ingredient group")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadValidIngredientGroup(ctx context.Context, validIngredientGroupID string) (*types.ValidIngredientGroup, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID)

	result, err := m.db.GetValidIngredientGroup(ctx, validIngredientGroupID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting valid ingredient group")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateValidIngredientGroup(ctx context.Context, validIngredientGroupID string, input *types.ValidIngredientGroupUpdateRequestInput) (*types.ValidIngredientGroup, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	existingValidIngredientGroup, err := m.db.GetValidIngredientGroup(ctx, validIngredientGroupID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient group")
	}

	existingValidIngredientGroup.Update(input)
	if err = m.db.UpdateValidIngredientGroup(ctx, existingValidIngredientGroup); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid ingredient group")
	}

	existingValidIngredientGroup, err = m.db.GetValidIngredientGroup(ctx, validIngredientGroupID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid ingredient group")
	}

	return existingValidIngredientGroup, nil
}

func (m *mealPlanningManager) ArchiveValidIngredientGroup(ctx context.Context, validIngredientGroupID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID)

	if err := m.db.ArchiveValidIngredientGroup(ctx, validIngredientGroupID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid ingredient group")
	}

	return nil
}
