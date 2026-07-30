package managers

import (
	"context"

	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/identifiers"
	"github.com/primandproper/platform-go/v8/observability"
	platformkeys "github.com/primandproper/platform-go/v8/observability/keys"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

func (m *mealPlanningManager) ListRecipes(ctx context.Context, status string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	logger = filter.AttachToLogger(logger)

	if status == "" {
		status = mealplanning.RecipeStatusApproved
	}

	results, err := m.db.GetRecipes(ctx, status, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching list of recipes")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateRecipe(ctx context.Context, creatorID string, input *mealplanning.RecipeCreationRequestInput) (*mealplanning.Recipe, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating recipe input")
	}

	if creatorID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	if err := m.recipeAnalyzer.ValidateRecipeCreationRequestInputIsDAG(ctx, input); err != nil {
		return nil, observability.PrepareError(err, span, "evaluating recipe cyclicity")
	}

	convertedInput, err := converters.ConvertRecipeCreationRequestInputToRecipeDatabaseCreationInput(input)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "converting recipe input")
	}

	convertedInput.CreatedByUser = creatorID
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, convertedInput.ID)

	created, err := m.db.CreateRecipe(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating recipe")
	}

	recipe, err := m.db.GetRecipe(ctx, created.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "retrieving recipe")
	}

	return recipe, nil
}

func (m *mealPlanningManager) ReadRecipe(ctx context.Context, recipeID string) (*mealplanning.Recipe, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	x, err := m.db.GetRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "retrieving recipe")
	}

	return x, nil
}

func (m *mealPlanningManager) SearchRecipes(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(platformkeys.SearchQueryKey, query).WithValue(platformkeys.UseDatabaseKey, !useSearchService)
	tracing.AttachToSpan(span, platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, platformkeys.UseDatabaseKey, !useSearchService)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	logger = filter.AttachToLogger(logger)

	var (
		recipes *filtering.QueryFilteredResult[mealplanning.Recipe]
		err     error
	)

	if useSearchService {
		recipes, err = m.searchRecipesViaIndex(ctx, query, filter)
	}

	if err != nil || recipes == nil {
		recipes, err = m.db.SearchForRecipes(ctx, query, filter)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "failed to search for recipes")
		}
	}

	return recipes, nil
}

// searchRecipesViaIndex searches recipes via the external search index. Returns (nil, err) on search failure, empty results, or GetRecipesWithIDs failure.
func (m *mealPlanningManager) searchRecipesViaIndex(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	recipeSubsets, err := m.recipeSearchIndex.Search(ctx, query)
	if err != nil || len(recipeSubsets) == 0 {
		return nil, err
	}

	ids := make([]string, 0, len(recipeSubsets))
	for _, recipeSubset := range recipeSubsets {
		ids = append(ids, recipeSubset.ID)
	}

	data, err := m.db.GetRecipesWithIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	return filtering.NewQueryFilteredResult(data, uint64(len(data)), uint64(len(data)), func(r *mealplanning.Recipe) string {
		return r.ID
	}, filter), nil
}

func (m *mealPlanningManager) SearchForMealEligibleRecipes(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, platformkeys.SearchQueryKey, query)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	logger = filter.AttachToLogger(logger)

	recipes, err := m.db.SearchForMealEligibleRecipes(ctx, query, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "failed to search external service for recipes")
	}

	return recipes, nil
}

func (m *mealPlanningManager) SearchRecipesWithInstrumentOwnership(ctx context.Context, accountID, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, platformkeys.SearchQueryKey, query)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	logger = filter.AttachToLogger(logger)

	recipes, err := m.db.SearchForRecipesWithInstrumentOwnership(ctx, accountID, query, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "failed to search for recipes with instrument ownership")
	}

	return recipes, nil
}

func (m *mealPlanningManager) UpdateRecipe(ctx context.Context, recipeID string, input *mealplanning.RecipeUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	existingRecipe, err := m.db.GetRecipe(ctx, recipeID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "retrieving existing recipe")
	}

	existingRecipe.Update(input)
	if err = m.db.UpdateRecipe(ctx, existingRecipe); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe")
	}

	return nil
}

func (m *mealPlanningManager) UpdateRecipeStatus(ctx context.Context, recipeID, newStatus string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.RecipeIDKey, recipeID).WithValue("new_status", newStatus)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, "new_status", newStatus)

	if err := m.db.UpdateRecipeStatus(ctx, recipeID, newStatus); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe status")
	}

	return nil
}

func (m *mealPlanningManager) ArchiveRecipe(ctx context.Context, recipeID, ownerID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey: recipeID,
		identitykeys.UserIDKey:       ownerID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, ownerID)

	if err := m.db.ArchiveRecipe(ctx, recipeID, ownerID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving recipe")
	}

	return nil
}

func (m *mealPlanningManager) AddRecipeImage(ctx context.Context, recipeID, uploadedMediaID, uploadedByUser string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if err := m.db.AddRecipeImage(ctx, recipeID, uploadedMediaID, uploadedByUser); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "adding recipe image")
	}

	return nil
}

func (m *mealPlanningManager) RecipeEstimatedPrepSteps(ctx context.Context, recipeID string) ([]*mealplanning.MealPlanTaskDatabaseCreationEstimate, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	x, err := m.db.GetRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "retrieving recipe")
	}

	stepInputs, err := m.recipeAnalyzer.GenerateMealPlanTasksForRecipe(ctx, "", x)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "generating meal plan tasks")
	}

	responseEvents := []*mealplanning.MealPlanTaskDatabaseCreationEstimate{}
	for _, input := range stepInputs {
		responseEvents = append(responseEvents, &mealplanning.MealPlanTaskDatabaseCreationEstimate{
			CreationExplanation: input.CreationExplanation,
		})
	}

	return responseEvents, nil
}

func (m *mealPlanningManager) RecipeImageUpload(ctx context.Context) error {
	_, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return nil
}

func (m *mealPlanningManager) MealMermaid(ctx context.Context, meal *mealplanning.Meal) (string, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.recipeAnalyzer.RenderMermaidDiagramForMeal(ctx, meal), nil
}

func (m *mealPlanningManager) RecipeMermaid(ctx context.Context, recipeID string) (string, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	recipe, err := m.db.GetRecipe(ctx, recipeID)
	if err != nil {
		return "", observability.PrepareAndLogError(err, logger, span, "retrieving recipe")
	}

	return m.recipeAnalyzer.RenderMermaidDiagramForRecipe(ctx, recipe), nil
}

func (m *mealPlanningManager) CloneRecipe(ctx context.Context, recipeID, newOwnerID string) (*mealplanning.Recipe, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.RecipeIDKey: recipeID,
		"new_owner":                  newOwnerID,
	})
	tracing.AttachToSpan(span, identitykeys.UserIDKey, newOwnerID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	original, err := m.db.GetRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe by id")
	}

	cloneInput := cloneRecipe(original, newOwnerID)
	cloneInput.ClonedFromRecipeID = &recipeID

	newRecipe, err := m.db.CreateRecipe(ctx, cloneInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating clone of recipe")
	}

	// Both the created and the cloned event are enqueued into the outbox by CreateRecipe,
	// inside the transaction that writes the clone; see internal/repositories/postgres/events.

	return newRecipe, nil
}

func cloneRecipe(x *mealplanning.Recipe, userID string) *mealplanning.RecipeDatabaseCreationInput {
	ingredientProductIndices := map[string]int{}
	instrumentProductIndices := map[string]int{}
	vesselProductIndices := map[string]int{}
	for _, step := range x.Steps {
		for _, ingredient := range step.Ingredients {
			if ingredient.RecipeStepProductID != nil {
				// nil-check the lookup: a product ID that doesn't resolve to a step in this recipe (e.g. a
				// cross-recipe link) would otherwise nil-deref, and a -1 index would become a huge uint64.
				if productStep := x.FindStepForRecipeStepProductID(*ingredient.RecipeStepProductID); productStep != nil {
					if idx := x.FindStepIndexByID(productStep.ID); idx >= 0 {
						ingredientProductIndices[ingredient.ID] = idx
					}
				}
			}
		}

		for _, instrument := range step.Instruments {
			if instrument.RecipeStepProductID != nil {
				if productStep := x.FindStepForRecipeStepProductID(*instrument.RecipeStepProductID); productStep != nil {
					if idx := x.FindStepIndexByID(productStep.ID); idx >= 0 {
						instrumentProductIndices[instrument.ID] = idx
					}
				}
			}
		}

		for _, vessel := range step.Vessels {
			if vessel.RecipeStepProductID != nil {
				if productStep := x.FindStepForRecipeStepProductID(*vessel.RecipeStepProductID); productStep != nil {
					if idx := x.FindStepIndexByID(productStep.ID); idx >= 0 {
						vesselProductIndices[vessel.ID] = idx
					}
				}
			}
		}
	}

	// clone recipe.
	cloneInput := converters.ConvertRecipeToRecipeDatabaseCreationInput(x)
	cloneInput.CreatedByUser = userID
	// TODO: cloneInput.ClonedFromRecipeID = &x.ID

	cloneInput.ID = identifiers.New()
	// oldToNewStepIDs maps each original recipe step ID to the freshly-generated clone step ID so that
	// references from other structures (notably prep-task task steps) can be rewritten to point at the
	// clone's steps rather than silently keeping the source recipe's step IDs.
	oldToNewStepIDs := make(map[string]string, len(cloneInput.Steps))
	for i := range cloneInput.Steps {
		newRecipeStepID := identifiers.New()
		oldToNewStepIDs[x.Steps[i].ID] = newRecipeStepID
		cloneInput.Steps[i].ID = newRecipeStepID
		for j := range cloneInput.Steps[i].Ingredients {
			if index, ok := ingredientProductIndices[x.Steps[i].Ingredients[j].ID]; ok {
				cloneInput.Steps[i].Ingredients[j].ProductOfRecipeStepIndex = new(uint64(index))
			}
			cloneInput.Steps[i].Ingredients[j].ID = identifiers.New()
			cloneInput.Steps[i].Ingredients[j].BelongsToRecipeStep = newRecipeStepID
		}
		for j := range cloneInput.Steps[i].Instruments {
			if index, ok := instrumentProductIndices[x.Steps[i].Instruments[j].ID]; ok {
				cloneInput.Steps[i].Instruments[j].ProductOfRecipeStepIndex = new(uint64(index))
			}
			cloneInput.Steps[i].Instruments[j].ID = identifiers.New()
			cloneInput.Steps[i].Instruments[j].BelongsToRecipeStep = newRecipeStepID
		}
		for j := range cloneInput.Steps[i].Vessels {
			if index, ok := vesselProductIndices[x.Steps[i].Vessels[j].ID]; ok {
				cloneInput.Steps[i].Vessels[j].ProductOfRecipeStepIndex = new(uint64(index))
			}
			cloneInput.Steps[i].Vessels[j].ID = identifiers.New()
			cloneInput.Steps[i].Vessels[j].BelongsToRecipeStep = newRecipeStepID
		}
		for j := range cloneInput.Steps[i].Products {
			cloneInput.Steps[i].Products[j].ID = identifiers.New()
			cloneInput.Steps[i].Products[j].BelongsToRecipeStep = newRecipeStepID
		}
		for j := range cloneInput.Steps[i].CompletionConditions {
			newCompletionConditionID := identifiers.New()
			cloneInput.Steps[i].CompletionConditions[j].ID = newCompletionConditionID
			cloneInput.Steps[i].CompletionConditions[j].BelongsToRecipeStep = newRecipeStepID
			for k := range cloneInput.Steps[i].CompletionConditions[j].Ingredients {
				cloneInput.Steps[i].CompletionConditions[j].Ingredients[k].ID = identifiers.New()
				cloneInput.Steps[i].CompletionConditions[j].Ingredients[k].BelongsToRecipeStepCompletionCondition = newCompletionConditionID
			}
		}
	}

	// TODO: handle media here eventually

	for i := range cloneInput.PrepTasks {
		newPrepTaskID := identifiers.New()
		cloneInput.PrepTasks[i].ID = newPrepTaskID
		for j := range cloneInput.PrepTasks[i].TaskSteps {
			cloneInput.PrepTasks[i].TaskSteps[j].ID = identifiers.New()
			cloneInput.PrepTasks[i].TaskSteps[j].BelongsToRecipePrepTask = newPrepTaskID
			// rewrite the task step's recipe-step reference to the cloned step ID; without this the clone's
			// prep tasks would keep pointing at the ORIGINAL recipe's steps (the FK still passes because
			// those steps exist, so the corruption is silent).
			if newStepID, ok := oldToNewStepIDs[cloneInput.PrepTasks[i].TaskSteps[j].BelongsToRecipeStep]; ok {
				cloneInput.PrepTasks[i].TaskSteps[j].BelongsToRecipeStep = newStepID
			}
		}
	}

	return cloneInput
}
