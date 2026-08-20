package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	recipeListsTableName = "recipe_lists"
)

func init() {
	registerTableName(recipeListsTableName)
}

var recipeListsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	belongsToUserColumn,
}

func buildRecipeListsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			applyToEach(recipeListsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s", recipeListsTableName, s)
			}),
			applyToEach(recipeListItemsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as recipe_list_item_%s", recipeListItemsTableName, s, s)
			}),
			2,
		)

		return slices.Concat(
			pgGen.StandardCRUD(recipeListsTableName, recipeListsColumns,
				querygen.WithEntity("RecipeList", "RecipeLists"),
				querygen.WithOwnership(belongsToUserColumn),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeLists",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	LEFT JOIN %s ON %s.%s = %s.%s AND %s.%s IS NULL
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(recipeListsTableName, recipeListsColumns, []string{}),
						pgGen.TotalCountSelect(recipeListsTableName, recipeListsColumns, []string{}),
						recipeListsTableName,
						recipeListItemsTableName, recipeListItemsTableName, "belongs_to_recipe_list", recipeListsTableName, idColumn, recipeListItemsTableName, archivedAtColumn,
						pgGen.FilterConditions(recipeListsTableName, recipeListsColumns),
						pgGen.CursorLimitClause(recipeListsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
