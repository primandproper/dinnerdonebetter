package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	mealsTableName = "meals"

	mealIDColumn = "meal_id"
)

func init() {
	registerTableName(mealsTableName)
}

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

		return slices.Concat(
			querygen.StandardCRUD(mealsTableName, mealsColumns,
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
						currentTimeExpression,
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
						currentTimeExpression,
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
						querygen.FilterCountSelect(mealsTableName, mealsColumns, []string{}),
						querygen.TotalCountSelect(mealsTableName, mealsColumns, []string{}),
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
						querygen.FilterConditions(mealsTableName, mealsColumns),
						querygen.CursorLimitClause(mealsTableName),
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
						querygen.FilterCountSelect(mealsTableName, mealsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealsTableName, createdByUserColumn, createdByUserColumn)),
						querygen.TotalCountSelect(mealsTableName, mealsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealsTableName, createdByUserColumn, createdByUserColumn)),
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
						querygen.FilterConditions(mealsTableName, mealsColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealsTableName, createdByUserColumn, createdByUserColumn),
						),
						querygen.CursorLimitClause(mealsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchForMeals",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
		AND %s.%s IS NULL
		AND EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.%s AND %s.%s IS NULL)
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						querygen.FilterCountSelect(mealsTableName, mealsColumns, []string{}),
						querygen.TotalCountSelect(mealsTableName, mealsColumns, []string{}),
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
						querygen.FilterConditions(mealsTableName, mealsColumns,
							fmt.Sprintf("%s.%s %s", mealsTableName, nameColumn, buildILIKEForArgument("query")),
						),
						querygen.CursorLimitClause(mealsTableName),
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
						currentTimeExpression,
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
