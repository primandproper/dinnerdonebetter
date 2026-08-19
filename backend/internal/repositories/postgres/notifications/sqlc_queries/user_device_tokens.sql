-- name: CreateUserDeviceToken :exec
INSERT INTO user_device_tokens (
	id,
	belongs_to_user,
	device_token,
	platform
) VALUES (
	sqlc.arg(id),
	sqlc.arg(belongs_to_user),
	sqlc.arg(device_token),
	sqlc.arg(platform)
);

-- name: UpsertUserDeviceToken :one
INSERT INTO user_device_tokens (
	id,
	belongs_to_user,
	device_token,
	platform
) VALUES (
	sqlc.arg(id),
	sqlc.arg(belongs_to_user),
	sqlc.arg(device_token),
	sqlc.arg(platform)
)
ON CONFLICT (belongs_to_user, device_token) WHERE (archived_at IS NULL)
DO UPDATE SET last_updated_at = NOW()
RETURNING *;

-- name: GetUserDeviceToken :one
SELECT
	user_device_tokens.id,
	user_device_tokens.belongs_to_user,
	user_device_tokens.device_token,
	user_device_tokens.platform,
	user_device_tokens.created_at,
	user_device_tokens.last_updated_at,
	user_device_tokens.archived_at
FROM user_device_tokens
WHERE user_device_tokens.belongs_to_user = sqlc.arg(belongs_to_user)
	AND user_device_tokens.id = sqlc.arg(id)
	AND user_device_tokens.archived_at IS NULL;

-- name: CheckUserDeviceTokenExistence :one
SELECT EXISTS(
	SELECT user_device_tokens.id
	FROM user_device_tokens
	WHERE user_device_tokens.id = sqlc.arg(id)
		AND user_device_tokens.belongs_to_user = sqlc.arg(belongs_to_user)
		AND user_device_tokens.archived_at IS NULL
);

-- name: GetUserDeviceTokensForUser :many
SELECT
	user_device_tokens.id,
	user_device_tokens.belongs_to_user,
	user_device_tokens.device_token,
	user_device_tokens.platform,
	user_device_tokens.created_at,
	user_device_tokens.last_updated_at,
	user_device_tokens.archived_at,
	(
		SELECT COUNT(user_device_tokens.id)
		FROM user_device_tokens
		WHERE
			(COALESCE(sqlc.narg(include_archived), false)::boolean OR user_device_tokens.archived_at IS NULL)
			AND user_device_tokens.belongs_to_user = sqlc.arg(user_id)
			AND (sqlc.narg(platform_filter)::TEXT IS NULL OR user_device_tokens.platform = sqlc.narg(platform_filter)::TEXT)
			AND user_device_tokens.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND user_device_tokens.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	) AS filtered_count,
	(
		SELECT COUNT(user_device_tokens.id)
		FROM user_device_tokens
		WHERE
			(COALESCE(sqlc.narg(include_archived), false)::boolean OR user_device_tokens.archived_at IS NULL)
			AND user_device_tokens.belongs_to_user = sqlc.arg(user_id)
			AND (sqlc.narg(platform_filter)::TEXT IS NULL OR user_device_tokens.platform = sqlc.narg(platform_filter)::TEXT)
	) AS total_count
FROM user_device_tokens
WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR user_device_tokens.archived_at IS NULL)
	AND user_device_tokens.belongs_to_user = sqlc.arg(user_id)
	AND (sqlc.narg(platform_filter)::TEXT IS NULL OR user_device_tokens.platform = sqlc.narg(platform_filter)::TEXT)
	AND user_device_tokens.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND user_device_tokens.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND user_device_tokens.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY user_device_tokens.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: UpdateUserDeviceToken :execrows
UPDATE user_device_tokens SET
	platform = sqlc.arg(platform),
	last_updated_at = NOW()
WHERE id = sqlc.arg(id)
	AND belongs_to_user = sqlc.arg(belongs_to_user)
	AND archived_at IS NULL;

-- name: ArchiveUserDeviceToken :execrows
UPDATE user_device_tokens SET
	archived_at = NOW()
WHERE id = sqlc.arg(id)
	AND belongs_to_user = sqlc.arg(belongs_to_user)
	AND archived_at IS NULL;
