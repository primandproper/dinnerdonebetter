package mealplanning

import (
	"context"
	"database/sql"

	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

var (
	_ types.RecipeRatingDataManager = (*repository)(nil)
)

// RecipeRatingExists fetches whether a recipe rating exists from the database.
func (q *repository) RecipeRatingExists(ctx context.Context, recipeID, recipeRatingID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if recipeRatingID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeRatingIDKey, recipeRatingID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeRatingIDKey, recipeRatingID)

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

	logger := q.logger.Clone()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if recipeRatingID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeRatingIDKey, recipeRatingID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeRatingIDKey, recipeRatingID)

	result, err := q.generatedQuerier.GetRecipeRating(ctx, q.readDB, recipeRatingID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe rating")
	}

	recipeRating := &types.RecipeRating{
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

	return recipeRating, nil
}

// GetRecipeRatingsForRecipe fetches a list of recipe ratings from the database that meet a particular filter.
func (q *repository) GetRecipeRatingsForRecipe(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.RecipeRating], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	var (
		data          []*types.RecipeRating
		filteredCount uint64
		totalCount    uint64
	)

	results, err := q.generatedQuerier.GetRecipeRatingsForRecipe(ctx, q.readDB, &generated.GetRecipeRatingsForRecipeParams{
		BelongsToRecipe: recipeID,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe ratings list retrieval query")
	}

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &types.RecipeRating{
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
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
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

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	var (
		data          []*types.RecipeRating
		filteredCount uint64
		totalCount    uint64
	)

	results, err := q.generatedQuerier.GetRecipeRatingsForUser(ctx, q.readDB, &generated.GetRecipeRatingsForUserParams{
		CreatedByUser:   userID,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe ratings list retrieval query")
	}

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &types.RecipeRating{
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
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
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
		return nil, platformerrors.ErrNilInputProvided
	}

	logger := q.logger.WithValue(mealplanningkeys.RecipeRatingIDKey, input.ID)

	// create the recipe rating.
	if err := q.withEvent(ctx, logger, types.RecipeRatingCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:       input.BelongsToRecipe,
		mealplanningkeys.RecipeRatingIDKey: input.ID,
	}, func(tx database.SQLQueryExecutor) error {
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
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeRatingIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeRatingIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, types.RecipeRatingUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:       updated.BelongsToRecipe,
		mealplanningkeys.RecipeRatingIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
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

	logger := q.logger.Clone()

	if recipeID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if recipeRatingID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeRatingIDKey, recipeRatingID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeRatingIDKey, recipeRatingID)

	if err := q.withEvent(ctx, logger, types.RecipeRatingArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:       recipeID,
		mealplanningkeys.RecipeRatingIDKey: recipeRatingID,
	}, func(tx database.SQLQueryExecutor) error {
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
