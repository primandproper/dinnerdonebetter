package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	mealPlansTableName = "meal_plans"

	mealPlanIDColumn                     = "meal_plan_id"
	mealPlanStatusColumn                 = "status"
	mealPlanVotingDeadlineColumn         = "voting_deadline"
	mealPlanGroceryListInitializedColumn = "grocery_list_initialized"
	mealPlanTasksCreatedColumn           = "tasks_created"
	mealPlanFinalizationSagaIDColumn     = "finalization_saga_id"
	electionMethodColumn                 = "election_method"
)

var mealPlansColumns = []string{
	idColumn,
	notesColumn,
	mealPlanStatusColumn,
	mealPlanVotingDeadlineColumn,
	mealPlanGroceryListInitializedColumn,
	mealPlanTasksCreatedColumn,
	electionMethodColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	belongsToAccountColumn,
	createdByUserColumn,
}

func buildMealPlansQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			pgGen.StandardCRUD(mealPlansTableName, mealPlansColumns,
				querygen.WithEntity("MealPlan", "MealPlans"),
				querygen.WithOwnership(belongsToAccountColumn),
				querygen.WithDatabaseOwned(mealPlanGroceryListInitializedColumn, mealPlanTasksCreatedColumn, electionMethodColumn),
				querygen.WithImmutable(createdByUserColumn),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "CheckMealPlanExistence",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s = sqlc.arg(%s)
);`,
						mealPlansTableName, idColumn,
						mealPlansTableName,
						mealPlansTableName, archivedAtColumn,
						mealPlansTableName, idColumn, mealPlanIDColumn,
						mealPlansTableName, belongsToAccountColumn, belongsToAccountColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "FinalizeMealPlan",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(
						`UPDATE %s SET %s = sqlc.arg(%s) WHERE %s IS NULL AND %s = sqlc.arg(%s);`,
						mealPlansTableName,
						mealPlanStatusColumn,
						mealPlanStatusColumn,
						archivedAtColumn,
						idColumn,
						idColumn,
					)),
				},
				{
					// The finalization saga's working set: every plan the pipeline still owes
					// something to and has no saga for yet. The two arms are the two ways a plan
					// gets here — one that has just run out of voting time, and one that was
					// finalized by a user request or by a build that predates the saga and never
					// had its tasks or its grocery list built.
					//
					// The second arm is also the migration. A plan left half-processed by the
					// three jobs this replaced is picked up by exactly the same predicate, and
					// the saga's steps skip whatever it already has.
					Annotation: QueryAnnotation{
						Name: "GetMealPlansAwaitingFinalizationSaga",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s.%s,
	%s.%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND (
		(%s.%s = 'awaiting_votes' AND %s.%s < %s)
		OR (%s.%s = 'finalized' AND (%s.%s IS FALSE OR %s.%s IS FALSE))
	)
ORDER BY %s.%s
LIMIT sqlc.arg(query_limit);`,
						mealPlansTableName, idColumn,
						mealPlansTableName, belongsToAccountColumn,
						mealPlansTableName,
						mealPlansTableName, archivedAtColumn,
						mealPlansTableName, mealPlanFinalizationSagaIDColumn,
						mealPlansTableName, mealPlanStatusColumn, mealPlansTableName, mealPlanVotingDeadlineColumn, querygen.NowExpression,
						mealPlansTableName, mealPlanStatusColumn, mealPlansTableName, mealPlanTasksCreatedColumn, mealPlansTableName, mealPlanGroceryListInitializedColumn,
						mealPlansTableName, mealPlanVotingDeadlineColumn,
					)),
				},
				{
					// Claims a plan for one saga. The IS NULL predicate is what makes starting
					// idempotent: the instance row is written in this same transaction, so a
					// second starter that loses the race matches no rows here and rolls its
					// instance back rather than running the pipeline twice.
					Annotation: QueryAnnotation{
						Name: "AttachMealPlanFinalizationSaga",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	%s = %s
WHERE %s IS NULL
	AND %s IS NULL
	AND %s = sqlc.arg(%s);`,
						mealPlansTableName,
						mealPlanFinalizationSagaIDColumn, mealPlanFinalizationSagaIDColumn,
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						mealPlanFinalizationSagaIDColumn,
						idColumn, idColumn,
					)),
				},
				{
					// The task creator's query, narrowed to one plan. It used to select every
					// finalized plan with tasks_created IS FALSE, because the job rediscovered
					// its own work; the saga already knows which plan it is running for, and the
					// flag it used to filter on is now the step's own idempotency guard.
					Annotation: QueryAnnotation{
						Name: "GetFinalizedMealPlanOptionsForMealPlan",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s.%s as meal_plan_id,
	%s.%s as meal_plan_option_id,
	%s.%s as meal_id,
	%s.%s as meal_plan_event_id,
	%s.%s as %s
FROM
	%s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s = 'finalized'
	AND %s.%s IS TRUE
	AND %s.%s = sqlc.arg(%s)
GROUP BY
	%s.%s,
	%s.%s,
	%s.%s,
	%s.%s,
	%s.%s
ORDER BY
	%s.%s,
	%s.%s,
	%s.%s,
	%s.%s,
	%s.%s;`,
						mealPlansTableName, idColumn,
						mealPlanOptionsTableName, idColumn,
						mealsTableName, idColumn,
						mealPlanEventsTableName, idColumn,
						mealComponentsTableName, recipeIDColumn, recipeIDColumn,
						mealPlanOptionsTableName,
						mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn,
						mealPlansTableName, mealPlanEventsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
						mealComponentsTableName, mealPlanOptionsTableName, mealIDColumn, mealComponentsTableName, belongsToMealColumn,
						mealsTableName, mealPlanOptionsTableName, mealIDColumn, mealsTableName, idColumn,
						mealPlansTableName, archivedAtColumn,
						mealPlansTableName, mealPlanStatusColumn,
						mealPlanOptionsTableName, mealPlanOptionsChosenColumn,
						mealPlansTableName, idColumn, mealPlanIDColumn,
						mealPlansTableName, idColumn,
						mealPlanOptionsTableName, idColumn,
						mealsTableName, idColumn,
						mealPlanEventsTableName, idColumn,
						mealComponentsTableName, recipeIDColumn,
						mealPlansTableName, idColumn,
						mealPlanOptionsTableName, idColumn,
						mealsTableName, idColumn,
						mealPlanEventsTableName, idColumn,
						mealComponentsTableName, recipeIDColumn,
					)),
				},
				{
					// Compensation for the grocery list step, paired with the DELETE of the items
					// the saga recorded. The flag and the rows it describes are cleared in one
					// transaction, exactly as they were written.
					Annotation: QueryAnnotation{
						Name: "UnmarkMealPlanGroceryListInitialized",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = FALSE,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						mealPlansTableName,
						mealPlanGroceryListInitializedColumn,
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					// Compensation for the task creation step. See above.
					Annotation: QueryAnnotation{
						Name: "UnmarkMealPlanPrepTasksCreated",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = FALSE,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						mealPlansTableName,
						mealPlanTasksCreatedColumn,
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetMealPlansForAccount",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(applyToEach(mealPlansColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", mealPlansTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(mealPlansTableName, mealPlansColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlansTableName, belongsToAccountColumn, belongsToAccountColumn)),
						pgGen.TotalCountSelect(mealPlansTableName, mealPlansColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlansTableName, belongsToAccountColumn, belongsToAccountColumn)),
						mealPlansTableName,
						pgGen.FilterConditions(mealPlansTableName, mealPlansColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlansTableName, belongsToAccountColumn, belongsToAccountColumn),
						),
						pgGen.CursorLimitClause(mealPlansTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetMealPlanPastVotingDeadline",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(meal_plan_id)
	AND %s.%s = sqlc.arg(account_id)
	AND %s.%s = 'awaiting_votes'
	AND %s > %s.%s;`,
						strings.Join(applyToEach(mealPlansColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", mealPlansTableName, s)
						}), ",\n\t"),
						mealPlansTableName,
						mealPlansTableName, archivedAtColumn,
						mealPlansTableName, idColumn,
						mealPlansTableName, belongsToAccountColumn,
						mealPlansTableName, mealPlanStatusColumn,
						querygen.NowExpression, mealPlansTableName, mealPlanVotingDeadlineColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "FindMealPlansForDates",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT DISTINCT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s <= sqlc.arg(end_time)
	AND %s.%s >= sqlc.arg(start_time)
ORDER BY %s.%s;`,
						strings.Join(applyToEach(mealPlansColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", mealPlansTableName, s)
						}), ",\n\t"),
						mealPlansTableName,
						mealPlanEventsTableName, mealPlansTableName, idColumn, mealPlanEventsTableName, belongsToMealPlanColumn,
						mealPlansTableName, archivedAtColumn,
						mealPlanEventsTableName, archivedAtColumn,
						mealPlansTableName, belongsToAccountColumn, belongsToAccountColumn,
						mealPlanEventsTableName, "starts_at",
						mealPlanEventsTableName, "ends_at",
						mealPlansTableName, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "MarkMealPlanAsGroceryListInitialized",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = TRUE,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						mealPlansTableName,
						mealPlanGroceryListInitializedColumn,
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "MarkMealPlanAsPrepTasksCreated",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = TRUE,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						mealPlansTableName,
						mealPlanTasksCreatedColumn,
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
