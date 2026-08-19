package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	validPrepTaskConfigsTableName = "valid_prep_task_configs"
)

func init() {
	registerTableName(validPrepTaskConfigsTableName)
}

var validPrepTaskConfigsColumns = []string{
	idColumn,
	validIngredientIDColumn,
	validPreparationIDColumn,
	"minimum_storage_duration_in_seconds",
	"maximum_storage_duration_in_seconds",
	"storage_container_type",
	"minimum_storage_temperature_in_celsius",
	"maximum_storage_temperature_in_celsius",
	"storage_instructions",
	notesColumn,
	"source",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidPrepTaskConfigsQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			mergeColumns(
				applyToEach(filterFromSlice(validPrepTaskConfigsColumns, "valid_preparation_id", "valid_ingredient_id"), func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_prep_task_config_%s", validPrepTaskConfigsTableName, s, s)
				}),
				applyToEach(validIngredientsColumns, func(i int, s string) string {
					return fmt.Sprintf("%s.%s as valid_ingredient_%s", validIngredientsTableName, s, s)
				}),
				2,
			),
			applyToEach(validPreparationsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_preparation_%s", validPreparationsTableName, s, s)
			}),
			2,
		)

		return slices.Concat(
			querygen.StandardCRUD(validPrepTaskConfigsTableName, validPrepTaskConfigsColumns,
				querygen.WithEntity("ValidPrepTaskConfig", "ValidPrepTaskConfigs"),
				querygen.WithNullable("maximum_storage_duration_in_seconds", "maximum_storage_temperature_in_celsius", "minimum_storage_temperature_in_celsius"),
				querygen.WithOmitted(querygen.GetQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPrepTaskConfigsForIngredient",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(validPrepTaskConfigsTableName, true, true, []string{}),
						buildTotalCountSelect(validPrepTaskConfigsTableName, true, []string{}),
						validPrepTaskConfigsTableName,
						validIngredientsTableName,
						validPrepTaskConfigsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validPrepTaskConfigsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validPrepTaskConfigsTableName,
						archivedAtColumn,
						validPrepTaskConfigsTableName,
						validIngredientIDColumn,
						idColumn,
						buildFilterConditions(validPrepTaskConfigsTableName, true, false),
						buildCursorLimitClause(validPrepTaskConfigsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPrepTaskConfigsForPreparation",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(validPrepTaskConfigsTableName, true, true, []string{}),
						buildTotalCountSelect(validPrepTaskConfigsTableName, true, []string{}),
						validPrepTaskConfigsTableName,
						validIngredientsTableName,
						validPrepTaskConfigsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validPrepTaskConfigsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validPrepTaskConfigsTableName,
						archivedAtColumn,
						validPrepTaskConfigsTableName,
						validPreparationIDColumn,
						idColumn,
						buildFilterConditions(validPrepTaskConfigsTableName, true, false),
						buildCursorLimitClause(validPrepTaskConfigsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPrepTaskConfigsForIngredientAndPreparation",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(validPrepTaskConfigsTableName, true, true, []string{}),
						buildTotalCountSelect(validPrepTaskConfigsTableName, true, []string{}),
						validPrepTaskConfigsTableName,
						validIngredientsTableName,
						validPrepTaskConfigsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validPrepTaskConfigsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validPrepTaskConfigsTableName,
						archivedAtColumn,
						validPrepTaskConfigsTableName,
						validIngredientIDColumn,
						validIngredientIDColumn,
						validPrepTaskConfigsTableName,
						validPreparationIDColumn,
						validPreparationIDColumn,
						buildFilterConditions(validPrepTaskConfigsTableName, true, false),
						buildCursorLimitClause(validPrepTaskConfigsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPrepTaskConfigs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(validPrepTaskConfigsTableName, true, true, []string{}),
						buildTotalCountSelect(validPrepTaskConfigsTableName, true, []string{}),
						validPrepTaskConfigsTableName,
						validIngredientsTableName,
						validPrepTaskConfigsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validPrepTaskConfigsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validPrepTaskConfigsTableName,
						archivedAtColumn,
						buildFilterConditions(validPrepTaskConfigsTableName, true, false),
						buildCursorLimitClause(validPrepTaskConfigsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidPrepTaskConfig",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE
	%s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullSelectColumns, ",\n\t"),
						validPrepTaskConfigsTableName,
						validIngredientsTableName,
						validPrepTaskConfigsTableName,
						validIngredientIDColumn,
						validIngredientsTableName,
						idColumn,
						validPreparationsTableName,
						validPrepTaskConfigsTableName,
						validPreparationIDColumn,
						validPreparationsTableName,
						idColumn,
						validPrepTaskConfigsTableName,
						archivedAtColumn,
						validIngredientsTableName,
						archivedAtColumn,
						validPreparationsTableName,
						archivedAtColumn,
						validPrepTaskConfigsTableName,
						idColumn,
						idColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
