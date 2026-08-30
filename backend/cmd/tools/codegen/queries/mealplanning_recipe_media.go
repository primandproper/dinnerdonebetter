package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	recipeMediaTableName = "recipe_media"
)

var recipeMediaColumns = []string{
	idColumn,
	belongsToRecipeColumn,
	belongsToRecipeStepColumn,
	"mime_type",
	"internal_path",
	"external_path",
	indexColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildRecipeMediaQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			pgGen.StandardCRUD(recipeMediaTableName, recipeMediaColumns,
				querygen.WithEntity("RecipeMedia", "RecipeMedias"),
				querygen.WithOmitted(querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeMediaForRecipe",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s IS NULL
GROUP BY %s.%s
ORDER BY %s.%s;`,
						strings.Join(applyToEach(recipeMediaColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", recipeMediaTableName, s)
						}), ",\n\t"),
						recipeMediaTableName,
						recipeMediaTableName, belongsToRecipeColumn, recipeIDColumn,
						recipeMediaTableName, belongsToRecipeStepColumn,
						recipeMediaTableName, archivedAtColumn,
						recipeMediaTableName, idColumn,
						recipeMediaTableName, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeMediaForRecipeStep",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
GROUP BY %s.%s
ORDER BY %s.%s;`,
						strings.Join(applyToEach(recipeMediaColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", recipeMediaTableName, s)
						}), ",\n\t"),
						recipeMediaTableName,
						recipeMediaTableName, belongsToRecipeColumn, recipeIDColumn,
						recipeMediaTableName, belongsToRecipeStepColumn, recipeStepIDColumn,
						recipeMediaTableName, archivedAtColumn,
						recipeMediaTableName, idColumn,
						recipeMediaTableName, idColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
