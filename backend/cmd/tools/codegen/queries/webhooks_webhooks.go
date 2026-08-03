package main

import (
	"fmt"
	"strings"

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

		insertColumns := filterForInsert(webhooksColumns)
		fullSelectColumns := mergeColumns(
			applyToEach(webhooksColumns, func(_ int, s string) string {
				return fullColumnName(webhooksTableName, s)
			}),
			applyToEach(webhookTriggerConfigsColumns, func(_ int, s string) string {
				return fullColumnName(webhookTriggerConfigsTableName, s)
			}),
			5,
		)

		return []*Query{
			{
				Annotation: QueryAnnotation{
					Name: "ArchiveWebhook",
					Type: ExecRowsType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`UPDATE %s SET
	%s = %s
WHERE %s IS NULL
	AND %s = sqlc.arg(%s)
	AND %s = sqlc.arg(%s);`,
					webhooksTableName,
					archivedAtColumn, currentTimeExpression,
					archivedAtColumn,
					idColumn, idColumn,
					belongsToAccountColumn, belongsToAccountColumn,
				)),
			},
			{
				Annotation: QueryAnnotation{
					Name: "CreateWebhook",
					Type: ExecType,
				},
				Content: buildRawQuery((&builq.Builder{}).Addf(`INSERT INTO %s (
	%s
) VALUES (
	%s
);`,
					webhooksTableName,
					strings.Join(insertColumns, ",\n\t"),
					strings.Join(applyToEach(insertColumns, func(_ int, s string) string {
						return fmt.Sprintf("sqlc.arg(%s)", s)
					}), ",\n\t"),
				)),
			},
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
WHERE %s.%s IS NULL
	%s
%s;`,
					strings.Join(fullSelectColumns, ",\n\t"),
					buildFilterCountSelect(webhooksTableName, true, true, []string{}, "webhooks.belongs_to_account = sqlc.arg(belongs_to_account)"),
					buildTotalCountSelect(
						webhooksTableName,
						true,
						nil,
						fmt.Sprintf("%s.%s = sqlc.arg(%s)", webhooksTableName, belongsToAccountColumn, belongsToAccountColumn),
					),
					webhooksTableName,
					webhookTriggerConfigsTableName, webhooksTableName, idColumn, webhookTriggerConfigsTableName, belongsToWebhookColumn,
					webhooksTableName, archivedAtColumn,
					buildFilterConditions(webhooksTableName, true, true, fmt.Sprintf("%s.%s = sqlc.arg(%s)", webhooksTableName, belongsToAccountColumn, belongsToAccountColumn), fmt.Sprintf("%s.%s IS NULL", webhookTriggerConfigsTableName, archivedAtColumn)),
					buildCursorLimitClause(webhooksTableName),
				)),
			},
			// There is deliberately no "webhooks for this account and event" query here any
			// more. Fan-out reads the platform's webhooks_subscriptions table, whose event
			// types carry the account, so that table is the single answer to "who wants this
			// event" — two tables that could answer it would eventually answer differently.
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
		}
	default:
		return nil
	}
}
