package managers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	eatingindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	platformerrors "github.com/primandproper/platform-go/v7/errors"
	"github.com/primandproper/platform-go/v7/filtering"
	"github.com/primandproper/platform-go/v7/observability"
	platformkeys "github.com/primandproper/platform-go/v7/observability/keys"
	"github.com/primandproper/platform-go/v7/observability/tracing"
)

func (m *mealPlanningManager) SearchValidMeasurementUnits(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnit], error) {
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
		results *filtering.QueryFilteredResult[types.ValidMeasurementUnit]
		err     error
	)
	if !useSearchService {
		results, err = m.db.SearchForValidMeasurementUnits(ctx, query, filter)
	} else {
		var validMeasurementUnitSubsets []*eatingindexing.ValidMeasurementUnitSearchSubset
		validMeasurementUnitSubsets, err = m.validMeasurementUnitSearchIndex.Search(ctx, query)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "searching for valid measurement units")
		}

		ids := []string{}
		for _, validMeasurementUnitSubset := range validMeasurementUnitSubsets {
			ids = append(ids, validMeasurementUnitSubset.ID)
		}

		searchResults, searchErr := m.db.GetValidMeasurementUnitsWithIDs(ctx, ids)
		if searchErr != nil {
			return nil, observability.PrepareAndLogError(searchErr, logger, span, "fetching valid measurement units")
		}

		results = filtering.NewQueryFilteredResult(searchResults, uint64(len(searchResults)), uint64(len(searchResults)), func(v *types.ValidMeasurementUnit) string {
			return v.ID
		}, filter)
	}

	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "searching for valid measurement units")
	}

	return results, nil
}

func (m *mealPlanningManager) SearchValidMeasurementUnitsByIngredientID(ctx context.Context, validIngredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnit], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	validMeasurementUnits, err := m.db.ValidMeasurementUnitsForIngredientID(ctx, validIngredientID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "searching valid measurement units for ingredient")
	}

	return validMeasurementUnits, nil
}

func (m *mealPlanningManager) ListValidMeasurementUnits(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnit], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidMeasurementUnits(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid measurement units")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateValidMeasurementUnit(ctx context.Context, input *types.ValidMeasurementUnitCreationRequestInput) (*types.ValidMeasurementUnit, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidMeasurementUnitCreationRequestInputToValidMeasurementUnitDatabaseCreationInput(input)
	created, err := m.db.CreateValidMeasurementUnit(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid measurement unit")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidMeasurementUnitCreatedServiceEventType, map[string]any{
		mealplanningkeys.ValidMeasurementUnitIDKey: created.ID,
	}))

	return created, nil
}

func (m *mealPlanningManager) ReadValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string) (*types.ValidMeasurementUnit, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)

	result, err := m.db.GetValidMeasurementUnit(ctx, validMeasurementUnitID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid measurement unit")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string, input *types.ValidMeasurementUnitUpdateRequestInput) (*types.ValidMeasurementUnit, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	existingValidMeasurementUnit, err := m.db.GetValidMeasurementUnit(ctx, validMeasurementUnitID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid measurement unit")
	}

	existingValidMeasurementUnit.Update(input)
	if err = m.db.UpdateValidMeasurementUnit(ctx, existingValidMeasurementUnit); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid measurement unit")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidMeasurementUnitUpdatedServiceEventType, map[string]any{
		mealplanningkeys.ValidMeasurementUnitIDKey: existingValidMeasurementUnit.ID,
	}))

	existingValidMeasurementUnit, err = m.db.GetValidMeasurementUnit(ctx, validMeasurementUnitID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid measurement unit")
	}

	return existingValidMeasurementUnit, nil
}

func (m *mealPlanningManager) ArchiveValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)

	if err := m.db.ArchiveValidMeasurementUnit(ctx, validMeasurementUnitID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid measurement unit")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.ValidMeasurementUnitArchivedServiceEventType, map[string]any{
		mealplanningkeys.ValidMeasurementUnitIDKey: validMeasurementUnitID,
	}))

	return nil
}
