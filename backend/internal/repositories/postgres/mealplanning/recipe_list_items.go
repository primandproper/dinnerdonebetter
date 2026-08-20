package mealplanning

import (
	"context"
	"database/sql"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v12/database"
	platformerrors "github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

var (
	_ types.RecipeListItemDataManager = (*repository)(nil)
)

// GetRecipeListItems fetches recipe list items for a given list with filtering.
func (q *repository) GetRecipeListItems(ctx context.Context, recipeListID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.RecipeListItem], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeListID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeListIDKey, recipeListID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeListIDKey, recipeListID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	var (
		data          []*types.RecipeListItem
		filteredCount uint64
		totalCount    uint64
	)
	recipeIDs := []string{}

	results, err := q.generatedQuerier.GetRecipeListItems(ctx, q.readDB, &generated.GetRecipeListItemsParams{
		RecipeListID:    recipeListID,
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe list items list retrieval query")
	}

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}

		data = append(data, &types.RecipeListItem{
			CreatedAt:           result.CreatedAt,
			LastUpdatedAt:       database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:          database.TimePointerFromNullTime(result.ArchivedAt),
			ID:                  result.ID,
			Recipe:              types.Recipe{ID: result.RecipeID},
			Notes:               result.Notes,
			BelongsToRecipeList: result.BelongsToRecipeList,
		})

		if result.RecipeID != "" {
			recipeIDs = append(recipeIDs, result.RecipeID)
		}
	}

	if len(recipeIDs) > 0 {
		recipes, recipesFetchErr := q.GetRecipesWithIDs(ctx, recipeIDs)
		if recipesFetchErr != nil {
			return nil, observability.PrepareAndLogError(recipesFetchErr, logger, span, "fetching recipes for recipe list items")
		}
		recipesByID := map[string]*types.Recipe{}
		for _, r := range recipes {
			recipesByID[r.ID] = r
		}
		for i, item := range data {
			if r, ok := recipesByID[item.Recipe.ID]; ok && r != nil {
				data[i].Recipe = *r
			}
		}
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(rli *types.RecipeListItem) string { return rli.ID },
		filter,
	)

	return x, nil
}

// CreateRecipeListItem creates a recipe list item in the database.
func (q *repository) CreateRecipeListItem(ctx context.Context, input *types.RecipeListItemDatabaseCreationInput) (*types.RecipeListItem, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, mealplanningkeys.RecipeListItemIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.RecipeListItemIDKey, input.ID)

	if err := q.generatedQuerier.CreateRecipeListItem(ctx, q.writeDB, &generated.CreateRecipeListItemParams{
		ID:                  input.ID,
		RecipeID:            input.RecipeID,
		Notes:               input.Notes,
		BelongsToRecipeList: input.BelongsToRecipeList,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing recipe list item creation query")
	}

	x := &types.RecipeListItem{
		ID:                  input.ID,
		Recipe:              types.Recipe{ID: input.RecipeID},
		Notes:               input.Notes,
		BelongsToRecipeList: input.BelongsToRecipeList,
		CreatedAt:           q.CurrentTime(),
	}

	logger.Info("recipe list item created")

	return x, nil
}

// UpdateRecipeListItem updates a particular recipe list item.
func (q *repository) UpdateRecipeListItem(ctx context.Context, updated *types.RecipeListItem) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeListItemIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeListItemIDKey, updated.ID)

	rowsAffected, err := q.generatedQuerier.UpdateRecipeListItem(ctx, q.writeDB, &generated.UpdateRecipeListItemParams{
		RecipeID:            updated.Recipe.ID,
		Notes:               updated.Notes,
		BelongsToRecipeList: updated.BelongsToRecipeList,
		ID:                  updated.ID,
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe list item")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	logger.Info("recipe list item updated")

	return nil
}

// ArchiveRecipeListItem archives a recipe list item from the database by its ID.
func (q *repository) ArchiveRecipeListItem(ctx context.Context, recipeListItemID, recipeListID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeListID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeListIDKey, recipeListID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeListIDKey, recipeListID)

	if recipeListItemID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeListItemIDKey, recipeListItemID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeListItemIDKey, recipeListItemID)

	rowsAffected, err := q.generatedQuerier.ArchiveRecipeListItem(ctx, q.writeDB, &generated.ArchiveRecipeListItemParams{
		BelongsToRecipeList: recipeListID,
		ID:                  recipeListItemID,
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving recipe list item")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	logger.Info("recipe list item archived")

	return nil
}
