package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

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

		fullSelectColumns := mergeColumns(
			applyToEach(mealListsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s", mealListsTableName, s)
			}),
			applyToEach(mealListItemsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as meal_list_item_%s", mealListItemsTableName, s, s)
			}),
			2,
		)

		return slices.Concat(
			querygen.StandardCRUD(mealListsTableName, mealListsColumns,
				querygen.WithEntity("MealList", "MealLists"),
				querygen.WithOwnership(belongsToUserColumn),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
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
			},
		)
	default:
		return nil
	}
}
