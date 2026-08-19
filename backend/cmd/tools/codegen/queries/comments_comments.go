package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	commentsTableName = "comments"
)

func init() {
	registerTableName(commentsTableName)
}

var commentsColumns = []string{
	idColumn,
	"content",
	"target_type",
	"referenced_id",
	"parent_comment_id",
	belongsToUserColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildCommentsQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			querygen.StandardCRUD(commentsTableName, commentsColumns,
				querygen.WithEntity("Comment", "Comments"),
				querygen.WithNullable("parent_comment_id"),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveCommentsForReference",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET %s = %s WHERE %s IS NULL
	AND %s = sqlc.arg(target_type)
	AND %s = sqlc.arg(referenced_id);`,
						commentsTableName,
						archivedAtColumn,
						currentTimeExpression,
						archivedAtColumn,
						"target_type",
						"referenced_id",
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetCommentsForReference",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(target_type)
	AND %s.%s = sqlc.arg(referenced_id)
	%s
%s;`,
						strings.Join(applyToEach(commentsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", commentsTableName, s)
						}), ",\n\t"),
						buildFilterCountSelect(commentsTableName, true, true, []string{},
							fmt.Sprintf("%s.%s = sqlc.arg(target_type)", commentsTableName, "target_type"),
							fmt.Sprintf("%s.%s = sqlc.arg(referenced_id)", commentsTableName, "referenced_id")),
						buildTotalCountSelect(commentsTableName, true, []string{},
							fmt.Sprintf("%s.%s = sqlc.arg(target_type)", commentsTableName, "target_type"),
							fmt.Sprintf("%s.%s = sqlc.arg(referenced_id)", commentsTableName, "referenced_id")),
						commentsTableName,
						commentsTableName, archivedAtColumn,
						commentsTableName, "target_type",
						commentsTableName, "referenced_id",
						buildFilterConditions(commentsTableName, true, true,
							fmt.Sprintf("%s.%s = sqlc.arg(target_type)", commentsTableName, "target_type"),
							fmt.Sprintf("%s.%s = sqlc.arg(referenced_id)", commentsTableName, "referenced_id")),
						buildCursorLimitClause(commentsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetCommentsForUser",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
						strings.Join(applyToEach(commentsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", commentsTableName, s)
						}), ",\n\t"),
						buildFilterCountSelect(commentsTableName, true, true, []string{},
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", commentsTableName, belongsToUserColumn, belongsToUserColumn)),
						buildTotalCountSelect(commentsTableName, true, []string{},
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", commentsTableName, belongsToUserColumn, belongsToUserColumn)),
						commentsTableName,
						commentsTableName, archivedAtColumn,
						commentsTableName, belongsToUserColumn, belongsToUserColumn,
						buildFilterConditions(commentsTableName, true, true,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", commentsTableName, belongsToUserColumn, belongsToUserColumn)),
						buildCursorLimitClause(commentsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "UpdateComment",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
						commentsTableName,
						strings.Join(applyToEach(filterForUpdate(commentsColumns, "target_type", "referenced_id", belongsToUserColumn, "parent_comment_id"), func(i int, s string) string {
							return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
						}), ",\n\t"),
						lastUpdatedAtColumn, currentTimeExpression,
						archivedAtColumn,
						idColumn, idColumn,
						belongsToUserColumn, belongsToUserColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
