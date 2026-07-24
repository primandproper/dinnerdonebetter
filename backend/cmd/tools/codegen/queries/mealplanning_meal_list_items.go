package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	mealListItemsTableName = "meal_list_items"

	mealListIDColumn = "meal_list_id"
)

func init() {
	registerTableName(mealListItemsTableName)
}

var mealListItemsColumns = []string{
	idColumn,
	mealIDColumn,
	notesColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	"belongs_to_meal_list",
}

func buildMealListItemsQueries(database string) []*Query {
	switch database {
	case postgres:

		insertColumns := filterForInsert(mealListItemsColumns)

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "CheckMealInMealList",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.belongs_to_meal_list = sqlc.arg(%s)
		AND %s.%s = sqlc.arg(%s)
);`,
					mealListItemsTableName, idColumn,
					mealListItemsTableName,
					mealListItemsTableName, archivedAtColumn,
					mealListItemsTableName, "belongs_to_meal_list",
					mealListItemsTableName, mealIDColumn, mealIDColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveMealListItem",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s IS NULL AND %s = sqlc.arg(%s) AND %s = sqlc.arg(%s);`,
					mealListItemsTableName,
					archivedAtColumn, currentTimeExpression,
					archivedAtColumn,
					"belongs_to_meal_list", "belongs_to_meal_list",
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CreateMealListItem",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					mealListItemsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetMealListItems",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	%s
	AND %s.belongs_to_meal_list = sqlc.arg(%s)
%s;`,
					strings.Join(applyToEach(mealListItemsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", mealListItemsTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(mealListItemsTableName, true, true, []string{}, fmt.Sprintf("%s.belongs_to_meal_list = sqlc.arg(%s)", mealListItemsTableName, mealListIDColumn), fmt.Sprintf("EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.belongs_to_meal_list AND %s.%s = sqlc.arg(%s))", mealListsTableName, mealListsTableName, idColumn, mealListItemsTableName, mealListsTableName, belongsToUserColumn, belongsToUserColumn)),
					buildTotalCountSelect(mealListItemsTableName, true, []string{}),
					mealListItemsTableName,
					mealListItemsTableName, archivedAtColumn,
					buildFilterConditions(mealListItemsTableName, true, false, fmt.Sprintf("%s.belongs_to_meal_list = sqlc.arg(%s)", mealListItemsTableName, mealListIDColumn), fmt.Sprintf("EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.belongs_to_meal_list AND %s.%s = sqlc.arg(%s))", mealListsTableName, mealListsTableName, idColumn, mealListItemsTableName, mealListsTableName, belongsToUserColumn, belongsToUserColumn)),
					mealListItemsTableName, mealListIDColumn,
					buildCursorLimitClause(mealListItemsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateMealListItem",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
					mealListItemsTableName,
					strings.Join(applyToEach(filterForUpdate(mealListItemsColumns, "belongs_to_meal_list"), func(i int, s string) string {
						return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
					}), ",\n\t"),
					lastUpdatedAtColumn, currentTimeExpression,
					archivedAtColumn,
					"belongs_to_meal_list", "belongs_to_meal_list",
					idColumn, idColumn,
				)),
			},
		}
	default:
		return nil
	}
}
