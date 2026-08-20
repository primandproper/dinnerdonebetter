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
	_ mealplanning.RecipeStepVesselDataManager = (*repository)(nil)
)

// RecipeStepVesselExists fetches whether a recipe step vessel exists from the database.
func (q *repository) RecipeStepVesselExists(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string) (exists bool, err error) {
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

	if recipeStepVesselID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepVesselIDKey, recipeStepVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepVesselIDKey, recipeStepVesselID)

	result, err := q.generatedQuerier.CheckRecipeStepVesselExistence(ctx, q.readDB, &generated.CheckRecipeStepVesselExistenceParams{
		RecipeStepID:       recipeStepID,
		RecipeStepVesselID: recipeStepVesselID,
		RecipeID:           recipeID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing recipe step vessel existence check")
	}

	return result, nil
}

// GetRecipeStepVessel fetches a recipe step vessel from the database.
func (q *repository) GetRecipeStepVessel(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string) (*mealplanning.RecipeStepVessel, error) {
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

	if recipeStepVesselID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepVesselIDKey, recipeStepVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepVesselIDKey, recipeStepVesselID)

	result, err := q.generatedQuerier.GetRecipeStepVessel(ctx, q.readDB, &generated.GetRecipeStepVesselParams{
		RecipeStepID:       recipeStepID,
		RecipeStepVesselID: recipeStepVesselID,
		RecipeID:           recipeID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting recipe step vessel")
	}

	scaleFactor := database.Float32FromString(result.ScaleFactor)
	if scaleFactor <= 0 {
		scaleFactor = 1.0
	}
	recipeStepVessel := &mealplanning.RecipeStepVessel{
		CreatedAt:            result.CreatedAt,
		MinQuantity:          uint16(result.MinimumQuantity),
		MaxQuantity:          database.Uint16PointerFromNullInt32(result.MaximumQuantity),
		LastUpdatedAt:        database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:           database.TimePointerFromNullTime(result.ArchivedAt),
		RecipeStepProductID:  database.StringPointerFromNullString(result.RecipeStepProductID),
		Vessel:               nil,
		ID:                   result.ID,
		Notes:                result.Notes,
		BelongsToRecipeStep:  result.BelongsToRecipeStep,
		VesselPreposition:    result.VesselPredicate,
		Name:                 result.Name,
		Index:                uint16(result.Index),
		OptionIndex:          uint16(result.OptionIndex),
		UnavailableAfterStep: result.UnavailableAfterStep,
		ScaleFactor:          scaleFactor,
	}

	if result.ValidVesselID.Valid {
		recipeStepVessel.Vessel = &mealplanning.ValidVessel{
			CreatedAt:                      result.ValidVesselCreatedAt.Time,
			ArchivedAt:                     database.TimePointerFromNullTime(result.ValidVesselArchivedAt),
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidVesselLastUpdatedAt),
			CapacityUnit:                   nil,
			IconPath:                       result.ValidVesselIconPath.String,
			PluralName:                     result.ValidVesselPluralName.String,
			Description:                    result.ValidVesselDescription.String,
			Name:                           result.ValidVesselName.String,
			Slug:                           result.ValidVesselSlug.String,
			Shape:                          string(result.ValidVesselShape.VesselShape),
			ID:                             result.ValidVesselID.String,
			WidthInMillimeters:             database.Float32FromNullString(result.ValidVesselWidthInMillimeters),
			LengthInMillimeters:            database.Float32FromNullString(result.ValidVesselLengthInMillimeters),
			HeightInMillimeters:            database.Float32FromNullString(result.ValidVesselHeightInMillimeters),
			Capacity:                       database.Float32FromNullString(result.ValidVesselCapacity),
			IncludeInGeneratedInstructions: result.ValidVesselIncludeInGeneratedInstructions.Bool,
			DisplayInSummaryLists:          result.ValidVesselDisplayInSummaryLists.Bool,
			UsableForStorage:               result.ValidVesselUsableForStorage.Bool,
		}

		if result.ValidMeasurementUnitID.Valid {
			recipeStepVessel.Vessel.CapacityUnit = &mealplanning.ValidMeasurementUnit{
				CreatedAt:     result.ValidMeasurementUnitCreatedAt.Time,
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitLastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitArchivedAt),
				Name:          result.ValidMeasurementUnitName.String,
				IconPath:      result.ValidMeasurementUnitIconPath.String,
				ID:            result.ValidMeasurementUnitID.String,
				Description:   result.ValidMeasurementUnitDescription.String,
				PluralName:    result.ValidMeasurementUnitPluralName.String,
				Slug:          result.ValidMeasurementUnitSlug.String,
				Volumetric:    result.ValidMeasurementUnitVolumetric.Bool,
				Universal:     result.ValidMeasurementUnitUniversal.Bool,
				Metric:        result.ValidMeasurementUnitMetric.Bool,
				Imperial:      result.ValidMeasurementUnitImperial.Bool,
			}
		}
	}

	return recipeStepVessel, nil
}

// GetRecipeStepVessels fetches a list of recipe step vessels from the database that meet a particular filter.
func (q *repository) GetRecipeStepVessels(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.RecipeStepVessel], err error) {
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

	results, err := q.generatedQuerier.GetRecipeStepVessels(ctx, q.readDB, &generated.GetRecipeStepVesselsParams{
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
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe step vessels list retrieval query")
	}

	var (
		data          []*mealplanning.RecipeStepVessel
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		scaleFactor := database.Float32FromString(result.ScaleFactor)
		if scaleFactor <= 0 {
			scaleFactor = 1.0
		}
		recipeStepVessel := &mealplanning.RecipeStepVessel{
			CreatedAt:            result.CreatedAt,
			MinQuantity:          uint16(result.MinimumQuantity),
			MaxQuantity:          database.Uint16PointerFromNullInt32(result.MaximumQuantity),
			LastUpdatedAt:        database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:           database.TimePointerFromNullTime(result.ArchivedAt),
			RecipeStepProductID:  database.StringPointerFromNullString(result.RecipeStepProductID),
			Vessel:               nil,
			ID:                   result.ID,
			Notes:                result.Notes,
			BelongsToRecipeStep:  result.BelongsToRecipeStep,
			VesselPreposition:    result.VesselPredicate,
			Name:                 result.Name,
			Index:                uint16(result.Index),
			OptionIndex:          uint16(result.OptionIndex),
			UnavailableAfterStep: result.UnavailableAfterStep,
			ScaleFactor:          scaleFactor,
		}

		if result.ValidVesselID.Valid {
			recipeStepVessel.Vessel = &mealplanning.ValidVessel{
				CreatedAt:                      result.ValidVesselCreatedAt.Time,
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidVesselArchivedAt),
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidVesselLastUpdatedAt),
				CapacityUnit:                   nil,
				IconPath:                       result.ValidVesselIconPath.String,
				PluralName:                     result.ValidVesselPluralName.String,
				Description:                    result.ValidVesselDescription.String,
				Name:                           result.ValidVesselName.String,
				Slug:                           result.ValidVesselSlug.String,
				Shape:                          string(result.ValidVesselShape.VesselShape),
				ID:                             result.ValidVesselID.String,
				WidthInMillimeters:             database.Float32FromNullString(result.ValidVesselWidthInMillimeters),
				LengthInMillimeters:            database.Float32FromNullString(result.ValidVesselLengthInMillimeters),
				HeightInMillimeters:            database.Float32FromNullString(result.ValidVesselHeightInMillimeters),
				Capacity:                       database.Float32FromNullString(result.ValidVesselCapacity),
				IncludeInGeneratedInstructions: result.ValidVesselIncludeInGeneratedInstructions.Bool,
				DisplayInSummaryLists:          result.ValidVesselDisplayInSummaryLists.Bool,
				UsableForStorage:               result.ValidVesselUsableForStorage.Bool,
			}

			if result.ValidMeasurementUnitID.Valid {
				recipeStepVessel.Vessel.CapacityUnit = &mealplanning.ValidMeasurementUnit{
					CreatedAt:     result.ValidMeasurementUnitCreatedAt.Time,
					LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitLastUpdatedAt),
					ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitArchivedAt),
					Name:          result.ValidMeasurementUnitName.String,
					IconPath:      result.ValidMeasurementUnitIconPath.String,
					ID:            result.ValidMeasurementUnitID.String,
					Description:   result.ValidMeasurementUnitDescription.String,
					PluralName:    result.ValidMeasurementUnitPluralName.String,
					Slug:          result.ValidMeasurementUnitSlug.String,
					Volumetric:    result.ValidMeasurementUnitVolumetric.Bool,
					Universal:     result.ValidMeasurementUnitUniversal.Bool,
					Metric:        result.ValidMeasurementUnitMetric.Bool,
					Imperial:      result.ValidMeasurementUnitImperial.Bool,
				}
			}
		}

		data = append(data, recipeStepVessel)
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(rsv *mealplanning.RecipeStepVessel) string { return rsv.ID },
		filter,
	)

	return x, nil
}

// getRecipeStepVesselsForRecipe fetches a list of recipe step vessels from the database that meet a particular filter.
func (q *repository) getRecipeStepVesselsForRecipe(ctx context.Context, recipeID string) ([]*mealplanning.RecipeStepVessel, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	results, err := q.generatedQuerier.GetRecipeStepVesselsForRecipe(ctx, q.readDB, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step vessels for recipe")
	}

	recipeStepVessels := []*mealplanning.RecipeStepVessel{}
	for _, result := range results {
		scaleFactor := database.Float32FromString(result.ScaleFactor)
		if scaleFactor <= 0 {
			scaleFactor = 1.0
		}
		recipeStepVessel := &mealplanning.RecipeStepVessel{
			CreatedAt:            result.CreatedAt,
			MinQuantity:          uint16(result.MinimumQuantity),
			MaxQuantity:          database.Uint16PointerFromNullInt32(result.MaximumQuantity),
			LastUpdatedAt:        database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:           database.TimePointerFromNullTime(result.ArchivedAt),
			RecipeStepProductID:  database.StringPointerFromNullString(result.RecipeStepProductID),
			Vessel:               nil,
			ID:                   result.ID,
			Notes:                result.Notes,
			BelongsToRecipeStep:  result.BelongsToRecipeStep,
			VesselPreposition:    result.VesselPredicate,
			Name:                 result.Name,
			Index:                uint16(result.Index),
			OptionIndex:          uint16(result.OptionIndex),
			UnavailableAfterStep: result.UnavailableAfterStep,
			ScaleFactor:          scaleFactor,
		}

		if result.ValidVesselID.Valid {
			recipeStepVessel.Vessel = &mealplanning.ValidVessel{
				CreatedAt:                      result.ValidVesselCreatedAt.Time,
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidVesselArchivedAt),
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidVesselLastUpdatedAt),
				CapacityUnit:                   nil,
				IconPath:                       result.ValidVesselIconPath.String,
				PluralName:                     result.ValidVesselPluralName.String,
				Description:                    result.ValidVesselDescription.String,
				Name:                           result.ValidVesselName.String,
				Slug:                           result.ValidVesselSlug.String,
				Shape:                          string(result.ValidVesselShape.VesselShape),
				ID:                             result.ValidVesselID.String,
				WidthInMillimeters:             database.Float32FromNullString(result.ValidVesselWidthInMillimeters),
				LengthInMillimeters:            database.Float32FromNullString(result.ValidVesselLengthInMillimeters),
				HeightInMillimeters:            database.Float32FromNullString(result.ValidVesselHeightInMillimeters),
				Capacity:                       database.Float32FromNullString(result.ValidVesselCapacity),
				IncludeInGeneratedInstructions: result.ValidVesselIncludeInGeneratedInstructions.Bool,
				DisplayInSummaryLists:          result.ValidVesselDisplayInSummaryLists.Bool,
				UsableForStorage:               result.ValidVesselUsableForStorage.Bool,
			}

			if result.ValidMeasurementUnitID.Valid {
				recipeStepVessel.Vessel.CapacityUnit = &mealplanning.ValidMeasurementUnit{
					CreatedAt:     result.ValidMeasurementUnitCreatedAt.Time,
					LastUpdatedAt: database.TimePointerFromNullTime(result.ValidMeasurementUnitLastUpdatedAt),
					ArchivedAt:    database.TimePointerFromNullTime(result.ValidMeasurementUnitArchivedAt),
					Name:          result.ValidMeasurementUnitName.String,
					IconPath:      result.ValidMeasurementUnitIconPath.String,
					ID:            result.ValidMeasurementUnitID.String,
					Description:   result.ValidMeasurementUnitDescription.String,
					PluralName:    result.ValidMeasurementUnitPluralName.String,
					Slug:          result.ValidMeasurementUnitSlug.String,
					Volumetric:    result.ValidMeasurementUnitVolumetric.Bool,
					Universal:     result.ValidMeasurementUnitUniversal.Bool,
					Metric:        result.ValidMeasurementUnitMetric.Bool,
					Imperial:      result.ValidMeasurementUnitImperial.Bool,
				}
			}
		}

		recipeStepVessels = append(recipeStepVessels, recipeStepVessel)
	}

	return recipeStepVessels, nil
}

// CreateRecipeStepVessel creates a recipe step vessel in the database.
func (q *repository) createRecipeStepVessel(ctx context.Context, querier database.SQLQueryExecutor, input *mealplanning.RecipeStepVesselDatabaseCreationInput) (*mealplanning.RecipeStepVessel, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := q.logger.WithValue(mealplanningkeys.RecipeStepVesselIDKey, input.ID).WithValue(mealplanningkeys.RecipeStepIDKey, input.BelongsToRecipeStep)

	// create the recipe step vessel.
	if err := q.generatedQuerier.CreateRecipeStepVessel(ctx, querier, &generated.CreateRecipeStepVesselParams{
		ID:                   input.ID,
		Name:                 input.Name,
		Notes:                input.Notes,
		BelongsToRecipeStep:  input.BelongsToRecipeStep,
		VesselPredicate:      input.VesselPreposition,
		RecipeStepProductID:  database.NullStringFromStringPointer(input.RecipeStepProductID),
		ValidVesselID:        database.NullStringFromStringPointer(input.VesselID),
		MaximumQuantity:      database.NullInt32FromUint16Pointer(input.MaxQuantity),
		MinimumQuantity:      int32(input.MinQuantity),
		Index:                int32(input.Index),
		OptionIndex:          int32(input.OptionIndex),
		UnavailableAfterStep: input.UnavailableAfterStep,
		ScaleFactor:          database.StringFromFloat32(input.ScaleFactor),
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing recipe step vessel creation query")
	}

	x := &mealplanning.RecipeStepVessel{
		ID:                   input.ID,
		RecipeStepProductID:  input.RecipeStepProductID,
		Name:                 input.Name,
		Notes:                input.Notes,
		BelongsToRecipeStep:  input.BelongsToRecipeStep,
		MinQuantity:          input.MinQuantity,
		MaxQuantity:          input.MaxQuantity,
		VesselPreposition:    input.VesselPreposition,
		Index:                input.Index,
		OptionIndex:          input.OptionIndex,
		UnavailableAfterStep: input.UnavailableAfterStep,
		ScaleFactor:          input.ScaleFactor,
		CreatedAt:            q.CurrentTime(),
	}

	if input.VesselID != nil {
		x.Vessel = &mealplanning.ValidVessel{ID: *input.VesselID}
	}

	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepVesselIDKey, x.ID)
	logger.Info("recipe step vessel created")

	return x, nil
}

// CreateRecipeStepVessel creates a recipe step vessel in the database.
func (q *repository) CreateRecipeStepVessel(ctx context.Context, recipeID string, input *mealplanning.RecipeStepVesselDatabaseCreationInput) (*mealplanning.RecipeStepVessel, error) {
	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	var created *mealplanning.RecipeStepVessel

	// The write and its event share a transaction.
	if err := q.withEvent(ctx, q.logger, mealplanning.RecipeStepVesselCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:           recipeID,
		mealplanningkeys.RecipeStepIDKey:       input.BelongsToRecipeStep,
		mealplanningkeys.RecipeStepVesselIDKey: input.ID,
	}, func(tx database.SQLQueryExecutor) error {
		var createErr error
		created, createErr = q.createRecipeStepVessel(ctx, tx, input)

		return createErr
	}); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateRecipeStepVessel updates a particular recipe step vessel.
func (q *repository) UpdateRecipeStepVessel(ctx context.Context, recipeID string, updated *mealplanning.RecipeStepVessel) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeStepVesselIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepVesselIDKey, updated.ID)

	var vesselID *string
	if updated.Vessel != nil {
		vesselID = &updated.Vessel.ID
	}

	if err := q.withEvent(ctx, logger, mealplanning.RecipeStepVesselUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:           recipeID,
		mealplanningkeys.RecipeStepIDKey:       updated.BelongsToRecipeStep,
		mealplanningkeys.RecipeStepVesselIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		_, updateErr := q.generatedQuerier.UpdateRecipeStepVessel(ctx, tx, &generated.UpdateRecipeStepVesselParams{
			Name:                 updated.Name,
			Notes:                updated.Notes,
			BelongsToRecipeStep:  updated.BelongsToRecipeStep,
			VesselPredicate:      updated.VesselPreposition,
			ID:                   updated.ID,
			RecipeStepProductID:  database.NullStringFromStringPointer(updated.RecipeStepProductID),
			ValidVesselID:        database.NullStringFromStringPointer(vesselID),
			MaximumQuantity:      database.NullInt32FromUint16Pointer(updated.MaxQuantity),
			MinimumQuantity:      int32(updated.MinQuantity),
			Index:                int32(updated.Index),
			OptionIndex:          int32(updated.OptionIndex),
			UnavailableAfterStep: updated.UnavailableAfterStep,
			ScaleFactor:          database.StringFromFloat32(updated.ScaleFactor),
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe step vessel")
	}

	logger.Info("recipe step vessel updated")

	return nil
}

// ArchiveRecipeStepVessel archives a recipe step vessel from the database by its ID.
func (q *repository) ArchiveRecipeStepVessel(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeStepID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if recipeStepVesselID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepVesselIDKey, recipeStepVesselID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepVesselIDKey, recipeStepVesselID)

	if err := q.withEvent(ctx, logger, mealplanning.RecipeStepVesselArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:           recipeID,
		mealplanningkeys.RecipeStepIDKey:       recipeStepID,
		mealplanningkeys.RecipeStepVesselIDKey: recipeStepVesselID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveRecipeStepVessel(ctx, tx, &generated.ArchiveRecipeStepVesselParams{
			BelongsToRecipeStep: recipeStepID,
			ID:                  recipeStepVesselID,
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving recipe step vessel")
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
