package managers

import (
	"context"
	"fmt"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	eatingindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/observability"
	platformkeys "github.com/primandproper/platform-go/v10/observability/keys"
	"github.com/primandproper/platform-go/v10/observability/tracing"
	searchpagination "github.com/primandproper/platform-go/v10/search/pagination"
)

func (m *mealPlanningManager) SearchValidIngredients(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredient], error) {
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
		results *filtering.QueryFilteredResult[types.ValidIngredient]
		err     error
	)
	if !useSearchService {
		var rawResults *filtering.QueryFilteredResult[types.ValidIngredient]
		rawResults, err = m.db.SearchForValidIngredients(ctx, query, filter)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "searching database for valid ingredients")
		}

		results = rawResults
	} else {
		results, err = searchpagination.Hydrated(ctx, m.validIngredientSearchIndex, query, filter,
			func(subset *eatingindexing.ValidIngredientSearchSubset) string { return subset.ID },
			m.db.GetValidIngredientsWithIDs,
		)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "searching valid ingredient search index for valid ingredients")
		}
	}
	for _, ing := range results.Data {
		if err = m.enrichValidIngredientWithMedia(ctx, ing); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "enriching valid ingredient with media")
		}
	}
	return results, nil
}

func (m *mealPlanningManager) ListValidIngredients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredient], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidIngredients(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing valid ingredients")
	}
	for _, ing := range results.Data {
		if err = m.enrichValidIngredientWithMedia(ctx, ing); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "enriching valid ingredient with media")
		}
	}
	return results, nil
}

func (m *mealPlanningManager) CreateValidIngredient(ctx context.Context, input *types.ValidIngredientCreationRequestInput) (*types.ValidIngredient, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidIngredientCreationRequestInputToValidIngredientDatabaseCreationInput(input)
	created, err := m.db.CreateValidIngredient(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid ingredient")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadValidIngredient(ctx context.Context, validIngredientID string) (*types.ValidIngredient, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	result, err := m.db.GetValidIngredient(ctx, validIngredientID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient")
	}
	if err = m.enrichValidIngredientWithMedia(ctx, result); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "enriching valid ingredient with media")
	}
	return result, nil
}

func (m *mealPlanningManager) RandomValidIngredient(ctx context.Context) (*types.ValidIngredient, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	result, err := m.db.GetRandomValidIngredient(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching random valid ingredient")
	}
	if err = m.enrichValidIngredientWithMedia(ctx, result); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "enriching valid ingredient with media")
	}
	return result, nil
}

func (m *mealPlanningManager) UpdateValidIngredient(ctx context.Context, validIngredientID string, input *types.ValidIngredientUpdateRequestInput) (*types.ValidIngredient, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, fmt.Errorf("validating update input: %w", err)
	}

	existingValidIngredient, err := m.db.GetValidIngredient(ctx, validIngredientID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient")
	}

	existingValidIngredient.Update(input)
	if err = m.db.UpdateValidIngredient(ctx, existingValidIngredient); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid ingredient")
	}

	existingValidIngredient, err = m.db.GetValidIngredient(ctx, validIngredientID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid ingredient")
	}
	if err = m.enrichValidIngredientWithMedia(ctx, existingValidIngredient); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "enriching valid ingredient with media")
	}
	return existingValidIngredient, nil
}

func (m *mealPlanningManager) ArchiveValidIngredient(ctx context.Context, validIngredientID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	if err := m.db.ArchiveValidIngredient(ctx, validIngredientID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid ingredient")
	}

	return nil
}

func (m *mealPlanningManager) AddIngredientMedia(ctx context.Context, validIngredientID, uploadedMediaID string, index int32) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	if err := m.db.AddIngredientMedia(ctx, validIngredientID, uploadedMediaID, index); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "adding ingredient media")
	}

	return nil
}

func (m *mealPlanningManager) SearchValidIngredientsByPreparationAndIngredientName(ctx context.Context, validPreparationID, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredient], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span).WithValue(platformkeys.SearchQueryKey, query).WithValue(mealplanningkeys.ValidPreparationIDKey, validPreparationID)
	tracing.AttachToSpan(span, platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, validPreparationID)

	validIngredients, err := m.db.SearchForValidIngredientsForPreparation(ctx, validPreparationID, query, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "searching for valid ingredient preparations")
	}
	for _, ing := range validIngredients.Data {
		if err = m.enrichValidIngredientWithMedia(ctx, ing); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "enriching valid ingredient with media")
		}
	}
	return validIngredients, nil
}

// enrichValidIngredientWithMedia loads and attaches media to a valid ingredient.
func (m *mealPlanningManager) enrichValidIngredientWithMedia(ctx context.Context, ing *types.ValidIngredient) error {
	if ing == nil {
		return nil
	}
	rows, err := m.db.GetIngredientMediaByIngredient(ctx, ing.ID)
	if err != nil || len(rows) == 0 {
		return err
	}
	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.UploadedMediaID
	}
	mediaList, err := m.db.GetUploadedMediaWithIDs(ctx, ids)
	if err != nil {
		return err
	}
	mediaByID := make(map[string]*uploadedmedia.UploadedMedia)
	for _, um := range mediaList {
		mediaByID[um.ID] = um
	}
	ing.Media = make([]*uploadedmedia.UploadedMedia, 0, len(rows))
	for _, r := range rows {
		if um := mediaByID[r.UploadedMediaID]; um != nil {
			ing.Media = append(ing.Media, um)
		}
	}
	return nil
}
