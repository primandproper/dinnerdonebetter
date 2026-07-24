package managers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v6/errors"
	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/observability"
	"github.com/primandproper/platform-go/v6/observability/tracing"
)

func (m *mealPlanningManager) ListMealPlanEvents(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanEvent], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	mealPlanEvents, err := m.db.GetMealPlanEvents(ctx, mealPlanID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan events")
	}

	return mealPlanEvents, nil
}

func (m *mealPlanningManager) CreateMealPlanEvent(ctx context.Context, mealPlanID string, input *types.MealPlanEventCreationRequestInput) (*types.MealPlanEvent, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertMealPlanEventCreationRequestInputToMealPlanEventDatabaseCreationInput(input)
	convertedInput.BelongsToMealPlan = mealPlanID
	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.MealPlanEventIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, convertedInput.ID)

	created, err := m.db.CreateMealPlanEvent(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "created meal plan event")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.MealPlanEventCreatedServiceEventType, map[string]any{
		mealplanningkeys.MealPlanEventIDKey: created.ID,
	}))

	return created, nil
}

func (m *mealPlanningManager) ReadMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string) (*types.MealPlanEvent, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:      mealPlanID,
		mealplanningkeys.MealPlanEventIDKey: mealPlanEventID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)

	mealPlanEvent, err := m.db.GetMealPlanEvent(ctx, mealPlanID, mealPlanEventID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan event")
	}

	return mealPlanEvent, nil
}

func (m *mealPlanningManager) UpdateMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string, input *types.MealPlanEventUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "validating input")
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:      mealPlanID,
		mealplanningkeys.MealPlanEventIDKey: mealPlanEventID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)

	existingMealPlanEvent, err := m.db.GetMealPlanEvent(ctx, mealPlanID, mealPlanEventID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching meal plan event to update")
	}

	oldStartsAt := existingMealPlanEvent.StartsAt
	existingMealPlanEvent.Update(input)
	if err = m.db.UpdateMealPlanEvent(ctx, existingMealPlanEvent); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal plan event")
	}

	if input.StartsAt != nil && !oldStartsAt.Equal(*input.StartsAt) {
		if err = m.db.ClearMealPlanTaskNotificationSentForEvent(ctx, mealPlanEventID); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "clearing meal plan task notification sent for event")
		}
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.MealPlanEventUpdatedServiceEventType, map[string]any{
		mealplanningkeys.MealPlanIDKey:      mealPlanID,
		mealplanningkeys.MealPlanEventIDKey: mealPlanEventID,
	}))

	return nil
}

func (m *mealPlanningManager) SwapMealPlanEvents(ctx context.Context, mealPlanID, mealPlanEventIDA, mealPlanEventIDB string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:      mealPlanID,
		mealplanningkeys.MealPlanEventIDKey: mealPlanEventIDA + "," + mealPlanEventIDB,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventIDA+","+mealPlanEventIDB)

	if err := m.db.SwapMealPlanEvents(ctx, mealPlanID, mealPlanEventIDA, mealPlanEventIDB); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "swapping meal plan events")
	}

	if err := m.db.ClearMealPlanTaskNotificationSentForEvent(ctx, mealPlanEventIDA); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "clearing meal plan task notification sent for event A")
	}
	if err := m.db.ClearMealPlanTaskNotificationSentForEvent(ctx, mealPlanEventIDB); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "clearing meal plan task notification sent for event B")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.MealPlanEventUpdatedServiceEventType, map[string]any{
		mealplanningkeys.MealPlanIDKey: mealPlanID,
	}))

	return nil
}

func (m *mealPlanningManager) ArchiveMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:      mealPlanID,
		mealplanningkeys.MealPlanEventIDKey: mealPlanEventID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)

	if err := m.db.ArchiveMealPlanEvent(ctx, mealPlanID, mealPlanEventID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving meal plan event")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.MealPlanEventArchivedServiceEventType, map[string]any{
		mealplanningkeys.MealPlanIDKey:      mealPlanID,
		mealplanningkeys.MealPlanEventIDKey: mealPlanEventID,
	}))

	return nil
}
