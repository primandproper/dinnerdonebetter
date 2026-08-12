package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	validIngredientStatesTableName = "valid_ingredient_states"
)

func init() {
	registerTableName(validIngredientStatesTableName)
}

var validIngredientStatesColumns = []string{
	idColumn,
	nameColumn,
	"past_tense",
	slugColumn,
	descriptionColumn,
	iconPathColumn,
	"attribute_type",
	lastIndexedAtColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidIngredientStatesQueries(database string) []*Query {
	switch database {
	case postgres:

		insertColumns := filterForInsert(validIngredientStatesColumns)

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveValidIngredientState",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s IS NULL AND %s = sqlc.arg(%s);`,
					validIngredientStatesTableName,
					archivedAtColumn,
					currentTimeExpression,
					archivedAtColumn,
					idColumn,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CreateValidIngredientState",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					validIngredientStatesTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CheckValidIngredientStateExistence",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
					validIngredientStatesTableName, idColumn,
					validIngredientStatesTableName,
					validIngredientStatesTableName,
					archivedAtColumn,
					validIngredientStatesTableName,
					idColumn,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetValidIngredientStates",
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
					strings.Join(applyToEach(validIngredientStatesColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validIngredientStatesTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(validIngredientStatesTableName, true, true, []string{}),
					buildTotalCountSelect(validIngredientStatesTableName, true, []string{}),
					validIngredientStatesTableName,
					validIngredientStatesTableName,
					archivedAtColumn,
					buildFilterConditions(
						validIngredientStatesTableName,
						true,
						true,
					),
					validIngredientStatesTableName, idColumn,
					buildCursorLimitClause(validIngredientStatesTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ScanValidIngredientStateIDsForReindex",
					Type: ManyType,
				},
				Content: buildReindexScanQuery(validIngredientStatesTableName),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetValidIngredientStatesNeedingIndexing",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
WHERE %s.%s IS NULL
	AND (
	%s.%s IS NULL
	OR %s.%s < %s - '24 hours'::INTERVAL
);`,
					validIngredientStatesTableName,
					idColumn,
					validIngredientStatesTableName,
					validIngredientStatesTableName,
					archivedAtColumn,
					validIngredientStatesTableName,
					lastIndexedAtColumn,
					validIngredientStatesTableName,
					lastIndexedAtColumn,
					currentTimeExpression,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetValidIngredientState",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
AND %s.%s = sqlc.arg(%s);`,
					strings.Join(applyToEach(validIngredientStatesColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validIngredientStatesTableName, s)
					}), ",\n\t"),
					validIngredientStatesTableName,
					validIngredientStatesTableName,
					archivedAtColumn,
					validIngredientStatesTableName,
					idColumn,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetValidIngredientStatesWithIDs",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[]);`,
					strings.Join(applyToEach(validIngredientStatesColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validIngredientStatesTableName, s)
					}), ",\n\t"),
					validIngredientStatesTableName,
					validIngredientStatesTableName,
					archivedAtColumn,
					validIngredientStatesTableName,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "SearchForValidIngredientStates",
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
					strings.Join(applyToEach(validIngredientStatesColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validIngredientStatesTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(validIngredientStatesTableName, true, true, []string{}),
					buildTotalCountSelect(validIngredientStatesTableName, true, []string{}),
					validIngredientStatesTableName,
					validIngredientStatesTableName,
					archivedAtColumn,
					validIngredientStatesTableName,
					nameColumn,
					buildILIKEForArgument("name_query"),
					buildFilterConditions(
						validIngredientStatesTableName,
						true,
						true,
					),
					buildCursorLimitClause(validIngredientStatesTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateValidIngredientState",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
					validIngredientStatesTableName,
					strings.Join(applyToEach(filterForUpdate(validIngredientStatesColumns), func(i int, s string) string {
						return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
					}), ",\n\t"),
					lastUpdatedAtColumn,
					currentTimeExpression,
					archivedAtColumn,
					idColumn,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateValidIngredientStateLastIndexedAt",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s = sqlc.arg(%s) AND %s IS NULL;`,
					validIngredientStatesTableName,
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
