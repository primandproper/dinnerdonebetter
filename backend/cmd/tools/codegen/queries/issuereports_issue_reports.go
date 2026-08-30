package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v13/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	issueReportsTableName  = "issue_reports"
	issueTypeColumn        = "issue_type"
	detailsColumn          = "details"
	relevantTableColumn    = "relevant_table"
	relevantRecordIDColumn = "relevant_record_id"
)

var issueReportsColumns = []string{
	idColumn,
	issueTypeColumn,
	detailsColumn,
	relevantTableColumn,
	relevantRecordIDColumn,
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
	createdByUserColumn,
	belongsToAccountColumn,
}

func buildIssueReportsQueries(database string) []*Query {
	switch database {
	case postgres:
		fullSelectColumns := applyToEach(issueReportsColumns, func(_ int, s string) string {
			return querygen.Qualify(issueReportsTableName, s)
		})

		return slices.Concat(
			pgGen.StandardCRUD(issueReportsTableName, issueReportsColumns,
				querygen.WithEntity("IssueReport", "IssueReports"),
				querygen.WithImmutable(createdByUserColumn, belongsToAccountColumn),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveIssueReport",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						issueReportsTableName,
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn, querygen.NowExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "CheckIssueReportExistence",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS(
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
						issueReportsTableName, idColumn,
						issueReportsTableName,
						issueReportsTableName, archivedAtColumn,
						issueReportsTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetIssueReports",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(issueReportsTableName, issueReportsColumns, nil),
						pgGen.TotalCountSelect(issueReportsTableName, issueReportsColumns, nil),
						issueReportsTableName,
						pgGen.FilterConditions(issueReportsTableName, issueReportsColumns, querygen.Ascending),
						pgGen.CursorLimitClause(issueReportsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetIssueReportsForAccount",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(issueReportsTableName, issueReportsColumns, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, belongsToAccountColumn, belongsToAccountColumn)),
						pgGen.TotalCountSelect(issueReportsTableName, issueReportsColumns, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, belongsToAccountColumn, belongsToAccountColumn)),
						issueReportsTableName,
						pgGen.FilterConditions(issueReportsTableName, issueReportsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, belongsToAccountColumn, belongsToAccountColumn),
						),
						pgGen.CursorLimitClause(issueReportsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetIssueReportsForTable",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(issueReportsTableName, issueReportsColumns, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantTableColumn, relevantTableColumn)),
						pgGen.TotalCountSelect(issueReportsTableName, issueReportsColumns, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantTableColumn, relevantTableColumn)),
						issueReportsTableName,
						pgGen.FilterConditions(issueReportsTableName, issueReportsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantTableColumn, relevantTableColumn),
						),
						pgGen.CursorLimitClause(issueReportsTableName, querygen.Ascending),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetIssueReportsForRecord",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(issueReportsTableName, issueReportsColumns, nil,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantTableColumn, relevantTableColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantRecordIDColumn, relevantRecordIDColumn)),
						pgGen.TotalCountSelect(issueReportsTableName, issueReportsColumns, nil,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantTableColumn, relevantTableColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantRecordIDColumn, relevantRecordIDColumn)),
						issueReportsTableName,
						pgGen.FilterConditions(issueReportsTableName, issueReportsColumns, querygen.Ascending,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantTableColumn, relevantTableColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantRecordIDColumn, relevantRecordIDColumn),
						),
						pgGen.CursorLimitClause(issueReportsTableName, querygen.Ascending),
					)),
				},
			},
		)
	default:
		return nil
	}
}
