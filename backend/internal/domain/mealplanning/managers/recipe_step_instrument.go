package managers

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v6/errors"
	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/observability"
	"github.com/primandproper/platform-go/v6/observability/tracing"
)

func (m *mealPlanningManager) ListRecipeStepInstruments(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepInstrument], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     recipeID,
		mealplanningkeys.RecipeStepIDKey: recipeStepID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	logger = filter.AttachToLogger(logger)

	results, err := m.db.GetRecipeStepInstruments(ctx, recipeID, recipeStepID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching list of recipe step instruments")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID string, input *types.RecipeStepInstrumentCreationRequestInput) (*types.RecipeStepInstrument, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     recipeID,
		mealplanningkeys.RecipeStepIDKey: recipeStepID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if input.Index == nil {
		return nil, fmt.Errorf("index is required when creating a recipe step instrument outside of initial recipe creation")
	}

	convertedInput := converters.ConvertRecipeStepInstrumentCreationRequestInputToRecipeStepInstrumentDatabaseCreationInput(input, 0)
	convertedInput.BelongsToRecipeStep = recipeStepID
	logger = logger.WithValue(mealplanningkeys.RecipeStepInstrumentIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepInstrumentIDKey, convertedInput.ID)

	if convertedInput.ValidPreparationInstrumentID != nil && *convertedInput.ValidPreparationInstrumentID != "" {
		vpi, err := m.db.GetValidPreparationInstrument(ctx, *convertedInput.ValidPreparationInstrumentID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation instrument")
		}
		convertedInput.InstrumentID = &vpi.Instrument.ID
	}

	created, err := m.db.CreateRecipeStepInstrument(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating recipe step instrument")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipeStepInstrumentCreatedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepInstrumentIDKey: convertedInput.ID,
	}))

	return created, nil
}

func (m *mealPlanningManager) ReadRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string) (*types.RecipeStepInstrument, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepInstrumentIDKey: recipeStepInstrumentID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepInstrumentIDKey, recipeStepInstrumentID)

	x, err := m.db.GetRecipeStepInstrument(ctx, recipeID, recipeStepID, recipeStepInstrumentID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "retrieving recipe step instrument")
	}

	return x, nil
}

func (m *mealPlanningManager) UpdateRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string, input *types.RecipeStepInstrumentUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepInstrumentIDKey: recipeStepInstrumentID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepInstrumentIDKey, recipeStepInstrumentID)

	existingRecipeStepInstrument, err := m.db.GetRecipeStepInstrument(ctx, recipeID, recipeStepID, recipeStepInstrumentID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "retrieving existing recipe step instrument")
	}

	existingRecipeStepInstrument.Update(input)
	if err = m.db.UpdateRecipeStepInstrument(ctx, existingRecipeStepInstrument); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe step instrument")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipeStepInstrumentUpdatedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepInstrumentIDKey: recipeStepInstrumentID,
	}))

	return nil
}

func (m *mealPlanningManager) ArchiveRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepInstrumentIDKey: recipeStepInstrumentID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepInstrumentIDKey, recipeStepInstrumentID)

	if err := m.db.ArchiveRecipeStepInstrument(ctx, recipeStepID, recipeStepInstrumentID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving recipe step instrument")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipeStepInstrumentArchivedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepInstrumentIDKey: recipeStepInstrumentID,
	}))

	return nil
}
