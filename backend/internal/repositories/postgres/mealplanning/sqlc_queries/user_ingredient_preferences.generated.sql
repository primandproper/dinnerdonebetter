-- name: CreateUserIngredientPreference :exec
INSERT INTO user_ingredient_preferences (
	id,
	ingredient,
	rating,
	notes,
	allergy,
	belongs_to_user
) VALUES (
	sqlc.arg(id),
	sqlc.arg(ingredient),
	sqlc.arg(rating),
	sqlc.arg(notes),
	sqlc.arg(allergy),
	sqlc.arg(belongs_to_user)
);

-- name: CheckUserIngredientPreferenceExistence :one
SELECT EXISTS (
	SELECT user_ingredient_preferences.id
	FROM user_ingredient_preferences
	WHERE user_ingredient_preferences.archived_at IS NULL
		AND user_ingredient_preferences.id = sqlc.arg(id)
		AND user_ingredient_preferences.belongs_to_user = sqlc.arg(belongs_to_user)
);

-- name: UpdateUserIngredientPreference :execrows
UPDATE user_ingredient_preferences SET
	ingredient = sqlc.arg(ingredient),
	rating = sqlc.arg(rating),
	notes = sqlc.arg(notes),
	allergy = sqlc.arg(allergy),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_user = sqlc.arg(belongs_to_user);

-- name: ArchiveUserIngredientPreference :execrows
UPDATE user_ingredient_preferences SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_user = sqlc.arg(belongs_to_user);

-- name: GetUserIngredientPreferencesForUser :many
SELECT
	user_ingredient_preferences.id,
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
	user_ingredient_preferences.rating,
	user_ingredient_preferences.notes,
	user_ingredient_preferences.allergy,
	user_ingredient_preferences.belongs_to_user,
	user_ingredient_preferences.created_at,
	user_ingredient_preferences.last_updated_at,
	user_ingredient_preferences.archived_at,
	(
		SELECT COUNT(user_ingredient_preferences.id)
		FROM user_ingredient_preferences
		WHERE user_ingredient_preferences.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND user_ingredient_preferences.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				user_ingredient_preferences.last_updated_at IS NULL
				OR user_ingredient_preferences.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				user_ingredient_preferences.last_updated_at IS NULL
				OR user_ingredient_preferences.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR user_ingredient_preferences.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(user_ingredient_preferences.id)
		FROM user_ingredient_preferences
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR user_ingredient_preferences.archived_at IS NULL)
	) AS total_count
FROM user_ingredient_preferences
	JOIN valid_ingredients ON valid_ingredients.id = user_ingredient_preferences.ingredient
WHERE user_ingredient_preferences.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND user_ingredient_preferences.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		user_ingredient_preferences.last_updated_at IS NULL
		OR user_ingredient_preferences.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		user_ingredient_preferences.last_updated_at IS NULL
		OR user_ingredient_preferences.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR user_ingredient_preferences.archived_at IS NULL)
	AND user_ingredient_preferences.belongs_to_user = sqlc.arg(belongs_to_user)
	AND valid_ingredients.archived_at IS NULL
	AND user_ingredient_preferences.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY user_ingredient_preferences.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetUserIngredientPreference :one
SELECT
	user_ingredient_preferences.id,
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
	user_ingredient_preferences.rating,
	user_ingredient_preferences.notes,
	user_ingredient_preferences.allergy,
	user_ingredient_preferences.belongs_to_user,
	user_ingredient_preferences.created_at,
	user_ingredient_preferences.last_updated_at,
	user_ingredient_preferences.archived_at
FROM user_ingredient_preferences
	JOIN valid_ingredients ON valid_ingredients.id = user_ingredient_preferences.ingredient
WHERE user_ingredient_preferences.archived_at IS NULL
	AND valid_ingredients.archived_at IS NULL
	AND user_ingredient_preferences.id = sqlc.arg(id)
	AND user_ingredient_preferences.belongs_to_user = sqlc.arg(belongs_to_user);
