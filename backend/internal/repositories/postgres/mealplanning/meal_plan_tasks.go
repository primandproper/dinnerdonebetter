package mealplanning

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v6/database"
	platformerrors "github.com/primandproper/platform-go/v6/errors"
	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/observability"
	"github.com/primandproper/platform-go/v6/observability/tracing"
)

var (
	_ types.MealPlanTaskDataManager = (*repository)(nil)
)

// MealPlanTaskExists checks if a meal plan task exists.
func (q *repository) MealPlanTaskExists(ctx context.Context, mealPlanID, mealPlanTaskID string) (bool, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if mealPlanTaskID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanTaskIDKey, mealPlanTaskID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanTaskIDKey, mealPlanTaskID)

	result, err := q.generatedQuerier.CheckMealPlanTaskExistence(ctx, q.readDB, &generated.CheckMealPlanTaskExistenceParams{
		MealPlanID:     mealPlanID,
		MealPlanTaskID: mealPlanTaskID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing meal plan task existence check")
	}

	logger.Info("meal plan task existence retrieved")

	return result, nil
}

// GetMealPlanTask fetches a meal plan task.
func (q *repository) GetMealPlanTask(ctx context.Context, mealPlanTaskID string) (*types.MealPlanTask, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanTaskID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanTaskIDKey, mealPlanTaskID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanTaskIDKey, mealPlanTaskID)

	result, err := q.generatedQuerier.GetMealPlanTask(ctx, q.readDB, mealPlanTaskID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting meal plan task")
	}

	mealPlanTask := &types.MealPlanTask{
		RecipePrepTask: types.RecipePrepTask{
			ID:   database.StringFromNullString(result.PrepTaskID),
			Name: database.StringFromNullString(result.PrepTaskName),
		},
		CreatedAt:           result.CreatedAt,
		LastUpdatedAt:       database.TimePointerFromNullTime(result.LastUpdatedAt),
		CompletedAt:         database.TimePointerFromNullTime(result.CompletedAt),
		NotificationSentAt:  database.TimePointerFromNullTime(result.NotificationSentAt),
		AssignedToUser:      database.StringPointerFromNullString(result.AssignedToUser),
		ID:                  result.ID,
		Status:              string(result.Status),
		CreationExplanation: result.CreationExplanation,
		StatusExplanation:   result.StatusExplanation,
		MealPlanOption: types.MealPlanOption{
			CreatedAt:              result.MealPlanOptionCreatedAt,
			LastUpdatedAt:          database.TimePointerFromNullTime(result.MealPlanOptionLastUpdatedAt),
			AssignedCook:           database.StringPointerFromNullString(result.MealPlanOptionAssignedCook),
			ArchivedAt:             database.TimePointerFromNullTime(result.MealPlanOptionArchivedAt),
			AssignedDishwasher:     database.StringPointerFromNullString(result.MealPlanOptionAssignedDishwasher),
			Notes:                  result.MealPlanOptionNotes,
			BelongsToMealPlanEvent: database.StringFromNullString(result.MealPlanOptionBelongsToMealPlanEvent),
			ID:                     result.MealPlanOptionID,
			Votes:                  nil,
			Meal: types.Meal{
				ID: result.MealPlanOptionMealID,
			},
			MealScale: database.Float32FromString(result.MealPlanOptionMealScale),
			Chosen:    result.MealPlanOptionChosen,
			TieBroken: result.MealPlanOptionTiebroken,
		},
	}

	return mealPlanTask, nil
}

// createMealPlanTask creates a meal plan task.
func (q *repository) createMealPlanTask(ctx context.Context, querier database.SQLQueryExecutorAndTransactionManager, input *types.MealPlanTaskDatabaseCreationInput) (*types.MealPlanTask, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanTaskIDKey, input.ID)

	// create the meal plan task. An empty RecipePrepTaskID means an ad-hoc task (e.g. a thaw
	// task) with no backing recipe prep task, persisted as NULL.
	if err := q.generatedQuerier.CreateMealPlanTask(ctx, querier, &generated.CreateMealPlanTaskParams{
		ID:                      input.ID,
		Status:                  types.MealPlanTaskStatusUnfinished,
		StatusExplanation:       input.StatusExplanation,
		CreationExplanation:     input.CreationExplanation,
		BelongsToMealPlanOption: input.MealPlanOptionID,
		BelongsToRecipePrepTask: sql.NullString{String: input.RecipePrepTaskID, Valid: input.RecipePrepTaskID != ""},
		AssignedToUser:          database.NullStringFromStringPointer(input.AssignedToUser),
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating meal plan task")
	}

	x := &types.MealPlanTask{
		CreatedAt:           q.CurrentTime(),
		ID:                  input.ID,
		AssignedToUser:      input.AssignedToUser,
		Status:              types.MealPlanTaskStatusUnfinished,
		CreationExplanation: input.CreationExplanation,
		StatusExplanation:   input.StatusExplanation,
		MealPlanOption: types.MealPlanOption{
			ID: input.MealPlanOptionID,
		},
		RecipePrepTask: types.RecipePrepTask{
			ID: input.RecipePrepTaskID,
		},
	}

	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, x.ID)
	logger.Info("meal plan task created")

	return x, nil
}

// CreateMealPlanTask creates a meal plan task.
func (q *repository) CreateMealPlanTask(ctx context.Context, input *types.MealPlanTaskDatabaseCreationInput) (*types.MealPlanTask, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanTaskIDKey, input.ID)

	var x *types.MealPlanTask
	if err := q.WithTransaction(ctx, func(tx database.SQLQueryExecutorAndTransactionManager) error {
		created, err := q.createMealPlanTask(ctx, tx, input)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating meal plan task")
		}

		x = created

		return nil
	}); err != nil {
		return nil, err
	}

	logger.Info("meal plan task created")

	return x, nil
}

// GetMealPlanTasksForMealPlan fetches a list of meal plan tasks.
func (q *repository) GetMealPlanTasksForMealPlan(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanTask], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	results, err := q.generatedQuerier.ListAllMealPlanTasksByMealPlan(ctx, q.readDB, mealPlanID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing meal plan tasks list retrieval query")
	}

	// Group results by task ID (since each task can have multiple rows - one per prep task step)
	taskMap := make(map[string]*types.MealPlanTask)

	for _, result := range results {
		var mealPlanTask *types.MealPlanTask
		var exists bool

		if mealPlanTask, exists = taskMap[result.ID]; !exists {
			// Create new task
			mealPlanTask = &types.MealPlanTask{
				CreatedAt:           result.CreatedAt,
				LastUpdatedAt:       database.TimePointerFromNullTime(result.LastUpdatedAt),
				CompletedAt:         database.TimePointerFromNullTime(result.CompletedAt),
				NotificationSentAt:  database.TimePointerFromNullTime(result.NotificationSentAt),
				AssignedToUser:      database.StringPointerFromNullString(result.AssignedToUser),
				ID:                  result.ID,
				Status:              string(result.Status),
				CreationExplanation: result.CreationExplanation,
				StatusExplanation:   result.StatusExplanation,
				MealPlanOption: types.MealPlanOption{
					CreatedAt:              result.MealPlanOptionCreatedAt,
					LastUpdatedAt:          database.TimePointerFromNullTime(result.MealPlanOptionLastUpdatedAt),
					AssignedCook:           database.StringPointerFromNullString(result.MealPlanOptionAssignedCook),
					ArchivedAt:             database.TimePointerFromNullTime(result.MealPlanOptionArchivedAt),
					AssignedDishwasher:     database.StringPointerFromNullString(result.MealPlanOptionAssignedDishwasher),
					Notes:                  result.MealPlanOptionNotes,
					BelongsToMealPlanEvent: database.StringFromNullString(result.MealPlanOptionBelongsToMealPlanEvent),
					ID:                     result.MealPlanOptionID,
					Votes:                  nil,
					Meal: types.Meal{
						ID: result.MealPlanOptionMealID,
					},
					MealScale: database.Float32FromString(result.MealPlanOptionMealScale),
					Chosen:    result.MealPlanOptionChosen,
					TieBroken: result.MealPlanOptionTiebroken,
				},
				RecipePrepTask: types.RecipePrepTask{
					ID:                                 database.StringFromNullString(result.PrepTaskID),
					BelongsToRecipe:                    database.StringFromNullString(result.PrepTaskBelongsToRecipe),
					Name:                               database.StringFromNullString(result.PrepTaskName),
					Description:                        database.StringFromNullString(result.PrepTaskDescription),
					Notes:                              database.StringFromNullString(result.PrepTaskNotes),
					ExplicitStorageInstructions:        database.StringFromNullString(result.PrepTaskExplicitStorageInstructions),
					Optional:                           database.BoolFromNullBool(result.PrepTaskOptional),
					CreatedAt:                          database.TimeFromNullTime(result.PrepTaskCreatedAt),
					LastUpdatedAt:                      database.TimePointerFromNullTime(result.PrepTaskLastUpdatedAt),
					ArchivedAt:                         database.TimePointerFromNullTime(result.PrepTaskArchivedAt),
					MinStorageTemperatureInCelsius:     database.Float32PointerFromNullString(result.PrepTaskMinimumStorageTemperatureInCelsius),
					MaxStorageTemperatureInCelsius:     database.Float32PointerFromNullString(result.PrepTaskMaximumStorageTemperatureInCelsius),
					MinTimeBufferBeforeRecipeInSeconds: uint32(result.PrepTaskMinimumTimeBufferBeforeRecipeInSeconds.Int32),
					MaxTimeBufferBeforeRecipeInSeconds: database.Uint32PointerFromNullInt32(result.PrepTaskMaximumTimeBufferBeforeRecipeInSeconds),
					TaskSteps:                          []*types.RecipePrepTaskStep{},
				},
			}

			// Set storage type if available
			if result.PrepTaskStorageType.Valid {
				mealPlanTask.RecipePrepTask.StorageType = string(result.PrepTaskStorageType.StorageContainerType)
			}

			taskMap[result.ID] = mealPlanTask
		}

		// Add prep task step if it exists and hasn't been added yet
		if prepTaskStepID := database.StringFromNullString(result.PrepTaskStepID); prepTaskStepID != "" {
			// Check if this step is already in the task steps
			stepExists := false
			for _, existingStep := range mealPlanTask.RecipePrepTask.TaskSteps {
				if existingStep.ID == prepTaskStepID {
					stepExists = true
					break
				}
			}

			if !stepExists {
				mealPlanTask.RecipePrepTask.TaskSteps = append(mealPlanTask.RecipePrepTask.TaskSteps, &types.RecipePrepTaskStep{
					ID:                      prepTaskStepID,
					BelongsToRecipeStep:     database.StringFromNullString(result.PrepTaskStepBelongsToRecipeStep),
					BelongsToRecipePrepTask: database.StringFromNullString(result.PrepTaskStepBelongsToRecipePrepTask),
					SatisfiesRecipeStep:     database.BoolFromNullBool(result.PrepTaskStepSatisfiesRecipeStep),
				})
			}
		}
	}

	// Convert map to slice
	data := make([]*types.MealPlanTask, 0, len(taskMap))
	for _, task := range taskMap {
		data = append(data, task)
	}

	var filteredCount, totalCount uint64

	x := filtering.NewQueryFilteredResult(data, filteredCount, totalCount, func(t *types.MealPlanTask) string {
		return t.ID
	}, filter)

	return x, nil
}

// CreateMealPlanTasksForMealPlanOption creates meal plan tasks.
func (q *repository) CreateMealPlanTasksForMealPlanOption(ctx context.Context, inputs []*types.MealPlanTaskDatabaseCreationInput) ([]*types.MealPlanTask, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	outputs := []*types.MealPlanTask{}
	if err := q.WithTransaction(ctx, func(tx database.SQLQueryExecutorAndTransactionManager) error {
		for _, input := range inputs {
			mealPlanTask, createMealPlanTaskErr := q.createMealPlanTask(ctx, tx, input)
			if createMealPlanTaskErr != nil {
				return observability.PrepareAndLogError(createMealPlanTaskErr, logger, span, "creating meal plan task")
			}

			outputs = append(outputs, mealPlanTask)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	logger.Info("meal plan tasks created")

	return outputs, nil
}

// MarkMealPlanAsHavingTasksCreated marks a meal plan as having all its tasks created.
func (q *repository) MarkMealPlanAsHavingTasksCreated(ctx context.Context, mealPlanID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if err := q.generatedQuerier.MarkMealPlanAsPrepTasksCreated(ctx, q.writeDB, mealPlanID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking meal plan as having tasks created")
	}

	return nil
}

// MarkMealPlanAsHavingGroceryListInitialized marks a meal plan as having all its tasks created.
func (q *repository) MarkMealPlanAsHavingGroceryListInitialized(ctx context.Context, mealPlanID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if err := q.generatedQuerier.MarkMealPlanAsGroceryListInitialized(ctx, q.writeDB, mealPlanID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking meal plan as having tasks created")
	}

	logger.Info("meal plan marked as grocery list initialized")

	return nil
}

// ChangeMealPlanTaskStatus changes a meal plan task's status.
func (q *repository) ChangeMealPlanTaskStatus(ctx context.Context, input *types.MealPlanTaskStatusChangeRequestInput) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if input == nil {
		return platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanTaskIDKey, input.MealPlanTaskID)
	logger = logger.WithValue(mealplanningkeys.MealPlanTaskIDKey, input.MealPlanTaskID)

	var settledAt *time.Time
	if input.Status != nil && *input.Status == types.MealPlanTaskStatusFinished {
		settledAt = new(q.CurrentTime())
	}

	var newStatus string
	if input.Status != nil {
		newStatus = *input.Status
	}

	if err := q.generatedQuerier.ChangeMealPlanTaskStatus(ctx, q.writeDB, &generated.ChangeMealPlanTaskStatusParams{
		ID:                input.MealPlanTaskID,
		Status:            generated.PrepStepStatus(newStatus),
		StatusExplanation: input.StatusExplanation,
		CompletedAt:       database.NullTimeFromTimePointer(settledAt),
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "changing meal plan task status")
	}

	logger.Info("meal plan task status changed")

	return nil
}

// MealPlanTaskNotificationHasBeenSent checks if a push notification has already been sent for a meal plan task.
func (q *repository) MealPlanTaskNotificationHasBeenSent(ctx context.Context, mealPlanTaskID string) (bool, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if mealPlanTaskID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}

	result, err := q.generatedQuerier.MealPlanTaskNotificationHasBeenSent(ctx, q.readDB, mealPlanTaskID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, q.logger.Clone(), span, "checking meal plan task notification status")
	}

	sent, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type from MealPlanTaskNotificationHasBeenSent: %T", result)
	}
	return sent, nil
}

// MarkMealPlanTaskNotificationSent records that a push notification was sent for a meal plan task.
func (q *repository) MarkMealPlanTaskNotificationSent(ctx context.Context, mealPlanTaskID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if mealPlanTaskID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	if err := q.generatedQuerier.MarkMealPlanTaskNotificationSent(ctx, q.writeDB, mealPlanTaskID); err != nil {
		return observability.PrepareAndLogError(err, q.logger.Clone(), span, "marking meal plan task notification sent")
	}

	return nil
}

// ClearMealPlanTaskNotificationSentForEvent clears notification_sent_at for all incomplete meal plan tasks
// belonging to options in the given event. Used when an event's start time changes or events are swapped.
func (q *repository) ClearMealPlanTaskNotificationSentForEvent(ctx context.Context, mealPlanEventID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if mealPlanEventID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	if err := q.generatedQuerier.ClearMealPlanTaskNotificationSentForEvent(ctx, q.writeDB, database.NullStringFromString(mealPlanEventID)); err != nil {
		return observability.PrepareAndLogError(err, q.logger.Clone(), span, "clearing meal plan task notification sent for event")
	}

	return nil
}

// GetMealPlanTaskIDsThatNeedNotification returns meal plan task IDs that need a push notification sent.
func (q *repository) GetMealPlanTaskIDsThatNeedNotification(ctx context.Context) ([]string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	results, err := q.generatedQuerier.GetMealPlanTaskIDsThatNeedNotification(ctx, q.readDB)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, q.logger.Clone(), span, "getting meal plan task IDs that need notification")
	}

	return results, nil
}

// GetMealPlanTaskAccountID returns the account ID (household) for a meal plan task.
func (q *repository) GetMealPlanTaskAccountID(ctx context.Context, mealPlanTaskID string) (string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if mealPlanTaskID == "" {
		return "", platformerrors.ErrInvalidIDProvided
	}

	accountID, err := q.generatedQuerier.GetMealPlanTaskAccountID(ctx, q.readDB, mealPlanTaskID)
	if err != nil {
		return "", observability.PrepareAndLogError(err, q.logger.Clone(), span, "getting meal plan task account ID")
	}

	return accountID, nil
}

// GetMealPlanTaskNotificationContext returns context for building a meal plan task notification.
func (q *repository) GetMealPlanTaskNotificationContext(ctx context.Context, mealPlanTaskID string) (*types.MealPlanTaskNotificationContext, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if mealPlanTaskID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	row, err := q.generatedQuerier.GetMealPlanTaskNotificationContext(ctx, q.readDB, mealPlanTaskID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, q.logger.Clone(), span, "getting meal plan task notification context")
	}

	return &types.MealPlanTaskNotificationContext{
		PrepTaskName:        row.PrepTaskName,
		CreationExplanation: row.CreationExplanation,
		MealName:            string(row.MealName),
		StartsAt:            row.StartsAt,
	}, nil
}
