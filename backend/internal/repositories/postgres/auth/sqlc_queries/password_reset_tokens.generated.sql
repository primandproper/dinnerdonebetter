-- name: CreatePasswordResetToken :exec
INSERT INTO password_reset_tokens (
	id,
	token,
	expires_at,
	belongs_to_user
) VALUES (
	sqlc.arg(id),
	sqlc.arg(token),
	CURRENT_TIMESTAMP + (30 * '1 minutes'::INTERVAL),
	sqlc.arg(belongs_to_user)
);

-- name: GetPasswordResetToken :one
SELECT
	password_reset_tokens.id,
	password_reset_tokens.token,
	password_reset_tokens.expires_at,
	password_reset_tokens.redeemed_at,
	password_reset_tokens.belongs_to_user,
	password_reset_tokens.created_at,
	password_reset_tokens.last_updated_at
FROM password_reset_tokens
WHERE password_reset_tokens.redeemed_at IS NULL
	AND CURRENT_TIMESTAMP < password_reset_tokens.expires_at
	AND password_reset_tokens.token = sqlc.arg(token);

-- name: GetPasswordResetTokenByID :one
SELECT
	password_reset_tokens.id,
	password_reset_tokens.token,
	password_reset_tokens.expires_at,
	password_reset_tokens.redeemed_at,
	password_reset_tokens.belongs_to_user,
	password_reset_tokens.created_at,
	password_reset_tokens.last_updated_at
FROM password_reset_tokens
WHERE password_reset_tokens.redeemed_at IS NULL
	AND CURRENT_TIMESTAMP < password_reset_tokens.expires_at
	AND password_reset_tokens.id = sqlc.arg(id);

-- name: RedeemPasswordResetToken :exec
UPDATE password_reset_tokens SET
	redeemed_at = CURRENT_TIMESTAMP
WHERE redeemed_at IS NULL
	AND id = sqlc.arg(id);
