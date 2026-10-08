package mealplanning

import (
	"context"
	"database/sql"
	"fmt"
	"maps"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v15/mediaregistry"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability"
	platformkeys "github.com/primandproper/primitives-go/v2/observability/keys"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/pointer"
)

var (
	_ mealplanning.RecipeDataManager = (*repository)(nil)
)

// RecipeExists fetches whether a recipe exists from the database.
func (q *repository) RecipeExists(ctx context.Context, recipeID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if recipeID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	result, err := q.generatedQuerier.CheckRecipeExistence(ctx, q.readDB, recipeID)
	if err != nil {
		return false, observability.PrepareError(err, span, "performing recipe existence check")
	}

	return result, nil
}

// RecipeIsOwnedBy fetches whether a recipe exists and was written by a given user.
func (q *repository) RecipeIsOwnedBy(ctx context.Context, recipeID, userID string) (bool, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if recipeID == "" || userID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, platformkeys.UserIDKey, userID)

	result, err := q.generatedQuerier.CheckRecipeOwnership(ctx, q.readDB, &generated.CheckRecipeOwnershipParams{
		ID:            recipeID,
		CreatedByUser: userID,
	})
	if err != nil {
		return false, observability.PrepareError(err, span, "performing recipe ownership check")
	}

	return result, nil
}

// getRecipe fetches a recipe from the database.
// visited is an optional set of recipe IDs already visited to prevent infinite recursion in circular dependencies.
func (q *repository) getRecipe(ctx context.Context, recipeID string, visited ...map[string]bool) (*mealplanning.Recipe, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	// Check for circular dependency to prevent infinite recursion
	var seen map[string]bool
	if len(visited) > 0 && visited[0] != nil {
		seen = visited[0]
	} else {
		seen = make(map[string]bool)
	}

	// Track if this is the initial call (recipeID not yet in seen) vs a recursive discovery
	// If recipeID is already in seen, it means we've encountered it in the call chain (cycle detected)
	// Return a minimal recipe to break the cycle
	if seen[recipeID] {
		return &mealplanning.Recipe{
			ID:    recipeID,
			Steps: []*mealplanning.RecipeStep{},
		}, nil
	}
	// Mark as seen before processing to detect cycles during processing
	seen[recipeID] = true

	var x *mealplanning.Recipe
	results, err := q.generatedQuerier.GetRecipeByID(ctx, q.readDB, recipeID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching recipe")
	}

	for _, result := range results {
		if x == nil {
			x = &mealplanning.Recipe{
				CreatedAt:            result.CreatedAt,
				InspiredByRecipeID:   database.StringPointerFromNullString(result.InspiredByRecipeID),
				LastUpdatedAt:        database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:           database.TimePointerFromNullTime(result.ArchivedAt),
				PluralPortionName:    result.PluralPortionName,
				Description:          result.Description,
				Name:                 result.Name,
				PortionName:          result.PortionName,
				ID:                   result.ID,
				CreatedByUser:        result.CreatedByUser,
				Source:               result.Source,
				SourceISBN:           result.SourceIsbn,
				Slug:                 result.Slug,
				YieldsComponentType:  string(result.YieldsComponentType),
				MinEstimatedPortions: database.Float32FromString(result.MinEstimatedPortions),
				MaxEstimatedPortions: database.Float32PointerFromNullString(result.MaxEstimatedPortions),
				Status:               string(result.Status),
				EligibleForMeals:     result.EligibleForMeals,
			}
		}

		// Only add the step if it actually exists (not NULL from LEFT JOIN)
		if result.RecipeStepID.Valid {
			x.Steps = append(x.Steps, &mealplanning.RecipeStep{
				CreatedAt:                 result.RecipeStepCreatedAt.Time,
				MinEstimatedTimeInSeconds: database.Uint32PointerFromNullInt64(result.RecipeStepMinimumEstimatedTimeInSeconds),
				MaxEstimatedTimeInSeconds: database.Uint32PointerFromNullInt64(result.RecipeStepMaximumEstimatedTimeInSeconds),
				MinTemperatureInCelsius:   database.Float32PointerFromNullString(result.RecipeStepMinimumTemperatureInCelsius),
				MaxTemperatureInCelsius:   database.Float32PointerFromNullString(result.RecipeStepMaximumTemperatureInCelsius),
				ArchivedAt:                database.TimePointerFromNullTime(result.RecipeStepArchivedAt),
				LastUpdatedAt:             database.TimePointerFromNullTime(result.RecipeStepLastUpdatedAt),
				BelongsToRecipe:           result.RecipeStepBelongsToRecipe.String,
				ConditionExpression:       result.RecipeStepConditionExpression.String,
				ID:                        result.RecipeStepID.String,
				Notes:                     result.RecipeStepNotes.String,
				ExplicitInstructions:      result.RecipeStepExplicitInstructions.String,
				Preparation: mealplanning.ValidPreparation{
					CreatedAt:                   result.RecipeStepPreparationCreatedAt.Time,
					MinInstrumentCount:          uint16(result.RecipeStepPreparationMinimumInstrumentCount.Int32),
					MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.RecipeStepPreparationMaximumInstrumentCount),
					MinIngredientCount:          uint16(result.RecipeStepPreparationMinimumIngredientCount.Int32),
					MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.RecipeStepPreparationMaximumIngredientCount),
					MinVesselCount:              uint16(result.RecipeStepPreparationMinimumVesselCount.Int32),
					MaxVesselCount:              database.Uint16PointerFromNullInt32(result.RecipeStepPreparationMaximumVesselCount),
					ArchivedAt:                  database.TimePointerFromNullTime(result.RecipeStepPreparationArchivedAt),
					LastUpdatedAt:               database.TimePointerFromNullTime(result.RecipeStepPreparationLastUpdatedAt),
					IconPath:                    result.RecipeStepPreparationIconPath.String,
					PastTense:                   result.RecipeStepPreparationPastTense.String,
					ID:                          result.RecipeStepPreparationID.String,
					Name:                        result.RecipeStepPreparationName.String,
					Description:                 result.RecipeStepPreparationDescription.String,
					Slug:                        result.RecipeStepPreparationSlug.String,
					RestrictToIngredients:       result.RecipeStepPreparationRestrictToIngredients.Bool,
					TemperatureRequired:         result.RecipeStepPreparationTemperatureRequired.Bool,
					TimeEstimateRequired:        result.RecipeStepPreparationTimeEstimateRequired.Bool,
					ConditionExpressionRequired: result.RecipeStepPreparationConditionExpressionRequired.Bool,
					ConsumesVessel:              result.RecipeStepPreparationConsumesVessel.Bool,
					OnlyForVessels:              result.RecipeStepPreparationOnlyForVessels.Bool,
					YieldsNothing:               result.RecipeStepPreparationYieldsNothing.Bool,
				},
				Index:                   uint32(result.RecipeStepIndex.Int32),
				Optional:                result.RecipeStepOptional.Bool,
				StartTimerAutomatically: result.RecipeStepStartTimerAutomatically.Bool,
				Media:                   []*mealplanning.RecipeMedia{},
			})
		}
	}

	if x == nil {
		return nil, sql.ErrNoRows
	}

	prepTasks, err := q.getRecipePrepTasksForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching recipe step prep tasks for recipe")
	}
	if prepTasks != nil {
		x.PrepTasks = prepTasks
	}

	recipeMedia, err := q.getRecipeMediaForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching recipe step media for recipe")
	}
	if recipeMedia != nil {
		x.Media = recipeMedia
	}

	ingredients, err := q.getRecipeStepIngredientsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching recipe step ingredients for recipe")
	}

	products, err := q.getRecipeStepProductsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching recipe step products for recipe")
	}

	instruments, err := q.getRecipeStepInstrumentsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching recipe step instruments for recipe")
	}

	vessels, err := q.getRecipeStepVesselsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching recipe step vessels for recipe")
	}

	completionConditions, err := q.getRecipeStepCompletionConditionsForRecipe(ctx, recipeID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching recipe step completion conditions for recipe")
	}

	for i, step := range x.Steps {
		for _, ingredient := range ingredients {
			if ingredient.BelongsToRecipeStep == step.ID {
				x.Steps[i].Ingredients = append(x.Steps[i].Ingredients, ingredient)
			}
		}

		for _, product := range products {
			if product.BelongsToRecipeStep == step.ID {
				x.Steps[i].Products = append(x.Steps[i].Products, product)
			}
		}

		for _, instrument := range instruments {
			if instrument.BelongsToRecipeStep == step.ID {
				x.Steps[i].Instruments = append(x.Steps[i].Instruments, instrument)
			}
		}

		for _, vessel := range vessels {
			if vessel.BelongsToRecipeStep == step.ID {
				x.Steps[i].Vessels = append(x.Steps[i].Vessels, vessel)
			}
		}

		for _, completionCondition := range completionConditions {
			if completionCondition.BelongsToRecipeStep == step.ID {
				x.Steps[i].CompletionConditions = append(x.Steps[i].CompletionConditions, completionCondition)
			}
		}

		recipeMedia, err = q.getRecipeMediaForRecipeStep(ctx, recipeID, step.ID)
		if err != nil {
			return nil, observability.PrepareError(err, span, "fetching recipe media for recipe step")
		}
		x.Steps[i].Media = recipeMedia

		var stepImages []*mediaregistry.Object
		stepImages, err = q.enrichRecipeStepWithStepImages(ctx, step.ID)
		if err != nil {
			return nil, observability.PrepareError(err, span, "fetching recipe step images")
		}
		x.Steps[i].StepImages = stepImages
	}

	// Check for cross-recipe product references and collect all related recipe IDs
	// We'll flatten the associated recipes so the root recipe contains all of them directly
	var relatedRecipeIDs []string
	for _, step := range x.Steps {
		for _, ingredient := range step.Ingredients {
			if ingredient.RecipeStepProductRecipeID != nil && *ingredient.RecipeStepProductRecipeID != "" && *ingredient.RecipeStepProductRecipeID != x.ID {
				relatedRecipeIDs = append(relatedRecipeIDs, pointer.Dereference(ingredient.RecipeStepProductRecipeID))
			}
		}
	}

	// Map to store fetched recipes by ID (before flattening)
	fetchedRecipes := make(map[string]*mealplanning.Recipe)

	// Track which recipes we've queued to prevent adding the same recipe multiple times
	queuedRecipeIDs := make(map[string]bool)
	for _, id := range relatedRecipeIDs {
		queuedRecipeIDs[id] = true
	}

	// Queue to process recipes and discover nested dependencies
	recipeQueue := make([]string, 0, len(relatedRecipeIDs))
	recipeQueue = append(recipeQueue, relatedRecipeIDs...)

	// Walk the recipes this one draws products from, and theirs, flattening them into
	// AssociatedRecipes. Writes refuse a recipe whose dependencies lead back to it (the manager
	// checks before anything is stored), so the seen set and the iteration cap are a read's
	// defense against a cycle two concurrent writes could still close between them, not the
	// place the rule lives.
	maxIterations := 1000
	iteration := 0
	loopSeen := seen
	if loopSeen == nil {
		loopSeen = make(map[string]bool)
	}
	loopSeen[x.ID] = true // Always include the current recipe to prevent cycles
	for len(recipeQueue) > 0 && iteration < maxIterations {
		iteration++
		rID := recipeQueue[0]
		recipeQueue = recipeQueue[1:]

		// Skip if already fetched
		if fetchedRecipes[rID] != nil {
			continue
		}

		// Already on the chain being read: it is, or will be, flattened in from the nested read
		// that found it.
		if loopSeen[rID] {
			continue
		}

		// Create a copy of loopSeen without rID so rID can be fetched initially
		seenForFetch := make(map[string]bool)
		for id := range loopSeen {
			if id != rID {
				seenForFetch[id] = true
			}
		}
		recipe, getErr := q.getRecipe(ctx, rID, seenForFetch)
		if getErr != nil {
			return nil, observability.PrepareError(getErr, span, "fetching associated recipe")
		}

		// Mark as seen after fetching to prevent cycles in future iterations
		loopSeen[rID] = true

		// Store the fetched recipe
		fetchedRecipes[rID] = recipe

		// Discover nested dependencies by looking at the recipe's ingredients
		// (not its AssociatedRecipes, since we want to flatten those)
		// We discover from ingredients to ensure we find all recipes in the dependency graph
		for _, step := range recipe.Steps {
			for _, ingredient := range step.Ingredients {
				if ingredient.RecipeStepProductRecipeID != nil && *ingredient.RecipeStepProductRecipeID != "" && *ingredient.RecipeStepProductRecipeID != x.ID {
					nestedRecipeID := pointer.Dereference(ingredient.RecipeStepProductRecipeID)
					// Only add to queue if not already fetched and not already queued
					// Note: we check fetchedRecipes, not loopSeen, because loopSeen may contain
					// recipes that were fetched in nested getRecipe calls but aren't in our fetchedRecipes yet
					if fetchedRecipes[nestedRecipeID] == nil && !queuedRecipeIDs[nestedRecipeID] {
						queuedRecipeIDs[nestedRecipeID] = true
						recipeQueue = append(recipeQueue, nestedRecipeID)
					}
				}
			}
		}
	}

	// Extract any recipes that were recursively fetched as part of other recipes' AssociatedRecipes
	// This ensures we capture all nested recipes even if they were fetched recursively
	// We need to do this iteratively until no new recipes are found
	// Limit iterations to prevent infinite loops in case of cycles
	maxExtractionIterations := 100
	extractionIteration := 0
	extracted := true
	for extracted && extractionIteration < maxExtractionIterations {
		extractionIteration++
		extracted = false
		// Collect all recipes to add in this iteration
		toAdd := make(map[string]*mealplanning.Recipe)
		for _, recipe := range fetchedRecipes {
			// Skip minimal recipes (those with empty Steps) as they were returned due to cycle detection
			// and shouldn't have valid AssociatedRecipes to extract from
			if len(recipe.Steps) == 0 {
				continue
			}
			for _, nestedAssociated := range recipe.AssociatedRecipes {
				if fetchedRecipes[nestedAssociated.ID] == nil && toAdd[nestedAssociated.ID] == nil {
					toAdd[nestedAssociated.ID] = nestedAssociated
					extracted = true
				}
			}
		}
		// Add all collected recipes at once
		maps.Copy(fetchedRecipes, toAdd)
	}

	// Second pass: flatten by clearing AssociatedRecipes from all fetched recipes
	// and adding all of them to the root recipe's AssociatedRecipes
	for _, fetchedRecipe := range fetchedRecipes {
		// Clear the AssociatedRecipes to flatten the structure
		fetchedRecipe.AssociatedRecipes = nil
		// Add to root-level AssociatedRecipes
		x.AssociatedRecipes = append(x.AssociatedRecipes, fetchedRecipe)
	}

	return x, nil
}

// GetRecipe fetches a recipe from the database.
func (q *repository) GetRecipe(ctx context.Context, recipeID string) (*mealplanning.Recipe, error) {
	return q.getRecipe(ctx, recipeID, nil)
}

// GetRecipes fetches a list of recipes from the database that meet a particular filter.
func (q *repository) GetRecipes(ctx context.Context, status string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.Recipe], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	filter, logger = filtering.Observe(ctx, logger, filter)

	if status == "" {
		status = mealplanning.RecipeStatusApproved
	}

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetRecipes(ctx, q.readDB, &generated.GetRecipesParams{
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
		Status:          generated.NullRecipeStatus{RecipeStatus: generated.RecipeStatus(status), Valid: true},
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipes list retrieval query")
	}

	x = filtering.Drain(
		results,
		func(result *generated.GetRecipesRow) *mealplanning.Recipe {
			return &mealplanning.Recipe{
				CreatedAt:            result.CreatedAt,
				InspiredByRecipeID:   database.StringPointerFromNullString(result.InspiredByRecipeID),
				LastUpdatedAt:        database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:           database.TimePointerFromNullTime(result.ArchivedAt),
				PluralPortionName:    result.PluralPortionName,
				Description:          result.Description,
				Name:                 result.Name,
				PortionName:          result.PortionName,
				ID:                   result.ID,
				CreatedByUser:        result.CreatedByUser,
				Source:               result.Source,
				SourceISBN:           result.SourceIsbn,
				Slug:                 result.Slug,
				YieldsComponentType:  string(result.YieldsComponentType),
				MinEstimatedPortions: database.Float32FromString(result.MinEstimatedPortions),
				MaxEstimatedPortions: database.Float32PointerFromNullString(result.MaxEstimatedPortions),
				Status:               string(result.Status),
				EligibleForMeals:     result.EligibleForMeals,
			}
		},
		func(result *generated.GetRecipesRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(r *mealplanning.Recipe) string { return r.ID },
		filter,
	)

	return x, nil
}

// GetRecipesCreatedByUser fetches a list of recipes from the database that meet a particular filter.
func (q *repository) GetRecipesCreatedByUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.Recipe], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	filter, logger = filtering.Observe(ctx, logger, filter)

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(platformkeys.UserIDKey, userID)
	tracing.AttachToSpan(span, platformkeys.UserIDKey, userID)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetRecipesCreatedByUser(ctx, q.readDB, &generated.GetRecipesCreatedByUserParams{
		CreatedByUser:   userID,
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipes list retrieval query")
	}

	x = filtering.Drain(
		results,
		func(result *generated.GetRecipesCreatedByUserRow) *mealplanning.Recipe {
			return &mealplanning.Recipe{
				CreatedAt:            result.CreatedAt,
				InspiredByRecipeID:   database.StringPointerFromNullString(result.InspiredByRecipeID),
				LastUpdatedAt:        database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:           database.TimePointerFromNullTime(result.ArchivedAt),
				PluralPortionName:    result.PluralPortionName,
				Description:          result.Description,
				Name:                 result.Name,
				PortionName:          result.PortionName,
				ID:                   result.ID,
				CreatedByUser:        result.CreatedByUser,
				Source:               result.Source,
				SourceISBN:           result.SourceIsbn,
				Slug:                 result.Slug,
				YieldsComponentType:  string(result.YieldsComponentType),
				MinEstimatedPortions: database.Float32FromString(result.MinEstimatedPortions),
				MaxEstimatedPortions: database.Float32PointerFromNullString(result.MaxEstimatedPortions),
				Status:               string(result.Status),
				EligibleForMeals:     result.EligibleForMeals,
			}
		},
		func(result *generated.GetRecipesCreatedByUserRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(r *mealplanning.Recipe) string { return r.ID },
		filter,
	)

	return x, nil
}

// GetRecipesWithIDs fetches a list of recipes from the database that meet a particular filter.
func (q *repository) GetRecipesWithIDs(ctx context.Context, ids []string) ([]*mealplanning.Recipe, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if len(ids) == 0 {
		return []*mealplanning.Recipe{}, nil
	}

	results, err := q.generatedQuerier.GetRecipesWithIDs(ctx, q.readDB, ids)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipes list retrieval by ids")
	}

	recipesByID := map[string]*mealplanning.Recipe{}
	for _, result := range results {
		r, exists := recipesByID[result.ID]
		if !exists {
			r = &mealplanning.Recipe{
				CreatedAt:            result.CreatedAt,
				InspiredByRecipeID:   database.StringPointerFromNullString(result.InspiredByRecipeID),
				LastUpdatedAt:        database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:           database.TimePointerFromNullTime(result.ArchivedAt),
				PluralPortionName:    result.PluralPortionName,
				Description:          result.Description,
				Name:                 result.Name,
				PortionName:          result.PortionName,
				ID:                   result.ID,
				CreatedByUser:        result.CreatedByUser,
				Source:               result.Source,
				SourceISBN:           result.SourceIsbn,
				Slug:                 result.Slug,
				YieldsComponentType:  string(result.YieldsComponentType),
				MinEstimatedPortions: database.Float32FromString(result.MinEstimatedPortions),
				MaxEstimatedPortions: database.Float32PointerFromNullString(result.MaxEstimatedPortions),
				Status:               string(result.Status),
				EligibleForMeals:     result.EligibleForMeals,
				Steps:                []*mealplanning.RecipeStep{},
			}
			recipesByID[result.ID] = r
		}

		// optional step
		if result.RecipeStepID.Valid && result.RecipeStepID.String != "" {
			var prep mealplanning.ValidPreparation
			if result.RecipeStepPreparationID.Valid {
				ingMin := uint16(0)
				if result.RecipeStepPreparationMinimumIngredientCount.Valid {
					ingMin = uint16(result.RecipeStepPreparationMinimumIngredientCount.Int32)
				}
				instMin := uint16(0)
				if result.RecipeStepPreparationMinimumInstrumentCount.Valid {
					instMin = uint16(result.RecipeStepPreparationMinimumInstrumentCount.Int32)
				}
				vesselMin := uint16(0)
				if result.RecipeStepPreparationMinimumVesselCount.Valid {
					vesselMin = uint16(result.RecipeStepPreparationMinimumVesselCount.Int32)
				}

				prep = mealplanning.ValidPreparation{
					ID:                          result.RecipeStepPreparationID.String,
					Name:                        result.RecipeStepPreparationName.String,
					Slug:                        result.RecipeStepPreparationSlug.String,
					Description:                 result.RecipeStepPreparationDescription.String,
					IconPath:                    result.RecipeStepPreparationIconPath.String,
					YieldsNothing:               database.BoolFromNullBool(result.RecipeStepPreparationYieldsNothing),
					RestrictToIngredients:       database.BoolFromNullBool(result.RecipeStepPreparationRestrictToIngredients),
					PastTense:                   result.RecipeStepPreparationPastTense.String,
					MinIngredientCount:          ingMin,
					MaxIngredientCount:          database.Uint16PointerFromNullInt32(result.RecipeStepPreparationMaximumIngredientCount),
					MinInstrumentCount:          instMin,
					MaxInstrumentCount:          database.Uint16PointerFromNullInt32(result.RecipeStepPreparationMaximumInstrumentCount),
					TemperatureRequired:         database.BoolFromNullBool(result.RecipeStepPreparationTemperatureRequired),
					TimeEstimateRequired:        database.BoolFromNullBool(result.RecipeStepPreparationTimeEstimateRequired),
					ConditionExpressionRequired: database.BoolFromNullBool(result.RecipeStepPreparationConditionExpressionRequired),
					ConsumesVessel:              database.BoolFromNullBool(result.RecipeStepPreparationConsumesVessel),
					OnlyForVessels:              database.BoolFromNullBool(result.RecipeStepPreparationOnlyForVessels),
					MinVesselCount:              vesselMin,
					MaxVesselCount:              database.Uint16PointerFromNullInt32(result.RecipeStepPreparationMaximumVesselCount),
					CreatedAt:                   database.TimeFromNullTime(result.RecipeStepPreparationCreatedAt),
					LastUpdatedAt:               database.TimePointerFromNullTime(result.RecipeStepPreparationLastUpdatedAt),
					ArchivedAt:                  database.TimePointerFromNullTime(result.RecipeStepPreparationArchivedAt),
				}
			}

			stepIndex := uint32(0)
			if result.RecipeStepIndex.Valid {
				stepIndex = uint32(result.RecipeStepIndex.Int32)
			}

			r.Steps = append(r.Steps, &mealplanning.RecipeStep{
				ID:                        result.RecipeStepID.String,
				BelongsToRecipe:           result.RecipeStepBelongsToRecipe.String,
				Index:                     stepIndex,
				MinEstimatedTimeInSeconds: database.Uint32PointerFromNullInt64(result.RecipeStepMinimumEstimatedTimeInSeconds),
				MaxEstimatedTimeInSeconds: database.Uint32PointerFromNullInt64(result.RecipeStepMaximumEstimatedTimeInSeconds),
				MinTemperatureInCelsius:   database.Float32PointerFromNullString(result.RecipeStepMinimumTemperatureInCelsius),
				MaxTemperatureInCelsius:   database.Float32PointerFromNullString(result.RecipeStepMaximumTemperatureInCelsius),
				Notes:                     result.RecipeStepNotes.String,
				ExplicitInstructions:      result.RecipeStepExplicitInstructions.String,
				ConditionExpression:       result.RecipeStepConditionExpression.String,
				Optional:                  database.BoolFromNullBool(result.RecipeStepOptional),
				StartTimerAutomatically:   database.BoolFromNullBool(result.RecipeStepStartTimerAutomatically),
				CreatedAt:                 database.TimeFromNullTime(result.RecipeStepCreatedAt),
				LastUpdatedAt:             database.TimePointerFromNullTime(result.RecipeStepLastUpdatedAt),
				ArchivedAt:                database.TimePointerFromNullTime(result.RecipeStepArchivedAt),
				Preparation:               prep,
			})
		}
	}

	out := make([]*mealplanning.Recipe, 0, len(recipesByID))
	for _, r := range recipesByID {
		out = append(out, r)
	}

	return out, nil
}

// ScanRecipeIDsForReindex returns up to limit IDs sorting strictly after `after`, in ascending byte order.
//
// It is the source half of a search reindex: searchsync.Reindexer walks this to find every
// document that should exist, and prunes the index of anything it does not name. It replaces
// the "IDs that need indexing" sampler platform-go v10 removed, which asked a different and
// weaker question — which rows look stale — and could only ever be probabilistically right,
// because a row the sampler had not reached was a row the index was wrong about.
func (q *repository) ScanRecipeIDsForReindex(ctx context.Context, after string, limit int) ([]string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	results, err := q.generatedQuerier.ScanRecipeIDsForReindex(ctx, q.readDB, &generated.ScanRecipeIDsForReindexParams{
		ReindexCursor: after,
		ResultLimit:   limit,
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing recipes reindex scan query")
	}

	return results, nil
}

// SearchForRecipes fetches a list of recipes from the database that match a query.
func (q *repository) SearchForRecipes(ctx context.Context, recipeNameQuery string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.Recipe], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	filter, logger = filtering.Observe(ctx, logger, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.RecipeSearch(ctx, q.readDB, &generated.RecipeSearchParams{
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
		Query:           recipeNameQuery,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipes search query")
	}

	x = filtering.Drain(
		results,
		func(result *generated.RecipeSearchRow) *mealplanning.Recipe {
			return &mealplanning.Recipe{
				CreatedAt:            result.CreatedAt,
				InspiredByRecipeID:   database.StringPointerFromNullString(result.InspiredByRecipeID),
				LastUpdatedAt:        database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:           database.TimePointerFromNullTime(result.ArchivedAt),
				PluralPortionName:    result.PluralPortionName,
				Description:          result.Description,
				Name:                 result.Name,
				PortionName:          result.PortionName,
				ID:                   result.ID,
				CreatedByUser:        result.CreatedByUser,
				Source:               result.Source,
				SourceISBN:           result.SourceIsbn,
				Slug:                 result.Slug,
				YieldsComponentType:  string(result.YieldsComponentType),
				MinEstimatedPortions: database.Float32FromString(result.MinEstimatedPortions),
				MaxEstimatedPortions: database.Float32PointerFromNullString(result.MaxEstimatedPortions),
				Status:               string(result.Status),
				EligibleForMeals:     result.EligibleForMeals,
			}
		},
		func(result *generated.RecipeSearchRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(r *mealplanning.Recipe) string { return r.ID },
		filter,
	)

	return x, nil
}

// SearchForMealEligibleRecipes fetches a list of recipes from the database that match a query.
func (q *repository) SearchForMealEligibleRecipes(ctx context.Context, recipeNameQuery string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.Recipe], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	filter, logger = filtering.Observe(ctx, logger, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.SearchForMealEligibleRecipes(ctx, q.readDB, &generated.SearchForMealEligibleRecipesParams{
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
		Query:           recipeNameQuery,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipes search query")
	}

	x = filtering.Drain(
		results,
		func(result *generated.SearchForMealEligibleRecipesRow) *mealplanning.Recipe {
			return &mealplanning.Recipe{
				CreatedAt:            result.CreatedAt,
				InspiredByRecipeID:   database.StringPointerFromNullString(result.InspiredByRecipeID),
				LastUpdatedAt:        database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:           database.TimePointerFromNullTime(result.ArchivedAt),
				PluralPortionName:    result.PluralPortionName,
				Description:          result.Description,
				Name:                 result.Name,
				PortionName:          result.PortionName,
				ID:                   result.ID,
				CreatedByUser:        result.CreatedByUser,
				Source:               result.Source,
				SourceISBN:           result.SourceIsbn,
				Slug:                 result.Slug,
				YieldsComponentType:  string(result.YieldsComponentType),
				MinEstimatedPortions: database.Float32FromString(result.MinEstimatedPortions),
				MaxEstimatedPortions: database.Float32PointerFromNullString(result.MaxEstimatedPortions),
				Status:               string(result.Status),
				EligibleForMeals:     result.EligibleForMeals,
			}
		},
		func(result *generated.SearchForMealEligibleRecipesRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(r *mealplanning.Recipe) string { return r.ID },
		filter,
	)

	return x, nil
}

// SearchForRecipesWithInstrumentOwnership fetches recipes that match a query and whose required instruments are owned by the account.
func (q *repository) SearchForRecipesWithInstrumentOwnership(ctx context.Context, accountID, recipeNameQuery string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.Recipe], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	filter, logger = filtering.Observe(ctx, logger, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.SearchForRecipesWithInstrumentOwnership(ctx, q.readDB, &generated.SearchForRecipesWithInstrumentOwnershipParams{
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
		Query:           recipeNameQuery,
		AccountID:       accountID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipes search with instrument ownership query")
	}

	x = filtering.Drain(
		results,
		func(result *generated.SearchForRecipesWithInstrumentOwnershipRow) *mealplanning.Recipe {
			return &mealplanning.Recipe{
				CreatedAt:            result.CreatedAt,
				InspiredByRecipeID:   database.StringPointerFromNullString(result.InspiredByRecipeID),
				LastUpdatedAt:        database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:           database.TimePointerFromNullTime(result.ArchivedAt),
				PluralPortionName:    result.PluralPortionName,
				Description:          result.Description,
				Name:                 result.Name,
				PortionName:          result.PortionName,
				ID:                   result.ID,
				CreatedByUser:        result.CreatedByUser,
				Source:               result.Source,
				SourceISBN:           result.SourceIsbn,
				Slug:                 result.Slug,
				YieldsComponentType:  string(result.YieldsComponentType),
				MinEstimatedPortions: database.Float32FromString(result.MinEstimatedPortions),
				MaxEstimatedPortions: database.Float32PointerFromNullString(result.MaxEstimatedPortions),
				Status:               string(result.Status),
				EligibleForMeals:     result.EligibleForMeals,
			}
		},
		func(result *generated.SearchForRecipesWithInstrumentOwnershipRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(r *mealplanning.Recipe) string { return r.ID },
		filter,
	)

	return x, nil
}

// CreateRecipe creates a recipe in the database.
func (q *repository) CreateRecipe(ctx context.Context, input *mealplanning.RecipeDatabaseCreationInput) (*mealplanning.Recipe, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeIDKey, input.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, input.ID)

	var err error
	var x *mealplanning.Recipe
	if err = q.WithTransaction(ctx, func(tx database.Tx) error {
		// create the recipe.
		if err = q.generatedQuerier.CreateRecipe(ctx, tx, &generated.CreateRecipeParams{
			MinEstimatedPortions: database.StringFromFloat32(input.MinEstimatedPortions),
			ID:                   input.ID,
			Slug:                 input.Slug,
			Source:               input.Source,
			SourceIsbn:           input.SourceISBN,
			Description:          input.Description,
			CreatedByUser:        input.CreatedByUser,
			Name:                 input.Name,
			YieldsComponentType:  generated.ComponentType(input.YieldsComponentType),
			PortionName:          input.PortionName,
			PluralPortionName:    input.PluralPortionName,
			MaxEstimatedPortions: database.NullStringFromFloat32Pointer(input.MaxEstimatedPortions),
			InspiredByRecipeID:   database.NullStringFromStringPointer(input.InspiredByRecipeID),
			Status:               mealplanning.RecipeStatusSubmitted,
			EligibleForMeals:     input.EligibleForMeals,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "performing recipe creation query")
		}

		x = &mealplanning.Recipe{
			ID:                   input.ID,
			Name:                 input.Name,
			Slug:                 input.Slug,
			Source:               input.Source,
			SourceISBN:           input.SourceISBN,
			Description:          input.Description,
			InspiredByRecipeID:   input.InspiredByRecipeID,
			CreatedByUser:        input.CreatedByUser,
			MinEstimatedPortions: input.MinEstimatedPortions,
			MaxEstimatedPortions: input.MaxEstimatedPortions,
			Status:               mealplanning.RecipeStatusSubmitted,
			EligibleForMeals:     input.EligibleForMeals,
			PortionName:          input.PortionName,
			PluralPortionName:    input.PluralPortionName,
			YieldsComponentType:  input.YieldsComponentType,
			CreatedAt:            q.CurrentTime(),
			PrepTasks:            []*mealplanning.RecipePrepTask{},
			Steps:                []*mealplanning.RecipeStep{},
			Media:                []*mealplanning.RecipeMedia{},
		}

		if err = q.findCreatedRecipeStepProductsForIngredients(ctx, input); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "finding recipe step products for ingredients")
		}
		q.findCreatedRecipeStepProductsForInstruments(ctx, input)
		q.findCreatedRecipeStepProductsForVessels(ctx, input)

		for i, stepInput := range input.Steps {
			stepInput.Index = uint32(i)
			stepInput.BelongsToRecipe = x.ID

			q.logger.Info(fmt.Sprintf("creating recipe step #%d", i+1))

			var s *mealplanning.RecipeStep
			s, err = q.createRecipeStep(ctx, tx, stepInput)
			if err != nil {
				return observability.PrepareError(err, span, "creating recipe step #%d", i+1)
			}

			x.Steps = append(x.Steps, s)
		}

		for i, prepTaskInput := range input.PrepTasks {
			var pt *mealplanning.RecipePrepTask
			pt, err = q.createRecipePrepTask(ctx, tx, prepTaskInput)
			if err != nil {
				return observability.PrepareError(err, span, "creating recipe prep task #%d", i+1)
			}

			x.PrepTasks = append(x.PrepTasks, pt)
		}

		if input.AlsoCreateMeal {
			if _, err = q.createMeal(ctx, tx, &mealplanning.MealDatabaseCreationInput{
				ID:                   identifiers.New(),
				Name:                 x.Name,
				Description:          x.Description,
				MinEstimatedPortions: x.MinEstimatedPortions,
				MaxEstimatedPortions: x.MaxEstimatedPortions,
				EligibleForMealPlans: x.EligibleForMeals,
				CreatedByUser:        x.CreatedByUser,
				Components: []*mealplanning.MealComponentDatabaseCreationInput{
					{
						RecipeID:      x.ID,
						RecipeScale:   1.0,
						ComponentType: mealplanning.MealComponentTypesMain,
					},
				},
			}); err != nil {
				return observability.PrepareError(err, span, "creating meal from recipe")
			}
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.emit(ctx, tx, logger, mealplanning.RecipeCreatedServiceEventType, "", map[string]any{
			mealplanningkeys.RecipeIDKey: input.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		// A clone is still a creation, so both events describe this one write and both
		// belong to this one transaction. The cloned event names the source, which is the
		// only thing that distinguishes it from an ordinary create — and is why it carries
		// no index event: the new recipe was indexed by the create above, and nothing about
		// the recipe it was cloned from changed.
		if input.ClonedFromRecipeID != nil {
			if emitErr := q.emit(ctx, tx, logger, mealplanning.RecipeClonedServiceEventType, "", map[string]any{
				mealplanningkeys.RecipeIDKey: *input.ClonedFromRecipeID,
			}); emitErr != nil {
				return observability.PrepareError(emitErr, span, "enqueuing recipe cloned event")
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	logger.Info("recipe created")

	return x, nil
}

// findCreatedRecipeStepProductsForIngredients finds and links recipe step products for ingredients.
// It handles both products from the same recipe and products from other recipes (via RecipeStepProductRecipeID).
func (q *repository) findCreatedRecipeStepProductsForIngredients(ctx context.Context, recipe *mealplanning.RecipeDatabaseCreationInput) error {
	for _, step := range recipe.Steps {
		for _, ingredient := range step.Ingredients {
			if ingredient.ProductOfRecipeStepIndex == nil || ingredient.ProductOfRecipeStepProductIndex == nil {
				continue
			}

			// Check if this references a product from a different recipe
			if ingredient.RecipeStepProductRecipeID != nil && *ingredient.RecipeStepProductRecipeID != recipe.ID {
				// Skip if recipe ID is empty (indicates cross-recipe reference that will be resolved later)
				// This can happen when getRecipeIDBySlug returns an empty string because the prerequisite
				// recipe hasn't been created yet. The recipe ID will be resolved in a later pass.
				if *ingredient.RecipeStepProductRecipeID == "" {
					continue
				}
				// Look up the referenced recipe
				referencedRecipe, err := q.getRecipe(ctx, *ingredient.RecipeStepProductRecipeID, nil)
				if err != nil {
					return fmt.Errorf("failed to get referenced recipe %s: %w", *ingredient.RecipeStepProductRecipeID, err)
				}

				// Find the product by step index and product index
				stepIndex := int(*ingredient.ProductOfRecipeStepIndex)
				if stepIndex >= len(referencedRecipe.Steps) {
					continue
				}

				referencedStep := referencedRecipe.Steps[stepIndex]
				productIndex := int(*ingredient.ProductOfRecipeStepProductIndex)
				if productIndex >= len(referencedStep.Products) {
					continue
				}

				product := referencedStep.Products[productIndex]
				ingredient.RecipeStepProductID = &product.ID
				// Inherit measurement unit from the product if not already set (for ingredient-type products)
				if product.Type == mealplanning.RecipeStepProductIngredientType && ingredient.MeasurementUnitID == "" && product.MeasurementUnit != nil {
					ingredient.MeasurementUnitID = product.MeasurementUnit.ID
				}
				continue
			}

			// Original logic: product from the same recipe
			enoughSteps := len(recipe.Steps) > int(*ingredient.ProductOfRecipeStepIndex)
			if !enoughSteps {
				continue
			}

			enoughRecipeStepProducts := len(recipe.Steps[int(*ingredient.ProductOfRecipeStepIndex)].Products) > int(*ingredient.ProductOfRecipeStepProductIndex)
			if !enoughRecipeStepProducts {
				continue
			}

			product := recipe.Steps[*ingredient.ProductOfRecipeStepIndex].Products[*ingredient.ProductOfRecipeStepProductIndex]
			ingredient.RecipeStepProductID = &product.ID
			// Inherit measurement unit from the product if not already set (for ingredient-type products)
			if product.Type == mealplanning.RecipeStepProductIngredientType && ingredient.MeasurementUnitID == "" && product.MeasurementUnitID != nil {
				ingredient.MeasurementUnitID = *product.MeasurementUnitID
			}
		}
	}
	return nil
}

func (q *repository) findCreatedRecipeStepProductsForInstruments(ctx context.Context, recipe *mealplanning.RecipeDatabaseCreationInput) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	for _, step := range recipe.Steps {
		for _, instrument := range step.Instruments {
			if instrument.ProductOfRecipeStepIndex != nil && instrument.ProductOfRecipeStepProductIndex != nil {
				enoughSteps := len(recipe.Steps) > int(*instrument.ProductOfRecipeStepIndex)
				enoughRecipeStepProducts := len(recipe.Steps[int(*instrument.ProductOfRecipeStepIndex)].Products) > int(*instrument.ProductOfRecipeStepProductIndex)
				if enoughSteps && enoughRecipeStepProducts {
					relevantProductIsInstrument := recipe.Steps[*instrument.ProductOfRecipeStepIndex].Products[*instrument.ProductOfRecipeStepProductIndex].Type == mealplanning.RecipeStepProductInstrumentType
					if relevantProductIsInstrument {
						instrument.RecipeStepProductID = &recipe.Steps[*instrument.ProductOfRecipeStepIndex].Products[*instrument.ProductOfRecipeStepProductIndex].ID
					}
				}
			}
		}
	}
}

func (q *repository) findCreatedRecipeStepProductsForVessels(ctx context.Context, recipe *mealplanning.RecipeDatabaseCreationInput) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithSpan(span)

	for _, step := range recipe.Steps {
		for _, vessel := range step.Vessels {
			if vessel.ProductOfRecipeStepIndex != nil && vessel.ProductOfRecipeStepProductIndex != nil {
				enoughSteps := len(recipe.Steps) > int(*vessel.ProductOfRecipeStepIndex)
				enoughRecipeStepProducts := len(recipe.Steps[int(*vessel.ProductOfRecipeStepIndex)].Products) > int(*vessel.ProductOfRecipeStepProductIndex)
				vesselLogger := logger.
					WithValue(mealplanningkeys.RecipeStepIDKey, step.ID).
					WithValue("vessel_id", vessel.ID).
					WithValue("enough_steps", enoughSteps).
					WithValue("enough_recipe_step_products", enoughRecipeStepProducts)
				if enoughSteps && enoughRecipeStepProducts {
					relevantProductIsVessel := recipe.Steps[*vessel.ProductOfRecipeStepIndex].Products[*vessel.ProductOfRecipeStepProductIndex].Type == mealplanning.RecipeStepProductVesselType
					if relevantProductIsVessel {
						vessel.RecipeStepProductID = &recipe.Steps[*vessel.ProductOfRecipeStepIndex].Products[*vessel.ProductOfRecipeStepProductIndex].ID
					} else {
						vesselLogger.WithValue("relevant_product_is_vessel", relevantProductIsVessel).Info("could not find created recipe step product for vessel")
					}
				} else {
					vesselLogger.Info("could not find created recipe step product for vessel")
				}
			}
		}
	}
}

// UpdateRecipe updates a particular recipe, provided ownerID wrote it.
func (q *repository) UpdateRecipe(ctx context.Context, updated *mealplanning.Recipe, ownerID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}

	if ownerID == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(mealplanningkeys.RecipeIDKey, updated.ID).WithValue(platformkeys.UserIDKey, ownerID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, updated.ID)
	tracing.AttachToSpan(span, platformkeys.UserIDKey, ownerID)

	return q.withEvent(ctx, logger, mealplanning.RecipeUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey: updated.ID,
	}, func(tx database.Tx) error {
		rowsAffected, err := q.generatedQuerier.UpdateRecipe(ctx, tx, &generated.UpdateRecipeParams{
			Name:                 updated.Name,
			Slug:                 updated.Slug,
			Source:               updated.Source,
			SourceIsbn:           updated.SourceISBN,
			Description:          updated.Description,
			InspiredByRecipeID:   database.NullStringFromStringPointer(updated.InspiredByRecipeID),
			MinEstimatedPortions: database.StringFromFloat32(updated.MinEstimatedPortions),
			MaxEstimatedPortions: database.NullStringFromFloat32Pointer(updated.MaxEstimatedPortions),
			PortionName:          updated.PortionName,
			PluralPortionName:    updated.PluralPortionName,
			EligibleForMeals:     updated.EligibleForMeals,
			YieldsComponentType:  generated.ComponentType(updated.YieldsComponentType),
			CreatedByUser:        ownerID,
			ID:                   updated.ID,
		})
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating recipe")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		logger.Info("recipe updated")

		return nil
	})
}

// UpdateRecipeStatus updates a particular recipe's status exclusively.
func (q *repository) UpdateRecipeStatus(ctx context.Context, recipeID, newStatus string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.WithSpan(span)

	if recipeID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	return q.withEvent(ctx, logger, mealplanning.RecipeUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey: recipeID,
	}, func(tx database.Tx) error {
		rowsAffected, err := q.generatedQuerier.UpdateRecipeStatus(ctx, tx, &generated.UpdateRecipeStatusParams{
			Status: generated.RecipeStatus(newStatus),
			ID:     recipeID,
		})
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating recipe status")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}

// MarkRecipesAsIndexed stamps last_indexed_at on the rows behind the documents an index has taken.
//
// It is the write half of search/sync's Stamper: the ids arrive already coalesced and ordered
// by the batching.Buffer the syncer stamps through, so this is one statement per flush rather
// than one per document. It is deliberately not guarded on archived_at — the syncer applies a
// vanished row as a delete and stamps nothing, and a row archived between apply and flush is
// harmless to stamp.
func (q *repository) MarkRecipesAsIndexed(ctx context.Context, ids []string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if len(ids) == 0 {
		return nil
	}

	logger := q.logger.Clone().WithValue("id_count", len(ids))
	tracing.AttachToSpan(span, "id_count", len(ids))

	if _, err := q.generatedQuerier.MarkRecipesAsIndexed(ctx, q.writeDB, ids); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking recipes as indexed")
	}

	return nil
}

// ArchiveRecipe archives a recipe from the database by its ID.
func (q *repository) ArchiveRecipe(ctx context.Context, recipeID, userID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if recipeID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(platformkeys.UserIDKey, userID)
	tracing.AttachToSpan(span, platformkeys.UserIDKey, userID)

	return q.withEvent(ctx, logger, mealplanning.RecipeArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey: recipeID,
	}, func(tx database.Tx) error {
		rowsAffected, err := q.generatedQuerier.ArchiveRecipe(ctx, tx, &generated.ArchiveRecipeParams{
			CreatedByUser: userID,
			ID:            recipeID,
		})
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving recipe")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}

// AddRecipeImage adds an uploaded media image to a recipe.
func (q *repository) AddRecipeImage(ctx context.Context, recipeID, uploadedMediaID, uploadedByUser string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if recipeID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	if uploadedMediaID == "" {
		return platformerrors.ErrEmptyInputProvided
	}
	if uploadedByUser == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	if err := q.withEvent(ctx, logger, mealplanning.RecipeImageCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:        recipeID,
		mealplanningkeys.UploadedMediaIDKey: uploadedMediaID,
	}, func(tx database.Tx) error {
		return q.generatedQuerier.CreateRecipeImage(ctx, tx, &generated.CreateRecipeImageParams{
			ID:              identifiers.New(),
			BelongsToRecipe: recipeID,
			UploadedMediaID: uploadedMediaID,
			UploadedByUser:  uploadedByUser,
		})
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "creating recipe image")
	}

	return nil
}
