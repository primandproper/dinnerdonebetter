package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	commentsTableName = "comments"
)

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
			pgGen.StandardCRUD(commentsTableName, commentsColumns,
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
						querygen.NowExpression,
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
WHERE %s
%s;`,
						strings.Join(applyToEach(commentsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", commentsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(commentsTableName, commentsColumns, []string{},
							fmt.Sprintf("%s.%s = sqlc.arg(target_type)", commentsTableName, "target_type"),
							fmt.Sprintf("%s.%s = sqlc.arg(referenced_id)", commentsTableName, "referenced_id")),
						pgGen.TotalCountSelect(commentsTableName, commentsColumns, []string{},
							fmt.Sprintf("%s.%s = sqlc.arg(target_type)", commentsTableName, "target_type"),
							fmt.Sprintf("%s.%s = sqlc.arg(referenced_id)", commentsTableName, "referenced_id")),
						commentsTableName,
						pgGen.FilterConditions(commentsTableName, commentsColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(target_type)", commentsTableName, "target_type"),
							fmt.Sprintf("%s.%s = sqlc.arg(referenced_id)", commentsTableName, "referenced_id"),
						),
						pgGen.CursorLimitClause(commentsTableName),
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
WHERE %s
%s;`,
						strings.Join(applyToEach(commentsColumns, func(i int, s string) string {
							return fmt.Sprintf("%s.%s", commentsTableName, s)
						}), ",\n\t"),
						pgGen.FilterCountSelect(commentsTableName, commentsColumns, []string{},
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", commentsTableName, belongsToUserColumn, belongsToUserColumn)),
						pgGen.TotalCountSelect(commentsTableName, commentsColumns, []string{},
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", commentsTableName, belongsToUserColumn, belongsToUserColumn)),
						commentsTableName,
						pgGen.FilterConditions(commentsTableName, commentsColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", commentsTableName, belongsToUserColumn, belongsToUserColumn),
						),
						pgGen.CursorLimitClause(commentsTableName),
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
						strings.Join(applyToEach(querygen.ForUpdate(commentsColumns, "target_type", "referenced_id", belongsToUserColumn, "parent_comment_id"), func(i int, s string) string {
							return fmt.Sprintf("%s = sqlc.arg(%s)", s, s)
						}), ",\n\t"),
						lastUpdatedAtColumn, querygen.NowExpression,
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
