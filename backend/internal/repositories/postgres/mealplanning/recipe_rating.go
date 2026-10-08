package mealplanning

import (
	"context"
	"database/sql"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/observability"
	platformkeys "github.com/primandproper/primitives-go/v2/observability/keys"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

var (
	_ types.RecipeRatingDataManager = (*repository)(nil)
)

// RecipeRatingExists fetches whether a recipe rating exists from the database.
func (q *repository) RecipeRatingExists(ctx context.Context, recipeID, recipeRatingID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span,
		idArg{mealplanningkeys.RecipeIDKey, recipeID},
		idArg{mealplanningkeys.RecipeRatingIDKey, recipeRatingID},
	)
	if err != nil {
		return false, err
	}

	result, err := q.generatedQuerier.CheckRecipeRatingExistence(ctx, q.readDB, recipeRatingID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing recipe rating existence check")
	}

	return result, nil
}

// GetRecipeRating fetches a recipe rating from the database.
func (q *repository) GetRecipeRating(ctx context.Context, recipeID, recipeRatingID string) (*types.RecipeRating, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span,
		idArg{mealplanningkeys.RecipeIDKey, recipeID},
		idArg{mealplanningkeys.RecipeRatingIDKey, recipeRatingID},
	)
	if err != nil {
		return nil, err
	}

	result, err := q.generatedQuerier.GetRecipeRating(ctx, q.readDB, recipeRatingID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe rating")
	}

	return recipeRatingFromRow(result), nil
}

// GetRecipeRatingsForRecipe fetches a list of recipe ratings from the database that meet a particular filter.
func (q *repository) GetRecipeRatingsForRecipe(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.RecipeRating], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	filter, logger = filtering.Observe(ctx, logger, filter)

	logger, err = guardIDs(logger, span, idArg{mealplanningkeys.RecipeIDKey, recipeID})
	if err != nil {
		return nil, err
	}

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetRecipeRatingsForRecipe(ctx, q.readDB, &generated.GetRecipeRatingsForRecipeParams{
		BelongsToRecipe: recipeID,
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe ratings list retrieval query")
	}

	x = filtering.Drain(
		results,
		recipeRatingFromListRow,
		func(result *generated.GetRecipeRatingsForRecipeRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(rr *types.RecipeRating) string { return rr.ID },
		filter,
	)

	return x, nil
}

// GetRecipeRatingsForUser fetches a list of recipe ratings from the database that meet a particular filter.
func (q *repository) GetRecipeRatingsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.RecipeRating], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	filter, logger = filtering.Observe(ctx, logger, filter)

	logger, err = guardIDs(logger, span, idArg{platformkeys.UserIDKey, userID})
	if err != nil {
		return nil, err
	}

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetRecipeRatingsForUser(ctx, q.readDB, &generated.GetRecipeRatingsForUserParams{
		CreatedByUser:   userID,
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe ratings list retrieval query")
	}

	x = filtering.Drain(
		results,
		func(result *generated.GetRecipeRatingsForUserRow) *types.RecipeRating {
			return recipeRatingFromListRow((*generated.GetRecipeRatingsForRecipeRow)(result))
		},
		func(result *generated.GetRecipeRatingsForUserRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(rr *types.RecipeRating) string { return rr.ID },
		filter,
	)

	return x, nil
}

// CreateRecipeRating creates a recipe rating in the database.
func (q *repository) CreateRecipeRating(ctx context.Context, input *types.RecipeRatingDatabaseCreationInput) (*types.RecipeRating, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := q.logger.WithValue(mealplanningkeys.RecipeRatingIDKey, input.ID)

	// create the recipe rating.
	if err := q.withEvent(ctx, logger, types.RecipeRatingCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:       input.BelongsToRecipe,
		mealplanningkeys.RecipeRatingIDKey: input.ID,
	}, func(tx database.Tx) error {
		return q.generatedQuerier.CreateRecipeRating(ctx, tx, &generated.CreateRecipeRatingParams{
			ID:              input.ID,
			BelongsToRecipe: input.BelongsToRecipe,
			Notes:           input.Notes,
			CreatedByUser:   input.CreatedByUser,
			Taste:           database.NullStringFromFloat32(input.Taste),
			Difficulty:      database.NullStringFromFloat32(input.Difficulty),
			Cleanup:         database.NullStringFromFloat32(input.Cleanup),
			Instructions:    database.NullStringFromFloat32(input.Instructions),
			Overall:         database.NullStringFromFloat32(input.Overall),
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing recipe rating creation query")
	}

	x := &types.RecipeRating{
		ID:              input.ID,
		BelongsToRecipe: input.BelongsToRecipe,
		Taste:           input.Taste,
		Difficulty:      input.Difficulty,
		Cleanup:         input.Cleanup,
		Instructions:    input.Instructions,
		Overall:         input.Overall,
		Notes:           input.Notes,
		CreatedByUser:   input.CreatedByUser,
		CreatedAt:       q.CurrentTime(),
	}

	tracing.AttachToSpan(span, mealplanningkeys.RecipeRatingIDKey, x.ID)
	logger.Info("recipe rating created")

	return x, nil
}

// UpdateRecipeRating updates a particular recipe rating.
func (q *repository) UpdateRecipeRating(ctx context.Context, updated *types.RecipeRating) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeRatingIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeRatingIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, types.RecipeRatingUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:       updated.BelongsToRecipe,
		mealplanningkeys.RecipeRatingIDKey: updated.ID,
	}, func(tx database.Tx) error {
		_, updateErr := q.generatedQuerier.UpdateRecipeRating(ctx, tx, &generated.UpdateRecipeRatingParams{
			BelongsToRecipe: updated.BelongsToRecipe,
			Taste:           database.NullStringFromFloat32(updated.Taste),
			Difficulty:      database.NullStringFromFloat32(updated.Difficulty),
			Cleanup:         database.NullStringFromFloat32(updated.Cleanup),
			Instructions:    database.NullStringFromFloat32(updated.Instructions),
			Overall:         database.NullStringFromFloat32(updated.Overall),
			Notes:           updated.Notes,
			ID:              updated.ID,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe rating")
	}

	logger.Info("recipe rating updated")

	return nil
}

// ArchiveRecipeRating archives a recipe rating from the database by its ID.
func (q *repository) ArchiveRecipeRating(ctx context.Context, recipeID, recipeRatingID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span,
		idArg{mealplanningkeys.RecipeIDKey, recipeID},
		idArg{mealplanningkeys.RecipeRatingIDKey, recipeRatingID},
	)
	if err != nil {
		return err
	}

	if err = q.withEvent(ctx, logger, types.RecipeRatingArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:       recipeID,
		mealplanningkeys.RecipeRatingIDKey: recipeRatingID,
	}, func(tx database.Tx) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveRecipeRating(ctx, tx, recipeRatingID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving recipe rating")
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

// recipeRatingFromRow maps a recipe rating row to its domain type. The single-rating read returns
// the table model, generated.RecipeRatings; the paginated reads go through recipeRatingFromListRow.
func recipeRatingFromRow(result *generated.RecipeRatings) *types.RecipeRating {
	return &types.RecipeRating{
		CreatedAt:       result.CreatedAt,
		LastUpdatedAt:   database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:      database.TimePointerFromNullTime(result.ArchivedAt),
		Notes:           result.Notes,
		ID:              result.ID,
		BelongsToRecipe: result.BelongsToRecipe,
		CreatedByUser:   result.CreatedByUser,
		Taste:           database.Float32FromNullString(result.Taste),
		Instructions:    database.Float32FromNullString(result.Instructions),
		Overall:         database.Float32FromNullString(result.Overall),
		Cleanup:         database.Float32FromNullString(result.Cleanup),
		Difficulty:      database.Float32FromNullString(result.Difficulty),
	}
}

// recipeRatingFromListRow is recipeRatingFromRow for the paginated reads, whose rows also carry the
// filtered and total counts. The by-recipe and by-user reads select the same columns, so the
// by-user row converts to generated.GetRecipeRatingsForRecipeRow.
func recipeRatingFromListRow(result *generated.GetRecipeRatingsForRecipeRow) *types.RecipeRating {
	return &types.RecipeRating{
		CreatedAt:       result.CreatedAt,
		LastUpdatedAt:   database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:      database.TimePointerFromNullTime(result.ArchivedAt),
		Notes:           result.Notes,
		ID:              result.ID,
		BelongsToRecipe: result.BelongsToRecipe,
		CreatedByUser:   result.CreatedByUser,
		Taste:           database.Float32FromNullString(result.Taste),
		Instructions:    database.Float32FromNullString(result.Instructions),
		Overall:         database.Float32FromNullString(result.Overall),
		Cleanup:         database.Float32FromNullString(result.Cleanup),
		Difficulty:      database.Float32FromNullString(result.Difficulty),
	}
}
