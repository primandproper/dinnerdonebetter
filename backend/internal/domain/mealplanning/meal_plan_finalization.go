package mealplanning

import (
	"context"

	"github.com/primandproper/platform-go/v9/database"
)

const (
	// MealPlanFinalizationSagaName names the saga definition that carries a meal plan from
	// "voting is over" to "there are tasks and a grocery list". Instances record it, so it is
	// the one string here that must stay stable across deploys.
	MealPlanFinalizationSagaName = "meal_plan_finalization"

	// MealPlanFinalizationStepFinalize tallies the ballots and marks the plan finalized.
	MealPlanFinalizationStepFinalize = "finalize_meal_plan"

	// MealPlanFinalizationStepCreateTasks derives the prep tasks for the chosen options.
	MealPlanFinalizationStepCreateTasks = "create_meal_plan_tasks"

	// MealPlanFinalizationStepInitializeGroceryList builds the plan's grocery list.
	MealPlanFinalizationStepInitializeGroceryList = "initialize_grocery_list"
)

type (
	// MealPlanFinalizationState is what one run of the finalization saga carries between its
	// steps. It is persisted as JSON after every step, so a step that runs after a crash sees
	// exactly what the last one left.
	//
	// The two ID slices are there for compensation and nothing else. An Undo that deleted
	// "the tasks for this meal plan" would also delete work another instance did, or work a
	// user added; these are the rows this instance created, and they are the only rows it may
	// take back.
	//
	// The two booleans are separate from those slices rather than derived from them, because
	// "this step created nothing" and "this step found the work already done and skipped" are
	// different facts with opposite compensations, and a plan with no chosen options
	// legitimately produces zero tasks. An Undo that read an empty slice as "nothing to undo"
	// would be right for the second case and would clear a flag over somebody else's rows in
	// the first.
	MealPlanFinalizationState struct {
		_ struct{} `json:"-"`

		MealPlanID string `json:"mealPlanID"`
		AccountID  string `json:"accountID"`

		CreatedTaskIDs            []string `json:"createdTaskIDs,omitempty"`
		CreatedGroceryListItemIDs []string `json:"createdGroceryListItemIDs,omitempty"`

		// TasksCreated records that this instance's create-tasks step did the writing, as
		// opposed to finding the plan already marked and skipping.
		TasksCreated bool `json:"tasksCreated,omitempty"`

		// GroceryListInitialized records the same for the grocery list step.
		GroceryListInitialized bool `json:"groceryListInitialized,omitempty"`
	}

	// MealPlanFinalizationCandidate is a meal plan the finalization pipeline still owes
	// something to and has no saga for yet.
	//
	// It carries the two fields the saga needs to start rather than a whole MealPlan: the
	// starter reads a page of these on every tick, and hydrating each one into a plan with its
	// events, options, and votes would be a query storm in service of data the saga's first
	// step re-reads anyway.
	MealPlanFinalizationCandidate struct {
		_ struct{} `json:"-"`

		MealPlanID string `json:"mealPlanID"`
		AccountID  string `json:"accountID"`
	}

	// MealPlanFinalizationSagaStarter writes a saga instance using the caller's transaction
	// and returns its ID.
	//
	// It exists so that starting the saga and claiming the plan for it are one commit. A saga
	// started in its own transaction, after the claim has committed, is a plan claimed by a
	// saga that does not exist if the process dies in between — and nothing would ever pick
	// the plan up again, because the claim is exactly what the starter's query filters on.
	MealPlanFinalizationSagaStarter func(ctx context.Context, q database.SQLQueryExecutor) (string, error)
)
