package managers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v5/errors"
	"github.com/primandproper/platform-go/v5/filtering"
	"github.com/primandproper/platform-go/v5/observability"
	"github.com/primandproper/platform-go/v5/observability/tracing"
)

func (m *mealPlanningManager) ListMealPlanOptions(ctx context.Context, mealPlanID, mealPlanEventID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanOption], error) {
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

	results, err := m.db.GetMealPlanOptions(ctx, mealPlanID, mealPlanEventID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan options")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateMealPlanOption(ctx context.Context, input *types.MealPlanOptionCreationRequestInput) (*types.MealPlanOption, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	convertedInput := converters.ConvertMealPlanOptionCreationRequestInputToMealPlanOptionDatabaseCreationInput(input)
	if convertedInput.BelongsToMealPlanEvent != "" {
		exists, err := m.db.MealExistsAsOptionInEvent(ctx, convertedInput.BelongsToMealPlanEvent, convertedInput.MealID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, m.logger.WithSpan(span), span, "checking if meal exists as option in event")
		}
		if exists {
			return nil, types.ErrDuplicateMealPlanOption
		}
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

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.MealPlanOptionCreatedServiceEventType, map[string]any{
		mealplanningkeys.MealPlanOptionIDKey: convertedInput.ID,
	}))

	return created, nil
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

func (m *mealPlanningManager) CreateMealPlanOptionWithEventID(ctx context.Context, mealPlanEventID string, input *types.MealPlanOptionCreationRequestInput) (*types.MealPlanOption, error) {
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

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.MealPlanOptionCreatedServiceEventType, map[string]any{
		mealplanningkeys.MealPlanOptionIDKey: convertedInput.ID,
	}))

	return created, nil
}

func (m *mealPlanningManager) ReadMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) (*types.MealPlanOption, error) {
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

	mealPlanOption, err := m.db.GetMealPlanOption(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan option")
	}

	return mealPlanOption, nil
}

func (m *mealPlanningManager) MealPlanOptionBelongsToAccount(ctx context.Context, mealPlanOptionID, accountID string) (bool, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanOptionIDKey: mealPlanOptionID,
		identitykeys.AccountIDKey:            accountID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	belongs, err := m.db.MealPlanOptionBelongsToAccount(ctx, mealPlanOptionID, accountID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "checking meal plan option account ownership")
	}

	return belongs, nil
}

func (m *mealPlanningManager) UpdateMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string, input *types.MealPlanOptionUpdateRequestInput) error {
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

	existingMealPlanOption, err := m.db.GetMealPlanOption(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching meal plan option to update")
	}

	existingMealPlanOption.Update(input)
	if err = m.db.UpdateMealPlanOption(ctx, existingMealPlanOption); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal plan option")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.MealPlanOptionUpdatedServiceEventType, map[string]any{
		mealplanningkeys.MealPlanIDKey:       mealPlanID,
		mealplanningkeys.MealPlanEventIDKey:  mealPlanEventID,
		mealplanningkeys.MealPlanOptionIDKey: mealPlanOptionID,
	}))

	return nil
}

func (m *mealPlanningManager) ArchiveMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) error {
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

	if err := m.db.ArchiveMealPlanOption(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving meal plan option")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.MealPlanOptionArchivedServiceEventType, map[string]any{
		mealplanningkeys.MealPlanIDKey:       mealPlanID,
		mealplanningkeys.MealPlanEventIDKey:  mealPlanEventID,
		mealplanningkeys.MealPlanOptionIDKey: mealPlanOptionID,
	}))

	return nil
}
