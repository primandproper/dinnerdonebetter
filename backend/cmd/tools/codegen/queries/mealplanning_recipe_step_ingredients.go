package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	recipeStepIngredientsTableName = "recipe_step_ingredients"

	recipeStepIngredientIDColumn = "recipe_step_ingredient_id"
	ingredientIDColumn           = "ingredient_id"
	measurementUnitColumn        = "measurement_unit"
)

func init() {
	registerTableName(recipeStepIngredientsTableName)
}

var recipeStepIngredientsColumns = []string{
	idColumn,
	nameColumn,
	"optional",
	ingredientIDColumn,
	measurementUnitColumn,
	"minimum_quantity_value",
	"maximum_quantity_value",
	"quantity_notes",
	"recipe_step_product_id",
	"ingredient_notes",
	"index",
	"option_index",
	"to_taste",
	"product_percentage_to_use",
	"vessel_index",
	"scale_factor",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	"recipe_step_product_recipe_id",
	belongsToRecipeStepColumn,
}

func buildRecipeStepIngredientsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumn := mergeColumns(
			applyToEach(filterFromSlice(recipeStepIngredientsColumns, ingredientIDColumn, measurementUnitColumn), func(i int, s string) string {
				return fmt.Sprintf("%s.%s", recipeStepIngredientsTableName, s)
			}),
			append(
				applyToEach(validIngredientsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_ingredient_%s", validIngredientsTableName, s, s)
				}),
				applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_measurement_unit_%s", validMeasurementUnitsTableName, s, s)
				})...,
			),
			3,
		)

		return slices.Concat(
			querygen.StandardCRUD(recipeStepIngredientsTableName, recipeStepIngredientsColumns,
				querygen.WithEntity("RecipeStepIngredient", "RecipeStepIngredients"),
				querygen.WithOwnership(belongsToRecipeStepColumn),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "CheckRecipeStepIngredientExistence",
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
						recipeStepIngredientsTableName, idColumn,
						recipeStepIngredientsTableName,
						recipeStepsTableName, recipeStepIngredientsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						recipeStepIngredientsTableName, archivedAtColumn,
						recipeStepIngredientsTableName, belongsToRecipeStepColumn, recipeStepIDColumn,
						recipeStepIngredientsTableName, idColumn, recipeStepIngredientIDColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipeStepsTableName, idColumn, recipeStepIDColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetAllRecipeStepIngredientsForRecipe",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	LEFT JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumn, ",\n\t"),
						recipeStepIngredientsTableName,
						recipeStepsTableName, recipeStepIngredientsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						validIngredientsTableName, recipeStepIngredientsTableName, ingredientIDColumn, validIngredientsTableName, idColumn,
						validMeasurementUnitsTableName, recipeStepIngredientsTableName, measurementUnitColumn, validMeasurementUnitsTableName, idColumn,
						recipeStepIngredientsTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeStepIngredients",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	LEFT JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
						strings.Join(fullSelectColumn, ",\n\t"),

						//

						buildFilterCountSelect(
							recipeStepIngredientsTableName,
							true,
							true,
							[]string{
								fmt.Sprintf("%s ON %s.%s = %s.%s", recipeStepsTableName, recipeStepIngredientsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn),
								fmt.Sprintf("%s ON %s.%s = %s.%s", recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn),
							},
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipesTableName, idColumn, recipeIDColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeStepsTableName, idColumn, recipeStepIDColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeStepIngredientsTableName, belongsToRecipeStepColumn, recipeStepIDColumn),
						),

						//

						buildTotalCountSelect(

							recipeStepIngredientsTableName,
							true,
							[]string{
								fmt.Sprintf("%s ON %s.%s = %s.%s", recipeStepsTableName, recipeStepIngredientsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn),
								fmt.Sprintf("%s ON %s.%s = %s.%s", recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn),
							},
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipesTableName, idColumn, recipeIDColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeStepsTableName, idColumn, recipeStepIDColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeStepIngredientsTableName, belongsToRecipeStepColumn, recipeStepIDColumn),
						),

						//

						recipeStepIngredientsTableName,
						recipeStepsTableName, recipeStepIngredientsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						validIngredientsTableName, recipeStepIngredientsTableName, ingredientIDColumn, validIngredientsTableName, idColumn,
						validMeasurementUnitsTableName, recipeStepIngredientsTableName, measurementUnitColumn, validMeasurementUnitsTableName, idColumn,

						//

						recipeStepIngredientsTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
						recipeStepsTableName, idColumn, recipeStepIDColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipeStepIngredientsTableName, belongsToRecipeStepColumn, recipeStepIDColumn,
						buildFilterConditions(recipeStepIngredientsTableName, true, false),
						buildCursorLimitClause(recipeStepIngredientsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeStepIngredient",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	LEFT JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumn, ",\n\t"),
						recipeStepIngredientsTableName,
						recipeStepsTableName, recipeStepIngredientsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						validIngredientsTableName, recipeStepIngredientsTableName, ingredientIDColumn, validIngredientsTableName, idColumn,
						validMeasurementUnitsTableName, recipeStepIngredientsTableName, measurementUnitColumn, validMeasurementUnitsTableName, idColumn,
						recipeStepIngredientsTableName, archivedAtColumn,
						recipeStepIngredientsTableName, belongsToRecipeStepColumn, recipeStepIDColumn,
						recipeStepIngredientsTableName, idColumn, recipeStepIngredientIDColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipeStepsTableName, idColumn, recipeStepIDColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
