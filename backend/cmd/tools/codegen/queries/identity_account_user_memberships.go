package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	accountUserMembershipsTableName = "account_user_memberships"

	defaultAccountColumn = "default_account"
)

func init() {
	registerTableName(accountUserMembershipsTableName)
}

var accountUserMembershipsColumns = []string{
	idColumn,
	belongsToAccountColumn,
	belongsToUserColumn,
	defaultAccountColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildAccountUserMembershipsQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			querygen.StandardCRUD(accountUserMembershipsTableName, accountUserMembershipsColumns,
				querygen.WithEntity("AccountUserMemberships", "AccountUserMembershipss"),
				querygen.WithDatabaseOwned(defaultAccountColumn),
				querygen.WithQueryName(querygen.CreateQuery, "AddUserToAccount"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveUserMemberships",
						Type: ExecRowsType,
					},
					Content: buildUpdateAccountMembershipsQuery(belongsToUserColumn, []string{}),
				},
				{
					Annotation: QueryAnnotation{
						Name: "CreateAccountUserMembershipForNewUser",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
						accountUserMembershipsTableName,
						strings.Join(querygen.ForInsert(accountUserMembershipsColumns), ",\n\t"),
						strings.Join(applyToEach(querygen.ForInsert(accountUserMembershipsColumns), func(i int, s string) string {
							return fmt.Sprintf("sqlc.arg(%s)", s)
						}), ",\n\t"),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetDefaultAccountIDForUser",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s = TRUE;`,
						accountsTableName, idColumn,
						accountsTableName,
						accountUserMembershipsTableName, accountUserMembershipsTableName, belongsToAccountColumn, accountsTableName, idColumn,
						accountUserMembershipsTableName, belongsToUserColumn, belongsToUserColumn,
						accountUserMembershipsTableName, defaultAccountColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetAccountUserMembershipsForUser",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(applyToEach(accountUserMembershipsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", accountUserMembershipsTableName, s)
						}), ",\n\t"),
						accountUserMembershipsTableName,
						accountsTableName, accountsTableName, idColumn, accountUserMembershipsTableName, belongsToAccountColumn,
						accountUserMembershipsTableName, archivedAtColumn,
						accountUserMembershipsTableName, belongsToUserColumn, belongsToUserColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "MarkAccountUserMembershipAsUserDefault",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = (%s = sqlc.arg(%s) AND %s = sqlc.arg(%s))
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						accountUserMembershipsTableName,
						defaultAccountColumn, belongsToUserColumn, belongsToUserColumn, belongsToAccountColumn, belongsToAccountColumn,
						archivedAtColumn,
						belongsToUserColumn, belongsToUserColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "RemoveUserFromAccount",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s,
	%s = 'false'
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s);`,
						accountUserMembershipsTableName,
						archivedAtColumn, querygen.NowExpression,
						defaultAccountColumn,
						accountUserMembershipsTableName, archivedAtColumn,
						accountUserMembershipsTableName, belongsToAccountColumn, belongsToAccountColumn,
						accountUserMembershipsTableName, belongsToUserColumn, belongsToUserColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "TransferAccountMembership",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s)
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						accountUserMembershipsTableName,
						belongsToUserColumn,
						belongsToUserColumn,
						archivedAtColumn,
						belongsToAccountColumn,
						belongsToAccountColumn,
						belongsToUserColumn,
						belongsToUserColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "TransferAccountOwnership",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(new_owner)
WHERE %s IS NULL
	AND %s = sqlc.arg(old_owner)
	AND %s = sqlc.arg(account_id);`,
						accountsTableName,
						belongsToUserColumn,
						archivedAtColumn,
						belongsToUserColumn,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UserIsAccountMember",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.%s = sqlc.arg(%s)
);`,
						accountUserMembershipsTableName, idColumn,
						accountUserMembershipsTableName,
						accountUserMembershipsTableName, archivedAtColumn,
						accountUserMembershipsTableName, belongsToAccountColumn, belongsToAccountColumn,
						accountUserMembershipsTableName, belongsToUserColumn, belongsToUserColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
