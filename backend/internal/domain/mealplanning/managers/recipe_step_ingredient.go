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

func (m *mealPlanningManager) ListRecipeStepIngredients(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepIngredient], error) {
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

	results, err := m.db.GetRecipeStepIngredients(ctx, recipeID, recipeStepID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching list of recipe step ingredients")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID string, input *types.RecipeStepIngredientCreationRequestInput) (*types.RecipeStepIngredient, error) {
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
		return nil, fmt.Errorf("index is required when creating a recipe step ingredient outside of initial recipe creation")
	}

	convertedInput := converters.ConvertRecipeStepIngredientCreationRequestInputToRecipeStepIngredientDatabaseCreationInput(input, 0)
	convertedInput.BelongsToRecipeStep = recipeStepID
	logger = logger.WithValue(mealplanningkeys.RecipeStepIngredientIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIngredientIDKey, convertedInput.ID)

	if convertedInput.ValidIngredientPreparationID != nil && *convertedInput.ValidIngredientPreparationID != "" {
		vip, err := m.db.GetValidIngredientPreparation(ctx, *convertedInput.ValidIngredientPreparationID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient preparation")
		}
		convertedInput.IngredientID = &vip.Ingredient.ID
	}

	if convertedInput.ValidIngredientMeasurementUnitID != nil && *convertedInput.ValidIngredientMeasurementUnitID != "" {
		vimu, err := m.db.GetValidIngredientMeasurementUnit(ctx, *convertedInput.ValidIngredientMeasurementUnitID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredient measurement unit")
		}
		convertedInput.MeasurementUnitID = vimu.MeasurementUnit.ID
	}

	created, err := m.db.CreateRecipeStepIngredient(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating recipe step ingredient")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipeStepIngredientCreatedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepIngredientIDKey: convertedInput.ID,
	}))

	return created, nil
}

func (m *mealPlanningManager) ReadRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string) (*types.RecipeStepIngredient, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepIngredientIDKey: recipeStepIngredientID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIngredientIDKey, recipeStepIngredientID)

	x, err := m.db.GetRecipeStepIngredient(ctx, recipeID, recipeStepID, recipeStepIngredientID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "retrieving recipe step ingredient")
	}

	return x, nil
}

func (m *mealPlanningManager) UpdateRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string, input *types.RecipeStepIngredientUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepIngredientIDKey: recipeStepIngredientID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	existingRecipeStepIngredient, err := m.db.GetRecipeStepIngredient(ctx, recipeID, recipeStepID, recipeStepIngredientID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "retrieving existing recipe step ingredient")
	}

	existingRecipeStepIngredient.Update(input)
	if err = m.db.UpdateRecipeStepIngredient(ctx, existingRecipeStepIngredient); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe step ingredient")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipeStepIngredientUpdatedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepIngredientIDKey: recipeStepIngredientID,
	}))

	return nil
}

func (m *mealPlanningManager) ArchiveRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepIngredientIDKey: recipeStepIngredientID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIngredientIDKey, recipeStepIngredientID)

	if err := m.db.ArchiveRecipeStepIngredient(ctx, recipeStepID, recipeStepIngredientID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving recipe step ingredient")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.RecipeStepIngredientArchivedServiceEventType, map[string]any{
		mealplanningkeys.RecipeIDKey:               recipeID,
		mealplanningkeys.RecipeStepIDKey:           recipeStepID,
		mealplanningkeys.RecipeStepIngredientIDKey: recipeStepIngredientID,
	}))

	return nil
}
