package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	validIngredientsTableName = "valid_ingredients"
	validIngredientIDColumn   = "valid_ingredient_id"
)

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

		return slices.Concat(
			pgGen.StandardCRUD(validIngredientsTableName, validIngredientsColumns,
				querygen.WithEntity("ValidIngredient", "ValidIngredients"),
				querygen.WithNullable(
					"minimum_ideal_storage_temperature_in_celsius",
					"maximum_ideal_storage_temperature_in_celsius",
				),
				querygen.WithOmitted(querygen.ListQuery),
			),
			[]*Query{
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
WHERE %s
GROUP BY %s.%s
%s;`,
						strings.Join(applyToEach(validIngredientsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validIngredientsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(validIngredientsTableName, validIngredientsColumns, []string{}),
						pgGen.TotalCountSelect(validIngredientsTableName, validIngredientsColumns, []string{}),
						validIngredientsTableName,
						pgGen.FilterConditions(validIngredientsTableName, validIngredientsColumns),
						validIngredientsTableName,
						idColumn,
						pgGen.CursorLimitClause(validIngredientsTableName),
					)),
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
						querygen.NowExpression,
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
WHERE %s
%s;`,
						strings.Join(applyToEach(validIngredientsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validIngredientsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(validIngredientsTableName, validIngredientsColumns, []string{}),
						pgGen.TotalCountSelect(validIngredientsTableName, validIngredientsColumns, []string{}),
						validIngredientsTableName,
						pgGen.FilterConditions(validIngredientsTableName, validIngredientsColumns,
							fmt.Sprintf("%s.%s %s", validIngredientsTableName, nameColumn, buildILIKEForArgument("name_query")),
						),
						pgGen.CursorLimitClause(validIngredientsTableName),
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
			},
		)
	default:
		return nil
	}
}
