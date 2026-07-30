package managers

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

func (m *mealPlanningManager) GetMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string) (*types.MealPlanRecipeOptionSelection, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanOptionIDKey: mealPlanOptionID,
		"recipe_step_id":                     recipeStepID,
		"ingredient_index":                   ingredientIndex,
		"selection_type":                     selectionType,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, "recipe_step_id", recipeStepID)
	tracing.AttachToSpan(span, "ingredient_index", ingredientIndex)
	tracing.AttachToSpan(span, "selection_type", selectionType)

	result, err := m.db.GetMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan recipe option selection")
	}

	return result, nil
}

func (m *mealPlanningManager) GetMealPlanRecipeOptionSelectionsForMealPlanOption(ctx context.Context, mealPlanOptionID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanRecipeOptionSelection], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := m.db.GetSelectionsForMealPlanOption(ctx, mealPlanOptionID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan recipe option selections")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID string, input *types.MealPlanRecipeOptionSelectionCreationRequestInput) (*types.MealPlanRecipeOptionSelection, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogError(err, m.logger.WithSpan(span), span, "validating meal plan recipe option selection creation input")
	}

	converted := converters.ConvertMealPlanRecipeOptionSelectionDatabaseCreationInputToMealPlanRecipeOptionSelectionDatabaseCreationInput(input, mealPlanOptionID)

	logger := m.logger.WithSpan(span).WithValue("meal_plan_recipe_option_selection_id", converted.ID)
	tracing.AttachToSpan(span, "meal_plan_recipe_option_selection_id", converted.ID)

	created, err := m.db.CreateMealPlanRecipeOptionSelection(ctx, converted)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating meal plan recipe option selection")
	}

	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return created, nil
}

func (m *mealPlanningManager) UpdateMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string, input *types.MealPlanRecipeOptionSelectionUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanOptionIDKey: mealPlanOptionID,
		"recipe_step_id":                     recipeStepID,
		"ingredient_index":                   ingredientIndex,
		"selection_type":                     selectionType,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, "recipe_step_id", recipeStepID)
	tracing.AttachToSpan(span, "ingredient_index", ingredientIndex)
	tracing.AttachToSpan(span, "selection_type", selectionType)

	existingSelection, err := m.db.GetMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching meal plan recipe option selection to update")
	}
	if existingSelection == nil {
		return fmt.Errorf("meal plan recipe option selection not found")
	}

	if err = m.db.UpdateMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType, input); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal plan recipe option selection")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.MealPlanRecipeOptionSelectionUpdatedServiceEventType, map[string]any{
		"meal_plan_recipe_option_selection_id": existingSelection.ID,
		mealplanningkeys.MealPlanOptionIDKey:   mealPlanOptionID,
	}))

	return nil
}

func (m *mealPlanningManager) ArchiveMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanOptionIDKey: mealPlanOptionID,
		"recipe_step_id":                     recipeStepID,
		"ingredient_index":                   ingredientIndex,
		"selection_type":                     selectionType,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, "recipe_step_id", recipeStepID)
	tracing.AttachToSpan(span, "ingredient_index", ingredientIndex)
	tracing.AttachToSpan(span, "selection_type", selectionType)

	if err := m.db.ArchiveMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving meal plan recipe option selection")
	}

	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return nil
}
