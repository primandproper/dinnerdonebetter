package managers

import (
	"context"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	eatingindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/observability"
	platformkeys "github.com/primandproper/platform-go/v9/observability/keys"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	textsearch "github.com/primandproper/platform-go/v9/search/text"
)

func (m *mealPlanningManager) SearchValidPreparations(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparation], error) {
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
		results *filtering.QueryFilteredResult[types.ValidPreparation]
		err     error
	)
	if !useSearchService {
		results, err = m.db.SearchForValidPreparations(ctx, query, filter)
	} else {
		var indexHits *textsearch.SearchResults[eatingindexing.ValidPreparationSearchSubset]
		indexHits, err = m.validPreparationsSearchIndex.Search(ctx, textsearch.SearchRequest{Query: query})
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "searching valid preparations")
		}

		validPreparationSubsets := indexHits.Hits

		ids := []string{}
		for _, validPreparationSubset := range validPreparationSubsets {
			ids = append(ids, validPreparationSubset.ID)
		}

		searchResults, searchErr := m.db.GetValidPreparationsWithIDs(ctx, ids)
		if searchErr != nil {
			return nil, observability.PrepareAndLogError(searchErr, logger, span, "fetching valid preparations from database")
		}

		results = filtering.NewQueryFilteredResult(searchResults, uint64(len(searchResults)), uint64(len(searchResults)), func(v *types.ValidPreparation) string {
			return v.ID
		}, filter)
	}

	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "searching valid preparations")
	}
	for _, prep := range results.Data {
		if err = m.enrichValidPreparationWithMedia(ctx, prep); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "enriching valid preparation with media")
		}
	}
	return results, nil
}

func (m *mealPlanningManager) ListValidPreparations(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparation], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetValidPreparations(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing valid preparations")
	}
	for _, prep := range results.Data {
		if err = m.enrichValidPreparationWithMedia(ctx, prep); err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "enriching valid preparation with media")
		}
	}
	return results, nil
}

func (m *mealPlanningManager) CreateValidPreparation(ctx context.Context, input *types.ValidPreparationCreationRequestInput) (*types.ValidPreparation, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertValidPreparationCreationRequestInputToValidPreparationDatabaseCreationInput(input)
	created, err := m.db.CreateValidPreparation(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating valid preparation")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadValidPreparation(ctx context.Context, validPreparationID string) (*types.ValidPreparation, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, validPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, validPreparationID)

	result, err := m.db.GetValidPreparation(ctx, validPreparationID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation")
	}
	if err = m.enrichValidPreparationWithMedia(ctx, result); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "enriching valid preparation with media")
	}
	return result, nil
}

func (m *mealPlanningManager) RandomValidPreparation(ctx context.Context) (*types.ValidPreparation, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	result, err := m.db.GetRandomValidPreparation(ctx)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching random valid preparation")
	}
	if err = m.enrichValidPreparationWithMedia(ctx, result); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "enriching valid preparation with media")
	}
	return result, nil
}

func (m *mealPlanningManager) UpdateValidPreparation(ctx context.Context, validPreparationID string, input *types.ValidPreparationUpdateRequestInput) (*types.ValidPreparation, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	existingValidPreparation, err := m.db.GetValidPreparation(ctx, validPreparationID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation")
	}

	existingValidPreparation.Update(input)
	if err = m.db.UpdateValidPreparation(ctx, existingValidPreparation); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "updating valid preparation")
	}

	existingValidPreparation, err = m.db.GetValidPreparation(ctx, validPreparationID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching updated valid preparation")
	}
	if err = m.enrichValidPreparationWithMedia(ctx, existingValidPreparation); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "enriching valid preparation with media")
	}
	return existingValidPreparation, nil
}

func (m *mealPlanningManager) ArchiveValidPreparation(ctx context.Context, validPreparationID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, validPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, validPreparationID)

	if err := m.db.ArchiveValidPreparation(ctx, validPreparationID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid preparation")
	}

	return nil
}

func (m *mealPlanningManager) AddPreparationMedia(ctx context.Context, validPreparationID string, forIngredientID *string, uploadedMediaID string, index int32) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.ValidPreparationIDKey, validPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, validPreparationID)

	if err := m.db.AddPreparationMedia(ctx, validPreparationID, forIngredientID, uploadedMediaID, index); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "adding preparation media")
	}

	return nil
}

// enrichValidPreparationWithMedia loads and attaches media to a valid preparation.
func (m *mealPlanningManager) enrichValidPreparationWithMedia(ctx context.Context, prep *types.ValidPreparation) error {
	if prep == nil {
		return nil
	}
	rows, err := m.db.GetPreparationMediaByPreparation(ctx, prep.ID)
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
	prep.Media = make([]*uploadedmedia.UploadedMedia, 0, len(rows))
	for _, r := range rows {
		if um := mediaByID[r.UploadedMediaID]; um != nil {
			prep.Media = append(prep.Media, um)
		}
	}
	return nil
}
