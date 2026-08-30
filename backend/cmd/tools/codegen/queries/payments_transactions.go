package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	paymentTransactionsTableName = "payment_transactions"
)

var paymentTransactionsColumns = []string{
	idColumn,
	belongsToAccountColumn,
	"subscription_id",
	"purchase_id",
	"external_transaction_id",
	"amount_cents",
	"currency",
	"status",
	createdAtColumn,
}

func buildPaymentsTransactionsQueries(database string) []*Query {
	switch database {
	case postgres:
		fullSelectColumns := applyToEach(paymentTransactionsColumns, func(_ int, s string) string {
			return querygen.Qualify(paymentTransactionsTableName, s)
		})
		accountCondition := fmt.Sprintf("%s.%s = sqlc.arg(%s)", paymentTransactionsTableName, belongsToAccountColumn, belongsToAccountColumn)

		return slices.Concat(
			pgGen.StandardCRUD(paymentTransactionsTableName, paymentTransactionsColumns,
				querygen.WithEntity("PaymentTransaction", "PaymentTransactions"),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetPaymentTransactionsForAccount",
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
						pgGen.FilterCountSelect(paymentTransactionsTableName, paymentTransactionsColumns, nil, accountCondition),
						pgGen.TotalCountSelect(paymentTransactionsTableName, paymentTransactionsColumns, nil, accountCondition),
						paymentTransactionsTableName,
						pgGen.FilterConditions(paymentTransactionsTableName, paymentTransactionsColumns,
							accountCondition,
							accountCondition,
						),
						pgGen.CursorLimitClause(paymentTransactionsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
