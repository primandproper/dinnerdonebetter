package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

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

		return slices.Concat(
			querygen.StandardCRUD(mealListItemsTableName, mealListItemsColumns,
				querygen.WithEntity("MealListItem", "MealListItems"),
				querygen.WithOwnership("belongs_to_meal_list"),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
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
						Name: "GetMealListItems",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(applyToEach(mealListItemsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", mealListItemsTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(mealListItemsTableName, mealListItemsColumns, []string{}, fmt.Sprintf("%s.belongs_to_meal_list = sqlc.arg(%s)", mealListItemsTableName, mealListIDColumn), fmt.Sprintf("EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.belongs_to_meal_list AND %s.%s = sqlc.arg(%s))", mealListsTableName, mealListsTableName, idColumn, mealListItemsTableName, mealListsTableName, belongsToUserColumn, belongsToUserColumn)),
						querygen.TotalCountSelect(mealListItemsTableName, mealListItemsColumns, []string{}),
						mealListItemsTableName,
						querygen.FilterConditions(mealListItemsTableName, mealListItemsColumns,
							fmt.Sprintf("%s.belongs_to_meal_list = sqlc.arg(%s)", mealListItemsTableName, mealListIDColumn),
							fmt.Sprintf("EXISTS (SELECT 1 FROM %s WHERE %s.%s = %s.belongs_to_meal_list AND %s.%s = sqlc.arg(%s))", mealListsTableName, mealListsTableName, idColumn, mealListItemsTableName, mealListsTableName, belongsToUserColumn, belongsToUserColumn),
						),
						querygen.CursorLimitClause(mealListItemsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
