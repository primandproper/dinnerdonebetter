package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	mealPlanEventsTableName = "meal_plan_events"

	belongsToMealPlanEventColumn = "belongs_to_meal_plan_event"
	mealPlanEventIDColumn        = "meal_plan_event_id"
	belongsToMealPlanColumn      = "belongs_to_meal_plan"
)

func init() {
	registerTableName(mealPlanEventsTableName)
}

var mealPlanEventsColumns = []string{
	idColumn,
	notesColumn,
	"starts_at",
	"ends_at",
	"meal_name",
	belongsToMealPlanColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildMealPlanEventsQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			pgGen.StandardCRUD(mealPlanEventsTableName, mealPlanEventsColumns,
				querygen.WithEntity("MealPlanEvent", "MealPlanEvents"),
				querygen.WithOwnership(belongsToMealPlanColumn),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "MealPlanEventIsEligibleForVoting",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
		JOIN %s ON %s.%s = %s.%s
	WHERE
		%s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s = 'awaiting_votes'
		AND %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s IS NULL
);`,
						mealPlanEventsTableName, idColumn,
						mealPlanEventsTableName,
						mealPlansTableName, mealPlanEventsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
						mealPlanEventsTableName, archivedAtColumn,
						mealPlansTableName, idColumn, mealPlanIDColumn,
						mealPlansTableName, mealPlanStatusColumn,
						mealPlansTableName, archivedAtColumn,
						mealPlanEventsTableName, idColumn, mealPlanEventIDColumn,
						mealPlanEventsTableName, archivedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "CheckMealPlanEventExistence",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s = sqlc.arg(%s)
);`,
						mealPlanEventsTableName, idColumn,
						mealPlanEventsTableName,
						mealPlanEventsTableName, archivedAtColumn,
						mealPlanEventsTableName, idColumn, idColumn,
						mealPlanEventsTableName, belongsToMealPlanColumn, mealPlanIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetMealPlanEvents",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
GROUP BY %s.%s
%s;`,
						strings.Join(applyToEach(mealPlanEventsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", mealPlanEventsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(mealPlanEventsTableName, mealPlanEventsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanEventsTableName, belongsToMealPlanColumn, mealPlanIDColumn)),
						pgGen.TotalCountSelect(mealPlanEventsTableName, mealPlanEventsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanEventsTableName, belongsToMealPlanColumn, mealPlanIDColumn)),
						mealPlanEventsTableName,
						pgGen.FilterConditions(mealPlanEventsTableName, mealPlanEventsColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanEventsTableName, belongsToMealPlanColumn, mealPlanIDColumn),
						),
						mealPlanEventsTableName,
						idColumn,
						pgGen.CursorLimitClause(mealPlanEventsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetAllMealPlanEventsForMealPlan",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE
	%s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
ORDER BY %s.%s ASC;`,
						strings.Join(applyToEach(mealPlanEventsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", mealPlanEventsTableName, s)
						}), ",\n\t"),
						mealPlanEventsTableName,
						mealPlanEventsTableName, archivedAtColumn,
						mealPlanEventsTableName, belongsToMealPlanColumn, mealPlanIDColumn,
						mealPlanEventsTableName, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateMealPlanEvent",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						mealPlanEventsTableName,
						strings.Join(applyToEach(querygen.ForUpdate(mealPlanEventsColumns), func(i int, s string) string {
							return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
						}), ",\n\t"),
						lastUpdatedAtColumn,
						querygen.NowExpression,
						archivedAtColumn,
						idColumn,
						idColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
