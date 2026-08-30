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
	_ mealplanning.RecipeStepProductDataManager = (*repository)(nil)
)

// RecipeStepProductExists fetches whether a recipe step product exists from the database.
func (q *repository) RecipeStepProductExists(ctx context.Context, recipeID, recipeStepID, recipeStepProductID string) (exists bool, err error) {
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

	if recipeStepProductID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepProductIDKey, recipeStepProductID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepProductIDKey, recipeStepProductID)

	result, err := q.generatedQuerier.CheckRecipeStepProductExistence(ctx, q.readDB, &generated.CheckRecipeStepProductExistenceParams{
		RecipeStepID:        recipeStepID,
		RecipeStepProductID: recipeStepProductID,
		RecipeID:            recipeID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing recipe step product existence check")
	}

	return result, nil
}

// GetRecipeStepProduct fetches a recipe step product from the database.
func (q *repository) GetRecipeStepProduct(ctx context.Context, recipeID, recipeStepID, recipeStepProductID string) (*mealplanning.RecipeStepProduct, error) {
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

	if recipeStepProductID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepProductIDKey, recipeStepProductID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepProductIDKey, recipeStepProductID)

	result, err := q.generatedQuerier.GetRecipeStepProduct(ctx, q.readDB, &generated.GetRecipeStepProductParams{
		RecipeStepID:        recipeStepID,
		RecipeStepProductID: recipeStepProductID,
		RecipeID:            recipeID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting recipe step product")
	}

	recipeStepProduct := &mealplanning.RecipeStepProduct{
		CreatedAt:                      result.CreatedAt,
		MinMeasurementQuantity:         database.Float32PointerFromNullString(result.MinimumMeasurementQuantityValue),
		MaxMeasurementQuantity:         database.Float32PointerFromNullString(result.MaximumMeasurementQuantityValue),
		MinItemQuantity:                database.Float32PointerFromNullString(result.MinimumItemQuantityValue),
		MaxItemQuantity:                database.Float32PointerFromNullString(result.MaximumItemQuantityValue),
		MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MinimumStorageTemperatureInCelsius),
		MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MaximumStorageTemperatureInCelsius),
		MaxStorageDurationInSeconds:    database.Uint32PointerFromNullInt32(result.MaximumStorageDurationInSeconds),
		ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
		MeasurementUnit:                nil,
		ContainedInVesselIndex:         database.Uint16PointerFromNullInt32(result.ContainedInVesselIndex),
		Name:                           result.Name,
		BelongsToRecipeStep:            result.BelongsToRecipeStep,
		Type:                           string(result.Type),
		ID:                             result.ID,
		StorageInstructions:            result.StorageInstructions,
		QuantityNotes:                  result.QuantityNotes,
		Index:                          uint16(result.Index),
		IsWaste:                        result.IsWaste,
		IsLiquid:                       result.IsLiquid,
		Compostable:                    result.Compostable,
	}

	if result.ValidMeasurementUnitID.Valid && result.ValidMeasurementUnitID.String != "" {
		recipeStepProduct.MeasurementUnit = &mealplanning.ValidMeasurementUnit{
			CreatedAt:     result.ValidMeasurementUnitCreatedAt.Time,
			LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitLastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitArchivedAt),
			Name:          result.ValidMeasurementUnitName.String,
			IconPath:      result.ValidMeasurementUnitIconPath.String,
			ID:            result.ValidMeasurementUnitID.String,
			Description:   result.ValidMeasurementUnitDescription.String,
			PluralName:    result.ValidMeasurementUnitPluralName.String,
			Slug:          result.ValidMeasurementUnitSlug.String,
			Volumetric:    database.BoolFromNullBool(result.ValidMeasurementUnitVolumetric),
			Universal:     result.ValidMeasurementUnitUniversal.Bool,
			Metric:        result.ValidMeasurementUnitMetric.Bool,
			Imperial:      result.ValidMeasurementUnitImperial.Bool,
		}
	}

	return recipeStepProduct, nil
}

// getRecipeStepProductsForRecipe fetches a list of recipe step products from the database that meet a particular filter.
func (q *repository) getRecipeStepProductsForRecipe(ctx context.Context, recipeID string) ([]*mealplanning.RecipeStepProduct, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	results, err := q.generatedQuerier.GetRecipeStepProductsForRecipe(ctx, q.readDB, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step products for recipe")
	}

	recipeStepProducts := []*mealplanning.RecipeStepProduct{}
	for _, result := range results {
		recipeStepProduct := &mealplanning.RecipeStepProduct{
			CreatedAt:                      result.CreatedAt,
			MinMeasurementQuantity:         database.Float32PointerFromNullString(result.MinimumMeasurementQuantityValue),
			MaxMeasurementQuantity:         database.Float32PointerFromNullString(result.MaximumMeasurementQuantityValue),
			MinItemQuantity:                database.Float32PointerFromNullString(result.MinimumItemQuantityValue),
			MaxItemQuantity:                database.Float32PointerFromNullString(result.MaximumItemQuantityValue),
			MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MinimumStorageTemperatureInCelsius),
			MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MaximumStorageTemperatureInCelsius),
			MaxStorageDurationInSeconds:    database.Uint32PointerFromNullInt32(result.MaximumStorageDurationInSeconds),
			ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
			MeasurementUnit:                nil,
			ContainedInVesselIndex:         database.Uint16PointerFromNullInt32(result.ContainedInVesselIndex),
			Name:                           result.Name,
			BelongsToRecipeStep:            result.BelongsToRecipeStep,
			Type:                           string(result.Type),
			ID:                             result.ID,
			StorageInstructions:            result.StorageInstructions,
			QuantityNotes:                  result.QuantityNotes,
			Index:                          uint16(result.Index),
			IsWaste:                        result.IsWaste,
			IsLiquid:                       result.IsLiquid,
			Compostable:                    result.Compostable,
		}

		if result.ValidMeasurementUnitID.Valid && result.ValidMeasurementUnitID.String != "" {
			recipeStepProduct.MeasurementUnit = &mealplanning.ValidMeasurementUnit{
				CreatedAt:     result.ValidMeasurementUnitCreatedAt.Time,
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitArchivedAt),
				Name:          result.ValidMeasurementUnitName.String,
				IconPath:      result.ValidMeasurementUnitIconPath.String,
				ID:            result.ValidMeasurementUnitID.String,
				Description:   result.ValidMeasurementUnitDescription.String,
				PluralName:    result.ValidMeasurementUnitPluralName.String,
				Slug:          result.ValidMeasurementUnitSlug.String,
				Volumetric:    database.BoolFromNullBool(result.ValidMeasurementUnitVolumetric),
				Universal:     result.ValidMeasurementUnitUniversal.Bool,
				Metric:        result.ValidMeasurementUnitMetric.Bool,
				Imperial:      result.ValidMeasurementUnitImperial.Bool,
			}
		}

		recipeStepProducts = append(recipeStepProducts, recipeStepProduct)
	}

	return recipeStepProducts, nil
}

// GetRecipeStepProducts fetches a list of recipe step products from the database that meet a particular filter.
func (q *repository) GetRecipeStepProducts(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.RecipeStepProduct], err error) {
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

	var (
		data          []*mealplanning.RecipeStepProduct
		filteredCount uint64
		totalCount    uint64
	)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetRecipeStepProducts(ctx, q.readDB, &generated.GetRecipeStepProductsParams{
		RecipeID:        recipeID,
		RecipeStepID:    recipeStepID,
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe step products list retrieval query")
	}

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		recipeStepProduct := &mealplanning.RecipeStepProduct{
			CreatedAt:                      result.CreatedAt,
			MinMeasurementQuantity:         database.Float32PointerFromNullString(result.MinimumMeasurementQuantityValue),
			MaxMeasurementQuantity:         database.Float32PointerFromNullString(result.MaximumMeasurementQuantityValue),
			MinItemQuantity:                database.Float32PointerFromNullString(result.MinimumItemQuantityValue),
			MaxItemQuantity:                database.Float32PointerFromNullString(result.MaximumItemQuantityValue),
			MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MinimumStorageTemperatureInCelsius),
			MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.MaximumStorageTemperatureInCelsius),
			MaxStorageDurationInSeconds:    database.Uint32PointerFromNullInt32(result.MaximumStorageDurationInSeconds),
			ArchivedAt:                     database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.LastUpdatedAt),
			MeasurementUnit:                nil,
			ContainedInVesselIndex:         database.Uint16PointerFromNullInt32(result.ContainedInVesselIndex),
			Name:                           result.Name,
			BelongsToRecipeStep:            result.BelongsToRecipeStep,
			Type:                           string(result.Type),
			ID:                             result.ID,
			StorageInstructions:            result.StorageInstructions,
			QuantityNotes:                  result.QuantityNotes,
			Index:                          uint16(result.Index),
			IsWaste:                        result.IsWaste,
			IsLiquid:                       result.IsLiquid,
			Compostable:                    result.Compostable,
		}

		if result.ValidMeasurementUnitID.Valid && result.ValidMeasurementUnitID.String != "" {
			recipeStepProduct.MeasurementUnit = &mealplanning.ValidMeasurementUnit{
				CreatedAt:     result.ValidMeasurementUnitCreatedAt.Time,
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitArchivedAt),
				Name:          result.ValidMeasurementUnitName.String,
				IconPath:      result.ValidMeasurementUnitIconPath.String,
				ID:            result.ValidMeasurementUnitID.String,
				Description:   result.ValidMeasurementUnitDescription.String,
				PluralName:    result.ValidMeasurementUnitPluralName.String,
				Slug:          result.ValidMeasurementUnitSlug.String,
				Volumetric:    database.BoolFromNullBool(result.ValidMeasurementUnitVolumetric),
				Universal:     result.ValidMeasurementUnitUniversal.Bool,
				Metric:        result.ValidMeasurementUnitMetric.Bool,
				Imperial:      result.ValidMeasurementUnitImperial.Bool,
			}
		}

		data = append(data, recipeStepProduct)
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(rsp *mealplanning.RecipeStepProduct) string { return rsp.ID },
		filter,
	)

	return x, nil
}

// CreateRecipeStepProduct creates a recipe step product in the database.
func (q *repository) createRecipeStepProduct(ctx context.Context, db database.SQLQueryExecutor, input *mealplanning.RecipeStepProductDatabaseCreationInput) (*mealplanning.RecipeStepProduct, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	// create the recipe step product.
	if err := q.generatedQuerier.CreateRecipeStepProduct(ctx, db, &generated.CreateRecipeStepProductParams{
		QuantityNotes:                      input.QuantityNotes,
		Name:                               input.Name,
		Type:                               generated.RecipeStepProductType(input.Type),
		BelongsToRecipeStep:                input.BelongsToRecipeStep,
		ID:                                 input.ID,
		StorageInstructions:                input.StorageInstructions,
		MinimumMeasurementQuantityValue:    database.NullStringFromFloat32Pointer(input.MinMeasurementQuantity),
		MaximumMeasurementQuantityValue:    database.NullStringFromFloat32Pointer(input.MaxMeasurementQuantity),
		MinimumItemQuantityValue:           database.NullStringFromFloat32Pointer(input.MinItemQuantity),
		MaximumItemQuantityValue:           database.NullStringFromFloat32Pointer(input.MaxItemQuantity),
		MinimumStorageTemperatureInCelsius: database.NullStringFromFloat32Pointer(input.MinStorageTemperatureInCelsius),
		MaximumStorageTemperatureInCelsius: database.NullStringFromFloat32Pointer(input.MaxStorageTemperatureInCelsius),
		MeasurementUnit:                    database.NullStringFromStringPointer(input.MeasurementUnitID),
		MaximumStorageDurationInSeconds:    database.NullInt32FromUint32Pointer(input.MaxStorageDurationInSeconds),
		ContainedInVesselIndex:             database.NullInt32FromUint16Pointer(input.ContainedInVesselIndex),
		Index:                              int32(input.Index),
		Compostable:                        input.Compostable,
		IsLiquid:                           input.IsLiquid,
		IsWaste:                            input.IsWaste,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "performing recipe step product creation query")
	}

	x := &mealplanning.RecipeStepProduct{
		ID:                             input.ID,
		Name:                           input.Name,
		Type:                           input.Type,
		QuantityNotes:                  input.QuantityNotes,
		BelongsToRecipeStep:            input.BelongsToRecipeStep,
		Compostable:                    input.Compostable,
		MinMeasurementQuantity:         input.MinMeasurementQuantity,
		MaxMeasurementQuantity:         input.MaxMeasurementQuantity,
		MinItemQuantity:                input.MinItemQuantity,
		MaxItemQuantity:                input.MaxItemQuantity,
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		MinStorageDurationInSeconds:    input.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    input.MaxStorageDurationInSeconds,
		StorageInstructions:            input.StorageInstructions,
		IsLiquid:                       input.IsLiquid,
		IsWaste:                        input.IsWaste,
		Index:                          input.Index,
		ContainedInVesselIndex:         input.ContainedInVesselIndex,
		CreatedAt:                      q.CurrentTime(),
	}

	if input.MeasurementUnitID != nil {
		x.MeasurementUnit = &mealplanning.ValidMeasurementUnit{ID: *input.MeasurementUnitID}
	}

	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepProductIDKey, x.ID)

	return x, nil
}

// CreateRecipeStepProduct creates a recipe step product in the database.
func (q *repository) CreateRecipeStepProduct(ctx context.Context, recipeID string, input *mealplanning.RecipeStepProductDatabaseCreationInput) (*mealplanning.RecipeStepProduct, error) {
	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	var created *mealplanning.RecipeStepProduct

	// The write and its event share a transaction.
	if err := q.withEvent(ctx, q.logger, mealplanning.RecipeStepProductCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:            recipeID,
		mealplanningkeys.RecipeStepIDKey:        input.BelongsToRecipeStep,
		mealplanningkeys.RecipeStepProductIDKey: input.ID,
	}, func(tx database.Tx) error {
		var createErr error
		created, createErr = q.createRecipeStepProduct(ctx, tx, input)

		return createErr
	}); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateRecipeStepProduct updates a particular recipe step product.
func (q *repository) UpdateRecipeStepProduct(ctx context.Context, recipeID string, updated *mealplanning.RecipeStepProduct) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeStepProductIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepProductIDKey, updated.ID)

	var measurementUnitID *string
	if updated.MeasurementUnit != nil {
		measurementUnitID = &updated.MeasurementUnit.ID
	}

	if err := q.withEvent(ctx, logger, mealplanning.RecipeStepProductUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:            recipeID,
		mealplanningkeys.RecipeStepIDKey:        updated.BelongsToRecipeStep,
		mealplanningkeys.RecipeStepProductIDKey: updated.ID,
	}, func(tx database.Tx) error {
		_, updateErr := q.generatedQuerier.UpdateRecipeStepProduct(ctx, tx, &generated.UpdateRecipeStepProductParams{
			Name:                               updated.Name,
			Type:                               generated.RecipeStepProductType(updated.Type),
			MeasurementUnit:                    database.NullStringFromStringPointer(measurementUnitID),
			MinimumMeasurementQuantityValue:    database.NullStringFromFloat32Pointer(updated.MinMeasurementQuantity),
			MaximumMeasurementQuantityValue:    database.NullStringFromFloat32Pointer(updated.MaxMeasurementQuantity),
			MinimumItemQuantityValue:           database.NullStringFromFloat32Pointer(updated.MinItemQuantity),
			MaximumItemQuantityValue:           database.NullStringFromFloat32Pointer(updated.MaxItemQuantity),
			QuantityNotes:                      updated.QuantityNotes,
			Compostable:                        updated.Compostable,
			MaximumStorageDurationInSeconds:    database.NullInt32FromUint32Pointer(updated.MaxStorageDurationInSeconds),
			MinimumStorageTemperatureInCelsius: database.NullStringFromFloat32Pointer(updated.MinStorageTemperatureInCelsius),
			MaximumStorageTemperatureInCelsius: database.NullStringFromFloat32Pointer(updated.MaxStorageTemperatureInCelsius),
			StorageInstructions:                updated.StorageInstructions,
			IsLiquid:                           updated.IsLiquid,
			IsWaste:                            updated.IsWaste,
			Index:                              int32(updated.Index),
			ContainedInVesselIndex:             database.NullInt32FromUint16Pointer(updated.ContainedInVesselIndex),
			BelongsToRecipeStep:                updated.BelongsToRecipeStep,
			ID:                                 updated.ID,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe step product")
	}

	logger.Info("recipe step product updated")

	return nil
}

// ArchiveRecipeStepProduct archives a recipe step product from the database by its ID.
func (q *repository) ArchiveRecipeStepProduct(ctx context.Context, recipeID, recipeStepID, recipeStepProductID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeStepID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if recipeStepProductID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepProductIDKey, recipeStepProductID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepProductIDKey, recipeStepProductID)

	if err := q.withEvent(ctx, logger, mealplanning.RecipeStepProductArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:            recipeID,
		mealplanningkeys.RecipeStepIDKey:        recipeStepID,
		mealplanningkeys.RecipeStepProductIDKey: recipeStepProductID,
	}, func(tx database.Tx) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveRecipeStepProduct(ctx, tx, &generated.ArchiveRecipeStepProductParams{
			BelongsToRecipeStep: recipeStepID,
			ID:                  recipeStepProductID,
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "updating recipe step product")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
