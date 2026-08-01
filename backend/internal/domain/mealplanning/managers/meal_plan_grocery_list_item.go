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

func (m *mealPlanningManager) ListMealPlanGroceryListItemsByMealPlan(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanGroceryListItem], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	results, err := m.db.GetMealPlanGroceryListItemsForMealPlan(ctx, mealPlanID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan grocery list items for meal plan")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateMealPlanGroceryListItem(ctx context.Context, input *types.MealPlanGroceryListItemCreationRequestInput) (*types.MealPlanGroceryListItem, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	convertedInput := converters.ConvertMealPlanGroceryListItemCreationRequestInputToMealPlanGroceryListItemDatabaseCreationInput(input)
	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.MealPlanGroceryListItemIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanGroceryListItemIDKey, convertedInput.ID)

	created, err := m.db.CreateMealPlanGroceryListItem(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating meal plan grocery list item")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadMealPlanGroceryListItem(ctx context.Context, mealPlanID, mealPlanGroceryListItemID string) (*types.MealPlanGroceryListItem, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:                mealPlanID,
		mealplanningkeys.MealPlanGroceryListItemIDKey: mealPlanGroceryListItemID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanGroceryListItemIDKey, mealPlanGroceryListItemID)

	result, err := m.db.GetMealPlanGroceryListItem(ctx, mealPlanID, mealPlanGroceryListItemID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan grocery list item")
	}

	return result, nil
}

func (m *mealPlanningManager) UpdateMealPlanGroceryListItem(ctx context.Context, mealPlanID, mealPlanGroceryListItemID string, input *types.MealPlanGroceryListItemUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:                mealPlanID,
		mealplanningkeys.MealPlanGroceryListItemIDKey: mealPlanGroceryListItemID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanGroceryListItemIDKey, mealPlanGroceryListItemID)

	existingMealPlanGroceryListItem, err := m.db.GetMealPlanGroceryListItem(ctx, mealPlanID, mealPlanGroceryListItemID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching meal plan grocery list item to update")
	}

	existingMealPlanGroceryListItem.Update(input)
	if err = m.db.UpdateMealPlanGroceryListItem(ctx, existingMealPlanGroceryListItem); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal plan grocery list item")
	}

	return nil
}

func (m *mealPlanningManager) ArchiveMealPlanGroceryListItem(ctx context.Context, mealPlanID, mealPlanGroceryListItemID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:                mealPlanID,
		mealplanningkeys.MealPlanGroceryListItemIDKey: mealPlanGroceryListItemID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanGroceryListItemIDKey, mealPlanGroceryListItemID)

	if err := m.db.ArchiveMealPlanGroceryListItem(ctx, mealPlanGroceryListItemID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving meal plan grocery list item")
	}

	return nil
}
