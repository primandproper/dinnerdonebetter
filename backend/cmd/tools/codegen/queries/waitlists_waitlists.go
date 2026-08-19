package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	waitlistsTableName = "waitlists"
)

func init() {
	registerTableName(waitlistsTableName)
}

var waitlistsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	"valid_until",
	createdAtColumn,
	lastUpdatedAtColumn,
	archivedAtColumn,
}

func buildWaitlistsQueries(database string) []*Query {
	switch database {
	case postgres:
		fullSelectColumns := applyToEach(waitlistsColumns, func(_ int, s string) string {
			return fullColumnName(waitlistsTableName, s)
		})

		return slices.Concat(
			querygen.StandardCRUD(waitlistsTableName, waitlistsColumns,
				querygen.WithEntity("Waitlist", "Waitlists"),
				querygen.WithOmitted(querygen.ArchiveQuery, querygen.ExistsQuery, querygen.ListQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "ArchiveWaitlist",
						Type: ExecRowsType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s,
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s);`,
						waitlistsTableName,
						lastUpdatedAtColumn, currentTimeExpression,
						archivedAtColumn, currentTimeExpression,
						archivedAtColumn,
						idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "CheckWaitlistExistence",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS(
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
);`,
						waitlistsTableName, idColumn,
						waitlistsTableName,
						waitlistsTableName, archivedAtColumn,
						waitlistsTableName, idColumn, idColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "WaitlistIsNotExpired",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS(
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
		AND %s.%s = sqlc.arg(%s)
		AND %s.valid_until >= %s
);`,
						waitlistsTableName, idColumn,
						waitlistsTableName,
						waitlistsTableName, archivedAtColumn,
						waitlistsTableName, idColumn, idColumn,
						waitlistsTableName, currentTimeExpression,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetWaitlists",
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
						buildFilterCountSelect(waitlistsTableName, true, true, nil),
						buildTotalCountSelect(waitlistsTableName, true, nil),
						waitlistsTableName,
						waitlistsTableName, archivedAtColumn,
						buildFilterConditions(waitlistsTableName, true, false),
						buildCursorLimitClause(waitlistsTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetActiveWaitlists",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.valid_until >= %s
	%s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						buildFilterCountSelect(waitlistsTableName, true, true, nil, fmt.Sprintf("%s.valid_until >= %s", waitlistsTableName, currentTimeExpression)),
						buildTotalCountSelect(waitlistsTableName, true, nil, fmt.Sprintf("%s.valid_until >= %s", waitlistsTableName, currentTimeExpression)),
						waitlistsTableName,
						waitlistsTableName, archivedAtColumn,
						waitlistsTableName, currentTimeExpression,
						buildFilterConditions(waitlistsTableName, true, false, fmt.Sprintf("%s.valid_until >= %s", waitlistsTableName, currentTimeExpression)),
						buildCursorLimitClause(waitlistsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
