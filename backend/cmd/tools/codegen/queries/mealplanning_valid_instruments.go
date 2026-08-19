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
WHERE %s
GROUP BY %s.%s
%s;`,
						strings.Join(applyToEach(validInstrumentsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validInstrumentsTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(validInstrumentsTableName, validInstrumentsColumns, []string{}),
						querygen.TotalCountSelect(validInstrumentsTableName, validInstrumentsColumns, []string{}),
						validInstrumentsTableName,
						querygen.FilterConditions(validInstrumentsTableName, validInstrumentsColumns),
						validInstrumentsTableName,
						idColumn,
						querygen.CursorLimitClause(validInstrumentsTableName),
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
						querygen.NowExpression,
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
WHERE %s
%s;`,
						strings.Join(applyToEach(validInstrumentsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validInstrumentsTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(validInstrumentsTableName, validInstrumentsColumns, []string{}),
						querygen.TotalCountSelect(validInstrumentsTableName, validInstrumentsColumns, []string{}),
						validInstrumentsTableName,
						querygen.FilterConditions(validInstrumentsTableName, validInstrumentsColumns,
							fmt.Sprintf("%s.%s %s", validInstrumentsTableName, nameColumn, buildILIKEForArgument("name_query")),
						),
						querygen.CursorLimitClause(validInstrumentsTableName),
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
WHERE %s
%s;`,
						strings.Join(applyToEach(validInstrumentsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validInstrumentsTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(validInstrumentsTableName, validInstrumentsColumns, []string{},
							validInstrumentsTableName+".id NOT IN (SELECT valid_instrument_id FROM account_instrument_ownerships WHERE account_instrument_ownerships.belongs_to_account = sqlc.arg(account_id) AND account_instrument_ownerships.archived_at IS NULL)"),
						querygen.TotalCountSelect(validInstrumentsTableName, validInstrumentsColumns, []string{},
							validInstrumentsTableName+".id NOT IN (SELECT valid_instrument_id FROM account_instrument_ownerships WHERE account_instrument_ownerships.belongs_to_account = sqlc.arg(account_id) AND account_instrument_ownerships.archived_at IS NULL)"),
						validInstrumentsTableName,
						querygen.FilterConditions(validInstrumentsTableName, validInstrumentsColumns,
							fmt.Sprintf("%s.%s %s", validInstrumentsTableName, nameColumn, buildILIKEForArgument("name_query")),
							validInstrumentsTableName+".id NOT IN (SELECT valid_instrument_id FROM account_instrument_ownerships WHERE account_instrument_ownerships.belongs_to_account = sqlc.arg(account_id) AND account_instrument_ownerships.archived_at IS NULL)",
						),
						querygen.CursorLimitClause(validInstrumentsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
