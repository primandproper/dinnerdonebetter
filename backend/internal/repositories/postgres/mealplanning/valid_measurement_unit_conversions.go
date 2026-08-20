package mealplanning

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v12/database"
	platformerrors "github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

var (
	_ mealplanning.ValidMeasurementUnitConversionDataManager = (*repository)(nil)
)

// ValidMeasurementUnitConversionExists fetches whether a valid measurement conversion exists from the database.
func (q *repository) ValidMeasurementUnitConversionExists(ctx context.Context, validMeasurementUnitConversionID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validMeasurementUnitConversionID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidMeasurementUnitConversionIDKey, validMeasurementUnitConversionID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitConversionIDKey, validMeasurementUnitConversionID)

	result, err := q.generatedQuerier.CheckValidMeasurementUnitConversionExistence(ctx, q.readDB, validMeasurementUnitConversionID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing valid measurement conversion existence check")
	}

	return result, nil
}

// GetValidMeasurementUnitConversion fetches a valid measurement conversion from the database.
func (q *repository) GetValidMeasurementUnitConversion(ctx context.Context, validMeasurementUnitConversionID string) (*mealplanning.ValidMeasurementUnitConversion, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validMeasurementUnitConversionID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidMeasurementUnitConversionIDKey, validMeasurementUnitConversionID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitConversionIDKey, validMeasurementUnitConversionID)

	result, err := q.generatedQuerier.GetValidMeasurementUnitConversion(ctx, q.readDB, validMeasurementUnitConversionID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "scanning valid measurement conversion")
	}

	validMeasurementUnitConversion := &mealplanning.ValidMeasurementUnitConversion{
		CreatedAt:     result.ValidMeasurementUnitConversionCreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitConversionLastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitConversionArchivedAt),
		Notes:         result.ValidMeasurementUnitConversionNotes,
		ID:            result.ValidMeasurementUnitConversionID,
		From: mealplanning.ValidMeasurementUnit{
			CreatedAt:     result.FromUnitCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.FromUnitLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.FromUnitArchivedAt),
			Name:          result.FromUnitName,
			IconPath:      result.FromUnitIconPath,
			ID:            result.FromUnitID,
			Description:   result.FromUnitDescription,
			PluralName:    result.FromUnitPluralName,
			Slug:          result.FromUnitSlug,
			Volumetric:    database.BoolFromNullBool(result.FromUnitVolumetric),
			Universal:     result.FromUnitUniversal,
			Metric:        result.FromUnitMetric,
			Imperial:      result.FromUnitImperial,
		},
		To: mealplanning.ValidMeasurementUnit{
			CreatedAt:     result.ToUnitCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ToUnitLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ToUnitArchivedAt),
			Name:          result.ToUnitName,
			IconPath:      result.ToUnitIconPath,
			ID:            result.ToUnitID,
			Description:   result.ToUnitDescription,
			PluralName:    result.ToUnitPluralName,
			Slug:          result.ToUnitSlug,
			Volumetric:    database.BoolFromNullBool(result.ToUnitVolumetric),
			Universal:     result.ToUnitUniversal,
			Metric:        result.ToUnitMetric,
			Imperial:      result.ToUnitImperial,
		},
		Modifier: database.Float32FromString(result.ValidMeasurementUnitConversionModifier),
	}

	if result.ValidIngredientID.Valid && result.ValidIngredientID.String != "" {
		validMeasurementUnitConversion.OnlyForIngredient = &mealplanning.ValidIngredient{
			CreatedAt:                      result.ValidIngredientCreatedAt.Time,
			LastUpdatedAt:                  &result.ValidIngredientLastUpdatedAt.Time,
			ArchivedAt:                     &result.ValidIngredientArchivedAt.Time,
			MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMinimumIdealStorageTemperatureInCelsius),
			MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMaximumIdealStorageTemperatureInCelsius),
			IconPath:                       result.ValidIngredientIconPath.String,
			Warning:                        result.ValidIngredientWarning.String,
			PluralName:                     result.ValidIngredientPluralName.String,
			StorageInstructions:            result.ValidIngredientStorageInstructions.String,
			Name:                           result.ValidIngredientName.String,
			ID:                             result.ValidIngredientID.String,
			Description:                    result.ValidIngredientDescription.String,
			Slug:                           result.ValidIngredientSlug.String,
			ShoppingSuggestions:            result.ValidIngredientShoppingSuggestions.String,
			ContainsShellfish:              result.ValidIngredientContainsShellfish.Bool,
			IsLiquid:                       result.ValidIngredientIsLiquid.Bool,
			ContainsPeanut:                 result.ValidIngredientContainsPeanut.Bool,
			ContainsTreeNut:                result.ValidIngredientContainsTreeNut.Bool,
			ContainsEgg:                    result.ValidIngredientContainsEgg.Bool,
			ContainsWheat:                  result.ValidIngredientContainsWheat.Bool,
			ContainsSoy:                    result.ValidIngredientContainsSoy.Bool,
			AnimalDerived:                  result.ValidIngredientAnimalDerived.Bool,
			RestrictToPreparations:         result.ValidIngredientRestrictToPreparations.Bool,
			ContaminatesEquipment:          result.ValidIngredientContaminatesEquipment.Bool,
			ContainsSesame:                 result.ValidIngredientContainsSesame.Bool,
			ContainsFish:                   result.ValidIngredientContainsFish.Bool,
			ContainsGluten:                 result.ValidIngredientContainsGluten.Bool,
			ContainsDairy:                  result.ValidIngredientContainsDairy.Bool,
			ContainsAlcohol:                result.ValidIngredientContainsAlcohol.Bool,
			AnimalFlesh:                    result.ValidIngredientAnimalFlesh.Bool,
			IsStarch:                       result.ValidIngredientIsStarch.Bool,
			IsProtein:                      result.ValidIngredientIsProtein.Bool,
			IsGrain:                        result.ValidIngredientIsGrain.Bool,
			IsFruit:                        result.ValidIngredientIsFruit.Bool,
			IsSalt:                         result.ValidIngredientIsSalt.Bool,
			IsFat:                          result.ValidIngredientIsFat.Bool,
			IsAcid:                         result.ValidIngredientIsAcid.Bool,
			IsHeat:                         result.ValidIngredientIsHeat.Bool,
		}
	}

	return validMeasurementUnitConversion, nil
}

// GetValidMeasurementUnitConversionsForIngredients fetches all valid measurement conversions applicable to the given ingredient IDs
// (universal conversions where only_for_ingredient IS NULL, plus ingredient-specific conversions).
func (q *repository) GetValidMeasurementUnitConversionsForIngredients(ctx context.Context, validIngredientIDs []string) ([]*mealplanning.ValidMeasurementUnitConversion, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	results, err := q.generatedQuerier.GetValidMeasurementUnitConversionsForIngredients(ctx, q.readDB, validIngredientIDs)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "querying for valid measurement conversions for ingredients")
	}

	data := make([]*mealplanning.ValidMeasurementUnitConversion, 0, len(results))
	for _, result := range results {
		conversion := &mealplanning.ValidMeasurementUnitConversion{
			CreatedAt:     result.ValidMeasurementUnitConversionCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitConversionLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitConversionArchivedAt),
			Notes:         result.ValidMeasurementUnitConversionNotes,
			ID:            result.ValidMeasurementUnitConversionID,
			From: mealplanning.ValidMeasurementUnit{
				CreatedAt:     result.FromUnitCreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.FromUnitLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.FromUnitArchivedAt),
				Name:          result.FromUnitName,
				IconPath:      result.FromUnitIconPath,
				ID:            result.FromUnitID,
				Description:   result.FromUnitDescription,
				PluralName:    result.FromUnitPluralName,
				Slug:          result.FromUnitSlug,
				Volumetric:    database.BoolFromNullBool(result.FromUnitVolumetric),
				Universal:     result.FromUnitUniversal,
				Metric:        result.FromUnitMetric,
				Imperial:      result.FromUnitImperial,
			},
			To: mealplanning.ValidMeasurementUnit{
				CreatedAt:     result.ToUnitCreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.ToUnitLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ToUnitArchivedAt),
				Name:          result.ToUnitName,
				IconPath:      result.ToUnitIconPath,
				ID:            result.ToUnitID,
				Description:   result.ToUnitDescription,
				PluralName:    result.ToUnitPluralName,
				Slug:          result.ToUnitSlug,
				Volumetric:    database.BoolFromNullBool(result.ToUnitVolumetric),
				Universal:     result.ToUnitUniversal,
				Metric:        result.ToUnitMetric,
				Imperial:      result.ToUnitImperial,
			},
			Modifier: database.Float32FromString(result.ValidMeasurementUnitConversionModifier),
		}

		if result.ValidIngredientID.Valid && result.ValidIngredientID.String != "" {
			conversion.OnlyForIngredient = &mealplanning.ValidIngredient{
				CreatedAt:                      result.ValidIngredientCreatedAt.Time,
				LastUpdatedAt:                  &result.ValidIngredientLastUpdatedAt.Time,
				ArchivedAt:                     &result.ValidIngredientArchivedAt.Time,
				MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMinimumIdealStorageTemperatureInCelsius),
				MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMaximumIdealStorageTemperatureInCelsius),
				IconPath:                       result.ValidIngredientIconPath.String,
				Warning:                        result.ValidIngredientWarning.String,
				PluralName:                     result.ValidIngredientPluralName.String,
				StorageInstructions:            result.ValidIngredientStorageInstructions.String,
				Name:                           result.ValidIngredientName.String,
				ID:                             result.ValidIngredientID.String,
				Description:                    result.ValidIngredientDescription.String,
				Slug:                           result.ValidIngredientSlug.String,
				ShoppingSuggestions:            result.ValidIngredientShoppingSuggestions.String,
				ContainsShellfish:              result.ValidIngredientContainsShellfish.Bool,
				IsLiquid:                       result.ValidIngredientIsLiquid.Bool,
				ContainsPeanut:                 result.ValidIngredientContainsPeanut.Bool,
				ContainsTreeNut:                result.ValidIngredientContainsTreeNut.Bool,
				ContainsEgg:                    result.ValidIngredientContainsEgg.Bool,
				ContainsWheat:                  result.ValidIngredientContainsWheat.Bool,
				ContainsSoy:                    result.ValidIngredientContainsSoy.Bool,
				AnimalDerived:                  result.ValidIngredientAnimalDerived.Bool,
				RestrictToPreparations:         result.ValidIngredientRestrictToPreparations.Bool,
				ContaminatesEquipment:          result.ValidIngredientContaminatesEquipment.Bool,
				ContainsSesame:                 result.ValidIngredientContainsSesame.Bool,
				ContainsFish:                   result.ValidIngredientContainsFish.Bool,
				ContainsGluten:                 result.ValidIngredientContainsGluten.Bool,
				ContainsDairy:                  result.ValidIngredientContainsDairy.Bool,
				ContainsAlcohol:                result.ValidIngredientContainsAlcohol.Bool,
				AnimalFlesh:                    result.ValidIngredientAnimalFlesh.Bool,
				IsStarch:                       result.ValidIngredientIsStarch.Bool,
				IsProtein:                      result.ValidIngredientIsProtein.Bool,
				IsGrain:                        result.ValidIngredientIsGrain.Bool,
				IsFruit:                        result.ValidIngredientIsFruit.Bool,
				IsSalt:                         result.ValidIngredientIsSalt.Bool,
				IsFat:                          result.ValidIngredientIsFat.Bool,
				IsAcid:                         result.ValidIngredientIsAcid.Bool,
				IsHeat:                         result.ValidIngredientIsHeat.Bool,
			}
		}

		data = append(data, conversion)
	}

	return data, nil
}

// GetValidMeasurementUnitConversionsForUnit fetches all valid measurement conversions involving a given measurement unit.
func (q *repository) GetValidMeasurementUnitConversionsForUnit(ctx context.Context, validMeasurementUnitID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnitConversion], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validMeasurementUnitID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitIDKey, validMeasurementUnitID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	logger = filter.AttachToLogger(logger)

	results, err := q.generatedQuerier.GetValidMeasurementUnitConversionsForMeasurementUnit(ctx, q.readDB, &generated.GetValidMeasurementUnitConversionsForMeasurementUnitParams{
		ID:              validMeasurementUnitID,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "querying for valid measurement conversions")
	}

	var (
		data          []*mealplanning.ValidMeasurementUnitConversion
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}

		conversion := &mealplanning.ValidMeasurementUnitConversion{
			CreatedAt:     result.ValidMeasurementUnitConversionCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitConversionLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitConversionArchivedAt),
			Notes:         result.ValidMeasurementUnitConversionNotes,
			ID:            result.ValidMeasurementUnitConversionID,
			From: mealplanning.ValidMeasurementUnit{
				CreatedAt:     result.FromUnitCreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.FromUnitLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.FromUnitArchivedAt),
				Name:          result.FromUnitName,
				IconPath:      result.FromUnitIconPath,
				ID:            result.FromUnitID,
				Description:   result.FromUnitDescription,
				PluralName:    result.FromUnitPluralName,
				Slug:          result.FromUnitSlug,
				Volumetric:    database.BoolFromNullBool(result.FromUnitVolumetric),
				Universal:     result.FromUnitUniversal,
				Metric:        result.FromUnitMetric,
				Imperial:      result.FromUnitImperial,
			},
			To: mealplanning.ValidMeasurementUnit{
				CreatedAt:     result.ToUnitCreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.ToUnitLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ToUnitArchivedAt),
				Name:          result.ToUnitName,
				IconPath:      result.ToUnitIconPath,
				ID:            result.ToUnitID,
				Description:   result.ToUnitDescription,
				PluralName:    result.ToUnitPluralName,
				Slug:          result.ToUnitSlug,
				Volumetric:    database.BoolFromNullBool(result.ToUnitVolumetric),
				Universal:     result.ToUnitUniversal,
				Metric:        result.ToUnitMetric,
				Imperial:      result.ToUnitImperial,
			},
			Modifier: database.Float32FromString(result.ValidMeasurementUnitConversionModifier),
		}

		if result.ValidIngredientID.Valid && result.ValidIngredientID.String != "" {
			conversion.OnlyForIngredient = &mealplanning.ValidIngredient{
				CreatedAt:                      result.ValidIngredientCreatedAt.Time,
				LastUpdatedAt:                  &result.ValidIngredientLastUpdatedAt.Time,
				ArchivedAt:                     &result.ValidIngredientArchivedAt.Time,
				MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMinimumIdealStorageTemperatureInCelsius),
				MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMaximumIdealStorageTemperatureInCelsius),
				IconPath:                       result.ValidIngredientIconPath.String,
				Warning:                        result.ValidIngredientWarning.String,
				PluralName:                     result.ValidIngredientPluralName.String,
				StorageInstructions:            result.ValidIngredientStorageInstructions.String,
				Name:                           result.ValidIngredientName.String,
				ID:                             result.ValidIngredientID.String,
				Description:                    result.ValidIngredientDescription.String,
				Slug:                           result.ValidIngredientSlug.String,
				ShoppingSuggestions:            result.ValidIngredientShoppingSuggestions.String,
				ContainsShellfish:              result.ValidIngredientContainsShellfish.Bool,
				IsLiquid:                       result.ValidIngredientIsLiquid.Bool,
				ContainsPeanut:                 result.ValidIngredientContainsPeanut.Bool,
				ContainsTreeNut:                result.ValidIngredientContainsTreeNut.Bool,
				ContainsEgg:                    result.ValidIngredientContainsEgg.Bool,
				ContainsWheat:                  result.ValidIngredientContainsWheat.Bool,
				ContainsSoy:                    result.ValidIngredientContainsSoy.Bool,
				AnimalDerived:                  result.ValidIngredientAnimalDerived.Bool,
				RestrictToPreparations:         result.ValidIngredientRestrictToPreparations.Bool,
				ContaminatesEquipment:          result.ValidIngredientContaminatesEquipment.Bool,
				ContainsSesame:                 result.ValidIngredientContainsSesame.Bool,
				ContainsFish:                   result.ValidIngredientContainsFish.Bool,
				ContainsGluten:                 result.ValidIngredientContainsGluten.Bool,
				ContainsDairy:                  result.ValidIngredientContainsDairy.Bool,
				ContainsAlcohol:                result.ValidIngredientContainsAlcohol.Bool,
				AnimalFlesh:                    result.ValidIngredientAnimalFlesh.Bool,
				IsStarch:                       result.ValidIngredientIsStarch.Bool,
				IsProtein:                      result.ValidIngredientIsProtein.Bool,
				IsGrain:                        result.ValidIngredientIsGrain.Bool,
				IsFruit:                        result.ValidIngredientIsFruit.Bool,
				IsSalt:                         result.ValidIngredientIsSalt.Bool,
				IsFat:                          result.ValidIngredientIsFat.Bool,
				IsAcid:                         result.ValidIngredientIsAcid.Bool,
				IsHeat:                         result.ValidIngredientIsHeat.Bool,
			}
		}

		data = append(data, conversion)
	}

	return filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(c *mealplanning.ValidMeasurementUnitConversion) string { return c.ID },
		filter,
	), nil
}

// CreateValidMeasurementUnitConversion creates a valid measurement conversion in the database.
func (q *repository) CreateValidMeasurementUnitConversion(ctx context.Context, input *mealplanning.ValidMeasurementUnitConversionDatabaseCreationInput) (*mealplanning.ValidMeasurementUnitConversion, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := q.logger.WithValue(mealplanningkeys.ValidMeasurementUnitConversionIDKey, input.ID)

	// Normalize to canonical ordering (smaller ID first) to satisfy CHECK constraint
	// This ensures duplicate entries hit the unique constraint rather than the check constraint
	fromUnit := input.From
	toUnit := input.To
	modifier := input.Modifier

	if fromUnit > toUnit {
		// Swap the units and invert the modifier
		fromUnit, toUnit = toUnit, fromUnit
		modifier = 1.0 / modifier
	}

	// create the valid measurement conversion.
	if err := q.withEvent(ctx, logger, mealplanning.ValidMeasurementUnitConversionCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidMeasurementUnitConversionIDKey: input.ID,
	}, func(tx database.SQLQueryExecutor) error {
		return q.generatedQuerier.CreateValidMeasurementUnitConversion(ctx, tx, &generated.CreateValidMeasurementUnitConversionParams{
			ID:                input.ID,
			FromUnit:          fromUnit,
			ToUnit:            toUnit,
			Modifier:          database.StringFromFloat32(modifier),
			Notes:             input.Notes,
			OnlyForIngredient: database.NullStringFromStringPointer(input.OnlyForIngredient),
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid measurement conversion creation query")
	}

	x := &mealplanning.ValidMeasurementUnitConversion{
		ID:        input.ID,
		From:      mealplanning.ValidMeasurementUnit{ID: fromUnit},
		To:        mealplanning.ValidMeasurementUnit{ID: toUnit},
		Modifier:  modifier,
		Notes:     input.Notes,
		CreatedAt: q.CurrentTime(),
	}

	if input.OnlyForIngredient != nil {
		ingredient, err := q.GetValidIngredient(ctx, *input.OnlyForIngredient)
		if err != nil {
			// basically impossible for this ingredient happen and not error out earlier
			return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation for valid preparation instrument")
		}
		if ingredient != nil {
			x.OnlyForIngredient = ingredient
		}
	}

	to, err := q.GetValidMeasurementUnit(ctx, toUnit)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching to valid measurement unit for valid measurement unit conversion")
	}
	if to != nil {
		x.To = *to
	}

	from, err := q.GetValidMeasurementUnit(ctx, fromUnit)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching from valid measurement unit for valid measurement unit conversion")
	}
	if from != nil {
		x.From = *from
	}

	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitConversionIDKey, x.ID)
	logger.Info("valid measurement conversion created")

	return x, nil
}

// UpdateValidMeasurementUnitConversion updates a particular valid measurement conversion.
func (q *repository) UpdateValidMeasurementUnitConversion(ctx context.Context, updated *mealplanning.ValidMeasurementUnitConversion) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := q.logger.WithValue(mealplanningkeys.ValidMeasurementUnitConversionIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitConversionIDKey, updated.ID)

	var ingredientID *string
	if updated.OnlyForIngredient != nil {
		ingredientID = &updated.OnlyForIngredient.ID
	}

	if err := q.withEvent(ctx, logger, mealplanning.ValidMeasurementUnitConversionUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidMeasurementUnitConversionIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		_, updateErr := q.generatedQuerier.UpdateValidMeasurementUnitConversion(ctx, tx, &generated.UpdateValidMeasurementUnitConversionParams{
			FromUnit:          updated.From.ID,
			ToUnit:            updated.To.ID,
			OnlyForIngredient: database.NullStringFromStringPointer(ingredientID),
			Modifier:          database.StringFromFloat32(updated.Modifier),
			Notes:             updated.Notes,
			ID:                updated.ID,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid measurement conversion")
	}

	logger.Info("valid measurement conversion updated")

	return nil
}

// ArchiveValidMeasurementUnitConversion archives a valid measurement conversion from the database by its ID.
func (q *repository) ArchiveValidMeasurementUnitConversion(ctx context.Context, validMeasurementUnitConversionID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validMeasurementUnitConversionID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidMeasurementUnitConversionIDKey, validMeasurementUnitConversionID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidMeasurementUnitConversionIDKey, validMeasurementUnitConversionID)

	return q.withEvent(ctx, logger, mealplanning.ValidMeasurementUnitConversionArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidMeasurementUnitConversionIDKey: validMeasurementUnitConversionID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveValidMeasurementUnitConversion(ctx, tx, validMeasurementUnitConversionID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving valid measurement conversion")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}

// GetMeasurementUnitConversionMismatches fetches ingredient-unit pairs that lack a conversion in the database.
func (q *repository) GetMeasurementUnitConversionMismatches(ctx context.Context) ([]*mealplanning.MeasurementUnitConversionMismatch, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	rows, err := q.generatedQuerier.GetMeasurementUnitConversionMismatches(ctx, q.readDB)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "querying for measurement unit conversion mismatches")
	}

	mismatches := make([]*mealplanning.MeasurementUnitConversionMismatch, 0, len(rows))
	for _, row := range rows {
		var ingredient *mealplanning.ValidIngredient
		var fromUnit, toUnit *mealplanning.ValidMeasurementUnit
		ingredient, err = q.GetValidIngredient(ctx, row.ValidIngredientID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "fetching ingredient for mismatch")
		}
		fromUnit, err = q.GetValidMeasurementUnit(ctx, row.FromUnitID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "fetching from unit for mismatch")
		}
		toUnit, err = q.GetValidMeasurementUnit(ctx, row.ToUnitID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "fetching to unit for mismatch")
		}
		mismatches = append(mismatches, &mealplanning.MeasurementUnitConversionMismatch{
			Ingredient: *ingredient,
			FromUnit:   derefValidMeasurementUnit(fromUnit),
			ToUnit:     derefValidMeasurementUnit(toUnit),
		})
	}

	return mismatches, nil
}

// derefValidMeasurementUnit returns the value or zero value if nil.
// GetValidMeasurementUnit returns *mealplanning.ValidMeasurementUnit (via types alias in valid_measurement_units).
func derefValidMeasurementUnit(u *mealplanning.ValidMeasurementUnit) mealplanning.ValidMeasurementUnit {
	if u == nil {
		return mealplanning.ValidMeasurementUnit{}
	}
	return *u
}
