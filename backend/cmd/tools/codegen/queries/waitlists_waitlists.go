package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v12/database/querygen"

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
			return querygen.Qualify(waitlistsTableName, s)
		})

		return slices.Concat(
			pgGen.StandardCRUD(waitlistsTableName, waitlistsColumns,
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
						lastUpdatedAtColumn, querygen.NowExpression,
						archivedAtColumn, querygen.NowExpression,
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
						waitlistsTableName, querygen.NowExpression,
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
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(waitlistsTableName, waitlistsColumns, nil),
						pgGen.TotalCountSelect(waitlistsTableName, waitlistsColumns, nil),
						waitlistsTableName,
						pgGen.FilterConditions(waitlistsTableName, waitlistsColumns),
						pgGen.CursorLimitClause(waitlistsTableName),
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
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						pgGen.FilterCountSelect(waitlistsTableName, waitlistsColumns, nil, fmt.Sprintf("%s.valid_until >= %s", waitlistsTableName, querygen.NowExpression)),
						pgGen.TotalCountSelect(waitlistsTableName, waitlistsColumns, nil, fmt.Sprintf("%s.valid_until >= %s", waitlistsTableName, querygen.NowExpression)),
						waitlistsTableName,
						pgGen.FilterConditions(waitlistsTableName, waitlistsColumns,
							fmt.Sprintf("%s.valid_until >= %s", waitlistsTableName, querygen.NowExpression),
						),
						pgGen.CursorLimitClause(waitlistsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
