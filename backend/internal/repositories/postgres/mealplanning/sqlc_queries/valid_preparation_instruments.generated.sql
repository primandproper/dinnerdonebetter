-- name: ArchiveValidPreparationInstrument :execrows
UPDATE valid_preparation_instruments SET archived_at = NOW() WHERE archived_at IS NULL AND id = sqlc.arg(id);

-- name: CreateValidPreparationInstrument :exec
INSERT INTO valid_preparation_instruments (
	id,
	notes,
	valid_preparation_id,
	valid_instrument_id
) VALUES (
	sqlc.arg(id),
	sqlc.arg(notes),
	sqlc.arg(valid_preparation_id),
	sqlc.arg(valid_instrument_id)
);

-- name: CheckValidPreparationInstrumentExistence :one
SELECT EXISTS (
	SELECT valid_preparation_instruments.id
	FROM valid_preparation_instruments
	WHERE valid_preparation_instruments.archived_at IS NULL
		AND valid_preparation_instruments.id = sqlc.arg(id)
);

-- name: GetValidPreparationInstrumentsForInstrument :many
SELECT
	valid_preparation_instruments.id as valid_preparation_instrument_id,
	valid_preparation_instruments.notes as valid_preparation_instrument_notes,
	valid_preparations.id as valid_preparation_id,
	valid_preparations.name as valid_preparation_name,
	valid_preparations.description as valid_preparation_description,
	valid_preparations.icon_path as valid_preparation_icon_path,
	valid_preparations.yields_nothing as valid_preparation_yields_nothing,
	valid_preparations.restrict_to_ingredients as valid_preparation_restrict_to_ingredients,
	valid_preparations.past_tense as valid_preparation_past_tense,
	valid_preparations.slug as valid_preparation_slug,
	valid_preparations.minimum_ingredient_count as valid_preparation_minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count as valid_preparation_maximum_ingredient_count,
	valid_preparations.minimum_instrument_count as valid_preparation_minimum_instrument_count,
	valid_preparations.maximum_instrument_count as valid_preparation_maximum_instrument_count,
	valid_preparations.temperature_required as valid_preparation_temperature_required,
	valid_preparations.time_estimate_required as valid_preparation_time_estimate_required,
	valid_preparations.condition_expression_required as valid_preparation_condition_expression_required,
	valid_preparations.consumes_vessel as valid_preparation_consumes_vessel,
	valid_preparations.only_for_vessels as valid_preparation_only_for_vessels,
	valid_preparations.minimum_vessel_count as valid_preparation_minimum_vessel_count,
	valid_preparations.maximum_vessel_count as valid_preparation_maximum_vessel_count,
	valid_preparations.last_indexed_at as valid_preparation_last_indexed_at,
	valid_preparations.created_at as valid_preparation_created_at,
	valid_preparations.last_updated_at as valid_preparation_last_updated_at,
	valid_preparations.archived_at as valid_preparation_archived_at,
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
	valid_preparation_instruments.created_at as valid_preparation_instrument_created_at,
	valid_preparation_instruments.last_updated_at as valid_preparation_instrument_last_updated_at,
	valid_preparation_instruments.archived_at as valid_preparation_instrument_archived_at,
	(
		SELECT COUNT(valid_preparation_instruments.id)
		FROM valid_preparation_instruments
		JOIN valid_instruments ON valid_preparation_instruments.valid_instrument_id = valid_instruments.id
	JOIN valid_preparations ON valid_preparation_instruments.valid_preparation_id = valid_preparations.id
		WHERE valid_preparation_instruments.archived_at IS NULL
			AND
			valid_preparation_instruments.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_preparation_instruments.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_preparation_instruments.last_updated_at IS NULL
				OR valid_preparation_instruments.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_preparation_instruments.last_updated_at IS NULL
				OR valid_preparation_instruments.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_instruments.archived_at IS NULL)
			AND valid_instruments.archived_at IS NULL
			AND valid_preparations.archived_at IS NULL
			AND valid_preparation_instruments.valid_instrument_id = sqlc.arg(id)
	) AS filtered_count,
	(
		SELECT COUNT(valid_preparation_instruments.id)
		FROM valid_preparation_instruments
		JOIN valid_instruments ON valid_preparation_instruments.valid_instrument_id = valid_instruments.id
	JOIN valid_preparations ON valid_preparation_instruments.valid_preparation_id = valid_preparations.id
		WHERE valid_preparation_instruments.archived_at IS NULL
			AND valid_instruments.archived_at IS NULL
			AND valid_preparations.archived_at IS NULL
			AND valid_preparation_instruments.valid_instrument_id = sqlc.arg(id)
	) AS total_count
FROM valid_preparation_instruments
	JOIN valid_instruments ON valid_preparation_instruments.valid_instrument_id = valid_instruments.id
	JOIN valid_preparations ON valid_preparation_instruments.valid_preparation_id = valid_preparations.id
WHERE
	valid_preparation_instruments.archived_at IS NULL
	AND valid_preparation_instruments.valid_instrument_id = sqlc.arg(id)
	AND valid_instruments.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_preparation_instruments.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_preparation_instruments.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_preparation_instruments.last_updated_at IS NULL
		OR valid_preparation_instruments.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_preparation_instruments.last_updated_at IS NULL
		OR valid_preparation_instruments.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_instruments.archived_at IS NULL)
	AND valid_preparation_instruments.id > COALESCE(sqlc.narg(cursor), '')
GROUP BY
	valid_preparation_instruments.id,
	valid_preparations.id,
	valid_instruments.id
ORDER BY valid_preparation_instruments.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidPreparationInstrumentsForPreparation :many
SELECT
	valid_preparation_instruments.id as valid_preparation_instrument_id,
	valid_preparation_instruments.notes as valid_preparation_instrument_notes,
	valid_preparations.id as valid_preparation_id,
	valid_preparations.name as valid_preparation_name,
	valid_preparations.description as valid_preparation_description,
	valid_preparations.icon_path as valid_preparation_icon_path,
	valid_preparations.yields_nothing as valid_preparation_yields_nothing,
	valid_preparations.restrict_to_ingredients as valid_preparation_restrict_to_ingredients,
	valid_preparations.past_tense as valid_preparation_past_tense,
	valid_preparations.slug as valid_preparation_slug,
	valid_preparations.minimum_ingredient_count as valid_preparation_minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count as valid_preparation_maximum_ingredient_count,
	valid_preparations.minimum_instrument_count as valid_preparation_minimum_instrument_count,
	valid_preparations.maximum_instrument_count as valid_preparation_maximum_instrument_count,
	valid_preparations.temperature_required as valid_preparation_temperature_required,
	valid_preparations.time_estimate_required as valid_preparation_time_estimate_required,
	valid_preparations.condition_expression_required as valid_preparation_condition_expression_required,
	valid_preparations.consumes_vessel as valid_preparation_consumes_vessel,
	valid_preparations.only_for_vessels as valid_preparation_only_for_vessels,
	valid_preparations.minimum_vessel_count as valid_preparation_minimum_vessel_count,
	valid_preparations.maximum_vessel_count as valid_preparation_maximum_vessel_count,
	valid_preparations.last_indexed_at as valid_preparation_last_indexed_at,
	valid_preparations.created_at as valid_preparation_created_at,
	valid_preparations.last_updated_at as valid_preparation_last_updated_at,
	valid_preparations.archived_at as valid_preparation_archived_at,
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
	valid_preparation_instruments.created_at as valid_preparation_instrument_created_at,
	valid_preparation_instruments.last_updated_at as valid_preparation_instrument_last_updated_at,
	valid_preparation_instruments.archived_at as valid_preparation_instrument_archived_at,
	(
		SELECT COUNT(valid_preparation_instruments.id)
		FROM valid_preparation_instruments
		JOIN valid_instruments ON valid_preparation_instruments.valid_instrument_id = valid_instruments.id
	JOIN valid_preparations ON valid_preparation_instruments.valid_preparation_id = valid_preparations.id
		WHERE valid_preparation_instruments.archived_at IS NULL
			AND
			valid_preparation_instruments.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_preparation_instruments.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_preparation_instruments.last_updated_at IS NULL
				OR valid_preparation_instruments.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_preparation_instruments.last_updated_at IS NULL
				OR valid_preparation_instruments.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_instruments.archived_at IS NULL)
			AND valid_instruments.archived_at IS NULL
			AND valid_preparations.archived_at IS NULL
			AND valid_preparation_instruments.valid_preparation_id = sqlc.arg(id)
	) AS filtered_count,
	(
		SELECT COUNT(valid_preparation_instruments.id)
		FROM valid_preparation_instruments
		JOIN valid_instruments ON valid_preparation_instruments.valid_instrument_id = valid_instruments.id
	JOIN valid_preparations ON valid_preparation_instruments.valid_preparation_id = valid_preparations.id
		WHERE valid_preparation_instruments.archived_at IS NULL
			AND valid_instruments.archived_at IS NULL
			AND valid_preparations.archived_at IS NULL
			AND valid_preparation_instruments.valid_preparation_id = sqlc.arg(id)
	) AS total_count
FROM valid_preparation_instruments
	JOIN valid_instruments ON valid_preparation_instruments.valid_instrument_id = valid_instruments.id
	JOIN valid_preparations ON valid_preparation_instruments.valid_preparation_id = valid_preparations.id
WHERE
	valid_preparation_instruments.archived_at IS NULL
	AND valid_preparation_instruments.valid_preparation_id = sqlc.arg(id)
	AND valid_instruments.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_preparation_instruments.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_preparation_instruments.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_preparation_instruments.last_updated_at IS NULL
		OR valid_preparation_instruments.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_preparation_instruments.last_updated_at IS NULL
		OR valid_preparation_instruments.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_instruments.archived_at IS NULL)
	AND valid_preparation_instruments.id > COALESCE(sqlc.narg(cursor), '')
GROUP BY
	valid_preparation_instruments.id,
	valid_preparations.id,
	valid_instruments.id
ORDER BY valid_preparation_instruments.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidPreparationInstruments :many
SELECT
	valid_preparation_instruments.id as valid_preparation_instrument_id,
	valid_preparation_instruments.notes as valid_preparation_instrument_notes,
	valid_preparations.id as valid_preparation_id,
	valid_preparations.name as valid_preparation_name,
	valid_preparations.description as valid_preparation_description,
	valid_preparations.icon_path as valid_preparation_icon_path,
	valid_preparations.yields_nothing as valid_preparation_yields_nothing,
	valid_preparations.restrict_to_ingredients as valid_preparation_restrict_to_ingredients,
	valid_preparations.past_tense as valid_preparation_past_tense,
	valid_preparations.slug as valid_preparation_slug,
	valid_preparations.minimum_ingredient_count as valid_preparation_minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count as valid_preparation_maximum_ingredient_count,
	valid_preparations.minimum_instrument_count as valid_preparation_minimum_instrument_count,
	valid_preparations.maximum_instrument_count as valid_preparation_maximum_instrument_count,
	valid_preparations.temperature_required as valid_preparation_temperature_required,
	valid_preparations.time_estimate_required as valid_preparation_time_estimate_required,
	valid_preparations.condition_expression_required as valid_preparation_condition_expression_required,
	valid_preparations.consumes_vessel as valid_preparation_consumes_vessel,
	valid_preparations.only_for_vessels as valid_preparation_only_for_vessels,
	valid_preparations.minimum_vessel_count as valid_preparation_minimum_vessel_count,
	valid_preparations.maximum_vessel_count as valid_preparation_maximum_vessel_count,
	valid_preparations.last_indexed_at as valid_preparation_last_indexed_at,
	valid_preparations.created_at as valid_preparation_created_at,
	valid_preparations.last_updated_at as valid_preparation_last_updated_at,
	valid_preparations.archived_at as valid_preparation_archived_at,
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
	valid_preparation_instruments.created_at as valid_preparation_instrument_created_at,
	valid_preparation_instruments.last_updated_at as valid_preparation_instrument_last_updated_at,
	valid_preparation_instruments.archived_at as valid_preparation_instrument_archived_at,
	(
		SELECT COUNT(valid_preparation_instruments.id)
		FROM valid_preparation_instruments
		JOIN valid_instruments ON valid_preparation_instruments.valid_instrument_id = valid_instruments.id
	JOIN valid_preparations ON valid_preparation_instruments.valid_preparation_id = valid_preparations.id
		WHERE valid_preparation_instruments.archived_at IS NULL
			AND
			valid_preparation_instruments.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_preparation_instruments.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_preparation_instruments.last_updated_at IS NULL
				OR valid_preparation_instruments.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_preparation_instruments.last_updated_at IS NULL
				OR valid_preparation_instruments.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_instruments.archived_at IS NULL)
			AND valid_instruments.archived_at IS NULL
			AND valid_preparations.archived_at IS NULL
	) AS filtered_count,
	(
		SELECT COUNT(valid_preparation_instruments.id)
		FROM valid_preparation_instruments
		JOIN valid_instruments ON valid_preparation_instruments.valid_instrument_id = valid_instruments.id
	JOIN valid_preparations ON valid_preparation_instruments.valid_preparation_id = valid_preparations.id
		WHERE valid_preparation_instruments.archived_at IS NULL
			AND valid_instruments.archived_at IS NULL
			AND valid_preparations.archived_at IS NULL
	) AS total_count
FROM valid_preparation_instruments
	JOIN valid_instruments ON valid_preparation_instruments.valid_instrument_id = valid_instruments.id
	JOIN valid_preparations ON valid_preparation_instruments.valid_preparation_id = valid_preparations.id
WHERE
	valid_preparation_instruments.archived_at IS NULL
	AND valid_instruments.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_preparation_instruments.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_preparation_instruments.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_preparation_instruments.last_updated_at IS NULL
		OR valid_preparation_instruments.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_preparation_instruments.last_updated_at IS NULL
		OR valid_preparation_instruments.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_instruments.archived_at IS NULL)
	AND valid_preparation_instruments.id > COALESCE(sqlc.narg(cursor), '')
GROUP BY
	valid_preparation_instruments.id,
	valid_preparations.id,
	valid_instruments.id
ORDER BY valid_preparation_instruments.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidPreparationInstrument :one
SELECT
	valid_preparation_instruments.id as valid_preparation_instrument_id,
	valid_preparation_instruments.notes as valid_preparation_instrument_notes,
	valid_preparations.id as valid_preparation_id,
	valid_preparations.name as valid_preparation_name,
	valid_preparations.description as valid_preparation_description,
	valid_preparations.icon_path as valid_preparation_icon_path,
	valid_preparations.yields_nothing as valid_preparation_yields_nothing,
	valid_preparations.restrict_to_ingredients as valid_preparation_restrict_to_ingredients,
	valid_preparations.past_tense as valid_preparation_past_tense,
	valid_preparations.slug as valid_preparation_slug,
	valid_preparations.minimum_ingredient_count as valid_preparation_minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count as valid_preparation_maximum_ingredient_count,
	valid_preparations.minimum_instrument_count as valid_preparation_minimum_instrument_count,
	valid_preparations.maximum_instrument_count as valid_preparation_maximum_instrument_count,
	valid_preparations.temperature_required as valid_preparation_temperature_required,
	valid_preparations.time_estimate_required as valid_preparation_time_estimate_required,
	valid_preparations.condition_expression_required as valid_preparation_condition_expression_required,
	valid_preparations.consumes_vessel as valid_preparation_consumes_vessel,
	valid_preparations.only_for_vessels as valid_preparation_only_for_vessels,
	valid_preparations.minimum_vessel_count as valid_preparation_minimum_vessel_count,
	valid_preparations.maximum_vessel_count as valid_preparation_maximum_vessel_count,
	valid_preparations.last_indexed_at as valid_preparation_last_indexed_at,
	valid_preparations.created_at as valid_preparation_created_at,
	valid_preparations.last_updated_at as valid_preparation_last_updated_at,
	valid_preparations.archived_at as valid_preparation_archived_at,
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
	valid_preparation_instruments.created_at as valid_preparation_instrument_created_at,
	valid_preparation_instruments.last_updated_at as valid_preparation_instrument_last_updated_at,
	valid_preparation_instruments.archived_at as valid_preparation_instrument_archived_at
FROM
	valid_preparation_instruments
	JOIN valid_instruments ON valid_preparation_instruments.valid_instrument_id = valid_instruments.id
	JOIN valid_preparations ON valid_preparation_instruments.valid_preparation_id = valid_preparations.id
WHERE
	valid_preparation_instruments.archived_at IS NULL
	AND valid_instruments.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_preparation_instruments.id = sqlc.arg(id);

-- name: GetValidPreparationInstrumentsByIDs :many
SELECT
	valid_preparation_instruments.id as valid_preparation_instrument_id,
	valid_preparation_instruments.notes as valid_preparation_instrument_notes,
	valid_preparations.id as valid_preparation_id,
	valid_preparations.name as valid_preparation_name,
	valid_preparations.description as valid_preparation_description,
	valid_preparations.icon_path as valid_preparation_icon_path,
	valid_preparations.yields_nothing as valid_preparation_yields_nothing,
	valid_preparations.restrict_to_ingredients as valid_preparation_restrict_to_ingredients,
	valid_preparations.past_tense as valid_preparation_past_tense,
	valid_preparations.slug as valid_preparation_slug,
	valid_preparations.minimum_ingredient_count as valid_preparation_minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count as valid_preparation_maximum_ingredient_count,
	valid_preparations.minimum_instrument_count as valid_preparation_minimum_instrument_count,
	valid_preparations.maximum_instrument_count as valid_preparation_maximum_instrument_count,
	valid_preparations.temperature_required as valid_preparation_temperature_required,
	valid_preparations.time_estimate_required as valid_preparation_time_estimate_required,
	valid_preparations.condition_expression_required as valid_preparation_condition_expression_required,
	valid_preparations.consumes_vessel as valid_preparation_consumes_vessel,
	valid_preparations.only_for_vessels as valid_preparation_only_for_vessels,
	valid_preparations.minimum_vessel_count as valid_preparation_minimum_vessel_count,
	valid_preparations.maximum_vessel_count as valid_preparation_maximum_vessel_count,
	valid_preparations.last_indexed_at as valid_preparation_last_indexed_at,
	valid_preparations.created_at as valid_preparation_created_at,
	valid_preparations.last_updated_at as valid_preparation_last_updated_at,
	valid_preparations.archived_at as valid_preparation_archived_at,
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
	valid_preparation_instruments.created_at as valid_preparation_instrument_created_at,
	valid_preparation_instruments.last_updated_at as valid_preparation_instrument_last_updated_at,
	valid_preparation_instruments.archived_at as valid_preparation_instrument_archived_at
FROM
	valid_preparation_instruments
	JOIN valid_instruments ON valid_preparation_instruments.valid_instrument_id = valid_instruments.id
	JOIN valid_preparations ON valid_preparation_instruments.valid_preparation_id = valid_preparations.id
WHERE
	valid_preparation_instruments.archived_at IS NULL
	AND valid_instruments.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_preparation_instruments.id = ANY(sqlc.arg(ids)::text[]);

-- name: ValidPreparationInstrumentPairIsValid :one
SELECT EXISTS(
	SELECT valid_preparation_instruments.id
	FROM valid_preparation_instruments
	WHERE valid_instrument_id = sqlc.arg(valid_instrument_id)
	AND valid_preparation_id = sqlc.arg(valid_preparation_id)
	AND archived_at IS NULL
);

-- name: UpdateValidPreparationInstrument :execrows
UPDATE valid_preparation_instruments SET
	notes = sqlc.arg(notes),
	valid_preparation_id = sqlc.arg(valid_preparation_id),
	valid_instrument_id = sqlc.arg(valid_instrument_id),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);
