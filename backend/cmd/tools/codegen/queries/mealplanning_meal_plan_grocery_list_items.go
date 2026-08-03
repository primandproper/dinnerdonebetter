package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	mealPlanGroceryListItemsTableName = "meal_plan_grocery_list_items"

	mealPlanGroceryListItemIDColumn = "meal_plan_grocery_list_item_id"
)

func init() {
	registerTableName(mealPlanGroceryListItemsTableName)
}

var mealPlanGroceryListItemsColumns = []string{
	idColumn,
	"belongs_to_meal_plan",
	validIngredientColumn,
	"valid_measurement_unit",
	"minimum_quantity_needed",
	"maximum_quantity_needed",
	"quantity_purchased",
	"purchased_measurement_unit",
	"purchased_upc",
	"purchase_price",
	"status_explanation",
	"status",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	belongsToMealPlanOptionColumn,
	recipeIDColumn,
	recipeStepIDColumn,
	"ingredient_index",
	"option_index",
}

func buildMealPlanGroceryListItemsQueries(database string) []*Query {
	switch database {
	case postgres:

		insertColumns := filterForInsert(mealPlanGroceryListItemsColumns)

		fullSelectColumns := mergeColumns(
			applyToEach(filterFromSlice(mealPlanGroceryListItemsColumns, validIngredientColumn, validMeasurementUnitColumn), func(i int, s string) string {
				return fmt.Sprintf("%s.%s", mealPlanGroceryListItemsTableName, s)
			}),
			append(
				applyToEach(validIngredientsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_ingredient_%s", validIngredientsTableName, s, s)
				}),
				applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_measurement_unit_%s", validMeasurementUnitsTableName, s, s)
				})...,
			),
			2,
		)

		return []*Query{
			{
				// Compensation for the finalization saga's grocery list step. Deleting rather
				// than archiving, for the same reason as DeleteMealPlanTasks: the step
				// regenerates the whole list when it runs again, and an archived row would
				// leave a tombstone beside every regenerated item forever.
				//
				// The IDs come from the saga's own state, so a user's own additions to the
				// list — which the saga never recorded — are untouched.
				Annotation: QueryAnnotation{
					Name: "DeleteMealPlanGroceryListItems",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`DELETE FROM %s WHERE %s = ANY(sqlc.arg(ids)::text[]);`,
					mealPlanGroceryListItemsTableName,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveMealPlanGroceryListItem",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s IS NULL AND %s = sqlc.arg(%s);`,
					mealPlanGroceryListItemsTableName,
					archivedAtColumn,
					currentTimeExpression,
					archivedAtColumn,
					idColumn,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CreateMealPlanGroceryListItem",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					mealPlanGroceryListItemsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CheckMealPlanGroceryListItemExistence",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s = sqlc.arg(%s)
);`,
					mealPlanGroceryListItemsTableName, idColumn,
					mealPlanGroceryListItemsTableName,
					mealPlanGroceryListItemsTableName, archivedAtColumn,
					mealPlanGroceryListItemsTableName, idColumn, mealPlanGroceryListItemIDColumn,
					mealPlanGroceryListItemsTableName, belongsToMealPlanColumn, mealPlanIDColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetMealPlanGroceryListItemsForMealPlan",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	%s
GROUP BY %s.%s,
	%s.%s,
	%s.%s,
	%s.%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(
						mealPlanGroceryListItemsTableName,
						true,
						true,
						[]string{
							fmt.Sprintf("%s ON %s.%s = %s.%s", mealPlansTableName, mealPlanGroceryListItemsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn),
							fmt.Sprintf("%s ON %s.%s = %s.%s", validIngredientsTableName, mealPlanGroceryListItemsTableName, validIngredientColumn, validIngredientsTableName, idColumn),
							fmt.Sprintf("%s ON %s.%s = %s.%s", validMeasurementUnitsTableName, mealPlanGroceryListItemsTableName, validMeasurementUnitColumn, validMeasurementUnitsTableName, idColumn),
						},
						fmt.Sprintf("%s.%s IS NULL", validMeasurementUnitsTableName, archivedAtColumn),
						fmt.Sprintf("%s.%s IS NULL", validIngredientsTableName, archivedAtColumn),
						fmt.Sprintf("%s.%s IS NULL", mealPlansTableName, archivedAtColumn),
						fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanGroceryListItemsTableName, belongsToMealPlanColumn, mealPlanIDColumn),
						"(meal_plan_grocery_list_items.belongs_to_meal_plan_option IS NULL OR NOT EXISTS (SELECT 1 FROM meal_plan_options o JOIN meal_plan_events e ON o.belongs_to_meal_plan_event = e.id WHERE o.id = meal_plan_grocery_list_items.belongs_to_meal_plan_option AND e.archived_at IS NOT NULL))",
					),
					buildTotalCountSelect(
						mealPlanGroceryListItemsTableName,
						true,
						[]string{
							fmt.Sprintf("%s ON %s.%s = %s.%s", mealPlansTableName, mealPlanGroceryListItemsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn),
							fmt.Sprintf("%s ON %s.%s = %s.%s", validIngredientsTableName, mealPlanGroceryListItemsTableName, validIngredientColumn, validIngredientsTableName, idColumn),
							fmt.Sprintf("%s ON %s.%s = %s.%s", validMeasurementUnitsTableName, mealPlanGroceryListItemsTableName, validMeasurementUnitColumn, validMeasurementUnitsTableName, idColumn),
						},
						fmt.Sprintf("%s.%s IS NULL", validMeasurementUnitsTableName, archivedAtColumn),
						fmt.Sprintf("%s.%s IS NULL", validIngredientsTableName, archivedAtColumn),
						fmt.Sprintf("%s.%s IS NULL", mealPlansTableName, archivedAtColumn),
						fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanGroceryListItemsTableName, belongsToMealPlanColumn, mealPlanIDColumn),
						"(meal_plan_grocery_list_items.belongs_to_meal_plan_option IS NULL OR NOT EXISTS (SELECT 1 FROM meal_plan_options o JOIN meal_plan_events e ON o.belongs_to_meal_plan_event = e.id WHERE o.id = meal_plan_grocery_list_items.belongs_to_meal_plan_option AND e.archived_at IS NOT NULL))",
					),
					mealPlanGroceryListItemsTableName,
					mealPlansTableName, mealPlanGroceryListItemsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
					validIngredientsTableName, mealPlanGroceryListItemsTableName, validIngredientColumn, validIngredientsTableName, idColumn,
					validMeasurementUnitsTableName, mealPlanGroceryListItemsTableName, validMeasurementUnitColumn, validMeasurementUnitsTableName, idColumn,
					mealPlanGroceryListItemsTableName, archivedAtColumn,
					buildFilterConditions(
						mealPlanGroceryListItemsTableName,
						true,
						true,
						fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanGroceryListItemsTableName, belongsToMealPlanColumn, mealPlanIDColumn),
						fmt.Sprintf("%s.%s IS NULL", validMeasurementUnitsTableName, archivedAtColumn),
						fmt.Sprintf("%s.%s IS NULL", validIngredientsTableName, archivedAtColumn),
						fmt.Sprintf("%s.%s IS NULL", mealPlansTableName, archivedAtColumn),
						"(meal_plan_grocery_list_items.belongs_to_meal_plan_option IS NULL OR NOT EXISTS (SELECT 1 FROM meal_plan_options o JOIN meal_plan_events e ON o.belongs_to_meal_plan_event = e.id WHERE o.id = meal_plan_grocery_list_items.belongs_to_meal_plan_option AND e.archived_at IS NOT NULL))",
					),
					mealPlanGroceryListItemsTableName, idColumn,
					validIngredientsTableName, idColumn,
					validMeasurementUnitsTableName, idColumn,
					mealPlansTableName, idColumn,
					buildCursorLimitClause(mealPlanGroceryListItemsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetMealPlanGroceryListItem",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s);`,
					strings.Join(fullSelectColumns, ",\n\t"),
					mealPlanGroceryListItemsTableName,
					mealPlansTableName, mealPlanGroceryListItemsTableName, belongsToMealPlanColumn, mealPlansTableName, idColumn,
					validIngredientsTableName, mealPlanGroceryListItemsTableName, validIngredientColumn, validIngredientsTableName, idColumn,
					validMeasurementUnitsTableName, mealPlanGroceryListItemsTableName, validMeasurementUnitColumn, validMeasurementUnitsTableName, idColumn,
					mealPlanGroceryListItemsTableName, archivedAtColumn,
					validMeasurementUnitsTableName, archivedAtColumn,
					validIngredientsTableName, archivedAtColumn,
					mealPlanGroceryListItemsTableName, idColumn, mealPlanGroceryListItemIDColumn,
					mealPlanGroceryListItemsTableName, belongsToMealPlanColumn, mealPlanIDColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateMealPlanGroceryListItem",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
					mealPlanGroceryListItemsTableName,
					strings.Join(applyToEach(filterForUpdate(mealPlanGroceryListItemsColumns, belongsToUserColumn), func(i int, s string) string {
						return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
					}), ",\n\t"),
					lastUpdatedAtColumn, currentTimeExpression,
					archivedAtColumn,
					idColumn, idColumn,
				)),
			},
		}
	default:
		return nil
	}
}
