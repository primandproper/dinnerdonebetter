package mealplanning

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"
	mealplanningindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v10/database"
	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/filtering"
	"github.com/primandproper/platform-go/v10/observability"
	"github.com/primandproper/platform-go/v10/observability/tracing"
)

var (
	_ mealplanning.RecipeStepIngredientDataManager = (*repository)(nil)
)

// RecipeStepIngredientExists fetches whether a recipe step ingredient exists from the database.
func (q *repository) RecipeStepIngredientExists(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if recipeStepID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if recipeStepIngredientID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIngredientIDKey, recipeStepIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIngredientIDKey, recipeStepIngredientID)

	result, err := q.generatedQuerier.CheckRecipeStepIngredientExistence(ctx, q.readDB, &generated.CheckRecipeStepIngredientExistenceParams{
		RecipeStepID:           recipeStepID,
		RecipeStepIngredientID: recipeStepIngredientID,
		RecipeID:               recipeID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing recipe step ingredient existence check")
	}

	return result, nil
}

// GetRecipeStepIngredient fetches a recipe step ingredient from the database.
func (q *repository) GetRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string) (*mealplanning.RecipeStepIngredient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if recipeStepID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if recipeStepIngredientID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIngredientIDKey, recipeStepIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIngredientIDKey, recipeStepIngredientID)

	result, err := q.generatedQuerier.GetRecipeStepIngredient(ctx, q.readDB, &generated.GetRecipeStepIngredientParams{
		RecipeStepID:           recipeStepID,
		RecipeStepIngredientID: recipeStepIngredientID,
		RecipeID:               recipeID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting recipe step ingredient")
	}

	scaleFactor := database.Float32FromString(result.ScaleFactor)
	if scaleFactor <= 0 {
		scaleFactor = 1.0
	}
	recipeStepIngredient := &mealplanning.RecipeStepIngredient{
		CreatedAt:                 result.CreatedAt,
		RecipeStepProductID:       database.StringPointerFromNullString(result.RecipeStepProductID),
		ArchivedAt:                database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt:             database.TimePointerFromNullTime(result.LastUpdatedAt),
		VesselIndex:               database.Uint16PointerFromNullInt32(result.VesselIndex),
		ProductPercentageToUse:    database.Float32PointerFromNullString(result.ProductPercentageToUse),
		RecipeStepProductRecipeID: database.StringPointerFromNullString(result.RecipeStepProductRecipeID),
		QuantityNotes:             result.QuantityNotes,
		ID:                        result.ID,
		BelongsToRecipeStep:       result.BelongsToRecipeStep,
		IngredientNotes:           result.IngredientNotes,
		Name:                      result.Name,
		ScaleFactor:               scaleFactor,
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
		MinQuantity: database.Float32FromString(result.MinimumQuantityValue),
		MaxQuantity: database.Float32PointerFromNullString(result.MaximumQuantityValue),
		Index:       uint16(result.Index),
		OptionIndex: uint16(result.OptionIndex),
		Optional:    result.Optional,
		ToTaste:     result.ToTaste,
	}

	if result.ValidIngredientID.Valid && result.ValidIngredientID.String != "" {
		recipeStepIngredient.Ingredient = &mealplanning.ValidIngredient{
			CreatedAt:                      result.ValidIngredientCreatedAt.Time,
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidIngredientLastUpdatedAt),
			ArchivedAt:                     database.TimePointerFromNullTime(result.ValidIngredientArchivedAt),
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
			IsLiquid:                       database.BoolFromNullBool(result.ValidIngredientIsLiquid),
			ContainsPeanut:                 result.ValidIngredientContainsPeanut.Bool,
			ContainsTreeNut:                result.ValidIngredientContainsTreeNut.Bool,
			ContainsEgg:                    result.ValidIngredientContainsEgg.Bool,
			ContainsWheat:                  result.ValidIngredientContainsWheat.Bool,
			ContainsSoy:                    result.ValidIngredientContainsSoy.Bool,
			AnimalDerived:                  result.ValidIngredientAnimalDerived.Bool,
			RestrictToPreparations:         result.ValidIngredientRestrictToPreparations.Bool,
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

	return recipeStepIngredient, nil
}

// getRecipeStepIngredientsForRecipe fetches a list of recipe step ingredients from the database that meet a particular filter.
func (q *repository) getRecipeStepIngredientsForRecipe(ctx context.Context, recipeID string) ([]*mealplanning.RecipeStepIngredient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	results, err := q.generatedQuerier.GetAllRecipeStepIngredientsForRecipe(ctx, q.readDB, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe step ingredients list retrieval query")
	}

	recipeStepIngredients := []*mealplanning.RecipeStepIngredient{}
	for _, result := range results {
		scaleFactor := database.Float32FromString(result.ScaleFactor)
		if scaleFactor <= 0 {
			scaleFactor = 1.0
		}
		recipeStepIngredient := &mealplanning.RecipeStepIngredient{
			CreatedAt:                 result.CreatedAt,
			RecipeStepProductID:       database.StringPointerFromNullString(result.RecipeStepProductID),
			ArchivedAt:                database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:             database.TimePointerFromNullTime(result.LastUpdatedAt),
			VesselIndex:               database.Uint16PointerFromNullInt32(result.VesselIndex),
			ProductPercentageToUse:    database.Float32PointerFromNullString(result.ProductPercentageToUse),
			RecipeStepProductRecipeID: database.StringPointerFromNullString(result.RecipeStepProductRecipeID),
			QuantityNotes:             result.QuantityNotes,
			ID:                        result.ID,
			BelongsToRecipeStep:       result.BelongsToRecipeStep,
			IngredientNotes:           result.IngredientNotes,
			Name:                      result.Name,
			ScaleFactor:               scaleFactor,
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
			MinQuantity: database.Float32FromString(result.MinimumQuantityValue),
			MaxQuantity: database.Float32PointerFromNullString(result.MaximumQuantityValue),
			Index:       uint16(result.Index),
			OptionIndex: uint16(result.OptionIndex),
			Optional:    result.Optional,
			ToTaste:     result.ToTaste,
		}

		if result.ValidIngredientID.Valid {
			recipeStepIngredient.Ingredient = &mealplanning.ValidIngredient{
				CreatedAt:                      result.ValidIngredientCreatedAt.Time,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidIngredientLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidIngredientArchivedAt),
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
				IsLiquid:                       database.BoolFromNullBool(result.ValidIngredientIsLiquid),
				ContainsPeanut:                 result.ValidIngredientContainsPeanut.Bool,
				ContainsTreeNut:                result.ValidIngredientContainsTreeNut.Bool,
				ContainsEgg:                    result.ValidIngredientContainsEgg.Bool,
				ContainsWheat:                  result.ValidIngredientContainsWheat.Bool,
				ContainsSoy:                    result.ValidIngredientContainsSoy.Bool,
				AnimalDerived:                  result.ValidIngredientAnimalDerived.Bool,
				RestrictToPreparations:         result.ValidIngredientRestrictToPreparations.Bool,
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

		recipeStepIngredients = append(recipeStepIngredients, recipeStepIngredient)
	}

	return recipeStepIngredients, nil
}

// GetRecipeStepIngredients fetches a list of recipe step ingredients from the database that meet a particular filter.
func (q *repository) GetRecipeStepIngredients(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepIngredient], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if recipeStepID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetRecipeStepIngredients(ctx, q.readDB, &generated.GetRecipeStepIngredientsParams{
		RecipeID:        recipeID,
		RecipeStepID:    recipeStepID,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe step ingredients list retrieval query")
	}

	var (
		data                      []*mealplanning.RecipeStepIngredient
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		scaleFactor := database.Float32FromString(result.ScaleFactor)
		if scaleFactor <= 0 {
			scaleFactor = 1.0
		}
		recipeStepIngredient := &mealplanning.RecipeStepIngredient{
			CreatedAt:                 result.CreatedAt,
			RecipeStepProductID:       database.StringPointerFromNullString(result.RecipeStepProductID),
			ArchivedAt:                database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:             database.TimePointerFromNullTime(result.LastUpdatedAt),
			VesselIndex:               database.Uint16PointerFromNullInt32(result.VesselIndex),
			ProductPercentageToUse:    database.Float32PointerFromNullString(result.ProductPercentageToUse),
			RecipeStepProductRecipeID: database.StringPointerFromNullString(result.RecipeStepProductRecipeID),
			QuantityNotes:             result.QuantityNotes,
			ID:                        result.ID,
			BelongsToRecipeStep:       result.BelongsToRecipeStep,
			IngredientNotes:           result.IngredientNotes,
			Name:                      result.Name,
			ScaleFactor:               scaleFactor,
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
			MinQuantity: database.Float32FromString(result.MinimumQuantityValue),
			MaxQuantity: database.Float32PointerFromNullString(result.MaximumQuantityValue),
			Index:       uint16(result.Index),
			OptionIndex: uint16(result.OptionIndex),
			Optional:    result.Optional,
			ToTaste:     result.ToTaste,
		}

		if result.ValidIngredientID.Valid {
			recipeStepIngredient.Ingredient = &mealplanning.ValidIngredient{
				CreatedAt:                      result.ValidIngredientCreatedAt.Time,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidIngredientLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidIngredientArchivedAt),
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
				IsLiquid:                       database.BoolFromNullBool(result.ValidIngredientIsLiquid),
				ContainsPeanut:                 result.ValidIngredientContainsPeanut.Bool,
				ContainsTreeNut:                result.ValidIngredientContainsTreeNut.Bool,
				ContainsEgg:                    result.ValidIngredientContainsEgg.Bool,
				ContainsWheat:                  result.ValidIngredientContainsWheat.Bool,
				ContainsSoy:                    result.ValidIngredientContainsSoy.Bool,
				AnimalDerived:                  result.ValidIngredientAnimalDerived.Bool,
				RestrictToPreparations:         result.ValidIngredientRestrictToPreparations.Bool,
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

		data = append(data, recipeStepIngredient)
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *mealplanning.RecipeStepIngredient) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// createRecipeStepIngredient creates a recipe step ingredient in the database.
func (q *repository) createRecipeStepIngredient(ctx context.Context, db database.SQLQueryExecutor, input *mealplanning.RecipeStepIngredientDatabaseCreationInput) (*mealplanning.RecipeStepIngredient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	// create the recipe step ingredient.
	var measurementUnit sql.NullString
	if input.MeasurementUnitID != "" {
		measurementUnit = database.NullStringFromString(input.MeasurementUnitID)
	}
	// Convert empty string to nil for RecipeStepProductRecipeID to avoid foreign key constraint violations
	// Empty string is used as a sentinel value to indicate cross-recipe references that will be resolved later
	var recipeStepProductRecipeID *string
	if input.RecipeStepProductRecipeID != nil && *input.RecipeStepProductRecipeID != "" {
		recipeStepProductRecipeID = input.RecipeStepProductRecipeID
	}
	if err := q.generatedQuerier.CreateRecipeStepIngredient(ctx, db, &generated.CreateRecipeStepIngredientParams{
		QuantityNotes:             input.QuantityNotes,
		Name:                      input.Name,
		BelongsToRecipeStep:       input.BelongsToRecipeStep,
		IngredientNotes:           input.IngredientNotes,
		ID:                        input.ID,
		MinimumQuantityValue:      database.StringFromFloat32(input.MinQuantity),
		RecipeStepProductID:       database.NullStringFromStringPointer(input.RecipeStepProductID),
		MaximumQuantityValue:      database.NullStringFromFloat32Pointer(input.MaxQuantity),
		MeasurementUnit:           measurementUnit,
		IngredientID:              database.NullStringFromStringPointer(input.IngredientID),
		ProductPercentageToUse:    database.NullStringFromFloat32Pointer(input.ProductPercentageToUse),
		RecipeStepProductRecipeID: database.NullStringFromStringPointer(recipeStepProductRecipeID),
		VesselIndex:               database.NullInt32FromUint16Pointer(input.VesselIndex),
		Index:                     int32(input.Index),
		OptionIndex:               int32(input.OptionIndex),
		ToTaste:                   input.ToTaste,
		Optional:                  input.Optional,
		ScaleFactor:               database.StringFromFloat32(input.ScaleFactor),
	}); err != nil {
		return nil, observability.PrepareError(err, span, "performing recipe step ingredient creation query")
	}

	x := &mealplanning.RecipeStepIngredient{
		ID:                        input.ID,
		Name:                      input.Name,
		Optional:                  input.Optional,
		MeasurementUnit:           mealplanning.ValidMeasurementUnit{ID: input.MeasurementUnitID},
		MinQuantity:               input.MinQuantity,
		MaxQuantity:               input.MaxQuantity,
		QuantityNotes:             input.QuantityNotes,
		IngredientNotes:           input.IngredientNotes,
		BelongsToRecipeStep:       input.BelongsToRecipeStep,
		RecipeStepProductID:       input.RecipeStepProductID,
		Index:                     input.Index,
		OptionIndex:               input.OptionIndex,
		ToTaste:                   input.ToTaste,
		ProductPercentageToUse:    input.ProductPercentageToUse,
		VesselIndex:               input.VesselIndex,
		RecipeStepProductRecipeID: input.RecipeStepProductRecipeID,
		ScaleFactor:               input.ScaleFactor,
		CreatedAt:                 q.CurrentTime(),
	}

	if input.IngredientID != nil {
		x.Ingredient = &mealplanning.ValidIngredient{ID: *input.IngredientID}
	}

	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIngredientIDKey, x.ID)

	return x, nil
}

// CreateRecipeStepIngredient creates a recipe step ingredient in the database.
func (q *repository) CreateRecipeStepIngredient(ctx context.Context, recipeID string, input *mealplanning.RecipeStepIngredientDatabaseCreationInput) (*mealplanning.RecipeStepIngredient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	// Get the recipe ID from the step
	step, err := q.getRecipeStepByID(ctx, q.readDB, input.BelongsToRecipeStep)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching recipe step")
	}

	// Validate no circular dependency if this ingredient has a cross-recipe reference
	if err = q.validateNoCircularDependencyForIngredient(ctx, step.BelongsToRecipe, input.RecipeStepProductRecipeID); err != nil {
		return nil, observability.PrepareError(err, span, "validating ingredient dependencies")
	}

	var created *mealplanning.RecipeStepIngredient

	// The write and its event share a transaction.
	if err = q.withEvent(ctx, q.logger, mealplanning.RecipeStepIngredientCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           input.BelongsToRecipeStep,
		mealplanningkeys.RecipeStepIngredientIDKey: input.ID,
	}, func(tx database.SQLQueryExecutor) error {
		var createErr error
		created, createErr = q.createRecipeStepIngredient(ctx, tx, input)

		return createErr
	}, events.WithIndexUpsert(mealplanningindexing.IndexTypeRecipes, recipeID)); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateRecipeStepIngredient updates a particular recipe step ingredient.
func (q *repository) UpdateRecipeStepIngredient(ctx context.Context, recipeID string, updated *mealplanning.RecipeStepIngredient) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeStepIngredientIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIngredientIDKey, updated.ID)

	// Get the recipe ID from the step
	step, err := q.getRecipeStepByID(ctx, q.readDB, updated.BelongsToRecipeStep)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching recipe step")
	}

	// Validate no circular dependency if this ingredient has a cross-recipe reference
	if err = q.validateNoCircularDependencyForIngredient(ctx, step.BelongsToRecipe, updated.RecipeStepProductRecipeID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "validating ingredient dependencies")
	}

	var ingredientID *string
	if updated.Ingredient != nil {
		ingredientID = &updated.Ingredient.ID
	}

	if err = q.withEvent(ctx, logger, mealplanning.RecipeStepIngredientUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           updated.BelongsToRecipeStep,
		mealplanningkeys.RecipeStepIngredientIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		_, updateErr := q.generatedQuerier.UpdateRecipeStepIngredient(ctx, tx, &generated.UpdateRecipeStepIngredientParams{
			IngredientID:              database.NullStringFromStringPointer(ingredientID),
			Name:                      updated.Name,
			Optional:                  updated.Optional,
			MeasurementUnit:           database.NullStringFromString(updated.MeasurementUnit.ID),
			MinimumQuantityValue:      database.StringFromFloat32(updated.MinQuantity),
			MaximumQuantityValue:      database.NullStringFromFloat32Pointer(updated.MaxQuantity),
			QuantityNotes:             updated.QuantityNotes,
			RecipeStepProductID:       database.NullStringFromStringPointer(updated.RecipeStepProductID),
			IngredientNotes:           updated.IngredientNotes,
			Index:                     int32(updated.Index),
			OptionIndex:               int32(updated.OptionIndex),
			ToTaste:                   updated.ToTaste,
			ProductPercentageToUse:    database.NullStringFromFloat32Pointer(updated.ProductPercentageToUse),
			VesselIndex:               database.NullInt32FromUint16Pointer(updated.VesselIndex),
			RecipeStepProductRecipeID: database.NullStringFromStringPointer(updated.RecipeStepProductRecipeID),
			BelongsToRecipeStep:       updated.BelongsToRecipeStep,
			ID:                        updated.ID,
			ScaleFactor:               database.StringFromFloat32(updated.ScaleFactor),
		})

		return updateErr
	}, events.WithIndexUpsert(mealplanningindexing.IndexTypeRecipes, recipeID)); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe step ingredient")
	}

	logger.Info("recipe step ingredient updated")

	return nil
}

// ArchiveRecipeStepIngredient archives a recipe step ingredient from the database by its ID.
func (q *repository) ArchiveRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeStepID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if recipeStepIngredientID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIngredientIDKey, recipeStepIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIngredientIDKey, recipeStepIngredientID)

	if err := q.withEvent(ctx, logger, mealplanning.RecipeStepIngredientArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepIngredientIDKey: recipeStepIngredientID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveRecipeStepIngredient(ctx, tx, &generated.ArchiveRecipeStepIngredientParams{
			BelongsToRecipeStep: recipeStepID,
			ID:                  recipeStepIngredientID,
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving recipe step ingredient")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	}, events.WithIndexUpsert(mealplanningindexing.IndexTypeRecipes, recipeID)); err != nil {
		return err
	}

	return nil
}
