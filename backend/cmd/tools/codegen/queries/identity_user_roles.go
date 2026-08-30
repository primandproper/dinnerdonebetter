package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	userRolesTableName = "user_roles"

	scopeColumn = "scope"
)

var userRolesColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	scopeColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildUserRolesQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			pgGen.StandardCRUD(userRolesTableName, userRolesColumns,
				querygen.WithEntity("UserRole", "UserRoles"),
				querygen.WithQueryName(querygen.GetQuery, "GetUserRoleByID"),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetUserRoleByName",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(applyToEach(userRolesColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", userRolesTableName, s)
						}), ",\n\t"),
						userRolesTableName,
						userRolesTableName, archivedAtColumn,
						userRolesTableName, nameColumn, nameColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetUserRoles",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(applyToEach(userRolesColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", userRolesTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(userRolesTableName, userRolesColumns, []string{}),
						pgGen.TotalCountSelect(userRolesTableName, userRolesColumns, []string{}),
						userRolesTableName,
						pgGen.FilterConditions(userRolesTableName, userRolesColumns),
						pgGen.CursorLimitClause(userRolesTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
