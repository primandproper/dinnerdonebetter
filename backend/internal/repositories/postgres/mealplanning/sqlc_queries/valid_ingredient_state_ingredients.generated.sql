-- name: CreateValidIngredientStateIngredient :exec
INSERT INTO valid_ingredient_state_ingredients (
	id,
	notes,
	valid_ingredient_state,
	valid_ingredient
) VALUES (
	sqlc.arg(id),
	sqlc.arg(notes),
	sqlc.arg(valid_ingredient_state),
	sqlc.arg(valid_ingredient)
);

-- name: CheckValidIngredientStateIngredientExistence :one
SELECT EXISTS (
	SELECT valid_ingredient_state_ingredients.id
	FROM valid_ingredient_state_ingredients
	WHERE valid_ingredient_state_ingredients.archived_at IS NULL
		AND valid_ingredient_state_ingredients.id = sqlc.arg(id)
);

-- name: UpdateValidIngredientStateIngredient :execrows
UPDATE valid_ingredient_state_ingredients SET
	notes = sqlc.arg(notes),
	valid_ingredient_state = sqlc.arg(valid_ingredient_state),
	valid_ingredient = sqlc.arg(valid_ingredient),
	last_updated_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveValidIngredientStateIngredient :execrows
UPDATE valid_ingredient_state_ingredients SET
	archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: GetValidIngredientStateIngredientsForIngredient :many
SELECT
	valid_ingredient_state_ingredients.id as valid_ingredient_state_ingredient_id,
	valid_ingredient_state_ingredients.notes as valid_ingredient_state_ingredient_notes,
	valid_ingredient_states.id as valid_ingredient_state_id,
	valid_ingredient_states.name as valid_ingredient_state_name,
	valid_ingredient_states.past_tense as valid_ingredient_state_past_tense,
	valid_ingredient_states.slug as valid_ingredient_state_slug,
	valid_ingredient_states.description as valid_ingredient_state_description,
	valid_ingredient_states.icon_path as valid_ingredient_state_icon_path,
	valid_ingredient_states.attribute_type as valid_ingredient_state_attribute_type,
	valid_ingredient_states.last_indexed_at as valid_ingredient_state_last_indexed_at,
	valid_ingredient_states.created_at as valid_ingredient_state_created_at,
	valid_ingredient_states.last_updated_at as valid_ingredient_state_last_updated_at,
	valid_ingredient_states.archived_at as valid_ingredient_state_archived_at,
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
	valid_ingredient_state_ingredients.valid_ingredient_state as valid_ingredient_state_ingredient_valid_ingredient_state,
	valid_ingredient_state_ingredients.valid_ingredient as valid_ingredient_state_ingredient_valid_ingredient,
	valid_ingredient_state_ingredients.created_at as valid_ingredient_state_ingredient_created_at,
	valid_ingredient_state_ingredients.last_updated_at as valid_ingredient_state_ingredient_last_updated_at,
	valid_ingredient_state_ingredients.archived_at as valid_ingredient_state_ingredient_archived_at,
	(
		SELECT COUNT(valid_ingredient_state_ingredients.id)
		FROM valid_ingredient_state_ingredients
		WHERE valid_ingredient_state_ingredients.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND valid_ingredient_state_ingredients.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				valid_ingredient_state_ingredients.last_updated_at IS NULL
				OR valid_ingredient_state_ingredients.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				valid_ingredient_state_ingredients.last_updated_at IS NULL
				OR valid_ingredient_state_ingredients.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_state_ingredients.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_ingredient_state_ingredients.id)
		FROM valid_ingredient_state_ingredients
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_state_ingredients.archived_at IS NULL)
	) AS total_count
FROM valid_ingredient_state_ingredients
	JOIN valid_ingredients ON valid_ingredient_state_ingredients.valid_ingredient = valid_ingredients.id
	JOIN valid_ingredient_states ON valid_ingredient_state_ingredients.valid_ingredient_state = valid_ingredient_states.id
WHERE valid_ingredient_state_ingredients.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND valid_ingredient_state_ingredients.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		valid_ingredient_state_ingredients.last_updated_at IS NULL
		OR valid_ingredient_state_ingredients.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		valid_ingredient_state_ingredients.last_updated_at IS NULL
		OR valid_ingredient_state_ingredients.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_state_ingredients.archived_at IS NULL)
	AND valid_ingredients.archived_at IS NULL
	AND valid_ingredient_states.archived_at IS NULL
	AND valid_ingredient_state_ingredients.valid_ingredient = sqlc.arg(valid_ingredient)
	AND valid_ingredient_state_ingredients.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY valid_ingredient_state_ingredients.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidIngredientStateIngredientsForIngredientState :many
SELECT
	valid_ingredient_state_ingredients.id as valid_ingredient_state_ingredient_id,
	valid_ingredient_state_ingredients.notes as valid_ingredient_state_ingredient_notes,
	valid_ingredient_states.id as valid_ingredient_state_id,
	valid_ingredient_states.name as valid_ingredient_state_name,
	valid_ingredient_states.past_tense as valid_ingredient_state_past_tense,
	valid_ingredient_states.slug as valid_ingredient_state_slug,
	valid_ingredient_states.description as valid_ingredient_state_description,
	valid_ingredient_states.icon_path as valid_ingredient_state_icon_path,
	valid_ingredient_states.attribute_type as valid_ingredient_state_attribute_type,
	valid_ingredient_states.last_indexed_at as valid_ingredient_state_last_indexed_at,
	valid_ingredient_states.created_at as valid_ingredient_state_created_at,
	valid_ingredient_states.last_updated_at as valid_ingredient_state_last_updated_at,
	valid_ingredient_states.archived_at as valid_ingredient_state_archived_at,
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
	valid_ingredient_state_ingredients.valid_ingredient_state as valid_ingredient_state_ingredient_valid_ingredient_state,
	valid_ingredient_state_ingredients.valid_ingredient as valid_ingredient_state_ingredient_valid_ingredient,
	valid_ingredient_state_ingredients.created_at as valid_ingredient_state_ingredient_created_at,
	valid_ingredient_state_ingredients.last_updated_at as valid_ingredient_state_ingredient_last_updated_at,
	valid_ingredient_state_ingredients.archived_at as valid_ingredient_state_ingredient_archived_at,
	(
		SELECT COUNT(valid_ingredient_state_ingredients.id)
		FROM valid_ingredient_state_ingredients
		WHERE valid_ingredient_state_ingredients.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND valid_ingredient_state_ingredients.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				valid_ingredient_state_ingredients.last_updated_at IS NULL
				OR valid_ingredient_state_ingredients.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				valid_ingredient_state_ingredients.last_updated_at IS NULL
				OR valid_ingredient_state_ingredients.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_state_ingredients.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_ingredient_state_ingredients.id)
		FROM valid_ingredient_state_ingredients
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_state_ingredients.archived_at IS NULL)
	) AS total_count
FROM valid_ingredient_state_ingredients
	JOIN valid_ingredients ON valid_ingredient_state_ingredients.valid_ingredient = valid_ingredients.id
	JOIN valid_ingredient_states ON valid_ingredient_state_ingredients.valid_ingredient_state = valid_ingredient_states.id
WHERE valid_ingredient_state_ingredients.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND valid_ingredient_state_ingredients.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		valid_ingredient_state_ingredients.last_updated_at IS NULL
		OR valid_ingredient_state_ingredients.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		valid_ingredient_state_ingredients.last_updated_at IS NULL
		OR valid_ingredient_state_ingredients.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_state_ingredients.archived_at IS NULL)
	AND valid_ingredients.archived_at IS NULL
	AND valid_ingredient_states.archived_at IS NULL
	AND valid_ingredient_state_ingredients.valid_ingredient_state = sqlc.arg(valid_ingredient_state)
	AND valid_ingredient_state_ingredients.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY valid_ingredient_state_ingredients.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidIngredientStateIngredients :many
SELECT
	valid_ingredient_state_ingredients.id as valid_ingredient_state_ingredient_id,
	valid_ingredient_state_ingredients.notes as valid_ingredient_state_ingredient_notes,
	valid_ingredient_states.id as valid_ingredient_state_id,
	valid_ingredient_states.name as valid_ingredient_state_name,
	valid_ingredient_states.past_tense as valid_ingredient_state_past_tense,
	valid_ingredient_states.slug as valid_ingredient_state_slug,
	valid_ingredient_states.description as valid_ingredient_state_description,
	valid_ingredient_states.icon_path as valid_ingredient_state_icon_path,
	valid_ingredient_states.attribute_type as valid_ingredient_state_attribute_type,
	valid_ingredient_states.last_indexed_at as valid_ingredient_state_last_indexed_at,
	valid_ingredient_states.created_at as valid_ingredient_state_created_at,
	valid_ingredient_states.last_updated_at as valid_ingredient_state_last_updated_at,
	valid_ingredient_states.archived_at as valid_ingredient_state_archived_at,
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
	valid_ingredient_state_ingredients.valid_ingredient_state as valid_ingredient_state_ingredient_valid_ingredient_state,
	valid_ingredient_state_ingredients.valid_ingredient as valid_ingredient_state_ingredient_valid_ingredient,
	valid_ingredient_state_ingredients.created_at as valid_ingredient_state_ingredient_created_at,
	valid_ingredient_state_ingredients.last_updated_at as valid_ingredient_state_ingredient_last_updated_at,
	valid_ingredient_state_ingredients.archived_at as valid_ingredient_state_ingredient_archived_at,
	(
		SELECT COUNT(valid_ingredient_state_ingredients.id)
		FROM valid_ingredient_state_ingredients
		WHERE valid_ingredient_state_ingredients.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND valid_ingredient_state_ingredients.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				valid_ingredient_state_ingredients.last_updated_at IS NULL
				OR valid_ingredient_state_ingredients.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				valid_ingredient_state_ingredients.last_updated_at IS NULL
				OR valid_ingredient_state_ingredients.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_state_ingredients.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_ingredient_state_ingredients.id)
		FROM valid_ingredient_state_ingredients
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_state_ingredients.archived_at IS NULL)
	) AS total_count
FROM valid_ingredient_state_ingredients
	JOIN valid_ingredients ON valid_ingredient_state_ingredients.valid_ingredient = valid_ingredients.id
	JOIN valid_ingredient_states ON valid_ingredient_state_ingredients.valid_ingredient_state = valid_ingredient_states.id
WHERE valid_ingredient_state_ingredients.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND valid_ingredient_state_ingredients.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		valid_ingredient_state_ingredients.last_updated_at IS NULL
		OR valid_ingredient_state_ingredients.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		valid_ingredient_state_ingredients.last_updated_at IS NULL
		OR valid_ingredient_state_ingredients.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_state_ingredients.archived_at IS NULL)
	AND valid_ingredients.archived_at IS NULL
	AND valid_ingredient_states.archived_at IS NULL
	AND valid_ingredient_state_ingredients.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY valid_ingredient_state_ingredients.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidIngredientStateIngredient :one
SELECT
	valid_ingredient_state_ingredients.id as valid_ingredient_state_ingredient_id,
	valid_ingredient_state_ingredients.notes as valid_ingredient_state_ingredient_notes,
	valid_ingredient_states.id as valid_ingredient_state_id,
	valid_ingredient_states.name as valid_ingredient_state_name,
	valid_ingredient_states.past_tense as valid_ingredient_state_past_tense,
	valid_ingredient_states.slug as valid_ingredient_state_slug,
	valid_ingredient_states.description as valid_ingredient_state_description,
	valid_ingredient_states.icon_path as valid_ingredient_state_icon_path,
	valid_ingredient_states.attribute_type as valid_ingredient_state_attribute_type,
	valid_ingredient_states.last_indexed_at as valid_ingredient_state_last_indexed_at,
	valid_ingredient_states.created_at as valid_ingredient_state_created_at,
	valid_ingredient_states.last_updated_at as valid_ingredient_state_last_updated_at,
	valid_ingredient_states.archived_at as valid_ingredient_state_archived_at,
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
	valid_ingredient_state_ingredients.valid_ingredient_state as valid_ingredient_state_ingredient_valid_ingredient_state,
	valid_ingredient_state_ingredients.valid_ingredient as valid_ingredient_state_ingredient_valid_ingredient,
	valid_ingredient_state_ingredients.created_at as valid_ingredient_state_ingredient_created_at,
	valid_ingredient_state_ingredients.last_updated_at as valid_ingredient_state_ingredient_last_updated_at,
	valid_ingredient_state_ingredients.archived_at as valid_ingredient_state_ingredient_archived_at
FROM valid_ingredient_state_ingredients
	JOIN valid_ingredients ON valid_ingredient_state_ingredients.valid_ingredient = valid_ingredients.id
	JOIN valid_ingredient_states ON valid_ingredient_state_ingredients.valid_ingredient_state = valid_ingredient_states.id
WHERE
	valid_ingredient_state_ingredients.archived_at IS NULL
	AND valid_ingredients.archived_at IS NULL
	AND valid_ingredient_states.archived_at IS NULL
	AND valid_ingredient_state_ingredients.id = sqlc.arg(id);

-- name: GetValidIngredientStateIngredientsWithIDs :many
SELECT
	valid_ingredient_state_ingredients.id as valid_ingredient_state_ingredient_id,
	valid_ingredient_state_ingredients.notes as valid_ingredient_state_ingredient_notes,
	valid_ingredient_states.id as valid_ingredient_state_id,
	valid_ingredient_states.name as valid_ingredient_state_name,
	valid_ingredient_states.past_tense as valid_ingredient_state_past_tense,
	valid_ingredient_states.slug as valid_ingredient_state_slug,
	valid_ingredient_states.description as valid_ingredient_state_description,
	valid_ingredient_states.icon_path as valid_ingredient_state_icon_path,
	valid_ingredient_states.attribute_type as valid_ingredient_state_attribute_type,
	valid_ingredient_states.last_indexed_at as valid_ingredient_state_last_indexed_at,
	valid_ingredient_states.created_at as valid_ingredient_state_created_at,
	valid_ingredient_states.last_updated_at as valid_ingredient_state_last_updated_at,
	valid_ingredient_states.archived_at as valid_ingredient_state_archived_at,
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
	valid_ingredient_state_ingredients.valid_ingredient_state as valid_ingredient_state_ingredient_valid_ingredient_state,
	valid_ingredient_state_ingredients.valid_ingredient as valid_ingredient_state_ingredient_valid_ingredient,
	valid_ingredient_state_ingredients.created_at as valid_ingredient_state_ingredient_created_at,
	valid_ingredient_state_ingredients.last_updated_at as valid_ingredient_state_ingredient_last_updated_at,
	valid_ingredient_state_ingredients.archived_at as valid_ingredient_state_ingredient_archived_at
FROM valid_ingredient_state_ingredients
	JOIN valid_ingredients ON valid_ingredient_state_ingredients.valid_ingredient = valid_ingredients.id
	JOIN valid_ingredient_states ON valid_ingredient_state_ingredients.valid_ingredient_state = valid_ingredient_states.id
WHERE
	valid_ingredient_state_ingredients.archived_at IS NULL
	AND valid_ingredients.archived_at IS NULL
	AND valid_ingredient_states.archived_at IS NULL
	AND valid_ingredient_state_ingredients.id = ANY(sqlc.arg(ids)::text[]);

-- name: CheckValidityOfValidIngredientStateIngredientPair :one
SELECT EXISTS(
	SELECT valid_ingredient_state_ingredients.id
	FROM valid_ingredient_state_ingredients
	WHERE valid_ingredient = sqlc.arg(valid_ingredient)
	AND valid_ingredient_state = sqlc.arg(valid_ingredient_state)
	AND archived_at IS NULL
);
