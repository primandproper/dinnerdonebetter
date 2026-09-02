-- name: AssignRoleToUser :exec
INSERT INTO user_role_assignments (
	id,
	user_id,
	role_name,
	account_id
) VALUES (
	sqlc.arg(id),
	sqlc.arg(user_id),
	sqlc.arg(role_name),
	sqlc.arg(account_id)
);

-- name: ArchiveRoleAssignmentsForUserAndAccount :exec
UPDATE user_role_assignments SET archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND user_id = sqlc.arg(user_id)
	AND account_id = sqlc.arg(account_id);

-- name: UpdateAccountRoleAssignment :exec
UPDATE user_role_assignments SET role_name = sqlc.arg(new_role_name)
WHERE archived_at IS NULL
	AND user_id = sqlc.arg(user_id)
	AND account_id = sqlc.arg(account_id);

-- name: GetRoleAssignmentsForUser :many
SELECT user_role_assignments.account_id, user_role_assignments.role_name
FROM user_role_assignments
WHERE user_role_assignments.user_id = sqlc.arg(user_id)
	AND user_role_assignments.archived_at IS NULL;
