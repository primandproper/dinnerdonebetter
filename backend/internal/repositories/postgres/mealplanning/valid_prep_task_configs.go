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
	_ mealplanning.ValidPrepTaskConfigDataManager = (*repository)(nil)
)

// ValidPrepTaskConfigExists fetches whether a valid prep task config exists from the database.
func (q *repository) ValidPrepTaskConfigExists(ctx context.Context, validPrepTaskConfigID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validPrepTaskConfigID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidPrepTaskConfigIDKey, validPrepTaskConfigID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPrepTaskConfigIDKey, validPrepTaskConfigID)

	result, err := q.generatedQuerier.CheckValidPrepTaskConfigExistence(ctx, q.readDB, validPrepTaskConfigID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing valid prep task config existence check")
	}

	return result, nil
}

// GetValidPrepTaskConfig fetches a valid prep task config from the database.
func (q *repository) GetValidPrepTaskConfig(ctx context.Context, validPrepTaskConfigID string) (*mealplanning.ValidPrepTaskConfig, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validPrepTaskConfigID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidPrepTaskConfigIDKey, validPrepTaskConfigID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPrepTaskConfigIDKey, validPrepTaskConfigID)

	result, err := q.generatedQuerier.GetValidPrepTaskConfig(ctx, q.readDB, validPrepTaskConfigID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid prep task config retrieval")
	}

	return convertValidPrepTaskConfigRow(result), nil
}

func convertValidPrepTaskConfigRow(result *generated.GetValidPrepTaskConfigRow) *mealplanning.ValidPrepTaskConfig {
	return &mealplanning.ValidPrepTaskConfig{
		CreatedAt:                      result.ValidPrepTaskConfigCreatedAt,
		LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidPrepTaskConfigLastUpdatedAt),
		ArchivedAt:                     database.TimePointerFromNullTime(result.ValidPrepTaskConfigArchivedAt),
		MinStorageDurationInSeconds:    uint32(result.ValidPrepTaskConfigMinimumStorageDurationInSeconds),
		MaxStorageDurationInSeconds:    database.Uint32PointerFromNullInt32(result.ValidPrepTaskConfigMaximumStorageDurationInSeconds),
		MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidPrepTaskConfigMinimumStorageTemperatureInCelsius),
		MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidPrepTaskConfigMaximumStorageTemperatureInCelsius),
		ID:                             result.ValidPrepTaskConfigID,
		StorageType:                    string(result.ValidPrepTaskConfigStorageContainerType),
		StorageInstructions:            result.ValidPrepTaskConfigStorageInstructions,
		Notes:                          result.ValidPrepTaskConfigNotes,
		Source:                         result.ValidPrepTaskConfigSource,
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
			MinIngredientCount:          uint16(result.ValidPreparationMinimumIngredientCount),
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

// GetValidPrepTaskConfigs fetches a list of valid prep task configs from the database that meet a particular filter.
func (q *repository) GetValidPrepTaskConfigs(ctx context.Context, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidPrepTaskConfigs(ctx, q.readDB, &generated.GetValidPrepTaskConfigsParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid prep task configs list retrieval query")
	}

	var (
		data          []*mealplanning.ValidPrepTaskConfig
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, convertValidPrepTaskConfigsRow(result))
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vptc *mealplanning.ValidPrepTaskConfig) string { return vptc.ID },
		filter,
	)

	return x, nil
}

func convertValidPrepTaskConfigsRow(result *generated.GetValidPrepTaskConfigsRow) *mealplanning.ValidPrepTaskConfig {
	return &mealplanning.ValidPrepTaskConfig{
		CreatedAt:                      result.ValidPrepTaskConfigCreatedAt,
		LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidPrepTaskConfigLastUpdatedAt),
		ArchivedAt:                     database.TimePointerFromNullTime(result.ValidPrepTaskConfigArchivedAt),
		MinStorageDurationInSeconds:    uint32(result.ValidPrepTaskConfigMinimumStorageDurationInSeconds),
		MaxStorageDurationInSeconds:    database.Uint32PointerFromNullInt32(result.ValidPrepTaskConfigMaximumStorageDurationInSeconds),
		MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidPrepTaskConfigMinimumStorageTemperatureInCelsius),
		MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidPrepTaskConfigMaximumStorageTemperatureInCelsius),
		ID:                             result.ValidPrepTaskConfigID,
		StorageType:                    string(result.ValidPrepTaskConfigStorageContainerType),
		StorageInstructions:            result.ValidPrepTaskConfigStorageInstructions,
		Notes:                          result.ValidPrepTaskConfigNotes,
		Source:                         result.ValidPrepTaskConfigSource,
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
			MinIngredientCount:          uint16(result.ValidPreparationMinimumIngredientCount),
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

// GetValidPrepTaskConfigsForPreparation fetches a list of valid prep task configs from the database for a particular preparation.
func (q *repository) GetValidPrepTaskConfigsForPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if preparationID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidPreparationIDKey, preparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, preparationID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidPrepTaskConfigsForPreparation(ctx, q.readDB, &generated.GetValidPrepTaskConfigsForPreparationParams{
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
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid prep task configs for preparation list retrieval query")
	}

	var (
		data          []*mealplanning.ValidPrepTaskConfig
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, convertValidPrepTaskConfigsForPreparationRow(result))
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vptc *mealplanning.ValidPrepTaskConfig) string { return vptc.ID },
		filter,
	)

	return x, nil
}

func convertValidPrepTaskConfigsForPreparationRow(result *generated.GetValidPrepTaskConfigsForPreparationRow) *mealplanning.ValidPrepTaskConfig {
	return &mealplanning.ValidPrepTaskConfig{
		CreatedAt:                      result.ValidPrepTaskConfigCreatedAt,
		LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidPrepTaskConfigLastUpdatedAt),
		ArchivedAt:                     database.TimePointerFromNullTime(result.ValidPrepTaskConfigArchivedAt),
		MinStorageDurationInSeconds:    uint32(result.ValidPrepTaskConfigMinimumStorageDurationInSeconds),
		MaxStorageDurationInSeconds:    database.Uint32PointerFromNullInt32(result.ValidPrepTaskConfigMaximumStorageDurationInSeconds),
		MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidPrepTaskConfigMinimumStorageTemperatureInCelsius),
		MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidPrepTaskConfigMaximumStorageTemperatureInCelsius),
		ID:                             result.ValidPrepTaskConfigID,
		StorageType:                    string(result.ValidPrepTaskConfigStorageContainerType),
		StorageInstructions:            result.ValidPrepTaskConfigStorageInstructions,
		Notes:                          result.ValidPrepTaskConfigNotes,
		Source:                         result.ValidPrepTaskConfigSource,
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
			MinIngredientCount:          uint16(result.ValidPreparationMinimumIngredientCount),
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

// GetValidPrepTaskConfigsForIngredient fetches a list of valid prep task configs from the database for a particular ingredient.
func (q *repository) GetValidPrepTaskConfigsForIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if ingredientID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientIDKey, ingredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, ingredientID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidPrepTaskConfigsForIngredient(ctx, q.readDB, &generated.GetValidPrepTaskConfigsForIngredientParams{
		ID:              ingredientID,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid prep task configs for ingredient list retrieval query")
	}

	var (
		data          []*mealplanning.ValidPrepTaskConfig
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, convertValidPrepTaskConfigsForIngredientRow(result))
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vptc *mealplanning.ValidPrepTaskConfig) string { return vptc.ID },
		filter,
	)

	return x, nil
}

func convertValidPrepTaskConfigsForIngredientRow(result *generated.GetValidPrepTaskConfigsForIngredientRow) *mealplanning.ValidPrepTaskConfig {
	return &mealplanning.ValidPrepTaskConfig{
		CreatedAt:                      result.ValidPrepTaskConfigCreatedAt,
		LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidPrepTaskConfigLastUpdatedAt),
		ArchivedAt:                     database.TimePointerFromNullTime(result.ValidPrepTaskConfigArchivedAt),
		MinStorageDurationInSeconds:    uint32(result.ValidPrepTaskConfigMinimumStorageDurationInSeconds),
		MaxStorageDurationInSeconds:    database.Uint32PointerFromNullInt32(result.ValidPrepTaskConfigMaximumStorageDurationInSeconds),
		MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidPrepTaskConfigMinimumStorageTemperatureInCelsius),
		MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidPrepTaskConfigMaximumStorageTemperatureInCelsius),
		ID:                             result.ValidPrepTaskConfigID,
		StorageType:                    string(result.ValidPrepTaskConfigStorageContainerType),
		StorageInstructions:            result.ValidPrepTaskConfigStorageInstructions,
		Notes:                          result.ValidPrepTaskConfigNotes,
		Source:                         result.ValidPrepTaskConfigSource,
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
			MinIngredientCount:          uint16(result.ValidPreparationMinimumIngredientCount),
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

// GetValidPrepTaskConfigsForIngredientAndPreparation fetches a list of valid prep task configs from the database for a particular ingredient and preparation.
func (q *repository) GetValidPrepTaskConfigsForIngredientAndPreparation(ctx context.Context, ingredientID, preparationID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if ingredientID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientIDKey, ingredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, ingredientID)

	if preparationID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidPreparationIDKey, preparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPreparationIDKey, preparationID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidPrepTaskConfigsForIngredientAndPreparation(ctx, q.readDB, &generated.GetValidPrepTaskConfigsForIngredientAndPreparationParams{
		ValidIngredientID:  ingredientID,
		ValidPreparationID: preparationID,
		CreatedBefore:      database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:       database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:      database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:       database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:             database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:        database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived:    database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid prep task configs for ingredient and preparation list retrieval query")
	}

	var (
		data          []*mealplanning.ValidPrepTaskConfig
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, convertValidPrepTaskConfigsForIngredientAndPreparationRow(result))
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vptc *mealplanning.ValidPrepTaskConfig) string { return vptc.ID },
		filter,
	)

	return x, nil
}

func convertValidPrepTaskConfigsForIngredientAndPreparationRow(result *generated.GetValidPrepTaskConfigsForIngredientAndPreparationRow) *mealplanning.ValidPrepTaskConfig {
	return &mealplanning.ValidPrepTaskConfig{
		CreatedAt:                      result.ValidPrepTaskConfigCreatedAt,
		LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidPrepTaskConfigLastUpdatedAt),
		ArchivedAt:                     database.TimePointerFromNullTime(result.ValidPrepTaskConfigArchivedAt),
		MinStorageDurationInSeconds:    uint32(result.ValidPrepTaskConfigMinimumStorageDurationInSeconds),
		MaxStorageDurationInSeconds:    database.Uint32PointerFromNullInt32(result.ValidPrepTaskConfigMaximumStorageDurationInSeconds),
		MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidPrepTaskConfigMinimumStorageTemperatureInCelsius),
		MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidPrepTaskConfigMaximumStorageTemperatureInCelsius),
		ID:                             result.ValidPrepTaskConfigID,
		StorageType:                    string(result.ValidPrepTaskConfigStorageContainerType),
		StorageInstructions:            result.ValidPrepTaskConfigStorageInstructions,
		Notes:                          result.ValidPrepTaskConfigNotes,
		Source:                         result.ValidPrepTaskConfigSource,
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
			MinIngredientCount:          uint16(result.ValidPreparationMinimumIngredientCount),
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

// CreateValidPrepTaskConfig creates a valid prep task config in the database.
func (q *repository) CreateValidPrepTaskConfig(ctx context.Context, input *mealplanning.ValidPrepTaskConfigDatabaseCreationInput) (*mealplanning.ValidPrepTaskConfig, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidPrepTaskConfigIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.ValidPrepTaskConfigIDKey, input.ID)

	// create the valid prep task config.
	if err := q.generatedQuerier.CreateValidPrepTaskConfig(ctx, q.writeDB, &generated.CreateValidPrepTaskConfigParams{
		ID:                                 input.ID,
		ValidIngredientID:                  input.ValidIngredientID,
		ValidPreparationID:                 input.ValidPreparationID,
		MinimumStorageDurationInSeconds:    int32(input.MinStorageDurationInSeconds),
		MaximumStorageDurationInSeconds:    database.NullInt32FromUint32Pointer(input.MaxStorageDurationInSeconds),
		StorageContainerType:               generated.StorageContainerType(input.StorageType),
		MinimumStorageTemperatureInCelsius: database.NullStringFromFloat32Pointer(input.MinStorageTemperatureInCelsius),
		MaximumStorageTemperatureInCelsius: database.NullStringFromFloat32Pointer(input.MaxStorageTemperatureInCelsius),
		StorageInstructions:                input.StorageInstructions,
		Notes:                              input.Notes,
		Source:                             input.Source,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid prep task config creation query")
	}

	x := &mealplanning.ValidPrepTaskConfig{
		ID:                             input.ID,
		MinStorageDurationInSeconds:    input.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    input.MaxStorageDurationInSeconds,
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		StorageType:                    input.StorageType,
		StorageInstructions:            input.StorageInstructions,
		Notes:                          input.Notes,
		Source:                         input.Source,
		Preparation:                    mealplanning.ValidPreparation{ID: input.ValidPreparationID},
		Ingredient:                     mealplanning.ValidIngredient{ID: input.ValidIngredientID},
		CreatedAt:                      q.CurrentTime(),
	}

	preparation, err := q.GetValidPreparation(ctx, input.ValidPreparationID)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation for valid prep task config")
	}
	if preparation != nil {
		x.Preparation = *preparation
	}

	ingredient, err := q.GetValidIngredient(ctx, input.ValidIngredientID)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient for valid prep task config")
	}
	if ingredient != nil {
		x.Ingredient = *ingredient
	}

	return x, nil
}

// UpdateValidPrepTaskConfig updates a particular valid prep task config.
func (q *repository) UpdateValidPrepTaskConfig(ctx context.Context, updated *mealplanning.ValidPrepTaskConfig) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidPrepTaskConfigIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPrepTaskConfigIDKey, updated.ID)

	if _, err := q.generatedQuerier.UpdateValidPrepTaskConfig(ctx, q.writeDB, &generated.UpdateValidPrepTaskConfigParams{
		ValidIngredientID:                  updated.Ingredient.ID,
		ValidPreparationID:                 updated.Preparation.ID,
		MinimumStorageDurationInSeconds:    int32(updated.MinStorageDurationInSeconds),
		MaximumStorageDurationInSeconds:    database.NullInt32FromUint32Pointer(updated.MaxStorageDurationInSeconds),
		StorageContainerType:               generated.StorageContainerType(updated.StorageType),
		MinimumStorageTemperatureInCelsius: database.NullStringFromFloat32Pointer(updated.MinStorageTemperatureInCelsius),
		MaximumStorageTemperatureInCelsius: database.NullStringFromFloat32Pointer(updated.MaxStorageTemperatureInCelsius),
		StorageInstructions:                updated.StorageInstructions,
		Notes:                              updated.Notes,
		Source:                             updated.Source,
		ID:                                 updated.ID,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid prep task config")
	}

	logger.Info("valid prep task config updated")

	return nil
}

// ArchiveValidPrepTaskConfig archives a valid prep task config from the database by its ID.
func (q *repository) ArchiveValidPrepTaskConfig(ctx context.Context, validPrepTaskConfigID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validPrepTaskConfigID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidPrepTaskConfigIDKey, validPrepTaskConfigID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidPrepTaskConfigIDKey, validPrepTaskConfigID)

	rowsAffected, err := q.generatedQuerier.ArchiveValidPrepTaskConfig(ctx, q.writeDB, validPrepTaskConfigID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid prep task config")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
