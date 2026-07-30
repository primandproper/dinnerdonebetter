package managers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

func (m *mealPlanningManager) ListRecipeStepCompletionConditions(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepCompletionCondition], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     recipeID,
		mealplanningkeys.RecipeStepIDKey: recipeStepID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	logger = filter.AttachToLogger(logger)

	results, err := m.db.GetRecipeStepCompletionConditions(ctx, recipeID, recipeStepID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching list of recipe step completion conditions")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID string, input *types.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput) (*types.RecipeStepCompletionCondition, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     recipeID,
		mealplanningkeys.RecipeStepIDKey: recipeStepID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	convertedInput := converters.ConvertRecipeStepCompletionConditionForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionDatabaseCreationInput(input)
	convertedInput.BelongsToRecipeStep = recipeStepID
	logger = logger.WithValue(mealplanningkeys.RecipeStepCompletionConditionIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepCompletionConditionIDKey, convertedInput.ID)

	created, err := m.db.CreateRecipeStepCompletionCondition(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating recipe step completion condition")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipeStepCompletionConditionCreatedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:                        recipeID,
		mealplanningkeys.RecipeStepIDKey:                    recipeStepID,
		mealplanningkeys.RecipeStepCompletionConditionIDKey: convertedInput.ID,
	}))

	return created, nil
}

func (m *mealPlanningManager) ReadRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID, recipeStepCompletionConditionID string) (*types.RecipeStepCompletionCondition, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:                        recipeID,
		mealplanningkeys.RecipeStepIDKey:                    recipeStepID,
		mealplanningkeys.RecipeStepCompletionConditionIDKey: recipeStepCompletionConditionID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepCompletionConditionIDKey, recipeStepCompletionConditionID)

	x, err := m.db.GetRecipeStepCompletionCondition(ctx, recipeID, recipeStepID, recipeStepCompletionConditionID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "retrieving recipe step completion condition")
	}

	return x, nil
}

func (m *mealPlanningManager) UpdateRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID, recipeStepCompletionConditionID string, input *types.RecipeStepCompletionConditionUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:                        recipeID,
		mealplanningkeys.RecipeStepIDKey:                    recipeStepID,
		mealplanningkeys.RecipeStepCompletionConditionIDKey: recipeStepCompletionConditionID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepCompletionConditionIDKey, recipeStepCompletionConditionID)

	existingRecipeStepCompletionCondition, err := m.db.GetRecipeStepCompletionCondition(ctx, recipeID, recipeStepID, recipeStepCompletionConditionID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "retrieving existing recipe step completion condition")
	}

	existingRecipeStepCompletionCondition.Update(input)
	if err = m.db.UpdateRecipeStepCompletionCondition(ctx, existingRecipeStepCompletionCondition); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe step completion condition")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipeStepCompletionConditionUpdatedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:                        recipeID,
		mealplanningkeys.RecipeStepIDKey:                    recipeStepID,
		mealplanningkeys.RecipeStepCompletionConditionIDKey: recipeStepCompletionConditionID,
	}))

	return nil
}

func (m *mealPlanningManager) ArchiveRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID, recipeStepCompletionConditionID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:                        recipeID,
		mealplanningkeys.RecipeStepIDKey:                    recipeStepID,
		mealplanningkeys.RecipeStepCompletionConditionIDKey: recipeStepCompletionConditionID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepCompletionConditionIDKey, recipeStepCompletionConditionID)

	if err := m.db.ArchiveRecipeStepCompletionCondition(ctx, recipeStepID, recipeStepCompletionConditionID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving recipe step completion condition")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipeStepCompletionConditionArchivedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:                        recipeID,
		mealplanningkeys.RecipeStepIDKey:                    recipeStepID,
		mealplanningkeys.RecipeStepCompletionConditionIDKey: recipeStepCompletionConditionID,
	}))

	return nil
}
