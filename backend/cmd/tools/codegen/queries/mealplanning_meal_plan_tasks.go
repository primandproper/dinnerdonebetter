package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	mealPlanTasksTableName = "meal_plan_tasks"

	mealPlanTaskIDColumn                 = "meal_plan_task_id"
	mealPlanTaskCompletedAtColumn        = "completed_at"
	mealPlanTaskNotificationSentAtColumn = "notification_sent_at"
	mealPlanTaskStatusExplanationColumn  = "status_explanation"
	mealPlanTaskStatusColumn             = "status"
)

func init() {
	registerTableName(mealPlanTasksTableName)
}

var mealPlanTasksColumns = []string{
	idColumn,
	mealPlanTaskStatusColumn,
	mealPlanTaskStatusExplanationColumn,
	"creation_explanation",
	"belongs_to_meal_plan_option",
	"belongs_to_recipe_prep_task",
	mealPlanTaskCompletedAtColumn,
	mealPlanTaskNotificationSentAtColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	"assigned_to_user",
}

func buildMealPlanTasksQueries(database string) []*Query {
	switch database {
	case postgres:

		insertColumns := filterForInsert(mealPlanTasksColumns, mealPlanTaskCompletedAtColumn, mealPlanTaskNotificationSentAtColumn)

		fullSelectColumns := mergeColumns(
			applyToEach(mealPlanTasksColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s", mealPlanTasksTableName, s)
			}),
			append(
				applyToEach(mealPlanOptionsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as meal_plan_option_%s", mealPlanOptionsTableName, s, s)
				}),
				append(
					applyToEach(recipePrepTasksColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s as prep_task_%s", recipePrepTasksTableName, s, s)
					}),
					applyToEach(recipePrepTaskStepsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s as prep_task_step_%s", recipePrepTaskStepsTableName, s, s)
					})...,
				)...,
			),
			1,
		)

		applyToEach(recipeStepsColumns, func(i int, s string) string {
			return fmt.Sprintf("%s.%s", s, s)
		})

		completeSelectColumns := mergeColumns(
			applyToEach(mealPlanTasksColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s", mealPlanTasksTableName, s)
			}),
			append(
				applyToEach(mealPlanOptionsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as meal_plan_option_%s", mealPlanOptionsTableName, s, s)
				}),

				mergeColumns(
					applyToEach(recipeStepsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s as recipe_step_%s", recipeStepsTableName, s, s)
					}),
					applyToEach(validPreparationsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s as valid_preparation_%s", validPreparationsTableName, s, s)
					}),
					2,
				)...,
			),
			1,
		)

		return []*Query{
			{
				// Compensation for the finalization saga's task creation step. It deletes
				// rather than archives because this table has no archived_at: a task is
				// derived from the plan's chosen options and is regenerated wholesale when
				// the step runs again, so a tombstone would be a row nothing could ever read.
				//
				// The IDs come from the saga's own state, so it only ever removes what that
				// instance created — not whatever else happens to hang off the plan.
				Annotation: QueryAnnotation{
					Name: "DeleteMealPlanTasks",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`DELETE FROM %s WHERE %s = ANY(sqlc.arg(ids)::text[]);`,
					mealPlanTasksTableName,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ChangeMealPlanTaskStatus",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s)
WHERE %s = sqlc.arg(%s);`,
					mealPlanTasksTableName,
					mealPlanTaskCompletedAtColumn, mealPlanTaskCompletedAtColumn,
					mealPlanTaskStatusExplanationColumn, mealPlanTaskStatusExplanationColumn,
					mealPlanTaskStatusColumn, mealPlanTaskStatusColumn,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CreateMealPlanTask",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					mealPlanTasksTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CheckMealPlanTaskExistence",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
		FULL OUTER JOIN %s ON %s.%s=%s.%s
		FULL OUTER JOIN %s ON %s.%s=%s.%s
		FULL OUTER JOIN %s ON %s.%s=%s.%s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
					mealPlanTasksTableName, idColumn,
					mealPlanTasksTableName,
					mealPlanOptionsTableName, mealPlanTasksTableName, belongsToMealPlanOptionColumn, mealPlanOptionsTableName, idColumn,
					mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn,
					mealPlansTableName, mealPlanEventsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
					mealPlanTasksTableName, mealPlanTaskCompletedAtColumn,
					mealPlansTableName, idColumn, mealPlanIDColumn,
					mealPlansTableName, archivedAtColumn,
					mealPlanTasksTableName, idColumn, mealPlanTaskIDColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetMealPlanTask",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	LEFT JOIN %s ON %s.%s=%s.%s
	LEFT JOIN %s ON %s.%s=%s.%s
	LEFT JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
					strings.Join(fullSelectColumns, ",\n\t"),
					mealPlanTasksTableName,
					mealPlanOptionsTableName, mealPlanTasksTableName, belongsToMealPlanOptionColumn, mealPlanOptionsTableName, idColumn,
					mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn,
					mealPlansTableName, mealPlanEventsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
					mealsTableName, mealPlanOptionsTableName, mealIDColumn, mealsTableName, idColumn,
					recipePrepTasksTableName, mealPlanTasksTableName, belongsToRecipePrepTaskColumn, recipePrepTasksTableName, idColumn,
					recipePrepTaskStepsTableName, recipePrepTaskStepsTableName, belongsToRecipePrepTaskColumn, recipePrepTasksTableName, idColumn,
					recipeStepsTableName, recipePrepTaskStepsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
					mealPlanOptionsTableName, archivedAtColumn,
					mealPlanEventsTableName, archivedAtColumn,
					mealPlansTableName, archivedAtColumn,
					mealsTableName, archivedAtColumn,
					recipeStepsTableName, archivedAtColumn,
					mealPlanTasksTableName, idColumn, mealPlanTaskIDColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ListAllMealPlanTasksByMealPlan",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	LEFT JOIN %s ON %s.%s=%s.%s
	LEFT JOIN %s ON %s.%s=%s.%s
	LEFT JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
					strings.Join(fullSelectColumns, ",\n\t"),
					mealPlanTasksTableName,
					mealPlanOptionsTableName, mealPlanTasksTableName, belongsToMealPlanOptionColumn, mealPlanOptionsTableName, idColumn,
					mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn,
					mealPlansTableName, mealPlanEventsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
					mealsTableName, mealPlanOptionsTableName, mealIDColumn, mealsTableName, idColumn,
					recipePrepTasksTableName, mealPlanTasksTableName, belongsToRecipePrepTaskColumn, recipePrepTasksTableName, idColumn,
					recipePrepTaskStepsTableName, recipePrepTaskStepsTableName, belongsToRecipePrepTaskColumn, recipePrepTasksTableName, idColumn,
					recipeStepsTableName, recipePrepTaskStepsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
					mealPlanOptionsTableName, archivedAtColumn,
					mealPlanEventsTableName, archivedAtColumn,
					mealPlansTableName, archivedAtColumn,
					mealsTableName, archivedAtColumn,
					recipeStepsTableName, archivedAtColumn,
					mealPlansTableName, idColumn, mealPlanIDColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ListIncompleteMealPlanTasksByMealPlanOption",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	 FULL OUTER JOIN %s ON %s.%s=%s.%s
	 FULL OUTER JOIN %s ON %s.%s=%s.%s
	 FULL OUTER JOIN %s ON %s.%s=%s.%s
	 JOIN %s ON %s.%s=%s.%s
	 JOIN %s ON %s.%s=%s.%s
WHERE %s.%s = sqlc.arg(%s)
AND %s.%s IS NULL;`,
					strings.Join(completeSelectColumns, ",\n\t"),
					mealPlanTasksTableName,
					mealPlanOptionsTableName, mealPlanTasksTableName, belongsToMealPlanOptionColumn, mealPlanOptionsTableName, idColumn,
					mealPlansTableName, mealPlanOptionsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
					mealsTableName, mealPlanOptionsTableName, mealIDColumn, mealsTableName, idColumn,
					recipeStepsTableName, mealPlanTasksTableName, satisfiesRecipeStepColumn, recipeStepsTableName, idColumn,
					validPreparationsTableName, recipeStepsTableName, preparationIDColumn, validPreparationsTableName, idColumn,
					mealPlanTasksTableName, belongsToMealPlanOptionColumn, belongsToMealPlanOptionColumn,
					mealPlanTasksTableName, mealPlanTaskCompletedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "MealPlanTaskNotificationHasBeenSent",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s IS NOT NULL
FROM %s
WHERE %s.%s = sqlc.arg(%s);`,
					mealPlanTasksTableName, mealPlanTaskNotificationSentAtColumn,
					mealPlanTasksTableName,
					mealPlanTasksTableName, idColumn, mealPlanTaskIDColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "MarkMealPlanTaskNotificationSent",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s
WHERE %s = sqlc.arg(%s);`,
					mealPlanTasksTableName,
					mealPlanTaskNotificationSentAtColumn, currentTimeExpression,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ClearMealPlanTaskNotificationSentForEvent",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = NULL
FROM %s
WHERE %s.%s = %s.%s
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL;`,
					mealPlanTasksTableName,
					mealPlanTaskNotificationSentAtColumn,
					mealPlanOptionsTableName,
					mealPlanTasksTableName, belongsToMealPlanOptionColumn, mealPlanOptionsTableName, idColumn,
					mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventIDColumn,
					mealPlanTasksTableName, mealPlanTaskCompletedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetMealPlanTaskIDsThatNeedNotification",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s > %s;`,
					mealPlanTasksTableName, idColumn,
					mealPlanTasksTableName,
					mealPlanOptionsTableName, mealPlanTasksTableName, belongsToMealPlanOptionColumn, mealPlanOptionsTableName, idColumn,
					mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn,
					mealPlansTableName, mealPlanEventsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
					mealPlanTasksTableName, mealPlanTaskCompletedAtColumn,
					mealPlanTasksTableName, mealPlanTaskNotificationSentAtColumn,
					mealPlanOptionsTableName, archivedAtColumn,
					mealPlanEventsTableName, archivedAtColumn,
					mealPlansTableName, archivedAtColumn,
					mealPlanEventsTableName, "starts_at",
					currentTimeExpression,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetMealPlanTaskAccountID",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE %s.%s = sqlc.arg(%s);`,
					mealPlansTableName, belongsToAccountColumn,
					mealPlanTasksTableName,
					mealPlanOptionsTableName, mealPlanTasksTableName, belongsToMealPlanOptionColumn, mealPlanOptionsTableName, idColumn,
					mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn,
					mealPlansTableName, mealPlanEventsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
					mealPlanTasksTableName, idColumn, mealPlanTaskIDColumn,
				)),
			},
		}
	default:
		return nil
	}
}
