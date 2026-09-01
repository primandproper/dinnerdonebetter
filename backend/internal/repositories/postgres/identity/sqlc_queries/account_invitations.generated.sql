-- name: CreateAccountInvitation :exec
INSERT INTO account_invitations (
	id,
	from_user,
	to_user,
	to_name,
	note,
	to_email,
	token,
	destination_account,
	expires_at
) VALUES (
	sqlc.arg(id),
	sqlc.arg(from_user),
	sqlc.arg(to_user),
	sqlc.arg(to_name),
	sqlc.arg(note),
	sqlc.arg(to_email),
	sqlc.arg(token),
	sqlc.arg(destination_account),
	sqlc.arg(expires_at)
);

-- name: CheckAccountInvitationExistence :one
SELECT EXISTS (
	SELECT account_invitations.id
	FROM account_invitations
	WHERE account_invitations.archived_at IS NULL
		AND account_invitations.id = sqlc.arg(id)
);

-- name: AttachAccountInvitationsToUserID :execrows
UPDATE account_invitations SET
	to_user = sqlc.arg(to_user),
	last_updated_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND to_email = LOWER(sqlc.arg(to_email));

-- name: AssignInvitationsToUserByEmail :execrows
UPDATE account_invitations SET
	to_user = sqlc.arg(to_user),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND to_email = LOWER(sqlc.arg(email_address));

-- name: GetAccountInvitationByEmailAndToken :one
SELECT
	account_invitations.id,
	accounts.id as account_id,
	accounts.name as account_name,
	accounts.billing_status as account_billing_status,
	accounts.contact_phone as account_contact_phone,
	accounts.payment_processor_customer_id as account_payment_processor_customer_id,
	accounts.subscription_plan_id as account_subscription_plan_id,
	accounts.belongs_to_user as account_belongs_to_user,
	accounts.time_zone as account_time_zone,
	accounts.address_line_1 as account_address_line_1,
	accounts.address_line_2 as account_address_line_2,
	accounts.city as account_city,
	accounts.state as account_state,
	accounts.zip_code as account_zip_code,
	accounts.country as account_country,
	accounts.latitude as account_latitude,
	accounts.longitude as account_longitude,
	accounts.last_payment_provider_sync_occurred_at as account_last_payment_provider_sync_occurred_at,
	accounts.webhook_hmac_secret as account_webhook_hmac_secret,
	accounts.created_at as account_created_at,
	accounts.last_updated_at as account_last_updated_at,
	accounts.archived_at as account_archived_at,
	account_invitations.from_user,
	account_invitations.to_user,
	users.id as user_id,
	users.username as user_username,
	users.email_address as user_email_address,
	users.hashed_password as user_hashed_password,
	users.password_last_changed_at as user_password_last_changed_at,
	users.requires_password_change as user_requires_password_change,
	users.two_factor_secret as user_two_factor_secret,
	users.two_factor_secret_verified_at as user_two_factor_secret_verified_at,
	users.user_account_status as user_user_account_status,
	users.user_account_status_explanation as user_user_account_status_explanation,
	users.birthday as user_birthday,
	users.email_address_verification_token as user_email_address_verification_token,
	users.email_address_verified_at as user_email_address_verified_at,
	users.first_name as user_first_name,
	users.last_name as user_last_name,
	users.last_accepted_terms_of_service as user_last_accepted_terms_of_service,
	users.last_accepted_privacy_policy as user_last_accepted_privacy_policy,
	users.last_indexed_at as user_last_indexed_at,
	users.created_at as user_created_at,
	users.last_updated_at as user_last_updated_at,
	users.archived_at as user_archived_at,
	user_avatars.uploaded_media_id as user_avatar_id,
	account_invitations.to_name,
	account_invitations.note,
	account_invitations.to_email,
	account_invitations.token,
	account_invitations.destination_account,
	account_invitations.expires_at,
	account_invitations.status,
	account_invitations.status_note,
	account_invitations.created_at,
	account_invitations.last_updated_at,
	account_invitations.archived_at
FROM account_invitations
	JOIN accounts ON account_invitations.destination_account = accounts.id
	JOIN users ON account_invitations.from_user = users.id
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
WHERE account_invitations.archived_at IS NULL
	AND account_invitations.expires_at > CURRENT_TIMESTAMP
	AND account_invitations.to_email = LOWER(sqlc.arg(to_email))
	AND account_invitations.token = sqlc.arg(token);

-- name: GetAccountInvitationByAccountAndID :one
SELECT
	account_invitations.id,
	accounts.id as account_id,
	accounts.name as account_name,
	accounts.billing_status as account_billing_status,
	accounts.contact_phone as account_contact_phone,
	accounts.payment_processor_customer_id as account_payment_processor_customer_id,
	accounts.subscription_plan_id as account_subscription_plan_id,
	accounts.belongs_to_user as account_belongs_to_user,
	accounts.time_zone as account_time_zone,
	accounts.address_line_1 as account_address_line_1,
	accounts.address_line_2 as account_address_line_2,
	accounts.city as account_city,
	accounts.state as account_state,
	accounts.zip_code as account_zip_code,
	accounts.country as account_country,
	accounts.latitude as account_latitude,
	accounts.longitude as account_longitude,
	accounts.last_payment_provider_sync_occurred_at as account_last_payment_provider_sync_occurred_at,
	accounts.webhook_hmac_secret as account_webhook_hmac_secret,
	accounts.created_at as account_created_at,
	accounts.last_updated_at as account_last_updated_at,
	accounts.archived_at as account_archived_at,
	account_invitations.from_user,
	account_invitations.to_user,
	users.id as user_id,
	users.username as user_username,
	users.email_address as user_email_address,
	users.hashed_password as user_hashed_password,
	users.password_last_changed_at as user_password_last_changed_at,
	users.requires_password_change as user_requires_password_change,
	users.two_factor_secret as user_two_factor_secret,
	users.two_factor_secret_verified_at as user_two_factor_secret_verified_at,
	users.user_account_status as user_user_account_status,
	users.user_account_status_explanation as user_user_account_status_explanation,
	users.birthday as user_birthday,
	users.email_address_verification_token as user_email_address_verification_token,
	users.email_address_verified_at as user_email_address_verified_at,
	users.first_name as user_first_name,
	users.last_name as user_last_name,
	users.last_accepted_terms_of_service as user_last_accepted_terms_of_service,
	users.last_accepted_privacy_policy as user_last_accepted_privacy_policy,
	users.last_indexed_at as user_last_indexed_at,
	users.created_at as user_created_at,
	users.last_updated_at as user_last_updated_at,
	users.archived_at as user_archived_at,
	user_avatars.uploaded_media_id as user_avatar_id,
	account_invitations.to_name,
	account_invitations.note,
	account_invitations.to_email,
	account_invitations.token,
	account_invitations.destination_account,
	account_invitations.expires_at,
	account_invitations.status,
	account_invitations.status_note,
	account_invitations.created_at,
	account_invitations.last_updated_at,
	account_invitations.archived_at
FROM account_invitations
	JOIN accounts ON account_invitations.destination_account = accounts.id
	JOIN users ON account_invitations.from_user = users.id
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
WHERE account_invitations.archived_at IS NULL
	AND account_invitations.expires_at > CURRENT_TIMESTAMP
	AND account_invitations.destination_account = sqlc.arg(destination_account)
	AND account_invitations.id = sqlc.arg(id);

-- name: GetAccountInvitationByTokenAndID :one
SELECT
	account_invitations.id,
	accounts.id as account_id,
	accounts.name as account_name,
	accounts.billing_status as account_billing_status,
	accounts.contact_phone as account_contact_phone,
	accounts.payment_processor_customer_id as account_payment_processor_customer_id,
	accounts.subscription_plan_id as account_subscription_plan_id,
	accounts.belongs_to_user as account_belongs_to_user,
	accounts.time_zone as account_time_zone,
	accounts.address_line_1 as account_address_line_1,
	accounts.address_line_2 as account_address_line_2,
	accounts.city as account_city,
	accounts.state as account_state,
	accounts.zip_code as account_zip_code,
	accounts.country as account_country,
	accounts.latitude as account_latitude,
	accounts.longitude as account_longitude,
	accounts.last_payment_provider_sync_occurred_at as account_last_payment_provider_sync_occurred_at,
	accounts.webhook_hmac_secret as account_webhook_hmac_secret,
	accounts.created_at as account_created_at,
	accounts.last_updated_at as account_last_updated_at,
	accounts.archived_at as account_archived_at,
	account_invitations.from_user,
	account_invitations.to_user,
	users.id as user_id,
	users.username as user_username,
	users.email_address as user_email_address,
	users.hashed_password as user_hashed_password,
	users.password_last_changed_at as user_password_last_changed_at,
	users.requires_password_change as user_requires_password_change,
	users.two_factor_secret as user_two_factor_secret,
	users.two_factor_secret_verified_at as user_two_factor_secret_verified_at,
	users.user_account_status as user_user_account_status,
	users.user_account_status_explanation as user_user_account_status_explanation,
	users.birthday as user_birthday,
	users.email_address_verification_token as user_email_address_verification_token,
	users.email_address_verified_at as user_email_address_verified_at,
	users.first_name as user_first_name,
	users.last_name as user_last_name,
	users.last_accepted_terms_of_service as user_last_accepted_terms_of_service,
	users.last_accepted_privacy_policy as user_last_accepted_privacy_policy,
	users.last_indexed_at as user_last_indexed_at,
	users.created_at as user_created_at,
	users.last_updated_at as user_last_updated_at,
	users.archived_at as user_archived_at,
	user_avatars.uploaded_media_id as user_avatar_id,
	account_invitations.to_name,
	account_invitations.note,
	account_invitations.to_email,
	account_invitations.token,
	account_invitations.destination_account,
	account_invitations.expires_at,
	account_invitations.status,
	account_invitations.status_note,
	account_invitations.created_at,
	account_invitations.last_updated_at,
	account_invitations.archived_at
FROM account_invitations
	JOIN accounts ON account_invitations.destination_account = accounts.id
	JOIN users ON account_invitations.from_user = users.id
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
WHERE account_invitations.archived_at IS NULL
	AND account_invitations.expires_at > CURRENT_TIMESTAMP
	AND account_invitations.token = sqlc.arg(token)
	AND account_invitations.id = sqlc.arg(id);

-- name: GetAccountInvitationByToken :one
SELECT
	account_invitations.id,
	accounts.id as account_id,
	accounts.name as account_name,
	accounts.billing_status as account_billing_status,
	accounts.contact_phone as account_contact_phone,
	accounts.payment_processor_customer_id as account_payment_processor_customer_id,
	accounts.subscription_plan_id as account_subscription_plan_id,
	accounts.belongs_to_user as account_belongs_to_user,
	accounts.time_zone as account_time_zone,
	accounts.address_line_1 as account_address_line_1,
	accounts.address_line_2 as account_address_line_2,
	accounts.city as account_city,
	accounts.state as account_state,
	accounts.zip_code as account_zip_code,
	accounts.country as account_country,
	accounts.latitude as account_latitude,
	accounts.longitude as account_longitude,
	accounts.last_payment_provider_sync_occurred_at as account_last_payment_provider_sync_occurred_at,
	accounts.webhook_hmac_secret as account_webhook_hmac_secret,
	accounts.created_at as account_created_at,
	accounts.last_updated_at as account_last_updated_at,
	accounts.archived_at as account_archived_at,
	account_invitations.from_user,
	account_invitations.to_user,
	users.id as user_id,
	users.username as user_username,
	users.email_address as user_email_address,
	users.hashed_password as user_hashed_password,
	users.password_last_changed_at as user_password_last_changed_at,
	users.requires_password_change as user_requires_password_change,
	users.two_factor_secret as user_two_factor_secret,
	users.two_factor_secret_verified_at as user_two_factor_secret_verified_at,
	users.user_account_status as user_user_account_status,
	users.user_account_status_explanation as user_user_account_status_explanation,
	users.birthday as user_birthday,
	users.email_address_verification_token as user_email_address_verification_token,
	users.email_address_verified_at as user_email_address_verified_at,
	users.first_name as user_first_name,
	users.last_name as user_last_name,
	users.last_accepted_terms_of_service as user_last_accepted_terms_of_service,
	users.last_accepted_privacy_policy as user_last_accepted_privacy_policy,
	users.last_indexed_at as user_last_indexed_at,
	users.created_at as user_created_at,
	users.last_updated_at as user_last_updated_at,
	users.archived_at as user_archived_at,
	user_avatars.uploaded_media_id as user_avatar_id,
	account_invitations.to_name,
	account_invitations.note,
	account_invitations.to_email,
	account_invitations.token,
	account_invitations.destination_account,
	account_invitations.expires_at,
	account_invitations.status,
	account_invitations.status_note,
	account_invitations.created_at,
	account_invitations.last_updated_at,
	account_invitations.archived_at
FROM account_invitations
	JOIN accounts ON account_invitations.destination_account = accounts.id
	JOIN users ON account_invitations.from_user = users.id
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
WHERE account_invitations.archived_at IS NULL
	AND account_invitations.expires_at > CURRENT_TIMESTAMP
	AND account_invitations.token = sqlc.arg(token);

-- name: GetPendingInvitesFromUser :many
SELECT
	account_invitations.id,
	accounts.id as account_id,
	accounts.name as account_name,
	accounts.billing_status as account_billing_status,
	accounts.contact_phone as account_contact_phone,
	accounts.payment_processor_customer_id as account_payment_processor_customer_id,
	accounts.subscription_plan_id as account_subscription_plan_id,
	accounts.belongs_to_user as account_belongs_to_user,
	accounts.time_zone as account_time_zone,
	accounts.address_line_1 as account_address_line_1,
	accounts.address_line_2 as account_address_line_2,
	accounts.city as account_city,
	accounts.state as account_state,
	accounts.zip_code as account_zip_code,
	accounts.country as account_country,
	accounts.latitude as account_latitude,
	accounts.longitude as account_longitude,
	accounts.last_payment_provider_sync_occurred_at as account_last_payment_provider_sync_occurred_at,
	accounts.webhook_hmac_secret as account_webhook_hmac_secret,
	accounts.created_at as account_created_at,
	accounts.last_updated_at as account_last_updated_at,
	accounts.archived_at as account_archived_at,
	account_invitations.from_user,
	account_invitations.to_user,
	users.id as user_id,
	users.username as user_username,
	users.email_address as user_email_address,
	users.hashed_password as user_hashed_password,
	users.password_last_changed_at as user_password_last_changed_at,
	users.requires_password_change as user_requires_password_change,
	users.two_factor_secret as user_two_factor_secret,
	users.two_factor_secret_verified_at as user_two_factor_secret_verified_at,
	users.user_account_status as user_user_account_status,
	users.user_account_status_explanation as user_user_account_status_explanation,
	users.birthday as user_birthday,
	users.email_address_verification_token as user_email_address_verification_token,
	users.email_address_verified_at as user_email_address_verified_at,
	users.first_name as user_first_name,
	users.last_name as user_last_name,
	users.last_accepted_terms_of_service as user_last_accepted_terms_of_service,
	users.last_accepted_privacy_policy as user_last_accepted_privacy_policy,
	users.last_indexed_at as user_last_indexed_at,
	users.created_at as user_created_at,
	users.last_updated_at as user_last_updated_at,
	users.archived_at as user_archived_at,
	user_avatars.uploaded_media_id as user_avatar_id,
	account_invitations.to_name,
	account_invitations.note,
	account_invitations.to_email,
	account_invitations.token,
	account_invitations.destination_account,
	account_invitations.expires_at,
	account_invitations.status,
	account_invitations.status_note,
	account_invitations.created_at,
	account_invitations.last_updated_at,
	account_invitations.archived_at,
	(
		SELECT COUNT(account_invitations.id)
		FROM account_invitations
		WHERE account_invitations.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND account_invitations.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				account_invitations.last_updated_at IS NULL
				OR account_invitations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				account_invitations.last_updated_at IS NULL
				OR account_invitations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR account_invitations.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(account_invitations.id)
		FROM account_invitations
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR account_invitations.archived_at IS NULL)
	) AS total_count
FROM account_invitations
	JOIN accounts ON account_invitations.destination_account = accounts.id
	JOIN users ON account_invitations.from_user = users.id
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
WHERE account_invitations.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND account_invitations.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		account_invitations.last_updated_at IS NULL
		OR account_invitations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		account_invitations.last_updated_at IS NULL
		OR account_invitations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR account_invitations.archived_at IS NULL)
	AND account_invitations.from_user = sqlc.arg(from_user)
	AND account_invitations.status = sqlc.arg(status)
	AND account_invitations.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY account_invitations.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetPendingInvitesForUser :many
SELECT
	account_invitations.id,
	accounts.id as account_id,
	accounts.name as account_name,
	accounts.billing_status as account_billing_status,
	accounts.contact_phone as account_contact_phone,
	accounts.payment_processor_customer_id as account_payment_processor_customer_id,
	accounts.subscription_plan_id as account_subscription_plan_id,
	accounts.belongs_to_user as account_belongs_to_user,
	accounts.time_zone as account_time_zone,
	accounts.address_line_1 as account_address_line_1,
	accounts.address_line_2 as account_address_line_2,
	accounts.city as account_city,
	accounts.state as account_state,
	accounts.zip_code as account_zip_code,
	accounts.country as account_country,
	accounts.latitude as account_latitude,
	accounts.longitude as account_longitude,
	accounts.last_payment_provider_sync_occurred_at as account_last_payment_provider_sync_occurred_at,
	accounts.webhook_hmac_secret as account_webhook_hmac_secret,
	accounts.created_at as account_created_at,
	accounts.last_updated_at as account_last_updated_at,
	accounts.archived_at as account_archived_at,
	account_invitations.from_user,
	account_invitations.to_user,
	users.id as user_id,
	users.username as user_username,
	users.email_address as user_email_address,
	users.hashed_password as user_hashed_password,
	users.password_last_changed_at as user_password_last_changed_at,
	users.requires_password_change as user_requires_password_change,
	users.two_factor_secret as user_two_factor_secret,
	users.two_factor_secret_verified_at as user_two_factor_secret_verified_at,
	users.user_account_status as user_user_account_status,
	users.user_account_status_explanation as user_user_account_status_explanation,
	users.birthday as user_birthday,
	users.email_address_verification_token as user_email_address_verification_token,
	users.email_address_verified_at as user_email_address_verified_at,
	users.first_name as user_first_name,
	users.last_name as user_last_name,
	users.last_accepted_terms_of_service as user_last_accepted_terms_of_service,
	users.last_accepted_privacy_policy as user_last_accepted_privacy_policy,
	users.last_indexed_at as user_last_indexed_at,
	users.created_at as user_created_at,
	users.last_updated_at as user_last_updated_at,
	users.archived_at as user_archived_at,
	user_avatars.uploaded_media_id as user_avatar_id,
	account_invitations.to_name,
	account_invitations.note,
	account_invitations.to_email,
	account_invitations.token,
	account_invitations.destination_account,
	account_invitations.expires_at,
	account_invitations.status,
	account_invitations.status_note,
	account_invitations.created_at,
	account_invitations.last_updated_at,
	account_invitations.archived_at,
	(
		SELECT COUNT(account_invitations.id)
		FROM account_invitations
		WHERE account_invitations.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND account_invitations.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				account_invitations.last_updated_at IS NULL
				OR account_invitations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				account_invitations.last_updated_at IS NULL
				OR account_invitations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR account_invitations.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(account_invitations.id)
		FROM account_invitations
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR account_invitations.archived_at IS NULL)
	) AS total_count
FROM account_invitations
	JOIN accounts ON account_invitations.destination_account = accounts.id
	JOIN users ON account_invitations.from_user = users.id
	LEFT JOIN user_avatars ON user_avatars.belongs_to_user = users.id AND user_avatars.archived_at IS NULL
WHERE account_invitations.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND account_invitations.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		account_invitations.last_updated_at IS NULL
		OR account_invitations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		account_invitations.last_updated_at IS NULL
		OR account_invitations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR account_invitations.archived_at IS NULL)
	AND account_invitations.to_user = sqlc.arg(to_user)
	AND account_invitations.status = sqlc.arg(status)
	AND account_invitations.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY account_invitations.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: SetAccountInvitationStatus :exec
UPDATE account_invitations SET
	status = sqlc.arg(status),
	status_note = sqlc.arg(status_note),
	last_updated_at = CURRENT_TIMESTAMP,
	archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);
