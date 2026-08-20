package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	validMeasurementUnitsTableName = "valid_measurement_units"

	validMeasurementUnitsUniversalColumn = "universal"
)

func init() {
	registerTableName(validMeasurementUnitsTableName)
}

var validMeasurementUnitsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	"volumetric",
	iconPathColumn,
	validMeasurementUnitsUniversalColumn,
	"metric",
	"imperial",
	slugColumn,
	pluralNameColumn,
	lastIndexedAtColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidMeasurementUnitsQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			pgGen.StandardCRUD(validMeasurementUnitsTableName, validMeasurementUnitsColumns,
				querygen.WithEntity("ValidMeasurementUnit", "ValidMeasurementUnits"),
				querygen.WithOmitted(querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetValidMeasurementUnits",
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
						strings.Join(applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validMeasurementUnitsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(validMeasurementUnitsTableName, validMeasurementUnitsColumns, []string{}),
						pgGen.TotalCountSelect(validMeasurementUnitsTableName, validMeasurementUnitsColumns, []string{}),
						validMeasurementUnitsTableName,
						pgGen.FilterConditions(validMeasurementUnitsTableName, validMeasurementUnitsColumns),
						validMeasurementUnitsTableName,
						idColumn,
						pgGen.CursorLimitClause(validMeasurementUnitsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUniversalValidMeasurementUnits",
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
						strings.Join(applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validMeasurementUnitsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(validMeasurementUnitsTableName, validMeasurementUnitsColumns, []string{}),
						pgGen.TotalCountSelect(validMeasurementUnitsTableName, validMeasurementUnitsColumns, []string{}),
						validMeasurementUnitsTableName,
						pgGen.FilterConditions(validMeasurementUnitsTableName, validMeasurementUnitsColumns,
							fmt.Sprintf("%s.%s = TRUE", validMeasurementUnitsTableName, validMeasurementUnitsUniversalColumn),
						),
						validMeasurementUnitsTableName,
						idColumn,
						pgGen.CursorLimitClause(validMeasurementUnitsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidMeasurementUnitsNeedingIndexing",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
WHERE %s.%s IS NULL
	AND (
	%s.%s IS NULL
	OR %s.%s < %s - '24 hours'::INTERVAL
);`,
						validMeasurementUnitsTableName,
						idColumn,
						validMeasurementUnitsTableName,
						validMeasurementUnitsTableName,
						archivedAtColumn,
						validMeasurementUnitsTableName,
						lastIndexedAtColumn,
						validMeasurementUnitsTableName,
						lastIndexedAtColumn,
						querygen.NowExpression,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRandomValidMeasurementUnit",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
ORDER BY RANDOM() LIMIT 1;`,
						strings.Join(applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validMeasurementUnitsTableName, s)
						}), ",\n\t"),
						validMeasurementUnitsTableName,
						validMeasurementUnitsTableName,
						archivedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidMeasurementUnitsWithIDs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[]);`,
						strings.Join(applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validMeasurementUnitsTableName, s)
						}), ",\n\t"),
						validMeasurementUnitsTableName,
						validMeasurementUnitsTableName,
						archivedAtColumn,
						validMeasurementUnitsTableName,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchForValidMeasurementUnits",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validMeasurementUnitsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(validMeasurementUnitsTableName, validMeasurementUnitsColumns, []string{}),
						pgGen.TotalCountSelect(validMeasurementUnitsTableName, validMeasurementUnitsColumns, []string{}),
						validMeasurementUnitsTableName,
						pgGen.FilterConditions(validMeasurementUnitsTableName, validMeasurementUnitsColumns,
							fmt.Sprintf("%s.%s %s", validMeasurementUnitsTableName, nameColumn, buildILIKEForArgument("name_query")),
						),
						pgGen.CursorLimitClause(validMeasurementUnitsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchValidMeasurementUnitsByIngredientID",
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
						strings.Join(applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
							if i == 0 {
								return fmt.Sprintf("DISTINCT(%s.%s)", validMeasurementUnitsTableName, s)
							}
							return fmt.Sprintf("%s.%s", validMeasurementUnitsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(validMeasurementUnitsTableName, validMeasurementUnitsColumns, []string{}, ` (
				valid_ingredient_measurement_units.valid_ingredient_id = sqlc.arg(valid_ingredient_id)
				OR valid_measurement_units.universal = true
			)`),
						pgGen.TotalCountSelect(validMeasurementUnitsTableName, validMeasurementUnitsColumns, []string{}),
						validMeasurementUnitsTableName,
						validIngredientMeasurementUnitsTableName,
						validIngredientMeasurementUnitsTableName,
						validMeasurementUnitIDColumn,
						validMeasurementUnitsTableName,
						idColumn,
						validIngredientsTableName,
						validIngredientMeasurementUnitsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						pgGen.FilterConditions(validMeasurementUnitsTableName, validMeasurementUnitsColumns,
							fmt.Sprintf("(\n\t\t%s.%s = sqlc.arg(%s)\n\t\tOR %s.%s = TRUE\n\t)", validIngredientMeasurementUnitsTableName, validIngredientIDColumn, validIngredientIDColumn, validMeasurementUnitsTableName, validMeasurementUnitsUniversalColumn),
							"valid_ingredients.archived_at IS NULL",
							"valid_ingredient_measurement_units.archived_at IS NULL",
						),
						pgGen.CursorLimitClause(validMeasurementUnitsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
