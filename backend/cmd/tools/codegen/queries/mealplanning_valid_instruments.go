package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	validInstrumentsTableName = "valid_instruments"

	validInstrumentIDColumn = "valid_instrument_id"
)

func init() {
	registerTableName(validInstrumentsTableName)
}

var validInstrumentsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	iconPathColumn,
	pluralNameColumn,
	"usable_for_storage",
	slugColumn,
	"display_in_summary_lists",
	"include_in_generated_instructions",
	lastIndexedAtColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidInstrumentsQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			querygen.StandardCRUD(validInstrumentsTableName, validInstrumentsColumns,
				querygen.WithEntity("ValidInstrument", "ValidInstruments"),
				querygen.WithOmitted(querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetValidInstruments",
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
						strings.Join(applyToEach(validInstrumentsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validInstrumentsTableName, s)
						}), ",\n\t"),
						buildFilterCountSelect(validInstrumentsTableName, true, true, []string{}),
						buildTotalCountSelect(validInstrumentsTableName, true, []string{}),
						validInstrumentsTableName,
						validInstrumentsTableName,
						archivedAtColumn,
						buildFilterConditions(
							validInstrumentsTableName,
							true,
							true,
						),
						validInstrumentsTableName, idColumn,
						buildCursorLimitClause(validInstrumentsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidInstrumentsNeedingIndexing",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
WHERE %s.%s IS NULL
	AND (
	%s.%s IS NULL
	OR %s.%s < %s - '24 hours'::INTERVAL
);`,
						validInstrumentsTableName,
						idColumn,
						validInstrumentsTableName,
						validInstrumentsTableName,
						archivedAtColumn,
						validInstrumentsTableName,
						lastIndexedAtColumn,
						validInstrumentsTableName,
						lastIndexedAtColumn,
						currentTimeExpression,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetRandomValidInstrument",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
ORDER BY RANDOM() LIMIT 1;`,
						strings.Join(applyToEach(validInstrumentsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validInstrumentsTableName, s)
						}), ",\n\t"),
						validInstrumentsTableName,
						validInstrumentsTableName,
						archivedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidInstrumentsWithIDs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[]);`,
						strings.Join(applyToEach(validInstrumentsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validInstrumentsTableName, s)
						}), ",\n\t"),
						validInstrumentsTableName,
						validInstrumentsTableName,
						archivedAtColumn,
						validInstrumentsTableName,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchForValidInstruments",
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
						strings.Join(applyToEach(validInstrumentsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validInstrumentsTableName, s)
						}), ",\n\t"),
						buildFilterCountSelect(validInstrumentsTableName, true, true, []string{}),
						buildTotalCountSelect(validInstrumentsTableName, true, []string{}),
						validInstrumentsTableName,
						validInstrumentsTableName,
						archivedAtColumn,
						validInstrumentsTableName,
						nameColumn,
						buildILIKEForArgument("name_query"),
						buildFilterConditions(
							validInstrumentsTableName,
							true,
							true,
						),
						buildCursorLimitClause(validInstrumentsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchForValidInstrumentsNotOwnedByAccount",
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
						strings.Join(applyToEach(validInstrumentsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validInstrumentsTableName, s)
						}), ",\n\t"),
						buildFilterCountSelect(validInstrumentsTableName, true, true, []string{},
							validInstrumentsTableName+".id NOT IN (SELECT valid_instrument_id FROM account_instrument_ownerships WHERE account_instrument_ownerships.belongs_to_account = sqlc.arg(account_id) AND account_instrument_ownerships.archived_at IS NULL)"),
						buildTotalCountSelect(validInstrumentsTableName, true, []string{},
							validInstrumentsTableName+".id NOT IN (SELECT valid_instrument_id FROM account_instrument_ownerships WHERE account_instrument_ownerships.belongs_to_account = sqlc.arg(account_id) AND account_instrument_ownerships.archived_at IS NULL)"),
						validInstrumentsTableName,
						validInstrumentsTableName,
						archivedAtColumn,
						validInstrumentsTableName,
						nameColumn,
						buildILIKEForArgument("name_query"),
						buildFilterConditions(
							validInstrumentsTableName,
							true,
							true,
							validInstrumentsTableName+".id NOT IN (SELECT valid_instrument_id FROM account_instrument_ownerships WHERE account_instrument_ownerships.belongs_to_account = sqlc.arg(account_id) AND account_instrument_ownerships.archived_at IS NULL)",
						),
						buildCursorLimitClause(validInstrumentsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateValidInstrumentLastIndexedAt",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s = sqlc.arg(%s) AND %s IS NULL;`,
						validInstrumentsTableName,
						lastIndexedAtColumn,
						currentTimeExpression,
						idColumn,
						idColumn,
						archivedAtColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
