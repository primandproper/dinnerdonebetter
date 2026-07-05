package mealplanning

import (
	"context"
	"database/sql"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v3/database"
	"github.com/primandproper/platform-go/v3/database/filtering"
	platformerrors "github.com/primandproper/platform-go/v3/errors"
	"github.com/primandproper/platform-go/v3/observability"
	platformkeys "github.com/primandproper/platform-go/v3/observability/keys"
	"github.com/primandproper/platform-go/v3/observability/tracing"
)

var (
	_ types.ValidIngredientStateDataManager = (*repository)(nil)
)

// ValidIngredientStateExists fetches whether a valid ingredient state exists from the database.
func (q *repository) ValidIngredientStateExists(ctx context.Context, validIngredientStateID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientStateID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)

	result, err := q.generatedQuerier.CheckValidIngredientStateExistence(ctx, q.readDB, validIngredientStateID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient state existence check")
	}

	return result, nil
}

// GetValidIngredientState fetches a valid ingredient state from the database.
func (q *repository) GetValidIngredientState(ctx context.Context, validIngredientStateID string) (*types.ValidIngredientState, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientStateID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)

	result, err := q.generatedQuerier.GetValidIngredientState(ctx, q.readDB, validIngredientStateID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredient state retrieval query")
	}

	validIngredientState := &types.ValidIngredientState{
		CreatedAt:     result.CreatedAt,
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		PastTense:     result.PastTense,
		Description:   result.Description,
		IconPath:      result.IconPath,
		ID:            result.ID,
		Name:          result.Name,
		AttributeType: string(result.AttributeType),
		Slug:          result.Slug,
	}

	return validIngredientState, nil
}

// SearchForValidIngredientStates fetches a valid ingredient state from the database.
func (q *repository) SearchForValidIngredientStates(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientState], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if query == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, query)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.SearchForValidIngredientStates(ctx, q.readDB, &generated.SearchForValidIngredientStatesParams{
		NameQuery:       query,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredient states list retrieval query")
	}

	var (
		data                      []*types.ValidIngredientState
		filteredCount, totalCount uint64
	)

	for _, result := range results {
		data = append(data, &types.ValidIngredientState{
			CreatedAt:     result.CreatedAt,
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			PastTense:     result.PastTense,
			Description:   result.Description,
			IconPath:      result.IconPath,
			ID:            result.ID,
			Name:          result.Name,
			AttributeType: string(result.AttributeType),
			Slug:          result.Slug,
		})
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(data, filteredCount, totalCount, func(vis *types.ValidIngredientState) string { return vis.ID }, filter)

	return x, nil
}

// GetValidIngredientStates fetches a list of valid ingredient states from the database that meet a particular filter.
func (q *repository) GetValidIngredientStates(ctx context.Context, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.ValidIngredientState], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidIngredientStates(ctx, q.readDB, &generated.GetValidIngredientStatesParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredient states list retrieval query")
	}

	var (
		data          []*types.ValidIngredientState
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &types.ValidIngredientState{
			CreatedAt:     result.CreatedAt,
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			PastTense:     result.PastTense,
			Description:   result.Description,
			IconPath:      result.IconPath,
			ID:            result.ID,
			Name:          result.Name,
			AttributeType: string(result.AttributeType),
			Slug:          result.Slug,
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vis *types.ValidIngredientState) string { return vis.ID },
		filter,
	)

	return x, nil
}

// GetValidIngredientStatesWithIDs fetches a list of valid ingredientStates from the database that meet a particular filter.
func (q *repository) GetValidIngredientStatesWithIDs(ctx context.Context, ids []string) ([]*types.ValidIngredientState, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	results, err := q.generatedQuerier.GetValidIngredientStatesWithIDs(ctx, q.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredient states id list retrieval query")
	}

	ingredientStates := []*types.ValidIngredientState{}
	for _, result := range results {
		ingredientStates = append(ingredientStates, &types.ValidIngredientState{
			CreatedAt:     result.CreatedAt,
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			PastTense:     result.PastTense,
			Description:   result.Description,
			IconPath:      result.IconPath,
			ID:            result.ID,
			Name:          result.Name,
			AttributeType: string(result.AttributeType),
			Slug:          result.Slug,
		})
	}

	return ingredientStates, nil
}

// GetValidIngredientStateIDsThatNeedSearchIndexing fetches a list of valid ingredientStates from the database that meet a particular filter.
func (q *repository) GetValidIngredientStateIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	results, err := q.generatedQuerier.GetValidIngredientStatesNeedingIndexing(ctx, q.readDB)
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing valid ingredient states list retrieval query")
	}

	return results, nil
}

// CreateValidIngredientState creates a valid ingredient state in the database.
func (q *repository) CreateValidIngredientState(ctx context.Context, input *types.ValidIngredientStateDatabaseCreationInput) (*types.ValidIngredientState, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientStateIDKey, input.ID)

	// create the valid ingredient state.
	if err := q.generatedQuerier.CreateValidIngredientState(ctx, q.writeDB, &generated.CreateValidIngredientStateParams{
		ID:            input.ID,
		Name:          input.Name,
		Description:   input.Description,
		IconPath:      input.IconPath,
		PastTense:     input.PastTense,
		Slug:          input.Slug,
		AttributeType: generated.IngredientAttributeType(input.AttributeType),
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient state creation query")
	}

	x := &types.ValidIngredientState{
		ID:            input.ID,
		Name:          input.Name,
		Description:   input.Description,
		IconPath:      input.IconPath,
		Slug:          input.Slug,
		PastTense:     input.PastTense,
		AttributeType: input.AttributeType,
		CreatedAt:     q.CurrentTime(),
	}

	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, x.ID)
	logger.Info("valid ingredient state created")

	return x, nil
}

// UpdateValidIngredientState updates a particular valid ingredient state.
func (q *repository) UpdateValidIngredientState(ctx context.Context, updated *types.ValidIngredientState) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientStateIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, updated.ID)

	if _, err := q.generatedQuerier.UpdateValidIngredientState(ctx, q.writeDB, &generated.UpdateValidIngredientStateParams{
		Name:          updated.Name,
		Description:   updated.Description,
		IconPath:      updated.IconPath,
		Slug:          updated.Slug,
		PastTense:     updated.PastTense,
		AttributeType: generated.IngredientAttributeType(updated.AttributeType),
		ID:            updated.ID,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid ingredient state")
	}

	logger.Info("valid ingredient state updated")

	return nil
}

// MarkValidIngredientStateAsIndexed updates a particular valid ingredient state's last_indexed_at value.
func (q *repository) MarkValidIngredientStateAsIndexed(ctx context.Context, validIngredientStateID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientStateID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)

	if _, err := q.generatedQuerier.UpdateValidIngredientStateLastIndexedAt(ctx, q.writeDB, validIngredientStateID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking valid ingredient state as indexed")
	}

	logger.Info("valid ingredient state marked as indexed")

	return nil
}

// ArchiveValidIngredientState archives a valid ingredient state from the database by its ID.
func (q *repository) ArchiveValidIngredientState(ctx context.Context, validIngredientStateID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientStateID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, validIngredientStateID)

	rowsAffected, err := q.generatedQuerier.ArchiveValidIngredientState(ctx, q.writeDB, validIngredientStateID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid ingredient state")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
