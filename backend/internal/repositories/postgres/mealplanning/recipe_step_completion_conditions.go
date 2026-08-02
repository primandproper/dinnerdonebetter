package mealplanning

import (
	"context"
	"database/sql"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

var (
	_ types.RecipeStepCompletionConditionDataManager = (*repository)(nil)
)

// RecipeStepCompletionConditionExists fetches whether a recipe step completion condition exists from the database.
func (q *repository) RecipeStepCompletionConditionExists(ctx context.Context, recipeID, recipeStepID, recipeStepCompletionConditionID string) (exists bool, err error) {
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

	if recipeStepCompletionConditionID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepCompletionConditionIDKey, recipeStepCompletionConditionID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepCompletionConditionIDKey, recipeStepCompletionConditionID)

	result, err := q.generatedQuerier.CheckRecipeStepCompletionConditionExistence(ctx, q.readDB, &generated.CheckRecipeStepCompletionConditionExistenceParams{
		RecipeStepID:                    recipeStepID,
		RecipeStepCompletionConditionID: recipeStepCompletionConditionID,
		RecipeID:                        recipeID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing recipe step completion condition existence check")
	}

	return result, nil
}

// GetRecipeStepCompletionCondition fetches a recipe step completion condition from the database.
func (q *repository) GetRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID, recipeStepCompletionConditionID string) (*types.RecipeStepCompletionCondition, error) {
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

	if recipeStepCompletionConditionID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepCompletionConditionIDKey, recipeStepCompletionConditionID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepCompletionConditionIDKey, recipeStepCompletionConditionID)

	results, err := q.generatedQuerier.GetRecipeStepCompletionConditionWithIngredients(ctx, q.readDB, &generated.GetRecipeStepCompletionConditionWithIngredientsParams{
		RecipeID:                        recipeID,
		RecipeStepID:                    recipeStepID,
		RecipeStepCompletionConditionID: recipeStepCompletionConditionID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "querying for recipe step completion condition")
	}

	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}

	recipeStepCompletionCondition := &types.RecipeStepCompletionCondition{}
	for _, result := range results {
		if recipeStepCompletionCondition.ID == "" {
			recipeStepCompletionCondition = &types.RecipeStepCompletionCondition{
				CreatedAt:     result.CreatedAt,
				ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
				LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
				IngredientState: types.ValidIngredientState{
					CreatedAt:     result.ValidIngredientStateCreatedAt,
					ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientStateArchivedAt),
					LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientStateLastUpdatedAt),
					PastTense:     result.ValidIngredientStatePastTense,
					Description:   result.ValidIngredientStateDescription,
					IconPath:      result.ValidIngredientStateIconPath,
					ID:            result.ValidIngredientStateID,
					Name:          result.ValidIngredientStateName,
					AttributeType: string(result.ValidIngredientStateAttributeType),
					Slug:          result.ValidIngredientStateSlug,
				},
				ID:                  result.ID,
				BelongsToRecipeStep: result.BelongsToRecipeStep,
				Notes:               result.Notes,
				Ingredients:         []*types.RecipeStepCompletionConditionIngredient{},
				Optional:            result.Optional,
			}
		}

		recipeStepCompletionCondition.Ingredients = append(recipeStepCompletionCondition.Ingredients, &types.RecipeStepCompletionConditionIngredient{
			CreatedAt:                              result.RecipeStepCompletionConditionIngredientCreatedAt,
			ArchivedAt:                             database.TimePointerFromNullTime(result.RecipeStepCompletionConditionIngredientArchivedAt),
			LastUpdatedAt:                          database.TimePointerFromNullTime(result.RecipeStepCompletionConditionIngredientLastUpdatedAt),
			ID:                                     result.RecipeStepCompletionConditionIngredientID,
			BelongsToRecipeStepCompletionCondition: result.RecipeStepCompletionConditionIngredientBelongsToRecipeS,
			RecipeStepIngredient:                   result.RecipeStepCompletionConditionIngredientRecipeStepIngredi,
		})
	}

	return recipeStepCompletionCondition, nil
}

// getRecipeStepCompletionConditionsForRecipe fetches a recipe step completion condition from the database.
func (q *repository) getRecipeStepCompletionConditionsForRecipe(ctx context.Context, recipeID string) ([]*types.RecipeStepCompletionCondition, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeIDKey, recipeID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeIDKey, recipeID)

	results, err := q.generatedQuerier.GetAllRecipeStepCompletionConditionsForRecipe(ctx, q.readDB, recipeID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "querying for recipe step completion condition")
	}

	idOrder := []string{}
	byID := map[string]*types.RecipeStepCompletionCondition{}
	for _, result := range results {
		recipeStepCompletionCondition := &types.RecipeStepCompletionCondition{
			CreatedAt:     result.CreatedAt,
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			IngredientState: types.ValidIngredientState{
				CreatedAt:     result.ValidIngredientStateCreatedAt,
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientStateArchivedAt),
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientStateLastUpdatedAt),
				PastTense:     result.ValidIngredientStatePastTense,
				Description:   result.ValidIngredientStateDescription,
				IconPath:      result.ValidIngredientStateIconPath,
				ID:            result.ValidIngredientStateID,
				Name:          result.ValidIngredientStateName,
				AttributeType: string(result.ValidIngredientStateAttributeType),
				Slug:          result.ValidIngredientStateSlug,
			},
			ID:                  result.ID,
			BelongsToRecipeStep: result.BelongsToRecipeStep,
			Notes:               result.Notes,
			Optional:            result.Optional,
		}

		if byID[recipeStepCompletionCondition.ID] == nil {
			byID[recipeStepCompletionCondition.ID] = recipeStepCompletionCondition
			idOrder = append(idOrder, recipeStepCompletionCondition.ID)
		}

		// Add ingredient to the completion condition
		byID[recipeStepCompletionCondition.ID].Ingredients = append(byID[recipeStepCompletionCondition.ID].Ingredients, &types.RecipeStepCompletionConditionIngredient{
			CreatedAt:                              result.RecipeStepCompletionConditionIngredientCreatedAt,
			ID:                                     result.RecipeStepCompletionConditionIngredientID,
			BelongsToRecipeStepCompletionCondition: result.RecipeStepCompletionConditionIngredientBelongsToRecipeS,
			RecipeStepIngredient:                   result.RecipeStepCompletionConditionIngredientRecipeStepIngredi,
		})
	}

	recipeStepConditions := []*types.RecipeStepCompletionCondition{}
	for _, id := range idOrder {
		recipeStepConditions = append(recipeStepConditions, byID[id])
	}

	return recipeStepConditions, nil
}

// GetRecipeStepCompletionConditions fetches a list of recipe step completion conditions from the database that meet a particular filter.
func (q *repository) GetRecipeStepCompletionConditions(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.RecipeStepCompletionCondition], err error) {
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

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	var (
		data          []*types.RecipeStepCompletionCondition
		filteredCount uint64
		totalCount    uint64
	)

	results, err := q.generatedQuerier.GetRecipeStepCompletionConditions(ctx, q.readDB, &generated.GetRecipeStepCompletionConditionsParams{
		RecipeStepID:    recipeStepID,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing recipe step completion conditions list retrieval query")
	}

	idOrder := []string{}
	byID := map[string]*types.RecipeStepCompletionCondition{}
	for _, result := range results {
		recipeStepCompletionCondition := &types.RecipeStepCompletionCondition{
			CreatedAt:     result.CreatedAt,
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			IngredientState: types.ValidIngredientState{
				CreatedAt:     result.ValidIngredientStateCreatedAt,
				ArchivedAt:    database.TimePointerFromNullTime(result.ValidIngredientStateArchivedAt),
				LastUpdatedAt: database.TimePointerFromNullTime(result.ValidIngredientStateLastUpdatedAt),
				PastTense:     result.ValidIngredientStatePastTense,
				Description:   result.ValidIngredientStateDescription,
				IconPath:      result.ValidIngredientStateIconPath,
				ID:            result.ValidIngredientStateID,
				Name:          result.ValidIngredientStateName,
				AttributeType: string(result.ValidIngredientStateAttributeType),
				Slug:          result.ValidIngredientStateSlug,
			},
			ID:                  result.ID,
			BelongsToRecipeStep: result.BelongsToRecipeStep,
			Notes:               result.Notes,
			Optional:            result.Optional,
		}

		if byID[recipeStepCompletionCondition.ID] == nil {
			byID[recipeStepCompletionCondition.ID] = recipeStepCompletionCondition
			idOrder = append(idOrder, recipeStepCompletionCondition.ID)
		}

		// Add ingredient to the completion condition
		byID[recipeStepCompletionCondition.ID].Ingredients = append(byID[recipeStepCompletionCondition.ID].Ingredients, &types.RecipeStepCompletionConditionIngredient{
			CreatedAt:                              result.RecipeStepCompletionConditionIngredientCreatedAt,
			ID:                                     result.RecipeStepCompletionConditionIngredientID,
			BelongsToRecipeStepCompletionCondition: result.RecipeStepCompletionConditionIngredientBelongsToRecipeS,
			RecipeStepIngredient:                   result.RecipeStepCompletionConditionIngredientRecipeStepIngredi,
		})

		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
	}

	for _, id := range idOrder {
		data = append(data, byID[id])
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(rscc *types.RecipeStepCompletionCondition) string { return rscc.ID },
		filter,
	)

	return x, nil
}

// createRecipeStepCompletionCondition creates a recipe step completion condition in the database.
func (q *repository) createRecipeStepCompletionCondition(ctx context.Context, db database.SQLQueryExecutor, input *types.RecipeStepCompletionConditionDatabaseCreationInput) (*types.RecipeStepCompletionCondition, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	// create the recipe step completion condition.
	if err := q.generatedQuerier.CreateRecipeStepCompletionCondition(ctx, db, &generated.CreateRecipeStepCompletionConditionParams{
		ID:                  input.ID,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		IngredientState:     input.IngredientStateID,
		Notes:               input.Notes,
		Optional:            input.Optional,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "performing recipe step completion condition creation query")
	}

	x := &types.RecipeStepCompletionCondition{
		ID:                  input.ID,
		Notes:               input.Notes,
		Optional:            input.Optional,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		IngredientState:     types.ValidIngredientState{ID: input.IngredientStateID},
		CreatedAt:           q.CurrentTime(),
	}

	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepCompletionConditionIDKey, x.ID)

	for _, ingredient := range input.Ingredients {
		ingredient.BelongsToRecipeStepCompletionCondition = x.ID
		completionConditionIngredient, err := q.createRecipeStepCompletionConditionIngredient(ctx, db, ingredient)
		if err != nil {
			return nil, observability.PrepareError(err, span, "creating ingredient for recipe step completion condition")
		}

		x.Ingredients = append(x.Ingredients, completionConditionIngredient)
	}

	q.logger.WithValue(mealplanningkeys.RecipeStepCompletionConditionIDKey, x.ID).Info("completion condition created")

	return x, nil
}

// createRecipeStepCompletionConditionIngredient creates a recipe step completion condition ingredient in the database.
func (q *repository) createRecipeStepCompletionConditionIngredient(ctx context.Context, db database.SQLQueryExecutor, input *types.RecipeStepCompletionConditionIngredientDatabaseCreationInput) (*types.RecipeStepCompletionConditionIngredient, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	// create the recipe step completion condition.
	if err := q.generatedQuerier.CreateRecipeStepCompletionConditionIngredient(ctx, db, &generated.CreateRecipeStepCompletionConditionIngredientParams{
		ID:                                     input.ID,
		BelongsToRecipeStepCompletionCondition: input.BelongsToRecipeStepCompletionCondition,
		RecipeStepIngredient:                   input.RecipeStepIngredient,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "performing recipe step completion condition ingredient creation query")
	}

	x := &types.RecipeStepCompletionConditionIngredient{
		ID:                                     input.ID,
		BelongsToRecipeStepCompletionCondition: input.BelongsToRecipeStepCompletionCondition,
		RecipeStepIngredient:                   input.RecipeStepIngredient,
		CreatedAt:                              q.CurrentTime(),
	}

	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepCompletionConditionIDKey, x.ID)

	return x, nil
}

// CreateRecipeStepCompletionCondition creates a recipe step completion condition in the database.
func (q *repository) CreateRecipeStepCompletionCondition(ctx context.Context, recipeID string, input *types.RecipeStepCompletionConditionDatabaseCreationInput) (*types.RecipeStepCompletionCondition, error) {
	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	var created *types.RecipeStepCompletionCondition

	// The write and its event share a transaction.
	if err := q.withEvent(ctx, q.logger, types.RecipeStepCompletionConditionCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:                        recipeID,
		mealplanningkeys.RecipeStepIDKey:                    input.BelongsToRecipeStep,
		mealplanningkeys.RecipeStepCompletionConditionIDKey: input.ID,
	}, func(tx database.SQLQueryExecutor) error {
		var createErr error
		created, createErr = q.createRecipeStepCompletionCondition(ctx, tx, input)

		return createErr
	}); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateRecipeStepCompletionCondition updates a particular recipe step completion condition.
func (q *repository) UpdateRecipeStepCompletionCondition(ctx context.Context, recipeID string, updated *types.RecipeStepCompletionCondition) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.RecipeStepCompletionConditionIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepCompletionConditionIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, types.RecipeStepCompletionConditionUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:                        recipeID,
		mealplanningkeys.RecipeStepIDKey:                    updated.BelongsToRecipeStep,
		mealplanningkeys.RecipeStepCompletionConditionIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		_, updateErr := q.generatedQuerier.UpdateRecipeStepCompletionCondition(ctx, tx, &generated.UpdateRecipeStepCompletionConditionParams{
			Optional:            updated.Optional,
			Notes:               updated.Notes,
			BelongsToRecipeStep: updated.BelongsToRecipeStep,
			IngredientState:     updated.IngredientState.ID,
			ID:                  updated.ID,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating recipe step completion condition")
	}

	logger.Info("recipe step completion condition updated")

	return nil
}

// ArchiveRecipeStepCompletionCondition archives a recipe step completion condition from the database by its ID.
func (q *repository) ArchiveRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID, recipeStepCompletionConditionID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if recipeStepID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepIDKey, recipeStepID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepIDKey, recipeStepID)

	if recipeStepCompletionConditionID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.RecipeStepCompletionConditionIDKey, recipeStepCompletionConditionID)
	tracing.AttachToSpan(span, mealplanningkeys.RecipeStepCompletionConditionIDKey, recipeStepCompletionConditionID)

	if err := q.withEvent(ctx, logger, types.RecipeStepCompletionConditionArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.RecipeIDKey:                        recipeID,
		mealplanningkeys.RecipeStepIDKey:                    recipeStepID,
		mealplanningkeys.RecipeStepCompletionConditionIDKey: recipeStepCompletionConditionID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveRecipeStepCompletionCondition(ctx, tx, &generated.ArchiveRecipeStepCompletionConditionParams{
			BelongsToRecipeStep: recipeStepID,
			ID:                  recipeStepCompletionConditionID,
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "updating recipe step completion condition")
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
