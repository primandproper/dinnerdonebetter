package managers

import (
	"context"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	eatingindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/observability"
	platformkeys "github.com/primandproper/platform-go/v11/observability/keys"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	searchpagination "github.com/primandproper/platform-go/v11/search/pagination"
)

func (m *mealPlanningManager) SearchValidIngredientStates(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientState], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(platformkeys.SearchQueryKey, query).WithValue(platformkeys.UseDatabaseKey, !useSearchService)
	tracing.AttachToSpan(span, platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, platformkeys.UseDatabaseKey, !useSearchService)

	var (
		results *filtering.QueryFilteredResult[types.ValidIngredientState]
		err     error
	)
	if !useSearchService {
		results, err = m.db.SearchForValidIngredientStates(ctx, query, filter)
	} else {
		results, err = searchpagination.Hydrated(ctx, m.validIngredientStatesSearchIndex, query, filter,
			func(subset *eatingindexing.ValidIngredientStateSearchSubset) string { return subset.ID },
			m.db.GetValidIngredientStatesWithIDs,
		)
	}

	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient states")
	}

	return results, nil
}

func (m *mealPlanningManager) ListValidIngredientStates(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientState], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidIngredientStates(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing valid ingredient states")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateValidIngredientState(ctx context.Context, input *types.ValidIngredientStateCreationRequestInput) (*types.ValidIngredientState, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidIngredientStateCreationRequestInputToValidIngredientStateDatabaseCreationInput(input)
	created, err := m.db.CreateValidIngredientState(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid ingredient state")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadValidIngredientState(ctx context.Context, validIngredientStateID string) (*types.ValidIngredientState, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)

	result, err := m.db.GetValidIngredientState(ctx, validIngredientStateID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient state")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateValidIngredientState(ctx context.Context, validIngredientStateID string, input *types.ValidIngredientStateUpdateRequestInput) (*types.ValidIngredientState, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	existingValidIngredientState, err := m.db.GetValidIngredientState(ctx, validIngredientStateID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient state")
	}

	existingValidIngredientState.Update(input)
	if err = m.db.UpdateValidIngredientState(ctx, existingValidIngredientState); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid ingredient state")
	}

	existingValidIngredientState, err = m.db.GetValidIngredientState(ctx, validIngredientStateID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid ingredient state")
	}

	return existingValidIngredientState, nil
}

func (m *mealPlanningManager) ArchiveValidIngredientState(ctx context.Context, validIngredientStateID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)

	if err := m.db.ArchiveValidIngredientState(ctx, validIngredientStateID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid ingredient state")
	}

	return nil
}
