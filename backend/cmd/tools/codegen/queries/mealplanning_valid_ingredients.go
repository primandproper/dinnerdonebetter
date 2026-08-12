package main

import (
	"fmt"
	"strings"

	"github.com/cristalhq/builq"
)

const (
	validIngredientsTableName = "valid_ingredients"
	validIngredientIDColumn   = "valid_ingredient_id"
)

func init() {
	registerTableName(validIngredientsTableName)
}

var validIngredientsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	"warning",
	"contains_egg",
	"contains_dairy",
	"contains_peanut",
	"contains_tree_nut",
	"contains_soy",
	"contains_wheat",
	"contains_shellfish",
	"contains_sesame",
	"contains_fish",
	"contains_gluten",
	"animal_flesh",
	"is_liquid",
	iconPathColumn,
	"animal_derived",
	pluralNameColumn,
	"restrict_to_preparations",
	"contaminates_equipment",
	"minimum_ideal_storage_temperature_in_celsius",
	"maximum_ideal_storage_temperature_in_celsius",
	"storage_instructions",
	slugColumn,
	"contains_alcohol",
	"shopping_suggestions",
	"is_starch",
	"is_protein",
	"is_grain",
	"is_fruit",
	"is_salt",
	"is_fat",
	"is_acid",
	"is_heat",
	lastIndexedAtColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidIngredientsQueries(database string) []*Query {
	switch database {
	case postgres:

		insertColumns := filterForInsert(validIngredientsColumns)

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveValidIngredient",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s IS NULL AND %s = sqlc.arg(%s);`,
					validIngredientsTableName,
					archivedAtColumn,
					currentTimeExpression,
					archivedAtColumn,
					idColumn,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CreateValidIngredient",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					validIngredientsTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(i int, s string) string {
						switch s {
						case "minimum_ideal_storage_temperature_in_celsius",
							"maximum_ideal_storage_temperature_in_celsius":
							return fmt.Sprintf("sqlc.narg(%s)", s)
						default:
							return fmt.Sprintf("sqlc.arg(%s)", s)
						}
					}), ",\n\t"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CheckValidIngredientExistence",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS (
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
					validIngredientsTableName, idColumn,
					validIngredientsTableName,
					validIngredientsTableName,
					archivedAtColumn,
					validIngredientsTableName,
					idColumn,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetValidIngredients",
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
					strings.Join(applyToEach(validIngredientsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validIngredientsTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(validIngredientsTableName, true, true, []string{}),
					buildTotalCountSelect(validIngredientsTableName, true, []string{}),
					validIngredientsTableName,
					validIngredientsTableName,
					archivedAtColumn,
					buildFilterConditions(
						validIngredientsTableName,
						true,
						true,
					),
					validIngredientsTableName, idColumn,
					buildCursorLimitClause(validIngredientsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "ScanValidIngredientIDsForReindex",
					Type: ManyType,
				},
				Content: buildReindexScanQuery(validIngredientsTableName),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetValidIngredientsNeedingIndexing",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT %s.%s
FROM %s
WHERE %s.%s IS NULL
	AND (
	%s.%s IS NULL
	OR %s.%s < %s - '24 hours'::INTERVAL
);`,
					validIngredientsTableName,
					idColumn,
					validIngredientsTableName,
					validIngredientsTableName,
					archivedAtColumn,
					validIngredientsTableName,
					lastIndexedAtColumn,
					validIngredientsTableName,
					lastIndexedAtColumn,
					currentTimeExpression,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetValidIngredient",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
AND %s.%s = sqlc.arg(%s);`,
					strings.Join(applyToEach(validIngredientsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validIngredientsTableName, s)
					}), ",\n\t"),
					validIngredientsTableName,
					validIngredientsTableName,
					archivedAtColumn,
					validIngredientsTableName,
					idColumn,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetRandomValidIngredient",
					Type: OneType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
ORDER BY RANDOM() LIMIT 1;`,
					strings.Join(applyToEach(validIngredientsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validIngredientsTableName, s)
					}), ",\n\t"),
					validIngredientsTableName,
					validIngredientsTableName,
					archivedAtColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "GetValidIngredientsWithIDs",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[]);`,
					strings.Join(applyToEach(validIngredientsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validIngredientsTableName, s)
					}), ",\n\t"),
					validIngredientsTableName,
					validIngredientsTableName,
					archivedAtColumn,
					validIngredientsTableName,
					idColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "SearchForValidIngredients",
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
					strings.Join(applyToEach(validIngredientsColumns, func(i int, s string) string {
						return fmt.Sprintf("%s.%s", validIngredientsTableName, s)
					}), ",\n\t"),
					buildFilterCountSelect(validIngredientsTableName, true, true, []string{}),
					buildTotalCountSelect(validIngredientsTableName, true, []string{}),
					validIngredientsTableName,
					validIngredientsTableName,
					archivedAtColumn,
					validIngredientsTableName,
					nameColumn,
					buildILIKEForArgument("name_query"),
					buildFilterConditions(
						validIngredientsTableName,
						true,
						true,
					),
					buildCursorLimitClause(validIngredientsTableName),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "SearchValidIngredientsByPreparationAndIngredientName",
					Type: ManyType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s %s;`,
					strings.Join(applyToEach(validIngredientsColumns, func(i int, s string) string {
						if i == 0 {
							return fmt.Sprintf("DISTINCT(%s.%s)", validIngredientsTableName, s)
						}
						return fmt.Sprintf("%s.%s", validIngredientsTableName, s)
					}), ",\n\t"),
					validIngredientPreparationsTableName,
					validIngredientsTableName, validIngredientPreparationsTableName, validIngredientIDColumn, validIngredientsTableName, idColumn,
					validPreparationsTableName, validIngredientPreparationsTableName, validPreparationIDColumn, validPreparationsTableName, idColumn,
					validIngredientPreparationsTableName, archivedAtColumn,
					validIngredientsTableName, archivedAtColumn,
					validPreparationsTableName, archivedAtColumn,
					validIngredientPreparationsTableName, validPreparationIDColumn, validPreparationIDColumn,
					validIngredientsTableName, nameColumn, buildILIKEForArgument("name_query"),
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "UpdateValidIngredient",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
					validIngredientsTableName,
					strings.Join(applyToEach(filterForUpdate(validIngredientsColumns), func(i int, s string) string {
						switch s {
						case "minimum_ideal_storage_temperature_in_celsius",
							"maximum_ideal_storage_temperature_in_celsius":
							return fmt.Sprintf("%s = sqlc.narg(%s)", s, s)
						default:
							return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
						}
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
					Name: "UpdateValidIngredientLastIndexedAt",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s = sqlc.arg(%s) AND %s IS NULL;`,
					validIngredientsTableName,
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
