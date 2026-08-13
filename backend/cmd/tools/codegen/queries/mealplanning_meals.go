package main

import (
	"fmt"
	"strings"

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

		insertColumns := filterForInsert(mealsColumns)

		fullSelectColumns := append(
			applyToEach(mealsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s", mealsTableName, s)
			}),
			applyToEach(mealComponentsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as component_%s", mealComponentsTableName, s, s)
			})...,
		)

		return []*Query{
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
					Name: "CreateMeal",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					mealsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CheckMealExistence",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
					mealsTableName, idColumn,
					mealsTableName,
					mealsTableName, archivedAtColumn,
					mealsTableName, idColumn, idColumn,
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
					Name: "ScanMealIDsForReindex",
					Type: ManyType,
				},
				Content: buildReindexScanQuery(mealsTableName),
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
WHERE
	%s.%s IS NULL
	%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(mealsTableName, true, true, []string{}),
					buildTotalCountSelect(mealsTableName, true, []string{}),
					mealsTableName,
					mealComponentsTableName, mealComponentsTableName, belongsToMealColumn, mealsTableName, idColumn,
					mealComponentsTableName, archivedAtColumn,
					recipesTableName, recipesTableName, idColumn, mealComponentsTableName, recipeIDColumn,
					recipesTableName, archivedAtColumn,
					mealsTableName, archivedAtColumn,
					buildFilterConditions(
						mealsTableName,
						true,
						true,
					),
					buildCursorLimitClause(mealsTableName),
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
WHERE
	%s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(mealsTableName, true, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealsTableName, createdByUserColumn, createdByUserColumn)),
					buildTotalCountSelect(mealsTableName, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealsTableName, createdByUserColumn, createdByUserColumn)),
					mealsTableName,
					mealComponentsTableName, mealComponentsTableName, belongsToMealColumn, mealsTableName, idColumn,
					mealComponentsTableName, archivedAtColumn,
					recipesTableName, recipesTableName, idColumn, mealComponentsTableName, recipeIDColumn,
					recipesTableName, archivedAtColumn,
					mealsTableName, archivedAtColumn,
					mealsTableName, createdByUserColumn, createdByUserColumn,
					buildFilterConditions(mealsTableName, true, true, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealsTableName, createdByUserColumn, createdByUserColumn)),
					buildCursorLimitClause(mealsTableName),
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
WHERE
	%s.%s IS NULL
	AND %s.%s %s
	%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(mealsTableName, true, true, []string{}),
					buildTotalCountSelect(mealsTableName, true, []string{}),
					mealsTableName,
					mealComponentsTableName, mealComponentsTableName, belongsToMealColumn, mealsTableName, idColumn,
					mealComponentsTableName, archivedAtColumn,
					recipesTableName, recipesTableName, idColumn, mealComponentsTableName, recipeIDColumn,
					recipesTableName, archivedAtColumn,
					mealsTableName, archivedAtColumn,
					mealsTableName, nameColumn, buildILIKEForArgument("query"),
					buildFilterConditions(
						mealsTableName,
						true,
						true,
					),
					buildCursorLimitClause(mealsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateMealLastIndexedAt",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s = sqlc.arg(%s) AND %s IS NULL;`,
					mealsTableName,
					lastIndexedAtColumn,
					currentTimeExpression,
					idColumn,
					idColumn,
					archivedAtColumn,
				)),
			},
		}
	default:
		return nil
	}
}
