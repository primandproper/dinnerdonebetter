package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	productsTableName = "products"
)

func init() {
	registerTableName(productsTableName)
}

var productsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	"kind",
	"amount_cents",
	"currency",
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
			querygen.StandardCRUD(productsTableName, productsColumns,
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
						querygen.FilterCountSelect(productsTableName, productsColumns, []string{}),
						querygen.TotalCountSelect(productsTableName, productsColumns, []string{}),
						productsTableName,
						querygen.FilterConditions(productsTableName, productsColumns),
						querygen.CursorLimitClause(productsTableName),
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
						querygen.FilterCountSelect(productsTableName, productsColumns, []string{}),
						querygen.TotalCountSelect(productsTableName, productsColumns, []string{}),
						productsTableName,
						querygen.FilterConditions(productsTableName, productsColumns,
							fmt.Sprintf("%s.%s %s", productsTableName, nameColumn, buildILIKEForArgument("name_query")),
						),
						querygen.CursorLimitClause(productsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
