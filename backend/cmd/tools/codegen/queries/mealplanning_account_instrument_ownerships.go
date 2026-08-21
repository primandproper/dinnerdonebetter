package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	accountInstrumentOwnershipsTableName = "account_instrument_ownerships"
)

var accountInstrumentOwnershipsColumns = []string{
	idColumn,
	notesColumn,
	"quantity",
	validInstrumentIDColumn,
	belongsToAccountColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildAccountInstrumentOwnershipQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			applyToEach(filterFromSlice(accountInstrumentOwnershipsColumns, validInstrumentIDColumn), func(i int, s string) string {
				return fmt.Sprintf("%s.%s", accountInstrumentOwnershipsTableName, s)
			}),
			applyToEach(validInstrumentsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_instrument_%s", validInstrumentsTableName, s, s)
			}),
			3,
		)

		return slices.Concat(
			pgGen.StandardCRUD(accountInstrumentOwnershipsTableName, accountInstrumentOwnershipsColumns,
				querygen.WithEntity("AccountInstrumentOwnership", "AccountInstrumentOwnerships"),
				querygen.WithOwnership(belongsToAccountColumn),
				querygen.WithOmitted(querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetAccountInstrumentOwnerships",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
INNER JOIN %s ON %s.%s = %s.%s
WHERE %s
GROUP BY
	%s.%s,
	%s.%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(accountInstrumentOwnershipsTableName, accountInstrumentOwnershipsColumns, []string{}, "account_instrument_ownerships.belongs_to_account = sqlc.arg(account_id)"),
						pgGen.TotalCountSelect(accountInstrumentOwnershipsTableName, accountInstrumentOwnershipsColumns, []string{}, "account_instrument_ownerships.belongs_to_account = sqlc.arg(account_id)"),
						accountInstrumentOwnershipsTableName,
						validInstrumentsTableName,
						accountInstrumentOwnershipsTableName,
						validInstrumentIDColumn,
						validInstrumentsTableName,
						idColumn,
						pgGen.FilterConditions(accountInstrumentOwnershipsTableName, accountInstrumentOwnershipsColumns,
							"account_instrument_ownerships.belongs_to_account = sqlc.arg(account_id)",
						),
						accountInstrumentOwnershipsTableName,
						idColumn,
						validInstrumentsTableName,
						idColumn,
						pgGen.CursorLimitClause(accountInstrumentOwnershipsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetAccountInstrumentOwnership",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
INNER JOIN %s ON %s.%s = %s.%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						accountInstrumentOwnershipsTableName,
						validInstrumentsTableName, accountInstrumentOwnershipsTableName, validInstrumentIDColumn, validInstrumentsTableName, idColumn,
						accountInstrumentOwnershipsTableName,
						archivedAtColumn,
						accountInstrumentOwnershipsTableName, idColumn, idColumn,
						accountInstrumentOwnershipsTableName, belongsToAccountColumn, belongsToAccountColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
