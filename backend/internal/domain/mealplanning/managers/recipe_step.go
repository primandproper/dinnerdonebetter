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

func (m *mealPlanningManager) ListRecipeSteps(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStep], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	logger = filter.AttachToLogger(logger)

	results, err := m.db.GetRecipeSteps(ctx, recipeID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching list of recipe steps")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateRecipeStep(ctx context.Context, recipeID string, input *types.RecipeStepCreationRequestInput) (*types.RecipeStep, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	convertedInput := converters.ConvertRecipeStepCreationInputToRecipeStepDatabaseCreationInput(input)
	convertedInput.BelongsToRecipe = recipeID
	logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, convertedInput.ID)

	for _, ingredient := range convertedInput.Ingredients {
		if ingredient.ValidIngredientPreparationID != nil && *ingredient.ValidIngredientPreparationID != "" {
			vip, err := m.db.GetValidIngredientPreparation(ctx, *ingredient.ValidIngredientPreparationID)
			if err != nil {
				return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient preparation")
			}
			ingredient.IngredientID = &vip.Ingredient.ID
		}
		if ingredient.ValidIngredientMeasurementUnitID != nil && *ingredient.ValidIngredientMeasurementUnitID != "" {
			vimu, err := m.db.GetValidIngredientMeasurementUnit(ctx, *ingredient.ValidIngredientMeasurementUnitID)
			if err != nil {
				return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient measurement unit")
			}
			ingredient.MeasurementUnitID = vimu.MeasurementUnit.ID
		}
	}

	for _, instrument := range convertedInput.Instruments {
		if instrument.ValidPreparationInstrumentID != nil && *instrument.ValidPreparationInstrumentID != "" {
			vpi, err := m.db.GetValidPreparationInstrument(ctx, *instrument.ValidPreparationInstrumentID)
			if err != nil {
				return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation instrument")
			}
			instrument.InstrumentID = &vpi.Instrument.ID
		}
	}

	for _, vessel := range convertedInput.Vessels {
		if vessel.ValidPreparationVesselID != nil && *vessel.ValidPreparationVesselID != "" {
			vpv, err := m.db.GetValidPreparationVessel(ctx, *vessel.ValidPreparationVesselID)
			if err != nil {
				return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation vessel")
			}
			vessel.VesselID = &vpv.Vessel.ID
		}
	}

	created, err := m.db.CreateRecipeStep(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating recipe step")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipeStepCreatedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:     recipeID,
		mealplanningkeys.RecipeStepIDKey: convertedInput.ID,
	}))

	return created, nil
}

func (m *mealPlanningManager) ReadRecipeStep(ctx context.Context, recipeID, recipeStepID string) (*types.RecipeStep, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     recipeID,
		mealplanningkeys.RecipeStepIDKey: recipeStepID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	x, err := m.db.GetRecipeStep(ctx, recipeID, recipeStepID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "retrieving recipe step")
	}

	return x, nil
}

func (m *mealPlanningManager) UpdateRecipeStep(ctx context.Context, recipeID, recipeStepID string, input *types.RecipeStepUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     recipeID,
		mealplanningkeys.RecipeStepIDKey: recipeStepID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	existingRecipeStep, err := m.db.GetRecipeStep(ctx, recipeID, recipeStepID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "retrieving existing recipe step")
	}

	existingRecipeStep.Update(input)
	if err = m.db.UpdateRecipeStep(ctx, existingRecipeStep); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe step")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipeStepUpdatedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:     recipeID,
		mealplanningkeys.RecipeStepIDKey: recipeStepID,
	}))

	return nil
}

func (m *mealPlanningManager) ArchiveRecipeStep(ctx context.Context, recipeID, recipeStepID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:     recipeID,
		mealplanningkeys.RecipeStepIDKey: recipeStepID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if err := m.db.ArchiveRecipeStep(ctx, recipeID, recipeStepID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving recipe step")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipeStepArchivedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:     recipeID,
		mealplanningkeys.RecipeStepIDKey: recipeStepID,
	}))

	return nil
}

func (m *mealPlanningManager) AddRecipeStepImage(ctx context.Context, recipeStepID, uploadedMediaID, uploadedByUser string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if err := m.db.AddRecipeStepImage(ctx, recipeStepID, uploadedMediaID, uploadedByUser); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "adding recipe step image")
	}

	return nil
}

func (m *mealPlanningManager) RecipeStepImageUpload(ctx context.Context) error {
	_, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return nil
}
