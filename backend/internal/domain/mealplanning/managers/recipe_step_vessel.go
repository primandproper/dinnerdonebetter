package managers

import (
	"context"
	"fmt"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

func (m *mealPlanningManager) ListRecipeStepVessels(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepVessel], error) {
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

	results, err := m.db.GetRecipeStepVessels(ctx, recipeID, recipeStepID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching list of recipe step vessels")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateRecipeStepVessel(ctx context.Context, recipeID, recipeStepID string, input *types.RecipeStepVesselCreationRequestInput) (*types.RecipeStepVessel, error) {
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
		return nil, fmt.Errorf("index is required when creating a recipe step vessel outside of initial recipe creation")
	}

	convertedInput := converters.ConvertRecipeStepVesselCreationRequestInputToRecipeStepVesselDatabaseCreationInput(input, 0)
	convertedInput.BelongsToRecipeStep = recipeStepID
	logger = logger.WithValue(mealplanningkeys.RecipeStepVesselIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepVesselIDKey, convertedInput.ID)

	if convertedInput.ValidPreparationVesselID != nil && *convertedInput.ValidPreparationVesselID != "" {
		vpv, err := m.db.GetValidPreparationVessel(ctx, *convertedInput.ValidPreparationVesselID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid preparation vessel")
		}
		convertedInput.VesselID = &vpv.Vessel.ID
	}

	created, err := m.db.CreateRecipeStepVessel(ctx, recipeID, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating recipe step vessel")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadRecipeStepVessel(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string) (*types.RecipeStepVessel, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:           recipeID,
		mealplanningkeys.RecipeStepIDKey:       recipeStepID,
		mealplanningkeys.RecipeStepVesselIDKey: recipeStepVesselID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepVesselIDKey, recipeStepVesselID)

	x, err := m.db.GetRecipeStepVessel(ctx, recipeID, recipeStepID, recipeStepVesselID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "retrieving recipe step vessel")
	}

	return x, nil
}

func (m *mealPlanningManager) UpdateRecipeStepVessel(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string, input *types.RecipeStepVesselUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:           recipeID,
		mealplanningkeys.RecipeStepIDKey:       recipeStepID,
		mealplanningkeys.RecipeStepVesselIDKey: recipeStepVesselID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepVesselIDKey, recipeStepVesselID)

	existingRecipeStepVessel, err := m.db.GetRecipeStepVessel(ctx, recipeID, recipeStepID, recipeStepVesselID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "retrieving existing recipe step vessel")
	}

	existingRecipeStepVessel.Update(input)
	if err = m.db.UpdateRecipeStepVessel(ctx, recipeID, existingRecipeStepVessel); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe step vessel")
	}

	return nil
}

func (m *mealPlanningManager) ArchiveRecipeStepVessel(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey:           recipeID,
		mealplanningkeys.RecipeStepIDKey:       recipeStepID,
		mealplanningkeys.RecipeStepVesselIDKey: recipeStepVesselID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepVesselIDKey, recipeStepVesselID)

	if err := m.db.ArchiveRecipeStepVessel(ctx, recipeID, recipeStepID, recipeStepVesselID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving recipe step vessel")
	}

	return nil
}
