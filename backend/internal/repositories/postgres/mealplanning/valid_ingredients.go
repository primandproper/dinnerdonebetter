package mealplanning

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/observability"
	platformkeys "github.com/primandproper/platform-go/v9/observability/keys"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

var (
	_ mealplanning.ValidIngredientDataManager = (*repository)(nil)
)

// ValidIngredientExists fetches whether a valid ingredient exists from the database.
func (q *repository) ValidIngredientExists(ctx context.Context, validIngredientID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	result, err := q.generatedQuerier.CheckValidIngredientExistence(ctx, q.readDB, validIngredientID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient existence check")
	}

	return result, nil
}

// GetValidIngredient fetches a valid ingredient from the database.
func (q *repository) GetValidIngredient(ctx context.Context, validIngredientID string) (*mealplanning.ValidIngredient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	result, err := q.generatedQuerier.GetValidIngredient(ctx, q.readDB, validIngredientID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient")
	}

	validIngredient := &mealplanning.ValidIngredient{
		CreatedAt:                      result.CreatedAt,
		LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
		MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MinimumIdealStorageTemperatureInCelsius),
		MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MaximumIdealStorageTemperatureInCelsius),
		IconPath:                       result.IconPath,
		Warning:                        result.Warning,
		PluralName:                     result.PluralName,
		StorageInstructions:            result.StorageInstructions,
		Name:                           result.Name,
		ID:                             result.ID,
		Description:                    result.Description,
		Slug:                           result.Slug,
		ShoppingSuggestions:            result.ShoppingSuggestions,
		ContainsShellfish:              result.ContainsShellfish,
		IsLiquid:                       database.BoolFromNullBool(result.IsLiquid),
		ContainsPeanut:                 result.ContainsPeanut,
		ContainsTreeNut:                result.ContainsTreeNut,
		ContainsEgg:                    result.ContainsEgg,
		ContainsWheat:                  result.ContainsWheat,
		ContainsSoy:                    result.ContainsSoy,
		AnimalDerived:                  result.AnimalDerived,
		RestrictToPreparations:         result.RestrictToPreparations,
		ContaminatesEquipment:          result.ContaminatesEquipment,
		ContainsSesame:                 result.ContainsSesame,
		ContainsFish:                   result.ContainsFish,
		ContainsGluten:                 result.ContainsGluten,
		ContainsDairy:                  result.ContainsDairy,
		ContainsAlcohol:                result.ContainsAlcohol,
		AnimalFlesh:                    result.AnimalFlesh,
		IsStarch:                       result.IsStarch,
		IsProtein:                      result.IsProtein,
		IsGrain:                        result.IsGrain,
		IsFruit:                        result.IsFruit,
		IsSalt:                         result.IsSalt,
		IsFat:                          result.IsFat,
		IsAcid:                         result.IsAcid,
		IsHeat:                         result.IsHeat,
	}

	return validIngredient, nil
}

// GetRandomValidIngredient fetches a valid ingredient from the database.
func (q *repository) GetRandomValidIngredient(ctx context.Context) (*mealplanning.ValidIngredient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	result, err := q.generatedQuerier.GetRandomValidIngredient(ctx, q.readDB)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching random valid ingredient")
	}

	validIngredient := &mealplanning.ValidIngredient{
		CreatedAt:                      result.CreatedAt,
		LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
		MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MinimumIdealStorageTemperatureInCelsius),
		MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MaximumIdealStorageTemperatureInCelsius),
		IconPath:                       result.IconPath,
		Warning:                        result.Warning,
		PluralName:                     result.PluralName,
		StorageInstructions:            result.StorageInstructions,
		Name:                           result.Name,
		ID:                             result.ID,
		Description:                    result.Description,
		Slug:                           result.Slug,
		ShoppingSuggestions:            result.ShoppingSuggestions,
		ContainsShellfish:              result.ContainsShellfish,
		IsLiquid:                       database.BoolFromNullBool(result.IsLiquid),
		ContainsPeanut:                 result.ContainsPeanut,
		ContainsTreeNut:                result.ContainsTreeNut,
		ContainsEgg:                    result.ContainsEgg,
		ContainsWheat:                  result.ContainsWheat,
		ContainsSoy:                    result.ContainsSoy,
		AnimalDerived:                  result.AnimalDerived,
		RestrictToPreparations:         result.RestrictToPreparations,
		ContaminatesEquipment:          result.ContaminatesEquipment,
		ContainsSesame:                 result.ContainsSesame,
		ContainsFish:                   result.ContainsFish,
		ContainsGluten:                 result.ContainsGluten,
		ContainsDairy:                  result.ContainsDairy,
		ContainsAlcohol:                result.ContainsAlcohol,
		AnimalFlesh:                    result.AnimalFlesh,
		IsStarch:                       result.IsStarch,
		IsProtein:                      result.IsProtein,
		IsGrain:                        result.IsGrain,
		IsFruit:                        result.IsFruit,
		IsSalt:                         result.IsSalt,
		IsFat:                          result.IsFat,
		IsAcid:                         result.IsAcid,
		IsHeat:                         result.IsHeat,
	}

	return validIngredient, nil
}

// SearchForValidIngredients fetches a valid ingredient from the database.
func (q *repository) SearchForValidIngredients(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredient], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if query == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, query)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.SearchForValidIngredients(ctx, q.readDB, &generated.SearchForValidIngredientsParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		NameQuery:       query,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient")
	}

	var data []*mealplanning.ValidIngredient
	for _, result := range results {
		validIngredient := &mealplanning.ValidIngredient{
			CreatedAt:                      result.CreatedAt,
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
			MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MinimumIdealStorageTemperatureInCelsius),
			MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MaximumIdealStorageTemperatureInCelsius),
			IconPath:                       result.IconPath,
			Warning:                        result.Warning,
			PluralName:                     result.PluralName,
			StorageInstructions:            result.StorageInstructions,
			Name:                           result.Name,
			ID:                             result.ID,
			Description:                    result.Description,
			Slug:                           result.Slug,
			ShoppingSuggestions:            result.ShoppingSuggestions,
			ContainsShellfish:              result.ContainsShellfish,
			IsLiquid:                       database.BoolFromNullBool(result.IsLiquid),
			ContainsPeanut:                 result.ContainsPeanut,
			ContainsTreeNut:                result.ContainsTreeNut,
			ContainsEgg:                    result.ContainsEgg,
			ContainsWheat:                  result.ContainsWheat,
			ContainsSoy:                    result.ContainsSoy,
			AnimalDerived:                  result.AnimalDerived,
			RestrictToPreparations:         result.RestrictToPreparations,
			ContaminatesEquipment:          result.ContaminatesEquipment,
			ContainsSesame:                 result.ContainsSesame,
			ContainsFish:                   result.ContainsFish,
			ContainsGluten:                 result.ContainsGluten,
			ContainsDairy:                  result.ContainsDairy,
			ContainsAlcohol:                result.ContainsAlcohol,
			AnimalFlesh:                    result.AnimalFlesh,
			IsStarch:                       result.IsStarch,
			IsProtein:                      result.IsProtein,
			IsGrain:                        result.IsGrain,
			IsFruit:                        result.IsFruit,
			IsSalt:                         result.IsSalt,
			IsFat:                          result.IsFat,
			IsAcid:                         result.IsAcid,
			IsHeat:                         result.IsHeat,
		}

		data = append(data, validIngredient)
	}

	return filtering.NewQueryFilteredResult(
		data,
		0,
		0,
		func(vi *mealplanning.ValidIngredient) string { return vi.ID },
		filter,
	), nil
}

// SearchForValidIngredientsForPreparation fetches a list of valid ingredient preparations from the database that meet a particular filter.
func (q *repository) SearchForValidIngredientsForPreparation(ctx context.Context, preparationID, query string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidIngredient], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if preparationID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidPreparationIDKey, preparationID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientPreparationIDKey, preparationID)

	if query == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, platformkeys.SearchQueryKey, query)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	logger = filter.AttachToLogger(logger)

	results, err := q.generatedQuerier.SearchValidIngredientsByPreparationAndIngredientName(ctx, q.readDB, &generated.SearchValidIngredientsByPreparationAndIngredientNameParams{
		ValidPreparationID: preparationID,
		NameQuery:          query,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredients for preparation")
	}

	var data []*mealplanning.ValidIngredient

	for _, result := range results {
		validIngredient := &mealplanning.ValidIngredient{
			CreatedAt:                      result.CreatedAt,
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
			MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MinimumIdealStorageTemperatureInCelsius),
			MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MaximumIdealStorageTemperatureInCelsius),
			IconPath:                       result.IconPath,
			Warning:                        result.Warning,
			PluralName:                     result.PluralName,
			StorageInstructions:            result.StorageInstructions,
			Name:                           result.Name,
			ID:                             result.ID,
			Description:                    result.Description,
			Slug:                           result.Slug,
			ShoppingSuggestions:            result.ShoppingSuggestions,
			ContainsShellfish:              result.ContainsShellfish,
			IsLiquid:                       database.BoolFromNullBool(result.IsLiquid),
			ContainsPeanut:                 result.ContainsPeanut,
			ContainsTreeNut:                result.ContainsTreeNut,
			ContainsEgg:                    result.ContainsEgg,
			ContainsWheat:                  result.ContainsWheat,
			ContainsSoy:                    result.ContainsSoy,
			AnimalDerived:                  result.AnimalDerived,
			RestrictToPreparations:         result.RestrictToPreparations,
			ContaminatesEquipment:          result.ContaminatesEquipment,
			ContainsSesame:                 result.ContainsSesame,
			ContainsFish:                   result.ContainsFish,
			ContainsGluten:                 result.ContainsGluten,
			ContainsDairy:                  result.ContainsDairy,
			ContainsAlcohol:                result.ContainsAlcohol,
			AnimalFlesh:                    result.AnimalFlesh,
			IsStarch:                       result.IsStarch,
			IsProtein:                      result.IsProtein,
			IsGrain:                        result.IsGrain,
			IsFruit:                        result.IsFruit,
			IsSalt:                         result.IsSalt,
			IsFat:                          result.IsFat,
			IsAcid:                         result.IsAcid,
			IsHeat:                         result.IsHeat,
		}

		data = append(data, validIngredient)
	}

	return filtering.NewQueryFilteredResult(
		data,
		0,
		0,
		func(vi *mealplanning.ValidIngredient) string { return vi.ID },
		filter,
	), nil
}

// GetValidIngredients fetches a list of valid ingredients from the database that meet a particular filter.
func (q *repository) GetValidIngredients(ctx context.Context, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidIngredient], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidIngredients(ctx, q.readDB, &generated.GetValidIngredientsParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredients list retrieval query")
	}

	var (
		data          []*mealplanning.ValidIngredient
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		validIngredient := &mealplanning.ValidIngredient{
			CreatedAt:                      result.CreatedAt,
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
			MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MinimumIdealStorageTemperatureInCelsius),
			MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MaximumIdealStorageTemperatureInCelsius),
			IconPath:                       result.IconPath,
			Warning:                        result.Warning,
			PluralName:                     result.PluralName,
			StorageInstructions:            result.StorageInstructions,
			Name:                           result.Name,
			ID:                             result.ID,
			Description:                    result.Description,
			Slug:                           result.Slug,
			ShoppingSuggestions:            result.ShoppingSuggestions,
			ContainsShellfish:              result.ContainsShellfish,
			IsLiquid:                       database.BoolFromNullBool(result.IsLiquid),
			ContainsPeanut:                 result.ContainsPeanut,
			ContainsTreeNut:                result.ContainsTreeNut,
			ContainsEgg:                    result.ContainsEgg,
			ContainsWheat:                  result.ContainsWheat,
			ContainsSoy:                    result.ContainsSoy,
			AnimalDerived:                  result.AnimalDerived,
			RestrictToPreparations:         result.RestrictToPreparations,
			ContaminatesEquipment:          result.ContaminatesEquipment,
			ContainsSesame:                 result.ContainsSesame,
			ContainsFish:                   result.ContainsFish,
			ContainsGluten:                 result.ContainsGluten,
			ContainsDairy:                  result.ContainsDairy,
			ContainsAlcohol:                result.ContainsAlcohol,
			AnimalFlesh:                    result.AnimalFlesh,
			IsStarch:                       result.IsStarch,
			IsProtein:                      result.IsProtein,
			IsGrain:                        result.IsGrain,
			IsFruit:                        result.IsFruit,
			IsSalt:                         result.IsSalt,
			IsFat:                          result.IsFat,
			IsAcid:                         result.IsAcid,
			IsHeat:                         result.IsHeat,
		}

		data = append(data, validIngredient)
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vi *mealplanning.ValidIngredient) string { return vi.ID },
		filter,
	)

	return x, nil
}

// GetValidIngredientsWithIDs fetches a list of valid ingredients from the database that meet a particular filter.
func (q *repository) GetValidIngredientsWithIDs(ctx context.Context, ids []string) ([]*mealplanning.ValidIngredient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if ids == nil {
		return nil, platformerrors.ErrEmptyInputProvided
	}

	results, err := q.generatedQuerier.GetValidIngredientsWithIDs(ctx, q.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing valid ingredients id list retrieval query")
	}

	var ingredients []*mealplanning.ValidIngredient
	for _, result := range results {
		validIngredient := &mealplanning.ValidIngredient{
			CreatedAt:                      result.CreatedAt,
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
			MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MinimumIdealStorageTemperatureInCelsius),
			MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MaximumIdealStorageTemperatureInCelsius),
			IconPath:                       result.IconPath,
			Warning:                        result.Warning,
			PluralName:                     result.PluralName,
			StorageInstructions:            result.StorageInstructions,
			Name:                           result.Name,
			ID:                             result.ID,
			Description:                    result.Description,
			Slug:                           result.Slug,
			ShoppingSuggestions:            result.ShoppingSuggestions,
			ContainsShellfish:              result.ContainsShellfish,
			IsLiquid:                       database.BoolFromNullBool(result.IsLiquid),
			ContainsPeanut:                 result.ContainsPeanut,
			ContainsTreeNut:                result.ContainsTreeNut,
			ContainsEgg:                    result.ContainsEgg,
			ContainsWheat:                  result.ContainsWheat,
			ContainsSoy:                    result.ContainsSoy,
			AnimalDerived:                  result.AnimalDerived,
			RestrictToPreparations:         result.RestrictToPreparations,
			ContaminatesEquipment:          result.ContaminatesEquipment,
			ContainsSesame:                 result.ContainsSesame,
			ContainsFish:                   result.ContainsFish,
			ContainsGluten:                 result.ContainsGluten,
			ContainsDairy:                  result.ContainsDairy,
			ContainsAlcohol:                result.ContainsAlcohol,
			AnimalFlesh:                    result.AnimalFlesh,
			IsStarch:                       result.IsStarch,
			IsProtein:                      result.IsProtein,
			IsGrain:                        result.IsGrain,
			IsFruit:                        result.IsFruit,
			IsSalt:                         result.IsSalt,
			IsFat:                          result.IsFat,
			IsAcid:                         result.IsAcid,
			IsHeat:                         result.IsHeat,
		}

		ingredients = append(ingredients, validIngredient)
	}

	return ingredients, nil
}

// GetValidIngredientIDsThatNeedSearchIndexing fetches a list of valid ingredients from the database that meet a particular filter.
func (q *repository) GetValidIngredientIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	results, err := q.generatedQuerier.GetValidIngredientsNeedingIndexing(ctx, q.readDB)
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing valid ingredients list retrieval query")
	}

	return results, err
}

// CreateValidIngredient creates a valid ingredient in the database.
func (q *repository) CreateValidIngredient(ctx context.Context, input *mealplanning.ValidIngredientDatabaseCreationInput) (*mealplanning.ValidIngredient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientIDKey, input.ID)

	// create the valid ingredient.
	if err := q.withEvent(ctx, logger, mealplanning.ValidIngredientCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientIDKey: input.ID,
	}, func(tx database.SQLQueryExecutor) error {
		return q.generatedQuerier.CreateValidIngredient(ctx, tx, &generated.CreateValidIngredientParams{
			ID:                                      input.ID,
			Name:                                    input.Name,
			Description:                             input.Description,
			Warning:                                 input.Warning,
			ContainsEgg:                             input.ContainsEgg,
			ContainsDairy:                           input.ContainsDairy,
			ContainsPeanut:                          input.ContainsPeanut,
			ContainsTreeNut:                         input.ContainsTreeNut,
			ContainsSoy:                             input.ContainsSoy,
			ContainsWheat:                           input.ContainsWheat,
			ContainsShellfish:                       input.ContainsShellfish,
			ContainsSesame:                          input.ContainsSesame,
			ContainsFish:                            input.ContainsFish,
			ContainsGluten:                          input.ContainsGluten,
			AnimalFlesh:                             input.AnimalFlesh,
			IsLiquid:                                database.NullBoolFromBool(input.IsLiquid),
			IconPath:                                input.IconPath,
			AnimalDerived:                           input.AnimalDerived,
			PluralName:                              input.PluralName,
			RestrictToPreparations:                  input.RestrictToPreparations,
			ContaminatesEquipment:                   input.ContaminatesEquipment,
			MaximumIdealStorageTemperatureInCelsius: database.NullStringFromFloat32Pointer(input.MaxStorageTemperatureInCelsius),
			MinimumIdealStorageTemperatureInCelsius: database.NullStringFromFloat32Pointer(input.MinStorageTemperatureInCelsius),
			StorageInstructions:                     input.StorageInstructions,
			Slug:                                    input.Slug,
			ContainsAlcohol:                         input.ContainsAlcohol,
			ShoppingSuggestions:                     input.ShoppingSuggestions,
			IsStarch:                                input.IsStarch,
			IsProtein:                               input.IsProtein,
			IsGrain:                                 input.IsGrain,
			IsFruit:                                 input.IsFruit,
			IsSalt:                                  input.IsSalt,
			IsFat:                                   input.IsFat,
			IsAcid:                                  input.IsAcid,
			IsHeat:                                  input.IsHeat,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient creation query")
	}

	x := &mealplanning.ValidIngredient{
		ID:                             input.ID,
		Name:                           input.Name,
		Description:                    input.Description,
		Warning:                        input.Warning,
		ContainsEgg:                    input.ContainsEgg,
		ContainsDairy:                  input.ContainsDairy,
		ContainsPeanut:                 input.ContainsPeanut,
		ContainsTreeNut:                input.ContainsTreeNut,
		ContainsSoy:                    input.ContainsSoy,
		ContainsWheat:                  input.ContainsWheat,
		ContainsShellfish:              input.ContainsShellfish,
		ContainsSesame:                 input.ContainsSesame,
		ContainsFish:                   input.ContainsFish,
		ContainsGluten:                 input.ContainsGluten,
		AnimalFlesh:                    input.AnimalFlesh,
		IsLiquid:                       input.IsLiquid,
		IconPath:                       input.IconPath,
		AnimalDerived:                  input.AnimalDerived,
		PluralName:                     input.PluralName,
		IsStarch:                       input.IsStarch,
		IsProtein:                      input.IsProtein,
		IsGrain:                        input.IsGrain,
		IsFruit:                        input.IsFruit,
		IsSalt:                         input.IsSalt,
		IsFat:                          input.IsFat,
		IsAcid:                         input.IsAcid,
		IsHeat:                         input.IsHeat,
		RestrictToPreparations:         input.RestrictToPreparations,
		ContaminatesEquipment:          input.ContaminatesEquipment,
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		StorageInstructions:            input.StorageInstructions,
		Slug:                           input.Slug,
		ContainsAlcohol:                input.ContainsAlcohol,
		ShoppingSuggestions:            input.ShoppingSuggestions,
		CreatedAt:                      q.CurrentTime(),
	}

	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, x.ID)
	logger.Info("valid ingredient created")

	return x, nil
}

// UpdateValidIngredient updates a particular valid ingredient.
func (q *repository) UpdateValidIngredient(ctx context.Context, updated *mealplanning.ValidIngredient) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, mealplanning.ValidIngredientUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		_, updateErr := q.generatedQuerier.UpdateValidIngredient(ctx, tx, &generated.UpdateValidIngredientParams{
			Description:                             updated.Description,
			Warning:                                 updated.Warning,
			ID:                                      updated.ID,
			ShoppingSuggestions:                     updated.ShoppingSuggestions,
			Slug:                                    updated.Slug,
			StorageInstructions:                     updated.StorageInstructions,
			Name:                                    updated.Name,
			PluralName:                              updated.PluralName,
			IconPath:                                updated.IconPath,
			MaximumIdealStorageTemperatureInCelsius: database.NullStringFromFloat32Pointer(updated.MaxStorageTemperatureInCelsius),
			MinimumIdealStorageTemperatureInCelsius: database.NullStringFromFloat32Pointer(updated.MinStorageTemperatureInCelsius),
			IsLiquid:                                database.NullBoolFromBool(updated.IsLiquid),
			ContainsWheat:                           updated.ContainsWheat,
			ContainsPeanut:                          updated.ContainsPeanut,
			ContainsGluten:                          updated.ContainsGluten,
			ContainsFish:                            updated.ContainsFish,
			AnimalDerived:                           updated.AnimalDerived,
			ContainsSesame:                          updated.ContainsSesame,
			RestrictToPreparations:                  updated.RestrictToPreparations,
			ContaminatesEquipment:                   updated.ContaminatesEquipment,
			ContainsShellfish:                       updated.ContainsShellfish,
			ContainsSoy:                             updated.ContainsSoy,
			ContainsTreeNut:                         updated.ContainsTreeNut,
			AnimalFlesh:                             updated.AnimalFlesh,
			ContainsAlcohol:                         updated.ContainsAlcohol,
			ContainsDairy:                           updated.ContainsDairy,
			IsStarch:                                updated.IsStarch,
			IsProtein:                               updated.IsProtein,
			IsGrain:                                 updated.IsGrain,
			IsFruit:                                 updated.IsFruit,
			IsSalt:                                  updated.IsSalt,
			IsFat:                                   updated.IsFat,
			IsAcid:                                  updated.IsAcid,
			IsHeat:                                  updated.IsHeat,
			ContainsEgg:                             updated.ContainsEgg,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid ingredient")
	}

	logger.Info("valid ingredient updated")

	return nil
}

// MarkValidIngredientAsIndexed updates a particular valid ingredient's last_indexed_at value.
func (q *repository) MarkValidIngredientAsIndexed(ctx context.Context, validIngredientID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	if _, err := q.generatedQuerier.UpdateValidIngredientLastIndexedAt(ctx, q.writeDB, validIngredientID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking valid ingredient as indexed")
	}

	logger.Info("valid ingredient marked as indexed")

	return nil
}

// ArchiveValidIngredient archives a valid ingredient from the database by its ID.
func (q *repository) ArchiveValidIngredient(ctx context.Context, validIngredientID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, validIngredientID)

	return q.withEvent(ctx, logger, mealplanning.ValidIngredientArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientIDKey: validIngredientID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveValidIngredient(ctx, tx, validIngredientID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving valid ingredient")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}
