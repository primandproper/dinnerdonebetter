package mealplanning

import (
	"context"
	"database/sql"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/observability"
	platformkeys "github.com/primandproper/platform-go/v13/observability/keys"
	"github.com/primandproper/platform-go/v13/observability/tracing"
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

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.SearchForValidIngredientStates(ctx, q.readDB, &generated.SearchForValidIngredientStatesParams{
		NameQuery:       query,
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
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

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetValidIngredientStates(ctx, q.readDB, &generated.GetValidIngredientStatesParams{
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
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

// ScanValidIngredientStateIDsForReindex returns up to limit IDs sorting strictly after `after`, in ascending byte order.
//
// It is the source half of a search reindex: searchsync.Reindexer walks this to find every
// document that should exist, and prunes the index of anything it does not name. It replaces
// the "IDs that need indexing" sampler platform-go v10 removed, which asked a different and
// weaker question — which rows look stale — and could only ever be probabilistically right,
// because a row the sampler had not reached was a row the index was wrong about.
func (q *repository) ScanValidIngredientStateIDsForReindex(ctx context.Context, after string, limit int) ([]string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	results, err := q.generatedQuerier.ScanValidIngredientStateIDsForReindex(ctx, q.readDB, &generated.ScanValidIngredientStateIDsForReindexParams{
		PageCursor:  after,
		ResultLimit: limit,
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing valid ingredient states reindex scan query")
	}

	return results, nil
}

// CreateValidIngredientState creates a valid ingredient state in the database.
func (q *repository) CreateValidIngredientState(ctx context.Context, input *types.ValidIngredientStateDatabaseCreationInput) (*types.ValidIngredientState, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientStateIDKey, input.ID)

	// create the valid ingredient state.
	if err := q.withEvent(ctx, logger, types.ValidIngredientStateCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientStateIDKey: input.ID,
	}, func(tx database.Tx) error {
		return q.generatedQuerier.CreateValidIngredientState(ctx, tx, &generated.CreateValidIngredientStateParams{
			ID:            input.ID,
			Name:          input.Name,
			Description:   input.Description,
			IconPath:      input.IconPath,
			PastTense:     input.PastTense,
			Slug:          input.Slug,
			AttributeType: generated.IngredientAttributeType(input.AttributeType),
		})
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
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientStateIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, types.ValidIngredientStateUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientStateIDKey: updated.ID,
	}, func(tx database.Tx) error {
		_, updateErr := q.generatedQuerier.UpdateValidIngredientState(ctx, tx, &generated.UpdateValidIngredientStateParams{
			Name:          updated.Name,
			Description:   updated.Description,
			IconPath:      updated.IconPath,
			Slug:          updated.Slug,
			PastTense:     updated.PastTense,
			AttributeType: generated.IngredientAttributeType(updated.AttributeType),
			ID:            updated.ID,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid ingredient state")
	}

	logger.Info("valid ingredient state updated")

	return nil
}

// MarkValidIngredientStatesAsIndexed stamps last_indexed_at on the rows behind the documents an index has taken.
//
// It is the write half of search/sync's Stamper: the ids arrive already coalesced and ordered
// by the batching.Buffer the syncer stamps through, so this is one statement per flush rather
// than one per document. It is deliberately not guarded on archived_at — the syncer applies a
// vanished row as a delete and stamps nothing, and a row archived between apply and flush is
// harmless to stamp.
func (q *repository) MarkValidIngredientStatesAsIndexed(ctx context.Context, ids []string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if len(ids) == 0 {
		return nil
	}

	logger := q.logger.Clone().WithValue("id_count", len(ids))
	tracing.AttachToSpan(span, "id_count", len(ids))

	if _, err := q.generatedQuerier.MarkValidIngredientStatesAsIndexed(ctx, q.writeDB, ids); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking valid ingredient states as indexed")
	}

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

	return q.withEvent(ctx, logger, types.ValidIngredientStateArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientStateIDKey: validIngredientStateID,
	}, func(tx database.Tx) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveValidIngredientState(ctx, tx, validIngredientStateID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving valid ingredient state")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}
