package mealplanning

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v10/database"
	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/observability"
	platformkeys "github.com/primandproper/platform-go/v10/observability/keys"
	"github.com/primandproper/platform-go/v10/observability/tracing"
)

var (
	_ mealplanning.ValidPreparationDataManager = (*repository)(nil)
)

// ValidPreparationExists fetches whether a valid preparation exists from the database.
func (q *repository) ValidPreparationExists(ctx context.Context, validPreparationID string) (bool, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validPreparationID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidPreparationIDKey, validPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, validPreparationID)

	exists, err := q.generatedQuerier.CheckValidPreparationExistence(ctx, q.readDB, validPreparationID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "checking valid preparation existence")
	}

	return exists, nil
}

// GetValidPreparation fetches a valid preparation from the database.
func (q *repository) GetValidPreparation(ctx context.Context, validPreparationID string) (*mealplanning.ValidPreparation, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validPreparationID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidPreparationIDKey, validPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, validPreparationID)

	result, err := q.generatedQuerier.GetValidPreparation(ctx, q.readDB, validPreparationID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting random valid preparation")
	}

	validPreparation := &mealplanning.ValidPreparation{
		CreatedAt:                   result.CreatedAt,
		ArchivedAt:                  database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt:               database.TimePointerFromNullTime(result.LastUpdatedAt),
		IconPath:                    result.IconPath,
		PastTense:                   result.PastTense,
		ID:                          result.ID,
		Name:                        result.Name,
		Description:                 result.Description,
		Slug:                        result.Slug,
		MinIngredientCount:          uint16(result.MinimumIngredientCount),
		MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.MaximumIngredientCount),
		MinInstrumentCount:          uint16(result.MinimumInstrumentCount),
		MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.MaximumInstrumentCount),
		MinVesselCount:              uint16(result.MinimumVesselCount),
		MaxVesselCount:              database.Uint16PointerFromNullInt32(result.MaximumVesselCount),
		RestrictToIngredients:       result.RestrictToIngredients,
		TemperatureRequired:         result.TemperatureRequired,
		TimeEstimateRequired:        result.TimeEstimateRequired,
		ConditionExpressionRequired: result.ConditionExpressionRequired,
		ConsumesVessel:              result.ConsumesVessel,
		OnlyForVessels:              result.OnlyForVessels,
		YieldsNothing:               result.YieldsNothing,
	}

	return validPreparation, nil
}

// GetRandomValidPreparation fetches a valid preparation from the database.
func (q *repository) GetRandomValidPreparation(ctx context.Context) (*mealplanning.ValidPreparation, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	result, err := q.generatedQuerier.GetRandomValidPreparation(ctx, q.readDB)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting random valid preparation")
	}

	validPreparation := &mealplanning.ValidPreparation{
		CreatedAt:                   result.CreatedAt,
		ArchivedAt:                  database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt:               database.TimePointerFromNullTime(result.LastUpdatedAt),
		IconPath:                    result.IconPath,
		PastTense:                   result.PastTense,
		ID:                          result.ID,
		Name:                        result.Name,
		Description:                 result.Description,
		Slug:                        result.Slug,
		MinIngredientCount:          uint16(result.MinimumIngredientCount),
		MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.MaximumIngredientCount),
		MinInstrumentCount:          uint16(result.MinimumInstrumentCount),
		MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.MaximumInstrumentCount),
		MinVesselCount:              uint16(result.MinimumVesselCount),
		MaxVesselCount:              database.Uint16PointerFromNullInt32(result.MaximumVesselCount),
		RestrictToIngredients:       result.RestrictToIngredients,
		TemperatureRequired:         result.TemperatureRequired,
		TimeEstimateRequired:        result.TimeEstimateRequired,
		ConditionExpressionRequired: result.ConditionExpressionRequired,
		ConsumesVessel:              result.ConsumesVessel,
		OnlyForVessels:              result.OnlyForVessels,
		YieldsNothing:               result.YieldsNothing,
	}

	return validPreparation, nil
}

// SearchForValidPreparations fetches a valid preparation from the database.
func (q *repository) SearchForValidPreparations(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparation], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if query == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, platformkeys.SearchQueryKey, query)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.SearchForValidPreparations(ctx, q.readDB, &generated.SearchForValidPreparationsParams{
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
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid preparations search")
	}

	var (
		data                      = []*mealplanning.ValidPreparation{}
		filteredCount, totalCount uint64
	)

	for _, result := range results {
		data = append(data, &mealplanning.ValidPreparation{
			CreatedAt:                   result.CreatedAt,
			ArchivedAt:                  database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:               database.TimePointerFromNullTime(result.LastUpdatedAt),
			IconPath:                    result.IconPath,
			PastTense:                   result.PastTense,
			ID:                          result.ID,
			Name:                        result.Name,
			Description:                 result.Description,
			Slug:                        result.Slug,
			MinIngredientCount:          uint16(result.MinimumIngredientCount),
			MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.MaximumIngredientCount),
			MinInstrumentCount:          uint16(result.MinimumInstrumentCount),
			MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.MaximumInstrumentCount),
			MinVesselCount:              uint16(result.MinimumVesselCount),
			MaxVesselCount:              database.Uint16PointerFromNullInt32(result.MaximumVesselCount),
			RestrictToIngredients:       result.RestrictToIngredients,
			TemperatureRequired:         result.TemperatureRequired,
			TimeEstimateRequired:        result.TimeEstimateRequired,
			ConditionExpressionRequired: result.ConditionExpressionRequired,
			ConsumesVessel:              result.ConsumesVessel,
			OnlyForVessels:              result.OnlyForVessels,
			YieldsNothing:               result.YieldsNothing,
		})
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(data, filteredCount, totalCount, func(vp *mealplanning.ValidPreparation) string { return vp.ID }, filter)

	return x, nil
}

// GetValidPreparations fetches a list of valid preparations from the database that meet a particular filter.
func (q *repository) GetValidPreparations(ctx context.Context, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidPreparation], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidPreparations(ctx, q.readDB, &generated.GetValidPreparationsParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid preparations list retrieval query")
	}

	var (
		data          []*mealplanning.ValidPreparation
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
		data = append(data, &mealplanning.ValidPreparation{
			CreatedAt:                   result.CreatedAt,
			MinIngredientCount:          uint16(result.MinimumIngredientCount),
			MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.MaximumIngredientCount),
			MinInstrumentCount:          uint16(result.MinimumInstrumentCount),
			MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.MaximumInstrumentCount),
			MinVesselCount:              uint16(result.MinimumVesselCount),
			MaxVesselCount:              database.Uint16PointerFromNullInt32(result.MaximumVesselCount),
			ArchivedAt:                  database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:               database.TimePointerFromNullTime(result.LastUpdatedAt),
			IconPath:                    result.IconPath,
			PastTense:                   result.PastTense,
			ID:                          result.ID,
			Name:                        result.Name,
			Description:                 result.Description,
			Slug:                        result.Slug,
			RestrictToIngredients:       result.RestrictToIngredients,
			TemperatureRequired:         result.TemperatureRequired,
			TimeEstimateRequired:        result.TimeEstimateRequired,
			ConditionExpressionRequired: result.ConditionExpressionRequired,
			ConsumesVessel:              result.ConsumesVessel,
			OnlyForVessels:              result.OnlyForVessels,
			YieldsNothing:               result.YieldsNothing,
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vp *mealplanning.ValidPreparation) string { return vp.ID },
		filter,
	)

	return x, nil
}

// GetValidPreparationsWithIDs fetches a list of valid preparations from the database that meet a particular filter.
func (q *repository) GetValidPreparationsWithIDs(ctx context.Context, ids []string) ([]*mealplanning.ValidPreparation, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if len(ids) == 0 {
		return nil, sql.ErrNoRows
	}
	logger := q.logger.WithValue("ids_count", len(ids))

	results, err := q.generatedQuerier.GetValidPreparationsWithIDs(ctx, q.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting valid preparations by MealPlanTaskID")
	}

	preparations := []*mealplanning.ValidPreparation{}
	for _, result := range results {
		preparations = append(preparations, &mealplanning.ValidPreparation{
			CreatedAt:                   result.CreatedAt,
			MinIngredientCount:          uint16(result.MinimumIngredientCount),
			MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.MaximumIngredientCount),
			MinInstrumentCount:          uint16(result.MinimumInstrumentCount),
			MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.MaximumInstrumentCount),
			MinVesselCount:              uint16(result.MinimumVesselCount),
			MaxVesselCount:              database.Uint16PointerFromNullInt32(result.MaximumVesselCount),
			ArchivedAt:                  database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:               database.TimePointerFromNullTime(result.LastUpdatedAt),
			IconPath:                    result.IconPath,
			PastTense:                   result.PastTense,
			ID:                          result.ID,
			Name:                        result.Name,
			Description:                 result.Description,
			Slug:                        result.Slug,
			RestrictToIngredients:       result.RestrictToIngredients,
			TemperatureRequired:         result.TemperatureRequired,
			TimeEstimateRequired:        result.TimeEstimateRequired,
			ConditionExpressionRequired: result.ConditionExpressionRequired,
			ConsumesVessel:              result.ConsumesVessel,
			OnlyForVessels:              result.OnlyForVessels,
			YieldsNothing:               result.YieldsNothing,
		})
	}

	return preparations, nil
}

// ScanValidPreparationIDsForReindex returns up to limit IDs sorting strictly after `after`, in ascending byte order.
//
// It is the source half of a search reindex: searchsync.Reindexer walks this to find every
// document that should exist, and prunes the index of anything it does not name. It replaces
// the "IDs that need indexing" sampler platform-go v10 removed, which asked a different and
// weaker question — which rows look stale — and could only ever be probabilistically right,
// because a row the sampler had not reached was a row the index was wrong about.
func (q *repository) ScanValidPreparationIDsForReindex(ctx context.Context, after string, limit int) ([]string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	results, err := q.generatedQuerier.ScanValidPreparationIDsForReindex(ctx, q.readDB, &generated.ScanValidPreparationIDsForReindexParams{
		Cursor:      after,
		ResultLimit: limit,
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing valid preparations reindex scan query")
	}

	return results, nil
}

// CreateValidPreparation creates a valid preparation in the database.
func (q *repository) CreateValidPreparation(ctx context.Context, input *mealplanning.ValidPreparationDatabaseCreationInput) (*mealplanning.ValidPreparation, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidPreparationIDKey, input.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, input.ID)

	// create the valid preparation.
	if err := q.withEvent(ctx, logger, mealplanning.ValidPreparationCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidPreparationIDKey: input.ID,
	}, func(tx database.SQLQueryExecutor) error {
		return q.generatedQuerier.CreateValidPreparation(ctx, tx, &generated.CreateValidPreparationParams{
			ID:                          input.ID,
			Name:                        input.Name,
			Description:                 input.Description,
			IconPath:                    input.IconPath,
			YieldsNothing:               input.YieldsNothing,
			RestrictToIngredients:       input.RestrictToIngredients,
			MinimumIngredientCount:      int32(input.MinIngredientCount),
			MaximumIngredientCount:      database.NullInt32FromUint16Pointer(input.MaxIngredientCount),
			MinimumInstrumentCount:      int32(input.MinInstrumentCount),
			MaximumInstrumentCount:      database.NullInt32FromUint16Pointer(input.MaxInstrumentCount),
			TemperatureRequired:         input.TemperatureRequired,
			TimeEstimateRequired:        input.TimeEstimateRequired,
			ConditionExpressionRequired: input.ConditionExpressionRequired,
			ConsumesVessel:              input.ConsumesVessel,
			OnlyForVessels:              input.OnlyForVessels,
			MinimumVesselCount:          int32(input.MinVesselCount),
			MaximumVesselCount:          database.NullInt32FromUint16Pointer(input.MaxVesselCount),
			PastTense:                   input.PastTense,
			Slug:                        input.Slug,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid preparation creation query")
	}

	x := &mealplanning.ValidPreparation{
		ID:                          input.ID,
		Name:                        input.Name,
		Description:                 input.Description,
		IconPath:                    input.IconPath,
		YieldsNothing:               input.YieldsNothing,
		RestrictToIngredients:       input.RestrictToIngredients,
		Slug:                        input.Slug,
		PastTense:                   input.PastTense,
		MinIngredientCount:          input.MinIngredientCount,
		MaxIngredientCount:          input.MaxIngredientCount,
		MinInstrumentCount:          input.MinInstrumentCount,
		MaxInstrumentCount:          input.MaxInstrumentCount,
		MinVesselCount:              input.MinVesselCount,
		MaxVesselCount:              input.MaxVesselCount,
		TemperatureRequired:         input.TemperatureRequired,
		TimeEstimateRequired:        input.TimeEstimateRequired,
		ConditionExpressionRequired: input.ConditionExpressionRequired,
		ConsumesVessel:              input.ConsumesVessel,
		OnlyForVessels:              input.OnlyForVessels,
		CreatedAt:                   q.CurrentTime(),
	}

	logger.Info("valid preparation created")

	return x, nil
}

// UpdateValidPreparation updates a particular valid preparation.
func (q *repository) UpdateValidPreparation(ctx context.Context, updated *mealplanning.ValidPreparation) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidPreparationIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, mealplanning.ValidPreparationUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidPreparationIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		_, updateErr := q.generatedQuerier.UpdateValidPreparation(ctx, tx, &generated.UpdateValidPreparationParams{
			Description:                 updated.Description,
			IconPath:                    updated.IconPath,
			ID:                          updated.ID,
			Name:                        updated.Name,
			PastTense:                   updated.PastTense,
			Slug:                        updated.Slug,
			MaximumIngredientCount:      database.NullInt32FromUint16Pointer(updated.MaxIngredientCount),
			MaximumInstrumentCount:      database.NullInt32FromUint16Pointer(updated.MaxInstrumentCount),
			MaximumVesselCount:          database.NullInt32FromUint16Pointer(updated.MaxVesselCount),
			MinimumVesselCount:          int32(updated.MinVesselCount),
			MinimumIngredientCount:      int32(updated.MinIngredientCount),
			MinimumInstrumentCount:      int32(updated.MinInstrumentCount),
			RestrictToIngredients:       updated.RestrictToIngredients,
			OnlyForVessels:              updated.OnlyForVessels,
			ConsumesVessel:              updated.ConsumesVessel,
			ConditionExpressionRequired: updated.ConditionExpressionRequired,
			TimeEstimateRequired:        updated.TimeEstimateRequired,
			TemperatureRequired:         updated.TemperatureRequired,
			YieldsNothing:               updated.YieldsNothing,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid preparation")
	}

	logger.Info("valid preparation updated")

	return nil
}

// MarkValidPreparationAsIndexed updates a particular valid preparation's last_indexed_at value.
func (q *repository) MarkValidPreparationAsIndexed(ctx context.Context, validPreparationID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validPreparationID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidPreparationIDKey, validPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, validPreparationID)

	if _, err := q.generatedQuerier.UpdateValidPreparationLastIndexedAt(ctx, q.writeDB, validPreparationID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking valid preparation as indexed")
	}

	logger.Info("valid preparation marked as indexed")

	return nil
}

// ArchiveValidPreparation archives a valid preparation from the database by its ID.
func (q *repository) ArchiveValidPreparation(ctx context.Context, validPreparationID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validPreparationID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidPreparationIDKey, validPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, validPreparationID)

	return q.withEvent(ctx, logger, mealplanning.ValidPreparationArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidPreparationIDKey: validPreparationID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveValidPreparation(ctx, tx, validPreparationID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "updating valid preparation")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}
