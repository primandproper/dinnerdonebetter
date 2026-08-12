package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	validVesselsTableName = "valid_vessels"
	validVesselIDColumn   = "valid_vessel_id"
	capacityUnitColumn    = "capacity_unit"
)

func init() {
	registerTableName(validVesselsTableName)
}

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

		insertColumns := filterForInsert(validVesselsColumns)

		fullSelectColumns := mergeColumns(
			applyToEach(filterFromSlice(validVesselsColumns, "capacity_unit"), func(i int, s string) string {
				return fmt.Sprintf("%s.%s", validVesselsTableName, s)
			}),
			applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_measurement_unit_%s", validMeasurementUnitsTableName, s, s)
			}),
			10,
		)

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveValidVessel",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s IS NULL AND %s = sqlc.arg(%s);`,
					validVesselsTableName,
					archivedAtColumn,
					currentTimeExpression,
					archivedAtColumn,
					idColumn,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CreateValidVessel",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					validVesselsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CheckValidVesselExistence",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.id
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
					validVesselsTableName,
					validVesselsTableName,
					validVesselsTableName,
					archivedAtColumn,
					validVesselsTableName,
					idColumn,
					idColumn,
				)),
			},
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
WHERE
	%s.%s IS NULL
	%s
GROUP BY %s.%s
%s;`,
					strings.Join(applyToEach(validVesselsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validVesselsTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(validVesselsTableName, true, true, []string{}),
					buildTotalCountSelect(validVesselsTableName, true, []string{}),
					validVesselsTableName,
					validVesselsTableName,
					archivedAtColumn,
					buildFilterConditions(
						validVesselsTableName,
						true,
						true,
					),
					validVesselsTableName, idColumn,
					buildCursorLimitClause(validVesselsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ScanValidVesselIDsForReindex",
					Type: ManyType,
				},
				Content: buildReindexScanQuery(validVesselsTableName),
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
					currentTimeExpression,
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
WHERE %s.%s IS NULL
	AND %s.%s %s
	%s
%s;`,
					strings.Join(applyToEach(validVesselsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validVesselsTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(validVesselsTableName, true, true, []string{}),
					buildTotalCountSelect(validVesselsTableName, true, []string{}),
					validVesselsTableName,
					validVesselsTableName,
					archivedAtColumn,
					validVesselsTableName,
					nameColumn,
					buildILIKEForArgument("name_query"),
					buildFilterConditions(
						validVesselsTableName,
						true,
						true,
					),
					buildCursorLimitClause(validVesselsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateValidVessel",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
					validVesselsTableName,
					strings.Join(applyToEach(filterForUpdate(validVesselsColumns), func(i int, s string) string {
						return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
					}), ",\n\t"),
					lastUpdatedAtColumn,
					currentTimeExpression,
					archivedAtColumn,
					idColumn,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateValidVesselLastIndexedAt",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s = sqlc.arg(%s) AND %s IS NULL;`,
					validVesselsTableName,
					lastIndexedAtColumn,
					currentTimeExpression,
					idColumn,
					idColumn,
					archivedAtColumn,
				)),
			},
		}
	default:
		return nil
	}
}
