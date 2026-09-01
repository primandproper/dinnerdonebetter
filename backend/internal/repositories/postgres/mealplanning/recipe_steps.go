package mealplanning

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexevents"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/uploads/registry"
)

var (
	_ mealplanning.RecipeStepDataManager = (*repository)(nil)
)

// RecipeStepExists fetches whether a recipe step exists from the database.
func (q *repository) RecipeStepExists(ctx context.Context, recipeID, recipeStepID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if recipeStepID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	result, err := q.generatedQuerier.CheckRecipeStepExistence(ctx, q.readDB, &generated.CheckRecipeStepExistenceParams{
		RecipeID:     recipeID,
		RecipeStepID: recipeStepID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing recipe step existence check")
	}

	return result, nil
}

// GetRecipeStep fetches a recipe step from the database.
func (q *repository) GetRecipeStep(ctx context.Context, recipeID, recipeStepID string) (*mealplanning.RecipeStep, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if recipeStepID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	result, err := q.generatedQuerier.GetRecipeStep(ctx, q.readDB, &generated.GetRecipeStepParams{
		RecipeID:     recipeID,
		RecipeStepID: recipeStepID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step")
	}

	recipeStep := &mealplanning.RecipeStep{
		CreatedAt:                 result.CreatedAt,
		MinEstimatedTimeInSeconds: database.Uint32PointerFromNullInt64(result.MinimumEstimatedTimeInSeconds),
		MaxEstimatedTimeInSeconds: database.Uint32PointerFromNullInt64(result.MaximumEstimatedTimeInSeconds),
		MinTemperatureInCelsius:   database.Float32PointerFromNullString(result.MinimumTemperatureInCelsius),
		MaxTemperatureInCelsius:   database.Float32PointerFromNullString(result.MaximumTemperatureInCelsius),
		ArchivedAt:                database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt:             database.TimePointerFromNullTime(result.LastUpdatedAt),
		BelongsToRecipe:           result.BelongsToRecipe,
		ConditionExpression:       result.ConditionExpression,
		ID:                        result.ID,
		Notes:                     result.Notes,
		ExplicitInstructions:      result.ExplicitInstructions,
		Media:                     []*mealplanning.RecipeMedia{},
		Products:                  []*mealplanning.RecipeStepProduct{},
		Instruments:               []*mealplanning.RecipeStepInstrument{},
		Vessels:                   []*mealplanning.RecipeStepVessel{},
		CompletionConditions:      []*mealplanning.RecipeStepCompletionCondition{},
		Ingredients:               []*mealplanning.RecipeStepIngredient{},
		Preparation: mealplanning.ValidPreparation{
			CreatedAt:                   result.ValidPreparationCreatedAt,
			MinInstrumentCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
			MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumInstrumentCount),
			MinIngredientCount:          uint16(result.ValidPreparationMinimumIngredientCount),
			MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumIngredientCount),
			MinVesselCount:              uint16(result.ValidPreparationMinimumVesselCount),
			MaxVesselCount:              database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumVesselCount),
			ArchivedAt:                  database.TimePointerFromNullTime(result.ValidPreparationArchivedAt),
			LastUpdatedAt:               database.TimePointerFromNullTime(result.ValidPreparationLastUpdatedAt),
			IconPath:                    result.ValidPreparationIconPath,
			PastTense:                   result.ValidPreparationPastTense,
			ID:                          result.ValidPreparationID,
			Name:                        result.ValidPreparationName,
			Description:                 result.ValidPreparationDescription,
			Slug:                        result.ValidPreparationSlug,
			RestrictToIngredients:       result.ValidPreparationRestrictToIngredients,
			TemperatureRequired:         result.ValidPreparationTemperatureRequired,
			TimeEstimateRequired:        result.ValidPreparationTimeEstimateRequired,
			ConditionExpressionRequired: result.ValidPreparationConditionExpressionRequired,
			ConsumesVessel:              result.ValidPreparationConsumesVessel,
			OnlyForVessels:              result.ValidPreparationOnlyForVessels,
			YieldsNothing:               result.ValidPreparationYieldsNothing,
		},
		Index:                   uint32(result.Index),
		Optional:                result.Optional,
		StartTimerAutomatically: result.StartTimerAutomatically,
	}

	// Fetch related data for this recipe step
	ingredients, err := q.getRecipeStepIngredientsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step ingredients for recipe step")
	}
	for _, ingredient := range ingredients {
		if ingredient.BelongsToRecipeStep == recipeStep.ID {
			recipeStep.Ingredients = append(recipeStep.Ingredients, ingredient)
		}
	}

	products, err := q.getRecipeStepProductsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step products for recipe step")
	}
	for _, product := range products {
		if product.BelongsToRecipeStep == recipeStep.ID {
			recipeStep.Products = append(recipeStep.Products, product)
		}
	}

	instruments, err := q.getRecipeStepInstrumentsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step instruments for recipe step")
	}
	for _, instrument := range instruments {
		if instrument.BelongsToRecipeStep == recipeStep.ID {
			recipeStep.Instruments = append(recipeStep.Instruments, instrument)
		}
	}

	vessels, err := q.getRecipeStepVesselsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step vessels for recipe step")
	}
	for _, vessel := range vessels {
		if vessel.BelongsToRecipeStep == recipeStep.ID {
			recipeStep.Vessels = append(recipeStep.Vessels, vessel)
		}
	}

	completionConditions, err := q.getRecipeStepCompletionConditionsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step completion conditions for recipe step")
	}
	for _, completionCondition := range completionConditions {
		if completionCondition.BelongsToRecipeStep == recipeStep.ID {
			recipeStep.CompletionConditions = append(recipeStep.CompletionConditions, completionCondition)
		}
	}

	recipeMedia, err := q.getRecipeMediaForRecipeStep(ctx, recipeID, recipeStep.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe media for recipe step")
	}
	recipeStep.Media = recipeMedia

	stepImages, err := q.enrichRecipeStepWithStepImages(ctx, recipeStep.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step images")
	}
	recipeStep.StepImages = stepImages

	return recipeStep, nil
}

// enrichRecipeStepWithStepImages fetches step images for a recipe step and returns the uploaded media.
func (q *repository) enrichRecipeStepWithStepImages(ctx context.Context, recipeStepID string) ([]*registry.Object, error) {
	rows, err := q.GetRecipeStepImagesByStep(ctx, recipeStepID)
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	ids := make([]string, len(rows))
	for i, r := range rows {
		ids[i] = r.UploadedMediaID
	}
	return q.GetUploadedMediaWithIDs(ctx, ids)
}

// getRecipeStepByID fetches a recipe step from the database.
func (q *repository) getRecipeStepByID(ctx context.Context, querier database.SQLQueryExecutor, recipeStepID string) (*mealplanning.RecipeStep, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeStepID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	result, err := q.generatedQuerier.GetRecipeStepByRecipeID(ctx, querier, recipeStepID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step")
	}

	recipeStep := &mealplanning.RecipeStep{
		CreatedAt:                 result.CreatedAt,
		MinEstimatedTimeInSeconds: database.Uint32PointerFromNullInt64(result.MinimumEstimatedTimeInSeconds),
		MaxEstimatedTimeInSeconds: database.Uint32PointerFromNullInt64(result.MaximumEstimatedTimeInSeconds),
		MinTemperatureInCelsius:   database.Float32PointerFromNullString(result.MinimumTemperatureInCelsius),
		MaxTemperatureInCelsius:   database.Float32PointerFromNullString(result.MaximumTemperatureInCelsius),
		ArchivedAt:                database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt:             database.TimePointerFromNullTime(result.LastUpdatedAt),
		BelongsToRecipe:           result.BelongsToRecipe,
		ConditionExpression:       result.ConditionExpression,
		ID:                        result.ID,
		Notes:                     result.Notes,
		ExplicitInstructions:      result.ExplicitInstructions,
		Media:                     []*mealplanning.RecipeMedia{},
		Products:                  []*mealplanning.RecipeStepProduct{},
		Instruments:               []*mealplanning.RecipeStepInstrument{},
		Vessels:                   []*mealplanning.RecipeStepVessel{},
		CompletionConditions:      []*mealplanning.RecipeStepCompletionCondition{},
		Ingredients:               []*mealplanning.RecipeStepIngredient{},
		Preparation: mealplanning.ValidPreparation{
			CreatedAt:                   result.ValidPreparationCreatedAt,
			MinInstrumentCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
			MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumInstrumentCount),
			MinIngredientCount:          uint16(result.ValidPreparationMinimumIngredientCount),
			MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumIngredientCount),
			MinVesselCount:              uint16(result.ValidPreparationMinimumVesselCount),
			MaxVesselCount:              database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumVesselCount),
			ArchivedAt:                  database.TimePointerFromNullTime(result.ValidPreparationArchivedAt),
			LastUpdatedAt:               database.TimePointerFromNullTime(result.ValidPreparationLastUpdatedAt),
			IconPath:                    result.ValidPreparationIconPath,
			PastTense:                   result.ValidPreparationPastTense,
			ID:                          result.ValidPreparationID,
			Name:                        result.ValidPreparationName,
			Description:                 result.ValidPreparationDescription,
			Slug:                        result.ValidPreparationSlug,
			RestrictToIngredients:       result.ValidPreparationRestrictToIngredients,
			TemperatureRequired:         result.ValidPreparationTemperatureRequired,
			TimeEstimateRequired:        result.ValidPreparationTimeEstimateRequired,
			ConditionExpressionRequired: result.ValidPreparationConditionExpressionRequired,
			ConsumesVessel:              result.ValidPreparationConsumesVessel,
			OnlyForVessels:              result.ValidPreparationOnlyForVessels,
			YieldsNothing:               result.ValidPreparationYieldsNothing,
		},
		Index:                   uint32(result.Index),
		Optional:                result.Optional,
		StartTimerAutomatically: result.StartTimerAutomatically,
	}

	// Fetch related data for this recipe step
	ingredients, err := q.getRecipeStepIngredientsForRecipe(ctx, result.BelongsToRecipe)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step ingredients for recipe step")
	}
	for _, ingredient := range ingredients {
		if ingredient.BelongsToRecipeStep == recipeStep.ID {
			recipeStep.Ingredients = append(recipeStep.Ingredients, ingredient)
		}
	}

	products, err := q.getRecipeStepProductsForRecipe(ctx, result.BelongsToRecipe)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step products for recipe step")
	}
	for _, product := range products {
		if product.BelongsToRecipeStep == recipeStep.ID {
			recipeStep.Products = append(recipeStep.Products, product)
		}
	}

	instruments, err := q.getRecipeStepInstrumentsForRecipe(ctx, result.BelongsToRecipe)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step instruments for recipe step")
	}
	for _, instrument := range instruments {
		if instrument.BelongsToRecipeStep == recipeStep.ID {
			recipeStep.Instruments = append(recipeStep.Instruments, instrument)
		}
	}

	vessels, err := q.getRecipeStepVesselsForRecipe(ctx, result.BelongsToRecipe)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step vessels for recipe step")
	}
	for _, vessel := range vessels {
		if vessel.BelongsToRecipeStep == recipeStep.ID {
			recipeStep.Vessels = append(recipeStep.Vessels, vessel)
		}
	}

	completionConditions, err := q.getRecipeStepCompletionConditionsForRecipe(ctx, result.BelongsToRecipe)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step completion conditions for recipe step")
	}
	for _, completionCondition := range completionConditions {
		if completionCondition.BelongsToRecipeStep == recipeStep.ID {
			recipeStep.CompletionConditions = append(recipeStep.CompletionConditions, completionCondition)
		}
	}

	recipeMedia, err := q.getRecipeMediaForRecipeStep(ctx, result.BelongsToRecipe, recipeStep.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe media for recipe step")
	}
	recipeStep.Media = recipeMedia

	stepImages, err := q.enrichRecipeStepWithStepImages(ctx, recipeStep.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step images")
	}
	recipeStep.StepImages = stepImages

	return recipeStep, nil
}

// GetRecipeSteps fetches a list of recipe steps from the database that meet a particular filter.
func (q *repository) GetRecipeSteps(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStep], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetRecipeSteps(ctx, q.readDB, &generated.GetRecipeStepsParams{
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
		RecipeID:        recipeID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe steps")
	}

	var (
		data                      = []*mealplanning.RecipeStep{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		recipeStep := &mealplanning.RecipeStep{
			CreatedAt:                 result.CreatedAt,
			MinEstimatedTimeInSeconds: database.Uint32PointerFromNullInt64(result.MinimumEstimatedTimeInSeconds),
			MaxEstimatedTimeInSeconds: database.Uint32PointerFromNullInt64(result.MaximumEstimatedTimeInSeconds),
			MinTemperatureInCelsius:   database.Float32PointerFromNullString(result.MinimumTemperatureInCelsius),
			MaxTemperatureInCelsius:   database.Float32PointerFromNullString(result.MaximumTemperatureInCelsius),
			ArchivedAt:                database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:             database.TimePointerFromNullTime(result.LastUpdatedAt),
			BelongsToRecipe:           result.BelongsToRecipe,
			ConditionExpression:       result.ConditionExpression,
			ID:                        result.ID,
			Notes:                     result.Notes,
			ExplicitInstructions:      result.ExplicitInstructions,
			Media:                     []*mealplanning.RecipeMedia{},
			Products:                  []*mealplanning.RecipeStepProduct{},
			Instruments:               []*mealplanning.RecipeStepInstrument{},
			Vessels:                   []*mealplanning.RecipeStepVessel{},
			CompletionConditions:      []*mealplanning.RecipeStepCompletionCondition{},
			Ingredients:               []*mealplanning.RecipeStepIngredient{},
			Preparation: mealplanning.ValidPreparation{
				CreatedAt:                   result.ValidPreparationCreatedAt,
				MinInstrumentCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
				MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumInstrumentCount),
				MinIngredientCount:          uint16(result.ValidPreparationMinimumInstrumentCount),
				MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumIngredientCount),
				MinVesselCount:              uint16(result.ValidPreparationMinimumVesselCount),
				MaxVesselCount:              database.Uint16PointerFromNullInt32(result.ValidPreparationMaximumVesselCount),
				ArchivedAt:                  database.TimePointerFromNullTime(result.ValidPreparationArchivedAt),
				LastUpdatedAt:               database.TimePointerFromNullTime(result.ValidPreparationLastUpdatedAt),
				IconPath:                    result.ValidPreparationIconPath,
				PastTense:                   result.ValidPreparationPastTense,
				ID:                          result.ValidPreparationID,
				Name:                        result.ValidPreparationName,
				Description:                 result.ValidPreparationDescription,
				Slug:                        result.ValidPreparationSlug,
				RestrictToIngredients:       result.ValidPreparationRestrictToIngredients,
				TemperatureRequired:         result.ValidPreparationTemperatureRequired,
				TimeEstimateRequired:        result.ValidPreparationTimeEstimateRequired,
				ConditionExpressionRequired: result.ValidPreparationConditionExpressionRequired,
				ConsumesVessel:              result.ValidPreparationConsumesVessel,
				OnlyForVessels:              result.ValidPreparationOnlyForVessels,
				YieldsNothing:               result.ValidPreparationYieldsNothing,
			},
			Index:                   uint32(result.Index),
			Optional:                result.Optional,
			StartTimerAutomatically: result.StartTimerAutomatically,
		}

		data = append(data, recipeStep)
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	// Fetch all related data for all recipe steps
	ingredients, err := q.getRecipeStepIngredientsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step ingredients for recipe steps")
	}

	products, err := q.getRecipeStepProductsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step products for recipe steps")
	}

	instruments, err := q.getRecipeStepInstrumentsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step instruments for recipe steps")
	}

	vessels, err := q.getRecipeStepVesselsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step vessels for recipe steps")
	}

	completionConditions, err := q.getRecipeStepCompletionConditionsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe step completion conditions for recipe steps")
	}

	// Populate each recipe step with its related data
	for i, step := range data {
		for _, ingredient := range ingredients {
			if ingredient.BelongsToRecipeStep == step.ID {
				data[i].Ingredients = append(data[i].Ingredients, ingredient)
			}
		}

		for _, product := range products {
			if product.BelongsToRecipeStep == step.ID {
				data[i].Products = append(data[i].Products, product)
			}
		}

		for _, instrument := range instruments {
			if instrument.BelongsToRecipeStep == step.ID {
				data[i].Instruments = append(data[i].Instruments, instrument)
			}
		}

		for _, vessel := range vessels {
			if vessel.BelongsToRecipeStep == step.ID {
				data[i].Vessels = append(data[i].Vessels, vessel)
			}
		}

		for _, completionCondition := range completionConditions {
			if completionCondition.BelongsToRecipeStep == step.ID {
				data[i].CompletionConditions = append(data[i].CompletionConditions, completionCondition)
			}
		}

		recipeMedia, mediaErr := q.getRecipeMediaForRecipeStep(ctx, recipeID, step.ID)
		if mediaErr != nil {
			return nil, observability.PrepareAndLogError(mediaErr, logger, span, "fetching recipe media for recipe step")
		}
		data[i].Media = recipeMedia
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *mealplanning.RecipeStep) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreateRecipeStep creates a recipe step in the database.
func (q *repository) createRecipeStep(ctx context.Context, db database.Tx, input *mealplanning.RecipeStepDatabaseCreationInput) (*mealplanning.RecipeStep, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	// create the recipe step.
	if err := q.generatedQuerier.CreateRecipeStep(ctx, db, &generated.CreateRecipeStepParams{
		ID:                            input.ID,
		BelongsToRecipe:               input.BelongsToRecipe,
		PreparationID:                 input.PreparationID,
		ConditionExpression:           input.ConditionExpression,
		ExplicitInstructions:          input.ExplicitInstructions,
		Notes:                         input.Notes,
		MaximumTemperatureInCelsius:   database.NullStringFromFloat32Pointer(input.MaxTemperatureInCelsius),
		MinimumTemperatureInCelsius:   database.NullStringFromFloat32Pointer(input.MinTemperatureInCelsius),
		MaximumEstimatedTimeInSeconds: database.NullInt64FromUint32Pointer(input.MaxEstimatedTimeInSeconds),
		MinimumEstimatedTimeInSeconds: database.NullInt64FromUint32Pointer(input.MinEstimatedTimeInSeconds),
		Index:                         int32(input.Index),
		Optional:                      input.Optional,
		StartTimerAutomatically:       input.StartTimerAutomatically,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "performing recipe step creation")
	}

	// Fetch the preparation data
	preparation, err := q.GetValidPreparation(ctx, input.PreparationID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching preparation data")
	}

	x := &mealplanning.RecipeStep{
		ID:                        input.ID,
		Index:                     input.Index,
		Preparation:               *preparation,
		MinEstimatedTimeInSeconds: input.MinEstimatedTimeInSeconds,
		MaxEstimatedTimeInSeconds: input.MaxEstimatedTimeInSeconds,
		MinTemperatureInCelsius:   input.MinTemperatureInCelsius,
		MaxTemperatureInCelsius:   input.MaxTemperatureInCelsius,
		Notes:                     input.Notes,
		ExplicitInstructions:      input.ExplicitInstructions,
		ConditionExpression:       input.ConditionExpression,
		Optional:                  input.Optional,
		BelongsToRecipe:           input.BelongsToRecipe,
		StartTimerAutomatically:   input.StartTimerAutomatically,
		CreatedAt:                 q.CurrentTime(),
	}
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, x.ID)

	for i, ingredientInput := range input.Ingredients {
		ingredientInput.BelongsToRecipeStep = x.ID
		ingredient, createErr := q.createRecipeStepIngredient(ctx, db, ingredientInput)
		if createErr != nil {
			return nil, observability.PrepareError(createErr, span, "creating recipe step ingredient #%d", i+1)
		}

		x.Ingredients = append(x.Ingredients, ingredient)
	}

	for i, productInput := range input.Products {
		productInput.BelongsToRecipeStep = x.ID
		product, createErr := q.createRecipeStepProduct(ctx, db, productInput)
		if createErr != nil {
			return nil, observability.PrepareError(createErr, span, "creating recipe step product #%d", i+1)
		}

		x.Products = append(x.Products, product)
	}

	for i, instrumentInput := range input.Instruments {
		instrumentInput.BelongsToRecipeStep = x.ID
		instrument, createErr := q.createRecipeStepInstrument(ctx, db, instrumentInput)
		if createErr != nil {
			return nil, observability.PrepareError(createErr, span, "creating recipe step instrument #%d", i+1)
		}

		x.Instruments = append(x.Instruments, instrument)
	}

	for i, vesselInput := range input.Vessels {
		vesselInput.BelongsToRecipeStep = x.ID
		vessel, createErr := q.createRecipeStepVessel(ctx, db, vesselInput)
		if createErr != nil {
			return nil, observability.PrepareError(createErr, span, "creating recipe step vessel #%d", i+1)
		}

		x.Vessels = append(x.Vessels, vessel)
	}

	for i, conditionInput := range input.CompletionConditions {
		conditionInput.BelongsToRecipeStep = x.ID
		condition, createErr := q.createRecipeStepCompletionCondition(ctx, db, conditionInput)
		if createErr != nil {
			return nil, observability.PrepareError(createErr, span, "creating recipe step completion condition #%d", i+1)
		}

		x.CompletionConditions = append(x.CompletionConditions, condition)
	}

	return x, nil
}

// CreateRecipeStep creates a recipe step in the database.
//
// The step and everything hanging off it — ingredients, products, instruments, vessels,
// completion conditions — are written in one transaction. They used to be written straight to
// the writer as a dozen independent statements, so a failure partway through left a step whose
// ingredients were half there and no way to tell.
//
// The transaction also carries the recipe's index event, because a new step changes the recipe
// document: the indexed subset holds each step's preparation name and the names of its
// ingredients, instruments and vessels.
//
// This derives the index event alone rather than going through withEvent, because there is no
// recipe_step_created data change event to attach it to. The constant existed and put the event
// in the generated webhook catalog, where it was subscribable and could never fire; it was
// removed rather than made to fire, because a step creation reaches subscribers as the recipe
// event that accompanies it. That still holds, so this passes a trigger rather than an event
// type — it reads out of the same table as everything else and puts nothing on the wire.
func (q *repository) CreateRecipeStep(ctx context.Context, input *mealplanning.RecipeStepDatabaseCreationInput) (*mealplanning.RecipeStep, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, input.BelongsToRecipe)

	var created *mealplanning.RecipeStep
	if err := q.WithTransaction(ctx, func(tx database.Tx) error {
		var createErr error
		if created, createErr = q.createRecipeStep(ctx, tx, input); createErr != nil {
			return createErr
		}

		return q.events.EmitIndex(ctx, tx, indexevents.RecipeStepCreatedIndexTrigger, map[string]any{
			mealplanningkeys.RecipeIDKey: input.BelongsToRecipe,
		})
	}); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateRecipeStep updates a particular recipe step.
func (q *repository) UpdateRecipeStep(ctx context.Context, updated *mealplanning.RecipeStep) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeStepIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, mealplanning.RecipeStepUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:     updated.BelongsToRecipe,
		mealplanningkeys.RecipeStepIDKey: updated.ID,
	}, func(tx database.Tx) error {
		_, updateErr := q.generatedQuerier.UpdateRecipeStep(ctx, tx, &generated.UpdateRecipeStepParams{
			ConditionExpression:           updated.ConditionExpression,
			PreparationID:                 updated.Preparation.ID,
			ID:                            updated.ID,
			BelongsToRecipe:               updated.BelongsToRecipe,
			Notes:                         updated.Notes,
			ExplicitInstructions:          updated.ExplicitInstructions,
			MaximumTemperatureInCelsius:   database.NullStringFromFloat32Pointer(updated.MaxTemperatureInCelsius),
			MinimumTemperatureInCelsius:   database.NullStringFromFloat32Pointer(updated.MinTemperatureInCelsius),
			MaximumEstimatedTimeInSeconds: database.NullInt64FromUint32Pointer(updated.MaxEstimatedTimeInSeconds),
			MinimumEstimatedTimeInSeconds: database.NullInt64FromUint32Pointer(updated.MinEstimatedTimeInSeconds),
			Index:                         int32(updated.Index),
			Optional:                      updated.Optional,
			StartTimerAutomatically:       updated.StartTimerAutomatically,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe step")
	}

	return nil
}

// ArchiveRecipeStep archives a recipe step from the database by its ID.
func (q *repository) ArchiveRecipeStep(ctx context.Context, recipeID, recipeStepID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if recipeStepID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if err := q.withEvent(ctx, logger, mealplanning.RecipeStepArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:     recipeID,
		mealplanningkeys.RecipeStepIDKey: recipeStepID,
	}, func(tx database.Tx) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveRecipeStep(ctx, tx, &generated.ArchiveRecipeStepParams{
			BelongsToRecipe: recipeID,
			ID:              recipeStepID,
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "updating recipe step")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
