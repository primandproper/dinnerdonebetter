-- name: CreateValidPreparationVessel :exec
INSERT INTO valid_preparation_vessels (
	id,
	notes,
	valid_preparation_id,
	valid_vessel_id
) VALUES (
	sqlc.arg(id),
	sqlc.arg(notes),
	sqlc.arg(valid_preparation_id),
	sqlc.arg(valid_vessel_id)
);

-- name: CheckValidPreparationVesselExistence :one
SELECT EXISTS (
	SELECT valid_preparation_vessels.id
	FROM valid_preparation_vessels
	WHERE valid_preparation_vessels.archived_at IS NULL
		AND valid_preparation_vessels.id = sqlc.arg(id)
);

-- name: UpdateValidPreparationVessel :execrows
UPDATE valid_preparation_vessels SET
	notes = sqlc.arg(notes),
	valid_preparation_id = sqlc.arg(valid_preparation_id),
	valid_vessel_id = sqlc.arg(valid_vessel_id),
	last_updated_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveValidPreparationVessel :execrows
UPDATE valid_preparation_vessels SET
	archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: GetValidPreparationVesselsForPreparation :many
SELECT
	valid_preparation_vessels.id as valid_preparation_vessel_id,
	valid_preparation_vessels.notes as valid_preparation_vessel_notes,
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
	valid_preparation_vessels.created_at as valid_preparation_vessel_created_at,
	valid_preparation_vessels.last_updated_at as valid_preparation_vessel_last_updated_at,
	valid_preparation_vessels.archived_at as valid_preparation_vessel_archived_at,
	valid_vessels.id as valid_vessel_id,
	valid_vessels.name as valid_vessel_name,
	valid_vessels.plural_name as valid_vessel_plural_name,
	valid_vessels.description as valid_vessel_description,
	valid_vessels.icon_path as valid_vessel_icon_path,
	valid_vessels.usable_for_storage as valid_vessel_usable_for_storage,
	valid_vessels.slug as valid_vessel_slug,
	valid_vessels.display_in_summary_lists as valid_vessel_display_in_summary_lists,
	valid_vessels.include_in_generated_instructions as valid_vessel_include_in_generated_instructions,
	valid_vessels.capacity as valid_vessel_capacity,
	valid_measurement_units.id as valid_measurement_unit_id,
	valid_measurement_units.name as valid_measurement_unit_name,
	valid_measurement_units.description as valid_measurement_unit_description,
	valid_measurement_units.volumetric as valid_measurement_unit_volumetric,
	valid_measurement_units.icon_path as valid_measurement_unit_icon_path,
	valid_measurement_units.universal as valid_measurement_unit_universal,
	valid_measurement_units.metric as valid_measurement_unit_metric,
	valid_measurement_units.imperial as valid_measurement_unit_imperial,
	valid_measurement_units.slug as valid_measurement_unit_slug,
	valid_measurement_units.plural_name as valid_measurement_unit_plural_name,
	valid_measurement_units.last_indexed_at as valid_measurement_unit_last_indexed_at,
	valid_measurement_units.created_at as valid_measurement_unit_created_at,
	valid_measurement_units.last_updated_at as valid_measurement_unit_last_updated_at,
	valid_measurement_units.archived_at as valid_measurement_unit_archived_at,
	valid_vessels.width_in_millimeters as valid_vessel_width_in_millimeters,
	valid_vessels.length_in_millimeters as valid_vessel_length_in_millimeters,
	valid_vessels.height_in_millimeters as valid_vessel_height_in_millimeters,
	valid_vessels.shape as valid_vessel_shape,
	valid_vessels.last_indexed_at as valid_vessel_last_indexed_at,
	valid_vessels.created_at as valid_vessel_created_at,
	valid_vessels.last_updated_at as valid_vessel_last_updated_at,
	valid_vessels.archived_at as valid_vessel_archived_at,
	(
		SELECT COUNT(valid_preparation_vessels.id)
		FROM valid_preparation_vessels
		WHERE valid_preparation_vessels.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND valid_preparation_vessels.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				valid_preparation_vessels.last_updated_at IS NULL
				OR valid_preparation_vessels.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				valid_preparation_vessels.last_updated_at IS NULL
				OR valid_preparation_vessels.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_vessels.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_preparation_vessels.id)
		FROM valid_preparation_vessels
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_vessels.archived_at IS NULL)
	) AS total_count
FROM valid_preparation_vessels
	JOIN valid_vessels ON valid_preparation_vessels.valid_vessel_id = valid_vessels.id
	JOIN valid_preparations ON valid_preparation_vessels.valid_preparation_id = valid_preparations.id
	LEFT JOIN valid_measurement_units ON valid_vessels.capacity_unit = valid_measurement_units.id
WHERE valid_preparation_vessels.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND valid_preparation_vessels.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		valid_preparation_vessels.last_updated_at IS NULL
		OR valid_preparation_vessels.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		valid_preparation_vessels.last_updated_at IS NULL
		OR valid_preparation_vessels.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_vessels.archived_at IS NULL)
	AND valid_vessels.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_measurement_units.archived_at IS NULL
	AND valid_preparation_vessels.valid_preparation_id = sqlc.arg(id)
	AND valid_preparation_vessels.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY valid_preparation_vessels.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidPreparationVesselsForVessel :many
SELECT
	valid_preparation_vessels.id as valid_preparation_vessel_id,
	valid_preparation_vessels.notes as valid_preparation_vessel_notes,
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
	valid_preparation_vessels.created_at as valid_preparation_vessel_created_at,
	valid_preparation_vessels.last_updated_at as valid_preparation_vessel_last_updated_at,
	valid_preparation_vessels.archived_at as valid_preparation_vessel_archived_at,
	valid_vessels.id as valid_vessel_id,
	valid_vessels.name as valid_vessel_name,
	valid_vessels.plural_name as valid_vessel_plural_name,
	valid_vessels.description as valid_vessel_description,
	valid_vessels.icon_path as valid_vessel_icon_path,
	valid_vessels.usable_for_storage as valid_vessel_usable_for_storage,
	valid_vessels.slug as valid_vessel_slug,
	valid_vessels.display_in_summary_lists as valid_vessel_display_in_summary_lists,
	valid_vessels.include_in_generated_instructions as valid_vessel_include_in_generated_instructions,
	valid_vessels.capacity as valid_vessel_capacity,
	valid_measurement_units.id as valid_measurement_unit_id,
	valid_measurement_units.name as valid_measurement_unit_name,
	valid_measurement_units.description as valid_measurement_unit_description,
	valid_measurement_units.volumetric as valid_measurement_unit_volumetric,
	valid_measurement_units.icon_path as valid_measurement_unit_icon_path,
	valid_measurement_units.universal as valid_measurement_unit_universal,
	valid_measurement_units.metric as valid_measurement_unit_metric,
	valid_measurement_units.imperial as valid_measurement_unit_imperial,
	valid_measurement_units.slug as valid_measurement_unit_slug,
	valid_measurement_units.plural_name as valid_measurement_unit_plural_name,
	valid_measurement_units.last_indexed_at as valid_measurement_unit_last_indexed_at,
	valid_measurement_units.created_at as valid_measurement_unit_created_at,
	valid_measurement_units.last_updated_at as valid_measurement_unit_last_updated_at,
	valid_measurement_units.archived_at as valid_measurement_unit_archived_at,
	valid_vessels.width_in_millimeters as valid_vessel_width_in_millimeters,
	valid_vessels.length_in_millimeters as valid_vessel_length_in_millimeters,
	valid_vessels.height_in_millimeters as valid_vessel_height_in_millimeters,
	valid_vessels.shape as valid_vessel_shape,
	valid_vessels.last_indexed_at as valid_vessel_last_indexed_at,
	valid_vessels.created_at as valid_vessel_created_at,
	valid_vessels.last_updated_at as valid_vessel_last_updated_at,
	valid_vessels.archived_at as valid_vessel_archived_at,
	(
		SELECT COUNT(valid_preparation_vessels.id)
		FROM valid_preparation_vessels
		WHERE valid_preparation_vessels.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND valid_preparation_vessels.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				valid_preparation_vessels.last_updated_at IS NULL
				OR valid_preparation_vessels.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				valid_preparation_vessels.last_updated_at IS NULL
				OR valid_preparation_vessels.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_vessels.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_preparation_vessels.id)
		FROM valid_preparation_vessels
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_vessels.archived_at IS NULL)
	) AS total_count
FROM valid_preparation_vessels
	JOIN valid_vessels ON valid_preparation_vessels.valid_vessel_id = valid_vessels.id
	JOIN valid_preparations ON valid_preparation_vessels.valid_preparation_id = valid_preparations.id
	LEFT JOIN valid_measurement_units ON valid_vessels.capacity_unit = valid_measurement_units.id
WHERE valid_preparation_vessels.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND valid_preparation_vessels.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		valid_preparation_vessels.last_updated_at IS NULL
		OR valid_preparation_vessels.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		valid_preparation_vessels.last_updated_at IS NULL
		OR valid_preparation_vessels.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_vessels.archived_at IS NULL)
	AND valid_vessels.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_measurement_units.archived_at IS NULL
	AND valid_preparation_vessels.valid_vessel_id = sqlc.arg(id)
	AND valid_preparation_vessels.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY valid_preparation_vessels.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidPreparationVessels :many
SELECT
	valid_preparation_vessels.id as valid_preparation_vessel_id,
	valid_preparation_vessels.notes as valid_preparation_vessel_notes,
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
	valid_preparation_vessels.created_at as valid_preparation_vessel_created_at,
	valid_preparation_vessels.last_updated_at as valid_preparation_vessel_last_updated_at,
	valid_preparation_vessels.archived_at as valid_preparation_vessel_archived_at,
	valid_vessels.id as valid_vessel_id,
	valid_vessels.name as valid_vessel_name,
	valid_vessels.plural_name as valid_vessel_plural_name,
	valid_vessels.description as valid_vessel_description,
	valid_vessels.icon_path as valid_vessel_icon_path,
	valid_vessels.usable_for_storage as valid_vessel_usable_for_storage,
	valid_vessels.slug as valid_vessel_slug,
	valid_vessels.display_in_summary_lists as valid_vessel_display_in_summary_lists,
	valid_vessels.include_in_generated_instructions as valid_vessel_include_in_generated_instructions,
	valid_vessels.capacity as valid_vessel_capacity,
	valid_measurement_units.id as valid_measurement_unit_id,
	valid_measurement_units.name as valid_measurement_unit_name,
	valid_measurement_units.description as valid_measurement_unit_description,
	valid_measurement_units.volumetric as valid_measurement_unit_volumetric,
	valid_measurement_units.icon_path as valid_measurement_unit_icon_path,
	valid_measurement_units.universal as valid_measurement_unit_universal,
	valid_measurement_units.metric as valid_measurement_unit_metric,
	valid_measurement_units.imperial as valid_measurement_unit_imperial,
	valid_measurement_units.slug as valid_measurement_unit_slug,
	valid_measurement_units.plural_name as valid_measurement_unit_plural_name,
	valid_measurement_units.last_indexed_at as valid_measurement_unit_last_indexed_at,
	valid_measurement_units.created_at as valid_measurement_unit_created_at,
	valid_measurement_units.last_updated_at as valid_measurement_unit_last_updated_at,
	valid_measurement_units.archived_at as valid_measurement_unit_archived_at,
	valid_vessels.width_in_millimeters as valid_vessel_width_in_millimeters,
	valid_vessels.length_in_millimeters as valid_vessel_length_in_millimeters,
	valid_vessels.height_in_millimeters as valid_vessel_height_in_millimeters,
	valid_vessels.shape as valid_vessel_shape,
	valid_vessels.last_indexed_at as valid_vessel_last_indexed_at,
	valid_vessels.created_at as valid_vessel_created_at,
	valid_vessels.last_updated_at as valid_vessel_last_updated_at,
	valid_vessels.archived_at as valid_vessel_archived_at,
	(
		SELECT COUNT(valid_preparation_vessels.id)
		FROM valid_preparation_vessels
		WHERE valid_preparation_vessels.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND valid_preparation_vessels.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				valid_preparation_vessels.last_updated_at IS NULL
				OR valid_preparation_vessels.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				valid_preparation_vessels.last_updated_at IS NULL
				OR valid_preparation_vessels.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_vessels.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_preparation_vessels.id)
		FROM valid_preparation_vessels
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_vessels.archived_at IS NULL)
	) AS total_count
FROM valid_preparation_vessels
	JOIN valid_vessels ON valid_preparation_vessels.valid_vessel_id = valid_vessels.id
	JOIN valid_preparations ON valid_preparation_vessels.valid_preparation_id = valid_preparations.id
	LEFT JOIN valid_measurement_units ON valid_vessels.capacity_unit = valid_measurement_units.id
WHERE valid_preparation_vessels.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND valid_preparation_vessels.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		valid_preparation_vessels.last_updated_at IS NULL
		OR valid_preparation_vessels.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		valid_preparation_vessels.last_updated_at IS NULL
		OR valid_preparation_vessels.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparation_vessels.archived_at IS NULL)
	AND valid_vessels.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_measurement_units.archived_at IS NULL
	AND valid_preparation_vessels.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY valid_preparation_vessels.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidPreparationVessel :one
SELECT
	valid_preparation_vessels.id as valid_preparation_vessel_id,
	valid_preparation_vessels.notes as valid_preparation_vessel_notes,
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
	valid_preparation_vessels.created_at as valid_preparation_vessel_created_at,
	valid_preparation_vessels.last_updated_at as valid_preparation_vessel_last_updated_at,
	valid_preparation_vessels.archived_at as valid_preparation_vessel_archived_at,
	valid_vessels.id as valid_vessel_id,
	valid_vessels.name as valid_vessel_name,
	valid_vessels.plural_name as valid_vessel_plural_name,
	valid_vessels.description as valid_vessel_description,
	valid_vessels.icon_path as valid_vessel_icon_path,
	valid_vessels.usable_for_storage as valid_vessel_usable_for_storage,
	valid_vessels.slug as valid_vessel_slug,
	valid_vessels.display_in_summary_lists as valid_vessel_display_in_summary_lists,
	valid_vessels.include_in_generated_instructions as valid_vessel_include_in_generated_instructions,
	valid_vessels.capacity as valid_vessel_capacity,
	valid_measurement_units.id as valid_measurement_unit_id,
	valid_measurement_units.name as valid_measurement_unit_name,
	valid_measurement_units.description as valid_measurement_unit_description,
	valid_measurement_units.volumetric as valid_measurement_unit_volumetric,
	valid_measurement_units.icon_path as valid_measurement_unit_icon_path,
	valid_measurement_units.universal as valid_measurement_unit_universal,
	valid_measurement_units.metric as valid_measurement_unit_metric,
	valid_measurement_units.imperial as valid_measurement_unit_imperial,
	valid_measurement_units.slug as valid_measurement_unit_slug,
	valid_measurement_units.plural_name as valid_measurement_unit_plural_name,
	valid_measurement_units.last_indexed_at as valid_measurement_unit_last_indexed_at,
	valid_measurement_units.created_at as valid_measurement_unit_created_at,
	valid_measurement_units.last_updated_at as valid_measurement_unit_last_updated_at,
	valid_measurement_units.archived_at as valid_measurement_unit_archived_at,
	valid_vessels.width_in_millimeters as valid_vessel_width_in_millimeters,
	valid_vessels.length_in_millimeters as valid_vessel_length_in_millimeters,
	valid_vessels.height_in_millimeters as valid_vessel_height_in_millimeters,
	valid_vessels.shape as valid_vessel_shape,
	valid_vessels.last_indexed_at as valid_vessel_last_indexed_at,
	valid_vessels.created_at as valid_vessel_created_at,
	valid_vessels.last_updated_at as valid_vessel_last_updated_at,
	valid_vessels.archived_at as valid_vessel_archived_at
FROM valid_preparation_vessels
	JOIN valid_vessels ON valid_preparation_vessels.valid_vessel_id = valid_vessels.id
	JOIN valid_preparations ON valid_preparation_vessels.valid_preparation_id = valid_preparations.id
	LEFT JOIN valid_measurement_units ON valid_vessels.capacity_unit = valid_measurement_units.id
WHERE
	valid_preparation_vessels.archived_at IS NULL
	AND valid_vessels.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_measurement_units.archived_at IS NULL
	AND valid_preparation_vessels.id = sqlc.arg(id);

-- name: GetValidPreparationVesselsByIDs :many
SELECT
	valid_preparation_vessels.id as valid_preparation_vessel_id,
	valid_preparation_vessels.notes as valid_preparation_vessel_notes,
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
	valid_preparation_vessels.created_at as valid_preparation_vessel_created_at,
	valid_preparation_vessels.last_updated_at as valid_preparation_vessel_last_updated_at,
	valid_preparation_vessels.archived_at as valid_preparation_vessel_archived_at,
	valid_vessels.id as valid_vessel_id,
	valid_vessels.name as valid_vessel_name,
	valid_vessels.plural_name as valid_vessel_plural_name,
	valid_vessels.description as valid_vessel_description,
	valid_vessels.icon_path as valid_vessel_icon_path,
	valid_vessels.usable_for_storage as valid_vessel_usable_for_storage,
	valid_vessels.slug as valid_vessel_slug,
	valid_vessels.display_in_summary_lists as valid_vessel_display_in_summary_lists,
	valid_vessels.include_in_generated_instructions as valid_vessel_include_in_generated_instructions,
	valid_vessels.capacity as valid_vessel_capacity,
	valid_measurement_units.id as valid_measurement_unit_id,
	valid_measurement_units.name as valid_measurement_unit_name,
	valid_measurement_units.description as valid_measurement_unit_description,
	valid_measurement_units.volumetric as valid_measurement_unit_volumetric,
	valid_measurement_units.icon_path as valid_measurement_unit_icon_path,
	valid_measurement_units.universal as valid_measurement_unit_universal,
	valid_measurement_units.metric as valid_measurement_unit_metric,
	valid_measurement_units.imperial as valid_measurement_unit_imperial,
	valid_measurement_units.slug as valid_measurement_unit_slug,
	valid_measurement_units.plural_name as valid_measurement_unit_plural_name,
	valid_measurement_units.last_indexed_at as valid_measurement_unit_last_indexed_at,
	valid_measurement_units.created_at as valid_measurement_unit_created_at,
	valid_measurement_units.last_updated_at as valid_measurement_unit_last_updated_at,
	valid_measurement_units.archived_at as valid_measurement_unit_archived_at,
	valid_vessels.width_in_millimeters as valid_vessel_width_in_millimeters,
	valid_vessels.length_in_millimeters as valid_vessel_length_in_millimeters,
	valid_vessels.height_in_millimeters as valid_vessel_height_in_millimeters,
	valid_vessels.shape as valid_vessel_shape,
	valid_vessels.last_indexed_at as valid_vessel_last_indexed_at,
	valid_vessels.created_at as valid_vessel_created_at,
	valid_vessels.last_updated_at as valid_vessel_last_updated_at,
	valid_vessels.archived_at as valid_vessel_archived_at
FROM valid_preparation_vessels
	JOIN valid_vessels ON valid_preparation_vessels.valid_vessel_id = valid_vessels.id
	JOIN valid_preparations ON valid_preparation_vessels.valid_preparation_id = valid_preparations.id
	LEFT JOIN valid_measurement_units ON valid_vessels.capacity_unit = valid_measurement_units.id
WHERE
	valid_preparation_vessels.archived_at IS NULL
	AND valid_vessels.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_measurement_units.archived_at IS NULL
	AND valid_preparation_vessels.id = ANY(sqlc.arg(ids)::text[]);

-- name: ValidPreparationVesselPairIsValid :one
SELECT EXISTS(
	SELECT id
	FROM valid_preparation_vessels
	WHERE valid_vessel_id = sqlc.arg(valid_vessel_id)
	AND valid_preparation_id = sqlc.arg(valid_preparation_id)
	AND archived_at IS NULL
);
