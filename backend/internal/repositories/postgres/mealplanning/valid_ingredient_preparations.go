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
	_ mealplanning.ValidIngredientPreparationDataManager = (*repository)(nil)
)

// ValidIngredientPreparationExists fetches whether a valid ingredient preparation exists from the database.
func (q *repository) ValidIngredientPreparationExists(ctx context.Context, validIngredientPreparationID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientPreparationID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientPreparationIDKey, validIngredientPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, validIngredientPreparationID)

	result, err := q.generatedQuerier.CheckValidIngredientPreparationExistence(ctx, q.readDB, validIngredientPreparationID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient preparation existence check")
	}

	return result, nil
}

// GetValidIngredientPreparation fetches a valid ingredient preparation from the database.
func (q *repository) GetValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string) (*mealplanning.ValidIngredientPreparation, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientPreparationID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientPreparationIDKey, validIngredientPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, validIngredientPreparationID)

	result, err := q.generatedQuerier.GetValidIngredientPreparation(ctx, q.readDB, validIngredientPreparationID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient preparation retrieval")
	}

	validIngredientPreparation := &mealplanning.ValidIngredientPreparation{
		CreatedAt:     result.ValidIngredientPreparationCreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientPreparationLastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientPreparationArchivedAt),
		Notes:         result.ValidIngredientPreparationNotes,
		ID:            result.ValidIngredientPreparationID,
		Ingredient: mealplanning.ValidIngredient{
			CreatedAt:                      result.ValidIngredientCreatedAt,
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidIngredientLastUpdatedAt),
			ArchivedAt:                     database.TimePointerFromNullTime(result.ValidIngredientArchivedAt),
			MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMinimumIdealStorageTemperatureInCelsius),
			MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMaximumIdealStorageTemperatureInCelsius),
			IconPath:                       result.ValidIngredientIconPath,
			Warning:                        result.ValidIngredientWarning,
			PluralName:                     result.ValidIngredientPluralName,
			StorageInstructions:            result.ValidIngredientStorageInstructions,
			Name:                           result.ValidIngredientName,
			ID:                             result.ValidIngredientID,
			Description:                    result.ValidIngredientDescription,
			Slug:                           result.ValidIngredientSlug,
			ShoppingSuggestions:            result.ValidIngredientShoppingSuggestions,
			ContainsShellfish:              result.ValidIngredientContainsShellfish,
			IsLiquid:                       database.BoolFromNullBool(result.ValidIngredientIsLiquid),
			ContainsPeanut:                 result.ValidIngredientContainsPeanut,
			ContainsTreeNut:                result.ValidIngredientContainsTreeNut,
			ContainsEgg:                    result.ValidIngredientContainsEgg,
			ContainsWheat:                  result.ValidIngredientContainsWheat,
			ContainsSoy:                    result.ValidIngredientContainsSoy,
			AnimalDerived:                  result.ValidIngredientAnimalDerived,
			RestrictToPreparations:         result.ValidIngredientRestrictToPreparations,
			ContaminatesEquipment:          result.ValidIngredientContaminatesEquipment,
			ContainsSesame:                 result.ValidIngredientContainsSesame,
			ContainsFish:                   result.ValidIngredientContainsFish,
			ContainsGluten:                 result.ValidIngredientContainsGluten,
			ContainsDairy:                  result.ValidIngredientContainsDairy,
			ContainsAlcohol:                result.ValidIngredientContainsAlcohol,
			AnimalFlesh:                    result.ValidIngredientAnimalFlesh,
			IsStarch:                       result.ValidIngredientIsStarch,
			IsProtein:                      result.ValidIngredientIsProtein,
			IsGrain:                        result.ValidIngredientIsGrain,
			IsFruit:                        result.ValidIngredientIsFruit,
			IsSalt:                         result.ValidIngredientIsSalt,
			IsFat:                          result.ValidIngredientIsFat,
			IsAcid:                         result.ValidIngredientIsAcid,
			IsHeat:                         result.ValidIngredientIsHeat,
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

	return validIngredientPreparation, nil
}

// GetValidIngredientPreparations fetches a list of valid ingredient preparations from the database that meet a particular filter.
func (q *repository) GetValidIngredientPreparations(ctx context.Context, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetValidIngredientPreparations(ctx, q.readDB, &generated.GetValidIngredientPreparationsParams{
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredient preparations list retrieval query")
	}

	var (
		data          []*mealplanning.ValidIngredientPreparation
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &mealplanning.ValidIngredientPreparation{
			CreatedAt:     result.ValidIngredientPreparationCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientPreparationLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientPreparationArchivedAt),
			Notes:         result.ValidIngredientPreparationNotes,
			ID:            result.ValidIngredientPreparationID,
			Ingredient: mealplanning.ValidIngredient{
				CreatedAt:                      result.ValidIngredientCreatedAt,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidIngredientLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidIngredientArchivedAt),
				MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMinimumIdealStorageTemperatureInCelsius),
				MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMaximumIdealStorageTemperatureInCelsius),
				IconPath:                       result.ValidIngredientIconPath,
				Warning:                        result.ValidIngredientWarning,
				PluralName:                     result.ValidIngredientPluralName,
				StorageInstructions:            result.ValidIngredientStorageInstructions,
				Name:                           result.ValidIngredientName,
				ID:                             result.ValidIngredientID,
				Description:                    result.ValidIngredientDescription,
				Slug:                           result.ValidIngredientSlug,
				ShoppingSuggestions:            result.ValidIngredientShoppingSuggestions,
				ContainsShellfish:              result.ValidIngredientContainsShellfish,
				IsLiquid:                       database.BoolFromNullBool(result.ValidIngredientIsLiquid),
				ContainsPeanut:                 result.ValidIngredientContainsPeanut,
				ContainsTreeNut:                result.ValidIngredientContainsTreeNut,
				ContainsEgg:                    result.ValidIngredientContainsEgg,
				ContainsWheat:                  result.ValidIngredientContainsWheat,
				ContainsSoy:                    result.ValidIngredientContainsSoy,
				AnimalDerived:                  result.ValidIngredientAnimalDerived,
				RestrictToPreparations:         result.ValidIngredientRestrictToPreparations,
				ContaminatesEquipment:          result.ValidIngredientContaminatesEquipment,
				ContainsSesame:                 result.ValidIngredientContainsSesame,
				ContainsFish:                   result.ValidIngredientContainsFish,
				ContainsGluten:                 result.ValidIngredientContainsGluten,
				ContainsDairy:                  result.ValidIngredientContainsDairy,
				ContainsAlcohol:                result.ValidIngredientContainsAlcohol,
				AnimalFlesh:                    result.ValidIngredientAnimalFlesh,
				IsStarch:                       result.ValidIngredientIsStarch,
				IsProtein:                      result.ValidIngredientIsProtein,
				IsGrain:                        result.ValidIngredientIsGrain,
				IsFruit:                        result.ValidIngredientIsFruit,
				IsSalt:                         result.ValidIngredientIsSalt,
				IsFat:                          result.ValidIngredientIsFat,
				IsAcid:                         result.ValidIngredientIsAcid,
				IsHeat:                         result.ValidIngredientIsHeat,
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
		func(vip *mealplanning.ValidIngredientPreparation) string { return vip.ID },
		filter,
	)

	return x, nil
}

// GetValidIngredientPreparationsForPreparation fetches a list of valid ingredient preparations from the database that meet a particular filter.
func (q *repository) GetValidIngredientPreparationsForPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if preparationID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidPreparationIDKey, preparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, preparationID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetValidIngredientPreparationsForPreparation(ctx, q.readDB, &generated.GetValidIngredientPreparationsForPreparationParams{
		ID:              preparationID,
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredient preparations list retrieval query")
	}

	var (
		data          []*mealplanning.ValidIngredientPreparation
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &mealplanning.ValidIngredientPreparation{
			CreatedAt:     result.ValidIngredientPreparationCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientPreparationLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientPreparationArchivedAt),
			Notes:         result.ValidIngredientPreparationNotes,
			ID:            result.ValidIngredientPreparationID,
			Ingredient: mealplanning.ValidIngredient{
				CreatedAt:                      result.ValidIngredientCreatedAt,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidIngredientLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidIngredientArchivedAt),
				MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMinimumIdealStorageTemperatureInCelsius),
				MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMaximumIdealStorageTemperatureInCelsius),
				IconPath:                       result.ValidIngredientIconPath,
				Warning:                        result.ValidIngredientWarning,
				PluralName:                     result.ValidIngredientPluralName,
				StorageInstructions:            result.ValidIngredientStorageInstructions,
				Name:                           result.ValidIngredientName,
				ID:                             result.ValidIngredientID,
				Description:                    result.ValidIngredientDescription,
				Slug:                           result.ValidIngredientSlug,
				ShoppingSuggestions:            result.ValidIngredientShoppingSuggestions,
				ContainsShellfish:              result.ValidIngredientContainsShellfish,
				IsLiquid:                       database.BoolFromNullBool(result.ValidIngredientIsLiquid),
				ContainsPeanut:                 result.ValidIngredientContainsPeanut,
				ContainsTreeNut:                result.ValidIngredientContainsTreeNut,
				ContainsEgg:                    result.ValidIngredientContainsEgg,
				ContainsWheat:                  result.ValidIngredientContainsWheat,
				ContainsSoy:                    result.ValidIngredientContainsSoy,
				AnimalDerived:                  result.ValidIngredientAnimalDerived,
				RestrictToPreparations:         result.ValidIngredientRestrictToPreparations,
				ContaminatesEquipment:          result.ValidIngredientContaminatesEquipment,
				ContainsSesame:                 result.ValidIngredientContainsSesame,
				ContainsFish:                   result.ValidIngredientContainsFish,
				ContainsGluten:                 result.ValidIngredientContainsGluten,
				ContainsDairy:                  result.ValidIngredientContainsDairy,
				ContainsAlcohol:                result.ValidIngredientContainsAlcohol,
				AnimalFlesh:                    result.ValidIngredientAnimalFlesh,
				IsStarch:                       result.ValidIngredientIsStarch,
				IsProtein:                      result.ValidIngredientIsProtein,
				IsGrain:                        result.ValidIngredientIsGrain,
				IsFruit:                        result.ValidIngredientIsFruit,
				IsSalt:                         result.ValidIngredientIsSalt,
				IsFat:                          result.ValidIngredientIsFat,
				IsAcid:                         result.ValidIngredientIsAcid,
				IsHeat:                         result.ValidIngredientIsHeat,
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
		func(vip *mealplanning.ValidIngredientPreparation) string { return vip.ID },
		filter,
	)

	return x, nil
}

// GetValidIngredientPreparationsForIngredient fetches a list of valid ingredient preparations from the database that meet a particular filter.
func (q *repository) GetValidIngredientPreparationsForIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if ingredientID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientIDKey, ingredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, ingredientID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetValidIngredientPreparationsForIngredient(ctx, q.readDB, &generated.GetValidIngredientPreparationsForIngredientParams{
		ID:              ingredientID,
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredient preparations list retrieval query")
	}

	var (
		data          []*mealplanning.ValidIngredientPreparation
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &mealplanning.ValidIngredientPreparation{
			CreatedAt:     result.ValidIngredientPreparationCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientPreparationLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientPreparationArchivedAt),
			Notes:         result.ValidIngredientPreparationNotes,
			ID:            result.ValidIngredientPreparationID,
			Ingredient: mealplanning.ValidIngredient{
				CreatedAt:                      result.ValidIngredientCreatedAt,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidIngredientLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidIngredientArchivedAt),
				MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMinimumIdealStorageTemperatureInCelsius),
				MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMaximumIdealStorageTemperatureInCelsius),
				IconPath:                       result.ValidIngredientIconPath,
				Warning:                        result.ValidIngredientWarning,
				PluralName:                     result.ValidIngredientPluralName,
				StorageInstructions:            result.ValidIngredientStorageInstructions,
				Name:                           result.ValidIngredientName,
				ID:                             result.ValidIngredientID,
				Description:                    result.ValidIngredientDescription,
				Slug:                           result.ValidIngredientSlug,
				ShoppingSuggestions:            result.ValidIngredientShoppingSuggestions,
				ContainsShellfish:              result.ValidIngredientContainsShellfish,
				IsLiquid:                       database.BoolFromNullBool(result.ValidIngredientIsLiquid),
				ContainsPeanut:                 result.ValidIngredientContainsPeanut,
				ContainsTreeNut:                result.ValidIngredientContainsTreeNut,
				ContainsEgg:                    result.ValidIngredientContainsEgg,
				ContainsWheat:                  result.ValidIngredientContainsWheat,
				ContainsSoy:                    result.ValidIngredientContainsSoy,
				AnimalDerived:                  result.ValidIngredientAnimalDerived,
				RestrictToPreparations:         result.ValidIngredientRestrictToPreparations,
				ContaminatesEquipment:          result.ValidIngredientContaminatesEquipment,
				ContainsSesame:                 result.ValidIngredientContainsSesame,
				ContainsFish:                   result.ValidIngredientContainsFish,
				ContainsGluten:                 result.ValidIngredientContainsGluten,
				ContainsDairy:                  result.ValidIngredientContainsDairy,
				ContainsAlcohol:                result.ValidIngredientContainsAlcohol,
				AnimalFlesh:                    result.ValidIngredientAnimalFlesh,
				IsStarch:                       result.ValidIngredientIsStarch,
				IsProtein:                      result.ValidIngredientIsProtein,
				IsGrain:                        result.ValidIngredientIsGrain,
				IsFruit:                        result.ValidIngredientIsFruit,
				IsSalt:                         result.ValidIngredientIsSalt,
				IsFat:                          result.ValidIngredientIsFat,
				IsAcid:                         result.ValidIngredientIsAcid,
				IsHeat:                         result.ValidIngredientIsHeat,
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
		func(vip *mealplanning.ValidIngredientPreparation) string { return vip.ID },
		filter,
	)

	return x, nil
}

// GetValidIngredientPreparationsByIDs fetches valid ingredient preparations by their IDs from the database.
func (q *repository) GetValidIngredientPreparationsByIDs(ctx context.Context, ids []string) (map[string]*mealplanning.ValidIngredientPreparation, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if len(ids) == 0 {
		return map[string]*mealplanning.ValidIngredientPreparation{}, nil
	}

	results, err := q.generatedQuerier.GetValidIngredientPreparationsByIDs(ctx, q.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient preparations by IDs")
	}

	resultMap := make(map[string]*mealplanning.ValidIngredientPreparation, len(results))
	for _, result := range results {
		resultMap[result.ValidIngredientPreparationID] = &mealplanning.ValidIngredientPreparation{
			CreatedAt:     result.ValidIngredientPreparationCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientPreparationLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientPreparationArchivedAt),
			Notes:         result.ValidIngredientPreparationNotes,
			ID:            result.ValidIngredientPreparationID,
			Ingredient: mealplanning.ValidIngredient{
				CreatedAt:                      result.ValidIngredientCreatedAt,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidIngredientLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidIngredientArchivedAt),
				MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMinimumIdealStorageTemperatureInCelsius),
				MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMaximumIdealStorageTemperatureInCelsius),
				IconPath:                       result.ValidIngredientIconPath,
				Warning:                        result.ValidIngredientWarning,
				PluralName:                     result.ValidIngredientPluralName,
				StorageInstructions:            result.ValidIngredientStorageInstructions,
				Name:                           result.ValidIngredientName,
				ID:                             result.ValidIngredientID,
				Description:                    result.ValidIngredientDescription,
				Slug:                           result.ValidIngredientSlug,
				ShoppingSuggestions:            result.ValidIngredientShoppingSuggestions,
				ContainsShellfish:              result.ValidIngredientContainsShellfish,
				IsLiquid:                       database.BoolFromNullBool(result.ValidIngredientIsLiquid),
				ContainsPeanut:                 result.ValidIngredientContainsPeanut,
				ContainsTreeNut:                result.ValidIngredientContainsTreeNut,
				ContainsEgg:                    result.ValidIngredientContainsEgg,
				ContainsWheat:                  result.ValidIngredientContainsWheat,
				ContainsSoy:                    result.ValidIngredientContainsSoy,
				AnimalDerived:                  result.ValidIngredientAnimalDerived,
				RestrictToPreparations:         result.ValidIngredientRestrictToPreparations,
				ContaminatesEquipment:          result.ValidIngredientContaminatesEquipment,
				ContainsSesame:                 result.ValidIngredientContainsSesame,
				ContainsFish:                   result.ValidIngredientContainsFish,
				ContainsGluten:                 result.ValidIngredientContainsGluten,
				ContainsDairy:                  result.ValidIngredientContainsDairy,
				ContainsAlcohol:                result.ValidIngredientContainsAlcohol,
				AnimalFlesh:                    result.ValidIngredientAnimalFlesh,
				IsStarch:                       result.ValidIngredientIsStarch,
				IsProtein:                      result.ValidIngredientIsProtein,
				IsGrain:                        result.ValidIngredientIsGrain,
				IsFruit:                        result.ValidIngredientIsFruit,
				IsSalt:                         result.ValidIngredientIsSalt,
				IsFat:                          result.ValidIngredientIsFat,
				IsAcid:                         result.ValidIngredientIsAcid,
				IsHeat:                         result.ValidIngredientIsHeat,
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

// CreateValidIngredientPreparation creates a valid ingredient preparation in the database.
func (q *repository) CreateValidIngredientPreparation(ctx context.Context, input *mealplanning.ValidIngredientPreparationDatabaseCreationInput) (*mealplanning.ValidIngredientPreparation, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientPreparationIDKey, input.ID)

	// create the valid ingredient preparation.
	if err := q.withEvent(ctx, logger, mealplanning.ValidIngredientPreparationCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientPreparationIDKey: input.ID,
	}, func(tx database.Tx) error {
		return q.generatedQuerier.CreateValidIngredientPreparation(ctx, tx, &generated.CreateValidIngredientPreparationParams{
			ID:                 input.ID,
			Notes:              input.Notes,
			ValidPreparationID: input.ValidPreparationID,
			ValidIngredientID:  input.ValidIngredientID,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient preparation creation query")
	}

	x := &mealplanning.ValidIngredientPreparation{
		ID:          input.ID,
		Notes:       input.Notes,
		Preparation: mealplanning.ValidPreparation{ID: input.ValidPreparationID},
		Ingredient:  mealplanning.ValidIngredient{ID: input.ValidIngredientID},
		CreatedAt:   q.CurrentTime(),
	}

	preparation, err := q.GetValidPreparation(ctx, input.ValidPreparationID)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation for valid ingredient preparation")
	}
	if preparation != nil {
		x.Preparation = *preparation
	}

	ingredient, err := q.GetValidIngredient(ctx, input.ValidIngredientID)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient for valid ingredient preparation")
	}
	if ingredient != nil {
		x.Ingredient = *ingredient
	}

	return x, nil
}

// UpdateValidIngredientPreparation updates a particular valid ingredient preparation.
func (q *repository) UpdateValidIngredientPreparation(ctx context.Context, updated *mealplanning.ValidIngredientPreparation) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientPreparationIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, mealplanning.ValidIngredientPreparationUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientPreparationIDKey: updated.ID,
	}, func(tx database.Tx) error {
		_, updateErr := q.generatedQuerier.UpdateValidIngredientPreparation(ctx, tx, &generated.UpdateValidIngredientPreparationParams{
			Notes:              updated.Notes,
			ValidPreparationID: updated.Preparation.ID,
			ValidIngredientID:  updated.Ingredient.ID,
			ID:                 updated.ID,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid ingredient preparation")
	}

	logger.Info("valid ingredient preparation updated")

	return nil
}

// ArchiveValidIngredientPreparation archives a valid ingredient preparation from the database by its ID.
func (q *repository) ArchiveValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientPreparationID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientPreparationIDKey, validIngredientPreparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, validIngredientPreparationID)

	return q.withEvent(ctx, logger, mealplanning.ValidIngredientPreparationArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientPreparationIDKey: validIngredientPreparationID,
	}, func(tx database.Tx) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveValidIngredientPreparation(ctx, tx, validIngredientPreparationID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving valid ingredient preparation")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}
