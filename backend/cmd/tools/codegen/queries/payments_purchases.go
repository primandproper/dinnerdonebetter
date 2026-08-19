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
			return fullColumnName(purchasesTableName, s)
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
WHERE %s.%s IS NULL
	AND %s
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(purchasesTableName, true, true, nil, accountCondition),
						buildTotalCountSelect(purchasesTableName, true, nil, accountCondition),
						purchasesTableName,
						purchasesTableName, archivedAtColumn,
						accountCondition,
						buildFilterConditions(purchasesTableName, true, false, accountCondition),
						buildCursorLimitClause(purchasesTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
