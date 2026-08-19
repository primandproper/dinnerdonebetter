package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	accountsTableName = "accounts"

	/* #nosec G101 */
	webhookHMACSecretColumn = "webhook_hmac_secret"
)

func init() {
	registerTableName(accountsTableName)
}

var accountsColumns = []string{
	idColumn,
	nameColumn,
	"billing_status",
	"contact_phone",
	"payment_processor_customer_id",
	"subscription_plan_id",
	belongsToUserColumn,
	"time_zone",
	"address_line_1",
	"address_line_2",
	"city",
	"state",
	"zip_code",
	"country",
	"latitude",
	"longitude",
	"last_payment_provider_sync_occurred_at",
	webhookHMACSecretColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildAccountsQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			querygen.StandardCRUD(accountsTableName, accountsColumns,
				querygen.WithEntity("Account", "Accounts"),
				querygen.WithOwnership(belongsToUserColumn),
				querygen.WithDatabaseOwned("payment_processor_customer_id", "subscription_plan_id", "time_zone", "last_payment_provider_sync_occurred_at"),
				querygen.WithImmutable("billing_status", webhookHMACSecretColumn),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "AddToAccountDuringCreation",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO account_user_memberships (
	%s
) VALUES (
	%s
);`,
						strings.Join(filterForInsert(accountUserMembershipsColumns, "default_account"), ",\n\t"),
						strings.Join(applyToEach(filterForInsert(accountUserMembershipsColumns, "default_account"), func(_ int, s string) string {
							return fmt.Sprintf("sqlc.arg(%s)", s)
						}), ",\n\t"),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveAccount",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						accountsTableName,
						lastUpdatedAtColumn,
						currentTimeExpression,
						archivedAtColumn,
						currentTimeExpression,
						archivedAtColumn,
						belongsToUserColumn,
						belongsToUserColumn,
						idColumn,
						idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetAccountByIDWithMemberships",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
	%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(append(
							append(
								append(
									applyToEach(accountsColumns, func(_ int, s string) string {
										return fmt.Sprintf("%s.%s", accountsTableName, s)
									}),
									applyToEach(usersColumns, func(_ int, s string) string {
										return fmt.Sprintf("%s.%s as user_%s", usersTableName, s, s)
									})...,
								),
								applyToEach(avatarJoinSelect("user_avatar"), func(_ int, s string) string {
									return s
								})...,
							),
							applyToEach(accountUserMembershipsColumns, func(_ int, s string) string {
								return fmt.Sprintf("%s.%s as membership_%s", accountUserMembershipsTableName, s, s)
							})...,
						), ",\n\t"),
						accountsTableName,
						accountUserMembershipsTableName, accountUserMembershipsTableName, belongsToAccountColumn, accountsTableName, idColumn,
						usersTableName, accountUserMembershipsTableName, belongsToUserColumn, usersTableName, idColumn,
						avatarJoinClause,
						accountsTableName, archivedAtColumn,
						accountUserMembershipsTableName, archivedAtColumn,
						accountsTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetAccountsForUser",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
JOIN %s ON %s.%s = %s.%s
WHERE %s
%s;`,
						strings.Join(applyToEach(accountsColumns, func(_ int, s string) string {
							return fmt.Sprintf("%s.%s", accountsTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(accountsTableName, accountsColumns, nil),
						querygen.TotalCountSelect(accountsTableName, accountsColumns, []string{}, fmt.Sprintf("%s.%s = sqlc.arg(%s)", accountUserMembershipsTableName, belongsToUserColumn, belongsToUserColumn)),
						accountsTableName,
						accountUserMembershipsTableName,
						accountUserMembershipsTableName,
						belongsToAccountColumn,
						accountsTableName,
						idColumn,
						querygen.FilterConditions(accountsTableName, accountsColumns,
							"account_user_memberships.archived_at IS NULL",
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", accountUserMembershipsTableName, belongsToUserColumn, belongsToUserColumn),
						),
						querygen.CursorLimitClause(accountsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateAccountBillingFields",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	billing_status = COALESCE(sqlc.narg(billing_status), billing_status),
	subscription_plan_id = COALESCE(sqlc.narg(subscription_plan_id), subscription_plan_id),
	payment_processor_customer_id = COALESCE(sqlc.narg(payment_processor_customer_id), payment_processor_customer_id),
	last_payment_provider_sync_occurred_at = COALESCE(sqlc.narg(last_payment_provider_sync_occurred_at), last_payment_provider_sync_occurred_at),
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						accountsTableName,
						lastUpdatedAtColumn, currentTimeExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateAccountWebhookEncryptionKey",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = sqlc.arg(%s),
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						accountsTableName,
						webhookHMACSecretColumn, webhookHMACSecretColumn,
						lastUpdatedAtColumn, currentTimeExpression,
						archivedAtColumn,
						belongsToUserColumn, belongsToUserColumn,
						idColumn, idColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
