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
	_ types.RecipeListDataManager = (*repository)(nil)
)

// RecipeListExists reports whether an unarchived recipe list with the given ID belongs to the given user.
func (q *repository) RecipeListExists(ctx context.Context, recipeListID, userID string) (bool, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span,
		idArg{mealplanningkeys.RecipeListIDKey, recipeListID},
		idArg{platformkeys.UserIDKey, userID},
	)
	if err != nil {
		return false, err
	}

	result, err := q.generatedQuerier.CheckRecipeListExistence(ctx, q.readDB, &generated.CheckRecipeListExistenceParams{
		ID:            recipeListID,
		BelongsToUser: userID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing recipe list existence check")
	}

	return result, nil
}

// GetRecipeLists fetches a list of the given user's recipe lists from the database that meet a particular filter.
func (q *repository) GetRecipeLists(ctx context.Context, userID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.RecipeList], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span, idArg{platformkeys.UserIDKey, userID})
	if err != nil {
		return nil, err
	}

	filter, logger = filtering.Observe(ctx, logger, filter)

	var (
		data          []*types.RecipeList
		filteredCount uint64
		totalCount    uint64
	)
	listsByID := map[string]*types.RecipeList{}

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetRecipeLists(ctx, q.readDB, &generated.GetRecipeListsParams{
		CreatedAfter:    filterArgs.CreatedAfter,
		CreatedBefore:   filterArgs.CreatedBefore,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
		BelongsToUser:   userID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe lists list retrieval query")
	}

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}

		rl, exists := listsByID[result.ID]
		if !exists {
			rl = &types.RecipeList{
				CreatedAt:     result.CreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
				ID:            result.ID,
				Name:          result.Name,
				Description:   result.Description,
				BelongsToUser: result.BelongsToUser,
				Items:         []*types.RecipeListItem{},
			}
			listsByID[result.ID] = rl
			data = append(data, rl)
		}

		if result.RecipeListItemID.Valid && result.RecipeListItemID.String != "" {
			rl.Items = append(rl.Items, &types.RecipeListItem{
				CreatedAt:           database.TimeFromNullTime(result.RecipeListItemCreatedAt),
				LastUpdatedAt:       database.TimePointerFromNullTime(result.RecipeListItemLastUpdatedAt),
				ArchivedAt:          database.TimePointerFromNullTime(result.RecipeListItemArchivedAt),
				ID:                  result.RecipeListItemID.String,
				Recipe:              types.Recipe{ID: result.RecipeListItemRecipeID.String},
				Notes:               result.RecipeListItemNotes.String,
				BelongsToRecipeList: result.RecipeListItemBelongsToRecipeList.String,
			})
		}
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(rl *types.RecipeList) string { return rl.ID },
		filter,
	)

	return x, nil
}

// CreateRecipeList creates a recipe list in the database.
func (q *repository) CreateRecipeList(ctx context.Context, input *types.RecipeListDatabaseCreationInput) (*types.RecipeList, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, mealplanningkeys.RecipeListIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.RecipeListIDKey, input.ID)

	if err := q.withEvent(ctx, logger, types.RecipeListCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeListIDKey: input.ID,
	}, func(tx database.Tx) error {
		return q.generatedQuerier.CreateRecipeList(ctx, tx, &generated.CreateRecipeListParams{
			ID:            input.ID,
			Name:          input.Name,
			Description:   input.Description,
			BelongsToUser: input.BelongsToUser,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing recipe list creation query")
	}

	x := &types.RecipeList{
		ID:            input.ID,
		Name:          input.Name,
		Description:   input.Description,
		BelongsToUser: input.BelongsToUser,
		CreatedAt:     q.CurrentTime(),
	}

	logger.Info("recipe list created")

	return x, nil
}

// UpdateRecipeList updates a particular recipe list.
func (q *repository) UpdateRecipeList(ctx context.Context, updated *types.RecipeList) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeListIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeListIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, types.RecipeListUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeListIDKey: updated.ID,
	}, func(tx database.Tx) error {
		rowsAffected, writeErr := q.generatedQuerier.UpdateRecipeList(ctx, tx, &generated.UpdateRecipeListParams{
			Name:          updated.Name,
			Description:   updated.Description,
			BelongsToUser: updated.BelongsToUser,
			ID:            updated.ID,
		})
		if writeErr != nil {
			return observability.PrepareAndLogError(writeErr, logger, span, "updating recipe list")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("recipe list updated")

	return nil
}

// ArchiveRecipeList archives a recipe list from the database by its ID.
func (q *repository) ArchiveRecipeList(ctx context.Context, recipeListID, userID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span,
		idArg{platformkeys.UserIDKey, userID},
		idArg{mealplanningkeys.RecipeListIDKey, recipeListID},
	)
	if err != nil {
		return err
	}

	if err = q.withEvent(ctx, logger, types.RecipeListArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeListIDKey: recipeListID,
	}, func(tx database.Tx) error {
		rowsAffected, writeErr := q.generatedQuerier.ArchiveRecipeList(ctx, tx, &generated.ArchiveRecipeListParams{
			BelongsToUser: userID,
			ID:            recipeListID,
		})
		if writeErr != nil {
			return observability.PrepareAndLogError(writeErr, logger, span, "archiving recipe list")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("recipe list archived")

	return nil
}
