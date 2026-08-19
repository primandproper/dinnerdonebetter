package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	issueReportsTableName  = "issue_reports"
	issueTypeColumn        = "issue_type"
	detailsColumn          = "details"
	relevantTableColumn    = "relevant_table"
	relevantRecordIDColumn = "relevant_record_id"
)

func init() {
	registerTableName(issueReportsTableName)
}

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
			return fullColumnName(issueReportsTableName, s)
		})

		return slices.Concat(
			querygen.StandardCRUD(issueReportsTableName, issueReportsColumns,
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
						lastUpdatedAtColumn, currentTimeExpression,
						archivedAtColumn, currentTimeExpression,
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
WHERE %s.%s IS NULL
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(issueReportsTableName, true, true, nil),
						buildTotalCountSelect(issueReportsTableName, true, nil),
						issueReportsTableName,
						issueReportsTableName, archivedAtColumn,
						buildFilterConditions(issueReportsTableName, true, false),
						buildCursorLimitClause(issueReportsTableName),
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
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(issueReportsTableName, true, true, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, belongsToAccountColumn, belongsToAccountColumn)),
						buildTotalCountSelect(issueReportsTableName, true, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, belongsToAccountColumn, belongsToAccountColumn)),
						issueReportsTableName,
						issueReportsTableName, archivedAtColumn,
						issueReportsTableName, belongsToAccountColumn, belongsToAccountColumn,
						buildFilterConditions(issueReportsTableName, true, false, fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, belongsToAccountColumn, belongsToAccountColumn)),
						buildCursorLimitClause(issueReportsTableName),
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
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(issueReportsTableName, true, true, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantTableColumn, relevantTableColumn)),
						buildTotalCountSelect(issueReportsTableName, true, nil, fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantTableColumn, relevantTableColumn)),
						issueReportsTableName,
						issueReportsTableName, archivedAtColumn,
						issueReportsTableName, relevantTableColumn, relevantTableColumn,
						buildFilterConditions(issueReportsTableName, true, false, fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantTableColumn, relevantTableColumn)),
						buildCursorLimitClause(issueReportsTableName),
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
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(issueReportsTableName, true, true, nil,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantTableColumn, relevantTableColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantRecordIDColumn, relevantRecordIDColumn)),
						buildTotalCountSelect(issueReportsTableName, true, nil,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantTableColumn, relevantTableColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantRecordIDColumn, relevantRecordIDColumn)),
						issueReportsTableName,
						issueReportsTableName, archivedAtColumn,
						issueReportsTableName, relevantTableColumn, relevantTableColumn,
						issueReportsTableName, relevantRecordIDColumn, relevantRecordIDColumn,
						buildFilterConditions(issueReportsTableName, true, false,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantTableColumn, relevantTableColumn),
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", issueReportsTableName, relevantRecordIDColumn, relevantRecordIDColumn)),
						buildCursorLimitClause(issueReportsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
