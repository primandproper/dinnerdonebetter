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
	_ mealplanning.RecipeStepInstrumentDataManager = (*repository)(nil)
)

// RecipeStepInstrumentExists fetches whether a recipe step instrument exists from the database.
func (q *repository) RecipeStepInstrumentExists(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string) (exists bool, err error) {
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

	if recipeStepInstrumentID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepInstrumentIDKey, recipeStepInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepInstrumentIDKey, recipeStepInstrumentID)

	result, err := q.generatedQuerier.CheckRecipeStepInstrumentExistence(ctx, q.readDB, &generated.CheckRecipeStepInstrumentExistenceParams{
		RecipeStepID:           recipeStepID,
		RecipeStepInstrumentID: recipeStepInstrumentID,
		RecipeID:               recipeID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing recipe step instrument existence check")
	}

	return result, nil
}

// GetRecipeStepInstrument fetches a recipe step instrument from the database.
func (q *repository) GetRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string) (*mealplanning.RecipeStepInstrument, error) {
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

	if recipeStepInstrumentID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepInstrumentIDKey, recipeStepInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepInstrumentIDKey, recipeStepInstrumentID)

	result, err := q.generatedQuerier.GetRecipeStepInstrument(ctx, q.readDB, &generated.GetRecipeStepInstrumentParams{
		RecipeStepID:           recipeStepID,
		RecipeStepInstrumentID: recipeStepInstrumentID,
		RecipeID:               recipeID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing recipe step instrument get")
	}

	scaleFactor := database.Float32FromString(result.ScaleFactor)
	if scaleFactor <= 0 {
		scaleFactor = 1.0
	}
	recipeStepInstrument := &mealplanning.RecipeStepInstrument{
		CreatedAt:           result.CreatedAt,
		Instrument:          nil,
		LastUpdatedAt:       database.TimePointerFromNullTime(result.LastUpdatedAt),
		RecipeStepProductID: database.StringPointerFromNullString(result.RecipeStepProductID),
		ArchivedAt:          database.TimePointerFromNullTime(result.ArchivedAt),
		MinQuantity:         uint32(result.MinimumQuantity),
		MaxQuantity:         database.Uint32PointerFromNullInt32(result.MaximumQuantity),
		Notes:               result.Notes,
		Name:                result.Name,
		BelongsToRecipeStep: result.BelongsToRecipeStep,
		ID:                  result.ID,
		Index:               uint16(result.Index),
		OptionIndex:         uint16(result.OptionIndex),
		PreferenceRank:      uint8(result.PreferenceRank),
		Optional:            result.Optional,
		ScaleFactor:         scaleFactor,
	}

	if result.ValidInstrumentID.Valid {
		recipeStepInstrument.Instrument = &mealplanning.ValidInstrument{
			CreatedAt:                      result.ValidInstrumentCreatedAt.Time,
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidInstrumentLastUpdatedAt),
			ArchivedAt:                     database.TimePointerFromNullTime(result.ValidInstrumentArchivedAt),
			IconPath:                       result.ValidInstrumentIconPath.String,
			ID:                             result.ValidInstrumentID.String,
			Name:                           result.ValidInstrumentName.String,
			PluralName:                     result.ValidInstrumentPluralName.String,
			Description:                    result.ValidInstrumentDescription.String,
			Slug:                           result.ValidInstrumentSlug.String,
			DisplayInSummaryLists:          result.ValidInstrumentDisplayInSummaryLists.Bool,
			IncludeInGeneratedInstructions: result.ValidInstrumentIncludeInGeneratedInstructions.Bool,
			UsableForStorage:               result.ValidInstrumentUsableForStorage.Bool,
		}
	}

	return recipeStepInstrument, nil
}

// GetRecipeStepInstruments fetches a list of recipe step instruments from the database that meet a particular filter.
func (q *repository) GetRecipeStepInstruments(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.RecipeStepInstrument], err error) {
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
		data          []*mealplanning.RecipeStepInstrument
		filteredCount uint64
		totalCount    uint64
	)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetRecipeStepInstruments(ctx, q.readDB, &generated.GetRecipeStepInstrumentsParams{
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
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe step instruments list retrieval query")
	}

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		scaleFactor := database.Float32FromString(result.ScaleFactor)
		if scaleFactor <= 0 {
			scaleFactor = 1.0
		}
		recipeStepInstrument := &mealplanning.RecipeStepInstrument{
			CreatedAt:           result.CreatedAt,
			Instrument:          nil,
			LastUpdatedAt:       database.TimePointerFromNullTime(result.LastUpdatedAt),
			RecipeStepProductID: database.StringPointerFromNullString(result.RecipeStepProductID),
			ArchivedAt:          database.TimePointerFromNullTime(result.ArchivedAt),
			Notes:               result.Notes,
			Name:                result.Name,
			BelongsToRecipeStep: result.BelongsToRecipeStep,
			ID:                  result.ID,
			MinQuantity:         uint32(result.MinimumQuantity),
			MaxQuantity:         database.Uint32PointerFromNullInt32(result.MaximumQuantity),
			Index:               uint16(result.Index),
			OptionIndex:         uint16(result.OptionIndex),
			PreferenceRank:      uint8(result.PreferenceRank),
			Optional:            result.Optional,
			ScaleFactor:         scaleFactor,
		}

		if result.ValidInstrumentID.Valid {
			recipeStepInstrument.Instrument = &mealplanning.ValidInstrument{
				CreatedAt:                      result.ValidInstrumentCreatedAt.Time,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidInstrumentLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidInstrumentArchivedAt),
				IconPath:                       result.ValidInstrumentIconPath.String,
				ID:                             result.ValidInstrumentID.String,
				Name:                           result.ValidInstrumentName.String,
				PluralName:                     result.ValidInstrumentPluralName.String,
				Description:                    result.ValidInstrumentDescription.String,
				Slug:                           result.ValidInstrumentSlug.String,
				DisplayInSummaryLists:          result.ValidInstrumentDisplayInSummaryLists.Bool,
				IncludeInGeneratedInstructions: result.ValidInstrumentIncludeInGeneratedInstructions.Bool,
				UsableForStorage:               result.ValidInstrumentUsableForStorage.Bool,
			}
		}

		data = append(data, recipeStepInstrument)
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(rsi *mealplanning.RecipeStepInstrument) string { return rsi.ID },
		filter,
	)

	return x, nil
}

// getRecipeStepInstrumentsForRecipe fetches a list of recipe step instruments from the database that meet a particular filter.
func (q *repository) getRecipeStepInstrumentsForRecipe(ctx context.Context, recipeID string) ([]*mealplanning.RecipeStepInstrument, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	results, err := q.generatedQuerier.GetRecipeStepInstrumentsForRecipe(ctx, q.readDB, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing recipe step instruments list retrieval")
	}

	recipeStepInstruments := []*mealplanning.RecipeStepInstrument{}
	for _, result := range results {
		scaleFactor := database.Float32FromString(result.ScaleFactor)
		if scaleFactor <= 0 {
			scaleFactor = 1.0
		}
		recipeStepInstrument := &mealplanning.RecipeStepInstrument{
			CreatedAt:           result.CreatedAt,
			Instrument:          nil,
			LastUpdatedAt:       database.TimePointerFromNullTime(result.LastUpdatedAt),
			RecipeStepProductID: database.StringPointerFromNullString(result.RecipeStepProductID),
			ArchivedAt:          database.TimePointerFromNullTime(result.ArchivedAt),
			Notes:               result.Notes,
			Name:                result.Name,
			BelongsToRecipeStep: result.BelongsToRecipeStep,
			ID:                  result.ID,
			MinQuantity:         uint32(result.MinimumQuantity),
			MaxQuantity:         database.Uint32PointerFromNullInt32(result.MaximumQuantity),
			Index:               uint16(result.Index),
			OptionIndex:         uint16(result.OptionIndex),
			PreferenceRank:      uint8(result.PreferenceRank),
			Optional:            result.Optional,
			ScaleFactor:         scaleFactor,
		}

		if result.ValidInstrumentID.Valid {
			recipeStepInstrument.Instrument = &mealplanning.ValidInstrument{
				CreatedAt:                      result.ValidInstrumentCreatedAt.Time,
				LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidInstrumentLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(result.ValidInstrumentArchivedAt),
				IconPath:                       result.ValidInstrumentIconPath.String,
				ID:                             result.ValidInstrumentID.String,
				Name:                           result.ValidInstrumentName.String,
				PluralName:                     result.ValidInstrumentPluralName.String,
				Description:                    result.ValidInstrumentDescription.String,
				Slug:                           result.ValidInstrumentSlug.String,
				DisplayInSummaryLists:          result.ValidInstrumentDisplayInSummaryLists.Bool,
				IncludeInGeneratedInstructions: result.ValidInstrumentIncludeInGeneratedInstructions.Bool,
				UsableForStorage:               result.ValidInstrumentUsableForStorage.Bool,
			}
		}

		recipeStepInstruments = append(recipeStepInstruments, recipeStepInstrument)
	}

	return recipeStepInstruments, nil
}

// CreateRecipeStepInstrument creates a recipe step instrument in the database.
func (q *repository) createRecipeStepInstrument(ctx context.Context, querier database.Tx, input *mealplanning.RecipeStepInstrumentDatabaseCreationInput) (*mealplanning.RecipeStepInstrument, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepInstrumentIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.RecipeStepIDKey, input.BelongsToRecipeStep).WithValue(mealplanningkeys.RecipeStepInstrumentIDKey, input.ID)

	// create the recipe step instrument.
	if err := q.generatedQuerier.CreateRecipeStepInstrument(ctx, querier, &generated.CreateRecipeStepInstrumentParams{
		ID:                  input.ID,
		Name:                input.Name,
		Notes:               input.Notes,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		InstrumentID:        database.NullStringFromStringPointer(input.InstrumentID),
		RecipeStepProductID: database.NullStringFromStringPointer(input.RecipeStepProductID),
		MaximumQuantity:     database.NullInt32FromUint32Pointer(input.MaxQuantity),
		PreferenceRank:      int32(input.PreferenceRank),
		Index:               int32(input.Index),
		OptionIndex:         int32(input.OptionIndex),
		MinimumQuantity:     int32(input.MinQuantity),
		Optional:            input.Optional,
		ScaleFactor:         database.StringFromFloat32(input.ScaleFactor),
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing recipe step instrument creation query")
	}

	x := &mealplanning.RecipeStepInstrument{
		ID:                  input.ID,
		Instrument:          nil,
		RecipeStepProductID: input.RecipeStepProductID,
		Name:                input.Name,
		Notes:               input.Notes,
		PreferenceRank:      input.PreferenceRank,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Optional:            input.Optional,
		Index:               input.Index,
		OptionIndex:         input.OptionIndex,
		MinQuantity:         input.MinQuantity,
		MaxQuantity:         input.MaxQuantity,
		ScaleFactor:         input.ScaleFactor,
		CreatedAt:           q.CurrentTime(),
	}

	if input.InstrumentID != nil {
		x.Instrument = &mealplanning.ValidInstrument{ID: *input.InstrumentID}
	}

	logger.Info("recipe step instrument created")

	return x, nil
}

// CreateRecipeStepInstrument creates a recipe step instrument in the database.
func (q *repository) CreateRecipeStepInstrument(ctx context.Context, recipeID string, input *mealplanning.RecipeStepInstrumentDatabaseCreationInput) (*mealplanning.RecipeStepInstrument, error) {
	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	var created *mealplanning.RecipeStepInstrument

	// The write and its event share a transaction.
	if err := q.withEvent(ctx, q.logger, mealplanning.RecipeStepInstrumentCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           input.BelongsToRecipeStep,
		mealplanningkeys.RecipeStepInstrumentIDKey: input.ID,
	}, func(tx database.Tx) error {
		var createErr error
		created, createErr = q.createRecipeStepInstrument(ctx, tx, input)

		return createErr
	}); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateRecipeStepInstrument updates a particular recipe step instrument.
func (q *repository) UpdateRecipeStepInstrument(ctx context.Context, recipeID string, updated *mealplanning.RecipeStepInstrument) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeStepInstrumentIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepInstrumentIDKey, updated.ID)

	var instrumentID *string
	if updated.Instrument != nil {
		instrumentID = &updated.Instrument.ID
	}

	if err := q.withEvent(ctx, logger, mealplanning.RecipeStepInstrumentUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           updated.BelongsToRecipeStep,
		mealplanningkeys.RecipeStepInstrumentIDKey: updated.ID,
	}, func(tx database.Tx) error {
		_, updateErr := q.generatedQuerier.UpdateRecipeStepInstrument(ctx, tx, &generated.UpdateRecipeStepInstrumentParams{
			InstrumentID:        database.NullStringFromStringPointer(instrumentID),
			RecipeStepProductID: database.NullStringFromStringPointer(updated.RecipeStepProductID),
			Name:                updated.Name,
			Notes:               updated.Notes,
			PreferenceRank:      int32(updated.PreferenceRank),
			Optional:            updated.Optional,
			Index:               int32(updated.Index),
			OptionIndex:         int32(updated.OptionIndex),
			MinimumQuantity:     int32(updated.MinQuantity),
			MaximumQuantity:     database.NullInt32FromUint32Pointer(updated.MaxQuantity),
			BelongsToRecipeStep: updated.BelongsToRecipeStep,
			ID:                  updated.ID,
			ScaleFactor:         database.StringFromFloat32(updated.ScaleFactor),
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe step instrument")
	}

	logger.Info("recipe step instrument updated")

	return nil
}

// ArchiveRecipeStepInstrument archives a recipe step instrument from the database by its ID.
func (q *repository) ArchiveRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeStepID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if recipeStepInstrumentID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepInstrumentIDKey, recipeStepInstrumentID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepInstrumentIDKey, recipeStepInstrumentID)

	if err := q.withEvent(ctx, logger, mealplanning.RecipeStepInstrumentArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepInstrumentIDKey: recipeStepInstrumentID,
	}, func(tx database.Tx) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveRecipeStepInstrument(ctx, tx, &generated.ArchiveRecipeStepInstrumentParams{
			BelongsToRecipeStep: recipeStepID,
			ID:                  recipeStepInstrumentID,
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving recipe step instrument")
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
