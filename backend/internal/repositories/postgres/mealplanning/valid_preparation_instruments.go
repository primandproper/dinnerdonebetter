package mealplanning

import (
	"context"
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v6/database"
	platformerrors "github.com/primandproper/platform-go/v6/errors"
	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/observability"
	"github.com/primandproper/platform-go/v6/observability/tracing"
)

var (
	_ mealplanning.ValidPreparationInstrumentDataManager = (*repository)(nil)
)

// ValidPreparationInstrumentExists fetches whether a valid preparation instrument exists from the database.
func (q *repository) ValidPreparationInstrumentExists(ctx context.Context, validPreparationInstrumentID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if validPreparationInstrumentID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, validPreparationInstrumentID)

	result, err := q.generatedQuerier.CheckValidPreparationInstrumentExistence(ctx, q.readDB, validPreparationInstrumentID)
	if err != nil {
		return false, observability.PrepareError(err, span, "checking valid preparation instrument existence")
	}

	return result, nil
}

// GetValidPreparationInstrument fetches a valid preparation instrument from the database.
func (q *repository) GetValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string) (*mealplanning.ValidPreparationInstrument, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if validPreparationInstrumentID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, validPreparationInstrumentID)

	result, err := q.generatedQuerier.GetValidPreparationInstrument(ctx, q.readDB, validPreparationInstrumentID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "getting valid preparation instrument")
	}

	x := &mealplanning.ValidPreparationInstrument{
		CreatedAt:     result.ValidPreparationInstrumentCreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.ValidPreparationInstrumentLastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ValidPreparationInstrumentArchivedAt),
		ID:            result.ValidPreparationInstrumentID,
		Notes:         result.ValidPreparationInstrumentNotes,
		Instrument: mealplanning.ValidInstrument{
			CreatedAt:                      result.ValidInstrumentCreatedAt,
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidInstrumentLastUpdatedAt),
			ArchivedAt:                     database.TimePointerFromNullTime(result.ValidInstrumentArchivedAt),
			IconPath:                       result.ValidInstrumentIconPath,
			ID:                             result.ValidInstrumentID,
			Name:                           result.ValidInstrumentName,
			PluralName:                     result.ValidInstrumentPluralName,
			Description:                    result.ValidInstrumentDescription,
			Slug:                           result.ValidInstrumentSlug,
			DisplayInSummaryLists:          result.ValidInstrumentDisplayInSummaryLists,
			IncludeInGeneratedInstructions: result.ValidInstrumentIncludeInGeneratedInstructions,
			UsableForStorage:               result.ValidInstrumentUsableForStorage,
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

	return x, nil
}

// GetValidPreparationInstruments fetches a list of valid preparation instruments from the database that meet a particular filter.
func (q *repository) GetValidPreparationInstruments(ctx context.Context, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidPreparationInstruments(ctx, q.readDB, &generated.GetValidPreparationInstrumentsParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid preparation instruments list retrieval query")
	}

	var (
		data          []*mealplanning.ValidPreparationInstrument
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &mealplanning.ValidPreparationInstrument{
			CreatedAt:     result.ValidPreparationInstrumentCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidPreparationInstrumentLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidPreparationInstrumentArchivedAt),
			ID:            result.ValidPreparationInstrumentID,
			Notes:         result.ValidPreparationInstrumentNotes,
			Instrument: mealplanning.ValidInstrument{
				CreatedAt:                      result.ValidInstrumentCreatedAt,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidInstrumentLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidInstrumentArchivedAt),
				IconPath:                       result.ValidInstrumentIconPath,
				ID:                             result.ValidInstrumentID,
				Name:                           result.ValidInstrumentName,
				PluralName:                     result.ValidInstrumentPluralName,
				Description:                    result.ValidInstrumentDescription,
				Slug:                           result.ValidInstrumentSlug,
				DisplayInSummaryLists:          result.ValidInstrumentDisplayInSummaryLists,
				IncludeInGeneratedInstructions: result.ValidInstrumentIncludeInGeneratedInstructions,
				UsableForStorage:               result.ValidInstrumentUsableForStorage,
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
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vpi *mealplanning.ValidPreparationInstrument) string { return vpi.ID },
		filter,
	)

	return x, nil
}

// GetValidPreparationInstrumentsForPreparation fetches a list of valid preparation instruments from the database that meet a particular filter.
func (q *repository) GetValidPreparationInstrumentsForPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if preparationID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, preparationID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidPreparationInstrumentsForPreparation(ctx, q.readDB, &generated.GetValidPreparationInstrumentsForPreparationParams{
		ID:              preparationID,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid preparation instruments list retrieval query")
	}

	var (
		data          []*mealplanning.ValidPreparationInstrument
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &mealplanning.ValidPreparationInstrument{
			CreatedAt:     result.ValidPreparationInstrumentCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidPreparationInstrumentLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidPreparationInstrumentArchivedAt),
			ID:            result.ValidPreparationInstrumentID,
			Notes:         result.ValidPreparationInstrumentNotes,
			Instrument: mealplanning.ValidInstrument{
				CreatedAt:                      result.ValidInstrumentCreatedAt,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidInstrumentLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidInstrumentArchivedAt),
				IconPath:                       result.ValidInstrumentIconPath,
				ID:                             result.ValidInstrumentID,
				Name:                           result.ValidInstrumentName,
				PluralName:                     result.ValidInstrumentPluralName,
				Description:                    result.ValidInstrumentDescription,
				Slug:                           result.ValidInstrumentSlug,
				DisplayInSummaryLists:          result.ValidInstrumentDisplayInSummaryLists,
				IncludeInGeneratedInstructions: result.ValidInstrumentIncludeInGeneratedInstructions,
				UsableForStorage:               result.ValidInstrumentUsableForStorage,
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
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vpi *mealplanning.ValidPreparationInstrument) string { return vpi.ID },
		filter,
	)

	return x, nil
}

// GetValidPreparationInstrumentsForInstrument fetches a list of valid preparation instruments from the database that meet a particular filter.
func (q *repository) GetValidPreparationInstrumentsForInstrument(ctx context.Context, instrumentID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if instrumentID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, instrumentID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidPreparationInstrumentsForInstrument(ctx, q.readDB, &generated.GetValidPreparationInstrumentsForInstrumentParams{
		ID:              instrumentID,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid preparation instruments list retrieval query")
	}

	var (
		data          []*mealplanning.ValidPreparationInstrument
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &mealplanning.ValidPreparationInstrument{
			CreatedAt:     result.ValidPreparationInstrumentCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidPreparationInstrumentLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidPreparationInstrumentArchivedAt),
			ID:            result.ValidPreparationInstrumentID,
			Notes:         result.ValidPreparationInstrumentNotes,
			Instrument: mealplanning.ValidInstrument{
				CreatedAt:                      result.ValidInstrumentCreatedAt,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidInstrumentLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidInstrumentArchivedAt),
				IconPath:                       result.ValidInstrumentIconPath,
				ID:                             result.ValidInstrumentID,
				Name:                           result.ValidInstrumentName,
				PluralName:                     result.ValidInstrumentPluralName,
				Description:                    result.ValidInstrumentDescription,
				Slug:                           result.ValidInstrumentSlug,
				DisplayInSummaryLists:          result.ValidInstrumentDisplayInSummaryLists,
				IncludeInGeneratedInstructions: result.ValidInstrumentIncludeInGeneratedInstructions,
				UsableForStorage:               result.ValidInstrumentUsableForStorage,
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
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vpi *mealplanning.ValidPreparationInstrument) string { return vpi.ID },
		filter,
	)

	return x, nil
}

// GetValidPreparationInstrumentsByIDs fetches valid preparation instruments by their IDs from the database.
func (q *repository) GetValidPreparationInstrumentsByIDs(ctx context.Context, ids []string) (map[string]*mealplanning.ValidPreparationInstrument, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if len(ids) == 0 {
		return map[string]*mealplanning.ValidPreparationInstrument{}, nil
	}

	results, err := q.generatedQuerier.GetValidPreparationInstrumentsByIDs(ctx, q.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation instruments by IDs")
	}

	resultMap := make(map[string]*mealplanning.ValidPreparationInstrument, len(results))
	for _, result := range results {
		resultMap[result.ValidPreparationInstrumentID] = &mealplanning.ValidPreparationInstrument{
			CreatedAt:     result.ValidPreparationInstrumentCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidPreparationInstrumentLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidPreparationInstrumentArchivedAt),
			ID:            result.ValidPreparationInstrumentID,
			Notes:         result.ValidPreparationInstrumentNotes,
			Instrument: mealplanning.ValidInstrument{
				CreatedAt:                      result.ValidInstrumentCreatedAt,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidInstrumentLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidInstrumentArchivedAt),
				IconPath:                       result.ValidInstrumentIconPath,
				ID:                             result.ValidInstrumentID,
				Name:                           result.ValidInstrumentName,
				PluralName:                     result.ValidInstrumentPluralName,
				Description:                    result.ValidInstrumentDescription,
				Slug:                           result.ValidInstrumentSlug,
				DisplayInSummaryLists:          result.ValidInstrumentDisplayInSummaryLists,
				IncludeInGeneratedInstructions: result.ValidInstrumentIncludeInGeneratedInstructions,
				UsableForStorage:               result.ValidInstrumentUsableForStorage,
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
	}

	return resultMap, nil
}

// CreateValidPreparationInstrument creates a valid preparation instrument in the database.
func (q *repository) CreateValidPreparationInstrument(ctx context.Context, input *mealplanning.ValidPreparationInstrumentDatabaseCreationInput) (*mealplanning.ValidPreparationInstrument, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.ValidPreparationInstrumentIDKey, input.ID)

	// create the valid preparation instrument.
	if err := q.generatedQuerier.CreateValidPreparationInstrument(ctx, q.writeDB, &generated.CreateValidPreparationInstrumentParams{
		ID:                 input.ID,
		Notes:              input.Notes,
		ValidPreparationID: input.ValidPreparationID,
		ValidInstrumentID:  input.ValidInstrumentID,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid preparation instrument creation query")
	}

	x := &mealplanning.ValidPreparationInstrument{
		ID:          input.ID,
		Notes:       input.Notes,
		Preparation: mealplanning.ValidPreparation{ID: input.ValidPreparationID},
		Instrument:  mealplanning.ValidInstrument{ID: input.ValidInstrumentID},
		CreatedAt:   q.CurrentTime(),
	}

	preparation, err := q.GetValidPreparation(ctx, input.ValidPreparationID)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation for valid preparation instrument")
	}
	if preparation != nil {
		x.Preparation = *preparation
	}

	instrument, err := q.GetValidInstrument(ctx, input.ValidInstrumentID)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid instrument for valid preparation instrument")
	}
	if instrument != nil {
		x.Instrument = *instrument
	}

	logger.Info("valid preparation instrument created")

	return x, nil
}

// UpdateValidPreparationInstrument updates a particular valid preparation instrument.
func (q *repository) UpdateValidPreparationInstrument(ctx context.Context, updated *mealplanning.ValidPreparationInstrument) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}

	logger := q.logger.WithValue(mealplanningkeys.ValidPreparationInstrumentIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, updated.ID)

	if _, err := q.generatedQuerier.UpdateValidPreparationInstrument(ctx, q.writeDB, &generated.UpdateValidPreparationInstrumentParams{
		Notes:              updated.Notes,
		ValidPreparationID: updated.Preparation.ID,
		ValidInstrumentID:  updated.Instrument.ID,
		ID:                 updated.ID,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid preparation instrument")
	}

	logger.Info("valid preparation instrument updated")

	return nil
}

// ArchiveValidPreparationInstrument archives a valid preparation instrument from the database by its ID.
func (q *repository) ArchiveValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if validPreparationInstrumentID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidPreparationInstrumentIDKey, validPreparationInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, validPreparationInstrumentID)

	rowsAffected, err := q.generatedQuerier.ArchiveValidPreparationInstrument(ctx, q.writeDB, validPreparationInstrumentID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid preparation instrument")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
