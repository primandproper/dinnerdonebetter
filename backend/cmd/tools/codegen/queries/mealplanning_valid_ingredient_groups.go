package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	validIngredientGroupsTableName = "valid_ingredient_groups"
)

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
			pgGen.StandardCRUD(validIngredientGroupsTableName, validIngredientGroupsColumns,
				querygen.WithEntity("ValidIngredientGroup", "ValidIngredientGroups"),
				querygen.WithOmitted(querygen.ListQuery),
			),
			pgGen.StandardCRUD(validIngredientGroupMembersTableName, validIngredientGroupMembersColumns,
				querygen.WithEntity("ValidIngredientGroupMember", "ValidIngredientGroupMembers"),
				querygen.WithOwnership(belongsToGroupColumn),
				querygen.WithOmitted(querygen.GetQuery, querygen.ExistsQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
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
WHERE %s
GROUP BY %s.%s
%s;`,
						strings.Join(applyToEach(validIngredientGroupsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validIngredientGroupsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(validIngredientGroupsTableName, validIngredientGroupsColumns, []string{}),
						pgGen.TotalCountSelect(validIngredientGroupsTableName, validIngredientGroupsColumns, []string{}),
						validIngredientGroupsTableName,
						pgGen.FilterConditions(validIngredientGroupsTableName, validIngredientGroupsColumns, querygen.Ascending),
						validIngredientGroupsTableName,
						idColumn,
						pgGen.CursorLimitClause(validIngredientGroupsTableName, querygen.Ascending),
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
WHERE %s
GROUP BY %s.%s
%s;`,
						strings.Join(applyToEach(validIngredientGroupsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", validIngredientGroupsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(validIngredientGroupsTableName, validIngredientGroupsColumns, []string{}, fmt.Sprintf("%s.%s %s", validIngredientGroupsTableName, nameColumn, buildILIKEForArgument("name"))),
						pgGen.TotalCountSelect(validIngredientGroupsTableName, validIngredientGroupsColumns, []string{}),
						validIngredientGroupsTableName,
						pgGen.FilterConditions(validIngredientGroupsTableName, validIngredientGroupsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s %s", validIngredientGroupsTableName, nameColumn, buildILIKEForArgument("name")),
						),
						validIngredientGroupsTableName,
						idColumn,
						pgGen.CursorLimitClause(validIngredientGroupsTableName, querygen.Ascending),
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
