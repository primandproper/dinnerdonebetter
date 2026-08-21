package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	validVesselsTableName = "valid_vessels"
	validVesselIDColumn   = "valid_vessel_id"
	capacityUnitColumn    = "capacity_unit"
)

var validVesselsColumns = []string{
	idColumn,
	nameColumn,
	pluralNameColumn,
	descriptionColumn,
	iconPathColumn,
	"usable_for_storage",
	slugColumn,
	"display_in_summary_lists",
	"include_in_generated_instructions",
	"capacity",
	capacityUnitColumn,
	"width_in_millimeters",
	"length_in_millimeters",
	"height_in_millimeters",
	"shape",
	lastIndexedAtColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidVesselsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			applyToEach(filterFromSlice(validVesselsColumns, "capacity_unit"), func(i int, s string) string {
				return fmt.Sprintf("%s.%s", validVesselsTableName, s)
			}),
			applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_measurement_unit_%s", validMeasurementUnitsTableName, s, s)
			}),
			10,
		)

		return slices.Concat(
			// GetValidVessel joins valid_measurement_units for the capacity unit, so it
			// is not the standard single-row read and stays below.
			pgGen.StandardCRUD(validVesselsTableName, validVesselsColumns,
				querygen.WithEntity("ValidVessel", "ValidVessels"),
				querygen.WithOmitted(querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetValidVessels",
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
						strings.Join(applyToEach(validVesselsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validVesselsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(validVesselsTableName, validVesselsColumns, []string{}),
						pgGen.TotalCountSelect(validVesselsTableName, validVesselsColumns, []string{}),
						validVesselsTableName,
						pgGen.FilterConditions(validVesselsTableName, validVesselsColumns),
						validVesselsTableName,
						idColumn,
						pgGen.CursorLimitClause(validVesselsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidVesselIDsNeedingIndexing",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
WHERE %s.%s IS NULL
	AND (
	%s.%s IS NULL
	OR %s.%s < %s - '24 hours'::INTERVAL
);`,
						validVesselsTableName,
						idColumn,
						validVesselsTableName,
						validVesselsTableName,
						archivedAtColumn,
						validVesselsTableName,
						lastIndexedAtColumn,
						validVesselsTableName,
						lastIndexedAtColumn,
						querygen.NowExpression,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidVessel",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	LEFT JOIN %s ON %s.%s=%s.id
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validVesselsTableName,
						validMeasurementUnitsTableName,
						validVesselsTableName,
						capacityUnitColumn,
						validMeasurementUnitsTableName,
						validVesselsTableName,
						archivedAtColumn,
						validMeasurementUnitsTableName,
						archivedAtColumn,
						validVesselsTableName,
						idColumn,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRandomValidVessel",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
ORDER BY RANDOM() LIMIT 1;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validVesselsTableName,
						validMeasurementUnitsTableName,
						validVesselsTableName,
						capacityUnitColumn,
						validMeasurementUnitsTableName,
						idColumn,
						validVesselsTableName,
						archivedAtColumn,
						validMeasurementUnitsTableName,
						archivedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidVesselsWithIDs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s=%s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[]);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validVesselsTableName,
						validMeasurementUnitsTableName,
						validVesselsTableName,
						capacityUnitColumn,
						validMeasurementUnitsTableName,
						idColumn,
						validVesselsTableName,
						archivedAtColumn,
						validMeasurementUnitsTableName,
						archivedAtColumn,
						validVesselsTableName,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchForValidVessels",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(applyToEach(validVesselsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validVesselsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(validVesselsTableName, validVesselsColumns, []string{}),
						pgGen.TotalCountSelect(validVesselsTableName, validVesselsColumns, []string{}),
						validVesselsTableName,
						pgGen.FilterConditions(validVesselsTableName, validVesselsColumns,
							fmt.Sprintf("%s.%s %s", validVesselsTableName, nameColumn, buildILIKEForArgument("name_query")),
						),
						pgGen.CursorLimitClause(validVesselsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
