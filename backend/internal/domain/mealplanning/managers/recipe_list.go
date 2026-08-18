package managers

import (
	"context"

	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/identifiers"
	"github.com/primandproper/platform-go/v11/observability"
	"github.com/primandproper/platform-go/v11/observability/tracing"
)

func (m *mealPlanningManager) ListRecipeLists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeList], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	res, err := m.db.GetRecipeLists(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing recipe lists")
	}

	return res, nil
}

func (m *mealPlanningManager) CreateRecipeList(ctx context.Context, userID string, input *types.RecipeListCreationRequestInput) (*types.RecipeList, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	if userID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating recipe list input")
	}

	recipeListID := identifiers.New()
	var items []*types.RecipeListItemDatabaseCreationInput
	for _, item := range input.Items {
		items = append(items, converters.ConvertRecipeListItemCreationRequestInputToRecipeListItemDatabaseCreationInput(item, recipeListID))
	}

	dbInput := &types.RecipeListDatabaseCreationInput{
		ID:            recipeListID,
		Name:          input.Name,
		Description:   input.Description,
		BelongsToUser: userID,
		Items:         items,
	}

	created, err := m.db.CreateRecipeList(ctx, dbInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating recipe list")
	}

	return created, nil
}

func (m *mealPlanningManager) ArchiveRecipeList(ctx context.Context, recipeListID, userID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if recipeListID == "" || userID == "" {
		return platformerrors.ErrEmptyInputParameter
	}

	if err := m.db.ArchiveRecipeList(ctx, recipeListID, userID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving recipe list")
	}

	return nil
}

func (m *mealPlanningManager) UpdateRecipeList(ctx context.Context, recipeListID, userID string, input *types.RecipeListUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeListIDKey: recipeListID,
		identitykeys.UserIDKey:           userID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeListIDKey, recipeListID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}
	if recipeListID == "" || userID == "" {
		return platformerrors.ErrEmptyInputParameter
	}
	if input.Name == nil || input.Description == nil {
		return platformerrors.ErrNilInputParameter
	}

	updated := &types.RecipeList{
		ID:            recipeListID,
		BelongsToUser: userID,
	}
	updated.Update(input)

	if err := m.db.UpdateRecipeList(ctx, updated); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe list")
	}

	return nil
}
