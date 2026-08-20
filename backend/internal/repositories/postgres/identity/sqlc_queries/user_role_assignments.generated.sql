-- name: AssignRoleToUser :exec
INSERT INTO user_role_assignments (
	id,
	user_id,
	role_id,
	account_id
) VALUES (
	sqlc.arg(id),
	sqlc.arg(user_id),
	sqlc.arg(role_id),
	sqlc.arg(account_id)
);

-- name: ArchiveRoleAssignment :exec
UPDATE user_role_assignments SET archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND user_id = sqlc.arg(user_id);

-- name: ArchiveRoleAssignmentsForUserAndAccount :exec
UPDATE user_role_assignments SET archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND user_id = sqlc.arg(user_id)
	AND account_id = sqlc.arg(account_id);

-- name: UpdateAccountRoleAssignment :exec
UPDATE user_role_assignments SET role_id = sqlc.arg(new_role_id)
WHERE archived_at IS NULL
	AND user_id = sqlc.arg(user_id)
	AND account_id = sqlc.arg(account_id);

-- name: GetServicePermissionsForUser :many
WITH RECURSIVE role_tree AS (
	SELECT user_roles.id AS role_id, user_roles.name AS role_name
	FROM user_role_assignments
	JOIN user_roles ON user_roles.id = user_role_assignments.role_id
	WHERE user_role_assignments.user_id = sqlc.arg(user_id)
		AND user_role_assignments.account_id IS NULL
		AND user_role_assignments.archived_at IS NULL
		AND user_roles.archived_at IS NULL
	UNION
	SELECT user_roles.id, user_roles.name
	FROM role_tree rt
	JOIN user_role_hierarchy ON user_role_hierarchy.child_role_id = rt.role_id
	JOIN user_roles ON user_roles.id = user_role_hierarchy.parent_role_id
	WHERE user_role_hierarchy.archived_at IS NULL
		AND user_roles.archived_at IS NULL
)
SELECT DISTINCT permissions.name AS permission_name
FROM role_tree rt
JOIN user_role_permissions ON user_role_permissions.role_id = rt.role_id
JOIN permissions ON permissions.id = user_role_permissions.permission_id
WHERE user_role_permissions.archived_at IS NULL
	AND permissions.archived_at IS NULL;

-- name: GetAccountPermissionsForUser :many
WITH RECURSIVE role_tree AS (
	SELECT user_roles.id AS role_id, user_roles.name AS role_name, user_role_assignments.account_id AS account_id
	FROM user_role_assignments
	JOIN user_roles ON user_roles.id = user_role_assignments.role_id
	WHERE user_role_assignments.user_id = sqlc.arg(user_id)
		AND user_role_assignments.account_id IS NOT NULL
		AND user_role_assignments.archived_at IS NULL
		AND user_roles.archived_at IS NULL
	UNION
	SELECT user_roles.id, user_roles.name, rt.account_id
	FROM role_tree rt
	JOIN user_role_hierarchy ON user_role_hierarchy.child_role_id = rt.role_id
	JOIN user_roles ON user_roles.id = user_role_hierarchy.parent_role_id
	WHERE user_role_hierarchy.archived_at IS NULL
		AND user_roles.archived_at IS NULL
)
SELECT DISTINCT rt.account_id, permissions.name AS permission_name
FROM role_tree rt
JOIN user_role_permissions ON user_role_permissions.role_id = rt.role_id
JOIN permissions ON permissions.id = user_role_permissions.permission_id
WHERE user_role_permissions.archived_at IS NULL
	AND permissions.archived_at IS NULL;

-- name: GetServiceRoleNamesForUser :many
SELECT user_roles.name AS role_name
FROM user_role_assignments
JOIN user_roles ON user_roles.id = user_role_assignments.role_id
WHERE user_role_assignments.user_id = sqlc.arg(user_id)
	AND user_role_assignments.account_id IS NULL
	AND user_role_assignments.archived_at IS NULL
	AND user_roles.archived_at IS NULL;
