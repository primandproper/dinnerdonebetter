-- name: CreateValidPreparation :exec
INSERT INTO valid_preparations (
	id,
	name,
	description,
	icon_path,
	yields_nothing,
	restrict_to_ingredients,
	past_tense,
	slug,
	minimum_ingredient_count,
	maximum_ingredient_count,
	minimum_instrument_count,
	maximum_instrument_count,
	temperature_required,
	time_estimate_required,
	condition_expression_required,
	consumes_vessel,
	only_for_vessels,
	minimum_vessel_count,
	maximum_vessel_count
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description),
	sqlc.arg(icon_path),
	sqlc.arg(yields_nothing),
	sqlc.arg(restrict_to_ingredients),
	sqlc.arg(past_tense),
	sqlc.arg(slug),
	sqlc.arg(minimum_ingredient_count),
	sqlc.arg(maximum_ingredient_count),
	sqlc.arg(minimum_instrument_count),
	sqlc.arg(maximum_instrument_count),
	sqlc.arg(temperature_required),
	sqlc.arg(time_estimate_required),
	sqlc.arg(condition_expression_required),
	sqlc.arg(consumes_vessel),
	sqlc.arg(only_for_vessels),
	sqlc.arg(minimum_vessel_count),
	sqlc.arg(maximum_vessel_count)
);

-- name: GetValidPreparation :one
SELECT
	valid_preparations.id,
	valid_preparations.name,
	valid_preparations.description,
	valid_preparations.icon_path,
	valid_preparations.yields_nothing,
	valid_preparations.restrict_to_ingredients,
	valid_preparations.past_tense,
	valid_preparations.slug,
	valid_preparations.minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count,
	valid_preparations.minimum_instrument_count,
	valid_preparations.maximum_instrument_count,
	valid_preparations.temperature_required,
	valid_preparations.time_estimate_required,
	valid_preparations.condition_expression_required,
	valid_preparations.consumes_vessel,
	valid_preparations.only_for_vessels,
	valid_preparations.minimum_vessel_count,
	valid_preparations.maximum_vessel_count,
	valid_preparations.last_indexed_at,
	valid_preparations.created_at,
	valid_preparations.last_updated_at,
	valid_preparations.archived_at
FROM valid_preparations
WHERE valid_preparations.archived_at IS NULL
	AND valid_preparations.id = sqlc.arg(id);

-- name: CheckValidPreparationExistence :one
SELECT EXISTS (
	SELECT valid_preparations.id
	FROM valid_preparations
	WHERE valid_preparations.archived_at IS NULL
		AND valid_preparations.id = sqlc.arg(id)
);

-- name: UpdateValidPreparation :execrows
UPDATE valid_preparations SET
	name = sqlc.arg(name),
	description = sqlc.arg(description),
	icon_path = sqlc.arg(icon_path),
	yields_nothing = sqlc.arg(yields_nothing),
	restrict_to_ingredients = sqlc.arg(restrict_to_ingredients),
	past_tense = sqlc.arg(past_tense),
	slug = sqlc.arg(slug),
	minimum_ingredient_count = sqlc.arg(minimum_ingredient_count),
	maximum_ingredient_count = sqlc.arg(maximum_ingredient_count),
	minimum_instrument_count = sqlc.arg(minimum_instrument_count),
	maximum_instrument_count = sqlc.arg(maximum_instrument_count),
	temperature_required = sqlc.arg(temperature_required),
	time_estimate_required = sqlc.arg(time_estimate_required),
	condition_expression_required = sqlc.arg(condition_expression_required),
	consumes_vessel = sqlc.arg(consumes_vessel),
	only_for_vessels = sqlc.arg(only_for_vessels),
	minimum_vessel_count = sqlc.arg(minimum_vessel_count),
	maximum_vessel_count = sqlc.arg(maximum_vessel_count),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveValidPreparation :execrows
UPDATE valid_preparations SET
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ScanValidPreparationIDsForReindex :many
SELECT valid_preparations.id
FROM valid_preparations
WHERE valid_preparations.archived_at IS NULL
	AND valid_preparations.id COLLATE "C" > sqlc.arg(cursor)
ORDER BY valid_preparations.id COLLATE "C"
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidPreparations :many
SELECT
	valid_preparations.id,
	valid_preparations.name,
	valid_preparations.description,
	valid_preparations.icon_path,
	valid_preparations.yields_nothing,
	valid_preparations.restrict_to_ingredients,
	valid_preparations.past_tense,
	valid_preparations.slug,
	valid_preparations.minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count,
	valid_preparations.minimum_instrument_count,
	valid_preparations.maximum_instrument_count,
	valid_preparations.temperature_required,
	valid_preparations.time_estimate_required,
	valid_preparations.condition_expression_required,
	valid_preparations.consumes_vessel,
	valid_preparations.only_for_vessels,
	valid_preparations.minimum_vessel_count,
	valid_preparations.maximum_vessel_count,
	valid_preparations.last_indexed_at,
	valid_preparations.created_at,
	valid_preparations.last_updated_at,
	valid_preparations.archived_at,
	(
		SELECT COUNT(valid_preparations.id)
		FROM valid_preparations
		WHERE valid_preparations.archived_at IS NULL
			AND
			valid_preparations.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_preparations.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_preparations.last_updated_at IS NULL
				OR valid_preparations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_preparations.last_updated_at IS NULL
				OR valid_preparations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparations.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_preparations.id)
		FROM valid_preparations
		WHERE valid_preparations.archived_at IS NULL
	) AS total_count
FROM valid_preparations
WHERE
	valid_preparations.archived_at IS NULL
	AND valid_preparations.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_preparations.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_preparations.last_updated_at IS NULL
		OR valid_preparations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_preparations.last_updated_at IS NULL
		OR valid_preparations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparations.archived_at IS NULL)
	AND valid_preparations.id > COALESCE(sqlc.narg(cursor), '')
GROUP BY valid_preparations.id
ORDER BY valid_preparations.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidPreparationsNeedingIndexing :many
SELECT valid_preparations.id
FROM valid_preparations
WHERE valid_preparations.archived_at IS NULL
	AND (
	valid_preparations.last_indexed_at IS NULL
	OR valid_preparations.last_indexed_at < NOW() - '24 hours'::INTERVAL
);

-- name: GetRandomValidPreparation :one
SELECT
	valid_preparations.id,
	valid_preparations.name,
	valid_preparations.description,
	valid_preparations.icon_path,
	valid_preparations.yields_nothing,
	valid_preparations.restrict_to_ingredients,
	valid_preparations.past_tense,
	valid_preparations.slug,
	valid_preparations.minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count,
	valid_preparations.minimum_instrument_count,
	valid_preparations.maximum_instrument_count,
	valid_preparations.temperature_required,
	valid_preparations.time_estimate_required,
	valid_preparations.condition_expression_required,
	valid_preparations.consumes_vessel,
	valid_preparations.only_for_vessels,
	valid_preparations.minimum_vessel_count,
	valid_preparations.maximum_vessel_count,
	valid_preparations.last_indexed_at,
	valid_preparations.created_at,
	valid_preparations.last_updated_at,
	valid_preparations.archived_at
FROM valid_preparations
WHERE valid_preparations.archived_at IS NULL
ORDER BY RANDOM() LIMIT 1;

-- name: GetValidPreparationsWithIDs :many
SELECT
	valid_preparations.id,
	valid_preparations.name,
	valid_preparations.description,
	valid_preparations.icon_path,
	valid_preparations.yields_nothing,
	valid_preparations.restrict_to_ingredients,
	valid_preparations.past_tense,
	valid_preparations.slug,
	valid_preparations.minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count,
	valid_preparations.minimum_instrument_count,
	valid_preparations.maximum_instrument_count,
	valid_preparations.temperature_required,
	valid_preparations.time_estimate_required,
	valid_preparations.condition_expression_required,
	valid_preparations.consumes_vessel,
	valid_preparations.only_for_vessels,
	valid_preparations.minimum_vessel_count,
	valid_preparations.maximum_vessel_count,
	valid_preparations.last_indexed_at,
	valid_preparations.created_at,
	valid_preparations.last_updated_at,
	valid_preparations.archived_at
FROM valid_preparations
WHERE valid_preparations.archived_at IS NULL
	AND valid_preparations.id = ANY(sqlc.arg(ids)::text[]);

-- name: SearchForValidPreparations :many
SELECT
	valid_preparations.id,
	valid_preparations.name,
	valid_preparations.description,
	valid_preparations.icon_path,
	valid_preparations.yields_nothing,
	valid_preparations.restrict_to_ingredients,
	valid_preparations.past_tense,
	valid_preparations.slug,
	valid_preparations.minimum_ingredient_count,
	valid_preparations.maximum_ingredient_count,
	valid_preparations.minimum_instrument_count,
	valid_preparations.maximum_instrument_count,
	valid_preparations.temperature_required,
	valid_preparations.time_estimate_required,
	valid_preparations.condition_expression_required,
	valid_preparations.consumes_vessel,
	valid_preparations.only_for_vessels,
	valid_preparations.minimum_vessel_count,
	valid_preparations.maximum_vessel_count,
	valid_preparations.last_indexed_at,
	valid_preparations.created_at,
	valid_preparations.last_updated_at,
	valid_preparations.archived_at,
	(
		SELECT COUNT(valid_preparations.id)
		FROM valid_preparations
		WHERE valid_preparations.archived_at IS NULL
			AND
			valid_preparations.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND valid_preparations.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				valid_preparations.last_updated_at IS NULL
				OR valid_preparations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				valid_preparations.last_updated_at IS NULL
				OR valid_preparations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparations.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_preparations.id)
		FROM valid_preparations
		WHERE valid_preparations.archived_at IS NULL
	) AS total_count
FROM valid_preparations
WHERE valid_preparations.archived_at IS NULL
	AND valid_preparations.name ILIKE '%' || sqlc.arg(name_query)::text || '%'
	AND valid_preparations.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND valid_preparations.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		valid_preparations.last_updated_at IS NULL
		OR valid_preparations.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		valid_preparations.last_updated_at IS NULL
		OR valid_preparations.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_preparations.archived_at IS NULL)
	AND valid_preparations.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY valid_preparations.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: UpdateValidPreparationLastIndexedAt :execrows
UPDATE valid_preparations SET last_indexed_at = NOW() WHERE id = sqlc.arg(id) AND archived_at IS NULL;
