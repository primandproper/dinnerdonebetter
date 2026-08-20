-- name: CreateUserRolePermission :exec
INSERT INTO user_role_permissions (
	id,
	role_id,
	permission_id
) VALUES (
	sqlc.arg(id),
	sqlc.arg(role_id),
	sqlc.arg(permission_id)
);

-- name: GetUserRolePermissionsForRole :many
SELECT
	user_role_permissions.id,
	user_role_permissions.role_id,
	user_role_permissions.permission_id,
	user_role_permissions.created_at,
	user_role_permissions.archived_at
FROM user_role_permissions
WHERE user_role_permissions.archived_at IS NULL
	AND user_role_permissions.role_id = sqlc.arg(role_id);

-- name: ArchiveUserRolePermission :exec
UPDATE user_role_permissions SET archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND role_id = sqlc.arg(role_id)
	AND permission_id = sqlc.arg(permission_id);
