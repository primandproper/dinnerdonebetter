-- name: CreateValidIngredientPreparation :exec
INSERT INTO valid_ingredient_preparations (
	id,
	notes,
	valid_preparation_id,
	valid_ingredient_id
) VALUES (
	sqlc.arg(id),
	sqlc.arg(notes),
	sqlc.arg(valid_preparation_id),
	sqlc.arg(valid_ingredient_id)
);

-- name: CheckValidIngredientPreparationExistence :one
SELECT EXISTS (
	SELECT valid_ingredient_preparations.id
	FROM valid_ingredient_preparations
	WHERE valid_ingredient_preparations.archived_at IS NULL
		AND valid_ingredient_preparations.id = sqlc.arg(id)
);

-- name: UpdateValidIngredientPreparation :execrows
UPDATE valid_ingredient_preparations SET
	notes = sqlc.arg(notes),
	valid_preparation_id = sqlc.arg(valid_preparation_id),
	valid_ingredient_id = sqlc.arg(valid_ingredient_id),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveValidIngredientPreparation :execrows
UPDATE valid_ingredient_preparations SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: GetValidIngredientPreparationsForIngredient :many
SELECT
	valid_ingredient_preparations.id as valid_ingredient_preparation_id,
	valid_ingredient_preparations.notes as valid_ingredient_preparation_notes,
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
	valid_ingredient_preparations.created_at as valid_ingredient_preparation_created_at,
	valid_ingredient_preparations.last_updated_at as valid_ingredient_preparation_last_updated_at,
	valid_ingredient_preparations.archived_at as valid_ingredient_preparation_archived_at,
	(
		SELECT COUNT(valid_ingredient_preparations.id)
		FROM valid_ingredient_preparations
		WHERE valid_ingredient_preparations.archived_at IS NULL
			AND
			valid_ingredient_preparations.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_ingredient_preparations.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_ingredient_preparations.last_updated_at IS NULL
				OR valid_ingredient_preparations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_ingredient_preparations.last_updated_at IS NULL
				OR valid_ingredient_preparations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_preparations.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_ingredient_preparations.id)
		FROM valid_ingredient_preparations
		WHERE valid_ingredient_preparations.archived_at IS NULL
	) AS total_count
FROM valid_ingredient_preparations
	JOIN valid_ingredients ON valid_ingredient_preparations.valid_ingredient_id = valid_ingredients.id
	JOIN valid_preparations ON valid_ingredient_preparations.valid_preparation_id = valid_preparations.id
WHERE
	valid_ingredient_preparations.archived_at IS NULL
	AND valid_ingredient_preparations.valid_ingredient_id = sqlc.arg(id)
	AND valid_ingredient_preparations.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_ingredient_preparations.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_ingredient_preparations.last_updated_at IS NULL
		OR valid_ingredient_preparations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_ingredient_preparations.last_updated_at IS NULL
		OR valid_ingredient_preparations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND valid_ingredient_preparations.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY valid_ingredient_preparations.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidIngredientPreparationsForPreparation :many
SELECT
	valid_ingredient_preparations.id as valid_ingredient_preparation_id,
	valid_ingredient_preparations.notes as valid_ingredient_preparation_notes,
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
	valid_ingredient_preparations.created_at as valid_ingredient_preparation_created_at,
	valid_ingredient_preparations.last_updated_at as valid_ingredient_preparation_last_updated_at,
	valid_ingredient_preparations.archived_at as valid_ingredient_preparation_archived_at,
	(
		SELECT COUNT(valid_ingredient_preparations.id)
		FROM valid_ingredient_preparations
		WHERE valid_ingredient_preparations.archived_at IS NULL
			AND
			valid_ingredient_preparations.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_ingredient_preparations.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_ingredient_preparations.last_updated_at IS NULL
				OR valid_ingredient_preparations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_ingredient_preparations.last_updated_at IS NULL
				OR valid_ingredient_preparations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_preparations.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_ingredient_preparations.id)
		FROM valid_ingredient_preparations
		WHERE valid_ingredient_preparations.archived_at IS NULL
	) AS total_count
FROM valid_ingredient_preparations
	JOIN valid_ingredients ON valid_ingredient_preparations.valid_ingredient_id = valid_ingredients.id
	JOIN valid_preparations ON valid_ingredient_preparations.valid_preparation_id = valid_preparations.id
WHERE
	valid_ingredient_preparations.archived_at IS NULL
	AND valid_ingredient_preparations.valid_preparation_id = sqlc.arg(id)
	AND valid_ingredient_preparations.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_ingredient_preparations.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_ingredient_preparations.last_updated_at IS NULL
		OR valid_ingredient_preparations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_ingredient_preparations.last_updated_at IS NULL
		OR valid_ingredient_preparations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND valid_ingredient_preparations.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY valid_ingredient_preparations.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidIngredientPreparations :many
SELECT
	valid_ingredient_preparations.id as valid_ingredient_preparation_id,
	valid_ingredient_preparations.notes as valid_ingredient_preparation_notes,
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
	valid_ingredient_preparations.created_at as valid_ingredient_preparation_created_at,
	valid_ingredient_preparations.last_updated_at as valid_ingredient_preparation_last_updated_at,
	valid_ingredient_preparations.archived_at as valid_ingredient_preparation_archived_at,
	(
		SELECT COUNT(valid_ingredient_preparations.id)
		FROM valid_ingredient_preparations
		WHERE valid_ingredient_preparations.archived_at IS NULL
			AND
			valid_ingredient_preparations.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_ingredient_preparations.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_ingredient_preparations.last_updated_at IS NULL
				OR valid_ingredient_preparations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_ingredient_preparations.last_updated_at IS NULL
				OR valid_ingredient_preparations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_preparations.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_ingredient_preparations.id)
		FROM valid_ingredient_preparations
		WHERE valid_ingredient_preparations.archived_at IS NULL
	) AS total_count
FROM valid_ingredient_preparations
	JOIN valid_ingredients ON valid_ingredient_preparations.valid_ingredient_id = valid_ingredients.id
	JOIN valid_preparations ON valid_ingredient_preparations.valid_preparation_id = valid_preparations.id
WHERE
	valid_ingredient_preparations.archived_at IS NULL
	AND valid_ingredient_preparations.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_ingredient_preparations.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_ingredient_preparations.last_updated_at IS NULL
		OR valid_ingredient_preparations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_ingredient_preparations.last_updated_at IS NULL
		OR valid_ingredient_preparations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND valid_ingredient_preparations.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY valid_ingredient_preparations.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidIngredientPreparation :one
SELECT
	valid_ingredient_preparations.id as valid_ingredient_preparation_id,
	valid_ingredient_preparations.notes as valid_ingredient_preparation_notes,
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
	valid_ingredient_preparations.created_at as valid_ingredient_preparation_created_at,
	valid_ingredient_preparations.last_updated_at as valid_ingredient_preparation_last_updated_at,
	valid_ingredient_preparations.archived_at as valid_ingredient_preparation_archived_at
FROM valid_ingredient_preparations
	JOIN valid_ingredients ON valid_ingredient_preparations.valid_ingredient_id = valid_ingredients.id
	JOIN valid_preparations ON valid_ingredient_preparations.valid_preparation_id = valid_preparations.id
WHERE
	valid_ingredient_preparations.archived_at IS NULL
	AND valid_ingredients.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_ingredient_preparations.id = sqlc.arg(id);

-- name: GetValidIngredientPreparationsByIDs :many
SELECT
	valid_ingredient_preparations.id as valid_ingredient_preparation_id,
	valid_ingredient_preparations.notes as valid_ingredient_preparation_notes,
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
	valid_ingredient_preparations.created_at as valid_ingredient_preparation_created_at,
	valid_ingredient_preparations.last_updated_at as valid_ingredient_preparation_last_updated_at,
	valid_ingredient_preparations.archived_at as valid_ingredient_preparation_archived_at
FROM valid_ingredient_preparations
	JOIN valid_ingredients ON valid_ingredient_preparations.valid_ingredient_id = valid_ingredients.id
	JOIN valid_preparations ON valid_ingredient_preparations.valid_preparation_id = valid_preparations.id
WHERE
	valid_ingredient_preparations.archived_at IS NULL
	AND valid_ingredients.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_ingredient_preparations.id = ANY(sqlc.arg(ids)::text[]);

-- name: ValidIngredientPreparationPairIsValid :one
SELECT EXISTS(
	SELECT id
	FROM valid_ingredient_preparations
	WHERE valid_ingredient_id = sqlc.arg(valid_ingredient_id)
	AND valid_preparation_id = sqlc.arg(valid_preparation_id)
	AND archived_at IS NULL
);

-- name: SearchValidIngredientPreparationsByPreparationAndIngredientName :many
SELECT
	valid_ingredient_preparations.id as valid_ingredient_preparation_id,
	valid_ingredient_preparations.notes as valid_ingredient_preparation_notes,
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
	valid_ingredient_preparations.created_at as valid_ingredient_preparation_created_at,
	valid_ingredient_preparations.last_updated_at as valid_ingredient_preparation_last_updated_at,
	valid_ingredient_preparations.archived_at as valid_ingredient_preparation_archived_at,
	(
		SELECT COUNT(valid_ingredient_preparations.id)
		FROM valid_ingredient_preparations
		WHERE valid_ingredient_preparations.archived_at IS NULL
			AND
			valid_ingredient_preparations.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_ingredient_preparations.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_ingredient_preparations.last_updated_at IS NULL
				OR valid_ingredient_preparations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_ingredient_preparations.last_updated_at IS NULL
				OR valid_ingredient_preparations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_preparations.archived_at IS NULL)
			AND valid_preparations.id = sqlc.arg(id)
	) AS filtered_count,
	(
		SELECT COUNT(valid_ingredient_preparations.id)
		FROM valid_ingredient_preparations
		WHERE valid_ingredient_preparations.archived_at IS NULL
			AND valid_preparations.id = sqlc.arg(id)
	) AS total_count
FROM valid_ingredient_preparations
	JOIN valid_ingredients ON valid_ingredient_preparations.valid_ingredient_id = valid_ingredients.id
	JOIN valid_preparations ON valid_ingredient_preparations.valid_preparation_id = valid_preparations.id
WHERE
	valid_ingredient_preparations.archived_at IS NULL
	AND valid_ingredients.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_preparations.id = sqlc.arg(id)
	AND valid_ingredients.name ILIKE '%' || sqlc.arg(name_query)::text || '%'
	AND valid_ingredient_preparations.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_ingredient_preparations.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_ingredient_preparations.last_updated_at IS NULL
		OR valid_ingredient_preparations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_ingredient_preparations.last_updated_at IS NULL
		OR valid_ingredient_preparations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_preparations.archived_at IS NULL)
	AND valid_preparations.id = sqlc.arg(id)
	AND valid_ingredient_preparations.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY valid_ingredient_preparations.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
