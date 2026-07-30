package managers

import (
	"context"

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

func (m *mealPlanningManager) SearchValidInstruments(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidInstrument], error) {
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
		results *filtering.QueryFilteredResult[types.ValidInstrument]
		err     error
	)
	if !useSearchService {
		results, err = m.db.SearchForValidInstruments(ctx, query, filter)
	} else {
		var validInstrumentSubsets []*eatingindexing.ValidInstrumentSearchSubset
		validInstrumentSubsets, err = m.validInstrumentSearchIndex.Search(ctx, query)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "searching for valid instruments")
		}

		ids := []string{}
		for _, validInstrumentSubset := range validInstrumentSubsets {
			ids = append(ids, validInstrumentSubset.ID)
		}

		searchResults, searchErr := m.db.GetValidInstrumentsWithIDs(ctx, ids)
		if searchErr != nil {
			return nil, observability.PrepareAndLogError(searchErr, logger, span, "fetching valid instruments")
		}

		results = filtering.NewQueryFilteredResult(
			searchResults,
			uint64(len(searchResults)),
			uint64(len(searchResults)),
			func(v *types.ValidInstrument) string {
				return v.ID
			},
			filter,
		)
	}

	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "searching for valid instruments")
	}

	return results, nil
}

func (m *mealPlanningManager) ListValidInstruments(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidInstrument], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidInstruments(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing valid instruments")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateValidInstrument(ctx context.Context, input *types.ValidInstrumentCreationRequestInput) (*types.ValidInstrument, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidInstrumentCreationRequestInputToValidInstrumentDatabaseCreationInput(input)
	created, err := m.db.CreateValidInstrument(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid instrument")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadValidInstrument(ctx context.Context, validInstrumentID string) (*types.ValidInstrument, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)

	result, err := m.db.GetValidInstrument(ctx, validInstrumentID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid instrument")
	}

	return result, nil
}

func (m *mealPlanningManager) RandomValidInstrument(ctx context.Context) (*types.ValidInstrument, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	result, err := m.db.GetRandomValidInstrument(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching random valid instrument")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateValidInstrument(ctx context.Context, validInstrumentID string, input *types.ValidInstrumentUpdateRequestInput) (*types.ValidInstrument, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	existingValidInstrument, err := m.db.GetValidInstrument(ctx, validInstrumentID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid instrument")
	}

	existingValidInstrument.Update(input)
	if err = m.db.UpdateValidInstrument(ctx, existingValidInstrument); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid instrument")
	}

	existingValidInstrument, err = m.db.GetValidInstrument(ctx, validInstrumentID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid instrument")
	}

	return existingValidInstrument, nil
}

func (m *mealPlanningManager) ArchiveValidInstrument(ctx context.Context, validInstrumentID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)

	if err := m.db.ArchiveValidInstrument(ctx, validInstrumentID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid instrument")
	}

	return nil
}
