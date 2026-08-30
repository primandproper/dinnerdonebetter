package mealplanfinalization

import (
	"context"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/retry"
	"github.com/primandproper/platform-go/v13/saga"
)

// errNotFinalizable indicates a plan the finalize step could not finalize.
//
// It is unretryable. AttemptToFinalizeMealPlan finalizes when every vote is in or the voting
// deadline has passed, and the starter only ever picks plans whose deadline is already behind
// them — so a plan that reaches this is one whose ballots the tally could not resolve, and no
// number of retries changes a ballot.
var errNotFinalizable = platformerrors.New("meal plan could not be finalized")

// steps builds the finalization saga's steps over the repository and the two generators the
// pipeline needs.
//
// # What compensation means here
//
// Finalizing is not undone. A finalized plan is a decision the account's members can see and
// have already been notified about, and taking it back because a downstream step failed would
// be a worse outcome than the failure. That step's Undo is nil, which the package reads as
// "this step needs no compensation" and skips.
//
// The other two delete what this instance wrote and clear the flag that says they wrote it, in
// the same transaction — see UndoMealPlanTaskCreation. Both are cheap to redo and neither is a
// user's own record: tasks and grocery items are derived wholesale from the plan's chosen
// options, so unwinding them costs a regeneration rather than data.
//
// Both compensations are gated on state written by the corresponding Do rather than on the
// presence of IDs, because compensation runs for the step that failed as well as the steps that
// succeeded, and a step that skipped over work somebody else had already done must not unwind
// it.
func steps(
	dataManager mealplanning.Repository,
	analyzer recipeanalysis.RecipeAnalyzer,
	groceryListCreator grocerylistpreparation.GroceryListCreator,
	logger logging.Logger,
) []saga.Step[mealplanning.MealPlanFinalizationState] {
	return []saga.Step[mealplanning.MealPlanFinalizationState]{
		{
			Name: mealplanning.MealPlanFinalizationStepFinalize,
			Do: func(ctx context.Context, state *mealplanning.MealPlanFinalizationState) error {
				return finalize(ctx, dataManager, logger, state)
			},
		},
		{
			Name: mealplanning.MealPlanFinalizationStepCreateTasks,
			Do: func(ctx context.Context, state *mealplanning.MealPlanFinalizationState) error {
				return createTasks(ctx, dataManager, analyzer, logger, state)
			},
			Undo: func(ctx context.Context, state *mealplanning.MealPlanFinalizationState) error {
				return undoCreateTasks(ctx, dataManager, state)
			},
		},
		{
			Name: mealplanning.MealPlanFinalizationStepInitializeGroceryList,
			Do: func(ctx context.Context, state *mealplanning.MealPlanFinalizationState) error {
				return initializeGroceryList(ctx, dataManager, groceryListCreator, logger, state)
			},
			Undo: func(ctx context.Context, state *mealplanning.MealPlanFinalizationState) error {
				return undoInitializeGroceryList(ctx, dataManager, state)
			},
		},
	}
}

// finalize tallies the plan's ballots and marks it finalized.
func finalize(
	ctx context.Context,
	dataManager mealplanning.Repository,
	logger logging.Logger,
	state *mealplanning.MealPlanFinalizationState,
) error {
	l := logger.WithValue(mealplanningkeys.MealPlanIDKey, state.MealPlanID)

	finalized, err := dataManager.AttemptToFinalizeMealPlan(ctx, state.MealPlanID, state.AccountID)
	switch {
	case errors.Is(err, mealplanningrepo.ErrAlreadyFinalized):
		// Already done, by a user request or by this step on an earlier attempt whose result
		// never made it into the instance row. Either way the step's effect is in place, which
		// is the whole of what it promises.
		l.Info("meal plan was already finalized")

		return nil
	case err != nil:
		return err
	case !finalized:
		return retry.Unretryable(platformerrors.Wrapf(errNotFinalizable, "meal plan %q", state.MealPlanID))
	}

	l.Info("meal plan finalized")

	return nil
}

// createTasks derives the plan's prep tasks from its chosen options.
func createTasks(
	ctx context.Context,
	dataManager mealplanning.Repository,
	analyzer recipeanalysis.RecipeAnalyzer,
	logger logging.Logger,
	state *mealplanning.MealPlanFinalizationState,
) error {
	l := logger.WithValue(mealplanningkeys.MealPlanIDKey, state.MealPlanID)

	mealPlan, err := dataManager.GetMealPlan(ctx, state.MealPlanID, state.AccountID)
	if err != nil {
		return err
	}

	// The flag commits with the tasks it describes, so it is the step's idempotency guard and a
	// stronger one than a key recorded in a separate transaction could be. A plan that already
	// carries it had its tasks written — by an earlier attempt of this saga, or by the job this
	// saga replaced — and the state records that this instance is not the one that wrote them,
	// so the compensation leaves them alone.
	if mealPlan.TasksCreated {
		l.Info("meal plan already had its tasks created")

		return nil
	}

	options, err := dataManager.GetFinalizedMealPlanOptionsForMealPlan(ctx, state.MealPlanID)
	if err != nil {
		return err
	}

	inputs := []*mealplanning.MealPlanTaskDatabaseCreationInput{}
	for _, option := range options {
		for _, recipeID := range option.RecipeIDs {
			recipe, recipeErr := dataManager.GetRecipe(ctx, recipeID)
			if recipeErr != nil {
				return recipeErr
			}

			steps, stepsErr := analyzer.GenerateMealPlanTasksForRecipe(ctx, option.MealPlanOptionID, recipe)
			if stepsErr != nil {
				return stepsErr
			}

			inputs = append(inputs, steps...)
		}
	}

	created, err := dataManager.CreateMealPlanTasksForMealPlan(ctx, state.MealPlanID, inputs)
	if err != nil {
		return err
	}

	state.TasksCreated = true
	state.CreatedTaskIDs = make([]string, 0, len(created))
	for _, task := range created {
		state.CreatedTaskIDs = append(state.CreatedTaskIDs, task.ID)
	}

	l.WithValue("created", len(created)).Info("meal plan tasks created")

	return nil
}

// undoCreateTasks removes the tasks this instance created.
func undoCreateTasks(
	ctx context.Context,
	dataManager mealplanning.Repository,
	state *mealplanning.MealPlanFinalizationState,
) error {
	if !state.TasksCreated {
		return nil
	}

	if err := dataManager.UndoMealPlanTaskCreation(ctx, state.MealPlanID, state.CreatedTaskIDs); err != nil {
		return err
	}

	state.TasksCreated = false
	state.CreatedTaskIDs = nil

	return nil
}

// initializeGroceryList builds the plan's grocery list.
func initializeGroceryList(
	ctx context.Context,
	dataManager mealplanning.Repository,
	groceryListCreator grocerylistpreparation.GroceryListCreator,
	logger logging.Logger,
	state *mealplanning.MealPlanFinalizationState,
) error {
	l := logger.WithValue(mealplanningkeys.MealPlanIDKey, state.MealPlanID)

	mealPlan, err := dataManager.GetMealPlan(ctx, state.MealPlanID, state.AccountID)
	if err != nil {
		return err
	}

	// Same guard, same reasoning as the task step's.
	if mealPlan.GroceryListInitialized {
		l.Info("meal plan already had its grocery list initialized")

		return nil
	}

	inputs, err := groceryListCreator.GenerateGroceryListInputs(ctx, mealPlan)
	if err != nil {
		return err
	}

	created, err := dataManager.InitializeMealPlanGroceryList(ctx, state.MealPlanID, state.AccountID, inputs)
	if err != nil {
		return err
	}

	state.GroceryListInitialized = true
	state.CreatedGroceryListItemIDs = make([]string, 0, len(created))
	for _, item := range created {
		state.CreatedGroceryListItemIDs = append(state.CreatedGroceryListItemIDs, item.ID)
	}

	l.WithValue("created", len(created)).Info("meal plan grocery list initialized")

	return nil
}

// undoInitializeGroceryList removes the grocery list items this instance created.
func undoInitializeGroceryList(
	ctx context.Context,
	dataManager mealplanning.Repository,
	state *mealplanning.MealPlanFinalizationState,
) error {
	if !state.GroceryListInitialized {
		return nil
	}

	if err := dataManager.UndoMealPlanGroceryListInitialization(ctx, state.MealPlanID, state.CreatedGroceryListItemIDs); err != nil {
		return err
	}

	state.GroceryListInitialized = false
	state.CreatedGroceryListItemIDs = nil

	return nil
}
