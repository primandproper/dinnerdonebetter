package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	mealListsTableName = "meal_lists"
)

func init() {
	registerTableName(mealListsTableName)
}

var mealListsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	belongsToUserColumn,
}

func buildMealListsQueries(database string) []*Query {
	switch database {
	case postgres:

		insertColumns := filterForInsert(mealListsColumns)
		fullSelectColumns := mergeColumns(
			applyToEach(mealListsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s", mealListsTableName, s)
			}),
			applyToEach(mealListItemsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as meal_list_item_%s", mealListItemsTableName, s, s)
			}),
			2,
		)

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveMealList",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s IS NULL AND %s = sqlc.arg(%s) AND %s = sqlc.arg(%s);`,
					mealListsTableName,
					archivedAtColumn, currentTimeExpression,
					archivedAtColumn,
					belongsToUserColumn, belongsToUserColumn,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CreateMealList",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					mealListsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetMealLists",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	LEFT JOIN %s ON %s.%s = %s.%s AND %s.%s IS NULL
	WHERE %s.%s IS NULL
	%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(mealListsTableName, true, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealListsTableName, belongsToUserColumn, belongsToUserColumn)),
					buildTotalCountSelect(mealListsTableName, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealListsTableName, belongsToUserColumn, belongsToUserColumn)),
					mealListsTableName,
					mealListItemsTableName, mealListItemsTableName, "belongs_to_meal_list", mealListsTableName, idColumn, mealListItemsTableName, archivedAtColumn,
					mealListsTableName, archivedAtColumn,
					buildFilterConditions(mealListsTableName, true, false, fmt.Sprintf("%s.%s = sqlc.arg(%s)", mealListsTableName, belongsToUserColumn, belongsToUserColumn)),
					buildCursorLimitClause(mealListsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateMealList",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
					mealListsTableName,
					strings.Join(applyToEach(filterForUpdate(mealListsColumns, belongsToUserColumn), func(i int, s string) string {
						return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
					}), ",\n\t"),
					lastUpdatedAtColumn, currentTimeExpression,
					archivedAtColumn,
					belongsToUserColumn, belongsToUserColumn,
					idColumn, idColumn,
				)),
			},
		}
	default:
		return nil
	}
}
