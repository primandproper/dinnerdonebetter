package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

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
			pgGen.StandardCRUD(recipeRatingsTableName, recipeRatingsColumns,
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
						pgGen.FilterCountSelect(recipeRatingsTableName, recipeRatingsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, belongsToRecipeColumn, belongsToRecipeColumn)),
						pgGen.TotalCountSelect(recipeRatingsTableName, recipeRatingsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, belongsToRecipeColumn, belongsToRecipeColumn)),
						recipeRatingsTableName,
						pgGen.FilterConditions(recipeRatingsTableName, recipeRatingsColumns,
							fmt.Sprintf("%s.%s IS NULL AND\n\t%s.%s = sqlc.arg(%s)", recipeRatingsTableName, archivedAtColumn, recipeRatingsTableName, belongsToRecipeColumn, belongsToRecipeColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, belongsToRecipeColumn, belongsToRecipeColumn),
						),
						recipeRatingsTableName,
						idColumn,
						pgGen.CursorLimitClause(recipeRatingsTableName),
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
						pgGen.FilterCountSelect(recipeRatingsTableName, recipeRatingsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, createdByUserColumn, createdByUserColumn)),
						pgGen.TotalCountSelect(recipeRatingsTableName, recipeRatingsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, createdByUserColumn, createdByUserColumn)),
						recipeRatingsTableName,
						pgGen.FilterConditions(recipeRatingsTableName, recipeRatingsColumns,
							fmt.Sprintf("%s.%s IS NULL AND\n\t%s.%s = sqlc.arg(%s)", recipeRatingsTableName, archivedAtColumn, recipeRatingsTableName, createdByUserColumn, createdByUserColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, createdByUserColumn, createdByUserColumn),
						),
						recipeRatingsTableName,
						idColumn,
						pgGen.CursorLimitClause(recipeRatingsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
