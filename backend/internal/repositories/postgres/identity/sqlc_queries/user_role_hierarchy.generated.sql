-- name: CreateUserRoleHierarchy :exec
INSERT INTO user_role_hierarchy (
	id,
	parent_role_id,
	child_role_id
) VALUES (
	sqlc.arg(id),
	sqlc.arg(parent_role_id),
	sqlc.arg(child_role_id)
);

-- name: GetUserRoleHierarchy :many
SELECT
	user_role_hierarchy.id,
	user_role_hierarchy.parent_role_id,
	user_role_hierarchy.child_role_id,
	user_role_hierarchy.created_at,
	user_role_hierarchy.archived_at
FROM user_role_hierarchy
WHERE user_role_hierarchy.archived_at IS NULL;

-- name: ArchiveUserRoleHierarchy :exec
UPDATE user_role_hierarchy SET archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND parent_role_id = sqlc.arg(parent_role_id)
	AND child_role_id = sqlc.arg(child_role_id);
