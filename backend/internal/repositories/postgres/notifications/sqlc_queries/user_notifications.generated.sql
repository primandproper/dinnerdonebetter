-- name: CreateUserNotification :exec
INSERT INTO user_notifications (
	id,
	content,
	belongs_to_user
) VALUES (
	sqlc.arg(id),
	sqlc.arg(content),
	sqlc.arg(belongs_to_user)
);

-- name: GetUserNotification :one
SELECT
	user_notifications.id,
	user_notifications.content,
	user_notifications.status,
	user_notifications.belongs_to_user,
	user_notifications.created_at,
	user_notifications.last_updated_at
FROM user_notifications
WHERE belongs_to_user = sqlc.arg(belongs_to_user)
AND user_notifications.id = sqlc.arg(id);

-- name: CheckUserNotificationExistence :one
SELECT EXISTS(
	SELECT user_notifications.id
	FROM user_notifications
	WHERE user_notifications.id = sqlc.arg(id)
	AND user_notifications.belongs_to_user = sqlc.arg(belongs_to_user)
);

-- name: GetUserNotificationsForUser :many
SELECT
	user_notifications.id,
	user_notifications.content,
	user_notifications.status,
	user_notifications.belongs_to_user,
	user_notifications.created_at,
	user_notifications.last_updated_at,
	(
		SELECT COUNT(user_notifications.id)
		FROM user_notifications
		WHERE user_notifications.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND user_notifications.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				user_notifications.last_updated_at IS NULL
				OR user_notifications.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				user_notifications.last_updated_at IS NULL
				OR user_notifications.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND user_notifications.status != 'dismissed'
			AND user_notifications.belongs_to_user = sqlc.arg(user_id)
	) AS filtered_count,
	(
		SELECT COUNT(user_notifications.id)
		FROM user_notifications
		WHERE user_notifications.status != 'dismissed'
			AND user_notifications.belongs_to_user = sqlc.arg(user_id)
	) AS total_count
FROM user_notifications
WHERE user_notifications.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND user_notifications.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		user_notifications.last_updated_at IS NULL
		OR user_notifications.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		user_notifications.last_updated_at IS NULL
		OR user_notifications.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND user_notifications.status != 'dismissed'
	AND user_notifications.belongs_to_user = sqlc.arg(user_id)
	AND user_notifications.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY user_notifications.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: UpdateUserNotification :execrows
UPDATE user_notifications SET
	status = sqlc.arg(status),
	last_updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id);
