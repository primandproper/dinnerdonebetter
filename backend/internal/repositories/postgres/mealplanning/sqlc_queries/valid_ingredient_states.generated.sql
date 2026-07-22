-- name: ArchiveValidIngredientState :execrows
UPDATE valid_ingredient_states SET archived_at = NOW() WHERE archived_at IS NULL AND id = sqlc.arg(id);

-- name: CreateValidIngredientState :exec
INSERT INTO valid_ingredient_states (
	id,
	name,
	past_tense,
	slug,
	description,
	icon_path,
	attribute_type
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(past_tense),
	sqlc.arg(slug),
	sqlc.arg(description),
	sqlc.arg(icon_path),
	sqlc.arg(attribute_type)
);

-- name: CheckValidIngredientStateExistence :one
SELECT EXISTS (
	SELECT valid_ingredient_states.id
	FROM valid_ingredient_states
	WHERE valid_ingredient_states.archived_at IS NULL
		AND valid_ingredient_states.id = sqlc.arg(id)
);

-- name: GetValidIngredientStates :many
SELECT
	valid_ingredient_states.id,
	valid_ingredient_states.name,
	valid_ingredient_states.past_tense,
	valid_ingredient_states.slug,
	valid_ingredient_states.description,
	valid_ingredient_states.icon_path,
	valid_ingredient_states.attribute_type,
	valid_ingredient_states.last_indexed_at,
	valid_ingredient_states.created_at,
	valid_ingredient_states.last_updated_at,
	valid_ingredient_states.archived_at,
	(
		SELECT COUNT(valid_ingredient_states.id)
		FROM valid_ingredient_states
		WHERE valid_ingredient_states.archived_at IS NULL
			AND
			valid_ingredient_states.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_ingredient_states.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_ingredient_states.last_updated_at IS NULL
				OR valid_ingredient_states.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_ingredient_states.last_updated_at IS NULL
				OR valid_ingredient_states.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_states.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_ingredient_states.id)
		FROM valid_ingredient_states
		WHERE valid_ingredient_states.archived_at IS NULL
	) AS total_count
FROM valid_ingredient_states
WHERE
	valid_ingredient_states.archived_at IS NULL
	AND valid_ingredient_states.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_ingredient_states.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_ingredient_states.last_updated_at IS NULL
		OR valid_ingredient_states.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_ingredient_states.last_updated_at IS NULL
		OR valid_ingredient_states.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_states.archived_at IS NULL)
	AND valid_ingredient_states.id > COALESCE(sqlc.narg(cursor), '')
GROUP BY valid_ingredient_states.id
ORDER BY valid_ingredient_states.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidIngredientStatesNeedingIndexing :many
SELECT valid_ingredient_states.id
FROM valid_ingredient_states
WHERE valid_ingredient_states.archived_at IS NULL
	AND (
	valid_ingredient_states.last_indexed_at IS NULL
	OR valid_ingredient_states.last_indexed_at < NOW() - '24 hours'::INTERVAL
);

-- name: GetValidIngredientState :one
SELECT
	valid_ingredient_states.id,
	valid_ingredient_states.name,
	valid_ingredient_states.past_tense,
	valid_ingredient_states.slug,
	valid_ingredient_states.description,
	valid_ingredient_states.icon_path,
	valid_ingredient_states.attribute_type,
	valid_ingredient_states.last_indexed_at,
	valid_ingredient_states.created_at,
	valid_ingredient_states.last_updated_at,
	valid_ingredient_states.archived_at
FROM valid_ingredient_states
WHERE valid_ingredient_states.archived_at IS NULL
AND valid_ingredient_states.id = sqlc.arg(id);

-- name: GetValidIngredientStatesWithIDs :many
SELECT
	valid_ingredient_states.id,
	valid_ingredient_states.name,
	valid_ingredient_states.past_tense,
	valid_ingredient_states.slug,
	valid_ingredient_states.description,
	valid_ingredient_states.icon_path,
	valid_ingredient_states.attribute_type,
	valid_ingredient_states.last_indexed_at,
	valid_ingredient_states.created_at,
	valid_ingredient_states.last_updated_at,
	valid_ingredient_states.archived_at
FROM valid_ingredient_states
WHERE valid_ingredient_states.archived_at IS NULL
	AND valid_ingredient_states.id = ANY(sqlc.arg(ids)::text[]);

-- name: SearchForValidIngredientStates :many
SELECT
	valid_ingredient_states.id,
	valid_ingredient_states.name,
	valid_ingredient_states.past_tense,
	valid_ingredient_states.slug,
	valid_ingredient_states.description,
	valid_ingredient_states.icon_path,
	valid_ingredient_states.attribute_type,
	valid_ingredient_states.last_indexed_at,
	valid_ingredient_states.created_at,
	valid_ingredient_states.last_updated_at,
	valid_ingredient_states.archived_at,
	(
		SELECT COUNT(valid_ingredient_states.id)
		FROM valid_ingredient_states
		WHERE valid_ingredient_states.archived_at IS NULL
			AND
			valid_ingredient_states.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_ingredient_states.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_ingredient_states.last_updated_at IS NULL
				OR valid_ingredient_states.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_ingredient_states.last_updated_at IS NULL
				OR valid_ingredient_states.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_states.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_ingredient_states.id)
		FROM valid_ingredient_states
		WHERE valid_ingredient_states.archived_at IS NULL
	) AS total_count
FROM valid_ingredient_states
WHERE valid_ingredient_states.archived_at IS NULL
	AND valid_ingredient_states.name ILIKE '%' || sqlc.arg(name_query)::text || '%'
	AND valid_ingredient_states.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_ingredient_states.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_ingredient_states.last_updated_at IS NULL
		OR valid_ingredient_states.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_ingredient_states.last_updated_at IS NULL
		OR valid_ingredient_states.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_ingredient_states.archived_at IS NULL)
	AND valid_ingredient_states.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY valid_ingredient_states.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: UpdateValidIngredientState :execrows
UPDATE valid_ingredient_states SET
	name = sqlc.arg(name),
	past_tense = sqlc.arg(past_tense),
	slug = sqlc.arg(slug),
	description = sqlc.arg(description),
	icon_path = sqlc.arg(icon_path),
	attribute_type = sqlc.arg(attribute_type),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: UpdateValidIngredientStateLastIndexedAt :execrows
UPDATE valid_ingredient_states SET last_indexed_at = NOW() WHERE id = sqlc.arg(id) AND archived_at IS NULL;
