-- name: CreateAccountInstrumentOwnership :exec
INSERT INTO account_instrument_ownerships (
	id,
	notes,
	quantity,
	valid_instrument_id,
	belongs_to_account
) VALUES (
	sqlc.arg(id),
	sqlc.arg(notes),
	sqlc.arg(quantity),
	sqlc.arg(valid_instrument_id),
	sqlc.arg(belongs_to_account)
);

-- name: CheckAccountInstrumentOwnershipExistence :one
SELECT EXISTS (
	SELECT account_instrument_ownerships.id
	FROM account_instrument_ownerships
	WHERE account_instrument_ownerships.archived_at IS NULL
		AND account_instrument_ownerships.id = sqlc.arg(id)
		AND account_instrument_ownerships.belongs_to_account = sqlc.arg(belongs_to_account)
);

-- name: UpdateAccountInstrumentOwnership :execrows
UPDATE account_instrument_ownerships SET
	notes = sqlc.arg(notes),
	quantity = sqlc.arg(quantity),
	valid_instrument_id = sqlc.arg(valid_instrument_id),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_account = sqlc.arg(belongs_to_account);

-- name: ArchiveAccountInstrumentOwnership :execrows
UPDATE account_instrument_ownerships SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_account = sqlc.arg(belongs_to_account);

-- name: GetAccountInstrumentOwnerships :many
SELECT
	account_instrument_ownerships.id,
	account_instrument_ownerships.notes,
	account_instrument_ownerships.quantity,
	valid_instruments.id as valid_instrument_id,
	valid_instruments.name as valid_instrument_name,
	valid_instruments.description as valid_instrument_description,
	valid_instruments.icon_path as valid_instrument_icon_path,
	valid_instruments.plural_name as valid_instrument_plural_name,
	valid_instruments.usable_for_storage as valid_instrument_usable_for_storage,
	valid_instruments.slug as valid_instrument_slug,
	valid_instruments.display_in_summary_lists as valid_instrument_display_in_summary_lists,
	valid_instruments.include_in_generated_instructions as valid_instrument_include_in_generated_instructions,
	valid_instruments.last_indexed_at as valid_instrument_last_indexed_at,
	valid_instruments.created_at as valid_instrument_created_at,
	valid_instruments.last_updated_at as valid_instrument_last_updated_at,
	valid_instruments.archived_at as valid_instrument_archived_at,
	account_instrument_ownerships.belongs_to_account,
	account_instrument_ownerships.created_at,
	account_instrument_ownerships.last_updated_at,
	account_instrument_ownerships.archived_at,
	(
		SELECT COUNT(account_instrument_ownerships.id)
		FROM account_instrument_ownerships
		WHERE account_instrument_ownerships.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND account_instrument_ownerships.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				account_instrument_ownerships.last_updated_at IS NULL
				OR account_instrument_ownerships.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				account_instrument_ownerships.last_updated_at IS NULL
				OR account_instrument_ownerships.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR account_instrument_ownerships.archived_at IS NULL)
			AND account_instrument_ownerships.belongs_to_account = sqlc.arg(account_id)
	) AS filtered_count,
	(
		SELECT COUNT(account_instrument_ownerships.id)
		FROM account_instrument_ownerships
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR account_instrument_ownerships.archived_at IS NULL)
			AND account_instrument_ownerships.belongs_to_account = sqlc.arg(account_id)
	) AS total_count
FROM account_instrument_ownerships
INNER JOIN valid_instruments ON account_instrument_ownerships.valid_instrument_id = valid_instruments.id
WHERE account_instrument_ownerships.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND account_instrument_ownerships.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		account_instrument_ownerships.last_updated_at IS NULL
		OR account_instrument_ownerships.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		account_instrument_ownerships.last_updated_at IS NULL
		OR account_instrument_ownerships.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR account_instrument_ownerships.archived_at IS NULL)
	AND account_instrument_ownerships.belongs_to_account = sqlc.arg(account_id)
	AND account_instrument_ownerships.id > COALESCE(sqlc.narg(cursor), '')
GROUP BY
	account_instrument_ownerships.id,
	valid_instruments.id
ORDER BY account_instrument_ownerships.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetAccountInstrumentOwnership :one
SELECT
	account_instrument_ownerships.id,
	account_instrument_ownerships.notes,
	account_instrument_ownerships.quantity,
	valid_instruments.id as valid_instrument_id,
	valid_instruments.name as valid_instrument_name,
	valid_instruments.description as valid_instrument_description,
	valid_instruments.icon_path as valid_instrument_icon_path,
	valid_instruments.plural_name as valid_instrument_plural_name,
	valid_instruments.usable_for_storage as valid_instrument_usable_for_storage,
	valid_instruments.slug as valid_instrument_slug,
	valid_instruments.display_in_summary_lists as valid_instrument_display_in_summary_lists,
	valid_instruments.include_in_generated_instructions as valid_instrument_include_in_generated_instructions,
	valid_instruments.last_indexed_at as valid_instrument_last_indexed_at,
	valid_instruments.created_at as valid_instrument_created_at,
	valid_instruments.last_updated_at as valid_instrument_last_updated_at,
	valid_instruments.archived_at as valid_instrument_archived_at,
	account_instrument_ownerships.belongs_to_account,
	account_instrument_ownerships.created_at,
	account_instrument_ownerships.last_updated_at,
	account_instrument_ownerships.archived_at
FROM account_instrument_ownerships
INNER JOIN valid_instruments ON account_instrument_ownerships.valid_instrument_id = valid_instruments.id
WHERE account_instrument_ownerships.archived_at IS NULL
	AND account_instrument_ownerships.id = sqlc.arg(id)
	AND account_instrument_ownerships.belongs_to_account = sqlc.arg(belongs_to_account);
