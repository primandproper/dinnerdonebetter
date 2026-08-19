package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	purchasesTableName = "purchases"
)

func init() {
	registerTableName(purchasesTableName)
}

var purchasesColumns = []string{
	idColumn,
	belongsToAccountColumn,
	"product_id",
	"amount_cents",
	"currency",
	"completed_at",
	"external_transaction_id",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildPaymentsPurchasesQueries(database string) []*Query {
	switch database {
	case postgres:
		fullSelectColumns := applyToEach(purchasesColumns, func(_ int, s string) string {
			return querygen.Qualify(purchasesTableName, s)
		})
		accountCondition := fmt.Sprintf("%s.%s = sqlc.arg(%s)", purchasesTableName, belongsToAccountColumn, belongsToAccountColumn)

		return slices.Concat(
			querygen.StandardCRUD(purchasesTableName, purchasesColumns,
				querygen.WithEntity("Purchase", "Purchases"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetPurchasesForAccount",
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
						querygen.FilterCountSelect(purchasesTableName, purchasesColumns, nil, accountCondition),
						querygen.TotalCountSelect(purchasesTableName, purchasesColumns, nil, accountCondition),
						purchasesTableName,
						querygen.FilterConditions(purchasesTableName, purchasesColumns,
							accountCondition,
							accountCondition,
						),
						querygen.CursorLimitClause(purchasesTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
