-- name: SetUserAccountStatus :execrows
UPDATE users SET
	last_updated_at = CURRENT_TIMESTAMP,
	user_account_status = sqlc.arg(user_account_status),
	user_account_status_explanation = sqlc.arg(user_account_status_explanation)
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: SetUserRequiresPasswordChange :execrows
UPDATE users SET
	last_updated_at = CURRENT_TIMESTAMP,
	requires_password_change = sqlc.arg(requires_password_change)
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);
