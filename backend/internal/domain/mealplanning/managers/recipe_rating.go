package managers

import (
	"context"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

func (m *mealPlanningManager) ListRecipeRatings(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeRating], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	logger = filter.AttachToLogger(logger)

	results, err := m.db.GetRecipeRatingsForRecipe(ctx, recipeID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching list of recipe ratings")
	}

	return results, nil
}

func (m *mealPlanningManager) ReadRecipeRating(ctx context.Context, recipeID, recipeRatingID string) (*types.RecipeRating, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:       recipeID,
		mealplanningkeys.RecipeRatingIDKey: recipeRatingID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeRatingIDKey, recipeRatingID)

	x, err := m.db.GetRecipeRating(ctx, recipeID, recipeRatingID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "retrieving recipe rating")
	}

	return x, nil
}

func (m *mealPlanningManager) CreateRecipeRating(ctx context.Context, recipeID string, input *types.RecipeRatingCreationRequestInput) (*types.RecipeRating, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	convertedInput := converters.ConvertRecipeRatingCreationRequestInputToRecipeRatingDatabaseCreationInput(input)
	logger = logger.WithValue(mealplanningkeys.RecipeRatingIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeRatingIDKey, convertedInput.ID)

	created, err := m.db.CreateRecipeRating(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating recipe rating")
	}

	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return created, nil
}

func (m *mealPlanningManager) UpdateRecipeRating(ctx context.Context, recipeID, recipeRatingID string, input *types.RecipeRatingUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:       recipeID,
		mealplanningkeys.RecipeRatingIDKey: recipeRatingID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeRatingIDKey, recipeRatingID)

	existingRecipeRating, err := m.db.GetRecipeRating(ctx, recipeID, recipeRatingID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "retrieving existing recipe rating")
	}

	existingRecipeRating.Update(input)
	if err = m.db.UpdateRecipeRating(ctx, existingRecipeRating); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe rating")
	}

	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return nil
}

func (m *mealPlanningManager) ArchiveRecipeRating(ctx context.Context, recipeID, recipeRatingID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:       recipeID,
		mealplanningkeys.RecipeRatingIDKey: recipeRatingID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeRatingIDKey, recipeRatingID)

	if err := m.db.ArchiveRecipeRating(ctx, recipeID, recipeRatingID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving recipe rating")
	}

	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return nil
}
