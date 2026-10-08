package mealplanning

import (
	"context"
	"database/sql"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

var (
	_ types.RecipeMediaDataManager = (*repository)(nil)
)

// RecipeMediaExists fetches whether a recipe media exists from the database.
func (q *repository) RecipeMediaExists(ctx context.Context, recipeMediaID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span, idArg{mealplanningkeys.RecipeMediaIDKey, recipeMediaID})
	if err != nil {
		return false, err
	}

	result, err := q.generatedQuerier.CheckRecipeMediaExistence(ctx, q.readDB, recipeMediaID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing recipe media existence check")
	}

	return result, nil
}

// GetRecipeMedia fetches a recipe media from the database.
func (q *repository) GetRecipeMedia(ctx context.Context, recipeMediaID string) (*types.RecipeMedia, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span, idArg{mealplanningkeys.RecipeMediaIDKey, recipeMediaID})
	if err != nil {
		return nil, err
	}

	result, err := q.generatedQuerier.GetRecipeMedia(ctx, q.readDB, recipeMediaID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting recipe media")
	}

	recipeMedia := &types.RecipeMedia{
		CreatedAt:           result.CreatedAt,
		ArchivedAt:          database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt:       database.TimePointerFromNullTime(result.LastUpdatedAt),
		ID:                  result.ID,
		BelongsToRecipe:     database.StringPointerFromNullString(result.BelongsToRecipe),
		BelongsToRecipeStep: database.StringPointerFromNullString(result.BelongsToRecipeStep),
		MimeType:            result.MimeType,
		InternalPath:        result.InternalPath,
		ExternalPath:        result.ExternalPath,
		Index:               uint16(result.Index),
	}

	return recipeMedia, nil
}

// getRecipeMediaForRecipe fetches a list of recipe media from the database that meet a particular filter.
func (q *repository) getRecipeMediaForRecipe(ctx context.Context, recipeID string) ([]*types.RecipeMedia, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span, idArg{mealplanningkeys.RecipeIDKey, recipeID})
	if err != nil {
		return nil, err
	}

	results, err := q.generatedQuerier.GetRecipeMediaForRecipe(ctx, q.readDB, database.NullStringFromString(recipeID))
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe media list retrieval query")
	}

	recipeMedia := make([]*types.RecipeMedia, len(results))
	for i, result := range results {
		recipeMedia[i] = &types.RecipeMedia{
			CreatedAt:           result.CreatedAt,
			ArchivedAt:          database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:       database.TimePointerFromNullTime(result.LastUpdatedAt),
			ID:                  result.ID,
			BelongsToRecipe:     database.StringPointerFromNullString(result.BelongsToRecipe),
			BelongsToRecipeStep: database.StringPointerFromNullString(result.BelongsToRecipeStep),
			MimeType:            result.MimeType,
			InternalPath:        result.InternalPath,
			ExternalPath:        result.ExternalPath,
			Index:               uint16(result.Index),
		}
	}

	return recipeMedia, nil
}

// getRecipeMediaForRecipeStep fetches a list of recipe media from the database that meet a particular filter.
func (q *repository) getRecipeMediaForRecipeStep(ctx context.Context, recipeID, recipeStepID string) ([]*types.RecipeMedia, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span,
		idArg{mealplanningkeys.RecipeIDKey, recipeID},
		idArg{mealplanningkeys.RecipeStepIDKey, recipeStepID},
	)
	if err != nil {
		return nil, err
	}

	results, err := q.generatedQuerier.GetRecipeMediaForRecipeStep(ctx, q.readDB, &generated.GetRecipeMediaForRecipeStepParams{
		RecipeID:     database.NullStringFromString(recipeID),
		RecipeStepID: database.NullStringFromString(recipeStepID),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe media list retrieval query")
	}

	recipeMedia := []*types.RecipeMedia{}
	for _, result := range results {
		recipeMedia = append(recipeMedia, &types.RecipeMedia{
			CreatedAt:           result.CreatedAt,
			ArchivedAt:          database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:       database.TimePointerFromNullTime(result.LastUpdatedAt),
			ID:                  result.ID,
			BelongsToRecipe:     database.StringPointerFromNullString(result.BelongsToRecipe),
			BelongsToRecipeStep: database.StringPointerFromNullString(result.BelongsToRecipeStep),
			MimeType:            result.MimeType,
			InternalPath:        result.InternalPath,
			ExternalPath:        result.ExternalPath,
			Index:               uint16(result.Index),
		})
	}

	return recipeMedia, nil
}

// CreateRecipeMedia creates a recipe media in the database.
func (q *repository) CreateRecipeMedia(ctx context.Context, input *types.RecipeMediaDatabaseCreationInput) (*types.RecipeMedia, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := q.logger.WithValue(mealplanningkeys.RecipeMediaIDKey, input.ID)

	// create the recipe media.
	if err := q.withEvent(ctx, logger, types.RecipeMediaCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeMediaIDKey: input.ID,
	}, func(tx database.Tx) error {
		return q.generatedQuerier.CreateRecipeMedia(ctx, tx, &generated.CreateRecipeMediaParams{
			ID:                  input.ID,
			MimeType:            input.MimeType,
			InternalPath:        input.InternalPath,
			ExternalPath:        input.ExternalPath,
			BelongsToRecipe:     database.NullStringFromStringPointer(input.BelongsToRecipe),
			BelongsToRecipeStep: database.NullStringFromStringPointer(input.BelongsToRecipeStep),
			Index:               int32(input.Index),
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing recipe media creation query")
	}

	x := &types.RecipeMedia{
		ID:                  input.ID,
		BelongsToRecipe:     input.BelongsToRecipe,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		MimeType:            input.MimeType,
		InternalPath:        input.InternalPath,
		ExternalPath:        input.ExternalPath,
		Index:               input.Index,
		CreatedAt:           q.CurrentTime(),
	}

	tracing.AttachToSpan(span, mealplanningkeys.RecipeMediaIDKey, x.ID)
	logger.Info("recipe media created")

	return x, nil
}

// UpdateRecipeMedia updates a particular recipe media.
func (q *repository) UpdateRecipeMedia(ctx context.Context, updated *types.RecipeMedia) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeMediaIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeMediaIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, types.RecipeMediaUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeMediaIDKey: updated.ID,
	}, func(tx database.Tx) error {
		_, writeErr := q.generatedQuerier.UpdateRecipeMedia(ctx, tx, &generated.UpdateRecipeMediaParams{
			ID:                  updated.ID,
			BelongsToRecipe:     database.NullStringFromStringPointer(updated.BelongsToRecipe),
			BelongsToRecipeStep: database.NullStringFromStringPointer(updated.BelongsToRecipeStep),
			MimeType:            updated.MimeType,
			InternalPath:        updated.InternalPath,
			ExternalPath:        updated.ExternalPath,
			Index:               int32(updated.Index),
		})

		return writeErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe media")
	}

	logger.Info("recipe media updated")

	return nil
}

// ArchiveRecipeMedia archives a recipe media from the database by its ID.
func (q *repository) ArchiveRecipeMedia(ctx context.Context, recipeMediaID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span, idArg{mealplanningkeys.RecipeMediaIDKey, recipeMediaID})
	if err != nil {
		return err
	}

	return q.withEvent(ctx, logger, types.RecipeMediaArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeMediaIDKey: recipeMediaID,
	}, func(tx database.Tx) error {
		rowsAffected, writeErr := q.generatedQuerier.ArchiveRecipeMedia(ctx, tx, recipeMediaID)
		if writeErr != nil {
			return observability.PrepareAndLogError(writeErr, logger, span, "archiving recipe media")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}
