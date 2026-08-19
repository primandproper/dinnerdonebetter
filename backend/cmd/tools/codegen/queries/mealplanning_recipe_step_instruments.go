package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	recipeStepInstrumentsTableName = "recipe_step_instruments"

	recipeStepInstrumentIDColumn = "recipe_step_instrument_id"
	instrumentIDColumn           = "instrument_id"
)

func init() {
	registerTableName(recipeStepInstrumentsTableName)
}

var recipeStepInstrumentsColumns = []string{
	idColumn,
	instrumentIDColumn,
	"recipe_step_product_id",
	nameColumn,
	notesColumn,
	"preference_rank",
	"optional",
	"minimum_quantity",
	"maximum_quantity",
	"index",
	"option_index",
	"scale_factor",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	belongsToRecipeStepColumn,
}

func buildRecipeStepInstrumentsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumn := mergeColumns(
			applyToEach(filterFromSlice(recipeStepInstrumentsColumns, instrumentIDColumn, measurementUnitColumn), func(i int, s string) string {
				return fmt.Sprintf("%s.%s", recipeStepInstrumentsTableName, s)
			}),
			applyToEach(validInstrumentsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_instrument_%s", validInstrumentsTableName, s, s)
			}),
			1,
		)

		return slices.Concat(
			querygen.StandardCRUD(recipeStepInstrumentsTableName, recipeStepInstrumentsColumns,
				querygen.WithEntity("RecipeStepInstrument", "RecipeStepInstruments"),
				querygen.WithOwnership(belongsToRecipeStepColumn),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "CheckRecipeStepInstrumentExistence",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
		JOIN %s ON %s.%s=%s.%s
		JOIN %s ON %s.%s=%s.%s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
						recipeStepInstrumentsTableName, idColumn,
						recipeStepInstrumentsTableName,
						recipeStepsTableName, recipeStepInstrumentsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						recipeStepInstrumentsTableName, archivedAtColumn,
						recipeStepInstrumentsTableName, belongsToRecipeStepColumn, recipeStepIDColumn,
						recipeStepInstrumentsTableName, idColumn, recipeStepInstrumentIDColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipeStepsTableName, idColumn, recipeStepIDColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeStepInstrumentsForRecipe",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	LEFT JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumn, ",\n\t"),
						recipeStepInstrumentsTableName,
						validInstrumentsTableName, recipeStepInstrumentsTableName, instrumentIDColumn, validInstrumentsTableName, idColumn,
						recipeStepsTableName, recipeStepInstrumentsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						recipeStepInstrumentsTableName, archivedAtColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeStepInstrument",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	LEFT JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumn, ",\n\t"),
						recipeStepInstrumentsTableName,
						validInstrumentsTableName, recipeStepInstrumentsTableName, instrumentIDColumn, validInstrumentsTableName, idColumn,
						recipeStepsTableName, recipeStepInstrumentsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						recipeStepInstrumentsTableName, archivedAtColumn,
						recipeStepInstrumentsTableName, belongsToRecipeStepColumn, recipeStepIDColumn,
						recipeStepInstrumentsTableName, idColumn, recipeStepInstrumentIDColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipeStepsTableName, idColumn, recipeStepIDColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeStepInstruments",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	LEFT JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
						strings.Join(fullSelectColumn, ",\n\t"),
						buildFilterCountSelect(recipeStepInstrumentsTableName, true, true, nil),
						buildTotalCountSelect(recipeStepInstrumentsTableName, true, []string{}),
						recipeStepInstrumentsTableName,
						validInstrumentsTableName, recipeStepInstrumentsTableName, instrumentIDColumn, validInstrumentsTableName, idColumn,
						recipeStepsTableName, recipeStepInstrumentsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						recipeStepInstrumentsTableName, archivedAtColumn,
						recipeStepInstrumentsTableName, belongsToRecipeStepColumn, recipeStepIDColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipeStepsTableName, idColumn, recipeStepIDColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
						buildFilterConditions(recipeStepInstrumentsTableName, true, false),
						buildCursorLimitClause(recipeStepInstrumentsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
