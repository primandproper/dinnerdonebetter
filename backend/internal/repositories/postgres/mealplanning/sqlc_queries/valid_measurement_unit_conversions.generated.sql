-- name: ArchiveValidMeasurementUnitConversion :execrows
UPDATE valid_measurement_unit_conversions SET archived_at = NOW() WHERE archived_at IS NULL AND id = sqlc.arg(id);

-- name: CreateValidMeasurementUnitConversion :exec
INSERT INTO valid_measurement_unit_conversions (
	id,
	from_unit,
	to_unit,
	only_for_ingredient,
	modifier,
	notes
) VALUES (
	sqlc.arg(id),
	sqlc.arg(from_unit),
	sqlc.arg(to_unit),
	sqlc.arg(only_for_ingredient),
	sqlc.arg(modifier),
	sqlc.arg(notes)
);

-- name: CheckValidMeasurementUnitConversionExistence :one
SELECT EXISTS (
	SELECT valid_measurement_unit_conversions.id
	FROM valid_measurement_unit_conversions
	WHERE valid_measurement_unit_conversions.archived_at IS NULL
		AND valid_measurement_unit_conversions.id = sqlc.arg(id)
);

-- name: GetValidMeasurementUnitConversionsForMeasurementUnit :many
SELECT
	valid_measurement_unit_conversions.id as valid_measurement_unit_conversion_id,
	valid_measurement_units_from.id as from_unit_id,
	valid_measurement_units_from.name as from_unit_name,
	valid_measurement_units_from.description as from_unit_description,
	valid_measurement_units_from.volumetric as from_unit_volumetric,
	valid_measurement_units_from.icon_path as from_unit_icon_path,
	valid_measurement_units_from.universal as from_unit_universal,
	valid_measurement_units_from.metric as from_unit_metric,
	valid_measurement_units_from.imperial as from_unit_imperial,
	valid_measurement_units_from.slug as from_unit_slug,
	valid_measurement_units_from.plural_name as from_unit_plural_name,
	valid_measurement_units_from.last_indexed_at as from_unit_last_indexed_at,
	valid_measurement_units_from.created_at as from_unit_created_at,
	valid_measurement_units_from.last_updated_at as from_unit_last_updated_at,
	valid_measurement_units_from.archived_at as from_unit_archived_at,
	valid_measurement_units_to.id as to_unit_id,
	valid_measurement_units_to.name as to_unit_name,
	valid_measurement_units_to.description as to_unit_description,
	valid_measurement_units_to.volumetric as to_unit_volumetric,
	valid_measurement_units_to.icon_path as to_unit_icon_path,
	valid_measurement_units_to.universal as to_unit_universal,
	valid_measurement_units_to.metric as to_unit_metric,
	valid_measurement_units_to.imperial as to_unit_imperial,
	valid_measurement_units_to.slug as to_unit_slug,
	valid_measurement_units_to.plural_name as to_unit_plural_name,
	valid_measurement_units_to.last_indexed_at as to_unit_last_indexed_at,
	valid_measurement_units_to.created_at as to_unit_created_at,
	valid_measurement_units_to.last_updated_at as to_unit_last_updated_at,
	valid_measurement_units_to.archived_at as to_unit_archived_at,
	valid_ingredients.id as valid_ingredient_id,
	valid_ingredients.name as valid_ingredient_name,
	valid_ingredients.description as valid_ingredient_description,
	valid_ingredients.warning as valid_ingredient_warning,
	valid_ingredients.contains_egg as valid_ingredient_contains_egg,
	valid_ingredients.contains_dairy as valid_ingredient_contains_dairy,
	valid_ingredients.contains_peanut as valid_ingredient_contains_peanut,
	valid_ingredients.contains_tree_nut as valid_ingredient_contains_tree_nut,
	valid_ingredients.contains_soy as valid_ingredient_contains_soy,
	valid_ingredients.contains_wheat as valid_ingredient_contains_wheat,
	valid_ingredients.contains_shellfish as valid_ingredient_contains_shellfish,
	valid_ingredients.contains_sesame as valid_ingredient_contains_sesame,
	valid_ingredients.contains_fish as valid_ingredient_contains_fish,
	valid_ingredients.contains_gluten as valid_ingredient_contains_gluten,
	valid_ingredients.animal_flesh as valid_ingredient_animal_flesh,
	valid_ingredients.is_liquid as valid_ingredient_is_liquid,
	valid_ingredients.icon_path as valid_ingredient_icon_path,
	valid_ingredients.animal_derived as valid_ingredient_animal_derived,
	valid_ingredients.plural_name as valid_ingredient_plural_name,
	valid_ingredients.restrict_to_preparations as valid_ingredient_restrict_to_preparations,
	valid_ingredients.contaminates_equipment as valid_ingredient_contaminates_equipment,
	valid_ingredients.minimum_ideal_storage_temperature_in_celsius as valid_ingredient_minimum_ideal_storage_temperature_in_celsius,
	valid_ingredients.maximum_ideal_storage_temperature_in_celsius as valid_ingredient_maximum_ideal_storage_temperature_in_celsius,
	valid_ingredients.storage_instructions as valid_ingredient_storage_instructions,
	valid_ingredients.slug as valid_ingredient_slug,
	valid_ingredients.contains_alcohol as valid_ingredient_contains_alcohol,
	valid_ingredients.shopping_suggestions as valid_ingredient_shopping_suggestions,
	valid_ingredients.is_starch as valid_ingredient_is_starch,
	valid_ingredients.is_protein as valid_ingredient_is_protein,
	valid_ingredients.is_grain as valid_ingredient_is_grain,
	valid_ingredients.is_fruit as valid_ingredient_is_fruit,
	valid_ingredients.is_salt as valid_ingredient_is_salt,
	valid_ingredients.is_fat as valid_ingredient_is_fat,
	valid_ingredients.is_acid as valid_ingredient_is_acid,
	valid_ingredients.is_heat as valid_ingredient_is_heat,
	valid_ingredients.last_indexed_at as valid_ingredient_last_indexed_at,
	valid_ingredients.created_at as valid_ingredient_created_at,
	valid_ingredients.last_updated_at as valid_ingredient_last_updated_at,
	valid_ingredients.archived_at as valid_ingredient_archived_at,
	valid_measurement_unit_conversions.from_unit as valid_measurement_unit_conversion_from_unit,
	valid_measurement_unit_conversions.to_unit as valid_measurement_unit_conversion_to_unit,
	valid_measurement_unit_conversions.only_for_ingredient as valid_measurement_unit_conversion_only_for_ingredient,
	valid_measurement_unit_conversions.modifier as valid_measurement_unit_conversion_modifier,
	valid_measurement_unit_conversions.notes as valid_measurement_unit_conversion_notes,
	valid_measurement_unit_conversions.created_at as valid_measurement_unit_conversion_created_at,
	valid_measurement_unit_conversions.last_updated_at as valid_measurement_unit_conversion_last_updated_at,
	valid_measurement_unit_conversions.archived_at as valid_measurement_unit_conversion_archived_at,
	(
		SELECT COUNT(valid_measurement_unit_conversions.id)
		FROM valid_measurement_unit_conversions
		JOIN valid_measurement_units AS valid_measurement_units_from ON valid_measurement_unit_conversions.from_unit = valid_measurement_units_from.id
	JOIN valid_measurement_units AS valid_measurement_units_to ON valid_measurement_unit_conversions.to_unit = valid_measurement_units_to.id
	JOIN valid_ingredients ON valid_measurement_unit_conversions.only_for_ingredient = valid_ingredients.id
		WHERE valid_measurement_unit_conversions.archived_at IS NULL
			AND
			valid_measurement_unit_conversions.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_measurement_unit_conversions.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_measurement_unit_conversions.last_updated_at IS NULL
				OR valid_measurement_unit_conversions.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_measurement_unit_conversions.last_updated_at IS NULL
				OR valid_measurement_unit_conversions.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_unit_conversions.archived_at IS NULL)
			AND (valid_measurement_units_from.id = sqlc.arg(id) OR valid_measurement_units_to.id = sqlc.arg(id))
			AND valid_measurement_units_from.archived_at IS NULL
			AND valid_measurement_units_to.archived_at IS NULL
	) AS filtered_count,
	(
		SELECT COUNT(valid_measurement_unit_conversions.id)
		FROM valid_measurement_unit_conversions
		JOIN valid_measurement_units AS valid_measurement_units_from ON valid_measurement_unit_conversions.from_unit = valid_measurement_units_from.id
	JOIN valid_measurement_units AS valid_measurement_units_to ON valid_measurement_unit_conversions.to_unit = valid_measurement_units_to.id
	JOIN valid_ingredients ON valid_measurement_unit_conversions.only_for_ingredient = valid_ingredients.id
		WHERE valid_measurement_unit_conversions.archived_at IS NULL
			AND (valid_measurement_units_from.id = sqlc.arg(id) OR valid_measurement_units_to.id = sqlc.arg(id))
			AND valid_measurement_units_from.archived_at IS NULL
			AND valid_measurement_units_to.archived_at IS NULL
	) AS total_count
FROM valid_measurement_unit_conversions
	JOIN valid_measurement_units AS valid_measurement_units_from ON valid_measurement_unit_conversions.from_unit = valid_measurement_units_from.id
	JOIN valid_measurement_units AS valid_measurement_units_to ON valid_measurement_unit_conversions.to_unit = valid_measurement_units_to.id
	LEFT JOIN valid_ingredients ON valid_measurement_unit_conversions.only_for_ingredient = valid_ingredients.id
WHERE
	(valid_measurement_units_from.id = sqlc.arg(id) OR valid_measurement_units_to.id = sqlc.arg(id))
	AND valid_measurement_unit_conversions.archived_at IS NULL
	AND valid_measurement_units_from.archived_at IS NULL
	AND valid_measurement_units_to.archived_at IS NULL
	AND valid_measurement_unit_conversions.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_measurement_unit_conversions.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_measurement_unit_conversions.last_updated_at IS NULL
		OR valid_measurement_unit_conversions.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_measurement_unit_conversions.last_updated_at IS NULL
		OR valid_measurement_unit_conversions.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_unit_conversions.archived_at IS NULL)
	AND valid_measurement_unit_conversions.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY valid_measurement_unit_conversions.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidMeasurementUnitConversion :one
SELECT
	valid_measurement_unit_conversions.id as valid_measurement_unit_conversion_id,
	valid_measurement_units_from.id as from_unit_id,
	valid_measurement_units_from.name as from_unit_name,
	valid_measurement_units_from.description as from_unit_description,
	valid_measurement_units_from.volumetric as from_unit_volumetric,
	valid_measurement_units_from.icon_path as from_unit_icon_path,
	valid_measurement_units_from.universal as from_unit_universal,
	valid_measurement_units_from.metric as from_unit_metric,
	valid_measurement_units_from.imperial as from_unit_imperial,
	valid_measurement_units_from.slug as from_unit_slug,
	valid_measurement_units_from.plural_name as from_unit_plural_name,
	valid_measurement_units_from.last_indexed_at as from_unit_last_indexed_at,
	valid_measurement_units_from.created_at as from_unit_created_at,
	valid_measurement_units_from.last_updated_at as from_unit_last_updated_at,
	valid_measurement_units_from.archived_at as from_unit_archived_at,
	valid_measurement_units_to.id as to_unit_id,
	valid_measurement_units_to.name as to_unit_name,
	valid_measurement_units_to.description as to_unit_description,
	valid_measurement_units_to.volumetric as to_unit_volumetric,
	valid_measurement_units_to.icon_path as to_unit_icon_path,
	valid_measurement_units_to.universal as to_unit_universal,
	valid_measurement_units_to.metric as to_unit_metric,
	valid_measurement_units_to.imperial as to_unit_imperial,
	valid_measurement_units_to.slug as to_unit_slug,
	valid_measurement_units_to.plural_name as to_unit_plural_name,
	valid_measurement_units_to.last_indexed_at as to_unit_last_indexed_at,
	valid_measurement_units_to.created_at as to_unit_created_at,
	valid_measurement_units_to.last_updated_at as to_unit_last_updated_at,
	valid_measurement_units_to.archived_at as to_unit_archived_at,
	valid_ingredients.id as valid_ingredient_id,
	valid_ingredients.name as valid_ingredient_name,
	valid_ingredients.description as valid_ingredient_description,
	valid_ingredients.warning as valid_ingredient_warning,
	valid_ingredients.contains_egg as valid_ingredient_contains_egg,
	valid_ingredients.contains_dairy as valid_ingredient_contains_dairy,
	valid_ingredients.contains_peanut as valid_ingredient_contains_peanut,
	valid_ingredients.contains_tree_nut as valid_ingredient_contains_tree_nut,
	valid_ingredients.contains_soy as valid_ingredient_contains_soy,
	valid_ingredients.contains_wheat as valid_ingredient_contains_wheat,
	valid_ingredients.contains_shellfish as valid_ingredient_contains_shellfish,
	valid_ingredients.contains_sesame as valid_ingredient_contains_sesame,
	valid_ingredients.contains_fish as valid_ingredient_contains_fish,
	valid_ingredients.contains_gluten as valid_ingredient_contains_gluten,
	valid_ingredients.animal_flesh as valid_ingredient_animal_flesh,
	valid_ingredients.is_liquid as valid_ingredient_is_liquid,
	valid_ingredients.icon_path as valid_ingredient_icon_path,
	valid_ingredients.animal_derived as valid_ingredient_animal_derived,
	valid_ingredients.plural_name as valid_ingredient_plural_name,
	valid_ingredients.restrict_to_preparations as valid_ingredient_restrict_to_preparations,
	valid_ingredients.contaminates_equipment as valid_ingredient_contaminates_equipment,
	valid_ingredients.minimum_ideal_storage_temperature_in_celsius as valid_ingredient_minimum_ideal_storage_temperature_in_celsius,
	valid_ingredients.maximum_ideal_storage_temperature_in_celsius as valid_ingredient_maximum_ideal_storage_temperature_in_celsius,
	valid_ingredients.storage_instructions as valid_ingredient_storage_instructions,
	valid_ingredients.slug as valid_ingredient_slug,
	valid_ingredients.contains_alcohol as valid_ingredient_contains_alcohol,
	valid_ingredients.shopping_suggestions as valid_ingredient_shopping_suggestions,
	valid_ingredients.is_starch as valid_ingredient_is_starch,
	valid_ingredients.is_protein as valid_ingredient_is_protein,
	valid_ingredients.is_grain as valid_ingredient_is_grain,
	valid_ingredients.is_fruit as valid_ingredient_is_fruit,
	valid_ingredients.is_salt as valid_ingredient_is_salt,
	valid_ingredients.is_fat as valid_ingredient_is_fat,
	valid_ingredients.is_acid as valid_ingredient_is_acid,
	valid_ingredients.is_heat as valid_ingredient_is_heat,
	valid_ingredients.last_indexed_at as valid_ingredient_last_indexed_at,
	valid_ingredients.created_at as valid_ingredient_created_at,
	valid_ingredients.last_updated_at as valid_ingredient_last_updated_at,
	valid_ingredients.archived_at as valid_ingredient_archived_at,
	valid_measurement_unit_conversions.from_unit as valid_measurement_unit_conversion_from_unit,
	valid_measurement_unit_conversions.to_unit as valid_measurement_unit_conversion_to_unit,
	valid_measurement_unit_conversions.only_for_ingredient as valid_measurement_unit_conversion_only_for_ingredient,
	valid_measurement_unit_conversions.modifier as valid_measurement_unit_conversion_modifier,
	valid_measurement_unit_conversions.notes as valid_measurement_unit_conversion_notes,
	valid_measurement_unit_conversions.created_at as valid_measurement_unit_conversion_created_at,
	valid_measurement_unit_conversions.last_updated_at as valid_measurement_unit_conversion_last_updated_at,
	valid_measurement_unit_conversions.archived_at as valid_measurement_unit_conversion_archived_at
FROM valid_measurement_unit_conversions
	JOIN valid_measurement_units AS valid_measurement_units_from ON valid_measurement_unit_conversions.from_unit = valid_measurement_units_from.id
	JOIN valid_measurement_units AS valid_measurement_units_to ON valid_measurement_unit_conversions.to_unit = valid_measurement_units_to.id
	LEFT JOIN valid_ingredients ON valid_measurement_unit_conversions.only_for_ingredient = valid_ingredients.id
WHERE
	valid_measurement_unit_conversions.id = sqlc.arg(id)
	AND valid_measurement_unit_conversions.archived_at IS NULL
	AND valid_measurement_units_from.archived_at IS NULL
	AND valid_measurement_units_to.archived_at IS NULL;

-- name: GetValidMeasurementUnitConversionsForIngredients :many
SELECT
	valid_measurement_unit_conversions.id as valid_measurement_unit_conversion_id,
	valid_measurement_units_from.id as from_unit_id,
	valid_measurement_units_from.name as from_unit_name,
	valid_measurement_units_from.description as from_unit_description,
	valid_measurement_units_from.volumetric as from_unit_volumetric,
	valid_measurement_units_from.icon_path as from_unit_icon_path,
	valid_measurement_units_from.universal as from_unit_universal,
	valid_measurement_units_from.metric as from_unit_metric,
	valid_measurement_units_from.imperial as from_unit_imperial,
	valid_measurement_units_from.slug as from_unit_slug,
	valid_measurement_units_from.plural_name as from_unit_plural_name,
	valid_measurement_units_from.last_indexed_at as from_unit_last_indexed_at,
	valid_measurement_units_from.created_at as from_unit_created_at,
	valid_measurement_units_from.last_updated_at as from_unit_last_updated_at,
	valid_measurement_units_from.archived_at as from_unit_archived_at,
	valid_measurement_units_to.id as to_unit_id,
	valid_measurement_units_to.name as to_unit_name,
	valid_measurement_units_to.description as to_unit_description,
	valid_measurement_units_to.volumetric as to_unit_volumetric,
	valid_measurement_units_to.icon_path as to_unit_icon_path,
	valid_measurement_units_to.universal as to_unit_universal,
	valid_measurement_units_to.metric as to_unit_metric,
	valid_measurement_units_to.imperial as to_unit_imperial,
	valid_measurement_units_to.slug as to_unit_slug,
	valid_measurement_units_to.plural_name as to_unit_plural_name,
	valid_measurement_units_to.last_indexed_at as to_unit_last_indexed_at,
	valid_measurement_units_to.created_at as to_unit_created_at,
	valid_measurement_units_to.last_updated_at as to_unit_last_updated_at,
	valid_measurement_units_to.archived_at as to_unit_archived_at,
	valid_ingredients.id as valid_ingredient_id,
	valid_ingredients.name as valid_ingredient_name,
	valid_ingredients.description as valid_ingredient_description,
	valid_ingredients.warning as valid_ingredient_warning,
	valid_ingredients.contains_egg as valid_ingredient_contains_egg,
	valid_ingredients.contains_dairy as valid_ingredient_contains_dairy,
	valid_ingredients.contains_peanut as valid_ingredient_contains_peanut,
	valid_ingredients.contains_tree_nut as valid_ingredient_contains_tree_nut,
	valid_ingredients.contains_soy as valid_ingredient_contains_soy,
	valid_ingredients.contains_wheat as valid_ingredient_contains_wheat,
	valid_ingredients.contains_shellfish as valid_ingredient_contains_shellfish,
	valid_ingredients.contains_sesame as valid_ingredient_contains_sesame,
	valid_ingredients.contains_fish as valid_ingredient_contains_fish,
	valid_ingredients.contains_gluten as valid_ingredient_contains_gluten,
	valid_ingredients.animal_flesh as valid_ingredient_animal_flesh,
	valid_ingredients.is_liquid as valid_ingredient_is_liquid,
	valid_ingredients.icon_path as valid_ingredient_icon_path,
	valid_ingredients.animal_derived as valid_ingredient_animal_derived,
	valid_ingredients.plural_name as valid_ingredient_plural_name,
	valid_ingredients.restrict_to_preparations as valid_ingredient_restrict_to_preparations,
	valid_ingredients.contaminates_equipment as valid_ingredient_contaminates_equipment,
	valid_ingredients.minimum_ideal_storage_temperature_in_celsius as valid_ingredient_minimum_ideal_storage_temperature_in_celsius,
	valid_ingredients.maximum_ideal_storage_temperature_in_celsius as valid_ingredient_maximum_ideal_storage_temperature_in_celsius,
	valid_ingredients.storage_instructions as valid_ingredient_storage_instructions,
	valid_ingredients.slug as valid_ingredient_slug,
	valid_ingredients.contains_alcohol as valid_ingredient_contains_alcohol,
	valid_ingredients.shopping_suggestions as valid_ingredient_shopping_suggestions,
	valid_ingredients.is_starch as valid_ingredient_is_starch,
	valid_ingredients.is_protein as valid_ingredient_is_protein,
	valid_ingredients.is_grain as valid_ingredient_is_grain,
	valid_ingredients.is_fruit as valid_ingredient_is_fruit,
	valid_ingredients.is_salt as valid_ingredient_is_salt,
	valid_ingredients.is_fat as valid_ingredient_is_fat,
	valid_ingredients.is_acid as valid_ingredient_is_acid,
	valid_ingredients.is_heat as valid_ingredient_is_heat,
	valid_ingredients.last_indexed_at as valid_ingredient_last_indexed_at,
	valid_ingredients.created_at as valid_ingredient_created_at,
	valid_ingredients.last_updated_at as valid_ingredient_last_updated_at,
	valid_ingredients.archived_at as valid_ingredient_archived_at,
	valid_measurement_unit_conversions.from_unit as valid_measurement_unit_conversion_from_unit,
	valid_measurement_unit_conversions.to_unit as valid_measurement_unit_conversion_to_unit,
	valid_measurement_unit_conversions.only_for_ingredient as valid_measurement_unit_conversion_only_for_ingredient,
	valid_measurement_unit_conversions.modifier as valid_measurement_unit_conversion_modifier,
	valid_measurement_unit_conversions.notes as valid_measurement_unit_conversion_notes,
	valid_measurement_unit_conversions.created_at as valid_measurement_unit_conversion_created_at,
	valid_measurement_unit_conversions.last_updated_at as valid_measurement_unit_conversion_last_updated_at,
	valid_measurement_unit_conversions.archived_at as valid_measurement_unit_conversion_archived_at
FROM valid_measurement_unit_conversions
	JOIN valid_measurement_units AS valid_measurement_units_from ON valid_measurement_unit_conversions.from_unit = valid_measurement_units_from.id
	JOIN valid_measurement_units AS valid_measurement_units_to ON valid_measurement_unit_conversions.to_unit = valid_measurement_units_to.id
	LEFT JOIN valid_ingredients ON valid_measurement_unit_conversions.only_for_ingredient = valid_ingredients.id
WHERE
	(valid_measurement_unit_conversions.only_for_ingredient IS NULL OR valid_measurement_unit_conversions.only_for_ingredient = ANY(sqlc.arg(valid_ingredient_ids)::text[]))
	AND valid_measurement_unit_conversions.archived_at IS NULL
	AND valid_measurement_units_from.archived_at IS NULL
	AND valid_measurement_units_to.archived_at IS NULL;

-- name: UpdateValidMeasurementUnitConversion :execrows
UPDATE valid_measurement_unit_conversions SET
	from_unit = sqlc.arg(from_unit),
	to_unit = sqlc.arg(to_unit),
	only_for_ingredient = sqlc.arg(only_for_ingredient),
	modifier = sqlc.arg(modifier),
	notes = sqlc.arg(notes),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: GetMeasurementUnitConversionMismatches :many
WITH ingredient_units AS (
	SELECT DISTINCT
		vimu.valid_ingredient_id,
		vimu.valid_measurement_unit_id
	FROM valid_ingredient_measurement_units vimu
	JOIN valid_ingredients vi ON vi.id = vimu.valid_ingredient_id
	JOIN valid_measurement_units vmu ON vmu.id = vimu.valid_measurement_unit_id
	WHERE vimu.archived_at IS NULL
		AND vi.archived_at IS NULL
		AND vmu.archived_at IS NULL
),
unit_pairs AS (
	SELECT
		a.valid_ingredient_id,
		a.valid_measurement_unit_id AS from_unit_id,
		b.valid_measurement_unit_id AS to_unit_id
	FROM ingredient_units a
	JOIN ingredient_units b ON a.valid_ingredient_id = b.valid_ingredient_id
		AND a.valid_measurement_unit_id < b.valid_measurement_unit_id
)
SELECT
	unit_pairs.valid_ingredient_id,
	unit_pairs.from_unit_id,
	unit_pairs.to_unit_id
FROM unit_pairs
LEFT JOIN valid_measurement_unit_conversions c ON c.from_unit = unit_pairs.from_unit_id
	AND c.to_unit = unit_pairs.to_unit_id
	AND (c.only_for_ingredient IS NULL OR c.only_for_ingredient = unit_pairs.valid_ingredient_id)
	AND c.archived_at IS NULL
WHERE c.id IS NULL;
