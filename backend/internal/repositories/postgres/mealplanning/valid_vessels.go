package mealplanning

import (
	"context"
	"database/sql"
	"fmt"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/observability"
	platformkeys "github.com/primandproper/platform-go/v9/observability/keys"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

var (
	_ types.ValidVesselDataManager = (*repository)(nil)
)

// ValidVesselExists fetches whether a valid vessel exists from the database.
func (q *repository) ValidVesselExists(ctx context.Context, validVesselID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validVesselID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidVesselIDKey, validVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, validVesselID)

	result, err := q.generatedQuerier.CheckValidVesselExistence(ctx, q.readDB, validVesselID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing valid vessel existence check")
	}

	return result, nil
}

// GetValidVessel fetches a valid vessel from the database.
func (q *repository) GetValidVessel(ctx context.Context, validVesselID string) (*types.ValidVessel, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validVesselID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidVesselIDKey, validVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, validVesselID)

	result, err := q.generatedQuerier.GetValidVessel(ctx, q.readDB, validVesselID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting valid vessels")
	}

	validVessel := &types.ValidVessel{
		CreatedAt:     result.CreatedAt,
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		CapacityUnit: &types.ValidMeasurementUnit{
			CreatedAt:     result.ValidMeasurementUnitCreatedAt.Time,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitArchivedAt),
			Name:          result.ValidMeasurementUnitName.String,
			IconPath:      result.ValidMeasurementUnitIconPath.String,
			ID:            result.ValidMeasurementUnitID.String,
			Description:   result.ValidMeasurementUnitDescription.String,
			PluralName:    result.ValidMeasurementUnitPluralName.String,
			Slug:          result.ValidMeasurementUnitSlug.String,
			Volumetric:    database.BoolFromNullBool(result.ValidMeasurementUnitVolumetric),
			Universal:     result.ValidMeasurementUnitUniversal.Bool,
			Metric:        result.ValidMeasurementUnitMetric.Bool,
			Imperial:      result.ValidMeasurementUnitImperial.Bool,
		},
		IconPath:                       result.IconPath,
		PluralName:                     result.PluralName,
		Description:                    result.Description,
		Name:                           result.Name,
		Slug:                           result.Slug,
		Shape:                          string(result.Shape),
		ID:                             result.ID,
		WidthInMillimeters:             database.Float32FromNullString(result.WidthInMillimeters),
		LengthInMillimeters:            database.Float32FromNullString(result.LengthInMillimeters),
		HeightInMillimeters:            database.Float32FromNullString(result.HeightInMillimeters),
		Capacity:                       database.Float32FromString(result.Capacity),
		IncludeInGeneratedInstructions: result.IncludeInGeneratedInstructions,
		DisplayInSummaryLists:          result.DisplayInSummaryLists,
		UsableForStorage:               result.UsableForStorage,
	}

	return validVessel, nil
}

// GetRandomValidVessel fetches a valid vessel from the database.
func (q *repository) GetRandomValidVessel(ctx context.Context) (*types.ValidVessel, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	result, err := q.generatedQuerier.GetRandomValidVessel(ctx, q.readDB)
	if err != nil {
		return nil, observability.PrepareError(err, span, "querying for random valid vessel")
	}

	validVessel := &types.ValidVessel{
		CreatedAt:     result.CreatedAt,
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		CapacityUnit: &types.ValidMeasurementUnit{
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			Name:          result.ValidMeasurementUnitName,
			IconPath:      result.ValidMeasurementUnitIconPath,
			ID:            result.ValidMeasurementUnitID,
			Description:   result.ValidMeasurementUnitDescription,
			PluralName:    result.ValidMeasurementUnitPluralName,
			Slug:          result.ValidMeasurementUnitSlug,
			Volumetric:    database.BoolFromNullBool(result.ValidMeasurementUnitVolumetric),
			Universal:     result.ValidMeasurementUnitUniversal,
			Metric:        result.ValidMeasurementUnitMetric,
			Imperial:      result.ValidMeasurementUnitImperial,
		},
		IconPath:                       result.IconPath,
		PluralName:                     result.PluralName,
		Description:                    result.Description,
		Name:                           result.Name,
		Slug:                           result.Slug,
		Shape:                          string(result.Shape),
		ID:                             result.ID,
		WidthInMillimeters:             database.Float32FromNullString(result.WidthInMillimeters),
		LengthInMillimeters:            database.Float32FromNullString(result.LengthInMillimeters),
		HeightInMillimeters:            database.Float32FromNullString(result.HeightInMillimeters),
		Capacity:                       database.Float32FromString(result.Capacity),
		IncludeInGeneratedInstructions: result.IncludeInGeneratedInstructions,
		DisplayInSummaryLists:          result.DisplayInSummaryLists,
		UsableForStorage:               result.UsableForStorage,
	}

	return validVessel, nil
}

// SearchForValidVessels fetches a valid vessel from the database.
func (q *repository) SearchForValidVessels(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidVessel], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if query == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, query)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.SearchForValidVessels(ctx, q.readDB, &generated.SearchForValidVesselsParams{
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
		return nil, observability.PrepareError(err, span, "querying for valid vessel")
	}

	var (
		data                      = []*types.ValidVessel{}
		filteredCount, totalCount uint64
	)

	for _, result := range results {
		validVessel := &types.ValidVessel{
			CreatedAt:                      result.CreatedAt,
			ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
			IconPath:                       result.IconPath,
			PluralName:                     result.PluralName,
			Description:                    result.Description,
			Name:                           result.Name,
			Slug:                           result.Slug,
			Shape:                          string(result.Shape),
			ID:                             result.ID,
			WidthInMillimeters:             database.Float32FromNullString(result.WidthInMillimeters),
			LengthInMillimeters:            database.Float32FromNullString(result.LengthInMillimeters),
			HeightInMillimeters:            database.Float32FromNullString(result.HeightInMillimeters),
			Capacity:                       database.Float32FromString(result.Capacity),
			IncludeInGeneratedInstructions: result.IncludeInGeneratedInstructions,
			DisplayInSummaryLists:          result.DisplayInSummaryLists,
			UsableForStorage:               result.UsableForStorage,
		}

		if result.CapacityUnit.Valid && result.CapacityUnit.String != "" {
			validVessel.CapacityUnit, err = q.GetValidMeasurementUnit(ctx, result.CapacityUnit.String)
			if err != nil {
				return nil, observability.PrepareAndLogError(err, logger, span, "getting valid measurement unit")
			}
		}

		data = append(data, validVessel)
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	if len(data) == 0 {
		return nil, sql.ErrNoRows
	}

	x := filtering.NewQueryFilteredResult(data, filteredCount, totalCount, func(vv *types.ValidVessel) string { return vv.ID }, filter)

	return x, nil
}

// GetValidVessels fetches a list of valid vessels from the database that meet a particular filter.
func (q *repository) GetValidVessels(ctx context.Context, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.ValidVessel], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidVessels(ctx, q.readDB, &generated.GetValidVesselsParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "querying for valid vessels")
	}

	var (
		data          []*types.ValidVessel
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		validVessel := &types.ValidVessel{
			CreatedAt:                      result.CreatedAt,
			ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
			IconPath:                       result.IconPath,
			PluralName:                     result.PluralName,
			Description:                    result.Description,
			Name:                           result.Name,
			Slug:                           result.Slug,
			Shape:                          string(result.Shape),
			ID:                             result.ID,
			WidthInMillimeters:             database.Float32FromNullString(result.WidthInMillimeters),
			LengthInMillimeters:            database.Float32FromNullString(result.LengthInMillimeters),
			HeightInMillimeters:            database.Float32FromNullString(result.HeightInMillimeters),
			Capacity:                       database.Float32FromString(result.Capacity),
			IncludeInGeneratedInstructions: result.IncludeInGeneratedInstructions,
			DisplayInSummaryLists:          result.DisplayInSummaryLists,
			UsableForStorage:               result.UsableForStorage,
		}

		if result.CapacityUnit.Valid && result.CapacityUnit.String != "" {
			validVessel.CapacityUnit, err = q.GetValidMeasurementUnit(ctx, result.CapacityUnit.String)
			if err != nil {
				return nil, observability.PrepareAndLogError(err, logger, span, "getting valid measurement unit")
			}
		}

		data = append(data, validVessel)
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vv *types.ValidVessel) string { return vv.ID },
		filter,
	)

	return x, nil
}

// GetValidVesselsWithIDs fetches a list of valid vessels from the database that meet a particular filter.
func (q *repository) GetValidVesselsWithIDs(ctx context.Context, ids []string) ([]*types.ValidVessel, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if len(ids) == 0 {
		return nil, sql.ErrNoRows
	}

	results, err := q.generatedQuerier.GetValidVesselsWithIDs(ctx, q.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid vessels id list retrieval query")
	}

	validVessels := []*types.ValidVessel{}
	for _, result := range results {
		validVessels = append(validVessels, &types.ValidVessel{
			CreatedAt:     result.CreatedAt,
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			CapacityUnit: &types.ValidMeasurementUnit{
				CreatedAt:     result.ValidMeasurementUnitCreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitArchivedAt),
				Name:          result.ValidMeasurementUnitName,
				IconPath:      result.ValidMeasurementUnitIconPath,
				ID:            result.ValidMeasurementUnitID,
				Description:   result.ValidMeasurementUnitDescription,
				PluralName:    result.ValidMeasurementUnitPluralName,
				Slug:          result.ValidMeasurementUnitSlug,
				Volumetric:    database.BoolFromNullBool(result.ValidMeasurementUnitVolumetric),
				Universal:     result.ValidMeasurementUnitUniversal,
				Metric:        result.ValidMeasurementUnitMetric,
				Imperial:      result.ValidMeasurementUnitImperial,
			},
			IconPath:                       result.IconPath,
			PluralName:                     result.PluralName,
			Description:                    result.Description,
			Name:                           result.Name,
			Slug:                           result.Slug,
			Shape:                          string(result.Shape),
			ID:                             result.ID,
			WidthInMillimeters:             database.Float32FromNullString(result.WidthInMillimeters),
			LengthInMillimeters:            database.Float32FromNullString(result.LengthInMillimeters),
			HeightInMillimeters:            database.Float32FromNullString(result.HeightInMillimeters),
			Capacity:                       database.Float32FromString(result.Capacity),
			IncludeInGeneratedInstructions: result.IncludeInGeneratedInstructions,
			DisplayInSummaryLists:          result.DisplayInSummaryLists,
			UsableForStorage:               result.UsableForStorage,
		})
	}

	return validVessels, nil
}

// GetValidVesselIDsThatNeedSearchIndexing fetches a list of valid vessels from the database that meet a particular filter.
func (q *repository) GetValidVesselIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	results, err := q.generatedQuerier.GetValidVesselIDsNeedingIndexing(ctx, q.readDB)
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing valid vessels list retrieval query")
	}

	return results, nil
}

// CreateValidVessel creates a valid vessel in the database.
func (q *repository) CreateValidVessel(ctx context.Context, input *types.ValidVesselDatabaseCreationInput) (*types.ValidVessel, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidVesselIDKey, input.ID)

	// create the valid vessel.
	if err := q.withEvent(ctx, logger, types.ValidVesselCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidVesselIDKey: input.ID,
	}, func(tx database.SQLQueryExecutor) error {
		return q.generatedQuerier.CreateValidVessel(ctx, tx, &generated.CreateValidVesselParams{
			Slug:                           input.Slug,
			ID:                             input.ID,
			PluralName:                     input.PluralName,
			Description:                    input.Description,
			IconPath:                       input.IconPath,
			Shape:                          generated.VesselShape(input.Shape),
			Name:                           input.Name,
			Capacity:                       database.StringFromFloat32(input.Capacity),
			CapacityUnit:                   database.NullStringFromStringPointer(input.CapacityUnitID),
			WidthInMillimeters:             database.NullStringFromFloat32(input.WidthInMillimeters),
			LengthInMillimeters:            database.NullStringFromFloat32(input.LengthInMillimeters),
			HeightInMillimeters:            database.NullStringFromFloat32(input.HeightInMillimeters),
			IncludeInGeneratedInstructions: input.IncludeInGeneratedInstructions,
			DisplayInSummaryLists:          input.DisplayInSummaryLists,
			UsableForStorage:               input.UsableForStorage,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid vessel creation query")
	}

	x := &types.ValidVessel{
		ID:                             input.ID,
		Name:                           input.Name,
		PluralName:                     input.PluralName,
		Description:                    input.Description,
		IconPath:                       input.IconPath,
		Slug:                           input.Slug,
		Shape:                          input.Shape,
		WidthInMillimeters:             input.WidthInMillimeters,
		LengthInMillimeters:            input.LengthInMillimeters,
		HeightInMillimeters:            input.HeightInMillimeters,
		Capacity:                       input.Capacity,
		IncludeInGeneratedInstructions: input.IncludeInGeneratedInstructions,
		DisplayInSummaryLists:          input.DisplayInSummaryLists,
		UsableForStorage:               input.UsableForStorage,
		CreatedAt:                      q.CurrentTime(),
	}

	if input.CapacityUnitID != nil {
		x.CapacityUnit = &types.ValidMeasurementUnit{ID: *input.CapacityUnitID}
	}

	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, x.ID)
	logger.Info("valid vessel created")

	return x, nil
}

// UpdateValidVessel updates a particular valid vessel.
func (q *repository) UpdateValidVessel(ctx context.Context, updated *types.ValidVessel) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidVesselIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, updated.ID)

	if updated.CapacityUnit == nil {
		return fmt.Errorf("capacity unit: %w", platformerrors.ErrNilInputProvided)
	}

	if err := q.withEvent(ctx, logger, types.ValidVesselUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidVesselIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		_, updateErr := q.generatedQuerier.UpdateValidVessel(ctx, tx, &generated.UpdateValidVesselParams{
			Name:                           updated.Name,
			PluralName:                     updated.PluralName,
			Description:                    updated.Description,
			IconPath:                       updated.IconPath,
			UsableForStorage:               updated.UsableForStorage,
			Slug:                           updated.Slug,
			DisplayInSummaryLists:          updated.DisplayInSummaryLists,
			IncludeInGeneratedInstructions: updated.IncludeInGeneratedInstructions,
			Capacity:                       database.StringFromFloat32(updated.Capacity),
			CapacityUnit:                   database.NullStringFromString(updated.CapacityUnit.ID),
			WidthInMillimeters:             database.NullStringFromFloat32(updated.WidthInMillimeters),
			LengthInMillimeters:            database.NullStringFromFloat32(updated.LengthInMillimeters),
			HeightInMillimeters:            database.NullStringFromFloat32(updated.HeightInMillimeters),
			Shape:                          generated.VesselShape(updated.Shape),
			ID:                             updated.ID,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid vessel")
	}

	logger.Info("valid vessel updated")

	return nil
}

// MarkValidVesselAsIndexed updates a particular valid vessel's last_indexed_at value.
func (q *repository) MarkValidVesselAsIndexed(ctx context.Context, validVesselID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validVesselID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidVesselIDKey, validVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, validVesselID)

	if _, err := q.generatedQuerier.UpdateValidVesselLastIndexedAt(ctx, q.writeDB, validVesselID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking valid vessel as indexed")
	}

	logger.Info("valid vessel marked as indexed")

	return nil
}

// ArchiveValidVessel archives a valid vessel from the database by its ID.
func (q *repository) ArchiveValidVessel(ctx context.Context, validVesselID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validVesselID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidVesselIDKey, validVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, validVesselID)

	return q.withEvent(ctx, logger, types.ValidVesselArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidVesselIDKey: validVesselID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveValidVessel(ctx, tx, validVesselID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving valid vessel")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}
