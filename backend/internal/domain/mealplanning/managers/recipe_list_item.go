package managers

import (
	"context"
	"database/sql"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// verifyRecipeListOwnership confirms the recipe list exists and belongs to userID. Another user's
// list answers sql.ErrNoRows, exactly as a missing one does, so a caller can't probe which list IDs exist.
func (m *mealPlanningManager) verifyRecipeListOwnership(ctx context.Context, recipeListID, userID string) error {
	owned, err := m.db.RecipeListExists(ctx, recipeListID, userID)
	if err != nil {
		return err
	}
	if !owned {
		return sql.ErrNoRows
	}

	return nil
}

func (m *mealPlanningManager) UpdateRecipeListItem(ctx context.Context, recipeListItemID, recipeListID, userID, recipeID string, input *types.RecipeListItemUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeListIDKey:     recipeListID,
		mealplanningkeys.RecipeListItemIDKey: recipeListItemID,
		mealplanningkeys.RecipeIDKey:         recipeID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeListIDKey, recipeListID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeListItemIDKey, recipeListItemID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}
	if recipeListItemID == "" || recipeListID == "" || userID == "" || recipeID == "" {
		return platformerrors.ErrEmptyInputParameter
	}
	if input.Notes == nil {
		return platformerrors.ErrNilInputParameter
	}

	if err := m.verifyRecipeListOwnership(ctx, recipeListID, userID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "verifying recipe list ownership")
	}

	updated := &types.RecipeListItem{
		ID:                  recipeListItemID,
		BelongsToRecipeList: recipeListID,
		Recipe:              types.Recipe{ID: recipeID},
	}
	updated.Update(input)

	if err := m.db.UpdateRecipeListItem(ctx, updated); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe list item")
	}

	return nil
}

func (m *mealPlanningManager) AddRecipeToRecipeList(ctx context.Context, recipeListID, userID, recipeID, notes string) (*types.RecipeListItem, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if recipeListID == "" || userID == "" || recipeID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	if err := m.verifyRecipeListOwnership(ctx, recipeListID, userID); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "verifying recipe list ownership")
	}

	input := &types.RecipeListItemDatabaseCreationInput{
		ID:                  identifiers.New(),
		RecipeID:            recipeID,
		Notes:               notes,
		BelongsToRecipeList: recipeListID,
	}

	item, err := m.db.CreateRecipeListItem(ctx, input)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "adding recipe to recipe list")
	}

	return item, nil
}

func (m *mealPlanningManager) RemoveRecipeFromRecipeList(ctx context.Context, recipeListID, userID, recipeListItemID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if recipeListID == "" || userID == "" || recipeListItemID == "" {
		return platformerrors.ErrEmptyInputParameter
	}

	if err := m.verifyRecipeListOwnership(ctx, recipeListID, userID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "verifying recipe list ownership")
	}

	if err := m.db.ArchiveRecipeListItem(ctx, recipeListItemID, recipeListID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "removing recipe from recipe list")
	}

	return nil
}

func (m *mealPlanningManager) ListRecipeListItems(ctx context.Context, recipeListID, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeListItem], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if recipeListID == "" || userID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	if err := m.verifyRecipeListOwnership(ctx, recipeListID, userID); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "verifying recipe list ownership")
	}

	res, err := m.db.GetRecipeListItems(ctx, recipeListID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing recipe list items")
	}

	return res, nil
}
