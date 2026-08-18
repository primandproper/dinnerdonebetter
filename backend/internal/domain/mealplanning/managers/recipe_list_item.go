package managers

import (
	"context"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/identifiers"
	"github.com/primandproper/platform-go/v11/observability"
	"github.com/primandproper/platform-go/v11/observability/tracing"
)

func (m *mealPlanningManager) UpdateRecipeListItem(ctx context.Context, recipeListItemID, recipeListID, recipeID string, input *types.RecipeListItemUpdateRequestInput) error {
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
	if recipeListItemID == "" || recipeListID == "" || recipeID == "" {
		return platformerrors.ErrEmptyInputParameter
	}
	if input.Notes == nil {
		return platformerrors.ErrNilInputParameter
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

func (m *mealPlanningManager) AddRecipeToRecipeList(ctx context.Context, recipeListID, recipeID, notes string) (*types.RecipeListItem, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if recipeListID == "" || recipeID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
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

func (m *mealPlanningManager) RemoveRecipeFromRecipeList(ctx context.Context, recipeListID, recipeListItemID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if recipeListID == "" || recipeListItemID == "" {
		return platformerrors.ErrEmptyInputParameter
	}

	if err := m.db.ArchiveRecipeListItem(ctx, recipeListItemID, recipeListID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "removing recipe from recipe list")
	}

	return nil
}

func (m *mealPlanningManager) ListRecipeListItems(ctx context.Context, recipeListID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeListItem], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if recipeListID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	res, err := m.db.GetRecipeListItems(ctx, recipeListID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing recipe list items")
	}

	return res, nil
}
