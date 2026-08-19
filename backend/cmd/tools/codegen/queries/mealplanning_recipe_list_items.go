package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	recipeListItemsTableName = "recipe_list_items"

	recipeListIDColumn = "recipe_list_id"
)

func init() {
	registerTableName(recipeListItemsTableName)
}

var recipeListItemsColumns = []string{
	idColumn,
	recipeIDColumn,
	notesColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	"belongs_to_recipe_list",
}

func buildRecipeListItemsQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			querygen.StandardCRUD(recipeListItemsTableName, recipeListItemsColumns,
				querygen.WithEntity("RecipeListItem", "RecipeListItems"),
				querygen.WithOwnership("belongs_to_recipe_list"),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeListItems",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	%s
	AND %s.belongs_to_recipe_list = sqlc.arg(%s)
%s;`,
						strings.Join(applyToEach(recipeListItemsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", recipeListItemsTableName, s)
						}), ",\n\t"),
						buildFilterCountSelect(recipeListItemsTableName, true, true, []string{}, fmt.Sprintf("%s.belongs_to_recipe_list = sqlc.arg(%s)", recipeListItemsTableName, recipeListIDColumn)),
						buildTotalCountSelect(recipeListItemsTableName, true, []string{}),
						recipeListItemsTableName,
						recipeListItemsTableName, archivedAtColumn,
						buildFilterConditions(recipeListItemsTableName, true, false, fmt.Sprintf("%s.belongs_to_recipe_list = sqlc.arg(%s)", recipeListItemsTableName, recipeListIDColumn)),
						recipeListItemsTableName, recipeListIDColumn,
						buildCursorLimitClause(recipeListItemsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
