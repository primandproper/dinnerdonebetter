package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	validIngredientMeasurementUnitsTableName = "valid_ingredient_measurement_units"
	validMeasurementUnitColumn               = "valid_measurement_unit"
	validMeasurementUnitIDColumn             = "valid_measurement_unit_id"
)

func init() {
	registerTableName(validIngredientMeasurementUnitsTableName)
}

var validIngredientMeasurementUnitsColumns = []string{
	idColumn,
	notesColumn,
	validMeasurementUnitIDColumn,
	validIngredientIDColumn,
	"minimum_allowable_quantity",
	"maximum_allowable_quantity",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidIngredientMeasurementUnitsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			applyToEach(filterFromSlice(validIngredientMeasurementUnitsColumns, "valid_ingredient_id", "valid_measurement_unit_id"), func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_ingredient_measurement_unit_%s", validIngredientMeasurementUnitsTableName, s, s)
			}),
			append(
				applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_measurement_unit_%s", validMeasurementUnitsTableName, s, s)
				}),
				applyToEach(validIngredientsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_ingredient_%s", validIngredientsTableName, s, s)
				})...),
			2,
		)

		return slices.Concat(
			pgGen.StandardCRUD(validIngredientMeasurementUnitsTableName, validIngredientMeasurementUnitsColumns,
				querygen.WithEntity("ValidIngredientMeasurementUnit", "ValidIngredientMeasurementUnits"),
				querygen.WithOmitted(querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientMeasurementUnitsForIngredient",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(validIngredientMeasurementUnitsTableName, validIngredientMeasurementUnitsColumns, []string{}),
						pgGen.TotalCountSelect(validIngredientMeasurementUnitsTableName, validIngredientMeasurementUnitsColumns, []string{}),
						validIngredientMeasurementUnitsTableName,
						validMeasurementUnitsTableName,
						validIngredientMeasurementUnitsTableName,
						validMeasurementUnitIDColumn,
						validMeasurementUnitsTableName,
						idColumn,
						validIngredientsTableName,
						validIngredientMeasurementUnitsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						pgGen.FilterConditions(validIngredientMeasurementUnitsTableName, validIngredientMeasurementUnitsColumns,
							"valid_measurement_units.archived_at IS NULL",
							"valid_ingredients.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validIngredientMeasurementUnitsTableName, validIngredientIDColumn, validIngredientIDColumn),
						),
						pgGen.CursorLimitClause(validIngredientMeasurementUnitsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientMeasurementUnitsForMeasurementUnit",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(validIngredientMeasurementUnitsTableName, validIngredientMeasurementUnitsColumns, []string{}),
						pgGen.TotalCountSelect(validIngredientMeasurementUnitsTableName, validIngredientMeasurementUnitsColumns, []string{}),
						validIngredientMeasurementUnitsTableName,
						validMeasurementUnitsTableName,
						validIngredientMeasurementUnitsTableName,
						validMeasurementUnitIDColumn,
						validMeasurementUnitsTableName,
						idColumn,
						validIngredientsTableName,
						validIngredientMeasurementUnitsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						pgGen.FilterConditions(validIngredientMeasurementUnitsTableName, validIngredientMeasurementUnitsColumns,
							"valid_measurement_units.archived_at IS NULL",
							"valid_ingredients.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validIngredientMeasurementUnitsTableName, validMeasurementUnitIDColumn, validMeasurementUnitIDColumn),
						),
						pgGen.CursorLimitClause(validIngredientMeasurementUnitsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientMeasurementUnits",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(validIngredientMeasurementUnitsTableName, validIngredientMeasurementUnitsColumns, []string{}),
						pgGen.TotalCountSelect(validIngredientMeasurementUnitsTableName, validIngredientMeasurementUnitsColumns, []string{}),
						validIngredientMeasurementUnitsTableName,
						validMeasurementUnitsTableName,
						validIngredientMeasurementUnitsTableName,
						validMeasurementUnitIDColumn,
						validMeasurementUnitsTableName,
						idColumn,
						validIngredientsTableName,
						validIngredientMeasurementUnitsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						pgGen.FilterConditions(validIngredientMeasurementUnitsTableName, validIngredientMeasurementUnitsColumns,
							"valid_measurement_units.archived_at IS NULL",
							"valid_ingredients.archived_at IS NULL",
						),
						pgGen.CursorLimitClause(validIngredientMeasurementUnitsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientMeasurementUnit",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validIngredientMeasurementUnitsTableName,
						validMeasurementUnitsTableName, validIngredientMeasurementUnitsTableName, validMeasurementUnitIDColumn, validMeasurementUnitsTableName, idColumn,
						validIngredientsTableName, validIngredientMeasurementUnitsTableName, validIngredientIDColumn, validIngredientsTableName, idColumn,
						validIngredientMeasurementUnitsTableName, archivedAtColumn,
						validMeasurementUnitsTableName, archivedAtColumn,
						validIngredientsTableName, archivedAtColumn,
						validIngredientMeasurementUnitsTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientMeasurementUnitsByIDs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[]);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validIngredientMeasurementUnitsTableName,
						validMeasurementUnitsTableName, validIngredientMeasurementUnitsTableName, validMeasurementUnitIDColumn, validMeasurementUnitsTableName, idColumn,
						validIngredientsTableName, validIngredientMeasurementUnitsTableName, validIngredientIDColumn, validIngredientsTableName, idColumn,
						validIngredientMeasurementUnitsTableName, archivedAtColumn,
						validMeasurementUnitsTableName, archivedAtColumn,
						validIngredientsTableName, archivedAtColumn,
						validIngredientMeasurementUnitsTableName, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "ValidIngredientMeasurementUnitPairIsValid",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS(
	SELECT %s
	FROM %s
	WHERE %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s)
	AND %s IS NULL
);`,
						idColumn,
						validIngredientMeasurementUnitsTableName,
						validMeasurementUnitIDColumn, validMeasurementUnitIDColumn,
						validIngredientIDColumn, validIngredientIDColumn,
						archivedAtColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
