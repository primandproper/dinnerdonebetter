package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	userRoleHierarchyTableName = "user_role_hierarchy"

	parentRoleIDColumn = "parent_role_id"
	childRoleIDColumn  = "child_role_id"
)

var userRoleHierarchyColumns = []string{
	idColumn,
	parentRoleIDColumn,
	childRoleIDColumn,
	createdAtColumn,
	archivedAtColumn,
}

func buildUserRoleHierarchyQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			pgGen.StandardCRUD(userRoleHierarchyTableName, userRoleHierarchyColumns,
				querygen.WithEntity("UserRoleHierarchy", "UserRoleHierarchys"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetUserRoleHierarchy",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL;`,
						strings.Join(applyToEach(userRoleHierarchyColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", userRoleHierarchyTableName, s)
						}), ",\n\t"),
						userRoleHierarchyTableName,
						userRoleHierarchyTableName, archivedAtColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveUserRoleHierarchy",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						userRoleHierarchyTableName,
						archivedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						parentRoleIDColumn, parentRoleIDColumn,
						childRoleIDColumn, childRoleIDColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
