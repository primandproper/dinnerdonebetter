package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	mealsTableName = "meals"

	mealIDColumn = "meal_id"
)

var mealsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	"min_estimated_portions",
	"max_estimated_portions",
	"eligible_for_meal_plans",
	lastIndexedAtColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	createdByUserColumn,
}

// mealComponentRecipeColumns are the recipe columns GetMeals and SearchForMeals select
// alongside each component. They are the RecipeSummary's columns, which is what a meal
// list or search response carries per component -- selecting them here is what lets those
// queries build a component without a getRecipe hydration each.
//
// recipes.id is left out because meal_components.recipe_id already arrives as
// component_recipe_id and holds the same value; last_indexed_at and last_validated_at are
// left out because the domain Recipe has no field for either.
var mealComponentRecipeColumns = filterFromSlice(recipesColumns, idColumn, lastIndexedAtColumn, lastValidatedAtColumn)

func buildMealsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := append(
			applyToEach(mealsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s", mealsTableName, s)
			}),
			applyToEach(mealComponentsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as component_%s", mealComponentsTableName, s, s)
			})...,
		)

		// The list and search selects carry each component's recipe as well, so the
		// repository can build the component's RecipeSummary from the row.
		summarySelectColumns := slices.Concat(
			fullSelectColumns,
			applyToEach(mealComponentRecipeColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as component_recipe_%s", recipesTableName, s, s)
			}),
		)

		return slices.Concat(
			pgGen.StandardCRUD(mealsTableName, mealsColumns,
				querygen.WithEntity("Meal", "Meals"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.GetQuery, querygen.ListQuery, querygen.MarkAsIndexedQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveMeal",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s IS NULL AND %s = sqlc.arg(%s) AND %s = sqlc.arg(%s);`,
						mealsTableName,
						archivedAtColumn,
						querygen.NowExpression,
						archivedAtColumn,
						createdByUserColumn,
						createdByUserColumn,
						idColumn,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetMealsByCreatorAndName",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
		AND %s.%s IS NULL
		AND EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.%s AND %s.%s IS NULL)
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
ORDER BY %s.%s ASC, %s.%s ASC;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						mealsTableName,
						mealComponentsTableName, mealComponentsTableName, belongsToMealColumn, mealsTableName, idColumn,
						mealComponentsTableName, archivedAtColumn,
						recipesTableName, recipesTableName, idColumn, mealComponentsTableName, recipeIDColumn,
						recipesTableName, archivedAtColumn,
						mealsTableName, archivedAtColumn,
						mealsTableName, createdByUserColumn, createdByUserColumn,
						mealsTableName, nameColumn, nameColumn,
						mealsTableName, idColumn,
						mealComponentsTableName, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetMealsNeedingIndexing",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
	AND (
		%s.%s IS NULL
		OR %s.%s < %s - '24 hours'::INTERVAL
	);`,
						mealsTableName, idColumn,
						mealsTableName,
						mealsTableName, archivedAtColumn,
						mealsTableName, lastIndexedAtColumn,
						mealsTableName, lastIndexedAtColumn,
						querygen.NowExpression,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetMeal",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
		AND %s.%s IS NULL
		AND EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.%s AND %s.%s IS NULL)
WHERE %s.%s IS NULL
  AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						mealsTableName,
						mealComponentsTableName, mealComponentsTableName, belongsToMealColumn, mealsTableName, idColumn,
						mealComponentsTableName, archivedAtColumn,
						recipesTableName, recipesTableName, idColumn, mealComponentsTableName, recipeIDColumn,
						recipesTableName, archivedAtColumn,
						mealsTableName, archivedAtColumn,
						mealsTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetMeals",
						Type: ManyType,
					},
					// Both joins are LEFT joins, and the recipes one replaces the EXISTS
					// that used to hang off the meal_components join condition. A
					// component whose recipe is archived now arrives with null recipe
					// columns instead of dropping out of the result, and the repository
					// skips it -- which is what it already did when the getRecipe this
					// join replaces came back with no rows.
					//
					// Chaining the joins rather than nesting the recipes one inside the
					// LEFT JOIN's right operand is deliberate: sqlc infers nullability
					// from the join tree, and it reads a parenthesized operand as
					// non-nullable, which would type a component-less meal's columns as
					// bare strings and fail to scan.
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	LEFT JOIN %s ON %s.%s=%s.%s AND %s.%s IS NULL
	LEFT JOIN %s ON %s.%s = %s.%s AND %s.%s IS NULL
WHERE %s
%s;`,
						strings.Join(summarySelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(mealsTableName, mealsColumns, []string{}),
						pgGen.TotalCountSelect(mealsTableName, mealsColumns, []string{}),
						mealsTableName,
						mealComponentsTableName,
						mealComponentsTableName,
						belongsToMealColumn,
						mealsTableName,
						idColumn,
						mealComponentsTableName,
						archivedAtColumn,
						recipesTableName, recipesTableName, idColumn, mealComponentsTableName, recipeIDColumn,
						recipesTableName, archivedAtColumn,
						pgGen.FilterConditions(mealsTableName, mealsColumns, querygen.Ascending),
						pgGen.CursorLimitClause(mealsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetMealsWithIDs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
		AND %s.%s IS NULL
		AND EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.%s AND %s.%s IS NULL)
WHERE %s.%s IS NULL
  AND %s.%s = ANY(sqlc.arg(ids)::text[])
ORDER BY %s.%s ASC;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						mealsTableName,
						mealComponentsTableName, mealComponentsTableName, belongsToMealColumn, mealsTableName, idColumn,
						mealComponentsTableName, archivedAtColumn,
						recipesTableName, recipesTableName, idColumn, mealComponentsTableName, recipeIDColumn,
						recipesTableName, archivedAtColumn,
						mealsTableName, archivedAtColumn,
						mealsTableName, idColumn,
						mealsTableName, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetMealsCreatedByUser",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	LEFT JOIN %s ON %s.%s=%s.%s AND %s.%s IS NULL
		AND EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.%s AND %s.%s IS NULL)
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(mealsTableName, mealsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealsTableName, createdByUserColumn, createdByUserColumn)),
						pgGen.TotalCountSelect(mealsTableName, mealsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealsTableName, createdByUserColumn, createdByUserColumn)),
						mealsTableName,
						mealComponentsTableName,
						mealComponentsTableName,
						belongsToMealColumn,
						mealsTableName,
						idColumn,
						mealComponentsTableName,
						archivedAtColumn,
						recipesTableName,
						recipesTableName,
						idColumn,
						mealComponentsTableName,
						recipeIDColumn,
						recipesTableName,
						archivedAtColumn,
						pgGen.FilterConditions(mealsTableName, mealsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealsTableName, createdByUserColumn, createdByUserColumn),
						),
						pgGen.CursorLimitClause(mealsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchForMeals",
						Type: ManyType,
					},
					// The EXISTS this join replaces filtered on exactly the same rows;
					// as an inner join it also carries the recipe's columns out, which
					// is what spares the repository a getRecipe per component.
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
		AND %s.%s IS NULL
	JOIN %s ON %s.%s = %s.%s AND %s.%s IS NULL
WHERE %s
%s;`,
						strings.Join(summarySelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(mealsTableName, mealsColumns, []string{}),
						pgGen.TotalCountSelect(mealsTableName, mealsColumns, []string{}),
						mealsTableName,
						mealComponentsTableName,
						mealComponentsTableName,
						belongsToMealColumn,
						mealsTableName,
						idColumn,
						mealComponentsTableName,
						archivedAtColumn,
						recipesTableName, recipesTableName, idColumn, mealComponentsTableName, recipeIDColumn,
						recipesTableName, archivedAtColumn,
						pgGen.FilterConditions(mealsTableName, mealsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s %s", mealsTableName, nameColumn, buildILIKEForArgument("query")),
						),
						pgGen.CursorLimitClause(mealsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "MarkMealsAsIndexed",
						Type: ExecRowsType,
					},
					// Bulk, and named for what querygen emits, so migrating this table to
					// StandardCRUD deletes this declaration without renaming anything.
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s = ANY(sqlc.arg(%s)::text[]);`,
						mealsTableName,
						lastIndexedAtColumn,
						querygen.NowExpression,
						idColumn,
						idsArg,
					)),
				},
			},
		)
	default:
		return nil
	}
}
