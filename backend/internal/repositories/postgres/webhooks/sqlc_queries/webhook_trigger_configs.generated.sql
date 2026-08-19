-- name: CreateWebhookTriggerConfig :exec
INSERT INTO webhook_trigger_configs (
	id,
	trigger_event,
	belongs_to_webhook
) VALUES (
	sqlc.arg(id),
	sqlc.arg(trigger_event),
	sqlc.arg(belongs_to_webhook)
);

-- name: ArchiveWebhookTriggerConfig :execrows
UPDATE webhook_trigger_configs SET
	archived_at = NOW()
WHERE webhook_trigger_configs.archived_at IS NULL
	AND webhook_trigger_configs.id = sqlc.arg(id)
	AND webhook_trigger_configs.belongs_to_webhook IN (
		SELECT webhooks.id FROM webhooks
		WHERE webhooks.id = sqlc.arg(belongs_to_webhook)
				AND webhooks.belongs_to_account = sqlc.arg(belongs_to_account)
				AND webhooks.archived_at IS NULL
	);
