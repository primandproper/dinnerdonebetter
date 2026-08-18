package mealplanning

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v11/database"
	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/observability"
	"github.com/primandproper/platform-go/v11/observability/tracing"
)

var (
	_ mealplanning.ValidIngredientMeasurementUnitDataManager = (*repository)(nil)
)

// ValidIngredientMeasurementUnitExists fetches whether a valid ingredient measurement unit exists from the database.
func (q *repository) ValidIngredientMeasurementUnitExists(ctx context.Context, validIngredientMeasurementUnitID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientMeasurementUnitID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientMeasurementUnitIDKey, validIngredientMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientMeasurementUnitIDKey, validIngredientMeasurementUnitID)

	result, err := q.generatedQuerier.CheckValidIngredientMeasurementUnitExistence(ctx, q.readDB, validIngredientMeasurementUnitID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient measurement unit existence check")
	}

	return result, nil
}

// GetValidIngredientMeasurementUnit fetches a valid ingredient measurement unit from the database.
func (q *repository) GetValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string) (*mealplanning.ValidIngredientMeasurementUnit, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientMeasurementUnitID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientMeasurementUnitIDKey, validIngredientMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientMeasurementUnitIDKey, validIngredientMeasurementUnitID)

	result, err := q.generatedQuerier.GetValidIngredientMeasurementUnit(ctx, q.readDB, validIngredientMeasurementUnitID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "scanning validIngredientMeasurementUnit")
	}

	validIngredientMeasurementUnit := &mealplanning.ValidIngredientMeasurementUnit{
		CreatedAt:            result.ValidIngredientMeasurementUnitCreatedAt,
		LastUpdatedAt:        database.TimePointerFromNullTime(result.ValidIngredientMeasurementUnitLastUpdatedAt),
		ArchivedAt:           database.TimePointerFromNullTime(result.ValidIngredientMeasurementUnitArchivedAt),
		MinAllowableQuantity: database.Float32FromString(result.ValidIngredientMeasurementUnitMinimumAllowableQuantity),
		MaxAllowableQuantity: database.Float32PointerFromNullString(result.ValidIngredientMeasurementUnitMaximumAllowableQuantity),
		Notes:                result.ValidIngredientMeasurementUnitNotes,
		ID:                   result.ValidIngredientMeasurementUnitID,
		MeasurementUnit: mealplanning.ValidMeasurementUnit{
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
	}

	return validIngredientMeasurementUnit, nil
}

// GetValidIngredientMeasurementUnitsForIngredient fetches a list of valid measurement units from the database that belong to a given ingredient ID.
func (q *repository) GetValidIngredientMeasurementUnitsForIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if ingredientID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientIDKey, ingredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, ingredientID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidIngredientMeasurementUnitsForIngredient(ctx, q.readDB, &generated.GetValidIngredientMeasurementUnitsForIngredientParams{
		ValidIngredientID: ingredientID,
		CreatedBefore:     database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:      database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:     database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:      database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:            database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:       database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived:   database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredient measurement units list retrieval query")
	}

	var (
		data          []*mealplanning.ValidIngredientMeasurementUnit
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &mealplanning.ValidIngredientMeasurementUnit{
			CreatedAt:     result.ValidIngredientMeasurementUnitCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientMeasurementUnitLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientMeasurementUnitArchivedAt),
			Notes:         result.ValidIngredientMeasurementUnitNotes,
			ID:            result.ValidIngredientMeasurementUnitID,
			MeasurementUnit: mealplanning.ValidMeasurementUnit{
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
			MinAllowableQuantity: database.Float32FromString(result.ValidIngredientMeasurementUnitMinimumAllowableQuantity),
			MaxAllowableQuantity: database.Float32PointerFromNullString(result.ValidIngredientMeasurementUnitMaximumAllowableQuantity),
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vimu *mealplanning.ValidIngredientMeasurementUnit) string { return vimu.ID },
		filter,
	)

	return x, nil
}

// GetValidIngredientMeasurementUnitsForMeasurementUnit fetches a list of valid measurement units from the database that belong to a given ingredient ID.
func (q *repository) GetValidIngredientMeasurementUnitsForMeasurementUnit(ctx context.Context, validMeasurementUnitID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validMeasurementUnitID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationInstrumentIDKey, validMeasurementUnitID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidIngredientMeasurementUnitsForMeasurementUnit(ctx, q.readDB, &generated.GetValidIngredientMeasurementUnitsForMeasurementUnitParams{
		ValidMeasurementUnitID: validMeasurementUnitID,
		CreatedBefore:          database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:           database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:          database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:           database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:                 database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:            database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived:        database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredient measurement units list retrieval query")
	}

	var (
		data          []*mealplanning.ValidIngredientMeasurementUnit
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &mealplanning.ValidIngredientMeasurementUnit{
			CreatedAt:     result.ValidIngredientMeasurementUnitCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientMeasurementUnitLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientMeasurementUnitArchivedAt),
			Notes:         result.ValidIngredientMeasurementUnitNotes,
			ID:            result.ValidIngredientMeasurementUnitID,
			MeasurementUnit: mealplanning.ValidMeasurementUnit{
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
			MinAllowableQuantity: database.Float32FromString(result.ValidIngredientMeasurementUnitMinimumAllowableQuantity),
			MaxAllowableQuantity: database.Float32PointerFromNullString(result.ValidIngredientMeasurementUnitMaximumAllowableQuantity),
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vimu *mealplanning.ValidIngredientMeasurementUnit) string { return vimu.ID },
		filter,
	)

	return x, nil
}

// GetValidIngredientMeasurementUnits fetches a list of valid ingredient measurement units from the database that meet a particular filter.
func (q *repository) GetValidIngredientMeasurementUnits(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidIngredientMeasurementUnits(ctx, q.readDB, &generated.GetValidIngredientMeasurementUnitsParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredient measurement units list retrieval query")
	}

	var (
		data          []*mealplanning.ValidIngredientMeasurementUnit
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &mealplanning.ValidIngredientMeasurementUnit{
			CreatedAt:            result.ValidIngredientMeasurementUnitCreatedAt,
			LastUpdatedAt:        database.TimePointerFromNullTime(result.ValidIngredientMeasurementUnitLastUpdatedAt),
			ArchivedAt:           database.TimePointerFromNullTime(result.ValidIngredientMeasurementUnitArchivedAt),
			MinAllowableQuantity: database.Float32FromString(result.ValidIngredientMeasurementUnitMinimumAllowableQuantity),
			MaxAllowableQuantity: database.Float32PointerFromNullString(result.ValidIngredientMeasurementUnitMaximumAllowableQuantity),
			Notes:                result.ValidIngredientMeasurementUnitNotes,
			ID:                   result.ValidIngredientMeasurementUnitID,
			MeasurementUnit: mealplanning.ValidMeasurementUnit{
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
		})
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vimu *mealplanning.ValidIngredientMeasurementUnit) string { return vimu.ID },
		filter,
	)

	return x, nil
}

// GetValidIngredientMeasurementUnitsByIDs fetches valid ingredient measurement units by their IDs from the database.
func (q *repository) GetValidIngredientMeasurementUnitsByIDs(ctx context.Context, ids []string) (map[string]*mealplanning.ValidIngredientMeasurementUnit, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if len(ids) == 0 {
		return map[string]*mealplanning.ValidIngredientMeasurementUnit{}, nil
	}

	results, err := q.generatedQuerier.GetValidIngredientMeasurementUnitsByIDs(ctx, q.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient measurement units by IDs")
	}

	resultMap := make(map[string]*mealplanning.ValidIngredientMeasurementUnit, len(results))
	for _, result := range results {
		resultMap[result.ValidIngredientMeasurementUnitID] = &mealplanning.ValidIngredientMeasurementUnit{
			CreatedAt:            result.ValidIngredientMeasurementUnitCreatedAt,
			LastUpdatedAt:        database.TimePointerFromNullTime(result.ValidIngredientMeasurementUnitLastUpdatedAt),
			ArchivedAt:           database.TimePointerFromNullTime(result.ValidIngredientMeasurementUnitArchivedAt),
			MinAllowableQuantity: database.Float32FromString(result.ValidIngredientMeasurementUnitMinimumAllowableQuantity),
			MaxAllowableQuantity: database.Float32PointerFromNullString(result.ValidIngredientMeasurementUnitMaximumAllowableQuantity),
			Notes:                result.ValidIngredientMeasurementUnitNotes,
			ID:                   result.ValidIngredientMeasurementUnitID,
			MeasurementUnit: mealplanning.ValidMeasurementUnit{
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
		}
	}

	return resultMap, nil
}

// CreateValidIngredientMeasurementUnit creates a valid ingredient measurement unit in the database.
func (q *repository) CreateValidIngredientMeasurementUnit(ctx context.Context, input *mealplanning.ValidIngredientMeasurementUnitDatabaseCreationInput) (*mealplanning.ValidIngredientMeasurementUnit, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientMeasurementUnitIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientMeasurementUnitIDKey, input.ID)

	// create the valid ingredient measurement unit.
	if err := q.withEvent(ctx, logger, mealplanning.ValidIngredientMeasurementUnitCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientMeasurementUnitIDKey: input.ID,
	}, func(tx database.SQLQueryExecutor) error {
		return q.generatedQuerier.CreateValidIngredientMeasurementUnit(ctx, tx, &generated.CreateValidIngredientMeasurementUnitParams{
			ID:                       input.ID,
			Notes:                    input.Notes,
			ValidMeasurementUnitID:   input.ValidMeasurementUnitID,
			ValidIngredientID:        input.ValidIngredientID,
			MinimumAllowableQuantity: database.StringFromFloat32(input.MinAllowableQuantity),
			MaximumAllowableQuantity: database.NullStringFromFloat32Pointer(input.MaxAllowableQuantity),
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient measurement unit creation query")
	}

	x := &mealplanning.ValidIngredientMeasurementUnit{
		ID:                   input.ID,
		Notes:                input.Notes,
		MeasurementUnit:      mealplanning.ValidMeasurementUnit{ID: input.ValidMeasurementUnitID},
		Ingredient:           mealplanning.ValidIngredient{ID: input.ValidIngredientID},
		MinAllowableQuantity: input.MinAllowableQuantity,
		MaxAllowableQuantity: input.MaxAllowableQuantity,
		CreatedAt:            q.CurrentTime(),
	}

	ingredient, err := q.GetValidIngredient(ctx, input.ValidIngredientID)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient for valid ingredient measurement unit")
	}
	if ingredient != nil {
		x.Ingredient = *ingredient
	}

	measurementUnit, err := q.GetValidMeasurementUnit(ctx, input.ValidMeasurementUnitID)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid measurement unit for valid ingredient measurement unit")
	}
	if measurementUnit != nil {
		x.MeasurementUnit = *measurementUnit
	}

	logger.Info("valid ingredient measurement unit created")

	return x, nil
}

// UpdateValidIngredientMeasurementUnit updates a particular valid ingredient measurement unit.
func (q *repository) UpdateValidIngredientMeasurementUnit(ctx context.Context, updated *mealplanning.ValidIngredientMeasurementUnit) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientMeasurementUnitIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientMeasurementUnitIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, mealplanning.ValidIngredientMeasurementUnitUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientMeasurementUnitIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		_, updateErr := q.generatedQuerier.UpdateValidIngredientMeasurementUnit(ctx, tx, &generated.UpdateValidIngredientMeasurementUnitParams{
			Notes:                    updated.Notes,
			ValidMeasurementUnitID:   updated.MeasurementUnit.ID,
			ValidIngredientID:        updated.Ingredient.ID,
			MinimumAllowableQuantity: database.StringFromFloat32(updated.MinAllowableQuantity),
			MaximumAllowableQuantity: database.NullStringFromFloat32Pointer(updated.MaxAllowableQuantity),
			ID:                       updated.ID,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid ingredient measurement unit")
	}

	logger.Info("valid ingredient measurement unit updated")

	return nil
}

// ArchiveValidIngredientMeasurementUnit archives a valid ingredient measurement unit from the database by its ID.
func (q *repository) ArchiveValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientMeasurementUnitID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientMeasurementUnitIDKey, validIngredientMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientMeasurementUnitIDKey, validIngredientMeasurementUnitID)

	return q.withEvent(ctx, logger, mealplanning.ValidIngredientMeasurementUnitArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientMeasurementUnitIDKey: validIngredientMeasurementUnitID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveValidIngredientMeasurementUnit(ctx, tx, validIngredientMeasurementUnitID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving valid ingredient measurement unit")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}
