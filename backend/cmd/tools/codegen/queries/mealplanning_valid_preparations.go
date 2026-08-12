package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	validPreparationsTableName  = "valid_preparations"
	preparationIDColumn         = "preparation_id"
	validPreparationIDColumn    = "valid_preparation_id"
	restrictToIngredientsColumn = "restrict_to_ingredients"
)

func init() {
	registerTableName(validPreparationsTableName)
}

var validPreparationsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	iconPathColumn,
	"yields_nothing",
	restrictToIngredientsColumn,
	"past_tense",
	slugColumn,
	"minimum_ingredient_count",
	"maximum_ingredient_count",
	"minimum_instrument_count",
	"maximum_instrument_count",
	"temperature_required",
	"time_estimate_required",
	"condition_expression_required",
	"consumes_vessel",
	"only_for_vessels",
	"minimum_vessel_count",
	"maximum_vessel_count",
	lastIndexedAtColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidPreparationsQueries(database string) []*Query {
	switch database {
	case postgres:

		insertColumns := filterForInsert(validPreparationsColumns)

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveValidPreparation",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s IS NULL AND %s = sqlc.arg(%s);`,
					validPreparationsTableName,
					archivedAtColumn,
					currentTimeExpression,
					archivedAtColumn,
					idColumn,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CreateValidPreparation",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					validPreparationsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CheckValidPreparationExistence",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
					validPreparationsTableName, idColumn,
					validPreparationsTableName,
					validPreparationsTableName, archivedAtColumn,
					validPreparationsTableName, idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetValidPreparations",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE
	%s.%s IS NULL
	%s
GROUP BY %s.%s
%s;`,
					strings.Join(applyToEach(validPreparationsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validPreparationsTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(validPreparationsTableName, true, true, []string{}),
					buildTotalCountSelect(validPreparationsTableName, true, []string{}),
					validPreparationsTableName,
					validPreparationsTableName, archivedAtColumn,
					buildFilterConditions(
						validPreparationsTableName,
						true,
						true,
					),
					validPreparationsTableName, idColumn, buildCursorLimitClause(validPreparationsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ScanValidPreparationIDsForReindex",
					Type: ManyType,
				},
				Content: buildReindexScanQuery(validPreparationsTableName),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetValidPreparationsNeedingIndexing",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
WHERE %s.%s IS NULL
	AND (
	%s.%s IS NULL
	OR %s.%s < %s - '24 hours'::INTERVAL
);`,
					validPreparationsTableName, idColumn,
					validPreparationsTableName,
					validPreparationsTableName, archivedAtColumn,
					validPreparationsTableName, lastIndexedAtColumn,
					validPreparationsTableName, lastIndexedAtColumn, currentTimeExpression,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetValidPreparation",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
					strings.Join(applyToEach(validPreparationsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validPreparationsTableName, s)
					}), ",\n\t"),
					validPreparationsTableName,
					validPreparationsTableName, archivedAtColumn,
					validPreparationsTableName, idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetRandomValidPreparation",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
ORDER BY RANDOM() LIMIT 1;`,
					strings.Join(applyToEach(validPreparationsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validPreparationsTableName, s)
					}), ",\n\t"),
					validPreparationsTableName,
					validPreparationsTableName, archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetValidPreparationsWithIDs",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[]);`,
					strings.Join(applyToEach(validPreparationsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validPreparationsTableName, s)
					}), ",\n\t"),
					validPreparationsTableName,
					validPreparationsTableName, archivedAtColumn,
					validPreparationsTableName, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "SearchForValidPreparations",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s %s
	%s
%s;`,
					strings.Join(applyToEach(validPreparationsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validPreparationsTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(validPreparationsTableName, true, true, []string{}),
					buildTotalCountSelect(validPreparationsTableName, true, []string{}),
					validPreparationsTableName,
					validPreparationsTableName,
					archivedAtColumn,
					validPreparationsTableName,
					nameColumn,
					buildILIKEForArgument("name_query"),
					buildFilterConditions(
						validPreparationsTableName,
						true,
						true,
					),
					buildCursorLimitClause(validPreparationsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateValidPreparation",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
					validPreparationsTableName,
					strings.Join(applyToEach(filterForUpdate(validPreparationsColumns), func(i int, s string) string {
						return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
					}), ",\n\t"),
					lastUpdatedAtColumn, currentTimeExpression,
					archivedAtColumn,
					idColumn, idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateValidPreparationLastIndexedAt",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s = sqlc.arg(%s) AND %s IS NULL;`,
					validPreparationsTableName,
					lastIndexedAtColumn,
					currentTimeExpression,
					idColumn,
					idColumn,
					archivedAtColumn,
				)),
			},
		}
	default:
		return nil
	}
}
