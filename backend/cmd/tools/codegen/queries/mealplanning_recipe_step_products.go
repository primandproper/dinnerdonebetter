package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	recipeStepProductsTableName = "recipe_step_products"

	recipeStepProductIDColumn = "recipe_step_product_id"
)

var recipeStepProductsColumns = []string{
	idColumn,
	nameColumn,
	"type",
	measurementUnitColumn,
	"minimum_measurement_quantity_value",
	"maximum_measurement_quantity_value",
	"minimum_item_quantity_value",
	"maximum_item_quantity_value",
	"quantity_notes",
	"compostable",
	"maximum_storage_duration_in_seconds",
	minimumStorageTemperatureInCelsiusColumn,
	maximumStorageTemperatureInCelsiusColumn,
	storageInstructionsColumn,
	"is_liquid",
	"is_waste",
	indexColumn,
	"contained_in_vessel_index",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	belongsToRecipeStepColumn,
}

func buildRecipeStepProductsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			applyToEach(filterFromSlice(recipeStepProductsColumns, measurementUnitColumn), func(i int, s string) string {
				return fmt.Sprintf("%s.%s", recipeStepProductsTableName, s)
			}),
			applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_measurement_unit_%s", validMeasurementUnitsTableName, s, s)
			}),
			3,
		)

		return slices.Concat(
			pgGen.StandardCRUD(recipeStepProductsTableName, recipeStepProductsColumns,
				querygen.WithEntity("RecipeStepProduct", "RecipeStepProducts"),
				querygen.WithOwnership(belongsToRecipeStepColumn),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "CheckRecipeStepProductExistence",
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
						recipeStepProductsTableName, idColumn,
						recipeStepProductsTableName,
						recipeStepsTableName, recipeStepProductsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						recipeStepProductsTableName, archivedAtColumn,
						recipeStepProductsTableName, belongsToRecipeStepColumn, recipeStepIDColumn,
						recipeStepProductsTableName, idColumn, recipeStepProductIDColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipeStepsTableName, idColumn, recipeStepIDColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeStepProductsForRecipe",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	LEFT JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						recipeStepProductsTableName,
						recipeStepsTableName, recipeStepProductsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						validMeasurementUnitsTableName, recipeStepProductsTableName, measurementUnitColumn, validMeasurementUnitsTableName, idColumn,
						recipeStepProductsTableName, archivedAtColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeStepProducts",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	(
		SELECT COUNT(%s.%s)
		FROM %s
		WHERE %s.%s IS NULL
				AND %s.%s = sqlc.arg(%s)
	) AS total_count
FROM %s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	LEFT JOIN %s ON %s.%s=%s.%s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(recipeStepProductsTableName, recipeStepProductsColumns, nil),
						recipeStepProductsTableName,
						idColumn,
						recipeStepProductsTableName,
						recipeStepProductsTableName,
						archivedAtColumn,
						recipeStepProductsTableName,
						belongsToRecipeStepColumn,
						recipeStepIDColumn,
						recipeStepProductsTableName,
						recipeStepsTableName,
						recipeStepProductsTableName,
						belongsToRecipeStepColumn,
						recipeStepsTableName,
						idColumn,
						recipesTableName,
						recipeStepsTableName,
						belongsToRecipeColumn,
						recipesTableName,
						idColumn,
						validMeasurementUnitsTableName,
						recipeStepProductsTableName,
						measurementUnitColumn,
						validMeasurementUnitsTableName,
						idColumn,
						pgGen.FilterConditions(recipeStepProductsTableName, recipeStepProductsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeStepProductsTableName, belongsToRecipeStepColumn, recipeStepIDColumn),
							"recipe_steps.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeStepsTableName, idColumn, recipeStepIDColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn),
							"recipes.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipesTableName, idColumn, recipeIDColumn),
						),
						pgGen.CursorLimitClause(recipeStepProductsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeStepProduct",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	LEFT JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						recipeStepProductsTableName,
						recipeStepsTableName, recipeStepProductsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						validMeasurementUnitsTableName, recipeStepProductsTableName, measurementUnitColumn, validMeasurementUnitsTableName, idColumn,
						recipeStepProductsTableName, archivedAtColumn,
						recipeStepProductsTableName, belongsToRecipeStepColumn, recipeStepIDColumn,
						recipeStepProductsTableName, idColumn, recipeStepProductIDColumn,
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
