package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	productsTableName = "products"
)

var productsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	"kind",
	amountCentsColumn,
	currencyColumn,
	"billing_interval_months",
	"external_product_id",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildPaymentsProductsQueries(database string) []*Query {
	switch database {
	case postgres:
		fullSelectColumns := applyToEach(productsColumns, func(_ int, s string) string {
			return querygen.Qualify(productsTableName, s)
		})

		return slices.Concat(
			pgGen.StandardCRUD(productsTableName, productsColumns,
				querygen.WithEntity("Product", "Products"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveProduct",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s, %s = %s WHERE %s IS NULL AND %s = sqlc.arg(%s);`,
						productsTableName,
						archivedAtColumn, querygen.NowExpression,
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetProductByExternalID",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
AND %s.external_product_id = sqlc.arg(external_product_id);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						productsTableName,
						productsTableName, archivedAtColumn,
						productsTableName,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetProducts",
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
						pgGen.FilterCountSelect(productsTableName, productsColumns, []string{}),
						pgGen.TotalCountSelect(productsTableName, productsColumns, []string{}),
						productsTableName,
						pgGen.FilterConditions(productsTableName, productsColumns, querygen.Ascending),
						pgGen.CursorLimitClause(productsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchForProducts",
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
						pgGen.FilterCountSelect(productsTableName, productsColumns, []string{}),
						pgGen.TotalCountSelect(productsTableName, productsColumns, []string{}),
						productsTableName,
						pgGen.FilterConditions(productsTableName, productsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s %s", productsTableName, nameColumn, buildILIKEForArgument("name_query")),
						),
						pgGen.CursorLimitClause(productsTableName, querygen.Ascending),
					)),
				},
			},
		)
	default:
		return nil
	}
}
