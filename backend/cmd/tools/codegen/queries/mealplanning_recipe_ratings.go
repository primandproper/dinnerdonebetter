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
WHERE
	%s.%s IS NULL AND
	%s.%s = sqlc.arg(%s)
	%s
GROUP BY %s.%s
%s;`,
						strings.Join(applyToEach(recipeRatingsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", recipeRatingsTableName, s)
						}), ",\n\t"),
						buildFilterCountSelect(recipeRatingsTableName, true, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, belongsToRecipeColumn, belongsToRecipeColumn)),
						buildTotalCountSelect(recipeRatingsTableName, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, belongsToRecipeColumn, belongsToRecipeColumn)),
						recipeRatingsTableName,
						recipeRatingsTableName, archivedAtColumn,
						recipeRatingsTableName, belongsToRecipeColumn, belongsToRecipeColumn,
						buildFilterConditions(recipeRatingsTableName, true, true, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, belongsToRecipeColumn, belongsToRecipeColumn)),
						recipeRatingsTableName, idColumn,
						buildCursorLimitClause(recipeRatingsTableName),
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
WHERE
	%s.%s IS NULL AND
	%s.%s = sqlc.arg(%s)
	%s
GROUP BY %s.%s
%s;`,
						strings.Join(applyToEach(recipeRatingsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", recipeRatingsTableName, s)
						}), ",\n\t"),
						buildFilterCountSelect(recipeRatingsTableName, true, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, createdByUserColumn, createdByUserColumn)),
						buildTotalCountSelect(recipeRatingsTableName, true, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, createdByUserColumn, createdByUserColumn)),
						recipeRatingsTableName,
						recipeRatingsTableName, archivedAtColumn,
						recipeRatingsTableName, createdByUserColumn, createdByUserColumn,
						buildFilterConditions(recipeRatingsTableName, true, true, fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeRatingsTableName, createdByUserColumn, createdByUserColumn)),
						recipeRatingsTableName, idColumn,
						buildCursorLimitClause(recipeRatingsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
