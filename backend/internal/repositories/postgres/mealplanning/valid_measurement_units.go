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
	_ types.ValidMeasurementUnitDataManager = (*repository)(nil)
)

// ValidMeasurementUnitExists fetches whether a valid measurement unit exists from the database.
func (q *repository) ValidMeasurementUnitExists(ctx context.Context, validMeasurementUnitID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validMeasurementUnitID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)

	result, err := q.generatedQuerier.CheckValidMeasurementUnitExistence(ctx, q.readDB, validMeasurementUnitID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing valid measurement unit existence check")
	}

	return result, nil
}

// GetValidMeasurementUnit fetches a valid measurement unit from the database.
func (q *repository) GetValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string) (*types.ValidMeasurementUnit, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validMeasurementUnitID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)

	result, err := q.generatedQuerier.GetValidMeasurementUnit(ctx, q.readDB, validMeasurementUnitID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "scanning valid measurement unit")
	}

	validMeasurementUnit := &types.ValidMeasurementUnit{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		Name:          result.Name,
		IconPath:      result.IconPath,
		ID:            result.ID,
		Description:   result.Description,
		PluralName:    result.PluralName,
		Slug:          result.Slug,
		Volumetric:    database.BoolFromNullBool(result.Volumetric),
		Universal:     result.Universal,
		Metric:        result.Metric,
		Imperial:      result.Imperial,
	}

	return validMeasurementUnit, nil
}

// GetRandomValidMeasurementUnit fetches a valid measurement unit from the database.
func (q *repository) GetRandomValidMeasurementUnit(ctx context.Context) (*types.ValidMeasurementUnit, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	result, err := q.generatedQuerier.GetRandomValidMeasurementUnit(ctx, q.readDB)
	if err != nil {
		return nil, observability.PrepareError(err, span, "scanning valid measurement unit")
	}

	validMeasurementUnit := &types.ValidMeasurementUnit{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		Name:          result.Name,
		IconPath:      result.IconPath,
		ID:            result.ID,
		Description:   result.Description,
		PluralName:    result.PluralName,
		Slug:          result.Slug,
		Volumetric:    database.BoolFromNullBool(result.Volumetric),
		Universal:     result.Universal,
		Metric:        result.Metric,
		Imperial:      result.Imperial,
	}

	return validMeasurementUnit, nil
}

// SearchForValidMeasurementUnits fetches a valid measurement unit from the database.
func (q *repository) SearchForValidMeasurementUnits(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnit], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if query == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, query)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.SearchForValidMeasurementUnits(ctx, q.readDB, &generated.SearchForValidMeasurementUnitsParams{
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
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid measurement units list retrieval query")
	}

	x := filtering.Drain(
		results,
		func(result *generated.SearchForValidMeasurementUnitsRow) *types.ValidMeasurementUnit {
			return &types.ValidMeasurementUnit{
				CreatedAt:     result.CreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
				Name:          result.Name,
				IconPath:      result.IconPath,
				ID:            result.ID,
				Description:   result.Description,
				PluralName:    result.PluralName,
				Slug:          result.Slug,
				Volumetric:    database.BoolFromNullBool(result.Volumetric),
				Universal:     result.Universal,
				Metric:        result.Metric,
				Imperial:      result.Imperial,
			}
		},
		func(result *generated.SearchForValidMeasurementUnitsRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(vmu *types.ValidMeasurementUnit) string { return vmu.ID },
		filter,
	)

	return x, nil
}

// ValidMeasurementUnitsForIngredientID fetches a valid measurement unit from the database.
func (q *repository) ValidMeasurementUnitsForIngredientID(ctx context.Context, validIngredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnit], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	if validIngredientID == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.SearchValidMeasurementUnitsByIngredientID(ctx, q.readDB, &generated.SearchValidMeasurementUnitsByIngredientIDParams{
		CreatedBefore:     filterArgs.CreatedBefore,
		CreatedAfter:      filterArgs.CreatedAfter,
		UpdatedBefore:     filterArgs.UpdatedBefore,
		UpdatedAfter:      filterArgs.UpdatedAfter,
		PageCursor:        filterArgs.Cursor,
		ResultLimit:       filterArgs.ResultLimit,
		IncludeArchived:   filterArgs.IncludeArchived,
		ValidIngredientID: validIngredientID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid measurement units list retrieval query")
	}

	x := filtering.Drain(
		results,
		func(result *generated.SearchValidMeasurementUnitsByIngredientIDRow) *types.ValidMeasurementUnit {
			return &types.ValidMeasurementUnit{
				CreatedAt:     result.CreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
				Name:          result.Name,
				IconPath:      result.IconPath,
				ID:            result.ID,
				Description:   result.Description,
				PluralName:    result.PluralName,
				Slug:          result.Slug,
				Volumetric:    database.BoolFromNullBool(result.Volumetric),
				Universal:     result.Universal,
				Metric:        result.Metric,
				Imperial:      result.Imperial,
			}
		},
		func(result *generated.SearchValidMeasurementUnitsByIngredientIDRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(vmu *types.ValidMeasurementUnit) string { return vmu.ID },
		filter,
	)

	return x, nil
}

// GetValidMeasurementUnits fetches a list of valid measurement units from the database that meet a particular filter.
func (q *repository) GetValidMeasurementUnits(ctx context.Context, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.ValidMeasurementUnit], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetValidMeasurementUnits(ctx, q.readDB, &generated.GetValidMeasurementUnitsParams{
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid measurement units list retrieval query")
	}

	x = filtering.Drain(
		results,
		func(result *generated.GetValidMeasurementUnitsRow) *types.ValidMeasurementUnit {
			return &types.ValidMeasurementUnit{
				CreatedAt:     result.CreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
				Name:          result.Name,
				IconPath:      result.IconPath,
				ID:            result.ID,
				Description:   result.Description,
				PluralName:    result.PluralName,
				Slug:          result.Slug,
				Volumetric:    database.BoolFromNullBool(result.Volumetric),
				Universal:     result.Universal,
				Metric:        result.Metric,
				Imperial:      result.Imperial,
			}
		},
		func(result *generated.GetValidMeasurementUnitsRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(vmu *types.ValidMeasurementUnit) string { return vmu.ID },
		filter,
	)

	return x, nil
}

// GetValidMeasurementUnitsWithIDs fetches a list of valid measurement unit from the database that meet a particular filter.
func (q *repository) GetValidMeasurementUnitsWithIDs(ctx context.Context, ids []string) ([]*types.ValidMeasurementUnit, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	results, err := q.generatedQuerier.GetValidMeasurementUnitsWithIDs(ctx, q.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid measurement unit id list retrieval query")
	}

	x := []*types.ValidMeasurementUnit{}
	for _, result := range results {
		x = append(x, &types.ValidMeasurementUnit{
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			Name:          result.Name,
			IconPath:      result.IconPath,
			ID:            result.ID,
			Description:   result.Description,
			PluralName:    result.PluralName,
			Slug:          result.Slug,
			Volumetric:    database.BoolFromNullBool(result.Volumetric),
			Universal:     result.Universal,
			Metric:        result.Metric,
			Imperial:      result.Imperial,
		})
	}

	return x, nil
}

// ScanValidMeasurementUnitIDsForReindex returns up to limit IDs sorting strictly after `after`, in ascending byte order.
//
// It is the source half of a search reindex: searchsync.Reindexer walks this to find every
// document that should exist, and prunes the index of anything it does not name. It replaces
// the "IDs that need indexing" sampler platform-go v10 removed, which asked a different and
// weaker question — which rows look stale — and could only ever be probabilistically right,
// because a row the sampler had not reached was a row the index was wrong about.
func (q *repository) ScanValidMeasurementUnitIDsForReindex(ctx context.Context, after string, limit int) ([]string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	results, err := q.generatedQuerier.ScanValidMeasurementUnitIDsForReindex(ctx, q.readDB, &generated.ScanValidMeasurementUnitIDsForReindexParams{
		PageCursor:  after,
		ResultLimit: limit,
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing valid measurement units reindex scan query")
	}

	return results, nil
}

// CreateValidMeasurementUnit creates a valid measurement unit in the database.
func (q *repository) CreateValidMeasurementUnit(ctx context.Context, input *types.ValidMeasurementUnitDatabaseCreationInput) (*types.ValidMeasurementUnit, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, input.ID)

	// create the valid measurement unit.
	if err := q.withEvent(ctx, logger, types.ValidMeasurementUnitCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidMeasurementUnitIDKey: input.ID,
	}, func(tx database.Tx) error {
		return q.generatedQuerier.CreateValidMeasurementUnit(ctx, tx, &generated.CreateValidMeasurementUnitParams{
			Name:        input.Name,
			Description: input.Description,
			IconPath:    input.IconPath,
			Slug:        input.Slug,
			PluralName:  input.PluralName,
			ID:          input.ID,
			Volumetric:  database.NullBoolFromBool(input.Volumetric),
			Universal:   input.Universal,
			Metric:      input.Metric,
			Imperial:    input.Imperial,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid measurement unit creation query")
	}

	x := &types.ValidMeasurementUnit{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
		Volumetric:  input.Volumetric,
		IconPath:    input.IconPath,
		Universal:   input.Universal,
		Metric:      input.Metric,
		Imperial:    input.Imperial,
		Slug:        input.Slug,
		PluralName:  input.PluralName,
		CreatedAt:   q.CurrentTime(),
	}

	logger.Info("valid measurement unit created")

	return x, nil
}

// UpdateValidMeasurementUnit updates a particular valid measurement unit.
func (q *repository) UpdateValidMeasurementUnit(ctx context.Context, updated *types.ValidMeasurementUnit) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, types.ValidMeasurementUnitUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidMeasurementUnitIDKey: updated.ID,
	}, func(tx database.Tx) error {
		_, updateErr := q.generatedQuerier.UpdateValidMeasurementUnit(ctx, tx, &generated.UpdateValidMeasurementUnitParams{
			Name:        updated.Name,
			Description: updated.Description,
			IconPath:    updated.IconPath,
			Slug:        updated.Slug,
			PluralName:  updated.PluralName,
			ID:          updated.ID,
			Volumetric:  database.NullBoolFromBool(updated.Volumetric),
			Universal:   updated.Universal,
			Metric:      updated.Metric,
			Imperial:    updated.Imperial,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid measurement unit")
	}

	logger.Info("valid measurement unit updated")

	return nil
}

// MarkValidMeasurementUnitsAsIndexed stamps last_indexed_at on the rows behind the documents an index has taken.
//
// It is the write half of search/sync's Stamper: the ids arrive already coalesced and ordered
// by the batching.Buffer the syncer stamps through, so this is one statement per flush rather
// than one per document. It is deliberately not guarded on archived_at — the syncer applies a
// vanished row as a delete and stamps nothing, and a row archived between apply and flush is
// harmless to stamp.
func (q *repository) MarkValidMeasurementUnitsAsIndexed(ctx context.Context, ids []string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if len(ids) == 0 {
		return nil
	}

	logger := q.logger.Clone().WithValue("id_count", len(ids))
	tracing.AttachToSpan(span, "id_count", len(ids))

	if _, err := q.generatedQuerier.MarkValidMeasurementUnitsAsIndexed(ctx, q.writeDB, ids); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking valid measurement units as indexed")
	}

	return nil
}

// ArchiveValidMeasurementUnit archives a valid measurement unit from the database by its ID.
func (q *repository) ArchiveValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validMeasurementUnitID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)

	return q.withEvent(ctx, logger, types.ValidMeasurementUnitArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidMeasurementUnitIDKey: validMeasurementUnitID,
	}, func(tx database.Tx) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveValidMeasurementUnit(ctx, tx, validMeasurementUnitID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving valid measurement unit")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}
