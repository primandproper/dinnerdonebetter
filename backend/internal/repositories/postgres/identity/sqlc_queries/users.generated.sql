-- name: CreateUser :exec
INSERT INTO users (
	id,
	username,
	email_address,
	hashed_password,
	requires_password_change,
	two_factor_secret,
	two_factor_secret_verified_at,
	user_account_status,
	user_account_status_explanation,
	birthday,
	email_address_verification_token,
	first_name,
	last_name
) VALUES (
	sqlc.arg(id),
	sqlc.arg(username),
	sqlc.arg(email_address),
	sqlc.arg(hashed_password),
	sqlc.arg(requires_password_change),
	sqlc.arg(two_factor_secret),
	sqlc.arg(two_factor_secret_verified_at),
	sqlc.arg(user_account_status),
	sqlc.arg(user_account_status_explanation),
	sqlc.arg(birthday),
	sqlc.arg(email_address_verification_token),
	sqlc.arg(first_name),
	sqlc.arg(last_name)
);

-- name: ArchiveUser :execrows
UPDATE users SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ScanUserIDsForReindex :many
SELECT users.id
FROM users
WHERE users.archived_at IS NULL
	AND users.id COLLATE "C" > sqlc.arg(cursor)
ORDER BY users.id COLLATE "C"
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: MarkUsersAsIndexed :execrows
UPDATE users SET
	last_indexed_at = NOW()
WHERE id = ANY(sqlc.arg(ids)::text[]);

-- name: AcceptPrivacyPolicyForUser :exec
UPDATE users SET
	last_accepted_privacy_policy = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: AcceptTermsOfServiceForUser :exec
UPDATE users SET
	last_accepted_terms_of_service = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: DeleteUser :execrows
DELETE FROM users WHERE id = sqlc.arg(id);

-- name: GetAdminUserByUsername :one
SELECT
	users.id,
	users.username,
	users.email_address,
	users.hashed_password,
	users.password_last_changed_at,
	users.requires_password_change,
	users.two_factor_secret,
	users.two_factor_secret_verified_at,
	users.user_account_status,
	users.user_account_status_explanation,
	users.birthday,
	users.email_address_verification_token,
	users.email_address_verified_at,
	users.first_name,
	users.last_name,
	users.last_accepted_terms_of_service,
	users.last_accepted_privacy_policy,
	users.last_indexed_at,
	users.created_at,
	users.last_updated_at,
	users.archived_at,
	uploaded_media.id as avatar_id,
	uploaded_media.storage_path as avatar_storage_path,
	uploaded_media.mime_type as avatar_mime_type,
	uploaded_media.created_at as avatar_created_at,
	uploaded_media.last_updated_at as avatar_last_updated_at,
	uploaded_media.archived_at as avatar_archived_at,
	uploaded_media.created_by_user as avatar_created_by_user
FROM users
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
	LEFT JOIN uploaded_media ON uploaded_media.id = user_avatars.uploaded_media_id AND uploaded_media.archived_at IS NULL
	JOIN user_role_assignments ON user_role_assignments.user_id = users.id AND user_role_assignments.account_id IS NULL AND user_role_assignments.archived_at IS NULL
	JOIN user_roles ON user_roles.id = user_role_assignments.role_id AND user_roles.archived_at IS NULL
WHERE users.archived_at IS NULL
	AND user_roles.name = 'service_admin'
	AND users.username = sqlc.arg(username)
	AND users.two_factor_secret_verified_at IS NOT NULL;

-- name: GetUserByEmail :one
SELECT
	users.id,
	users.username,
	users.email_address,
	users.hashed_password,
	users.password_last_changed_at,
	users.requires_password_change,
	users.two_factor_secret,
	users.two_factor_secret_verified_at,
	users.user_account_status,
	users.user_account_status_explanation,
	users.birthday,
	users.email_address_verification_token,
	users.email_address_verified_at,
	users.first_name,
	users.last_name,
	users.last_accepted_terms_of_service,
	users.last_accepted_privacy_policy,
	users.last_indexed_at,
	users.created_at,
	users.last_updated_at,
	users.archived_at,
	uploaded_media.id as avatar_id,
	uploaded_media.storage_path as avatar_storage_path,
	uploaded_media.mime_type as avatar_mime_type,
	uploaded_media.created_at as avatar_created_at,
	uploaded_media.last_updated_at as avatar_last_updated_at,
	uploaded_media.archived_at as avatar_archived_at,
	uploaded_media.created_by_user as avatar_created_by_user
FROM users
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
	LEFT JOIN uploaded_media ON uploaded_media.id = user_avatars.uploaded_media_id AND uploaded_media.archived_at IS NULL
WHERE users.archived_at IS NULL
	AND users.email_address = sqlc.arg(email_address);

-- name: GetUserByEmailAddressVerificationToken :one
SELECT
	users.id,
	users.username,
	users.email_address,
	users.hashed_password,
	users.password_last_changed_at,
	users.requires_password_change,
	users.two_factor_secret,
	users.two_factor_secret_verified_at,
	users.user_account_status,
	users.user_account_status_explanation,
	users.birthday,
	users.email_address_verification_token,
	users.email_address_verified_at,
	users.first_name,
	users.last_name,
	users.last_accepted_terms_of_service,
	users.last_accepted_privacy_policy,
	users.last_indexed_at,
	users.created_at,
	users.last_updated_at,
	users.archived_at,
	uploaded_media.id as avatar_id,
	uploaded_media.storage_path as avatar_storage_path,
	uploaded_media.mime_type as avatar_mime_type,
	uploaded_media.created_at as avatar_created_at,
	uploaded_media.last_updated_at as avatar_last_updated_at,
	uploaded_media.archived_at as avatar_archived_at,
	uploaded_media.created_by_user as avatar_created_by_user
FROM users
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
	LEFT JOIN uploaded_media ON uploaded_media.id = user_avatars.uploaded_media_id AND uploaded_media.archived_at IS NULL
WHERE users.archived_at IS NULL
	AND users.email_address_verification_token = sqlc.arg(email_address_verification_token);

-- name: GetUserByID :one
SELECT
	users.id,
	users.username,
	users.email_address,
	users.hashed_password,
	users.password_last_changed_at,
	users.requires_password_change,
	users.two_factor_secret,
	users.two_factor_secret_verified_at,
	users.user_account_status,
	users.user_account_status_explanation,
	users.birthday,
	users.email_address_verification_token,
	users.email_address_verified_at,
	users.first_name,
	users.last_name,
	users.last_accepted_terms_of_service,
	users.last_accepted_privacy_policy,
	users.last_indexed_at,
	users.created_at,
	users.last_updated_at,
	users.archived_at,
	uploaded_media.id as avatar_id,
	uploaded_media.storage_path as avatar_storage_path,
	uploaded_media.mime_type as avatar_mime_type,
	uploaded_media.created_at as avatar_created_at,
	uploaded_media.last_updated_at as avatar_last_updated_at,
	uploaded_media.archived_at as avatar_archived_at,
	uploaded_media.created_by_user as avatar_created_by_user
FROM users
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
	LEFT JOIN uploaded_media ON uploaded_media.id = user_avatars.uploaded_media_id AND uploaded_media.archived_at IS NULL
WHERE users.archived_at IS NULL
	AND users.id = sqlc.arg(id);

-- name: GetUserByUsername :one
SELECT
	users.id,
	users.username,
	users.email_address,
	users.hashed_password,
	users.password_last_changed_at,
	users.requires_password_change,
	users.two_factor_secret,
	users.two_factor_secret_verified_at,
	users.user_account_status,
	users.user_account_status_explanation,
	users.birthday,
	users.email_address_verification_token,
	users.email_address_verified_at,
	users.first_name,
	users.last_name,
	users.last_accepted_terms_of_service,
	users.last_accepted_privacy_policy,
	users.last_indexed_at,
	users.created_at,
	users.last_updated_at,
	users.archived_at,
	uploaded_media.id as avatar_id,
	uploaded_media.storage_path as avatar_storage_path,
	uploaded_media.mime_type as avatar_mime_type,
	uploaded_media.created_at as avatar_created_at,
	uploaded_media.last_updated_at as avatar_last_updated_at,
	uploaded_media.archived_at as avatar_archived_at,
	uploaded_media.created_by_user as avatar_created_by_user
FROM users
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
	LEFT JOIN uploaded_media ON uploaded_media.id = user_avatars.uploaded_media_id AND uploaded_media.archived_at IS NULL
WHERE users.archived_at IS NULL
	AND users.username = sqlc.arg(username);

-- name: GetEmailVerificationTokenByUserID :one
SELECT
	users.email_address_verification_token
FROM users
WHERE users.archived_at IS NULL
	AND users.email_address_verified_at IS NULL
	AND users.id = sqlc.arg(id);

-- name: GetUsers :many
SELECT
	users.id,
	users.username,
	users.email_address,
	users.hashed_password,
	users.password_last_changed_at,
	users.requires_password_change,
	users.two_factor_secret,
	users.two_factor_secret_verified_at,
	users.user_account_status,
	users.user_account_status_explanation,
	users.birthday,
	users.email_address_verification_token,
	users.email_address_verified_at,
	users.first_name,
	users.last_name,
	users.last_accepted_terms_of_service,
	users.last_accepted_privacy_policy,
	users.last_indexed_at,
	users.created_at,
	users.last_updated_at,
	users.archived_at,
	uploaded_media.id as avatar_id,
	uploaded_media.storage_path as avatar_storage_path,
	uploaded_media.mime_type as avatar_mime_type,
	uploaded_media.created_at as avatar_created_at,
	uploaded_media.last_updated_at as avatar_last_updated_at,
	uploaded_media.archived_at as avatar_archived_at,
	uploaded_media.created_by_user as avatar_created_by_user,
	(
		SELECT COUNT(users.id)
		FROM users
		WHERE users.archived_at IS NULL
			AND
			users.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND users.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				users.last_updated_at IS NULL
				OR users.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				users.last_updated_at IS NULL
				OR users.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR users.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(users.id)
		FROM users
		WHERE users.archived_at IS NULL
	) AS total_count
FROM users
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
	LEFT JOIN uploaded_media ON uploaded_media.id = user_avatars.uploaded_media_id AND uploaded_media.archived_at IS NULL
WHERE users.archived_at IS NULL
	AND users.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND users.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		users.last_updated_at IS NULL
		OR users.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		users.last_updated_at IS NULL
		OR users.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR users.archived_at IS NULL)
	AND users.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY users.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetUsersForAccount :many
SELECT
	users.id,
	users.username,
	users.email_address,
	users.hashed_password,
	users.password_last_changed_at,
	users.requires_password_change,
	users.two_factor_secret,
	users.two_factor_secret_verified_at,
	users.user_account_status,
	users.user_account_status_explanation,
	users.birthday,
	users.email_address_verification_token,
	users.email_address_verified_at,
	users.first_name,
	users.last_name,
	users.last_accepted_terms_of_service,
	users.last_accepted_privacy_policy,
	users.last_indexed_at,
	users.created_at,
	users.last_updated_at,
	users.archived_at,
	uploaded_media.id as avatar_id,
	uploaded_media.storage_path as avatar_storage_path,
	uploaded_media.mime_type as avatar_mime_type,
	uploaded_media.created_at as avatar_created_at,
	uploaded_media.last_updated_at as avatar_last_updated_at,
	uploaded_media.archived_at as avatar_archived_at,
	uploaded_media.created_by_user as avatar_created_by_user,
	(
		SELECT COUNT(users.id)
		FROM users
		WHERE users.archived_at IS NULL
			AND
			users.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND users.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				users.last_updated_at IS NULL
				OR users.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				users.last_updated_at IS NULL
				OR users.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR users.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(users.id)
		FROM users
		WHERE users.archived_at IS NULL
			AND account_user_memberships.belongs_to_account = sqlc.arg(belongs_to_account)
	) AS total_count
FROM users
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
	LEFT JOIN uploaded_media ON uploaded_media.id = user_avatars.uploaded_media_id AND uploaded_media.archived_at IS NULL
JOIN account_user_memberships ON account_user_memberships.belongs_to_user = users.id
WHERE users.archived_at IS NULL
	AND users.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND users.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		users.last_updated_at IS NULL
		OR users.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		users.last_updated_at IS NULL
		OR users.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR users.archived_at IS NULL)
	AND account_user_memberships.belongs_to_account = sqlc.arg(belongs_to_account)
	AND account_user_memberships.archived_at IS NULL
	AND users.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY users.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetUsersWithIDs :many
SELECT
	users.id,
	users.username,
	users.email_address,
	users.hashed_password,
	users.password_last_changed_at,
	users.requires_password_change,
	users.two_factor_secret,
	users.two_factor_secret_verified_at,
	users.user_account_status,
	users.user_account_status_explanation,
	users.birthday,
	users.email_address_verification_token,
	users.email_address_verified_at,
	users.first_name,
	users.last_name,
	users.last_accepted_terms_of_service,
	users.last_accepted_privacy_policy,
	users.last_indexed_at,
	users.created_at,
	users.last_updated_at,
	users.archived_at,
	uploaded_media.id as avatar_id,
	uploaded_media.storage_path as avatar_storage_path,
	uploaded_media.mime_type as avatar_mime_type,
	uploaded_media.created_at as avatar_created_at,
	uploaded_media.last_updated_at as avatar_last_updated_at,
	uploaded_media.archived_at as avatar_archived_at,
	uploaded_media.created_by_user as avatar_created_by_user
FROM users
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
	LEFT JOIN uploaded_media ON uploaded_media.id = user_avatars.uploaded_media_id AND uploaded_media.archived_at IS NULL
WHERE users.archived_at IS NULL
	AND users.id = ANY(sqlc.arg(ids)::text[]);

-- name: GetUserRequiresPasswordChange :one
SELECT users.requires_password_change
FROM users
WHERE users.archived_at IS NULL
	AND users.id = sqlc.arg(id);

-- name: GetUserIDsNeedingIndexing :many
SELECT users.id
FROM users
WHERE users.archived_at IS NULL
	AND users.last_indexed_at IS NULL
	OR users.last_indexed_at < NOW() - '24 hours'::INTERVAL;

-- name: GetUserWithUnverifiedTwoFactor :one
SELECT
	users.id,
	users.username,
	users.email_address,
	users.hashed_password,
	users.password_last_changed_at,
	users.requires_password_change,
	users.two_factor_secret,
	users.two_factor_secret_verified_at,
	users.user_account_status,
	users.user_account_status_explanation,
	users.birthday,
	users.email_address_verification_token,
	users.email_address_verified_at,
	users.first_name,
	users.last_name,
	users.last_accepted_terms_of_service,
	users.last_accepted_privacy_policy,
	users.last_indexed_at,
	users.created_at,
	users.last_updated_at,
	users.archived_at,
	uploaded_media.id as avatar_id,
	uploaded_media.storage_path as avatar_storage_path,
	uploaded_media.mime_type as avatar_mime_type,
	uploaded_media.created_at as avatar_created_at,
	uploaded_media.last_updated_at as avatar_last_updated_at,
	uploaded_media.archived_at as avatar_archived_at,
	uploaded_media.created_by_user as avatar_created_by_user
FROM users
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
	LEFT JOIN uploaded_media ON uploaded_media.id = user_avatars.uploaded_media_id AND uploaded_media.archived_at IS NULL
WHERE users.archived_at IS NULL
	AND users.id = sqlc.arg(id)
	AND users.two_factor_secret_verified_at IS NULL;

-- name: GetUserWithVerifiedTwoFactor :one
SELECT
	users.id,
	users.username,
	users.email_address,
	users.hashed_password,
	users.password_last_changed_at,
	users.requires_password_change,
	users.two_factor_secret,
	users.two_factor_secret_verified_at,
	users.user_account_status,
	users.user_account_status_explanation,
	users.birthday,
	users.email_address_verification_token,
	users.email_address_verified_at,
	users.first_name,
	users.last_name,
	users.last_accepted_terms_of_service,
	users.last_accepted_privacy_policy,
	users.last_indexed_at,
	users.created_at,
	users.last_updated_at,
	users.archived_at,
	uploaded_media.id as avatar_id,
	uploaded_media.storage_path as avatar_storage_path,
	uploaded_media.mime_type as avatar_mime_type,
	uploaded_media.created_at as avatar_created_at,
	uploaded_media.last_updated_at as avatar_last_updated_at,
	uploaded_media.archived_at as avatar_archived_at,
	uploaded_media.created_by_user as avatar_created_by_user
FROM users
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
	LEFT JOIN uploaded_media ON uploaded_media.id = user_avatars.uploaded_media_id AND uploaded_media.archived_at IS NULL
WHERE users.archived_at IS NULL
	AND users.id = sqlc.arg(id)
	AND users.two_factor_secret_verified_at IS NOT NULL;

-- name: MarkEmailAddressAsVerified :exec
UPDATE users SET
	email_address_verified_at = NOW(),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND email_address_verified_at IS NULL
	AND id = sqlc.arg(id)
	AND email_address_verification_token = sqlc.arg(email_address_verification_token);

-- name: MarkEmailAddressAsUnverified :exec
UPDATE users SET
	email_address_verified_at = NULL,
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND email_address_verified_at IS NOT NULL
	AND id = sqlc.arg(id);

-- name: MarkTwoFactorSecretAsUnverified :exec
UPDATE users SET
	two_factor_secret_verified_at = NULL,
	two_factor_secret = sqlc.arg(two_factor_secret),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: MarkTwoFactorSecretAsVerified :exec
UPDATE users SET
	two_factor_secret_verified_at = NOW(),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: SearchUsersByUsername :many
SELECT
	users.id,
	users.username,
	users.email_address,
	users.hashed_password,
	users.password_last_changed_at,
	users.requires_password_change,
	users.two_factor_secret,
	users.two_factor_secret_verified_at,
	users.user_account_status,
	users.user_account_status_explanation,
	users.birthday,
	users.email_address_verification_token,
	users.email_address_verified_at,
	users.first_name,
	users.last_name,
	users.last_accepted_terms_of_service,
	users.last_accepted_privacy_policy,
	users.last_indexed_at,
	users.created_at,
	users.last_updated_at,
	users.archived_at,
	uploaded_media.id as avatar_id,
	uploaded_media.storage_path as avatar_storage_path,
	uploaded_media.mime_type as avatar_mime_type,
	uploaded_media.created_at as avatar_created_at,
	uploaded_media.last_updated_at as avatar_last_updated_at,
	uploaded_media.archived_at as avatar_archived_at,
	uploaded_media.created_by_user as avatar_created_by_user,
	(
		SELECT COUNT(users.id)
		FROM users
		WHERE users.archived_at IS NULL
			AND
			users.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND users.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				users.last_updated_at IS NULL
				OR users.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				users.last_updated_at IS NULL
				OR users.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR users.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(users.id)
		FROM users
		WHERE users.archived_at IS NULL
	) AS total_count
FROM users
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
	LEFT JOIN uploaded_media ON uploaded_media.id = user_avatars.uploaded_media_id AND uploaded_media.archived_at IS NULL
WHERE users.archived_at IS NULL
	AND users.username ILIKE '%' || sqlc.arg(username)::text || '%'
	AND users.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND users.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		users.last_updated_at IS NULL
		OR users.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		users.last_updated_at IS NULL
		OR users.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR users.archived_at IS NULL)
	AND users.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY users.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: UpdateUserDetails :execrows
UPDATE users SET
	first_name = sqlc.arg(first_name),
	last_name = sqlc.arg(last_name),
	birthday = sqlc.arg(birthday),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: UpdateUserEmailAddress :execrows
UPDATE users SET
	email_address = sqlc.arg(email_address),
	email_address_verified_at = NULL,
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: UpdateUserPassword :execrows
UPDATE users SET
	hashed_password = sqlc.arg(hashed_password),
	requires_password_change = FALSE,
	password_last_changed_at = NOW(),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: UpdateUserTwoFactorSecret :execrows
UPDATE users SET
	two_factor_secret_verified_at = NULL,
	two_factor_secret = sqlc.arg(two_factor_secret),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: UpdateUserUsername :execrows
UPDATE users SET
	username = sqlc.arg(username),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);
