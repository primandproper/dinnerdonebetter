package mealplanning

import (
	"context"
	"database/sql"

	mealplanning "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v11/database"
	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/observability"
	"github.com/primandproper/platform-go/v11/observability/tracing"
)

var (
	_ mealplanning.ValidIngredientStateIngredientDataManager = (*repository)(nil)
)

// ValidIngredientStateIngredientExists fetches whether a valid ingredient state ingredient exists from the database.
func (q *repository) ValidIngredientStateIngredientExists(ctx context.Context, validIngredientStateIngredientID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientStateIngredientID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientStateIngredientIDKey, validIngredientStateIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, validIngredientStateIngredientID)

	result, err := q.generatedQuerier.CheckValidIngredientStateIngredientExistence(ctx, q.readDB, validIngredientStateIngredientID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient state ingredient existence check")
	}

	return result, nil
}

// GetValidIngredientStateIngredient fetches a valid ingredient state ingredient from the database.
func (q *repository) GetValidIngredientStateIngredient(ctx context.Context, validIngredientStateIngredientID string) (*mealplanning.ValidIngredientStateIngredient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientStateIngredientID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientStateIngredientIDKey, validIngredientStateIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, validIngredientStateIngredientID)

	result, err := q.generatedQuerier.GetValidIngredientStateIngredient(ctx, q.readDB, validIngredientStateIngredientID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting valid ingredient state ingredient")
	}

	validIngredientStateIngredient := &mealplanning.ValidIngredientStateIngredient{
		CreatedAt:     result.ValidIngredientStateIngredientCreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientStateIngredientLastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientStateIngredientArchivedAt),
		Notes:         result.ValidIngredientStateIngredientNotes,
		ID:            result.ValidIngredientStateIngredientID,
		IngredientState: mealplanning.ValidIngredientState{
			CreatedAt:     result.ValidIngredientStateCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientStateLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientStateArchivedAt),
			PastTense:     result.ValidIngredientStatePastTense,
			Description:   result.ValidIngredientStateDescription,
			IconPath:      result.ValidIngredientStateIconPath,
			ID:            result.ValidIngredientStateID,
			Name:          result.ValidIngredientStateName,
			AttributeType: string(result.ValidIngredientStateAttributeType),
			Slug:          result.ValidIngredientStateSlug,
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

	return validIngredientStateIngredient, nil
}

// GetValidIngredientStateIngredients fetches a list of valid ingredient state ingredients from the database that meet a particular filter.
func (q *repository) GetValidIngredientStateIngredients(ctx context.Context, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidIngredientStateIngredients(ctx, q.readDB, &generated.GetValidIngredientStateIngredientsParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredient state ingredients list retrieval query")
	}

	var (
		data          []*mealplanning.ValidIngredientStateIngredient
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		validIngredientStateIngredient := &mealplanning.ValidIngredientStateIngredient{
			CreatedAt:     result.ValidIngredientStateIngredientCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientStateIngredientLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientStateIngredientArchivedAt),
			Notes:         result.ValidIngredientStateIngredientNotes,
			ID:            result.ValidIngredientStateIngredientID,
			IngredientState: mealplanning.ValidIngredientState{
				CreatedAt:     result.ValidIngredientStateCreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientStateLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientStateArchivedAt),
				PastTense:     result.ValidIngredientStatePastTense,
				Description:   result.ValidIngredientStateDescription,
				IconPath:      result.ValidIngredientStateIconPath,
				ID:            result.ValidIngredientStateID,
				Name:          result.ValidIngredientStateName,
				AttributeType: string(result.ValidIngredientStateAttributeType),
				Slug:          result.ValidIngredientStateSlug,
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

		data = append(data, validIngredientStateIngredient)
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(visi *mealplanning.ValidIngredientStateIngredient) string { return visi.ID },
		filter,
	)

	return x, nil
}

// GetValidIngredientStateIngredientsForIngredientState fetches a list of valid ingredient state ingredients from the database that meet a particular filter.
func (q *repository) GetValidIngredientStateIngredientsForIngredientState(ctx context.Context, ingredientStateID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if ingredientStateID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientStateIDKey, ingredientStateID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, ingredientStateID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidIngredientStateIngredientsForIngredientState(ctx, q.readDB, &generated.GetValidIngredientStateIngredientsForIngredientStateParams{
		CreatedBefore:        database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:         database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:        database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:         database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:               database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:          database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived:      database.NullBoolFromBoolPointer(filter.IncludeArchived),
		ValidIngredientState: ingredientStateID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredient state ingredients list retrieval query")
	}

	var (
		data          []*mealplanning.ValidIngredientStateIngredient
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		validIngredientStateIngredient := &mealplanning.ValidIngredientStateIngredient{
			CreatedAt:     result.ValidIngredientStateIngredientCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientStateIngredientLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientStateIngredientArchivedAt),
			Notes:         result.ValidIngredientStateIngredientNotes,
			ID:            result.ValidIngredientStateIngredientID,
			IngredientState: mealplanning.ValidIngredientState{
				CreatedAt:     result.ValidIngredientStateCreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientStateLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientStateArchivedAt),
				PastTense:     result.ValidIngredientStatePastTense,
				Description:   result.ValidIngredientStateDescription,
				IconPath:      result.ValidIngredientStateIconPath,
				ID:            result.ValidIngredientStateID,
				Name:          result.ValidIngredientStateName,
				AttributeType: string(result.ValidIngredientStateAttributeType),
				Slug:          result.ValidIngredientStateSlug,
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

		data = append(data, validIngredientStateIngredient)
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(visi *mealplanning.ValidIngredientStateIngredient) string { return visi.ID },
		filter,
	)

	return x, nil
}

// GetValidIngredientStateIngredientsForIngredient fetches a list of valid ingredient state ingredients from the database that meet a particular filter.
func (q *repository) GetValidIngredientStateIngredientsForIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if ingredientID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientIDKey, ingredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, ingredientID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidIngredientStateIngredientsForIngredient(ctx, q.readDB, &generated.GetValidIngredientStateIngredientsForIngredientParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		ValidIngredient: ingredientID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredient state ingredients list retrieval query")
	}

	var (
		data          []*mealplanning.ValidIngredientStateIngredient
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		validIngredientStateIngredient := &mealplanning.ValidIngredientStateIngredient{
			CreatedAt:     result.ValidIngredientStateIngredientCreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientStateIngredientLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientStateIngredientArchivedAt),
			Notes:         result.ValidIngredientStateIngredientNotes,
			ID:            result.ValidIngredientStateIngredientID,
			IngredientState: mealplanning.ValidIngredientState{
				CreatedAt:     result.ValidIngredientStateCreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientStateLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientStateArchivedAt),
				PastTense:     result.ValidIngredientStatePastTense,
				Description:   result.ValidIngredientStateDescription,
				IconPath:      result.ValidIngredientStateIconPath,
				ID:            result.ValidIngredientStateID,
				Name:          result.ValidIngredientStateName,
				AttributeType: string(result.ValidIngredientStateAttributeType),
				Slug:          result.ValidIngredientStateSlug,
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

		data = append(data, validIngredientStateIngredient)
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(visi *mealplanning.ValidIngredientStateIngredient) string { return visi.ID },
		filter,
	)

	return x, nil
}

// CreateValidIngredientStateIngredient creates a valid ingredient state ingredient in the database.
func (q *repository) CreateValidIngredientStateIngredient(ctx context.Context, input *mealplanning.ValidIngredientStateIngredientDatabaseCreationInput) (*mealplanning.ValidIngredientStateIngredient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientStateIngredientIDKey, input.ID)

	// create the valid ingredient state ingredient.
	if err := q.withEvent(ctx, logger, mealplanning.ValidIngredientStateIngredientCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientStateIngredientIDKey: input.ID,
	}, func(tx database.SQLQueryExecutor) error {
		return q.generatedQuerier.CreateValidIngredientStateIngredient(ctx, tx, &generated.CreateValidIngredientStateIngredientParams{
			ID:                   input.ID,
			Notes:                input.Notes,
			ValidIngredientState: input.ValidIngredientStateID,
			ValidIngredient:      input.ValidIngredientID,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient state ingredient creation query")
	}

	x := &mealplanning.ValidIngredientStateIngredient{
		ID:              input.ID,
		Notes:           input.Notes,
		IngredientState: mealplanning.ValidIngredientState{ID: input.ValidIngredientStateID},
		Ingredient:      mealplanning.ValidIngredient{ID: input.ValidIngredientID},
		CreatedAt:       q.CurrentTime(),
	}

	ingredientState, err := q.GetValidIngredientState(ctx, input.ValidIngredientStateID)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient state for valid ingredient state ingredient")
	}
	if ingredientState != nil {
		x.IngredientState = *ingredientState
	}

	vessel, err := q.GetValidIngredient(ctx, input.ValidIngredientID)
	if err != nil {
		// basically impossible for this to happen and not error out earlier
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient for valid ingredient state ingredient")
	}
	if vessel != nil {
		x.Ingredient = *vessel
	}

	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, x.ID)
	logger.Info("valid ingredient state ingredient created")

	return x, nil
}

// UpdateValidIngredientStateIngredient updates a particular valid ingredient state ingredient.
func (q *repository) UpdateValidIngredientStateIngredient(ctx context.Context, updated *mealplanning.ValidIngredientStateIngredient) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientStateIngredientIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, mealplanning.ValidIngredientStateIngredientUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientStateIngredientIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		_, updateErr := q.generatedQuerier.UpdateValidIngredientStateIngredient(ctx, tx, &generated.UpdateValidIngredientStateIngredientParams{
			Notes:                updated.Notes,
			ValidIngredientState: updated.IngredientState.ID,
			ValidIngredient:      updated.Ingredient.ID,
			ID:                   updated.ID,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid ingredient state ingredient")
	}

	logger.Info("valid ingredient state ingredient updated")

	return nil
}

// ArchiveValidIngredientStateIngredient archives a valid ingredient state ingredient from the database by its ID.
func (q *repository) ArchiveValidIngredientStateIngredient(ctx context.Context, validIngredientStateIngredientID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientStateIngredientID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientStateIngredientIDKey, validIngredientStateIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientStateIngredientIDKey, validIngredientStateIngredientID)

	return q.withEvent(ctx, logger, mealplanning.ValidIngredientStateIngredientArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientStateIngredientIDKey: validIngredientStateIngredientID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveValidIngredientStateIngredient(ctx, tx, validIngredientStateIngredientID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving valid ingredient state ingredient")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}
