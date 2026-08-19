package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	oauth2ClientsTableName = "oauth2_clients"
	clientIDColumn         = "client_id"
)

func init() {
	registerTableName(oauth2ClientsTableName)
}

var oauth2ClientsColumns = []string{
	idColumn,
	nameColumn,
	descriptionColumn,
	clientIDColumn,
	"client_secret",
	createdAtColumn,
	archivedAtColumn,
}

func buildOAuth2ClientsQueries(database string) []*Query {
	switch database {
	case postgres:

		return slices.Concat(
			querygen.StandardCRUD(oauth2ClientsTableName, oauth2ClientsColumns,
				querygen.WithEntity("OAuth2Client", "OAuth2Clients"),
				querygen.WithQueryName(querygen.GetQuery, "GetOAuth2ClientByDatabaseID"),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "GetOAuth2ClientByClientID",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(applyToEach(oauth2ClientsColumns, func(_ int, s string) string {
							return fmt.Sprintf("%s.%s", oauth2ClientsTableName, s)
						}), ",\n\t"),
						oauth2ClientsTableName,
						oauth2ClientsTableName, archivedAtColumn,
						oauth2ClientsTableName, clientIDColumn, clientIDColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetOAuth2Clients",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
WHERE %s
%s;
`,
						strings.Join(applyToEach(oauth2ClientsColumns, func(_ int, s string) string {
							return fmt.Sprintf("%s.%s", oauth2ClientsTableName, s)
						}), ",\n\t"),
						querygen.FilterCountSelect(oauth2ClientsTableName, oauth2ClientsColumns, nil),
						querygen.TotalCountSelect(oauth2ClientsTableName, oauth2ClientsColumns, nil),
						oauth2ClientsTableName,
						querygen.FilterConditions(oauth2ClientsTableName, oauth2ClientsColumns),
						querygen.CursorLimitClause(oauth2ClientsTableName),
					)),
				},
			},
		)
	default:
		return nil
	}
}
