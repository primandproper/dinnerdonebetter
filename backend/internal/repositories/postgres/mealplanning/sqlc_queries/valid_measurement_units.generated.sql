-- name: CreateValidMeasurementUnit :exec
INSERT INTO valid_measurement_units (
	id,
	name,
	description,
	volumetric,
	icon_path,
	universal,
	metric,
	imperial,
	slug,
	plural_name
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description),
	sqlc.arg(volumetric),
	sqlc.arg(icon_path),
	sqlc.arg(universal),
	sqlc.arg(metric),
	sqlc.arg(imperial),
	sqlc.arg(slug),
	sqlc.arg(plural_name)
);

-- name: GetValidMeasurementUnit :one
SELECT
	valid_measurement_units.id,
	valid_measurement_units.name,
	valid_measurement_units.description,
	valid_measurement_units.volumetric,
	valid_measurement_units.icon_path,
	valid_measurement_units.universal,
	valid_measurement_units.metric,
	valid_measurement_units.imperial,
	valid_measurement_units.slug,
	valid_measurement_units.plural_name,
	valid_measurement_units.last_indexed_at,
	valid_measurement_units.created_at,
	valid_measurement_units.last_updated_at,
	valid_measurement_units.archived_at
FROM valid_measurement_units
WHERE valid_measurement_units.archived_at IS NULL
	AND valid_measurement_units.id = sqlc.arg(id);

-- name: CheckValidMeasurementUnitExistence :one
SELECT EXISTS (
	SELECT valid_measurement_units.id
	FROM valid_measurement_units
	WHERE valid_measurement_units.archived_at IS NULL
		AND valid_measurement_units.id = sqlc.arg(id)
);

-- name: UpdateValidMeasurementUnit :execrows
UPDATE valid_measurement_units SET
	name = sqlc.arg(name),
	description = sqlc.arg(description),
	volumetric = sqlc.arg(volumetric),
	icon_path = sqlc.arg(icon_path),
	universal = sqlc.arg(universal),
	metric = sqlc.arg(metric),
	imperial = sqlc.arg(imperial),
	slug = sqlc.arg(slug),
	plural_name = sqlc.arg(plural_name),
	last_updated_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveValidMeasurementUnit :execrows
UPDATE valid_measurement_units SET
	archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ScanValidMeasurementUnitIDsForReindex :many
SELECT valid_measurement_units.id
FROM valid_measurement_units
WHERE valid_measurement_units.archived_at IS NULL
	AND valid_measurement_units.id COLLATE "C" > sqlc.arg(page_cursor)
ORDER BY valid_measurement_units.id COLLATE "C"
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: MarkValidMeasurementUnitsAsIndexed :execrows
UPDATE valid_measurement_units SET
	last_indexed_at = CURRENT_TIMESTAMP
WHERE id = ANY(sqlc.arg(ids)::text[]);

-- name: GetValidMeasurementUnits :many
SELECT
	valid_measurement_units.id,
	valid_measurement_units.name,
	valid_measurement_units.description,
	valid_measurement_units.volumetric,
	valid_measurement_units.icon_path,
	valid_measurement_units.universal,
	valid_measurement_units.metric,
	valid_measurement_units.imperial,
	valid_measurement_units.slug,
	valid_measurement_units.plural_name,
	valid_measurement_units.last_indexed_at,
	valid_measurement_units.created_at,
	valid_measurement_units.last_updated_at,
	valid_measurement_units.archived_at,
	(
		SELECT COUNT(valid_measurement_units.id)
		FROM valid_measurement_units
		WHERE valid_measurement_units.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND valid_measurement_units.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				valid_measurement_units.last_updated_at IS NULL
				OR valid_measurement_units.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				valid_measurement_units.last_updated_at IS NULL
				OR valid_measurement_units.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_units.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_measurement_units.id)
		FROM valid_measurement_units
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_units.archived_at IS NULL)
	) AS total_count
FROM valid_measurement_units
WHERE valid_measurement_units.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND valid_measurement_units.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		valid_measurement_units.last_updated_at IS NULL
		OR valid_measurement_units.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		valid_measurement_units.last_updated_at IS NULL
		OR valid_measurement_units.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_units.archived_at IS NULL)
	AND valid_measurement_units.id > COALESCE(sqlc.narg(page_cursor), '')
GROUP BY valid_measurement_units.id
ORDER BY valid_measurement_units.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetUniversalValidMeasurementUnits :many
SELECT
	valid_measurement_units.id,
	valid_measurement_units.name,
	valid_measurement_units.description,
	valid_measurement_units.volumetric,
	valid_measurement_units.icon_path,
	valid_measurement_units.universal,
	valid_measurement_units.metric,
	valid_measurement_units.imperial,
	valid_measurement_units.slug,
	valid_measurement_units.plural_name,
	valid_measurement_units.last_indexed_at,
	valid_measurement_units.created_at,
	valid_measurement_units.last_updated_at,
	valid_measurement_units.archived_at,
	(
		SELECT COUNT(valid_measurement_units.id)
		FROM valid_measurement_units
		WHERE valid_measurement_units.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND valid_measurement_units.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				valid_measurement_units.last_updated_at IS NULL
				OR valid_measurement_units.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				valid_measurement_units.last_updated_at IS NULL
				OR valid_measurement_units.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_units.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_measurement_units.id)
		FROM valid_measurement_units
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_units.archived_at IS NULL)
	) AS total_count
FROM valid_measurement_units
WHERE valid_measurement_units.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND valid_measurement_units.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		valid_measurement_units.last_updated_at IS NULL
		OR valid_measurement_units.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		valid_measurement_units.last_updated_at IS NULL
		OR valid_measurement_units.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_units.archived_at IS NULL)
	AND valid_measurement_units.universal = TRUE
	AND valid_measurement_units.id > COALESCE(sqlc.narg(page_cursor), '')
GROUP BY valid_measurement_units.id
ORDER BY valid_measurement_units.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetValidMeasurementUnitsNeedingIndexing :many
SELECT valid_measurement_units.id
FROM valid_measurement_units
WHERE valid_measurement_units.archived_at IS NULL
	AND (
	valid_measurement_units.last_indexed_at IS NULL
	OR valid_measurement_units.last_indexed_at < CURRENT_TIMESTAMP - '24 hours'::INTERVAL
);

-- name: GetRandomValidMeasurementUnit :one
SELECT
	valid_measurement_units.id,
	valid_measurement_units.name,
	valid_measurement_units.description,
	valid_measurement_units.volumetric,
	valid_measurement_units.icon_path,
	valid_measurement_units.universal,
	valid_measurement_units.metric,
	valid_measurement_units.imperial,
	valid_measurement_units.slug,
	valid_measurement_units.plural_name,
	valid_measurement_units.last_indexed_at,
	valid_measurement_units.created_at,
	valid_measurement_units.last_updated_at,
	valid_measurement_units.archived_at
FROM valid_measurement_units
WHERE valid_measurement_units.archived_at IS NULL
ORDER BY RANDOM() LIMIT 1;

-- name: GetValidMeasurementUnitsWithIDs :many
SELECT
	valid_measurement_units.id,
	valid_measurement_units.name,
	valid_measurement_units.description,
	valid_measurement_units.volumetric,
	valid_measurement_units.icon_path,
	valid_measurement_units.universal,
	valid_measurement_units.metric,
	valid_measurement_units.imperial,
	valid_measurement_units.slug,
	valid_measurement_units.plural_name,
	valid_measurement_units.last_indexed_at,
	valid_measurement_units.created_at,
	valid_measurement_units.last_updated_at,
	valid_measurement_units.archived_at
FROM valid_measurement_units
WHERE valid_measurement_units.archived_at IS NULL
	AND valid_measurement_units.id = ANY(sqlc.arg(ids)::text[]);

-- name: SearchForValidMeasurementUnits :many
SELECT
	valid_measurement_units.id,
	valid_measurement_units.name,
	valid_measurement_units.description,
	valid_measurement_units.volumetric,
	valid_measurement_units.icon_path,
	valid_measurement_units.universal,
	valid_measurement_units.metric,
	valid_measurement_units.imperial,
	valid_measurement_units.slug,
	valid_measurement_units.plural_name,
	valid_measurement_units.last_indexed_at,
	valid_measurement_units.created_at,
	valid_measurement_units.last_updated_at,
	valid_measurement_units.archived_at,
	(
		SELECT COUNT(valid_measurement_units.id)
		FROM valid_measurement_units
		WHERE valid_measurement_units.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND valid_measurement_units.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				valid_measurement_units.last_updated_at IS NULL
				OR valid_measurement_units.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				valid_measurement_units.last_updated_at IS NULL
				OR valid_measurement_units.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_units.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(valid_measurement_units.id)
		FROM valid_measurement_units
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_units.archived_at IS NULL)
	) AS total_count
FROM valid_measurement_units
WHERE valid_measurement_units.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND valid_measurement_units.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		valid_measurement_units.last_updated_at IS NULL
		OR valid_measurement_units.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		valid_measurement_units.last_updated_at IS NULL
		OR valid_measurement_units.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_units.archived_at IS NULL)
	AND valid_measurement_units.name ILIKE '%' || sqlc.arg(name_query)::text || '%'
	AND valid_measurement_units.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY valid_measurement_units.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: SearchValidMeasurementUnitsByIngredientID :many
SELECT
	DISTINCT(valid_measurement_units.id),
	valid_measurement_units.name,
	valid_measurement_units.description,
	valid_measurement_units.volumetric,
	valid_measurement_units.icon_path,
	valid_measurement_units.universal,
	valid_measurement_units.metric,
	valid_measurement_units.imperial,
	valid_measurement_units.slug,
	valid_measurement_units.plural_name,
	valid_measurement_units.last_indexed_at,
	valid_measurement_units.created_at,
	valid_measurement_units.last_updated_at,
	valid_measurement_units.archived_at,
	(
		SELECT COUNT(valid_measurement_units.id)
		FROM valid_measurement_units
		WHERE valid_measurement_units.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND valid_measurement_units.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (
				valid_measurement_units.last_updated_at IS NULL
				OR valid_measurement_units.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			)
			AND (
				valid_measurement_units.last_updated_at IS NULL
				OR valid_measurement_units.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_units.archived_at IS NULL)
			AND  (
							valid_ingredient_measurement_units.valid_ingredient_id = sqlc.arg(valid_ingredient_id)
							OR valid_measurement_units.universal = true
						)
	) AS filtered_count,
	(
		SELECT COUNT(valid_measurement_units.id)
		FROM valid_measurement_units
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_units.archived_at IS NULL)
	) AS total_count
FROM valid_measurement_units
	JOIN valid_ingredient_measurement_units ON valid_ingredient_measurement_units.valid_measurement_unit_id = valid_measurement_units.id
	JOIN valid_ingredients ON valid_ingredient_measurement_units.valid_ingredient_id = valid_ingredients.id
WHERE valid_measurement_units.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND valid_measurement_units.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (
		valid_measurement_units.last_updated_at IS NULL
		OR valid_measurement_units.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	)
	AND (
		valid_measurement_units.last_updated_at IS NULL
		OR valid_measurement_units.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR valid_measurement_units.archived_at IS NULL)
	AND (
			valid_ingredient_measurement_units.valid_ingredient_id = sqlc.arg(valid_ingredient_id)
			OR valid_measurement_units.universal = TRUE
		)
	AND valid_ingredients.archived_at IS NULL
	AND valid_ingredient_measurement_units.archived_at IS NULL
	AND valid_measurement_units.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY valid_measurement_units.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
