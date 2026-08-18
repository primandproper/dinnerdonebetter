package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	validIngredientGroupsTableName = "valid_ingredient_groups"
)

func init() {
	registerTableName(validIngredientGroupsTableName)
}

var validIngredientGroupsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	slugColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildValidIngredientGroupsQueries(database string) []*Query {
	switch database {
	case postgres:

		groupMemberInsertColumns := filterForInsert(validIngredientGroupMembersColumns)

		fullMemberSelectColumns := mergeColumns(
			applyToEach(filterFromSlice(validIngredientGroupMembersColumns, "valid_ingredient"), func(i int, s string) string {
				return fmt.Sprintf("%s.%s", validIngredientGroupMembersTableName, s)
			}),
			applyToEach(validIngredientsColumns, func(i int, s string) string {
				return fmt.Sprintf("%s.%s as valid_ingredient_%s", validIngredientsTableName, s, s)
			}),
			2,
		)

		return slices.Concat(
			querygen.StandardCRUD(validIngredientGroupsTableName, validIngredientGroupsColumns,
				querygen.WithEntity("ValidIngredientGroup", "ValidIngredientGroups"),
				querygen.WithOmitted(querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveValidIngredientGroupMember",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						validIngredientGroupMembersTableName,
						archivedAtColumn, currentTimeExpression,
						archivedAtColumn,
						idColumn, idColumn,
						belongsToGroupColumn, belongsToGroupColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "CreateValidIngredientGroupMember",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
						validIngredientGroupMembersTableName,
						strings.Join(groupMemberInsertColumns, ",\n\t"),
						strings.Join(applyToEach(groupMemberInsertColumns, func(i int, s string) string {
							return fmt.Sprintf("sqlc.arg(%s)", s)
						}), ",\n\t"),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientGroups",
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
						strings.Join(applyToEach(validIngredientGroupsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validIngredientGroupsTableName, s)
						}), ",\n\t"),
						buildFilterCountSelect(validIngredientGroupsTableName, true, true, []string{}),
						buildTotalCountSelect(validIngredientGroupsTableName, true, []string{}),
						validIngredientGroupsTableName,
						validIngredientGroupsTableName,
						archivedAtColumn,
						buildFilterConditions(
							validIngredientGroupsTableName,
							true,
							true,
						),
						validIngredientGroupsTableName, idColumn,
						buildCursorLimitClause(validIngredientGroupsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientGroupMembers",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	JOIN %s ON %s.%s = %s.%s
	JOIN %s ON %s.%s = %s.%s
WHERE 
	%s.%s IS NULL
	AND %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(fullMemberSelectColumns, ",\n\t"),
						validIngredientGroupMembersTableName,
						validIngredientGroupsTableName, validIngredientGroupsTableName, idColumn, validIngredientGroupMembersTableName, belongsToGroupColumn,
						validIngredientsTableName, validIngredientsTableName, idColumn, validIngredientGroupMembersTableName, validIngredientGroupMemberValidIngredientColumn,
						validIngredientGroupsTableName, archivedAtColumn,
						validIngredientGroupMembersTableName, archivedAtColumn,
						validIngredientGroupMembersTableName, belongsToGroupColumn, belongsToGroupColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "SearchForValidIngredientGroups",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE
	%s.%s IS NULL
	AND %s.%s %s
	%s
GROUP BY %s.%s
%s;`,
						strings.Join(applyToEach(validIngredientGroupsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validIngredientGroupsTableName, s)
						}), ",\n\t"),
						buildFilterCountSelect(validIngredientGroupsTableName, true, true, []string{}, fmt.Sprintf("%s.%s %s", validIngredientGroupsTableName, nameColumn, buildILIKEForArgument("name"))),
						buildTotalCountSelect(validIngredientGroupsTableName, true, []string{}),
						validIngredientGroupsTableName,
						validIngredientGroupsTableName,
						archivedAtColumn,
						validIngredientGroupsTableName, nameColumn, buildILIKEForArgument("name"),
						buildFilterConditions(
							validIngredientGroupsTableName,
							true,
							true,
						),
						validIngredientGroupsTableName, idColumn,
						buildCursorLimitClause(validIngredientGroupsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetValidIngredientGroupsWithIDs",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = ANY(sqlc.arg(ids)::text[]);`,
						strings.Join(applyToEach(validIngredientGroupsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validIngredientGroupsTableName, s)
						}), ",\n\t"),
						validIngredientGroupsTableName,
						validIngredientGroupsTableName,
						archivedAtColumn,
						validIngredientGroupsTableName,
						idColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
