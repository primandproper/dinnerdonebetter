package mealplanning

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

var (
	_ mealplanning.ValidPreparationVesselDataManager = (*repository)(nil)
)

// ValidPreparationVesselExists fetches whether a valid preparation vessel exists from the database.
func (q *repository) ValidPreparationVesselExists(ctx context.Context, validPreparationVesselID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if validPreparationVesselID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, validPreparationVesselID)

	exists, err = q.generatedQuerier.CheckValidPreparationVesselExistence(ctx, q.readDB, validPreparationVesselID)
	if err != nil {
		return false, observability.PrepareError(err, span, "performing valid preparation vessel existence check")
	}

	return exists, nil
}

// GetValidPreparationVessel fetches a valid preparation vessel from the database.
func (q *repository) GetValidPreparationVessel(ctx context.Context, validPreparationVesselID string) (*mealplanning.ValidPreparationVessel, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if validPreparationVesselID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, validPreparationVesselID)

	result, err := q.generatedQuerier.GetValidPreparationVessel(ctx, q.readDB, validPreparationVesselID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "reading valid preparation vessel from database")
	}

	validPreparationVessel := &mealplanning.ValidPreparationVessel{
		CreatedAt:     result.ValidPreparationVesselCreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.ValidPreparationVesselLastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ValidPreparationVesselArchivedAt),
		ID:            result.ValidPreparationVesselID,
		Notes:         result.ValidPreparationVesselNotes,
		Vessel: mealplanning.ValidVessel{
			CreatedAt:     result.ValidVesselCreatedAt,
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidVesselArchivedAt),
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidVesselLastUpdatedAt),
			CapacityUnit: &mealplanning.ValidMeasurementUnit{
				CreatedAt:     result.ValidMeasurementUnitCreatedAt.Time,
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitArchivedAt),
				Name:          result.ValidMeasurementUnitName.String,
				IconPath:      result.ValidMeasurementUnitIconPath.String,
				ID:            result.ValidMeasurementUnitID.String,
				Description:   result.ValidMeasurementUnitDescription.String,
				PluralName:    result.ValidMeasurementUnitPluralName.String,
				Slug:          result.ValidMeasurementUnitSlug.String,
				Volumetric:    result.ValidMeasurementUnitVolumetric.Bool,
				Universal:     result.ValidMeasurementUnitUniversal.Bool,
				Metric:        result.ValidMeasurementUnitMetric.Bool,
				Imperial:      result.ValidMeasurementUnitImperial.Bool,
			},
			IconPath:                       result.ValidVesselIconPath,
			PluralName:                     result.ValidVesselPluralName,
			Description:                    result.ValidVesselDescription,
			Name:                           result.ValidVesselName,
			Slug:                           result.ValidVesselSlug,
			Shape:                          string(result.ValidVesselShape),
			ID:                             result.ValidVesselID,
			WidthInMillimeters:             database.Float32FromNullString(result.ValidVesselWidthInMillimeters),
			LengthInMillimeters:            database.Float32FromNullString(result.ValidVesselLengthInMillimeters),
			HeightInMillimeters:            database.Float32FromNullString(result.ValidVesselHeightInMillimeters),
			Capacity:                       database.Float32FromString(result.ValidVesselCapacity),
			IncludeInGeneratedInstructions: result.ValidVesselIncludeInGeneratedInstructions,
			DisplayInSummaryLists:          result.ValidVesselDisplayInSummaryLists,
			UsableForStorage:               result.ValidVesselUsableForStorage,
		},
		Preparation: mealplanning.ValidPreparation{
			CreatedAt:                   result.ValidPreparationCreatedAt,
			MinInstrumentCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
			MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumInstrumentCount),
			MinIngredientCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
			MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumIngredientCount),
			MinVesselCount:              uint16(result.ValidPreparationMinimumVesselCount),
			MaxVesselCount:              database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumVesselCount),
			ArchivedAt:                  database.TimePointerFromNullTime(result.ValidPreparationArchivedAt),
			LastUpdatedAt:               database.TimePointerFromNullTime(result.ValidPreparationLastUpdatedAt),
			IconPath:                    result.ValidPreparationIconPath,
			PastTense:                   result.ValidPreparationPastTense,
			ID:                          result.ValidPreparationID,
			Name:                        result.ValidPreparationName,
			Description:                 result.ValidPreparationDescription,
			Slug:                        result.ValidPreparationSlug,
			RestrictToIngredients:       result.ValidPreparationRestrictToIngredients,
			TemperatureRequired:         result.ValidPreparationTemperatureRequired,
			TimeEstimateRequired:        result.ValidPreparationTimeEstimateRequired,
			ConditionExpressionRequired: result.ValidPreparationConditionExpressionRequired,
			ConsumesVessel:              result.ValidPreparationConsumesVessel,
			OnlyForVessels:              result.ValidPreparationOnlyForVessels,
			YieldsNothing:               result.ValidPreparationYieldsNothing,
		},
	}

	return validPreparationVessel, nil
}

// GetValidPreparationVessels fetches a list of valid preparation vessels from the database that meet a particular filter.
func (q *repository) GetValidPreparationVessels(ctx context.Context, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidPreparationVessels(ctx, q.readDB, &generated.GetValidPreparationVesselsParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		PageCursor:      database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid preparation vessels list retrieval query")
	}

	var (
		data          []*mealplanning.ValidPreparationVessel
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		validPreparationVessel := &mealplanning.ValidPreparationVessel{
			CreatedAt:     result.ValidPreparationVesselCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidPreparationVesselLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidPreparationVesselArchivedAt),
			ID:            result.ValidPreparationVesselID,
			Notes:         result.ValidPreparationVesselNotes,
			Vessel: mealplanning.ValidVessel{
				CreatedAt:     result.ValidVesselCreatedAt,
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidVesselArchivedAt),
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidVesselLastUpdatedAt),
				CapacityUnit: &mealplanning.ValidMeasurementUnit{
					CreatedAt:     result.ValidMeasurementUnitCreatedAt.Time,
					LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitLastUpdatedAt),
					ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitArchivedAt),
					Name:          result.ValidMeasurementUnitName.String,
					IconPath:      result.ValidMeasurementUnitIconPath.String,
					ID:            result.ValidMeasurementUnitID.String,
					Description:   result.ValidMeasurementUnitDescription.String,
					PluralName:    result.ValidMeasurementUnitPluralName.String,
					Slug:          result.ValidMeasurementUnitSlug.String,
					Volumetric:    result.ValidMeasurementUnitVolumetric.Bool,
					Universal:     result.ValidMeasurementUnitUniversal.Bool,
					Metric:        result.ValidMeasurementUnitMetric.Bool,
					Imperial:      result.ValidMeasurementUnitImperial.Bool,
				},
				IconPath:                       result.ValidVesselIconPath,
				PluralName:                     result.ValidVesselPluralName,
				Description:                    result.ValidVesselDescription,
				Name:                           result.ValidVesselName,
				Slug:                           result.ValidVesselSlug,
				Shape:                          string(result.ValidVesselShape),
				ID:                             result.ValidVesselID,
				WidthInMillimeters:             database.Float32FromNullString(result.ValidVesselWidthInMillimeters),
				LengthInMillimeters:            database.Float32FromNullString(result.ValidVesselLengthInMillimeters),
				HeightInMillimeters:            database.Float32FromNullString(result.ValidVesselHeightInMillimeters),
				Capacity:                       database.Float32FromString(result.ValidVesselCapacity),
				IncludeInGeneratedInstructions: result.ValidVesselIncludeInGeneratedInstructions,
				DisplayInSummaryLists:          result.ValidVesselDisplayInSummaryLists,
				UsableForStorage:               result.ValidVesselUsableForStorage,
			},
			Preparation: mealplanning.ValidPreparation{
				CreatedAt:                   result.ValidPreparationCreatedAt,
				MinInstrumentCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
				MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumInstrumentCount),
				MinIngredientCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
				MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumIngredientCount),
				MinVesselCount:              uint16(result.ValidPreparationMinimumVesselCount),
				MaxVesselCount:              database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumVesselCount),
				ArchivedAt:                  database.TimePointerFromNullTime(result.ValidPreparationArchivedAt),
				LastUpdatedAt:               database.TimePointerFromNullTime(result.ValidPreparationLastUpdatedAt),
				IconPath:                    result.ValidPreparationIconPath,
				PastTense:                   result.ValidPreparationPastTense,
				ID:                          result.ValidPreparationID,
				Name:                        result.ValidPreparationName,
				Description:                 result.ValidPreparationDescription,
				Slug:                        result.ValidPreparationSlug,
				RestrictToIngredients:       result.ValidPreparationRestrictToIngredients,
				TemperatureRequired:         result.ValidPreparationTemperatureRequired,
				TimeEstimateRequired:        result.ValidPreparationTimeEstimateRequired,
				ConditionExpressionRequired: result.ValidPreparationConditionExpressionRequired,
				ConsumesVessel:              result.ValidPreparationConsumesVessel,
				OnlyForVessels:              result.ValidPreparationOnlyForVessels,
				YieldsNothing:               result.ValidPreparationYieldsNothing,
			},
		}
		if validPreparationVessel.Vessel.CapacityUnit.ID == "" {
			validPreparationVessel.Vessel.CapacityUnit = nil
		}

		data = append(data, validPreparationVessel)
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vpv *mealplanning.ValidPreparationVessel) string { return vpv.ID },
		filter,
	)

	return x, nil
}

// GetValidPreparationVesselsForPreparation fetches a list of valid preparation vessels from the database that meet a particular filter.
func (q *repository) GetValidPreparationVesselsForPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if preparationID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidPreparationIDKey, preparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, preparationID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidPreparationVesselsForPreparation(ctx, q.readDB, &generated.GetValidPreparationVesselsForPreparationParams{
		ID:              preparationID,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		PageCursor:      database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid preparation vessels list retrieval query")
	}

	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}

	var (
		data          []*mealplanning.ValidPreparationVessel
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		validPreparationVessel := &mealplanning.ValidPreparationVessel{
			CreatedAt:     result.ValidPreparationVesselCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidPreparationVesselLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidPreparationVesselArchivedAt),
			ID:            result.ValidPreparationVesselID,
			Notes:         result.ValidPreparationVesselNotes,
			Vessel: mealplanning.ValidVessel{
				CreatedAt:     result.ValidVesselCreatedAt,
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidVesselArchivedAt),
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidVesselLastUpdatedAt),
				CapacityUnit: &mealplanning.ValidMeasurementUnit{
					CreatedAt:     result.ValidMeasurementUnitCreatedAt.Time,
					LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitLastUpdatedAt),
					ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitArchivedAt),
					Name:          result.ValidMeasurementUnitName.String,
					IconPath:      result.ValidMeasurementUnitIconPath.String,
					ID:            result.ValidMeasurementUnitID.String,
					Description:   result.ValidMeasurementUnitDescription.String,
					PluralName:    result.ValidMeasurementUnitPluralName.String,
					Slug:          result.ValidMeasurementUnitSlug.String,
					Volumetric:    result.ValidMeasurementUnitVolumetric.Bool,
					Universal:     result.ValidMeasurementUnitUniversal.Bool,
					Metric:        result.ValidMeasurementUnitMetric.Bool,
					Imperial:      result.ValidMeasurementUnitImperial.Bool,
				},
				IconPath:                       result.ValidVesselIconPath,
				PluralName:                     result.ValidVesselPluralName,
				Description:                    result.ValidVesselDescription,
				Name:                           result.ValidVesselName,
				Slug:                           result.ValidVesselSlug,
				Shape:                          string(result.ValidVesselShape),
				ID:                             result.ValidVesselID,
				WidthInMillimeters:             database.Float32FromNullString(result.ValidVesselWidthInMillimeters),
				LengthInMillimeters:            database.Float32FromNullString(result.ValidVesselLengthInMillimeters),
				HeightInMillimeters:            database.Float32FromNullString(result.ValidVesselHeightInMillimeters),
				Capacity:                       database.Float32FromString(result.ValidVesselCapacity),
				IncludeInGeneratedInstructions: result.ValidVesselIncludeInGeneratedInstructions,
				DisplayInSummaryLists:          result.ValidVesselDisplayInSummaryLists,
				UsableForStorage:               result.ValidVesselUsableForStorage,
			},
			Preparation: mealplanning.ValidPreparation{
				CreatedAt:                   result.ValidPreparationCreatedAt,
				MinInstrumentCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
				MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumInstrumentCount),
				MinIngredientCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
				MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumIngredientCount),
				MinVesselCount:              uint16(result.ValidPreparationMinimumVesselCount),
				MaxVesselCount:              database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumVesselCount),
				ArchivedAt:                  database.TimePointerFromNullTime(result.ValidPreparationArchivedAt),
				LastUpdatedAt:               database.TimePointerFromNullTime(result.ValidPreparationLastUpdatedAt),
				IconPath:                    result.ValidPreparationIconPath,
				PastTense:                   result.ValidPreparationPastTense,
				ID:                          result.ValidPreparationID,
				Name:                        result.ValidPreparationName,
				Description:                 result.ValidPreparationDescription,
				Slug:                        result.ValidPreparationSlug,
				RestrictToIngredients:       result.ValidPreparationRestrictToIngredients,
				TemperatureRequired:         result.ValidPreparationTemperatureRequired,
				TimeEstimateRequired:        result.ValidPreparationTimeEstimateRequired,
				ConditionExpressionRequired: result.ValidPreparationConditionExpressionRequired,
				ConsumesVessel:              result.ValidPreparationConsumesVessel,
				OnlyForVessels:              result.ValidPreparationOnlyForVessels,
				YieldsNothing:               result.ValidPreparationYieldsNothing,
			},
		}
		if validPreparationVessel.Vessel.CapacityUnit.ID == "" {
			validPreparationVessel.Vessel.CapacityUnit = nil
		}

		data = append(data, validPreparationVessel)
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vpv *mealplanning.ValidPreparationVessel) string { return vpv.ID },
		filter,
	)

	return x, nil
}

// GetValidPreparationVesselsForVessel fetches a list of valid preparation vessels from the database that meet a particular filter.
func (q *repository) GetValidPreparationVesselsForVessel(ctx context.Context, vesselID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if vesselID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, vesselID)
	logger = logger.WithValue(mealplanningkeys.ValidVesselIDKey, vesselID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidPreparationVesselsForVessel(ctx, q.readDB, &generated.GetValidPreparationVesselsForVesselParams{
		ID:              vesselID,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		PageCursor:      database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid preparation vessels list retrieval query")
	}

	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}

	var (
		data          []*mealplanning.ValidPreparationVessel
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		validPreparationVessel := &mealplanning.ValidPreparationVessel{
			CreatedAt:     result.ValidPreparationVesselCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidPreparationVesselLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidPreparationVesselArchivedAt),
			ID:            result.ValidPreparationVesselID,
			Notes:         result.ValidPreparationVesselNotes,
			Vessel: mealplanning.ValidVessel{
				CreatedAt:     result.ValidVesselCreatedAt,
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidVesselArchivedAt),
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidVesselLastUpdatedAt),
				CapacityUnit: &mealplanning.ValidMeasurementUnit{
					CreatedAt:     result.ValidMeasurementUnitCreatedAt.Time,
					LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitLastUpdatedAt),
					ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitArchivedAt),
					Name:          result.ValidMeasurementUnitName.String,
					IconPath:      result.ValidMeasurementUnitIconPath.String,
					ID:            result.ValidMeasurementUnitID.String,
					Description:   result.ValidMeasurementUnitDescription.String,
					PluralName:    result.ValidMeasurementUnitPluralName.String,
					Slug:          result.ValidMeasurementUnitSlug.String,
					Volumetric:    result.ValidMeasurementUnitVolumetric.Bool,
					Universal:     result.ValidMeasurementUnitUniversal.Bool,
					Metric:        result.ValidMeasurementUnitMetric.Bool,
					Imperial:      result.ValidMeasurementUnitImperial.Bool,
				},
				IconPath:                       result.ValidVesselIconPath,
				PluralName:                     result.ValidVesselPluralName,
				Description:                    result.ValidVesselDescription,
				Name:                           result.ValidVesselName,
				Slug:                           result.ValidVesselSlug,
				Shape:                          string(result.ValidVesselShape),
				ID:                             result.ValidVesselID,
				WidthInMillimeters:             database.Float32FromNullString(result.ValidVesselWidthInMillimeters),
				LengthInMillimeters:            database.Float32FromNullString(result.ValidVesselLengthInMillimeters),
				HeightInMillimeters:            database.Float32FromNullString(result.ValidVesselHeightInMillimeters),
				Capacity:                       database.Float32FromString(result.ValidVesselCapacity),
				IncludeInGeneratedInstructions: result.ValidVesselIncludeInGeneratedInstructions,
				DisplayInSummaryLists:          result.ValidVesselDisplayInSummaryLists,
				UsableForStorage:               result.ValidVesselUsableForStorage,
			},
			Preparation: mealplanning.ValidPreparation{
				CreatedAt:                   result.ValidPreparationCreatedAt,
				MinInstrumentCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
				MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumInstrumentCount),
				MinIngredientCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
				MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumIngredientCount),
				MinVesselCount:              uint16(result.ValidPreparationMinimumVesselCount),
				MaxVesselCount:              database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumVesselCount),
				ArchivedAt:                  database.TimePointerFromNullTime(result.ValidPreparationArchivedAt),
				LastUpdatedAt:               database.TimePointerFromNullTime(result.ValidPreparationLastUpdatedAt),
				IconPath:                    result.ValidPreparationIconPath,
				PastTense:                   result.ValidPreparationPastTense,
				ID:                          result.ValidPreparationID,
				Name:                        result.ValidPreparationName,
				Description:                 result.ValidPreparationDescription,
				Slug:                        result.ValidPreparationSlug,
				RestrictToIngredients:       result.ValidPreparationRestrictToIngredients,
				TemperatureRequired:         result.ValidPreparationTemperatureRequired,
				TimeEstimateRequired:        result.ValidPreparationTimeEstimateRequired,
				ConditionExpressionRequired: result.ValidPreparationConditionExpressionRequired,
				ConsumesVessel:              result.ValidPreparationConsumesVessel,
				OnlyForVessels:              result.ValidPreparationOnlyForVessels,
				YieldsNothing:               result.ValidPreparationYieldsNothing,
			},
		}
		if validPreparationVessel.Vessel.CapacityUnit.ID == "" {
			validPreparationVessel.Vessel.CapacityUnit = nil
		}

		data = append(data, validPreparationVessel)
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vpv *mealplanning.ValidPreparationVessel) string { return vpv.ID },
		filter,
	)

	return x, nil
}

// GetValidPreparationVesselsByIDs fetches valid preparation vessels by their IDs from the database.
func (q *repository) GetValidPreparationVesselsByIDs(ctx context.Context, ids []string) (map[string]*mealplanning.ValidPreparationVessel, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if len(ids) == 0 {
		return map[string]*mealplanning.ValidPreparationVessel{}, nil
	}

	results, err := q.generatedQuerier.GetValidPreparationVesselsByIDs(ctx, q.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation vessels by IDs")
	}

	resultMap := make(map[string]*mealplanning.ValidPreparationVessel, len(results))
	for _, result := range results {
		validPreparationVessel := &mealplanning.ValidPreparationVessel{
			CreatedAt:     result.ValidPreparationVesselCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidPreparationVesselLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidPreparationVesselArchivedAt),
			ID:            result.ValidPreparationVesselID,
			Notes:         result.ValidPreparationVesselNotes,
			Vessel: mealplanning.ValidVessel{
				CreatedAt:     result.ValidVesselCreatedAt,
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidVesselArchivedAt),
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidVesselLastUpdatedAt),
				CapacityUnit: &mealplanning.ValidMeasurementUnit{
					CreatedAt:     result.ValidMeasurementUnitCreatedAt.Time,
					LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitLastUpdatedAt),
					ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitArchivedAt),
					Name:          result.ValidMeasurementUnitName.String,
					IconPath:      result.ValidMeasurementUnitIconPath.String,
					ID:            result.ValidMeasurementUnitID.String,
					Description:   result.ValidMeasurementUnitDescription.String,
					PluralName:    result.ValidMeasurementUnitPluralName.String,
					Slug:          result.ValidMeasurementUnitSlug.String,
					Volumetric:    result.ValidMeasurementUnitVolumetric.Bool,
					Universal:     result.ValidMeasurementUnitUniversal.Bool,
					Metric:        result.ValidMeasurementUnitMetric.Bool,
					Imperial:      result.ValidMeasurementUnitImperial.Bool,
				},
				IconPath:                       result.ValidVesselIconPath,
				PluralName:                     result.ValidVesselPluralName,
				Description:                    result.ValidVesselDescription,
				Name:                           result.ValidVesselName,
				Slug:                           result.ValidVesselSlug,
				Shape:                          string(result.ValidVesselShape),
				ID:                             result.ValidVesselID,
				WidthInMillimeters:             database.Float32FromNullString(result.ValidVesselWidthInMillimeters),
				LengthInMillimeters:            database.Float32FromNullString(result.ValidVesselLengthInMillimeters),
				HeightInMillimeters:            database.Float32FromNullString(result.ValidVesselHeightInMillimeters),
				Capacity:                       database.Float32FromString(result.ValidVesselCapacity),
				IncludeInGeneratedInstructions: result.ValidVesselIncludeInGeneratedInstructions,
				DisplayInSummaryLists:          result.ValidVesselDisplayInSummaryLists,
				UsableForStorage:               result.ValidVesselUsableForStorage,
			},
			Preparation: mealplanning.ValidPreparation{
				CreatedAt:                   result.ValidPreparationCreatedAt,
				MinInstrumentCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
				MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumInstrumentCount),
				MinIngredientCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
				MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumIngredientCount),
				MinVesselCount:              uint16(result.ValidPreparationMinimumVesselCount),
				MaxVesselCount:              database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumVesselCount),
				ArchivedAt:                  database.TimePointerFromNullTime(result.ValidPreparationArchivedAt),
				LastUpdatedAt:               database.TimePointerFromNullTime(result.ValidPreparationLastUpdatedAt),
				IconPath:                    result.ValidPreparationIconPath,
				PastTense:                   result.ValidPreparationPastTense,
				ID:                          result.ValidPreparationID,
				Name:                        result.ValidPreparationName,
				Description:                 result.ValidPreparationDescription,
				Slug:                        result.ValidPreparationSlug,
				RestrictToIngredients:       result.ValidPreparationRestrictToIngredients,
				TemperatureRequired:         result.ValidPreparationTemperatureRequired,
				TimeEstimateRequired:        result.ValidPreparationTimeEstimateRequired,
				ConditionExpressionRequired: result.ValidPreparationConditionExpressionRequired,
				ConsumesVessel:              result.ValidPreparationConsumesVessel,
				OnlyForVessels:              result.ValidPreparationOnlyForVessels,
				YieldsNothing:               result.ValidPreparationYieldsNothing,
			},
		}
		if validPreparationVessel.Vessel.CapacityUnit.ID == "" {
			validPreparationVessel.Vessel.CapacityUnit = nil
		}

		resultMap[result.ValidPreparationVesselID] = validPreparationVessel
	}

	return resultMap, nil
}

// CreateValidPreparationVessel creates a valid preparation vessel in the database.
func (q *repository) CreateValidPreparationVessel(ctx context.Context, input *mealplanning.ValidPreparationVesselDatabaseCreationInput) (*mealplanning.ValidPreparationVessel, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidPreparationVesselIDKey, input.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, input.ID)

	// create the valid preparation vessel.
	if err := q.withEvent(ctx, logger, mealplanning.ValidPreparationVesselCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidPreparationVesselIDKey: input.ID,
	}, func(tx database.Tx) error {
		return q.generatedQuerier.CreateValidPreparationVessel(ctx, tx, &generated.CreateValidPreparationVesselParams{
			ID:                 input.ID,
			Notes:              input.Notes,
			ValidPreparationID: input.ValidPreparationID,
			ValidVesselID:      input.ValidVesselID,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid preparation vessel creation query")
	}

	x := &mealplanning.ValidPreparationVessel{
		ID:          input.ID,
		Notes:       input.Notes,
		Preparation: mealplanning.ValidPreparation{ID: input.ValidPreparationID},
		Vessel:      mealplanning.ValidVessel{ID: input.ValidVesselID},
		CreatedAt:   q.CurrentTime(),
	}

	preparation, err := q.GetValidPreparation(ctx, input.ValidPreparationID)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation for valid preparation vessel")
	}
	if preparation != nil {
		x.Preparation = *preparation
	}

	vessel, err := q.GetValidVessel(ctx, input.ValidVesselID)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid vessel for valid preparation vessel")
	}
	if vessel != nil {
		x.Vessel = *vessel
	}

	logger.Info("valid preparation vessel created")

	return x, nil
}

// UpdateValidPreparationVessel updates a particular valid preparation vessel.
func (q *repository) UpdateValidPreparationVessel(ctx context.Context, updated *mealplanning.ValidPreparationVessel) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidPreparationVesselIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, mealplanning.ValidPreparationVesselUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidPreparationVesselIDKey: updated.ID,
	}, func(tx database.Tx) error {
		_, updateErr := q.generatedQuerier.UpdateValidPreparationVessel(ctx, tx, &generated.UpdateValidPreparationVesselParams{
			Notes:              updated.Notes,
			ValidPreparationID: updated.Preparation.ID,
			ValidVesselID:      updated.Vessel.ID,
			ID:                 updated.ID,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid preparation vessel")
	}

	logger.Info("valid preparation vessel updated")

	return nil
}

// ArchiveValidPreparationVessel archives a valid preparation vessel from the database by its ID.
func (q *repository) ArchiveValidPreparationVessel(ctx context.Context, validPreparationVesselID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if validPreparationVesselID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidPreparationVesselIDKey, validPreparationVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidVesselIDKey, validPreparationVesselID)

	return q.withEvent(ctx, logger, mealplanning.ValidPreparationVesselArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidPreparationVesselIDKey: validPreparationVesselID,
	}, func(tx database.Tx) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveValidPreparationVessel(ctx, tx, validPreparationVesselID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "updating valid preparation vessel")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}
