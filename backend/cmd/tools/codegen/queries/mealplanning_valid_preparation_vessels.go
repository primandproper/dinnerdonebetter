package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	validPreparationVesselsTableName = "valid_preparation_vessels"
)

var validPreparationVesselsColumns = []string{
	idColumn,
	notesColumn,
	validPreparationIDColumn,
	validVesselIDColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidPreparationVesselsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := append(
			mergeColumns(
				applyToEach(filterFromSlice(validPreparationVesselsColumns, "valid_preparation_id", "valid_vessel_id"), func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_preparation_vessel_%s", validPreparationVesselsTableName, s, s)
				}),
				applyToEach(validPreparationsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_preparation_%s", validPreparationsTableName, s, s)
				}),
				2,
			),
			mergeColumns(
				applyToEach(filterFromSlice(validVesselsColumns, "capacity_unit"), func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_vessel_%s", validVesselsTableName, s, s)
				}),
				applyToEach(validMeasurementUnitsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_measurement_unit_%s", validMeasurementUnitsTableName, s, s)
				}),
				10,
			)...,
		)

		return slices.Concat(
			pgGen.StandardCRUD(validPreparationVesselsTableName, validPreparationVesselsColumns,
				querygen.WithEntity("ValidPreparationVessel", "ValidPreparationVessels"),
				querygen.WithOmitted(querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPreparationVesselsForPreparation",
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
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(validPreparationVesselsTableName, validPreparationVesselsColumns, []string{}),
						pgGen.TotalCountSelect(validPreparationVesselsTableName, validPreparationVesselsColumns, []string{}),
						validPreparationVesselsTableName,
						validVesselsTableName,
						validPreparationVesselsTableName,
						validVesselIDColumn,
						validVesselsTableName,
						idColumn,
						validPreparationsTableName,
						validPreparationVesselsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validMeasurementUnitsTableName,
						validVesselsTableName,
						capacityUnitColumn,
						validMeasurementUnitsTableName,
						idColumn,
						pgGen.FilterConditions(validPreparationVesselsTableName, validPreparationVesselsColumns,
							"valid_vessels.archived_at IS NULL",
							"valid_preparations.archived_at IS NULL",
							"valid_measurement_units.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationVesselsTableName, validPreparationIDColumn, idColumn),
						),
						pgGen.CursorLimitClause(validPreparationVesselsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPreparationVesselsForVessel",
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
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(validPreparationVesselsTableName, validPreparationVesselsColumns, []string{}),
						pgGen.TotalCountSelect(validPreparationVesselsTableName, validPreparationVesselsColumns, []string{}),
						validPreparationVesselsTableName,
						validVesselsTableName,
						validPreparationVesselsTableName,
						validVesselIDColumn,
						validVesselsTableName,
						idColumn,
						validPreparationsTableName,
						validPreparationVesselsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validMeasurementUnitsTableName,
						validVesselsTableName,
						capacityUnitColumn,
						validMeasurementUnitsTableName,
						idColumn,
						pgGen.FilterConditions(validPreparationVesselsTableName, validPreparationVesselsColumns,
							"valid_vessels.archived_at IS NULL",
							"valid_preparations.archived_at IS NULL",
							"valid_measurement_units.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationVesselsTableName, validVesselIDColumn, idColumn),
						),
						pgGen.CursorLimitClause(validPreparationVesselsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPreparationVessels",
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
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(validPreparationVesselsTableName, validPreparationVesselsColumns, []string{}),
						pgGen.TotalCountSelect(validPreparationVesselsTableName, validPreparationVesselsColumns, []string{}),
						validPreparationVesselsTableName,
						validVesselsTableName,
						validPreparationVesselsTableName,
						validVesselIDColumn,
						validVesselsTableName,
						idColumn,
						validPreparationsTableName,
						validPreparationVesselsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validMeasurementUnitsTableName,
						validVesselsTableName,
						capacityUnitColumn,
						validMeasurementUnitsTableName,
						idColumn,
						pgGen.FilterConditions(validPreparationVesselsTableName, validPreparationVesselsColumns,
							"valid_vessels.archived_at IS NULL",
							"valid_preparations.archived_at IS NULL",
							"valid_measurement_units.archived_at IS NULL",
						),
						pgGen.CursorLimitClause(validPreparationVesselsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPreparationVessel",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	LEFT JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validPreparationVesselsTableName,
						validVesselsTableName, validPreparationVesselsTableName, validVesselIDColumn, validVesselsTableName, idColumn,
						validPreparationsTableName, validPreparationVesselsTableName, validPreparationIDColumn, validPreparationsTableName, idColumn,
						validMeasurementUnitsTableName, validVesselsTableName, capacityUnitColumn, validMeasurementUnitsTableName, idColumn,
						validPreparationVesselsTableName, archivedAtColumn,
						validVesselsTableName, archivedAtColumn,
						validPreparationsTableName, archivedAtColumn,
						validMeasurementUnitsTableName, archivedAtColumn,
						validPreparationVesselsTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPreparationVesselsByIDs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	LEFT JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[]);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validPreparationVesselsTableName,
						validVesselsTableName, validPreparationVesselsTableName, validVesselIDColumn, validVesselsTableName, idColumn,
						validPreparationsTableName, validPreparationVesselsTableName, validPreparationIDColumn, validPreparationsTableName, idColumn,
						validMeasurementUnitsTableName, validVesselsTableName, capacityUnitColumn, validMeasurementUnitsTableName, idColumn,
						validPreparationVesselsTableName, archivedAtColumn,
						validVesselsTableName, archivedAtColumn,
						validPreparationsTableName, archivedAtColumn,
						validMeasurementUnitsTableName, archivedAtColumn,
						validPreparationVesselsTableName, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "ValidPreparationVesselPairIsValid",
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
						validPreparationVesselsTableName,
						validVesselIDColumn,
						validVesselIDColumn,
						validPreparationIDColumn,
						validPreparationIDColumn,
						archivedAtColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
