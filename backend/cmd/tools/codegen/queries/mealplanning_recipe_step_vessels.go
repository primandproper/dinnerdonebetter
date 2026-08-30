package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	recipeStepVesselsTableName = "recipe_step_vessels"

	recipeStepVesselIDColumn = "recipe_step_vessel_id"
)

var recipeStepVesselsColumns = []string{
	idColumn,
	nameColumn,
	notesColumn,
	belongsToRecipeStepColumn,
	recipeStepProductIDColumn,
	"valid_vessel_id",
	"vessel_predicate",
	"minimum_quantity",
	"maximum_quantity",
	"unavailable_after_step",
	indexColumn,
	optionIndexColumn,
	scaleFactorColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildRecipeStepVesselsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			applyToEach(filterFromSlice(recipeStepVesselsColumns, validVesselIDColumn), func(i int, s string) string {
				return fmt.Sprintf("%s.%s", recipeStepVesselsTableName, s)
			}),
			mergeColumns(
				applyToEach(filterFromSlice(validVesselsColumns, capacityUnitColumn), func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_vessel_%s", validVesselsTableName, s, s)
				}),
				applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_measurement_unit_%s", validMeasurementUnitsTableName, s, s)
				}),
				10,
			),
			1,
		)

		return slices.Concat(
			pgGen.StandardCRUD(recipeStepVesselsTableName, recipeStepVesselsColumns,
				querygen.WithEntity("RecipeStepVessel", "RecipeStepVessels"),
				querygen.WithOwnership(belongsToRecipeStepColumn),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "CheckRecipeStepVesselExistence",
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
						recipeStepVesselsTableName, idColumn,
						recipeStepVesselsTableName,
						recipeStepsTableName, recipeStepVesselsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						recipeStepVesselsTableName, archivedAtColumn,
						recipeStepVesselsTableName, belongsToRecipeStepColumn, recipeStepIDColumn,
						recipeStepVesselsTableName, idColumn, recipeStepVesselIDColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipeStepsTableName, idColumn, recipeStepIDColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeStepVesselsForRecipe",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	LEFT JOIN %s ON %s.%s=%s.%s
	LEFT JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						recipeStepVesselsTableName,
						validVesselsTableName, recipeStepVesselsTableName, validVesselIDColumn, validVesselsTableName, idColumn,
						validMeasurementUnitsTableName, validVesselsTableName, capacityUnitColumn, validMeasurementUnitsTableName, idColumn,
						recipeStepsTableName, recipeStepVesselsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						recipeStepVesselsTableName, archivedAtColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeStepVessel",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	LEFT JOIN %s ON %s.%s=%s.%s
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
						strings.Join(fullSelectColumns, ",\n\t"),
						recipeStepVesselsTableName,
						validVesselsTableName, recipeStepVesselsTableName, validVesselIDColumn, validVesselsTableName, idColumn,
						validMeasurementUnitsTableName, validVesselsTableName, capacityUnitColumn, validMeasurementUnitsTableName, idColumn,
						recipeStepsTableName, recipeStepVesselsTableName, belongsToRecipeStepColumn, recipeStepsTableName, idColumn,
						recipesTableName, recipeStepsTableName, belongsToRecipeColumn, recipesTableName, idColumn,
						recipeStepVesselsTableName, archivedAtColumn,
						recipeStepVesselsTableName, belongsToRecipeStepColumn, recipeStepIDColumn,
						recipeStepVesselsTableName, idColumn, recipeStepVesselIDColumn,
						recipeStepsTableName, archivedAtColumn,
						recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn,
						recipeStepsTableName, idColumn, recipeStepIDColumn,
						recipesTableName, archivedAtColumn,
						recipesTableName, idColumn, recipeIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRecipeStepVessels",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	 LEFT JOIN %s ON %s.%s=%s.%s
	 LEFT JOIN %s ON %s.%s=%s.%s
	 JOIN %s ON %s.%s=%s.%s
	 JOIN %s ON %s.%s=%s.%s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(recipeStepVesselsTableName, recipeStepVesselsColumns, []string{}),
						pgGen.TotalCountSelect(recipeStepVesselsTableName, recipeStepVesselsColumns, []string{}),
						recipeStepVesselsTableName,
						validVesselsTableName,
						recipeStepVesselsTableName,
						validVesselIDColumn,
						validVesselsTableName,
						idColumn,
						validMeasurementUnitsTableName,
						validVesselsTableName,
						capacityUnitColumn,
						validMeasurementUnitsTableName,
						idColumn,
						recipeStepsTableName,
						recipeStepVesselsTableName,
						belongsToRecipeStepColumn,
						recipeStepsTableName,
						idColumn,
						recipesTableName,
						recipeStepsTableName,
						belongsToRecipeColumn,
						recipesTableName,
						idColumn,
						pgGen.FilterConditions(recipeStepVesselsTableName, recipeStepVesselsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeStepVesselsTableName, belongsToRecipeStepColumn, recipeStepIDColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeStepsTableName, belongsToRecipeColumn, recipeIDColumn),
							"recipe_steps.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipeStepsTableName, idColumn, recipeStepIDColumn),
							"recipes.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", recipesTableName, idColumn, recipeIDColumn),
						),
						pgGen.CursorLimitClause(recipeStepVesselsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateRecipeStepVessel",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						recipeStepVesselsTableName,
						strings.Join(applyToEach(querygen.ForUpdate(recipeStepVesselsColumns), func(i int, s string) string {
							return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
						}), ",\n\t"),
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						belongsToRecipeStepColumn, belongsToRecipeStepColumn,
						idColumn, idColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
