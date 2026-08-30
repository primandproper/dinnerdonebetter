package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	purchasesTableName = "purchases"
)

var purchasesColumns = []string{
	idColumn,
	belongsToAccountColumn,
	"product_id",
	amountCentsColumn,
	currencyColumn,
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
			pgGen.StandardCRUD(purchasesTableName, purchasesColumns,
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
						pgGen.FilterCountSelect(purchasesTableName, purchasesColumns, nil, accountCondition),
						pgGen.TotalCountSelect(purchasesTableName, purchasesColumns, nil, accountCondition),
						purchasesTableName,
						pgGen.FilterConditions(purchasesTableName, purchasesColumns, querygen.Ascending,
							accountCondition,
							accountCondition,
						),
						pgGen.CursorLimitClause(purchasesTableName, querygen.Ascending),
					)),
				},
			},
		)
	default:
		return nil
	}
}
