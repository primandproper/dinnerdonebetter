-- name: CreateWebhookTriggerEvent :exec
INSERT INTO webhook_trigger_events (
	id,
	name,
	description
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description)
);

-- name: GetWebhookTriggerEvent :one
SELECT
	webhook_trigger_events.id,
	webhook_trigger_events.name,
	webhook_trigger_events.description,
	webhook_trigger_events.created_at,
	webhook_trigger_events.last_updated_at,
	webhook_trigger_events.archived_at
FROM webhook_trigger_events
WHERE webhook_trigger_events.archived_at IS NULL
	AND webhook_trigger_events.id = sqlc.arg(id);

-- name: CheckWebhookTriggerEventExistence :one
SELECT EXISTS(
	SELECT webhook_trigger_events.id
	FROM webhook_trigger_events
	WHERE webhook_trigger_events.archived_at IS NULL
		AND webhook_trigger_events.id = sqlc.arg(id)
);

-- name: GetWebhookTriggerEvents :many
SELECT
	webhook_trigger_events.id,
	webhook_trigger_events.name,
	webhook_trigger_events.description,
	webhook_trigger_events.created_at,
	webhook_trigger_events.last_updated_at,
	webhook_trigger_events.archived_at,
	(
		SELECT COUNT(webhook_trigger_events.id)
		FROM webhook_trigger_events
		WHERE webhook_trigger_events.archived_at IS NULL
			AND
			webhook_trigger_events.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND webhook_trigger_events.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				webhook_trigger_events.last_updated_at IS NULL
				OR webhook_trigger_events.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				webhook_trigger_events.last_updated_at IS NULL
				OR webhook_trigger_events.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR webhook_trigger_events.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(webhook_trigger_events.id)
		FROM webhook_trigger_events
		WHERE webhook_trigger_events.archived_at IS NULL
	) AS total_count
FROM webhook_trigger_events
WHERE webhook_trigger_events.archived_at IS NULL
	AND webhook_trigger_events.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND webhook_trigger_events.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		webhook_trigger_events.last_updated_at IS NULL
		OR webhook_trigger_events.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		webhook_trigger_events.last_updated_at IS NULL
		OR webhook_trigger_events.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR webhook_trigger_events.archived_at IS NULL)
	AND webhook_trigger_events.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY webhook_trigger_events.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: UpdateWebhookTriggerEvent :execrows
UPDATE webhook_trigger_events SET
	name = sqlc.arg(name),
	description = sqlc.arg(description),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveWebhookTriggerEvent :execrows
UPDATE webhook_trigger_events SET
	last_updated_at = NOW(),
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);
