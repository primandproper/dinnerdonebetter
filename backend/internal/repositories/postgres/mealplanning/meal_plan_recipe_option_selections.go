package mealplanning

import (
	"context"
	"database/sql"
	"errors"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v8/database"
	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

var (
	_ types.MealPlanRecipeOptionSelectionDataManager = (*repository)(nil)
)

// GetSelection fetches a meal plan recipe option selection from the database.
func (q *repository) GetMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string) (*types.MealPlanRecipeOptionSelection, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanOptionID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	if recipeStepID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue("recipe_step_id", recipeStepID)
	tracing.AttachToSpan(span, "recipe_step_id", recipeStepID)

	if selectionType == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue("selection_type", selectionType)
	tracing.AttachToSpan(span, "selection_type", selectionType)

	result, err := q.generatedQuerier.GetMealPlanRecipeOptionSelection(ctx, q.readDB, &generated.GetMealPlanRecipeOptionSelectionParams{
		MealPlanOptionID: mealPlanOptionID,
		RecipeStepID:     recipeStepID,
		IngredientIndex:  int32(ingredientIndex),
		SelectionType:    selectionType,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan recipe option selection")
	}

	selection := &types.MealPlanRecipeOptionSelection{
		ID:                      result.ID,
		BelongsToMealPlanOption: result.BelongsToMealPlanOption,
		RecipeID:                result.RecipeID,
		RecipeStepID:            result.RecipeStepID,
		IngredientIndex:         uint16(result.IngredientIndex),
		SelectedOptionIndex:     uint16(result.SelectedOptionIndex),
		SelectionType:           result.SelectionType,
		CreatedAt:               result.CreatedAt,
		LastUpdatedAt:           database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:              database.TimePointerFromNullTime(result.ArchivedAt),
	}

	return selection, nil
}

// GetSelectionsForMealPlanOption fetches a list of meal plan recipe option selections from the database that meet a particular filter.
func (q *repository) GetSelectionsForMealPlanOption(ctx context.Context, mealPlanOptionID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanRecipeOptionSelection], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanOptionID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetMealPlanRecipeOptionSelectionsForMealPlanOption(ctx, q.readDB, &generated.GetMealPlanRecipeOptionSelectionsForMealPlanOptionParams{
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:    database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:     database.NullTimeFromTimePointer(filter.UpdatedAfter),
		MealPlanOptionID: mealPlanOptionID,
		Cursor:           database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing meal plan recipe option selections list retrieval query")
	}

	var (
		x                         = []*types.MealPlanRecipeOptionSelection{}
		filteredCount, totalCount uint64
	)

	for _, result := range results {
		selection := &types.MealPlanRecipeOptionSelection{
			ID:                      result.ID,
			BelongsToMealPlanOption: result.BelongsToMealPlanOption,
			RecipeID:                result.RecipeID,
			RecipeStepID:            result.RecipeStepID,
			IngredientIndex:         uint16(result.IngredientIndex),
			SelectedOptionIndex:     uint16(result.SelectedOptionIndex),
			SelectionType:           result.SelectionType,
			CreatedAt:               result.CreatedAt,
			LastUpdatedAt:           database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:              database.TimePointerFromNullTime(result.ArchivedAt),
		}

		if filteredCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}

		x = append(x, selection)
	}

	y := filtering.NewQueryFilteredResult(x, filteredCount, totalCount, func(s *types.MealPlanRecipeOptionSelection) string { return s.ID }, filter)

	return y, nil
}

// GetSelectionsForMealPlan fetches all meal plan recipe option selections for a meal plan from the database.
func (q *repository) GetSelectionsForMealPlan(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) ([]*types.MealPlanRecipeOptionSelection, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	logger = filter.AttachToLogger(logger)

	if mealPlanID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	results, err := q.generatedQuerier.GetMealPlanRecipeOptionSelectionsForMealPlan(ctx, q.readDB, &generated.GetMealPlanRecipeOptionSelectionsForMealPlanParams{
		MealPlanID:    mealPlanID,
		CreatedAfter:  database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore: database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore: database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:  database.NullTimeFromTimePointer(filter.UpdatedAfter),
		ResultLimit:   nil, // fetch everything always
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing meal plan recipe option selections for meal plan retrieval query")
	}

	if len(results) == 0 {
		return nil, nil
	}

	x := make([]*types.MealPlanRecipeOptionSelection, 0, len(results))
	for _, result := range results {
		selection := &types.MealPlanRecipeOptionSelection{
			ID:                      result.ID,
			BelongsToMealPlanOption: result.BelongsToMealPlanOption,
			RecipeID:                result.RecipeID,
			RecipeStepID:            result.RecipeStepID,
			IngredientIndex:         uint16(result.IngredientIndex),
			SelectedOptionIndex:     uint16(result.SelectedOptionIndex),
			SelectionType:           result.SelectionType,
			CreatedAt:               result.CreatedAt,
			LastUpdatedAt:           database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:              database.TimePointerFromNullTime(result.ArchivedAt),
		}

		x = append(x, selection)
	}

	return x, nil
}

// CreateSelection creates a meal plan recipe option selection in the database.
func (q *repository) CreateMealPlanRecipeOptionSelection(ctx context.Context, input *types.MealPlanRecipeOptionSelectionDatabaseCreationInput) (*types.MealPlanRecipeOptionSelection, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	logger := q.logger.WithValue("meal_plan_recipe_option_selection_id", input.ID)
	tracing.AttachToSpan(span, "meal_plan_recipe_option_selection_id", input.ID)

	// create the selection
	if err := q.withEvent(ctx, logger, types.MealPlanRecipeOptionSelectionCreatedServiceEventType, "", map[string]any{
		"meal_plan_recipe_option_selection_id": input.ID,
		mealplanningkeys.MealPlanOptionIDKey:   input.BelongsToMealPlanOption,
	}, func(tx database.SQLQueryExecutor) error {
		return q.generatedQuerier.CreateMealPlanRecipeOptionSelection(ctx, tx, &generated.CreateMealPlanRecipeOptionSelectionParams{
			ID:                      input.ID,
			BelongsToMealPlanOption: input.BelongsToMealPlanOption,
			RecipeID:                input.RecipeID,
			RecipeStepID:            input.RecipeStepID,
			IngredientIndex:         int32(input.IngredientIndex),
			SelectedOptionIndex:     int32(input.SelectedOptionIndex),
			SelectionType:           input.SelectionType,
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing meal plan recipe option selection creation query")
	}

	x := &types.MealPlanRecipeOptionSelection{
		ID:                      input.ID,
		BelongsToMealPlanOption: input.BelongsToMealPlanOption,
		RecipeID:                input.RecipeID,
		RecipeStepID:            input.RecipeStepID,
		IngredientIndex:         input.IngredientIndex,
		SelectedOptionIndex:     input.SelectedOptionIndex,
		SelectionType:           input.SelectionType,
		CreatedAt:               q.CurrentTime(),
	}

	tracing.AttachToSpan(span, "meal_plan_recipe_option_selection_id", x.ID)
	logger.Info("meal plan recipe option selection created")

	return x, nil
}

// UpdateSelection updates a meal plan recipe option selection in the database.
func (q *repository) UpdateMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string, input *types.MealPlanRecipeOptionSelectionUpdateRequestInput) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputProvided
	}

	logger := q.logger.Clone()

	if mealPlanOptionID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	if recipeStepID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue("recipe_step_id", recipeStepID)
	tracing.AttachToSpan(span, "recipe_step_id", recipeStepID)

	if selectionType == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue("selection_type", selectionType)
	tracing.AttachToSpan(span, "selection_type", selectionType)

	if input.SelectedOptionIndex == nil {
		return platformerrors.ErrInvalidIDProvided
	}

	// Get existing selection to retrieve recipe_id (needed for update query until SQL is regenerated)
	existing, err := q.GetMealPlanRecipeOptionSelection(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching existing selection for update")
	}
	if existing == nil {
		return sql.ErrNoRows
	}

	rowsAffected, err := q.generatedQuerier.UpdateMealPlanRecipeOptionSelection(ctx, q.writeDB, &generated.UpdateMealPlanRecipeOptionSelectionParams{
		RecipeID:            existing.RecipeID,
		MealPlanOptionID:    mealPlanOptionID,
		RecipeStepID:        recipeStepID,
		IngredientIndex:     int32(ingredientIndex),
		SelectionType:       selectionType,
		SelectedOptionIndex: int32(*input.SelectedOptionIndex),
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal plan recipe option selection")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	logger.Info("meal plan recipe option selection updated")

	return nil
}

// ArchiveSelection archives a meal plan recipe option selection from the database.
func (q *repository) ArchiveMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanOptionID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	if recipeStepID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue("recipe_step_id", recipeStepID)
	tracing.AttachToSpan(span, "recipe_step_id", recipeStepID)

	if selectionType == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue("selection_type", selectionType)
	tracing.AttachToSpan(span, "selection_type", selectionType)

	if err := q.withEvent(ctx, logger, types.MealPlanRecipeOptionSelectionArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.MealPlanOptionIDKey: mealPlanOptionID,
		"recipe_step_id":                     recipeStepID,
		"ingredient_index":                   ingredientIndex,
		"selection_type":                     selectionType,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveMealPlanRecipeOptionSelection(ctx, tx, &generated.ArchiveMealPlanRecipeOptionSelectionParams{
			MealPlanOptionID: mealPlanOptionID,
			RecipeStepID:     recipeStepID,
			IngredientIndex:  int32(ingredientIndex),
			SelectionType:    selectionType,
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving meal plan recipe option selection")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("meal plan recipe option selection archived")

	return nil
}
