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
	_ mealplanning.MealPlanGroceryListItemDataManager = (*repository)(nil)
)

// MealPlanGroceryListItemExists fetches whether a meal plan grocery list exists from the database.
func (q *repository) MealPlanGroceryListItemExists(ctx context.Context, mealPlanID, mealPlanGroceryListItemID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if mealPlanGroceryListItemID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanGroceryListItemIDKey, mealPlanGroceryListItemID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanGroceryListItemIDKey, mealPlanGroceryListItemID)

	result, err := q.generatedQuerier.CheckMealPlanGroceryListItemExistence(ctx, q.readDB, &generated.CheckMealPlanGroceryListItemExistenceParams{
		MealPlanID:                mealPlanID,
		MealPlanGroceryListItemID: mealPlanGroceryListItemID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing meal plan grocery list existence check")
	}

	return result, nil
}

func (q *repository) fleshOutMealPlanGroceryListItem(ctx context.Context, mealPlanGroceryListItem *mealplanning.MealPlanGroceryListItem) (*mealplanning.MealPlanGroceryListItem, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanGroceryListItem == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanGroceryListItemIDKey, mealPlanGroceryListItem.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanGroceryListItemIDKey, mealPlanGroceryListItem.ID)

	validIngredient, err := q.GetValidIngredient(ctx, mealPlanGroceryListItem.Ingredient.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching grocery list item ingredient")
	}
	mealPlanGroceryListItem.Ingredient = *validIngredient

	validMeasurementUnit, err := q.GetValidMeasurementUnit(ctx, mealPlanGroceryListItem.MeasurementUnit.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching grocery list item measurement unit")
	}
	mealPlanGroceryListItem.MeasurementUnit = *validMeasurementUnit

	if mealPlanGroceryListItem.PurchasedMeasurementUnit != nil {
		purchasedMeasurementUnit, getPurchasedMeasurementUnitErr := q.GetValidMeasurementUnit(ctx, mealPlanGroceryListItem.PurchasedMeasurementUnit.ID)
		if getPurchasedMeasurementUnitErr != nil {
			return nil, observability.PrepareAndLogError(getPurchasedMeasurementUnitErr, logger, span, "fetching grocery list item purchased measurement unit")
		}
		mealPlanGroceryListItem.PurchasedMeasurementUnit = purchasedMeasurementUnit
	}

	return mealPlanGroceryListItem, nil
}

// GetMealPlanGroceryListItem fetches a meal plan grocery list from the database.
func (q *repository) GetMealPlanGroceryListItem(ctx context.Context, mealPlanID, mealPlanGroceryListItemID string) (*mealplanning.MealPlanGroceryListItem, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if mealPlanGroceryListItemID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanGroceryListItemIDKey, mealPlanGroceryListItemID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanGroceryListItemIDKey, mealPlanGroceryListItemID)

	result, err := q.generatedQuerier.GetMealPlanGroceryListItem(ctx, q.readDB, &generated.GetMealPlanGroceryListItemParams{
		MealPlanID:                mealPlanID,
		MealPlanGroceryListItemID: mealPlanGroceryListItemID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan grocery list item")
	}

	mealPlanGroceryListItem := &mealplanning.MealPlanGroceryListItem{
		CreatedAt:         result.CreatedAt,
		LastUpdatedAt:     database.TimePointerFromNullTime(result.LastUpdatedAt),
		PurchasePrice:     database.Float32PointerFromNullString(result.PurchasePrice),
		PurchasedUPC:      database.StringPointerFromNullString(result.PurchasedUpc),
		ArchivedAt:        database.TimePointerFromNullTime(result.ArchivedAt),
		QuantityPurchased: database.Float32PointerFromNullString(result.QuantityPurchased),

		BelongsToMealPlan:       result.BelongsToMealPlan,
		Status:                  string(result.Status),
		StatusExplanation:       result.StatusExplanation,
		ID:                      result.ID,
		MinQuantityNeeded:       database.Float32FromString(result.MinimumQuantityNeeded),
		MaxQuantityNeeded:       database.Float32PointerFromNullString(result.MaximumQuantityNeeded),
		BelongsToMealPlanOption: database.StringPointerFromNullString(result.BelongsToMealPlanOption),
		RecipeID:                database.StringPointerFromNullString(result.RecipeID),
		RecipeStepID:            database.StringPointerFromNullString(result.RecipeStepID),
		IngredientIndex:         database.Uint16PointerFromNullInt32(result.IngredientIndex),
		OptionIndex:             database.Uint16PointerFromNullInt32(result.OptionIndex),
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

	if result.PurchasedMeasurementUnit.Valid {
		mealPlanGroceryListItem.PurchasedMeasurementUnit = &mealplanning.ValidMeasurementUnit{
			ID: result.PurchasedMeasurementUnit.String,
		}
	}

	if mealPlanGroceryListItem.PurchasedMeasurementUnit != nil {
		purchasedMeasurementUnit, getPurchasedMeasurementUnitErr := q.GetValidMeasurementUnit(ctx, mealPlanGroceryListItem.PurchasedMeasurementUnit.ID)
		if getPurchasedMeasurementUnitErr != nil {
			return nil, observability.PrepareAndLogError(getPurchasedMeasurementUnitErr, logger, span, "fetching grocery list item purchased measurement unit")
		}
		mealPlanGroceryListItem.PurchasedMeasurementUnit = purchasedMeasurementUnit
	}

	return mealPlanGroceryListItem, nil
}

// GetMealPlanGroceryListItemsForMealPlan fetches a list of meal plan grocery lists from the database that meet a particular filter.
func (q *repository) GetMealPlanGroceryListItemsForMealPlan(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanGroceryListItem], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetMealPlanGroceryListItemsForMealPlan(ctx, q.readDB, &generated.GetMealPlanGroceryListItemsForMealPlanParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		MealPlanID:      mealPlanID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing meal plan grocery list items list retrieval query")
	}

	var (
		x                         = []*mealplanning.MealPlanGroceryListItem{}
		filteredCount, totalCount uint64
	)

	for _, result := range results {
		mealPlanGroceryListItem := &mealplanning.MealPlanGroceryListItem{
			CreatedAt:               result.CreatedAt,
			LastUpdatedAt:           database.TimePointerFromNullTime(result.LastUpdatedAt),
			PurchasePrice:           database.Float32PointerFromNullString(result.PurchasePrice),
			PurchasedUPC:            database.StringPointerFromNullString(result.PurchasedUpc),
			ArchivedAt:              database.TimePointerFromNullTime(result.ArchivedAt),
			QuantityPurchased:       database.Float32PointerFromNullString(result.QuantityPurchased),
			BelongsToMealPlan:       result.BelongsToMealPlan,
			Status:                  string(result.Status),
			StatusExplanation:       result.StatusExplanation,
			ID:                      result.ID,
			MinQuantityNeeded:       database.Float32FromString(result.MinimumQuantityNeeded),
			MaxQuantityNeeded:       database.Float32PointerFromNullString(result.MaximumQuantityNeeded),
			BelongsToMealPlanOption: database.StringPointerFromNullString(result.BelongsToMealPlanOption),
			RecipeID:                database.StringPointerFromNullString(result.RecipeID),
			RecipeStepID:            database.StringPointerFromNullString(result.RecipeStepID),
			IngredientIndex:         database.Uint16PointerFromNullInt32(result.IngredientIndex),
			OptionIndex:             database.Uint16PointerFromNullInt32(result.OptionIndex),
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

		if result.PurchasedMeasurementUnit.Valid {
			mealPlanGroceryListItem.PurchasedMeasurementUnit = &mealplanning.ValidMeasurementUnit{
				ID: result.PurchasedMeasurementUnit.String,
			}
		}

		if mealPlanGroceryListItem.PurchasedMeasurementUnit != nil {
			purchasedMeasurementUnit, getPurchasedMeasurementUnitErr := q.GetValidMeasurementUnit(ctx, mealPlanGroceryListItem.PurchasedMeasurementUnit.ID)
			if getPurchasedMeasurementUnitErr != nil {
				return nil, observability.PrepareAndLogError(getPurchasedMeasurementUnitErr, logger, span, "fetching grocery list item purchased measurement unit")
			}
			mealPlanGroceryListItem.PurchasedMeasurementUnit = purchasedMeasurementUnit
		}

		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)

		x = append(x, mealPlanGroceryListItem)
	}

	for i := range x {
		x[i], err = q.fleshOutMealPlanGroceryListItem(ctx, x[i])
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "augmenting grocery list item data")
		}
	}

	y := filtering.NewQueryFilteredResult(x, filteredCount, totalCount, func(m *mealplanning.MealPlanGroceryListItem) string { return m.ID }, filter)

	return y, nil
}

// createMealPlanGroceryListItem creates a meal plan grocery list in the database.
func (q *repository) createMealPlanGroceryListItem(ctx context.Context, querier database.SQLQueryExecutor, input *mealplanning.MealPlanGroceryListItemDatabaseCreationInput) (*mealplanning.MealPlanGroceryListItem, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	logger := q.logger.WithValue(mealplanningkeys.MealPlanGroceryListItemIDKey, input.ID)

	// create the meal plan grocery list.
	if err := q.generatedQuerier.CreateMealPlanGroceryListItem(ctx, querier, &generated.CreateMealPlanGroceryListItemParams{
		ID:                       input.ID,
		BelongsToMealPlan:        input.BelongsToMealPlan,
		ValidIngredient:          input.ValidIngredientID,
		ValidMeasurementUnit:     input.ValidMeasurementUnitID,
		MinimumQuantityNeeded:    database.StringFromFloat32(input.MinQuantityNeeded),
		StatusExplanation:        input.StatusExplanation,
		Status:                   generated.GroceryListItemStatus(input.Status),
		MaximumQuantityNeeded:    database.NullStringFromFloat32Pointer(input.MaxQuantityNeeded),
		QuantityPurchased:        database.NullStringFromFloat32Pointer(input.QuantityPurchased),
		PurchasedMeasurementUnit: database.NullStringFromStringPointer(input.PurchasedMeasurementUnitID),
		PurchasedUpc:             database.NullStringFromStringPointer(input.PurchasedUPC),
		PurchasePrice:            database.NullStringFromFloat32Pointer(input.PurchasePrice),
		BelongsToMealPlanOption:  database.NullStringFromStringPointer(input.BelongsToMealPlanOption),
		RecipeID:                 database.NullStringFromStringPointer(input.RecipeID),
		RecipeStepID:             database.NullStringFromStringPointer(input.RecipeStepID),
		IngredientIndex:          database.NullInt32FromUint16Pointer(input.IngredientIndex),
		OptionIndex:              database.NullInt32FromUint16Pointer(input.OptionIndex),
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing meal plan grocery list creation query")
	}

	x := &mealplanning.MealPlanGroceryListItem{
		ID:                      input.ID,
		BelongsToMealPlan:       input.BelongsToMealPlan,
		Ingredient:              mealplanning.ValidIngredient{ID: input.ValidIngredientID},
		MeasurementUnit:         mealplanning.ValidMeasurementUnit{ID: input.ValidMeasurementUnitID},
		MinQuantityNeeded:       input.MinQuantityNeeded,
		MaxQuantityNeeded:       input.MaxQuantityNeeded,
		QuantityPurchased:       input.QuantityPurchased,
		PurchasedUPC:            input.PurchasedUPC,
		PurchasePrice:           input.PurchasePrice,
		StatusExplanation:       input.StatusExplanation,
		Status:                  input.Status,
		CreatedAt:               q.CurrentTime(),
		BelongsToMealPlanOption: input.BelongsToMealPlanOption,
		RecipeID:                input.RecipeID,
		RecipeStepID:            input.RecipeStepID,
		IngredientIndex:         input.IngredientIndex,
		OptionIndex:             input.OptionIndex,
	}

	if input.PurchasedMeasurementUnitID != nil {
		x.PurchasedMeasurementUnit = &mealplanning.ValidMeasurementUnit{ID: *input.PurchasedMeasurementUnitID}
	}

	tracing.AttachToSpan(span, mealplanningkeys.MealPlanGroceryListItemIDKey, x.ID)
	logger.Info("meal plan grocery list created")

	return x, nil
}

// CreateMealPlanGroceryListItem creates a meal plan grocery list in the database.
func (q *repository) CreateMealPlanGroceryListItem(ctx context.Context, input *mealplanning.MealPlanGroceryListItemDatabaseCreationInput) (*mealplanning.MealPlanGroceryListItem, error) {
	return q.createMealPlanGroceryListItem(ctx, q.writeDB, input)
}

// UpdateMealPlanGroceryListItem updates a particular meal plan grocery list.
func (q *repository) UpdateMealPlanGroceryListItem(ctx context.Context, updated *mealplanning.MealPlanGroceryListItem) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.MealPlanGroceryListItemIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanGroceryListItemIDKey, updated.ID)

	var purchasedMeasurementUnitID *string
	if updated.PurchasedMeasurementUnit != nil {
		purchasedMeasurementUnitID = &updated.PurchasedMeasurementUnit.ID
	}

	if _, err := q.generatedQuerier.UpdateMealPlanGroceryListItem(ctx, q.writeDB, &generated.UpdateMealPlanGroceryListItemParams{
		BelongsToMealPlanOption:  database.NullStringFromStringPointer(updated.BelongsToMealPlanOption),
		RecipeID:                 database.NullStringFromStringPointer(updated.RecipeID),
		RecipeStepID:             database.NullStringFromStringPointer(updated.RecipeStepID),
		IngredientIndex:          database.NullInt32FromUint16Pointer(updated.IngredientIndex),
		OptionIndex:              database.NullInt32FromUint16Pointer(updated.OptionIndex),
		BelongsToMealPlan:        updated.BelongsToMealPlan,
		ValidIngredient:          updated.Ingredient.ID,
		ValidMeasurementUnit:     updated.MeasurementUnit.ID,
		MinimumQuantityNeeded:    database.StringFromFloat32(updated.MinQuantityNeeded),
		StatusExplanation:        updated.StatusExplanation,
		Status:                   generated.GroceryListItemStatus(updated.Status),
		ID:                       updated.ID,
		MaximumQuantityNeeded:    database.NullStringFromFloat32Pointer(updated.MaxQuantityNeeded),
		QuantityPurchased:        database.NullStringFromFloat32Pointer(updated.QuantityPurchased),
		PurchasedMeasurementUnit: database.NullStringFromStringPointer(purchasedMeasurementUnitID),
		PurchasedUpc:             database.NullStringFromStringPointer(updated.PurchasedUPC),
		PurchasePrice:            database.NullStringFromFloat32Pointer(updated.PurchasePrice),
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal plan grocery list")
	}

	logger.Info("meal plan grocery list updated")

	return nil
}

// ArchiveMealPlanGroceryListItem archives a meal plan grocery list from the database by its ID.
func (q *repository) ArchiveMealPlanGroceryListItem(ctx context.Context, mealPlanGroceryListItemID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanGroceryListItemID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanGroceryListItemIDKey, mealPlanGroceryListItemID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanGroceryListItemIDKey, mealPlanGroceryListItemID)

	rowsAffected, err := q.generatedQuerier.ArchiveMealPlanGroceryListItem(ctx, q.writeDB, mealPlanGroceryListItemID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving meal plan grocery list")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
