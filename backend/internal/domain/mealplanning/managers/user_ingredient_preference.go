package managers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v6/errors"
	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/observability"
	"github.com/primandproper/platform-go/v6/observability/tracing"
)

func (m *mealPlanningManager) ListUserIngredientPreferences(ctx context.Context, ownerID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.UserIngredientPreference], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)

	if ownerID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}
	logger := m.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, ownerID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, ownerID)

	results, err := m.db.GetUserIngredientPreferences(ctx, ownerID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching ingredient preferences")
	}

	return results, nil
}

func (m *mealPlanningManager) ReadUserIngredientPreference(ctx context.Context, ownerID, ingredientPreferenceID string) (*types.UserIngredientPreference, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, ownerID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, ownerID)

	result, err := m.db.GetUserIngredientPreference(ctx, ingredientPreferenceID, ownerID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching ingredient preferences")
	}

	return result, nil
}

func (m *mealPlanningManager) CreateUserIngredientPreference(ctx context.Context, ownerID string, input *types.UserIngredientPreferenceCreationRequestInput) ([]*types.UserIngredientPreference, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if ownerID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.ValidIngredientGroupIDKey: input.ValidIngredientGroupID,
		mealplanningkeys.ValidIngredientIDKey:      input.ValidIngredientID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, input.ValidIngredientGroupID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientIDKey, input.ValidIngredientID)

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "validating ingredient preference creation request input")
	}

	convertedInput := converters.ConvertUserIngredientPreferenceCreationRequestInputToUserIngredientPreferenceDatabaseCreationInput(input)
	convertedInput.CreatedByUser = ownerID

	created, err := m.db.CreateUserIngredientPreference(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating ingredient preference")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.UserIngredientPreferenceCreatedServiceEventType, map[string]any{
		mealplanningkeys.ValidIngredientGroupIDKey: input.ValidIngredientGroupID,
		mealplanningkeys.ValidIngredientIDKey:      input.ValidIngredientID,
		"created":                                  len(created),
	}))

	return created, nil
}

func (m *mealPlanningManager) UpdateUserIngredientPreference(ctx context.Context, ingredientPreferenceID, ownerID string, input *types.UserIngredientPreferenceUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.UserIngredientPreferenceIDKey: ingredientPreferenceID,
		identitykeys.UserIDKey:                         ownerID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.UserIngredientPreferenceIDKey, ingredientPreferenceID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, ownerID)

	existingUserIngredientPreference, err := m.db.GetUserIngredientPreference(ctx, ingredientPreferenceID, ownerID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching UserIngredientPreference to update")
	}

	existingUserIngredientPreference.Update(input)
	if err = m.db.UpdateUserIngredientPreference(ctx, existingUserIngredientPreference); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating ingredient preference")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.UserIngredientPreferenceUpdatedServiceEventType, map[string]any{
		mealplanningkeys.UserIngredientPreferenceIDKey: ingredientPreferenceID,
	}))

	return nil
}

func (m *mealPlanningManager) ArchiveUserIngredientPreference(ctx context.Context, ownerID, ingredientPreferenceID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.UserIngredientPreferenceIDKey: ingredientPreferenceID,
		identitykeys.UserIDKey:                         ownerID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.UserIngredientPreferenceIDKey, ingredientPreferenceID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, ownerID)

	if err := m.db.ArchiveUserIngredientPreference(ctx, ingredientPreferenceID, ownerID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving ingredient preference")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.UserIngredientPreferenceArchivedServiceEventType, map[string]any{
		mealplanningkeys.UserIngredientPreferenceIDKey: ingredientPreferenceID,
	}))

	return nil
}
