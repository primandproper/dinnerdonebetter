package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	paymentTransactionsTableName = "payment_transactions"
)

func init() {
	registerTableName(paymentTransactionsTableName)
}

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
			return fullColumnName(paymentTransactionsTableName, s)
		})
		accountCondition := fmt.Sprintf("%s.%s = sqlc.arg(%s)", paymentTransactionsTableName, belongsToAccountColumn, belongsToAccountColumn)

		return slices.Concat(
			querygen.StandardCRUD(paymentTransactionsTableName, paymentTransactionsColumns,
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
						querygen.FilterCountSelect(paymentTransactionsTableName, paymentTransactionsColumns, nil, accountCondition),
						querygen.TotalCountSelect(paymentTransactionsTableName, paymentTransactionsColumns, nil, accountCondition),
						paymentTransactionsTableName,
						querygen.FilterConditions(paymentTransactionsTableName, paymentTransactionsColumns,
							accountCondition,
							accountCondition,
						),
						querygen.CursorLimitClause(paymentTransactionsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
