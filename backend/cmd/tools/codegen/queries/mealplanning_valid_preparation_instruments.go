package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	// The two joins every read of this bridge table makes, spelled once so the
	// projections below cannot drift apart.
	joinValidInstruments  = "JOIN valid_instruments ON valid_preparation_instruments.valid_instrument_id = valid_instruments.id"
	joinValidPreparations = "JOIN valid_preparations ON valid_preparation_instruments.valid_preparation_id = valid_preparations.id"

	validPreparationInstrumentsTableName = "valid_preparation_instruments"
)

var validPreparationInstrumentsColumns = []string{
	idColumn,
	notesColumn,
	validPreparationIDColumn,
	validInstrumentIDColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidPreparationInstrumentsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			mergeColumns(
				applyToEach(filterFromSlice(validPreparationInstrumentsColumns, "valid_preparation_id", "valid_instrument_id"), func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_preparation_instrument_%s", validPreparationInstrumentsTableName, s, s)
				}),
				applyToEach(validInstrumentsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_instrument_%s", validInstrumentsTableName, s, s)
				}),
				2,
			),
			applyToEach(validPreparationsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_preparation_%s", validPreparationsTableName, s, s)
			}),
			2,
		)

		return slices.Concat(
			pgGen.StandardCRUD(validPreparationInstrumentsTableName, validPreparationInstrumentsColumns,
				querygen.WithEntity("ValidPreparationInstrument", "ValidPreparationInstruments"),
				querygen.WithOmitted(querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPreparationInstrumentsForInstrument",
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
GROUP BY
	%s.%s,
	%s.%s,
	%s.%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(
							validPreparationInstrumentsTableName,
							validPreparationInstrumentsColumns,
							[]string{
								joinValidInstruments,
								joinValidPreparations,
							},
							fmt.Sprintf("%s.%s IS NULL", validInstrumentsTableName, archivedAtColumn),
							fmt.Sprintf("%s.%s IS NULL", validPreparationsTableName, archivedAtColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationInstrumentsTableName, validInstrumentIDColumn, idColumn),
						),
						pgGen.TotalCountSelect(
							validPreparationInstrumentsTableName,
							validPreparationInstrumentsColumns,
							[]string{
								joinValidInstruments,
								joinValidPreparations,
							},
							fmt.Sprintf("%s.%s IS NULL", validInstrumentsTableName, archivedAtColumn),
							fmt.Sprintf("%s.%s IS NULL", validPreparationsTableName, archivedAtColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationInstrumentsTableName, validInstrumentIDColumn, idColumn),
						),
						validPreparationInstrumentsTableName,
						validInstrumentsTableName,
						validPreparationInstrumentsTableName,
						validInstrumentIDColumn,
						validInstrumentsTableName,
						idColumn,
						validPreparationsTableName,
						validPreparationInstrumentsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						pgGen.FilterConditions(validPreparationInstrumentsTableName, validPreparationInstrumentsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationInstrumentsTableName, validInstrumentIDColumn, idColumn),
							fmt.Sprintf("%s.%s IS NULL", ///
								validInstrumentsTableName, archivedAtColumn),
							"valid_preparations.archived_at IS NULL",
						),
						validPreparationInstrumentsTableName,
						idColumn,
						validPreparationsTableName,
						idColumn,
						validInstrumentsTableName,
						idColumn,
						pgGen.CursorLimitClause(validPreparationInstrumentsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPreparationInstrumentsForPreparation",
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
GROUP BY
	%s.%s,
	%s.%s,
	%s.%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(
							validPreparationInstrumentsTableName,
							validPreparationInstrumentsColumns,
							[]string{
								joinValidInstruments,
								joinValidPreparations,
							},
							fmt.Sprintf("%s.%s IS NULL", validInstrumentsTableName, archivedAtColumn),
							fmt.Sprintf("%s.%s IS NULL", validPreparationsTableName, archivedAtColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationInstrumentsTableName, validPreparationIDColumn, idColumn),
						),
						pgGen.TotalCountSelect(
							validPreparationInstrumentsTableName,
							validPreparationInstrumentsColumns,
							[]string{
								joinValidInstruments,
								joinValidPreparations,
							},
							fmt.Sprintf("%s.%s IS NULL", validInstrumentsTableName, archivedAtColumn),
							fmt.Sprintf("%s.%s IS NULL", validPreparationsTableName, archivedAtColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationInstrumentsTableName, validPreparationIDColumn, idColumn),
						),
						validPreparationInstrumentsTableName,
						validInstrumentsTableName,
						validPreparationInstrumentsTableName,
						validInstrumentIDColumn,
						validInstrumentsTableName,
						idColumn,
						validPreparationsTableName,
						validPreparationInstrumentsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						pgGen.FilterConditions(validPreparationInstrumentsTableName, validPreparationInstrumentsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", validPreparationInstrumentsTableName, validPreparationIDColumn, idColumn),
							fmt.Sprintf("%s.%s IS NULL", ///
								validInstrumentsTableName, archivedAtColumn),
							"valid_preparations.archived_at IS NULL",
						),
						validPreparationInstrumentsTableName,
						idColumn,
						validPreparationsTableName,
						idColumn,
						validInstrumentsTableName,
						idColumn,
						pgGen.CursorLimitClause(validPreparationInstrumentsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPreparationInstruments",
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
GROUP BY
	%s.%s,
	%s.%s,
	%s.%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(
							validPreparationInstrumentsTableName,
							validPreparationInstrumentsColumns,
							[]string{
								joinValidInstruments,
								joinValidPreparations,
							},
							fmt.Sprintf("%s.%s IS NULL", validInstrumentsTableName, archivedAtColumn),
							fmt.Sprintf("%s.%s IS NULL", validPreparationsTableName, archivedAtColumn),
						),
						pgGen.TotalCountSelect(
							validPreparationInstrumentsTableName,
							validPreparationInstrumentsColumns,
							[]string{
								joinValidInstruments,
								joinValidPreparations,
							},
							fmt.Sprintf("%s.%s IS NULL", validInstrumentsTableName, archivedAtColumn),
							fmt.Sprintf("%s.%s IS NULL", validPreparationsTableName, archivedAtColumn),
						),
						validPreparationInstrumentsTableName,
						validInstrumentsTableName,
						validPreparationInstrumentsTableName,
						validInstrumentIDColumn,
						validInstrumentsTableName,
						idColumn,
						validPreparationsTableName,
						validPreparationInstrumentsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						pgGen.FilterConditions(validPreparationInstrumentsTableName, validPreparationInstrumentsColumns, querygen.Ascending,
							"valid_instruments.archived_at IS NULL",
							"valid_preparations.archived_at IS NULL",
						),
						validPreparationInstrumentsTableName,
						idColumn,
						validPreparationsTableName,
						idColumn,
						validInstrumentsTableName,
						idColumn,
						pgGen.CursorLimitClause(validPreparationInstrumentsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPreparationInstrument",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM
	%s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validPreparationInstrumentsTableName,
						validInstrumentsTableName, validPreparationInstrumentsTableName, validInstrumentIDColumn, validInstrumentsTableName, idColumn,
						validPreparationsTableName, validPreparationInstrumentsTableName, validPreparationIDColumn, validPreparationsTableName, idColumn,
						validPreparationInstrumentsTableName, archivedAtColumn,
						validInstrumentsTableName, archivedAtColumn,
						validPreparationsTableName, archivedAtColumn,
						validPreparationInstrumentsTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPreparationInstrumentsByIDs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM
	%s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[]);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validPreparationInstrumentsTableName,
						validInstrumentsTableName, validPreparationInstrumentsTableName, validInstrumentIDColumn, validInstrumentsTableName, idColumn,
						validPreparationsTableName, validPreparationInstrumentsTableName, validPreparationIDColumn, validPreparationsTableName, idColumn,
						validPreparationInstrumentsTableName, archivedAtColumn,
						validInstrumentsTableName, archivedAtColumn,
						validPreparationsTableName, archivedAtColumn,
						validPreparationInstrumentsTableName, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "ValidPreparationInstrumentPairIsValid",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS(
	SELECT %s.%s
	FROM %s
	WHERE %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s)
	AND %s IS NULL
);`,
						validPreparationInstrumentsTableName, idColumn,
						validPreparationInstrumentsTableName,
						validInstrumentIDColumn, validInstrumentIDColumn,
						validPreparationIDColumn, validPreparationIDColumn,
						archivedAtColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
