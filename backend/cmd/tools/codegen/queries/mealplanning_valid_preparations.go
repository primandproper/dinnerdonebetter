package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

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

		return slices.Concat(
			pgGen.StandardCRUD(validPreparationsTableName, validPreparationsColumns,
				querygen.WithEntity("ValidPreparation", "ValidPreparations"),
				querygen.WithOmitted(querygen.ListQuery),
			),
			[]*Query{
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
WHERE %s
GROUP BY %s.%s
%s;`,
						strings.Join(applyToEach(validPreparationsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validPreparationsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(validPreparationsTableName, validPreparationsColumns, []string{}),
						pgGen.TotalCountSelect(validPreparationsTableName, validPreparationsColumns, []string{}),
						validPreparationsTableName,
						pgGen.FilterConditions(validPreparationsTableName, validPreparationsColumns),
						validPreparationsTableName,
						idColumn,
						pgGen.CursorLimitClause(validPreparationsTableName),
					)),
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
						validPreparationsTableName, lastIndexedAtColumn, querygen.NowExpression,
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
WHERE %s
%s;`,
						strings.Join(applyToEach(validPreparationsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validPreparationsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(validPreparationsTableName, validPreparationsColumns, []string{}),
						pgGen.TotalCountSelect(validPreparationsTableName, validPreparationsColumns, []string{}),
						validPreparationsTableName,
						pgGen.FilterConditions(validPreparationsTableName, validPreparationsColumns,
							fmt.Sprintf("%s.%s %s", validPreparationsTableName, nameColumn, buildILIKEForArgument("name_query")),
						),
						pgGen.CursorLimitClause(validPreparationsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
