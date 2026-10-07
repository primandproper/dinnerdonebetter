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

func (m *mealPlanningManager) ListMealPlanOptions(ctx context.Context, mealPlanID, mealPlanEventID, ownerID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanOption], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:      mealPlanID,
		mealplanningkeys.MealPlanEventIDKey: mealPlanEventID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)

	if err := m.requireMealPlanAccess(ctx, mealPlanID, ownerID); err != nil {
		return nil, observability.PrepareError(err, span, "checking meal plan ownership")
	}

	results, err := m.db.GetMealPlanOptions(ctx, mealPlanID, mealPlanEventID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan options")
	}

	return results, nil
}

// createSelectionsForNewOption persists the recipe-option selections provided inline on an option creation input.
func (m *mealPlanningManager) createSelectionsForNewOption(ctx context.Context, mealPlanOptionID string, selections []*types.MealPlanRecipeOptionSelectionCreationRequestInput) error {
	for _, selection := range selections {
		if selection == nil {
			continue
		}

		if err := selection.ValidateWithContext(ctx); err != nil {
			return observability.PrepareError(err, nil, "validating inline selection")
		}

		converted := converters.ConvertMealPlanRecipeOptionSelectionDatabaseCreationInputToMealPlanRecipeOptionSelectionDatabaseCreationInput(selection, mealPlanOptionID)
		if _, err := m.db.CreateMealPlanRecipeOptionSelection(ctx, converted); err != nil {
			return err
		}
	}

	return nil
}

func (m *mealPlanningManager) CreateMealPlanOptionWithEventID(ctx context.Context, mealPlanID, mealPlanEventID, ownerID string, input *types.MealPlanOptionCreationRequestInput) (*types.MealPlanOption, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	if mealPlanEventID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	if err := m.requireMealPlanEventAccess(ctx, mealPlanID, mealPlanEventID, ownerID); err != nil {
		return nil, observability.PrepareError(err, span, "checking meal plan ownership")
	}

	convertedInput := converters.ConvertMealPlanOptionCreationRequestInputToMealPlanOptionDatabaseCreationInput(input)

	convertedInput.BelongsToMealPlanEvent = mealPlanEventID

	exists, err := m.db.MealExistsAsOptionInEvent(ctx, mealPlanEventID, convertedInput.MealID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, m.logger.WithSpan(span), span, "checking if meal exists as option in event")
	}
	if exists {
		return nil, types.ErrDuplicateMealPlanOption
	}

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.MealPlanOptionIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, convertedInput.ID)

	created, err := m.db.CreateMealPlanOption(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "created meal plan option")
	}

	if err = m.createSelectionsForNewOption(ctx, created.ID, input.Selections); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating selections for meal plan option")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, ownerID string) (*types.MealPlanOption, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:       mealPlanID,
		mealplanningkeys.MealPlanEventIDKey:  mealPlanEventID,
		mealplanningkeys.MealPlanOptionIDKey: mealPlanOptionID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	if err := m.requireMealPlanAccess(ctx, mealPlanID, ownerID); err != nil {
		return nil, observability.PrepareError(err, span, "checking meal plan ownership")
	}

	mealPlanOption, err := m.db.GetMealPlanOption(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan option")
	}

	return mealPlanOption, nil
}

func (m *mealPlanningManager) UpdateMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, ownerID string, input *types.MealPlanOptionUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "validating input")
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:       mealPlanID,
		mealplanningkeys.MealPlanEventIDKey:  mealPlanEventID,
		mealplanningkeys.MealPlanOptionIDKey: mealPlanOptionID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	if err := m.requireMealPlanAccess(ctx, mealPlanID, ownerID); err != nil {
		return observability.PrepareError(err, span, "checking meal plan ownership")
	}

	existingMealPlanOption, err := m.db.GetMealPlanOption(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching meal plan option to update")
	}

	existingMealPlanOption.Update(input)
	if err = m.db.UpdateMealPlanOption(ctx, existingMealPlanOption); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal plan option")
	}

	return nil
}

func (m *mealPlanningManager) ArchiveMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, ownerID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:       mealPlanID,
		mealplanningkeys.MealPlanEventIDKey:  mealPlanEventID,
		mealplanningkeys.MealPlanOptionIDKey: mealPlanOptionID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	if err := m.requireMealPlanEventAccess(ctx, mealPlanID, mealPlanEventID, ownerID); err != nil {
		return observability.PrepareError(err, span, "checking meal plan ownership")
	}

	if err := m.db.ArchiveMealPlanOption(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving meal plan option")
	}

	return nil
}
