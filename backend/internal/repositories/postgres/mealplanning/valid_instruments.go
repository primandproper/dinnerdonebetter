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
	_ types.ValidInstrumentDataManager = (*repository)(nil)
)

// ValidInstrumentExists fetches whether a valid instrument exists from the database.
func (q *repository) ValidInstrumentExists(ctx context.Context, validInstrumentID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validInstrumentID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)

	result, err := q.generatedQuerier.CheckValidInstrumentExistence(ctx, q.readDB, validInstrumentID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing valid instrument existence check")
	}

	return result, nil
}

// GetValidInstrument fetches a valid instrument from the database.
func (q *repository) GetValidInstrument(ctx context.Context, validInstrumentID string) (*types.ValidInstrument, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validInstrumentID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)

	result, err := q.generatedQuerier.GetValidInstrument(ctx, q.readDB, validInstrumentID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting valid instrument")
	}

	validInstrument := &types.ValidInstrument{
		CreatedAt:                      result.CreatedAt,
		LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
		IconPath:                       result.IconPath,
		ID:                             result.ID,
		Name:                           result.Name,
		PluralName:                     result.PluralName,
		Description:                    result.Description,
		Slug:                           result.Slug,
		DisplayInSummaryLists:          result.DisplayInSummaryLists,
		IncludeInGeneratedInstructions: result.IncludeInGeneratedInstructions,
		UsableForStorage:               result.UsableForStorage,
	}

	return validInstrument, nil
}

// GetRandomValidInstrument fetches a valid instrument from the database.
func (q *repository) GetRandomValidInstrument(ctx context.Context) (*types.ValidInstrument, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	result, err := q.generatedQuerier.GetRandomValidInstrument(ctx, q.readDB)
	if err != nil {
		return nil, observability.PrepareError(err, span, "scanning validInstrument")
	}

	validInstrument := &types.ValidInstrument{
		CreatedAt:                      result.CreatedAt,
		LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
		IconPath:                       result.IconPath,
		ID:                             result.ID,
		Name:                           result.Name,
		PluralName:                     result.PluralName,
		Description:                    result.Description,
		Slug:                           result.Slug,
		DisplayInSummaryLists:          result.DisplayInSummaryLists,
		IncludeInGeneratedInstructions: result.IncludeInGeneratedInstructions,
		UsableForStorage:               result.UsableForStorage,
	}

	return validInstrument, nil
}

// SearchForValidInstruments fetches a valid instrument from the database.
func (q *repository) SearchForValidInstruments(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidInstrument], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if query == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, query)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.SearchForValidInstruments(ctx, q.readDB, &generated.SearchForValidInstrumentsParams{
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
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid instruments list retrieval query")
	}

	x := filtering.Drain(
		results,
		func(result *generated.SearchForValidInstrumentsRow) *types.ValidInstrument {
			return &types.ValidInstrument{
				CreatedAt:                      result.CreatedAt,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
				IconPath:                       result.IconPath,
				ID:                             result.ID,
				Name:                           result.Name,
				PluralName:                     result.PluralName,
				Description:                    result.Description,
				Slug:                           result.Slug,
				DisplayInSummaryLists:          result.DisplayInSummaryLists,
				IncludeInGeneratedInstructions: result.IncludeInGeneratedInstructions,
				UsableForStorage:               result.UsableForStorage,
			}
		},
		func(result *generated.SearchForValidInstrumentsRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(vi *types.ValidInstrument) string { return vi.ID },
		filter,
	)

	return x, nil
}

// SearchForValidInstrumentsNotOwnedByAccount fetches valid instruments not owned by the account.
func (q *repository) SearchForValidInstrumentsNotOwnedByAccount(ctx context.Context, accountID, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidInstrument], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.SearchForValidInstrumentsNotOwnedByAccount(ctx, q.readDB, &generated.SearchForValidInstrumentsNotOwnedByAccountParams{
		AccountID:       accountID,
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
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid instruments not owned by account search query")
	}

	x := filtering.Drain(
		results,
		func(result *generated.SearchForValidInstrumentsNotOwnedByAccountRow) *types.ValidInstrument {
			return &types.ValidInstrument{
				CreatedAt:                      result.CreatedAt,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
				IconPath:                       result.IconPath,
				ID:                             result.ID,
				Name:                           result.Name,
				PluralName:                     result.PluralName,
				Description:                    result.Description,
				Slug:                           result.Slug,
				DisplayInSummaryLists:          result.DisplayInSummaryLists,
				IncludeInGeneratedInstructions: result.IncludeInGeneratedInstructions,
				UsableForStorage:               result.UsableForStorage,
			}
		},
		func(result *generated.SearchForValidInstrumentsNotOwnedByAccountRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(vi *types.ValidInstrument) string { return vi.ID },
		filter,
	)

	return x, nil
}

// GetValidInstruments fetches a list of valid instruments from the database that meet a particular filter.
func (q *repository) GetValidInstruments(ctx context.Context, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.ValidInstrument], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetValidInstruments(ctx, q.readDB, &generated.GetValidInstrumentsParams{
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid instruments list retrieval query")
	}

	x = filtering.Drain(
		results,
		func(result *generated.GetValidInstrumentsRow) *types.ValidInstrument {
			return &types.ValidInstrument{
				CreatedAt:                      result.CreatedAt,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
				IconPath:                       result.IconPath,
				ID:                             result.ID,
				Name:                           result.Name,
				PluralName:                     result.PluralName,
				Description:                    result.Description,
				Slug:                           result.Slug,
				DisplayInSummaryLists:          result.DisplayInSummaryLists,
				IncludeInGeneratedInstructions: result.IncludeInGeneratedInstructions,
				UsableForStorage:               result.UsableForStorage,
			}
		},
		func(result *generated.GetValidInstrumentsRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(vi *types.ValidInstrument) string { return vi.ID },
		filter,
	)

	return x, nil
}

// GetValidInstrumentsWithIDs fetches a list of valid instruments from the database that meet a particular filter.
func (q *repository) GetValidInstrumentsWithIDs(ctx context.Context, ids []string) ([]*types.ValidInstrument, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	results, err := q.generatedQuerier.GetValidInstrumentsWithIDs(ctx, q.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid instruments id list retrieval query")
	}

	instruments := []*types.ValidInstrument{}
	for _, result := range results {
		instruments = append(instruments, &types.ValidInstrument{
			CreatedAt:                      result.CreatedAt,
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
			IconPath:                       result.IconPath,
			ID:                             result.ID,
			Name:                           result.Name,
			PluralName:                     result.PluralName,
			Description:                    result.Description,
			Slug:                           result.Slug,
			DisplayInSummaryLists:          result.DisplayInSummaryLists,
			IncludeInGeneratedInstructions: result.IncludeInGeneratedInstructions,
			UsableForStorage:               result.UsableForStorage,
		})
	}

	return instruments, nil
}

// ScanValidInstrumentIDsForReindex returns up to limit IDs sorting strictly after `after`, in ascending byte order.
//
// It is the source half of a search reindex: searchsync.Reindexer walks this to find every
// document that should exist, and prunes the index of anything it does not name. It replaces
// the "IDs that need indexing" sampler platform-go v10 removed, which asked a different and
// weaker question — which rows look stale — and could only ever be probabilistically right,
// because a row the sampler had not reached was a row the index was wrong about.
func (q *repository) ScanValidInstrumentIDsForReindex(ctx context.Context, after string, limit int) ([]string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	results, err := q.generatedQuerier.ScanValidInstrumentIDsForReindex(ctx, q.readDB, &generated.ScanValidInstrumentIDsForReindexParams{
		PageCursor:  after,
		ResultLimit: limit,
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing valid instruments reindex scan query")
	}

	return results, nil
}

// CreateValidInstrument creates a valid instrument in the database.
func (q *repository) CreateValidInstrument(ctx context.Context, input *types.ValidInstrumentDatabaseCreationInput) (*types.ValidInstrument, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.ValidInstrumentIDKey, input.ID)

	// create the valid instrument.
	if err := q.withEvent(ctx, logger, types.ValidInstrumentCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidInstrumentIDKey: input.ID,
	}, func(tx database.Tx) error {
		return q.generatedQuerier.CreateValidInstrument(ctx, tx, &generated.CreateValidInstrumentParams{
			ID:                             input.ID,
			Name:                           input.Name,
			PluralName:                     input.PluralName,
			Description:                    input.Description,
			IconPath:                       input.IconPath,
			Slug:                           input.Slug,
			UsableForStorage:               input.UsableForStorage,
			DisplayInSummaryLists:          input.DisplayInSummaryLists,
			IncludeInGeneratedInstructions: input.IncludeInGeneratedInstructions,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid instrument creation query")
	}

	x := &types.ValidInstrument{
		ID:                             input.ID,
		Name:                           input.Name,
		PluralName:                     input.PluralName,
		Description:                    input.Description,
		IconPath:                       input.IconPath,
		UsableForStorage:               input.UsableForStorage,
		Slug:                           input.Slug,
		DisplayInSummaryLists:          input.DisplayInSummaryLists,
		IncludeInGeneratedInstructions: input.IncludeInGeneratedInstructions,
		CreatedAt:                      q.CurrentTime(),
	}

	logger.Info("valid instrument created")

	return x, nil
}

// UpdateValidInstrument updates a particular valid instrument.
func (q *repository) UpdateValidInstrument(ctx context.Context, updated *types.ValidInstrument) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidInstrumentIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, types.ValidInstrumentUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidInstrumentIDKey: updated.ID,
	}, func(tx database.Tx) error {
		_, updateErr := q.generatedQuerier.UpdateValidInstrument(ctx, tx, &generated.UpdateValidInstrumentParams{
			Name:                           updated.Name,
			PluralName:                     updated.PluralName,
			Description:                    updated.Description,
			IconPath:                       updated.IconPath,
			Slug:                           updated.Slug,
			ID:                             updated.ID,
			UsableForStorage:               updated.UsableForStorage,
			DisplayInSummaryLists:          updated.DisplayInSummaryLists,
			IncludeInGeneratedInstructions: updated.IncludeInGeneratedInstructions,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid instrument")
	}

	logger.Info("valid instrument updated")

	return nil
}

// MarkValidInstrumentsAsIndexed stamps last_indexed_at on the rows behind the documents an index has taken.
//
// It is the write half of search/sync's Stamper: the ids arrive already coalesced and ordered
// by the batching.Buffer the syncer stamps through, so this is one statement per flush rather
// than one per document. It is deliberately not guarded on archived_at — the syncer applies a
// vanished row as a delete and stamps nothing, and a row archived between apply and flush is
// harmless to stamp.
func (q *repository) MarkValidInstrumentsAsIndexed(ctx context.Context, ids []string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if len(ids) == 0 {
		return nil
	}

	logger := q.logger.Clone().WithValue("id_count", len(ids))
	tracing.AttachToSpan(span, "id_count", len(ids))

	if _, err := q.generatedQuerier.MarkValidInstrumentsAsIndexed(ctx, q.writeDB, ids); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking valid instruments as indexed")
	}

	return nil
}

// ArchiveValidInstrument archives a valid instrument from the database by its ID.
func (q *repository) ArchiveValidInstrument(ctx context.Context, validInstrumentID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validInstrumentID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)

	return q.withEvent(ctx, logger, types.ValidInstrumentArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidInstrumentIDKey: validInstrumentID,
	}, func(tx database.Tx) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveValidInstrument(ctx, tx, validInstrumentID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving valid instrument")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}
