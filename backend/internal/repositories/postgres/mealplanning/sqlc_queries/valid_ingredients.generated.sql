-- name: CreateValidIngredient :exec
INSERT INTO valid_ingredients (
	id,
	name,
	description,
	warning,
	contains_egg,
	contains_dairy,
	contains_peanut,
	contains_tree_nut,
	contains_soy,
	contains_wheat,
	contains_shellfish,
	contains_sesame,
	contains_fish,
	contains_gluten,
	animal_flesh,
	is_liquid,
	icon_path,
	animal_derived,
	plural_name,
	restrict_to_preparations,
	contaminates_equipment,
	minimum_ideal_storage_temperature_in_celsius,
	maximum_ideal_storage_temperature_in_celsius,
	storage_instructions,
	slug,
	contains_alcohol,
	shopping_suggestions,
	is_starch,
	is_protein,
	is_grain,
	is_fruit,
	is_salt,
	is_fat,
	is_acid,
	is_heat
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description),
	sqlc.arg(warning),
	sqlc.arg(contains_egg),
	sqlc.arg(contains_dairy),
	sqlc.arg(contains_peanut),
	sqlc.arg(contains_tree_nut),
	sqlc.arg(contains_soy),
	sqlc.arg(contains_wheat),
	sqlc.arg(contains_shellfish),
	sqlc.arg(contains_sesame),
	sqlc.arg(contains_fish),
	sqlc.arg(contains_gluten),
	sqlc.arg(animal_flesh),
	sqlc.arg(is_liquid),
	sqlc.arg(icon_path),
	sqlc.arg(animal_derived),
	sqlc.arg(plural_name),
	sqlc.arg(restrict_to_preparations),
	sqlc.arg(contaminates_equipment),
	sqlc.narg(minimum_ideal_storage_temperature_in_celsius),
	sqlc.narg(maximum_ideal_storage_temperature_in_celsius),
	sqlc.arg(storage_instructions),
	sqlc.arg(slug),
	sqlc.arg(contains_alcohol),
	sqlc.arg(shopping_suggestions),
	sqlc.arg(is_starch),
	sqlc.arg(is_protein),
	sqlc.arg(is_grain),
	sqlc.arg(is_fruit),
	sqlc.arg(is_salt),
	sqlc.arg(is_fat),
	sqlc.arg(is_acid),
	sqlc.arg(is_heat)
);

-- name: GetValidIngredient :one
SELECT
	valid_ingredients.id,
	valid_ingredients.name,
	valid_ingredients.description,
	valid_ingredients.warning,
	valid_ingredients.contains_egg,
	valid_ingredients.contains_dairy,
	valid_ingredients.contains_peanut,
	valid_ingredients.contains_tree_nut,
	valid_ingredients.contains_soy,
	valid_ingredients.contains_wheat,
	valid_ingredients.contains_shellfish,
	valid_ingredients.contains_sesame,
	valid_ingredients.contains_fish,
	valid_ingredients.contains_gluten,
	valid_ingredients.animal_flesh,
	valid_ingredients.is_liquid,
	valid_ingredients.icon_path,
	valid_ingredients.animal_derived,
	valid_ingredients.plural_name,
	valid_ingredients.restrict_to_preparations,
	valid_ingredients.contaminates_equipment,
	valid_ingredients.minimum_ideal_storage_temperature_in_celsius,
	valid_ingredients.maximum_ideal_storage_temperature_in_celsius,
	valid_ingredients.storage_instructions,
	valid_ingredients.slug,
	valid_ingredients.contains_alcohol,
	valid_ingredients.shopping_suggestions,
	valid_ingredients.is_starch,
	valid_ingredients.is_protein,
	valid_ingredients.is_grain,
	valid_ingredients.is_fruit,
	valid_ingredients.is_salt,
	valid_ingredients.is_fat,
	valid_ingredients.is_acid,
	valid_ingredients.is_heat,
	valid_ingredients.last_indexed_at,
	valid_ingredients.created_at,
	valid_ingredients.last_updated_at,
	valid_ingredients.archived_at
FROM valid_ingredients
WHERE valid_ingredients.archived_at IS NULL
	AND valid_ingredients.id = sqlc.arg(id);

-- name: CheckValidIngredientExistence :one
SELECT EXISTS (
	SELECT valid_ingredients.id
	FROM valid_ingredients
	WHERE valid_ingredients.archived_at IS NULL
		AND valid_ingredients.id = sqlc.arg(id)
);

-- name: UpdateValidIngredient :execrows
UPDATE valid_ingredients SET
	name = sqlc.arg(name),
	description = sqlc.arg(description),
	warning = sqlc.arg(warning),
	contains_egg = sqlc.arg(contains_egg),
	contains_dairy = sqlc.arg(contains_dairy),
	contains_peanut = sqlc.arg(contains_peanut),
	contains_tree_nut = sqlc.arg(contains_tree_nut),
	contains_soy = sqlc.arg(contains_soy),
	contains_wheat = sqlc.arg(contains_wheat),
	contains_shellfish = sqlc.arg(contains_shellfish),
	contains_sesame = sqlc.arg(contains_sesame),
	contains_fish = sqlc.arg(contains_fish),
	contains_gluten = sqlc.arg(contains_gluten),
	animal_flesh = sqlc.arg(animal_flesh),
	is_liquid = sqlc.arg(is_liquid),
	icon_path = sqlc.arg(icon_path),
	animal_derived = sqlc.arg(animal_derived),
	plural_name = sqlc.arg(plural_name),
	restrict_to_preparations = sqlc.arg(restrict_to_preparations),
	contaminates_equipment = sqlc.arg(contaminates_equipment),
	minimum_ideal_storage_temperature_in_celsius = sqlc.narg(minimum_ideal_storage_temperature_in_celsius),
	maximum_ideal_storage_temperature_in_celsius = sqlc.narg(maximum_ideal_storage_temperature_in_celsius),
	storage_instructions = sqlc.arg(storage_instructions),
	slug = sqlc.arg(slug),
	contains_alcohol = sqlc.arg(contains_alcohol),
	shopping_suggestions = sqlc.arg(shopping_suggestions),
	is_starch = sqlc.arg(is_starch),
	is_protein = sqlc.arg(is_protein),
	is_grain = sqlc.arg(is_grain),
	is_fruit = sqlc.arg(is_fruit),
	is_salt = sqlc.arg(is_salt),
	is_fat = sqlc.arg(is_fat),
	is_acid = sqlc.arg(is_acid),
	is_heat = sqlc.arg(is_heat),
	last_updated_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveValidIngredient :execrows
UPDATE valid_ingredients SET
	archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ScanValidIngredientIDsForReindex :many
SELECT valid_ingredients.id
FROM valid_ingredients
WHERE valid_ingredients.archived_at IS NULL
	AND valid_ingredients.id COLLATE "C" > sqlc.arg(cursor)
ORDER BY valid_ingredients.id COLLATE "C"
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: MarkValidIngredientsAsIndexed :execrows
UPDATE valid_ingredients SET
	last_indexed_at = CURRENT_TIMESTAMP
WHERE id = ANY(sqlc.arg(ids)::text[]);

-- name: GetValidIngredients :many
SELECT
	valid_ingredients.id,
	valid_ingredients.name,
	valid_ingredients.description,
	valid_ingredients.warning,
	valid_ingredients.contains_egg,
	valid_ingredients.contains_dairy,
	valid_ingredients.contains_peanut,
	valid_ingredients.contains_tree_nut,
	valid_ingredients.contains_soy,
	valid_ingredients.contains_wheat,
	valid_ingredients.contains_shellfish,
	valid_ingredients.contains_sesame,
	valid_ingredients.contains_fish,
	valid_ingredients.contains_gluten,
	valid_ingredients.animal_flesh,
	valid_ingredients.is_liquid,
	valid_ingredients.icon_path,
	valid_ingredients.animal_derived,
	valid_ingredients.plural_name,
	valid_ingredients.restrict_to_preparations,
	valid_ingredients.contaminates_equipment,
	valid_ingredients.minimum_ideal_storage_temperature_in_celsius,
	valid_ingredients.maximum_ideal_storage_temperature_in_celsius,
	valid_ingredients.storage_instructions,
	valid_ingredients.slug,
	valid_ingredients.contains_alcohol,
	valid_ingredients.shopping_suggestions,
	valid_ingredients.is_starch,
	valid_ingredients.is_protein,
	valid_ingredients.is_grain,
	valid_ingredients.is_fruit,
	valid_ingredients.is_salt,
	valid_ingredients.is_fat,
	valid_ingredients.is_acid,
	valid_ingredients.is_heat,
	valid_ingredients.last_indexed_at,
	valid_ingredients.created_at,
	valid_ingredients.last_updated_at,
	valid_ingredients.archived_at,
	(
		SELECT COUNT(valid_ingredients.id)
		FROM valid_ingredients
		WHERE valid_ingredients.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND valid_ingredients.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				valid_ingredients.last_updated_at IS NULL
				OR valid_ingredients.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				valid_ingredients.last_updated_at IS NULL
				OR valid_ingredients.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredients.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_ingredients.id)
		FROM valid_ingredients
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredients.archived_at IS NULL)
	) AS total_count
FROM valid_ingredients
WHERE valid_ingredients.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND valid_ingredients.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		valid_ingredients.last_updated_at IS NULL
		OR valid_ingredients.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		valid_ingredients.last_updated_at IS NULL
		OR valid_ingredients.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredients.archived_at IS NULL)
	AND valid_ingredients.id > COALESCE(sqlc.narg(cursor), '')
GROUP BY valid_ingredients.id
ORDER BY valid_ingredients.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidIngredientsNeedingIndexing :many
SELECT valid_ingredients.id
FROM valid_ingredients
WHERE valid_ingredients.archived_at IS NULL
	AND (
	valid_ingredients.last_indexed_at IS NULL
	OR valid_ingredients.last_indexed_at < CURRENT_TIMESTAMP - '24 hours'::INTERVAL
);

-- name: GetRandomValidIngredient :one
SELECT
	valid_ingredients.id,
	valid_ingredients.name,
	valid_ingredients.description,
	valid_ingredients.warning,
	valid_ingredients.contains_egg,
	valid_ingredients.contains_dairy,
	valid_ingredients.contains_peanut,
	valid_ingredients.contains_tree_nut,
	valid_ingredients.contains_soy,
	valid_ingredients.contains_wheat,
	valid_ingredients.contains_shellfish,
	valid_ingredients.contains_sesame,
	valid_ingredients.contains_fish,
	valid_ingredients.contains_gluten,
	valid_ingredients.animal_flesh,
	valid_ingredients.is_liquid,
	valid_ingredients.icon_path,
	valid_ingredients.animal_derived,
	valid_ingredients.plural_name,
	valid_ingredients.restrict_to_preparations,
	valid_ingredients.contaminates_equipment,
	valid_ingredients.minimum_ideal_storage_temperature_in_celsius,
	valid_ingredients.maximum_ideal_storage_temperature_in_celsius,
	valid_ingredients.storage_instructions,
	valid_ingredients.slug,
	valid_ingredients.contains_alcohol,
	valid_ingredients.shopping_suggestions,
	valid_ingredients.is_starch,
	valid_ingredients.is_protein,
	valid_ingredients.is_grain,
	valid_ingredients.is_fruit,
	valid_ingredients.is_salt,
	valid_ingredients.is_fat,
	valid_ingredients.is_acid,
	valid_ingredients.is_heat,
	valid_ingredients.last_indexed_at,
	valid_ingredients.created_at,
	valid_ingredients.last_updated_at,
	valid_ingredients.archived_at
FROM valid_ingredients
WHERE valid_ingredients.archived_at IS NULL
ORDER BY RANDOM() LIMIT 1;

-- name: GetValidIngredientsWithIDs :many
SELECT
	valid_ingredients.id,
	valid_ingredients.name,
	valid_ingredients.description,
	valid_ingredients.warning,
	valid_ingredients.contains_egg,
	valid_ingredients.contains_dairy,
	valid_ingredients.contains_peanut,
	valid_ingredients.contains_tree_nut,
	valid_ingredients.contains_soy,
	valid_ingredients.contains_wheat,
	valid_ingredients.contains_shellfish,
	valid_ingredients.contains_sesame,
	valid_ingredients.contains_fish,
	valid_ingredients.contains_gluten,
	valid_ingredients.animal_flesh,
	valid_ingredients.is_liquid,
	valid_ingredients.icon_path,
	valid_ingredients.animal_derived,
	valid_ingredients.plural_name,
	valid_ingredients.restrict_to_preparations,
	valid_ingredients.contaminates_equipment,
	valid_ingredients.minimum_ideal_storage_temperature_in_celsius,
	valid_ingredients.maximum_ideal_storage_temperature_in_celsius,
	valid_ingredients.storage_instructions,
	valid_ingredients.slug,
	valid_ingredients.contains_alcohol,
	valid_ingredients.shopping_suggestions,
	valid_ingredients.is_starch,
	valid_ingredients.is_protein,
	valid_ingredients.is_grain,
	valid_ingredients.is_fruit,
	valid_ingredients.is_salt,
	valid_ingredients.is_fat,
	valid_ingredients.is_acid,
	valid_ingredients.is_heat,
	valid_ingredients.last_indexed_at,
	valid_ingredients.created_at,
	valid_ingredients.last_updated_at,
	valid_ingredients.archived_at
FROM valid_ingredients
WHERE valid_ingredients.archived_at IS NULL
	AND valid_ingredients.id = ANY(sqlc.arg(ids)::text[]);

-- name: SearchForValidIngredients :many
SELECT
	valid_ingredients.id,
	valid_ingredients.name,
	valid_ingredients.description,
	valid_ingredients.warning,
	valid_ingredients.contains_egg,
	valid_ingredients.contains_dairy,
	valid_ingredients.contains_peanut,
	valid_ingredients.contains_tree_nut,
	valid_ingredients.contains_soy,
	valid_ingredients.contains_wheat,
	valid_ingredients.contains_shellfish,
	valid_ingredients.contains_sesame,
	valid_ingredients.contains_fish,
	valid_ingredients.contains_gluten,
	valid_ingredients.animal_flesh,
	valid_ingredients.is_liquid,
	valid_ingredients.icon_path,
	valid_ingredients.animal_derived,
	valid_ingredients.plural_name,
	valid_ingredients.restrict_to_preparations,
	valid_ingredients.contaminates_equipment,
	valid_ingredients.minimum_ideal_storage_temperature_in_celsius,
	valid_ingredients.maximum_ideal_storage_temperature_in_celsius,
	valid_ingredients.storage_instructions,
	valid_ingredients.slug,
	valid_ingredients.contains_alcohol,
	valid_ingredients.shopping_suggestions,
	valid_ingredients.is_starch,
	valid_ingredients.is_protein,
	valid_ingredients.is_grain,
	valid_ingredients.is_fruit,
	valid_ingredients.is_salt,
	valid_ingredients.is_fat,
	valid_ingredients.is_acid,
	valid_ingredients.is_heat,
	valid_ingredients.last_indexed_at,
	valid_ingredients.created_at,
	valid_ingredients.last_updated_at,
	valid_ingredients.archived_at,
	(
		SELECT COUNT(valid_ingredients.id)
		FROM valid_ingredients
		WHERE valid_ingredients.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND valid_ingredients.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				valid_ingredients.last_updated_at IS NULL
				OR valid_ingredients.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				valid_ingredients.last_updated_at IS NULL
				OR valid_ingredients.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredients.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_ingredients.id)
		FROM valid_ingredients
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredients.archived_at IS NULL)
	) AS total_count
FROM valid_ingredients
WHERE valid_ingredients.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND valid_ingredients.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		valid_ingredients.last_updated_at IS NULL
		OR valid_ingredients.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		valid_ingredients.last_updated_at IS NULL
		OR valid_ingredients.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredients.archived_at IS NULL)
	AND valid_ingredients.name ILIKE '%' || sqlc.arg(name_query)::text || '%'
	AND valid_ingredients.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY valid_ingredients.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: SearchValidIngredientsByPreparationAndIngredientName :many
SELECT
	DISTINCT(valid_ingredients.id),
	valid_ingredients.name,
	valid_ingredients.description,
	valid_ingredients.warning,
	valid_ingredients.contains_egg,
	valid_ingredients.contains_dairy,
	valid_ingredients.contains_peanut,
	valid_ingredients.contains_tree_nut,
	valid_ingredients.contains_soy,
	valid_ingredients.contains_wheat,
	valid_ingredients.contains_shellfish,
	valid_ingredients.contains_sesame,
	valid_ingredients.contains_fish,
	valid_ingredients.contains_gluten,
	valid_ingredients.animal_flesh,
	valid_ingredients.is_liquid,
	valid_ingredients.icon_path,
	valid_ingredients.animal_derived,
	valid_ingredients.plural_name,
	valid_ingredients.restrict_to_preparations,
	valid_ingredients.contaminates_equipment,
	valid_ingredients.minimum_ideal_storage_temperature_in_celsius,
	valid_ingredients.maximum_ideal_storage_temperature_in_celsius,
	valid_ingredients.storage_instructions,
	valid_ingredients.slug,
	valid_ingredients.contains_alcohol,
	valid_ingredients.shopping_suggestions,
	valid_ingredients.is_starch,
	valid_ingredients.is_protein,
	valid_ingredients.is_grain,
	valid_ingredients.is_fruit,
	valid_ingredients.is_salt,
	valid_ingredients.is_fat,
	valid_ingredients.is_acid,
	valid_ingredients.is_heat,
	valid_ingredients.last_indexed_at,
	valid_ingredients.created_at,
	valid_ingredients.last_updated_at,
	valid_ingredients.archived_at
FROM valid_ingredient_preparations
	JOIN valid_ingredients ON valid_ingredient_preparations.valid_ingredient_id = valid_ingredients.id
	JOIN valid_preparations ON valid_ingredient_preparations.valid_preparation_id = valid_preparations.id
WHERE valid_ingredient_preparations.archived_at IS NULL
	AND valid_ingredients.archived_at IS NULL
	AND valid_preparations.archived_at IS NULL
	AND valid_ingredient_preparations.valid_preparation_id = sqlc.arg(valid_preparation_id)
	AND valid_ingredients.name ILIKE '%' || sqlc.arg(name_query)::text || '%';
