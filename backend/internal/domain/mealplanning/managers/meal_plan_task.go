package managers

import (
	"context"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

func (m *mealPlanningManager) ListMealPlanTasksByMealPlan(ctx context.Context, mealPlanID, ownerID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanTask], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if err := m.requireMealPlanAccess(ctx, mealPlanID, ownerID); err != nil {
		return nil, observability.PrepareError(err, span, "checking meal plan ownership")
	}

	results, err := m.db.GetMealPlanTasksForMealPlan(ctx, mealPlanID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting meal plan tasks for meal plan")
	}

	return results, nil
}

func (m *mealPlanningManager) ReadMealPlanTask(ctx context.Context, mealPlanID, mealPlanTaskID, ownerID string) (*types.MealPlanTask, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:     mealPlanID,
		mealplanningkeys.MealPlanTaskIDKey: mealPlanTaskID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanTaskIDKey, mealPlanTaskID)

	if err := m.requireMealPlanTaskAccess(ctx, mealPlanID, mealPlanTaskID, ownerID); err != nil {
		return nil, observability.PrepareError(err, span, "checking meal plan ownership")
	}

	result, err := m.db.GetMealPlanTask(ctx, mealPlanTaskID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan task")
	}

	return result, nil
}

func (m *mealPlanningManager) CreateMealPlanTask(ctx context.Context, mealPlanID, ownerID string, input *types.MealPlanTaskCreationRequestInput) (*types.MealPlanTask, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	// The plan the request names, and the option the task is for: a task names its option and
	// nothing above it, so checking only the plan would let an option from another account in.
	if err := m.requireMealPlanAccess(ctx, mealPlanID, ownerID); err != nil {
		return nil, observability.PrepareError(err, span, "checking meal plan ownership")
	}

	if err := m.requireMealPlanOptionAccess(ctx, input.MealPlanOptionID, ownerID); err != nil {
		return nil, observability.PrepareError(err, span, "checking meal plan option ownership")
	}

	convertedInput := converters.ConvertMealPlanTaskCreationRequestInputToMealPlanTaskDatabaseCreationInput(input)
	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.MealPlanTaskIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanTaskIDKey, convertedInput.ID)

	created, err := m.db.CreateMealPlanTask(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating meal plan task")
	}

	return created, nil
}

func (m *mealPlanningManager) MealPlanTaskStatusChange(ctx context.Context, mealPlanID, ownerID string, input *types.MealPlanTaskStatusChangeRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.MealPlanTaskIDKey, input.MealPlanTaskID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanTaskIDKey, input.MealPlanTaskID)

	if err := m.requireMealPlanTaskAccess(ctx, mealPlanID, input.MealPlanTaskID, ownerID); err != nil {
		return observability.PrepareError(err, span, "checking meal plan ownership")
	}

	if err := m.db.ChangeMealPlanTaskStatus(ctx, input); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "changing meal plan task status")
	}

	return nil
}
