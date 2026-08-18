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
		results, err = searchpagination.Hydrated(ctx, m.validInstrumentSearchIndex, query, filter,
			func(subset *eatingindexing.ValidInstrumentSearchSubset) string { return subset.ID },
			m.db.GetValidInstrumentsWithIDs,
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
