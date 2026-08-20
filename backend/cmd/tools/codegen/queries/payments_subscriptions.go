package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	subscriptionsTableName       = "subscriptions"
	externalSubscriptionIDColumn = "external_subscription_id"
)

func init() {
	registerTableName(subscriptionsTableName)
}

var subscriptionsColumns = []string{
	idColumn,
	belongsToAccountColumn,
	"product_id",
	externalSubscriptionIDColumn,
	"status",
	"current_period_start",
	"current_period_end",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildPaymentsSubscriptionsQueries(database string) []*Query {
	switch database {
	case postgres:
		fullSelectColumns := applyToEach(subscriptionsColumns, func(_ int, s string) string {
			return querygen.Qualify(subscriptionsTableName, s)
		})
		accountCondition := fmt.Sprintf("%s.%s = sqlc.arg(%s)", subscriptionsTableName, belongsToAccountColumn, belongsToAccountColumn)

		return slices.Concat(
			pgGen.StandardCRUD(subscriptionsTableName, subscriptionsColumns,
				querygen.WithEntity("Subscription", "Subscriptions"),
				querygen.WithImmutable(belongsToAccountColumn, "product_id"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveSubscription",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s, %s = %s WHERE %s IS NULL AND %s = sqlc.arg(%s);`,
						subscriptionsTableName,
						archivedAtColumn, querygen.NowExpression,
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetSubscriptionByExternalID",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						subscriptionsTableName,
						subscriptionsTableName, archivedAtColumn,
						subscriptionsTableName, externalSubscriptionIDColumn, externalSubscriptionIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetSubscriptionsForAccount",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(subscriptionsTableName, subscriptionsColumns, nil, accountCondition),
						pgGen.TotalCountSelect(subscriptionsTableName, subscriptionsColumns, nil, accountCondition),
						subscriptionsTableName,
						pgGen.FilterConditions(subscriptionsTableName, subscriptionsColumns,
							accountCondition,
							accountCondition,
						),
						pgGen.CursorLimitClause(subscriptionsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateSubscriptionStatus",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET status = sqlc.arg(status), %s = %s
WHERE %s IS NULL AND %s = sqlc.arg(%s);`,
						subscriptionsTableName,
						lastUpdatedAtColumn, querygen.NowExpression,
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
