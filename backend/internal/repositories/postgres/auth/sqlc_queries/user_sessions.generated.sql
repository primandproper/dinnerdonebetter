-- name: CreateUserSession :exec
INSERT INTO user_sessions (
	id,
	belongs_to_user,
	session_token_id,
	refresh_token_id,
	client_ip,
	user_agent,
	device_name,
	login_method,
	expires_at
) VALUES (
	sqlc.arg(id),
	sqlc.arg(belongs_to_user),
	sqlc.arg(session_token_id),
	sqlc.arg(refresh_token_id),
	sqlc.arg(client_ip),
	sqlc.arg(user_agent),
	sqlc.arg(device_name),
	sqlc.arg(login_method),
	sqlc.arg(expires_at)
);

-- name: GetUserSessionBySessionTokenID :one
SELECT
	user_sessions.id,
	user_sessions.belongs_to_user,
	user_sessions.session_token_id,
	user_sessions.refresh_token_id,
	user_sessions.client_ip,
	user_sessions.user_agent,
	user_sessions.device_name,
	user_sessions.login_method,
	user_sessions.created_at,
	user_sessions.last_active_at,
	user_sessions.expires_at,
	user_sessions.revoked_at
FROM user_sessions
WHERE user_sessions.session_token_id = sqlc.arg(session_token_id)
	AND user_sessions.revoked_at IS NULL
	AND user_sessions.expires_at > CURRENT_TIMESTAMP;

-- name: GetUserSessionByRefreshTokenID :one
SELECT
	user_sessions.id,
	user_sessions.belongs_to_user,
	user_sessions.session_token_id,
	user_sessions.refresh_token_id,
	user_sessions.client_ip,
	user_sessions.user_agent,
	user_sessions.device_name,
	user_sessions.login_method,
	user_sessions.created_at,
	user_sessions.last_active_at,
	user_sessions.expires_at,
	user_sessions.revoked_at
FROM user_sessions
WHERE user_sessions.refresh_token_id = sqlc.arg(refresh_token_id)
	AND user_sessions.revoked_at IS NULL;

-- name: GetActiveSessionsForUser :many
SELECT
	user_sessions.id,
	user_sessions.belongs_to_user,
	user_sessions.session_token_id,
	user_sessions.refresh_token_id,
	user_sessions.client_ip,
	user_sessions.user_agent,
	user_sessions.device_name,
	user_sessions.login_method,
	user_sessions.created_at,
	user_sessions.last_active_at,
	user_sessions.expires_at,
	user_sessions.revoked_at,
	(
		SELECT COUNT(user_sessions.id)
		FROM user_sessions
		WHERE user_sessions.belongs_to_user = sqlc.arg(belongs_to_user)
				AND user_sessions.revoked_at IS NULL
				AND user_sessions.expires_at > CURRENT_TIMESTAMP
				AND user_sessions.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
				AND user_sessions.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	) AS filtered_count,
	(
		SELECT COUNT(user_sessions.id)
		FROM user_sessions
		WHERE user_sessions.belongs_to_user = sqlc.arg(belongs_to_user)
				AND user_sessions.revoked_at IS NULL
				AND user_sessions.expires_at > CURRENT_TIMESTAMP
	) AS total_count
FROM user_sessions
WHERE user_sessions.belongs_to_user = sqlc.arg(belongs_to_user)
	AND user_sessions.revoked_at IS NULL
	AND user_sessions.expires_at > CURRENT_TIMESTAMP
	AND user_sessions.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND user_sessions.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND user_sessions.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY user_sessions.last_active_at DESC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: RevokeUserSession :execrows
UPDATE user_sessions SET
	revoked_at = CURRENT_TIMESTAMP
WHERE user_sessions.id = sqlc.arg(id)
	AND user_sessions.belongs_to_user = sqlc.arg(belongs_to_user)
	AND user_sessions.revoked_at IS NULL;

-- name: RevokeAllSessionsForUser :execrows
UPDATE user_sessions SET
	revoked_at = CURRENT_TIMESTAMP
WHERE user_sessions.belongs_to_user = sqlc.arg(belongs_to_user)
	AND user_sessions.revoked_at IS NULL;

-- name: RevokeAllSessionsForUserExcept :execrows
UPDATE user_sessions SET
	revoked_at = CURRENT_TIMESTAMP
WHERE user_sessions.belongs_to_user = sqlc.arg(belongs_to_user)
	AND user_sessions.id != sqlc.arg(session_id)
	AND user_sessions.revoked_at IS NULL;

-- name: UpdateSessionTokenIDs :execrows
UPDATE user_sessions SET
	session_token_id = sqlc.arg(session_token_id),
	refresh_token_id = sqlc.arg(refresh_token_id),
	expires_at = sqlc.arg(expires_at),
	last_active_at = CURRENT_TIMESTAMP
WHERE user_sessions.id = sqlc.arg(id)
	AND user_sessions.revoked_at IS NULL;

-- name: TouchSessionLastActive :execrows
UPDATE user_sessions SET
	last_active_at = CURRENT_TIMESTAMP
WHERE user_sessions.session_token_id = sqlc.arg(session_token_id)
	AND user_sessions.revoked_at IS NULL;

-- name: CleanupExpiredSessions :execrows
UPDATE user_sessions SET
	revoked_at = CURRENT_TIMESTAMP
WHERE user_sessions.revoked_at IS NULL
	AND user_sessions.expires_at < CURRENT_TIMESTAMP;
