package main

import (
	"fmt"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	mealPlanRecipeOptionSelectionsTableName = "meal_plan_recipe_option_selections"

	mealPlanRecipeOptionSelectionIDColumn = "meal_plan_recipe_option_selection_id"
	selectedOptionIndexColumn             = "selected_option_index"
	selectionTypeColumn                   = "selection_type"
	ingredientIndexColumn                 = "ingredient_index"
)

func init() {
	registerTableName(mealPlanRecipeOptionSelectionsTableName)
}

var mealPlanRecipeOptionSelectionsColumns = []string{
	idColumn,
	belongsToMealPlanOptionColumn,
	recipeIDColumn,
	recipeStepIDColumn,
	ingredientIndexColumn,
	selectedOptionIndexColumn,
	selectionTypeColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildMealPlanRecipeOptionSelectionsQueries(database string) []*Query {
	switch database {
	case postgres:

		insertColumns := querygen.ForInsert(mealPlanRecipeOptionSelectionsColumns)

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "CreateMealPlanRecipeOptionSelection",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
) ON CONFLICT (%s, %s, %s, %s) DO UPDATE SET
	%s = EXCLUDED.%s,
	%s = %s;`,
					mealPlanRecipeOptionSelectionsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
					belongsToMealPlanOptionColumn,
					recipeStepIDColumn,
					ingredientIndexColumn,
					selectionTypeColumn,
					selectedOptionIndexColumn,
					selectedOptionIndexColumn,
					lastUpdatedAtColumn,
					querygen.NowExpression,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetMealPlanRecipeOptionSelection",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
					strings.Join(applyToEach(mealPlanRecipeOptionSelectionsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", mealPlanRecipeOptionSelectionsTableName, s)
					}), ",\n\t"),
					mealPlanRecipeOptionSelectionsTableName,
					belongsToMealPlanOptionColumn, mealPlanOptionIDColumn,
					recipeStepIDColumn, recipeStepIDColumn,
					ingredientIndexColumn, ingredientIndexColumn,
					selectionTypeColumn, selectionTypeColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetMealPlanRecipeOptionSelectionsForMealPlanOption",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
					strings.Join(applyToEach(mealPlanRecipeOptionSelectionsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", mealPlanRecipeOptionSelectionsTableName, s)
					}), ",\n\t"),
					pgGen.FilterCountSelect(mealPlanRecipeOptionSelectionsTableName, mealPlanRecipeOptionSelectionsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanRecipeOptionSelectionsTableName, belongsToMealPlanOptionColumn, mealPlanOptionIDColumn)),
					pgGen.TotalCountSelect(mealPlanRecipeOptionSelectionsTableName, mealPlanRecipeOptionSelectionsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanRecipeOptionSelectionsTableName, belongsToMealPlanOptionColumn, mealPlanOptionIDColumn)),
					mealPlanRecipeOptionSelectionsTableName,
					pgGen.FilterConditions(mealPlanRecipeOptionSelectionsTableName, mealPlanRecipeOptionSelectionsColumns,
						fmt.Sprintf("%s = sqlc.arg(%s)", belongsToMealPlanOptionColumn, mealPlanOptionIDColumn),
						fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanRecipeOptionSelectionsTableName, belongsToMealPlanOptionColumn, mealPlanOptionIDColumn),
					),
					pgGen.CursorLimitClause(mealPlanRecipeOptionSelectionsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetMealPlanRecipeOptionSelectionsForMealPlan",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE %s
%s;`,
					strings.Join(applyToEach(mealPlanRecipeOptionSelectionsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", mealPlanRecipeOptionSelectionsTableName, s)
					}), ",\n\t"),
					pgGen.FilterCountSelect(mealPlanRecipeOptionSelectionsTableName, mealPlanRecipeOptionSelectionsColumns,
						[]string{
							fmt.Sprintf("JOIN %s ON %s.%s = %s.%s", mealPlanOptionsTableName, mealPlanRecipeOptionSelectionsTableName, belongsToMealPlanOptionColumn, mealPlanOptionsTableName, idColumn),
							fmt.Sprintf("JOIN %s ON %s.%s = %s.%s", mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn),
						},
						fmt.Sprintf("%s.%s = sqlc.arg(%s) AND %s.%s IS NULL AND %s.%s IS NULL", mealPlanEventsTableName, belongsToMealPlanColumn, mealPlanIDColumn, mealPlanOptionsTableName, archivedAtColumn, mealPlanEventsTableName, archivedAtColumn)),
					pgGen.TotalCountSelect(mealPlanRecipeOptionSelectionsTableName, mealPlanRecipeOptionSelectionsColumns,
						[]string{
							fmt.Sprintf("JOIN %s ON %s.%s = %s.%s", mealPlanOptionsTableName, mealPlanRecipeOptionSelectionsTableName, belongsToMealPlanOptionColumn, mealPlanOptionsTableName, idColumn),
							fmt.Sprintf("JOIN %s ON %s.%s = %s.%s", mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn),
						},
						fmt.Sprintf("%s.%s = sqlc.arg(%s) AND %s.%s IS NULL AND %s.%s IS NULL", mealPlanEventsTableName, belongsToMealPlanColumn, mealPlanIDColumn, mealPlanOptionsTableName, archivedAtColumn, mealPlanEventsTableName, archivedAtColumn)),
					mealPlanRecipeOptionSelectionsTableName,
					mealPlanOptionsTableName, mealPlanRecipeOptionSelectionsTableName, belongsToMealPlanOptionColumn, mealPlanOptionsTableName, idColumn,
					mealPlanEventsTableName, mealPlanOptionsTableName, belongsToMealPlanEventColumn, mealPlanEventsTableName, idColumn,
					pgGen.FilterConditions(mealPlanRecipeOptionSelectionsTableName, mealPlanRecipeOptionSelectionsColumns,
						fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealPlanEventsTableName, belongsToMealPlanColumn, mealPlanIDColumn),
						fmt.Sprintf("%s.%s IS NULL", mealPlanOptionsTableName, archivedAtColumn),
						fmt.Sprintf("%s.%s IS NULL", mealPlanEventsTableName, archivedAtColumn),
					),
					pgGen.CursorLimitClause(mealPlanRecipeOptionSelectionsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateMealPlanRecipeOptionSelection",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
					mealPlanRecipeOptionSelectionsTableName,
					strings.Join(applyToEach(querygen.ForUpdate(mealPlanRecipeOptionSelectionsColumns, belongsToMealPlanOptionColumn, recipeStepIDColumn, ingredientIndexColumn, selectionTypeColumn), func(i int, s string) string {
						return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
					}), ",\n\t"),
					lastUpdatedAtColumn, querygen.NowExpression,
					belongsToMealPlanOptionColumn, mealPlanOptionIDColumn,
					recipeStepIDColumn, recipeStepIDColumn,
					ingredientIndexColumn, ingredientIndexColumn,
					selectionTypeColumn, selectionTypeColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveMealPlanRecipeOptionSelection",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
					mealPlanRecipeOptionSelectionsTableName,
					archivedAtColumn, querygen.NowExpression,
					belongsToMealPlanOptionColumn, mealPlanOptionIDColumn,
					recipeStepIDColumn, recipeStepIDColumn,
					ingredientIndexColumn, ingredientIndexColumn,
					selectionTypeColumn, selectionTypeColumn,
				)),
			},
		}
	default:
		return nil
	}
}
