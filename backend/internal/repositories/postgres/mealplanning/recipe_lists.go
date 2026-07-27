package mealplanning

import (
	"context"
	"database/sql"

	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v7/database"
	platformerrors "github.com/primandproper/platform-go/v7/errors"
	"github.com/primandproper/platform-go/v7/filtering"
	"github.com/primandproper/platform-go/v7/observability"
	"github.com/primandproper/platform-go/v7/observability/tracing"
)

var (
	_ types.RecipeListDataManager = (*repository)(nil)
)

// GetRecipeLists fetches a list of recipe lists from the database that meet a particular filter.
func (q *repository) GetRecipeLists(ctx context.Context, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.RecipeList], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	var (
		data          []*types.RecipeList
		filteredCount uint64
		totalCount    uint64
	)
	listsByID := map[string]*types.RecipeList{}

	results, err := q.generatedQuerier.GetRecipeLists(ctx, q.readDB, &generated.GetRecipeListsParams{
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
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
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.RecipeListIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.RecipeListIDKey, input.ID)

	if err := q.generatedQuerier.CreateRecipeList(ctx, q.writeDB, &generated.CreateRecipeListParams{
		ID:            input.ID,
		Name:          input.Name,
		Description:   input.Description,
		BelongsToUser: input.BelongsToUser,
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
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeListIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeListIDKey, updated.ID)

	rowsAffected, err := q.generatedQuerier.UpdateRecipeList(ctx, q.writeDB, &generated.UpdateRecipeListParams{
		Name:          updated.Name,
		Description:   updated.Description,
		BelongsToUser: updated.BelongsToUser,
		ID:            updated.ID,
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe list")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	logger.Info("recipe list updated")

	return nil
}

// ArchiveRecipeList archives a recipe list from the database by its ID.
func (q *repository) ArchiveRecipeList(ctx context.Context, recipeListID, userID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if recipeListID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeListIDKey, recipeListID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeListIDKey, recipeListID)

	rowsAffected, err := q.generatedQuerier.ArchiveRecipeList(ctx, q.writeDB, &generated.ArchiveRecipeListParams{
		BelongsToUser: userID,
		ID:            recipeListID,
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving recipe list")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	logger.Info("recipe list archived")

	return nil
}
