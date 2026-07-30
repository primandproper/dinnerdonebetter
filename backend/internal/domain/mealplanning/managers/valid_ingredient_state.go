package managers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	eatingindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/observability"
	platformkeys "github.com/primandproper/platform-go/v8/observability/keys"
	"github.com/primandproper/platform-go/v8/observability/tracing"
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
		var validIngredientStateSubsets []*eatingindexing.ValidIngredientStateSearchSubset
		validIngredientStateSubsets, err = m.validIngredientStatesSearchIndex.Search(ctx, query)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "searching valid ingredient states")
		}

		ids := []string{}
		for _, validIngredientStateSubset := range validIngredientStateSubsets {
			ids = append(ids, validIngredientStateSubset.ID)
		}

		searchResults, searchErr := m.db.GetValidIngredientStatesWithIDs(ctx, ids)
		if searchErr != nil {
			return nil, observability.PrepareAndLogError(searchErr, logger, span, "fetching valid ingredient states")
		}

		results = filtering.NewQueryFilteredResult(searchResults, uint64(len(searchResults)), uint64(len(searchResults)), func(v *types.ValidIngredientState) string {
			return v.ID
		}, filter)
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

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidIngredientStateCreatedServiceEventType, map[string]any{
		mealplanningkeys.ValidIngredientStateIDKey: created.ID,
	}))

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

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidIngredientStateUpdatedServiceEventType, map[string]any{
		mealplanningkeys.ValidIngredientStateIDKey: existingValidIngredientState.ID,
	}))

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

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidIngredientStateArchivedServiceEventType, map[string]any{
		mealplanningkeys.ValidIngredientStateIDKey: validIngredientStateID,
	}))

	return nil
}
