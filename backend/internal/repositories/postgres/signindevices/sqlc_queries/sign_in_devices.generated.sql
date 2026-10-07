-- name: UpsertSignInDevice :exec
INSERT INTO sign_in_devices (family_id, belongs_to_user, ip_address, user_agent, device_name, expires_at)
VALUES (sqlc.arg(family_id), sqlc.arg(belongs_to_user), sqlc.arg(ip_address), sqlc.arg(user_agent), sqlc.arg(device_name), sqlc.arg(expires_at))
ON CONFLICT (family_id) DO UPDATE SET
	ip_address = EXCLUDED.ip_address,
	user_agent = EXCLUDED.user_agent,
	device_name = EXCLUDED.device_name,
	expires_at = EXCLUDED.expires_at,
	last_seen_at = NOW()
WHERE sign_in_devices.belongs_to_user = EXCLUDED.belongs_to_user;

-- name: GetSignInDevicesForFamilies :many
SELECT family_id, ip_address, user_agent, device_name, created_at, last_seen_at
FROM sign_in_devices
WHERE belongs_to_user = sqlc.arg(belongs_to_user)
	AND family_id = ANY(sqlc.arg(family_ids)::TEXT[]);

-- name: GetSignInDevicesForUser :many
SELECT family_id, ip_address, user_agent, device_name, created_at, last_seen_at, expires_at
FROM sign_in_devices
WHERE belongs_to_user = sqlc.arg(belongs_to_user)
ORDER BY created_at;

-- name: DeleteExpiredSignInDevices :execrows
DELETE FROM sign_in_devices WHERE expires_at < NOW();
