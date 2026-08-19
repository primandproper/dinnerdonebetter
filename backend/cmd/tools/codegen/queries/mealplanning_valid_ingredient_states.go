package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

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

		return slices.Concat(
			querygen.StandardCRUD(validIngredientStatesTableName, validIngredientStatesColumns,
				querygen.WithEntity("ValidIngredientState", "ValidIngredientStates"),
				querygen.WithOmitted(querygen.ListQuery),
			),
			[]*Query{
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
WHERE %s
GROUP BY %s.%s
%s;`,
						strings.Join(applyToEach(validIngredientStatesColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validIngredientStatesTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(validIngredientStatesTableName, validIngredientStatesColumns, []string{}),
						querygen.TotalCountSelect(validIngredientStatesTableName, validIngredientStatesColumns, []string{}),
						validIngredientStatesTableName,
						querygen.FilterConditions(validIngredientStatesTableName, validIngredientStatesColumns),
						validIngredientStatesTableName,
						idColumn,
						querygen.CursorLimitClause(validIngredientStatesTableName),
					)),
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
						querygen.NowExpression,
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
WHERE %s
%s;`,
						strings.Join(applyToEach(validIngredientStatesColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validIngredientStatesTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(validIngredientStatesTableName, validIngredientStatesColumns, []string{}),
						querygen.TotalCountSelect(validIngredientStatesTableName, validIngredientStatesColumns, []string{}),
						validIngredientStatesTableName,
						querygen.FilterConditions(validIngredientStatesTableName, validIngredientStatesColumns,
							fmt.Sprintf("%s.%s %s", validIngredientStatesTableName, nameColumn, buildILIKEForArgument("name_query")),
						),
						querygen.CursorLimitClause(validIngredientStatesTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
