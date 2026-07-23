package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	mealPlanOptionsTableName = "meal_plan_options"

	mealPlanOptionIDColumn         = "meal_plan_option_id"
	mealPlanOptionsChosenColumn    = "chosen"
	mealPlanOptionsTiebrokenColumn = "tiebroken"
	mealPlanOptionsMealScaleColumn = "meal_scale"
)

func init() {
	registerTableName(mealPlanOptionsTableName)
}

var mealPlanOptionsColumns = []string{
	idColumn,
	"assigned_cook",
	"assigned_dishwasher",
	mealPlanOptionsChosenColumn,
	mealPlanOptionsTiebrokenColumn,
	mealPlanOptionsMealScaleColumn,
	"meal_id",
	notesColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	belongsToMealPlanEventColumn,
}

func buildMealPlanOptionsQueries(database string) []*Query {
	switch database {
	case postgres:

		insertColumns := filterForInsert(mealPlanOptionsColumns,
			mealPlanOptionsTiebrokenColumn,
		)

		fullSelectColumns := append(
			applyToEach(mealPlanOptionsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s", mealPlanOptionsTableName, s)
			}),
			applyToEach(mealsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as meal_%s", mealsTableName, s, s)
			})...,
		)

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "CheckMealInMealPlanEvent",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s = sqlc.arg(%s)
);`,
					mealPlanOptionsTableName, idColumn,
					mealPlanOptionsTableName,
					mealPlanOptionsTableName, archivedAtColumn,
					mealPlanOptionsTableName, belongsToMealPlanEventColumn, belongsToMealPlanEventColumn,
					mealPlanOptionsTableName, mealIDColumn, mealIDColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveMealPlanOption",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
					mealPlanOptionsTableName,
					archivedAtColumn, currentTimeExpression,
					archivedAtColumn,
					belongsToMealPlanEventColumn, belongsToMealPlanEventColumn,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CreateMealPlanOption",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					mealPlanOptionsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CheckMealPlanOptionExistence",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
		JOIN meal_plan_events ON meal_plan_options.belongs_to_meal_plan_event = meal_plan_events.id
		JOIN meal_plans ON meal_plan_events.belongs_to_meal_plan = meal_plans.id
	WHERE %s.%s IS NULL
		AND meal_plan_options.belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id)
		AND meal_plan_options.id = sqlc.arg(meal_plan_option_id)
		AND meal_plan_events.archived_at IS NULL
		AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
		AND meal_plan_events.id = sqlc.arg(meal_plan_event_id)
		AND meal_plans.archived_at IS NULL
		AND meal_plans.id = sqlc.arg(meal_plan_id)
);`,
					mealPlanOptionsTableName, idColumn,
					mealPlanOptionsTableName,
					mealPlanOptionsTableName, archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CheckMealPlanOptionBelongsToAccount",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
		JOIN meal_plan_events ON meal_plan_options.belongs_to_meal_plan_event = meal_plan_events.id
		JOIN meal_plans ON meal_plan_events.belongs_to_meal_plan = meal_plans.id
	WHERE %s.%s IS NULL
		AND meal_plan_options.id = sqlc.arg(meal_plan_option_id)
		AND meal_plans.archived_at IS NULL
		AND meal_plans.belongs_to_account = sqlc.arg(belongs_to_account)
);`,
					mealPlanOptionsTableName, idColumn,
					mealPlanOptionsTableName,
					mealPlanOptionsTableName, archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "FinalizeMealPlanOption",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = (%s = sqlc.arg(%s) AND %s = sqlc.arg(%s)),
	%s = sqlc.arg(%s)
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
					mealPlanOptionsTableName,
					mealPlanOptionsChosenColumn, belongsToMealPlanEventColumn, mealPlanEventIDColumn, idColumn, idColumn,
					mealPlanOptionsTiebrokenColumn, mealPlanOptionsTiebrokenColumn,
					archivedAtColumn,
					belongsToMealPlanEventColumn, mealPlanEventIDColumn,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetAllMealPlanOptionsForMealPlanEvent",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN meal_plan_events ON meal_plan_options.belongs_to_meal_plan_event = meal_plan_events.id
	JOIN meal_plans ON meal_plan_events.belongs_to_meal_plan = meal_plans.id
	JOIN meals ON meal_plan_options.meal_id = meals.id
WHERE
	meal_plan_options.archived_at IS NULL
	AND meal_plan_options.belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id)
	AND meal_plan_events.id = sqlc.arg(meal_plan_event_id)
	AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
	AND meal_plans.archived_at IS NULL
	AND meal_plans.id = sqlc.arg(meal_plan_id);`,
					strings.Join(fullSelectColumns, ",\n\t"),
					mealPlanOptionsTableName,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetMealPlanOptions",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM meal_plan_options
	JOIN meal_plan_events ON meal_plan_options.belongs_to_meal_plan_event = meal_plan_events.id
	JOIN meal_plans ON meal_plan_events.belongs_to_meal_plan = meal_plans.id
	JOIN meals ON meal_plan_options.meal_id = meals.id
WHERE
	meal_plan_options.archived_at IS NULL
	AND meal_plan_options.belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id)
	AND meal_plan_events.id = sqlc.arg(meal_plan_event_id)
	AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
	AND meal_plans.archived_at IS NULL
	AND meal_plans.id = sqlc.arg(meal_plan_id)
	%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(mealPlanOptionsTableName, true, true, []string{}, "meal_plan_options.belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id)"),
					buildTotalCountSelect(mealPlanOptionsTableName, true, []string{}),
					buildFilterConditions(mealPlanOptionsTableName, true, false, "meal_plan_options.belongs_to_meal_plan_event = sqlc.arg(meal_plan_event_id)"),
					buildCursorLimitClause(mealPlanOptionsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetMealPlanOption",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
					strings.Join(fullSelectColumns, ",\n\t"),
					mealPlanOptionsTableName,
					mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn,
					mealPlansTableName, mealPlanEventsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
					mealsTableName, mealPlanOptionsTableName, mealIDColumn, mealsTableName, idColumn,
					mealPlanOptionsTableName, archivedAtColumn,

					mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventIDColumn,
					mealPlanOptionsTableName, idColumn, mealPlanOptionIDColumn,
					mealPlanEventsTableName, idColumn, mealPlanEventIDColumn,
					mealPlanEventsTableName, belongsToMealPlanColumn, mealPlanIDColumn,
					mealPlansTableName, archivedAtColumn,
					mealPlansTableName, idColumn, mealPlanIDColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetMealPlanOptionByID",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
					strings.Join(fullSelectColumns, ",\n\t"),
					mealPlanOptionsTableName,
					mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn,
					mealPlansTableName, mealPlanEventsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
					mealsTableName, mealPlanOptionsTableName, mealIDColumn, mealsTableName, idColumn,
					mealPlanOptionsTableName, archivedAtColumn,
					mealPlanOptionsTableName, idColumn, mealPlanOptionIDColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateMealPlanOption",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
					mealPlanOptionsTableName,
					strings.Join(applyToEach(filterForUpdate(mealPlanOptionsColumns, mealPlanOptionsChosenColumn, mealPlanOptionsTiebrokenColumn, belongsToMealPlanEventColumn), func(i int, s string) string {
						return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
					}), ",\n\t"),
					lastUpdatedAtColumn,
					currentTimeExpression,
					archivedAtColumn,
					belongsToMealPlanEventColumn, mealPlanEventIDColumn,
					idColumn, mealPlanOptionIDColumn,
				)),
			},
		}
	default:
		return nil
	}
}
