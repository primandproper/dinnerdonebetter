-- name: CreateWebhook :exec
INSERT INTO webhooks (
	id,
	name,
	content_type,
	url,
	method,
	created_by_user,
	belongs_to_account
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(content_type),
	sqlc.arg(url),
	sqlc.arg(method),
	sqlc.arg(created_by_user),
	sqlc.arg(belongs_to_account)
);

-- name: ArchiveWebhook :execrows
UPDATE webhooks SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_account = sqlc.arg(belongs_to_account);

-- name: CheckWebhookExistence :one
SELECT EXISTS(
	SELECT webhooks.id
	FROM webhooks
	WHERE webhooks.archived_at IS NULL
	AND webhooks.id = sqlc.arg(id)
	AND webhooks.belongs_to_account = sqlc.arg(belongs_to_account)
);

-- name: GetWebhooksForAccount :many
SELECT
	webhooks.id,
	webhooks.name,
	webhooks.content_type,
	webhooks.url,
	webhooks.method,
	webhook_trigger_configs.id,
	webhook_trigger_configs.trigger_event,
	webhook_trigger_configs.belongs_to_webhook,
	webhook_trigger_configs.created_at,
	webhook_trigger_configs.archived_at,
	webhooks.created_at,
	webhooks.last_updated_at,
	webhooks.archived_at,
	webhooks.created_by_user,
	webhooks.belongs_to_account,
	(
		SELECT COUNT(webhooks.id)
		FROM webhooks
		WHERE webhooks.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND webhooks.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				webhooks.last_updated_at IS NULL
				OR webhooks.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				webhooks.last_updated_at IS NULL
				OR webhooks.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR webhooks.archived_at IS NULL)
			AND webhooks.belongs_to_account = sqlc.arg(belongs_to_account)
	) AS filtered_count,
	(
		SELECT COUNT(webhooks.id)
		FROM webhooks
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR webhooks.archived_at IS NULL)
			AND webhooks.belongs_to_account = sqlc.arg(belongs_to_account)
	) AS total_count
FROM webhooks
	LEFT JOIN webhook_trigger_configs ON webhooks.id = webhook_trigger_configs.belongs_to_webhook
WHERE webhooks.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND webhooks.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		webhooks.last_updated_at IS NULL
		OR webhooks.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		webhooks.last_updated_at IS NULL
		OR webhooks.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR webhooks.archived_at IS NULL)
	AND webhooks.belongs_to_account = sqlc.arg(belongs_to_account)
	AND webhook_trigger_configs.archived_at IS NULL
	AND webhooks.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY webhooks.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetWebhook :many
SELECT
	webhooks.id as webhook_id,
	webhooks.name as webhook_name,
	webhooks.content_type as webhook_content_type,
	webhooks.url as webhook_url,
	webhooks.method as webhook_method,
	webhook_trigger_configs.id as webhook_trigger_config_id,
	webhook_trigger_configs.trigger_event as webhook_trigger_config_trigger_event,
	webhook_trigger_configs.belongs_to_webhook as webhook_trigger_config_belongs_to_webhook,
	webhook_trigger_configs.created_at as webhook_trigger_config_created_at,
	webhook_trigger_configs.archived_at as webhook_trigger_config_archived_at,
	webhooks.created_at as webhook_created_at,
	webhooks.last_updated_at as webhook_last_updated_at,
	webhooks.archived_at as webhook_archived_at,
	webhooks.created_by_user as webhook_created_by_user,
	webhooks.belongs_to_account as webhook_belongs_to_account
FROM webhooks
	LEFT JOIN webhook_trigger_configs ON webhooks.id = webhook_trigger_configs.belongs_to_webhook AND webhook_trigger_configs.archived_at IS NULL
WHERE webhooks.archived_at IS NULL
	AND webhooks.belongs_to_account = sqlc.arg(belongs_to_account)
	AND webhooks.id = sqlc.arg(id);
