package mealplanning

import (
	"context"
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v8/database"
	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

var (
	_ mealplanning.RecipePrepTaskDataManager = (*repository)(nil)
)

// RecipePrepTaskExists checks if a recipe prep task exists.
func (q *repository) RecipePrepTaskExists(ctx context.Context, recipeID, recipePrepTaskID string) (bool, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if recipePrepTaskID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipePrepTaskIDKey, recipePrepTaskID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipePrepTaskIDKey, recipePrepTaskID)

	result, err := q.generatedQuerier.CheckRecipePrepTaskExistence(ctx, q.readDB, &generated.CheckRecipePrepTaskExistenceParams{
		RecipeID:         recipeID,
		RecipePrepTaskID: recipePrepTaskID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing recipe prep task existence check")
	}

	logger.Info("recipe prep task existence retrieved")

	return result, nil
}

// GetRecipePrepTask fetches a recipe prep task.
func (q *repository) GetRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string) (x *mealplanning.RecipePrepTask, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if recipePrepTaskID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipePrepTaskIDKey, recipePrepTaskID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipePrepTaskIDKey, recipePrepTaskID)

	results, err := q.generatedQuerier.GetRecipePrepTask(ctx, q.readDB, recipePrepTaskID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting recipe prep task")
	}

	for _, result := range results {
		if x == nil {
			x = &mealplanning.RecipePrepTask{
				CreatedAt:                          result.CreatedAt,
				MinStorageTemperatureInCelsius:     database.Float32PointerFromNullString(result.MinimumStorageTemperatureInCelsius),
				MaxStorageTemperatureInCelsius:     database.Float32PointerFromNullString(result.MaximumStorageTemperatureInCelsius),
				MinTimeBufferBeforeRecipeInSeconds: uint32(result.MinimumTimeBufferBeforeRecipeInSeconds),
				MaxTimeBufferBeforeRecipeInSeconds: database.Uint32PointerFromNullInt32(result.MaximumTimeBufferBeforeRecipeInSeconds),
				ArchivedAt:                         database.TimePointerFromNullTime(result.ArchivedAt),
				LastUpdatedAt:                      database.TimePointerFromNullTime(result.LastUpdatedAt),
				ID:                                 result.ID,
				BelongsToRecipe:                    result.BelongsToRecipe,
				ExplicitStorageInstructions:        result.ExplicitStorageInstructions,
				Notes:                              result.Notes,
				Name:                               result.Name,
				Description:                        result.Description,
				TaskSteps:                          []*mealplanning.RecipePrepTaskStep{},
				Optional:                           result.Optional,
			}

			logger.WithValue("storage_type", result.StorageType).Info("storage type")

			if result.StorageType.Valid {
				x.StorageType = string(result.StorageType.StorageContainerType)
			}
		}

		x.TaskSteps = append(x.TaskSteps, &mealplanning.RecipePrepTaskStep{
			ID:                      result.TaskStepID,
			BelongsToRecipeStep:     result.TaskStepBelongsToRecipeStep,
			BelongsToRecipePrepTask: result.TaskStepBelongsToRecipePrepTask,
			SatisfiesRecipeStep:     result.TaskStepSatisfiesRecipeStep,
		})
	}

	if x == nil || x.ID == "" {
		return nil, sql.ErrNoRows
	}

	return x, nil
}

// createRecipePrepTask creates a recipe prep task.
func (q *repository) createRecipePrepTask(ctx context.Context, querier database.SQLQueryExecutor, input *mealplanning.RecipePrepTaskDatabaseCreationInput) (*mealplanning.RecipePrepTask, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, input.ID)
	logger = logger.WithValue(mealplanningkeys.RecipePrepTaskIDKey, input.ID)

	// create the recipe prep task.
	if err := q.generatedQuerier.CreateRecipePrepTask(ctx, querier, &generated.CreateRecipePrepTaskParams{
		ID:                                     input.ID,
		Name:                                   input.Name,
		Description:                            input.Description,
		Notes:                                  input.Notes,
		ExplicitStorageInstructions:            input.ExplicitStorageInstructions,
		BelongsToRecipe:                        input.BelongsToRecipe,
		StorageType:                            generated.NullStorageContainerType{StorageContainerType: generated.StorageContainerType(input.StorageType), Valid: true},
		MinimumStorageTemperatureInCelsius:     database.NullStringFromFloat32Pointer(input.MinStorageTemperatureInCelsius),
		MaximumStorageTemperatureInCelsius:     database.NullStringFromFloat32Pointer(input.MaxStorageTemperatureInCelsius),
		MaximumTimeBufferBeforeRecipeInSeconds: database.NullInt32FromUint32Pointer(input.MaxTimeBufferBeforeRecipeInSeconds),
		MinimumTimeBufferBeforeRecipeInSeconds: int32(input.MinTimeBufferBeforeRecipeInSeconds),
		Optional:                               input.Optional,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating recipe prep task")
	}

	x := &mealplanning.RecipePrepTask{
		CreatedAt:                          q.CurrentTime(),
		ID:                                 input.ID,
		Name:                               input.Name,
		Description:                        input.Description,
		Notes:                              input.Notes,
		Optional:                           input.Optional,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
		StorageType:                        input.StorageType,
		BelongsToRecipe:                    input.BelongsToRecipe,
	}

	for _, recipePrepTaskStep := range input.TaskSteps {
		recipePrepTaskStep.BelongsToRecipePrepTask = input.ID
		s, err := q.createRecipePrepTaskStep(ctx, querier, recipePrepTaskStep)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "creating recipe prep task")
		}

		x.TaskSteps = append(x.TaskSteps, s)
	}

	logger.Info("recipe prep task created")

	return x, nil
}

// CreateRecipePrepTask creates a recipe prep task.
func (q *repository) CreateRecipePrepTask(ctx context.Context, input *mealplanning.RecipePrepTaskDatabaseCreationInput) (*mealplanning.RecipePrepTask, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipePrepTaskIDKey, input.ID)

	var x *mealplanning.RecipePrepTask
	if err := q.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		created, err := q.createRecipePrepTask(ctx, tx, input)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating recipe prep task")
		}

		x = created

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, mealplanning.RecipePrepTaskCreatedServiceEventType, "", map[string]any{
			mealplanningkeys.RecipeIDKey:         input.BelongsToRecipe,
			mealplanningkeys.RecipePrepTaskIDKey: input.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	logger.Info("recipe prep task created")

	return x, nil
}

// createRecipePrepTaskStep creates a recipe prep task step.
func (q *repository) createRecipePrepTaskStep(ctx context.Context, querier database.SQLQueryExecutor, input *mealplanning.RecipePrepTaskStepDatabaseCreationInput) (*mealplanning.RecipePrepTaskStep, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipePrepTaskIDKey, input.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipePrepTaskIDKey, input.BelongsToRecipePrepTask)

	// create the meal plan.
	if err := q.generatedQuerier.CreateRecipePrepTaskStep(ctx, querier, &generated.CreateRecipePrepTaskStepParams{
		ID:                      input.ID,
		BelongsToRecipePrepTask: input.BelongsToRecipePrepTask,
		BelongsToRecipeStep:     input.BelongsToRecipeStep,
		SatisfiesRecipeStep:     input.SatisfiesRecipeStep,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating recipe prep task step")
	}

	x := &mealplanning.RecipePrepTaskStep{
		ID:                      input.ID,
		BelongsToRecipeStep:     input.BelongsToRecipeStep,
		BelongsToRecipePrepTask: input.BelongsToRecipePrepTask,
		SatisfiesRecipeStep:     input.SatisfiesRecipeStep,
	}

	logger.Info("recipe prep task step created")

	return x, nil
}

// getRecipePrepTasksForRecipe gets a recipe prep task.
func (q *repository) getRecipePrepTasksForRecipe(ctx context.Context, recipeID string) ([]*mealplanning.RecipePrepTask, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	results, err := q.generatedQuerier.ListAllRecipePrepTasksByRecipe(ctx, q.readDB, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe prep tasks list retrieval query")
	}

	x := []*mealplanning.RecipePrepTask{}

	var currentRecipePrepTask *mealplanning.RecipePrepTask
	for _, result := range results {
		prepTaskStep := &mealplanning.RecipePrepTaskStep{
			ID:                      result.TaskStepID,
			BelongsToRecipeStep:     result.TaskStepBelongsToRecipeStep,
			BelongsToRecipePrepTask: result.TaskStepBelongsToRecipePrepTask,
			SatisfiesRecipeStep:     result.TaskStepSatisfiesRecipeStep,
		}

		if currentRecipePrepTask != nil && currentRecipePrepTask.ID != result.ID {
			x = append(x, currentRecipePrepTask)
			currentRecipePrepTask = nil
		}

		if currentRecipePrepTask == nil {
			currentRecipePrepTask = &mealplanning.RecipePrepTask{
				CreatedAt:                          result.CreatedAt,
				ArchivedAt:                         database.TimePointerFromNullTime(result.ArchivedAt),
				LastUpdatedAt:                      database.TimePointerFromNullTime(result.LastUpdatedAt),
				MinStorageTemperatureInCelsius:     database.Float32PointerFromNullString(result.MinimumStorageTemperatureInCelsius),
				MaxStorageTemperatureInCelsius:     database.Float32PointerFromNullString(result.MaximumStorageTemperatureInCelsius),
				MinTimeBufferBeforeRecipeInSeconds: uint32(result.MinimumTimeBufferBeforeRecipeInSeconds),
				MaxTimeBufferBeforeRecipeInSeconds: database.Uint32PointerFromNullInt32(result.MaximumTimeBufferBeforeRecipeInSeconds),
				ID:                                 result.ID,
				StorageType:                        string(result.StorageType.StorageContainerType),
				BelongsToRecipe:                    result.BelongsToRecipe,
				ExplicitStorageInstructions:        result.ExplicitStorageInstructions,
				Notes:                              result.Notes,
				Name:                               result.Name,
				Description:                        result.Description,
				TaskSteps:                          []*mealplanning.RecipePrepTaskStep{},
				Optional:                           result.Optional,
			}
		}
		currentRecipePrepTask.TaskSteps = append(currentRecipePrepTask.TaskSteps, prepTaskStep)
	}

	if currentRecipePrepTask != nil && currentRecipePrepTask.ID != "" {
		x = append(x, currentRecipePrepTask)
	}

	// Return nil instead of empty slice to match database behavior
	if len(x) == 0 {
		return nil, nil
	}

	return x, nil
}

// GetRecipePrepTasks gets recipe prep tasks with filtering support.
func (q *repository) GetRecipePrepTasks(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipePrepTask], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	tasks, err := q.getRecipePrepTasksForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe prep tasks")
	}

	x := filtering.NewQueryFilteredResult(
		tasks,
		uint64(len(tasks)),
		uint64(len(tasks)),
		func(rpt *mealplanning.RecipePrepTask) string { return rpt.ID },
		filter,
	)

	return x, nil
}

// GetRecipePrepTasksForRecipe gets a recipe prep task.
func (q *repository) GetRecipePrepTasksForRecipe(ctx context.Context, recipeID string) (x []*mealplanning.RecipePrepTask, err error) {
	return q.getRecipePrepTasksForRecipe(ctx, recipeID)
}

// UpdateRecipePrepTask updates a recipe prep task.
func (q *repository) UpdateRecipePrepTask(ctx context.Context, updated *mealplanning.RecipePrepTask) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()
	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipePrepTaskIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, mealplanning.RecipePrepTaskUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:         updated.BelongsToRecipe,
		mealplanningkeys.RecipePrepTaskIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		_, updateErr := q.generatedQuerier.UpdateRecipePrepTask(ctx, tx, &generated.UpdateRecipePrepTaskParams{
			Name:                                   updated.Name,
			Description:                            updated.Description,
			Notes:                                  updated.Notes,
			Optional:                               updated.Optional,
			ExplicitStorageInstructions:            updated.ExplicitStorageInstructions,
			MinimumTimeBufferBeforeRecipeInSeconds: int32(updated.MinTimeBufferBeforeRecipeInSeconds),
			MaximumTimeBufferBeforeRecipeInSeconds: database.NullInt32FromUint32Pointer(updated.MaxTimeBufferBeforeRecipeInSeconds),
			StorageType:                            generated.NullStorageContainerType{StorageContainerType: generated.StorageContainerType(updated.StorageType), Valid: true},
			MinimumStorageTemperatureInCelsius:     database.NullStringFromFloat32Pointer(updated.MinStorageTemperatureInCelsius),
			MaximumStorageTemperatureInCelsius:     database.NullStringFromFloat32Pointer(updated.MaxStorageTemperatureInCelsius),
			BelongsToRecipe:                        updated.BelongsToRecipe,
			ID:                                     updated.ID,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe prep task")
	}

	logger.Info("recipe prep task updated")

	return nil
}

// ArchiveRecipePrepTask marks a recipe prep task as archived.
func (q *repository) ArchiveRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if recipePrepTaskID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipePrepTaskIDKey, recipePrepTaskID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipePrepTaskIDKey, recipePrepTaskID)

	if err := q.withEvent(ctx, logger, mealplanning.RecipePrepTaskArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:         recipeID,
		mealplanningkeys.RecipePrepTaskIDKey: recipePrepTaskID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveRecipePrepTask(ctx, tx, recipePrepTaskID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "updating recipe prep task")
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
