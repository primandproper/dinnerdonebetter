package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/primandproper/platform-go/v11/database/querygen"

	"github.com/cristalhq/builq"
)

const (
	webhooksTableName = "webhooks"
)

func init() {
	registerTableName(webhooksTableName)
}

var (
	webhooksColumns = []string{
		idColumn,
		nameColumn,
		"content_type",
		"url",
		"method",
		createdAtColumn,
		lastUpdatedAtColumn,
		archivedAtColumn,
		createdByUserColumn,
		belongsToAccountColumn,
	}
)

func buildWebhooksQueries(database string) []*Query {
	switch database {
	case postgres:

		fullSelectColumns := mergeColumns(
			applyToEach(webhooksColumns, func(_ int, s string) string {
				return fullColumnName(webhooksTableName, s)
			}),
			applyToEach(webhookTriggerConfigsColumns, func(_ int, s string) string {
				return fullColumnName(webhookTriggerConfigsTableName, s)
			}),
			5,
		)

		return slices.Concat(
			querygen.StandardCRUD(webhooksTableName, webhooksColumns,
				querygen.WithEntity("Webhook", "Webhooks"),
				querygen.WithOwnership(belongsToAccountColumn),
				querygen.WithOmitted(querygen.ExistsQuery, querygen.GetQuery, querygen.ListQuery, querygen.UpdateQuery),
			),
			[]*Query{
				{
					Annotation: QueryAnnotation{
						Name: "CheckWebhookExistence",
						Type: OneType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT EXISTS(
	SELECT %s.%s
	FROM %s
	WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s)
);`,
						webhooksTableName, idColumn,
						webhooksTableName,
						webhooksTableName, archivedAtColumn,
						webhooksTableName, idColumn, idColumn,
						webhooksTableName, belongsToAccountColumn, belongsToAccountColumn,
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetWebhooksForAccount",
						Type: ManyType,
					},
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s,
	%s,
	%s
FROM %s
	LEFT JOIN %s ON %s.%s = %s.%s
WHERE %s
%s;`,
						strings.Join(fullSelectColumns, ",\n\t"),
						querygen.FilterCountSelect(webhooksTableName, webhooksColumns, []string{}, "webhooks.belongs_to_account = sqlc.arg(belongs_to_account)"),
						querygen.TotalCountSelect(
							webhooksTableName,
							webhooksColumns,
							nil,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", webhooksTableName, belongsToAccountColumn, belongsToAccountColumn),
						),
						webhooksTableName,
						webhookTriggerConfigsTableName,
						webhooksTableName,
						idColumn,
						webhookTriggerConfigsTableName,
						belongsToWebhookColumn,
						querygen.FilterConditions(webhooksTableName, webhooksColumns,
							fmt.Sprintf("%s.%s = sqlc.arg(%s)", webhooksTableName, belongsToAccountColumn, belongsToAccountColumn),
							fmt.Sprintf("%s.%s IS NULL", webhookTriggerConfigsTableName, archivedAtColumn),
						),
						querygen.CursorLimitClause(webhooksTableName),
					)),
				},
				{
					Annotation: QueryAnnotation{
						Name: "GetWebhook",
						Type: ManyType,
					},
					// The archived-config filter belongs in the JOIN, not the WHERE.
					//
					// In the WHERE it silently turns the LEFT JOIN into an inner one: a webhook
					// whose every trigger config is archived matches no rows and disappears
					// from its owner's view entirely, which is what unsubscribing from your
					// last event used to do to a webhook.
					Content: buildRawQuery((&builq.Builder{}).Addf(`SELECT
	%s
FROM %s
	LEFT JOIN %s ON %s.%s = %s.%s AND %s.%s IS NULL
WHERE %s.%s IS NULL
	AND %s.%s = sqlc.arg(%s)
	AND %s.%s = sqlc.arg(%s);`,
						strings.Join(applyToEach(fullSelectColumns, func(_ int, s string) string {
							parts := strings.Split(s, ".")
							return fmt.Sprintf("%s as %s_%s",
								s, strings.TrimSuffix(parts[0], "s"), parts[1],
							)
						}), ",\n\t"),
						webhooksTableName,
						webhookTriggerConfigsTableName, webhooksTableName, idColumn, webhookTriggerConfigsTableName, belongsToWebhookColumn,
						webhookTriggerConfigsTableName, archivedAtColumn,
						webhooksTableName, archivedAtColumn,
						webhooksTableName, belongsToAccountColumn, belongsToAccountColumn,
						webhooksTableName, idColumn, idColumn,
					)),
				},
			},
		)
	default:
		return nil
	}
}
