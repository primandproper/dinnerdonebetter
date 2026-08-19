-- name: CreateSubscription :exec
INSERT INTO subscriptions (
	id,
	belongs_to_account,
	product_id,
	external_subscription_id,
	status,
	current_period_start,
	current_period_end
) VALUES (
	sqlc.arg(id),
	sqlc.arg(belongs_to_account),
	sqlc.arg(product_id),
	sqlc.arg(external_subscription_id),
	sqlc.arg(status),
	sqlc.arg(current_period_start),
	sqlc.arg(current_period_end)
);

-- name: GetSubscription :one
SELECT
	subscriptions.id,
	subscriptions.belongs_to_account,
	subscriptions.product_id,
	subscriptions.external_subscription_id,
	subscriptions.status,
	subscriptions.current_period_start,
	subscriptions.current_period_end,
	subscriptions.created_at,
	subscriptions.last_updated_at,
	subscriptions.archived_at
FROM subscriptions
WHERE subscriptions.archived_at IS NULL
	AND subscriptions.id = sqlc.arg(id);

-- name: UpdateSubscription :execrows
UPDATE subscriptions SET
	external_subscription_id = sqlc.arg(external_subscription_id),
	status = sqlc.arg(status),
	current_period_start = sqlc.arg(current_period_start),
	current_period_end = sqlc.arg(current_period_end),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveSubscription :execrows
UPDATE subscriptions SET archived_at = NOW(), last_updated_at = NOW() WHERE archived_at IS NULL AND id = sqlc.arg(id);

-- name: GetSubscriptionByExternalID :one
SELECT
	subscriptions.id,
	subscriptions.belongs_to_account,
	subscriptions.product_id,
	subscriptions.external_subscription_id,
	subscriptions.status,
	subscriptions.current_period_start,
	subscriptions.current_period_end,
	subscriptions.created_at,
	subscriptions.last_updated_at,
	subscriptions.archived_at
FROM subscriptions
WHERE subscriptions.archived_at IS NULL
AND subscriptions.external_subscription_id = sqlc.arg(external_subscription_id);

-- name: GetSubscriptionsForAccount :many
SELECT
	subscriptions.id,
	subscriptions.belongs_to_account,
	subscriptions.product_id,
	subscriptions.external_subscription_id,
	subscriptions.status,
	subscriptions.current_period_start,
	subscriptions.current_period_end,
	subscriptions.created_at,
	subscriptions.last_updated_at,
	subscriptions.archived_at,
	(
		SELECT COUNT(subscriptions.id)
		FROM subscriptions
		WHERE subscriptions.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND subscriptions.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				subscriptions.last_updated_at IS NULL
				OR subscriptions.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				subscriptions.last_updated_at IS NULL
				OR subscriptions.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR subscriptions.archived_at IS NULL)
			AND subscriptions.belongs_to_account = sqlc.arg(belongs_to_account)
	) AS filtered_count,
	(
		SELECT COUNT(subscriptions.id)
		FROM subscriptions
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR subscriptions.archived_at IS NULL)
			AND subscriptions.belongs_to_account = sqlc.arg(belongs_to_account)
	) AS total_count
FROM subscriptions
WHERE subscriptions.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND subscriptions.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		subscriptions.last_updated_at IS NULL
		OR subscriptions.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		subscriptions.last_updated_at IS NULL
		OR subscriptions.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR subscriptions.archived_at IS NULL)
	AND subscriptions.belongs_to_account = sqlc.arg(belongs_to_account)
	AND subscriptions.belongs_to_account = sqlc.arg(belongs_to_account)
	AND subscriptions.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY subscriptions.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: UpdateSubscriptionStatus :execrows
UPDATE subscriptions SET status = sqlc.arg(status), last_updated_at = NOW()
WHERE archived_at IS NULL AND id = sqlc.arg(id);
