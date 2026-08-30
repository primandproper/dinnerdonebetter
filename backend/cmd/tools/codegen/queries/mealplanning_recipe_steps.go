package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	recipeStepsTableName = "recipe_steps"

	indexColumn               = "index"
	recipeStepIDColumn        = "recipe_step_id"
	belongsToRecipeStepColumn = "belongs_to_recipe_step"
)

var recipeStepsColumns = []string{
	idColumn,
	indexColumn,
	preparationIDColumn,
	"minimum_estimated_time_in_seconds",
	"maximum_estimated_time_in_seconds",
	"minimum_temperature_in_celsius",
	"maximum_temperature_in_celsius",
	notesColumn,
	"explicit_instructions",
	"condition_expression",
	optionalColumn,
	"start_timer_automatically",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	belongsToRecipeColumn,
}

func buildRecipeStepsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			applyToEach(filterFromSlice(recipeStepsColumns, preparationIDColumn), func(i int, s string) string {
				return fmt.Sprintf("%s.%s", recipeStepsTableName, s)
			}),
			applyToEach(validPreparationsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_preparation_%s", validPreparationsTableName, s, s)
			}),
			2,
		)

		return slices.Concat(
			pgGen.StandardCRUD(recipeStepsTableName, recipeStepsColumns,
				querygen.WithEntity("RecipeStep", "RecipeSteps"),
				querygen.WithOwnership(belongsToRecipeColumn),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "CheckRecipeStepExistence",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
		JOIN %s ON %s.%s=%s.%s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
						recipeStepsTableName, idColumn,
						recipeStepsTableName,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipeStepsTableName, idColumn, recipeStepIDColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeStep",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						recipeStepsTableName,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						validPreparationsTableName, recipeStepsTableName, preparationIDColumn, validPreparationsTableName, idColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipeStepsTableName, idColumn, recipeStepIDColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeSteps",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(recipeStepsTableName, recipeStepsColumns, []string{}),
						pgGen.TotalCountSelect(recipeStepsTableName, recipeStepsColumns, []string{}),
						recipeStepsTableName,
						recipesTableName,
						recipeStepsTableName,
						belongsToRecipeColumn,
						recipesTableName,
						idColumn,
						validPreparationsTableName,
						recipeStepsTableName,
						preparationIDColumn,
						validPreparationsTableName,
						idColumn,
						pgGen.FilterConditions(recipeStepsTableName, recipeStepsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn),
							"recipes.archived_at IS NULL",
						),
						pgGen.CursorLimitClause(recipeStepsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeStepByRecipeID",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						recipeStepsTableName,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						validPreparationsTableName, recipeStepsTableName, preparationIDColumn, validPreparationsTableName, idColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, idColumn, idColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
