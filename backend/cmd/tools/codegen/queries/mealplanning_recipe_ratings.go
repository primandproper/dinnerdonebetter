package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	recipeRatingsTableName = "recipe_ratings"
)

func init() {
	registerTableName(recipeRatingsTableName)
}

var recipeRatingsColumns = []string{
	idColumn,
	belongsToRecipeColumn,
	"taste",
	"difficulty",
	"cleanup",
	"instructions",
	"overall",
	notesColumn,
	createdByUserColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildRecipeRatingsQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			querygen.StandardCRUD(recipeRatingsTableName, recipeRatingsColumns,
				querygen.WithEntity("RecipeRating", "RecipeRatings"),
				querygen.WithImmutable(createdByUserColumn),
				querygen.WithOmitted(querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeRatingsForRecipe",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
GROUP BY %s.%s
%s;`,
						strings.Join(applyToEach(recipeRatingsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", recipeRatingsTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(recipeRatingsTableName, recipeRatingsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, belongsToRecipeColumn, belongsToRecipeColumn)),
						querygen.TotalCountSelect(recipeRatingsTableName, recipeRatingsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, belongsToRecipeColumn, belongsToRecipeColumn)),
						recipeRatingsTableName,
						querygen.FilterConditions(recipeRatingsTableName, recipeRatingsColumns,
							fmt.Sprintf("%s.%s IS NULL AND\n\t%s.%s = sqlc.arg(%s)", recipeRatingsTableName, archivedAtColumn, recipeRatingsTableName, belongsToRecipeColumn, belongsToRecipeColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, belongsToRecipeColumn, belongsToRecipeColumn),
						),
						recipeRatingsTableName,
						idColumn,
						querygen.CursorLimitClause(recipeRatingsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeRatingsForUser",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
GROUP BY %s.%s
%s;`,
						strings.Join(applyToEach(recipeRatingsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", recipeRatingsTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(recipeRatingsTableName, recipeRatingsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, createdByUserColumn, createdByUserColumn)),
						querygen.TotalCountSelect(recipeRatingsTableName, recipeRatingsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, createdByUserColumn, createdByUserColumn)),
						recipeRatingsTableName,
						querygen.FilterConditions(recipeRatingsTableName, recipeRatingsColumns,
							fmt.Sprintf("%s.%s IS NULL AND\n\t%s.%s = sqlc.arg(%s)", recipeRatingsTableName, archivedAtColumn, recipeRatingsTableName, createdByUserColumn, createdByUserColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, createdByUserColumn, createdByUserColumn),
						),
						recipeRatingsTableName,
						idColumn,
						querygen.CursorLimitClause(recipeRatingsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
