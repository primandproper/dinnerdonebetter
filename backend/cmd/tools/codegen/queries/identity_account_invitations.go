package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	accountInvitationsTableName        = "account_invitations"
	destinationAccountColumn           = "destination_account"
	fromUserColumn                     = "from_user"
	toUserColumn                       = "to_user"
	toEmailColumn                      = "to_email"
	accountInvitationsTokenColumn      = "token"
	accountInvitationsStatusColumn     = "status"
	accountInvitationsStatusNoteColumn = "status_note"
	accountInvitationsExpiresAtColumn  = "expires_at"
)

func init() {
	registerTableName(accountInvitationsTableName)
}

var accountInvitationsColumns = []string{
	idColumn,
	fromUserColumn,
	toUserColumn,
	"to_name",
	"note",
	toEmailColumn,
	accountInvitationsTokenColumn,
	destinationAccountColumn,
	accountInvitationsExpiresAtColumn,
	accountInvitationsStatusColumn,
	accountInvitationsStatusNoteColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildAccountInvitationsQueries(database string) []*Query {
	switch database {
	case postgres:

		userWithAvatarColumns := append(
			applyToEach(usersColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as user_%s", usersTableName, s, s)
			}),
			avatarJoinSelect("user_avatar")...,
		)
		fullSelectColumns := mergeColumns(mergeColumns(
			applyToEach(accountInvitationsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s", accountInvitationsTableName, s)
			}),
			userWithAvatarColumns,
			3,
		),
			applyToEach(accountsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as account_%s", accountsTableName, s, s)
			}),
			1,
		)

		return slices.Concat(
			pgGen.StandardCRUD(accountInvitationsTableName, accountInvitationsColumns,
				querygen.WithEntity("AccountInvitation", "AccountInvitations"),
				querygen.WithDatabaseOwned("status", accountInvitationsStatusNoteColumn),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "AttachAccountInvitationsToUserID",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	%s = %s
WHERE %s IS NULL
	AND %s = LOWER(sqlc.arg(%s));`,
						accountInvitationsTableName,
						toUserColumn, toUserColumn,
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						toEmailColumn, toEmailColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "AssignInvitationsToUserByEmail",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND %s = LOWER(sqlc.arg(%s))`,
						accountInvitationsTableName,
						toUserColumn, toUserColumn,
						toEmailColumn, emailAddressColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetAccountInvitationByEmailAndToken",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	%s
WHERE %s.%s IS NULL
	AND %s.%s > %s
	AND %s.%s = LOWER(sqlc.arg(%s))
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						accountInvitationsTableName,
						accountsTableName, accountInvitationsTableName, destinationAccountColumn, accountsTableName, idColumn,
						usersTableName, accountInvitationsTableName, fromUserColumn, usersTableName, idColumn,
						avatarJoinClause,
						accountInvitationsTableName, archivedAtColumn,
						accountInvitationsTableName, accountInvitationsExpiresAtColumn, querygen.NowExpression,
						accountInvitationsTableName, toEmailColumn, toEmailColumn,
						accountInvitationsTableName, accountInvitationsTokenColumn, accountInvitationsTokenColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetAccountInvitationByAccountAndID",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	%s
WHERE %s.%s IS NULL
	AND %s.%s > %s
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						accountInvitationsTableName,
						accountsTableName, accountInvitationsTableName, destinationAccountColumn, accountsTableName, idColumn,
						usersTableName, accountInvitationsTableName, fromUserColumn, usersTableName, idColumn,
						avatarJoinClause,
						accountInvitationsTableName, archivedAtColumn,
						accountInvitationsTableName, accountInvitationsExpiresAtColumn, querygen.NowExpression,
						accountInvitationsTableName, destinationAccountColumn, destinationAccountColumn,
						accountInvitationsTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetAccountInvitationByTokenAndID",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	%s
WHERE %s.%s IS NULL
	AND %s.%s > %s
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						accountInvitationsTableName,
						accountsTableName, accountInvitationsTableName, destinationAccountColumn, accountsTableName, idColumn,
						usersTableName, accountInvitationsTableName, fromUserColumn, usersTableName, idColumn,
						avatarJoinClause,
						accountInvitationsTableName, archivedAtColumn,
						accountInvitationsTableName, accountInvitationsExpiresAtColumn, querygen.NowExpression,
						accountInvitationsTableName, accountInvitationsTokenColumn, accountInvitationsTokenColumn,
						accountInvitationsTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetAccountInvitationByToken",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	%s
WHERE %s.%s IS NULL
	AND %s.%s > %s
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						accountInvitationsTableName,
						accountsTableName, accountInvitationsTableName, destinationAccountColumn, accountsTableName, idColumn,
						usersTableName, accountInvitationsTableName, fromUserColumn, usersTableName, idColumn,
						avatarJoinClause,
						accountInvitationsTableName, archivedAtColumn,
						accountInvitationsTableName, accountInvitationsExpiresAtColumn, querygen.NowExpression,
						accountInvitationsTableName, accountInvitationsTokenColumn, accountInvitationsTokenColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetPendingInvitesFromUser",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	%s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(accountInvitationsTableName, accountInvitationsColumns, []string{}),
						pgGen.TotalCountSelect(accountInvitationsTableName, accountInvitationsColumns, []string{}),
						accountInvitationsTableName,
						accountsTableName,
						accountInvitationsTableName,
						destinationAccountColumn,
						accountsTableName,
						idColumn,
						usersTableName,
						accountInvitationsTableName,
						fromUserColumn,
						usersTableName,
						idColumn,
						avatarJoinClause,
						pgGen.FilterConditions(accountInvitationsTableName, accountInvitationsColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", accountInvitationsTableName, fromUserColumn, fromUserColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", accountInvitationsTableName, accountInvitationsStatusColumn, accountInvitationsStatusColumn),
						),
						pgGen.CursorLimitClause(accountInvitationsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetPendingInvitesForUser",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	%s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(accountInvitationsTableName, accountInvitationsColumns, []string{}),
						pgGen.TotalCountSelect(accountInvitationsTableName, accountInvitationsColumns, []string{}),
						accountInvitationsTableName,
						accountsTableName,
						accountInvitationsTableName,
						destinationAccountColumn,
						accountsTableName,
						idColumn,
						usersTableName,
						accountInvitationsTableName,
						fromUserColumn,
						usersTableName,
						idColumn,
						avatarJoinClause,
						pgGen.FilterConditions(accountInvitationsTableName, accountInvitationsColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", accountInvitationsTableName, toUserColumn, toUserColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", accountInvitationsTableName, accountInvitationsStatusColumn, accountInvitationsStatusColumn),
						),
						pgGen.CursorLimitClause(accountInvitationsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SetAccountInvitationStatus",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	%s = sqlc.arg(%s),
	%s = %s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						accountInvitationsTableName,
						accountInvitationsStatusColumn, accountInvitationsStatusColumn,
						accountInvitationsStatusNoteColumn, accountInvitationsStatusNoteColumn,
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
