package mealplanning

import (
	"context"
	"database/sql"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v10/database"
	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/observability"
	platformkeys "github.com/primandproper/platform-go/v10/observability/keys"
	"github.com/primandproper/platform-go/v10/observability/tracing"
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

	results, err := q.generatedQuerier.SearchForValidInstruments(ctx, q.readDB, &generated.SearchForValidInstrumentsParams{
		NameQuery:       query,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid instruments list retrieval query")
	}

	var (
		data                      = []*types.ValidInstrument{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
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
		data = append(data, validInstrument)

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(data, filteredCount, totalCount, func(vi *types.ValidInstrument) string { return vi.ID }, filter)

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

	results, err := q.generatedQuerier.SearchForValidInstrumentsNotOwnedByAccount(ctx, q.readDB, &generated.SearchForValidInstrumentsNotOwnedByAccountParams{
		AccountID:       accountID,
		NameQuery:       query,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid instruments not owned by account search query")
	}

	var (
		data                      = []*types.ValidInstrument{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
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
		data = append(data, validInstrument)

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(data, filteredCount, totalCount, func(vi *types.ValidInstrument) string { return vi.ID }, filter)

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

	results, err := q.generatedQuerier.GetValidInstruments(ctx, q.readDB, &generated.GetValidInstrumentsParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid instruments list retrieval query")
	}

	var (
		data          []*types.ValidInstrument
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &types.ValidInstrument{
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

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
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
		Cursor:      after,
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
	}, func(tx database.SQLQueryExecutor) error {
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
	}, events.WithIndexUpsert(mealplanningindexing.IndexTypeValidInstruments, input.ID)); err != nil {
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
	}, func(tx database.SQLQueryExecutor) error {
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
	}, events.WithIndexUpsert(mealplanningindexing.IndexTypeValidInstruments, updated.ID)); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid instrument")
	}

	logger.Info("valid instrument updated")

	return nil
}

// MarkValidInstrumentAsIndexed updates a particular valid instrument's last_indexed_at value.
func (q *repository) MarkValidInstrumentAsIndexed(ctx context.Context, validInstrumentID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validInstrumentID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidInstrumentIDKey, validInstrumentID)

	if _, err := q.generatedQuerier.UpdateValidInstrumentLastIndexedAt(ctx, q.writeDB, validInstrumentID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking valid instrument as indexed")
	}

	logger.Info("valid instrument marked as indexed")

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
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveValidInstrument(ctx, tx, validInstrumentID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving valid instrument")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	}, events.WithIndexDelete(mealplanningindexing.IndexTypeValidInstruments, validInstrumentID))
}
