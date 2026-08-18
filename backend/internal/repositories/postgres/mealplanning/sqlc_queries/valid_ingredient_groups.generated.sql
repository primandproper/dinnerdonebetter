-- name: CreateValidIngredientGroup :exec
INSERT INTO valid_ingredient_groups (
	id,
	name,
	description,
	slug
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description),
	sqlc.arg(slug)
);

-- name: GetValidIngredientGroup :one
SELECT
	valid_ingredient_groups.id,
	valid_ingredient_groups.name,
	valid_ingredient_groups.description,
	valid_ingredient_groups.slug,
	valid_ingredient_groups.created_at,
	valid_ingredient_groups.last_updated_at,
	valid_ingredient_groups.archived_at
FROM valid_ingredient_groups
WHERE valid_ingredient_groups.archived_at IS NULL
	AND valid_ingredient_groups.id = sqlc.arg(id);

-- name: CheckValidIngredientGroupExistence :one
SELECT EXISTS (
	SELECT valid_ingredient_groups.id
	FROM valid_ingredient_groups
	WHERE valid_ingredient_groups.archived_at IS NULL
		AND valid_ingredient_groups.id = sqlc.arg(id)
);

-- name: UpdateValidIngredientGroup :execrows
UPDATE valid_ingredient_groups SET
	name = sqlc.arg(name),
	description = sqlc.arg(description),
	slug = sqlc.arg(slug),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveValidIngredientGroup :execrows
UPDATE valid_ingredient_groups SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveValidIngredientGroupMember :execrows
UPDATE valid_ingredient_group_members SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id)
	AND belongs_to_group = sqlc.arg(belongs_to_group);

-- name: CreateValidIngredientGroupMember :exec
INSERT INTO valid_ingredient_group_members (
	id,
	belongs_to_group,
	valid_ingredient
) VALUES (
	sqlc.arg(id),
	sqlc.arg(belongs_to_group),
	sqlc.arg(valid_ingredient)
);

-- name: GetValidIngredientGroups :many
SELECT
	valid_ingredient_groups.id,
	valid_ingredient_groups.name,
	valid_ingredient_groups.description,
	valid_ingredient_groups.slug,
	valid_ingredient_groups.created_at,
	valid_ingredient_groups.last_updated_at,
	valid_ingredient_groups.archived_at,
	(
		SELECT COUNT(valid_ingredient_groups.id)
		FROM valid_ingredient_groups
		WHERE valid_ingredient_groups.archived_at IS NULL
			AND
			valid_ingredient_groups.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_ingredient_groups.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_ingredient_groups.last_updated_at IS NULL
				OR valid_ingredient_groups.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_ingredient_groups.last_updated_at IS NULL
				OR valid_ingredient_groups.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_groups.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_ingredient_groups.id)
		FROM valid_ingredient_groups
		WHERE valid_ingredient_groups.archived_at IS NULL
	) AS total_count
FROM valid_ingredient_groups
WHERE
	valid_ingredient_groups.archived_at IS NULL
	AND valid_ingredient_groups.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_ingredient_groups.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_ingredient_groups.last_updated_at IS NULL
		OR valid_ingredient_groups.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_ingredient_groups.last_updated_at IS NULL
		OR valid_ingredient_groups.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_groups.archived_at IS NULL)
	AND valid_ingredient_groups.id > COALESCE(sqlc.narg(cursor), '')
GROUP BY valid_ingredient_groups.id
ORDER BY valid_ingredient_groups.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidIngredientGroupMembers :many
SELECT
	valid_ingredient_group_members.id,
	valid_ingredient_group_members.belongs_to_group,
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
	valid_ingredient_group_members.created_at,
	valid_ingredient_group_members.archived_at
FROM valid_ingredient_group_members
	JOIN valid_ingredient_groups ON valid_ingredient_groups.id = valid_ingredient_group_members.belongs_to_group
	JOIN valid_ingredients ON valid_ingredients.id = valid_ingredient_group_members.valid_ingredient
WHERE
	valid_ingredient_groups.archived_at IS NULL
	AND valid_ingredient_group_members.archived_at IS NULL
	AND valid_ingredient_group_members.belongs_to_group = sqlc.arg(belongs_to_group);

-- name: SearchForValidIngredientGroups :many
SELECT
	valid_ingredient_groups.id,
	valid_ingredient_groups.name,
	valid_ingredient_groups.description,
	valid_ingredient_groups.slug,
	valid_ingredient_groups.created_at,
	valid_ingredient_groups.last_updated_at,
	valid_ingredient_groups.archived_at,
	(
		SELECT COUNT(valid_ingredient_groups.id)
		FROM valid_ingredient_groups
		WHERE valid_ingredient_groups.archived_at IS NULL
			AND
			valid_ingredient_groups.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_ingredient_groups.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_ingredient_groups.last_updated_at IS NULL
				OR valid_ingredient_groups.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_ingredient_groups.last_updated_at IS NULL
				OR valid_ingredient_groups.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_groups.archived_at IS NULL)
			AND valid_ingredient_groups.name ILIKE '%' || sqlc.arg(name)::text || '%'
	) AS filtered_count,
	(
		SELECT COUNT(valid_ingredient_groups.id)
		FROM valid_ingredient_groups
		WHERE valid_ingredient_groups.archived_at IS NULL
	) AS total_count
FROM valid_ingredient_groups
WHERE
	valid_ingredient_groups.archived_at IS NULL
	AND valid_ingredient_groups.name ILIKE '%' || sqlc.arg(name)::text || '%'
	AND valid_ingredient_groups.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_ingredient_groups.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_ingredient_groups.last_updated_at IS NULL
		OR valid_ingredient_groups.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_ingredient_groups.last_updated_at IS NULL
		OR valid_ingredient_groups.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_groups.archived_at IS NULL)
	AND valid_ingredient_groups.id > COALESCE(sqlc.narg(cursor), '')
GROUP BY valid_ingredient_groups.id
ORDER BY valid_ingredient_groups.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidIngredientGroupsWithIDs :many
SELECT
	valid_ingredient_groups.id,
	valid_ingredient_groups.name,
	valid_ingredient_groups.description,
	valid_ingredient_groups.slug,
	valid_ingredient_groups.created_at,
	valid_ingredient_groups.last_updated_at,
	valid_ingredient_groups.archived_at
FROM valid_ingredient_groups
WHERE valid_ingredient_groups.archived_at IS NULL
	AND valid_ingredient_groups.id = ANY(sqlc.arg(ids)::text[]);
