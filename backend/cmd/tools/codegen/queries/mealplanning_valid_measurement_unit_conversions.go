package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	validMeasurementUnitConversionsTableName = "valid_measurement_unit_conversions"

	validMeasurementUnitConversionsFromUnitColumn          = "from_unit"
	validMeasurementUnitConversionsToUnitColumn            = "to_unit"
	validMeasurementUnitConversionsOnlyForIngredientColumn = "only_for_ingredient"
)

func init() {
	registerTableName(validMeasurementUnitConversionsTableName)
}

var validMeasurementUnitConversionsColumns = []string{
	idColumn,
	"from_unit",
	"to_unit",
	validMeasurementUnitConversionsOnlyForIngredientColumn,
	"modifier",
	notesColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidMeasurementUnitConversionsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			applyToEach(validMeasurementUnitConversionsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_measurement_unit_conversion_%s", validMeasurementUnitConversionsTableName, s, s)
			}),
			append(
				append(
					applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s_from.%s as from_unit_%s", validMeasurementUnitsTableName, s, s)
					}),
					applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s_to.%s as to_unit_%s", validMeasurementUnitsTableName, s, s)
					})...,
				),
				applyToEach(validIngredientsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_ingredient_%s", validIngredientsTableName, s, s)
				})...,
			),
			1,
		)

		return slices.Concat(
			querygen.StandardCRUD(validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsColumns,
				querygen.WithEntity("ValidMeasurementUnitConversion", "ValidMeasurementUnitConversions"),
				querygen.WithOmitted(querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetValidMeasurementUnitConversionsForMeasurementUnit",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s AS %s_from ON %s.%s = %s_from.%s
	JOIN %s AS %s_to ON %s.%s = %s_to.%s
	LEFT JOIN %s ON %s.%s = %s.%s
WHERE
	(%s_from.%s = sqlc.arg(%s) OR %s_to.%s = sqlc.arg(%s))
	AND %s.%s IS NULL
	AND %s_from.%s IS NULL
	AND %s_to.%s IS NULL
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(validMeasurementUnitConversionsTableName, true, true, []string{
							fmt.Sprintf("%s AS %s_from ON %s.%s = %s_from.%s", validMeasurementUnitsTableName, validMeasurementUnitsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsFromUnitColumn, validMeasurementUnitsTableName, idColumn),
							fmt.Sprintf("%s AS %s_to ON %s.%s = %s_to.%s", validMeasurementUnitsTableName, validMeasurementUnitsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsToUnitColumn, validMeasurementUnitsTableName, idColumn),
							fmt.Sprintf("%s ON %s.%s = %s.%s", validIngredientsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsOnlyForIngredientColumn, validIngredientsTableName, idColumn),
						}, fmt.Sprintf("(%s_from.%s = sqlc.arg(%s) OR %s_to.%s = sqlc.arg(%s))", validMeasurementUnitsTableName, idColumn, idColumn, validMeasurementUnitsTableName, idColumn, idColumn), fmt.Sprintf("%s_from.%s IS NULL", validMeasurementUnitsTableName, archivedAtColumn), fmt.Sprintf("%s_to.%s IS NULL", validMeasurementUnitsTableName, archivedAtColumn)),
						buildTotalCountSelect(validMeasurementUnitConversionsTableName, true, []string{
							fmt.Sprintf("%s AS %s_from ON %s.%s = %s_from.%s", validMeasurementUnitsTableName, validMeasurementUnitsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsFromUnitColumn, validMeasurementUnitsTableName, idColumn),
							fmt.Sprintf("%s AS %s_to ON %s.%s = %s_to.%s", validMeasurementUnitsTableName, validMeasurementUnitsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsToUnitColumn, validMeasurementUnitsTableName, idColumn),
							fmt.Sprintf("%s ON %s.%s = %s.%s", validIngredientsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsOnlyForIngredientColumn, validIngredientsTableName, idColumn),
						}, fmt.Sprintf("(%s_from.%s = sqlc.arg(%s) OR %s_to.%s = sqlc.arg(%s))", validMeasurementUnitsTableName, idColumn, idColumn, validMeasurementUnitsTableName, idColumn, idColumn), fmt.Sprintf("%s_from.%s IS NULL", validMeasurementUnitsTableName, archivedAtColumn), fmt.Sprintf("%s_to.%s IS NULL", validMeasurementUnitsTableName, archivedAtColumn)),
						validMeasurementUnitConversionsTableName,
						validMeasurementUnitsTableName, validMeasurementUnitsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsFromUnitColumn, validMeasurementUnitsTableName, idColumn,
						validMeasurementUnitsTableName, validMeasurementUnitsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsToUnitColumn, validMeasurementUnitsTableName, idColumn,
						validIngredientsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsOnlyForIngredientColumn, validIngredientsTableName, idColumn,
						validMeasurementUnitsTableName, idColumn, idColumn, validMeasurementUnitsTableName, idColumn, idColumn,
						validMeasurementUnitConversionsTableName, archivedAtColumn,
						validMeasurementUnitsTableName, archivedAtColumn,
						validMeasurementUnitsTableName, archivedAtColumn,
						buildFilterConditions(validMeasurementUnitConversionsTableName, true, true),
						buildCursorLimitClause(validMeasurementUnitConversionsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidMeasurementUnitConversion",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s AS %s_from ON %s.%s = %s_from.%s
	JOIN %s AS %s_to ON %s.%s = %s_to.%s
	LEFT JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL
	AND %s_from.%s IS NULL
	AND %s_to.%s IS NULL;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validMeasurementUnitConversionsTableName,
						validMeasurementUnitsTableName, validMeasurementUnitsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsFromUnitColumn, validMeasurementUnitsTableName, idColumn,
						validMeasurementUnitsTableName, validMeasurementUnitsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsToUnitColumn, validMeasurementUnitsTableName, idColumn,
						validIngredientsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsOnlyForIngredientColumn, validIngredientsTableName, idColumn,
						validMeasurementUnitConversionsTableName, idColumn, idColumn,
						validMeasurementUnitConversionsTableName, archivedAtColumn,
						validMeasurementUnitsTableName, archivedAtColumn,
						validMeasurementUnitsTableName, archivedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidMeasurementUnitConversionsForIngredients",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s AS %s_from ON %s.%s = %s_from.%s
	JOIN %s AS %s_to ON %s.%s = %s_to.%s
	LEFT JOIN %s ON %s.%s = %s.%s
WHERE
	(%s.%s IS NULL OR %s.%s = ANY(sqlc.arg(valid_ingredient_ids)::text[]))
	AND %s.%s IS NULL
	AND %s_from.%s IS NULL
	AND %s_to.%s IS NULL;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validMeasurementUnitConversionsTableName,
						validMeasurementUnitsTableName, validMeasurementUnitsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsFromUnitColumn, validMeasurementUnitsTableName, idColumn,
						validMeasurementUnitsTableName, validMeasurementUnitsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsToUnitColumn, validMeasurementUnitsTableName, idColumn,
						validIngredientsTableName, validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsOnlyForIngredientColumn, validIngredientsTableName, idColumn,
						validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsOnlyForIngredientColumn,
						validMeasurementUnitConversionsTableName, validMeasurementUnitConversionsOnlyForIngredientColumn,
						validMeasurementUnitConversionsTableName, archivedAtColumn,
						validMeasurementUnitsTableName, archivedAtColumn,
						validMeasurementUnitsTableName, archivedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetMeasurementUnitConversionMismatches",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`WITH ingredient_units AS (
	SELECT DISTINCT
		vimu.valid_ingredient_id,
		vimu.valid_measurement_unit_id
	FROM valid_ingredient_measurement_units vimu
	JOIN valid_ingredients vi ON vi.id = vimu.valid_ingredient_id
	JOIN valid_measurement_units vmu ON vmu.id = vimu.valid_measurement_unit_id
	WHERE vimu.archived_at IS NULL
		AND vi.archived_at IS NULL
		AND vmu.archived_at IS NULL
),
unit_pairs AS (
	SELECT
		a.valid_ingredient_id,
		a.valid_measurement_unit_id AS from_unit_id,
		b.valid_measurement_unit_id AS to_unit_id
	FROM ingredient_units a
	JOIN ingredient_units b ON a.valid_ingredient_id = b.valid_ingredient_id
		AND a.valid_measurement_unit_id < b.valid_measurement_unit_id
)
SELECT
	unit_pairs.valid_ingredient_id,
	unit_pairs.from_unit_id,
	unit_pairs.to_unit_id
FROM unit_pairs
LEFT JOIN valid_measurement_unit_conversions c ON c.from_unit = unit_pairs.from_unit_id
	AND c.to_unit = unit_pairs.to_unit_id
	AND (c.only_for_ingredient IS NULL OR c.only_for_ingredient = unit_pairs.valid_ingredient_id)
	AND c.archived_at IS NULL
WHERE c.id IS NULL;`)),
				},
			},
		)
	default:
		return nil
	}
}
