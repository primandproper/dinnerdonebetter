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

func (m *mealPlanningManager) ListRecipePrepTask(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipePrepTask], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	logger = filter.AttachToLogger(logger)

	results, err := m.db.GetRecipePrepTasks(ctx, recipeID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching list of recipe prep tasks")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateRecipePrepTask(ctx context.Context, recipeID string, input *types.RecipePrepTaskCreationRequestInput) (*types.RecipePrepTask, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	convertedInput := converters.ConvertRecipePrepTaskCreationRequestInputToRecipePrepTaskDatabaseCreationInput(input)
	convertedInput.BelongsToRecipe = recipeID
	logger = logger.WithValue(mealplanningkeys.RecipePrepTaskIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipePrepTaskIDKey, convertedInput.ID)

	created, err := m.db.CreateRecipePrepTask(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating recipe prep task")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipePrepTaskCreatedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:         recipeID,
		mealplanningkeys.RecipePrepTaskIDKey: convertedInput.ID,
	}))

	return created, nil
}

func (m *mealPlanningManager) ReadRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string) (*types.RecipePrepTask, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:         recipeID,
		mealplanningkeys.RecipePrepTaskIDKey: recipePrepTaskID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipePrepTaskIDKey, recipePrepTaskID)

	x, err := m.db.GetRecipePrepTask(ctx, recipeID, recipePrepTaskID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "retrieving recipe prep task")
	}

	return x, nil
}

func (m *mealPlanningManager) UpdateRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string, input *types.RecipePrepTaskUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:         recipeID,
		mealplanningkeys.RecipePrepTaskIDKey: recipePrepTaskID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipePrepTaskIDKey, recipePrepTaskID)

	existingRecipePrepTask, err := m.db.GetRecipePrepTask(ctx, recipeID, recipePrepTaskID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "retrieving existing recipe prep task")
	}

	existingRecipePrepTask.Update(input)
	if err = m.db.UpdateRecipePrepTask(ctx, existingRecipePrepTask); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe prep task")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipePrepTaskUpdatedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:         recipeID,
		mealplanningkeys.RecipePrepTaskIDKey: recipePrepTaskID,
	}))

	return nil
}

func (m *mealPlanningManager) ArchiveRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:         recipeID,
		mealplanningkeys.RecipePrepTaskIDKey: recipePrepTaskID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipePrepTaskIDKey, recipePrepTaskID)

	if err := m.db.ArchiveRecipePrepTask(ctx, recipeID, recipePrepTaskID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving recipe prep task")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipePrepTaskArchivedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:         recipeID,
		mealplanningkeys.RecipePrepTaskIDKey: recipePrepTaskID,
	}))

	return nil
}
