package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	permissionsTableName = "permissions"
)

func init() {
	registerTableName(permissionsTableName)
}

var permissionsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildPermissionsQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			querygen.StandardCRUD(permissionsTableName, permissionsColumns,
				querygen.WithEntity("Permission", "Permissions"),
				querygen.WithQueryName(querygen.GetQuery, "GetPermissionByID"),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetPermissionByName",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(applyToEach(permissionsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", permissionsTableName, s)
						}), ",\n\t"),
						permissionsTableName,
						permissionsTableName, archivedAtColumn,
						permissionsTableName, nameColumn, nameColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetPermissions",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	%s
%s;`,
						strings.Join(applyToEach(permissionsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", permissionsTableName, s)
						}), ",\n\t"),
						buildFilterCountSelect(permissionsTableName, true, true, []string{}),
						buildTotalCountSelect(permissionsTableName, true, []string{}),
						permissionsTableName,
						permissionsTableName, archivedAtColumn,
						buildFilterConditions(permissionsTableName, true, true),
						buildCursorLimitClause(permissionsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
