package managers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	eatingindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v3/database/filtering"
	platformerrors "github.com/primandproper/platform-go/v3/errors"
	"github.com/primandproper/platform-go/v3/observability"
	platformkeys "github.com/primandproper/platform-go/v3/observability/keys"
	"github.com/primandproper/platform-go/v3/observability/tracing"
)

func (m *mealPlanningManager) SearchValidVessels(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidVessel], error) {
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
		validVessels *filtering.QueryFilteredResult[types.ValidVessel]
		err          error
	)
	if !useSearchService {
		validVessels, err = m.db.SearchForValidVessels(ctx, query, filter)
	} else {
		var validVesselSubsets []*eatingindexing.ValidVesselSearchSubset
		validVesselSubsets, err = m.validVesselsSearchIndex.Search(ctx, query)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "searching index for valid vessels")
		}

		ids := []string{}
		for _, validVesselSubset := range validVesselSubsets {
			ids = append(ids, validVesselSubset.ID)
		}

		searchResults, searchErr := m.db.GetValidVesselsWithIDs(ctx, ids)
		if searchErr != nil {
			return nil, observability.PrepareAndLogError(searchErr, logger, span, "searching database for valid vessels")
		}

		validVessels = filtering.NewQueryFilteredResult(searchResults, uint64(len(searchResults)), uint64(len(searchResults)), func(v *types.ValidVessel) string {
			return v.ID
		}, filter)
	}
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "searching valid vessels")
	}

	return validVessels, nil
}

func (m *mealPlanningManager) ListValidVessels(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidVessel], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidVessels(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing valid vessels")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateValidVessel(ctx context.Context, input *types.ValidVesselCreationRequestInput) (*types.ValidVessel, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidVesselCreationRequestInputToValidVesselDatabaseCreationInput(input)
	created, err := m.db.CreateValidVessel(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid vessel")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidVesselCreatedServiceEventType, map[string]any{
		mealplanningkeys.ValidVesselIDKey: created.ID,
	}))

	return created, nil
}

func (m *mealPlanningManager) ReadValidVessel(ctx context.Context, validVesselID string) (*types.ValidVessel, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidVesselIDKey, validVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, validVesselID)

	result, err := m.db.GetValidVessel(ctx, validVesselID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid vessel")
	}

	return result, nil
}

func (m *mealPlanningManager) RandomValidVessel(ctx context.Context) (*types.ValidVessel, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	result, err := m.db.GetRandomValidVessel(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching random valid vessel")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateValidVessel(ctx context.Context, validVesselID string, input *types.ValidVesselUpdateRequestInput) (*types.ValidVessel, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidVesselIDKey, validVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, validVesselID)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	existingValidVessel, err := m.db.GetValidVessel(ctx, validVesselID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid vessel")
	}

	existingValidVessel.Update(input)
	if err = m.db.UpdateValidVessel(ctx, existingValidVessel); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid vessel")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidVesselUpdatedServiceEventType, map[string]any{
		mealplanningkeys.ValidVesselIDKey: existingValidVessel.ID,
	}))

	existingValidVessel, err = m.db.GetValidVessel(ctx, validVesselID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid vessel")
	}

	return existingValidVessel, nil
}

func (m *mealPlanningManager) ArchiveValidVessel(ctx context.Context, validVesselID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidVesselIDKey, validVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, validVesselID)

	if err := m.db.ArchiveValidVessel(ctx, validVesselID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid vessel")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidVesselArchivedServiceEventType, map[string]any{
		mealplanningkeys.ValidVesselIDKey: validVesselID,
	}))

	return nil
}
