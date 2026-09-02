package main

import (
	"fmt"
	"slices"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	userRoleAssignmentsTableName = "user_role_assignments"

	userIDColumn    = "user_id"
	accountIDColumn = "account_id"
	roleNameColumn  = "role_name"
)

var userRoleAssignmentsColumns = []string{
	idColumn,
	userIDColumn,
	roleNameColumn,
	accountIDColumn,
	createdAtColumn,
	archivedAtColumn,
}

// buildUserRoleAssignmentsQueries renders the statements over which roles a principal
// holds.
//
// What those roles grant is not here and has no statements in this repository at all:
// the policy tables belong to platform-go's authorization/database, whose own recursive
// statement resolves them. That is why the two recursive CTEs this file used to render
// are gone — they walked the role hierarchy and joined the grants, which is exactly the
// work that package does, and which sqlc can no longer see the tables for.
//
// An assignment names a role by name rather than by id for the same reason. There is no
// join left to make.
func buildUserRoleAssignmentsQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			pgGen.StandardCRUD(userRoleAssignmentsTableName, userRoleAssignmentsColumns,
				querygen.WithEntity("UserRoleAssignments", "UserRoleAssignmentss"),
				querygen.WithQueryName(querygen.CreateQuery, "AssignRoleToUser"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveRoleAssignmentsForUserAndAccount",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						userRoleAssignmentsTableName,
						archivedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						userIDColumn, userIDColumn,
						accountIDColumn, accountIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateAccountRoleAssignment",
						Type: ExecType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = sqlc.arg(new_role_name)
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						userRoleAssignmentsTableName,
						roleNameColumn,
						archivedAtColumn,
						userIDColumn, userIDColumn,
						accountIDColumn, accountIDColumn,
					)),
				},
				{
					// One statement for both scopes. account_id IS NULL is a
					// service-wide assignment and anything else is scoped to that
					// account, so splitting this in two would be two reads of one
					// table differing only in a predicate the caller has to apply
					// anyway to group the result.
					Annotation: QueryAnnotation{
						Name: "GetRoleAssignmentsForUser",
						Type: ManyType,
					},
					Content: fmt.Sprintf(`SELECT %s.%s, %s.%s
FROM %s
WHERE %s.%s = sqlc.arg(%s)
	AND %s.%s IS NULL`,
						userRoleAssignmentsTableName, accountIDColumn,
						userRoleAssignmentsTableName, roleNameColumn,
						userRoleAssignmentsTableName,
						userRoleAssignmentsTableName, userIDColumn, userIDColumn,
						userRoleAssignmentsTableName, archivedAtColumn,
					),
				},
			},
		)
	default:
		return nil
	}
}
