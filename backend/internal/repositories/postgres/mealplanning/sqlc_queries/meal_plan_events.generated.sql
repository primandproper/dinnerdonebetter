-- name: ArchiveMealPlanEvent :execrows
UPDATE meal_plan_events SET archived_at = NOW() WHERE archived_at IS NULL AND id = sqlc.arg(id) AND belongs_to_meal_plan = sqlc.arg(belongs_to_meal_plan);

-- name: CreateMealPlanEvent :exec
INSERT INTO meal_plan_events (
	id,
	notes,
	starts_at,
	ends_at,
	meal_name,
	belongs_to_meal_plan
) VALUES (
	sqlc.arg(id),
	sqlc.arg(notes),
	sqlc.arg(starts_at),
	sqlc.arg(ends_at),
	sqlc.arg(meal_name),
	sqlc.arg(belongs_to_meal_plan)
);

-- name: MealPlanEventIsEligibleForVoting :one
SELECT EXISTS (
	SELECT meal_plan_events.id
	FROM meal_plan_events
		JOIN meal_plans ON meal_plan_events.belongs_to_meal_plan = meal_plans.id
	WHERE
		meal_plan_events.archived_at IS NULL
		AND meal_plans.id = sqlc.arg(meal_plan_id)
		AND meal_plans.status = 'awaiting_votes'
		AND meal_plans.archived_at IS NULL
		AND meal_plan_events.id = sqlc.arg(meal_plan_event_id)
		AND meal_plan_events.archived_at IS NULL
);

-- name: CheckMealPlanEventExistence :one
SELECT EXISTS (
	SELECT meal_plan_events.id
	FROM meal_plan_events
	WHERE meal_plan_events.archived_at IS NULL
		AND meal_plan_events.id = sqlc.arg(id)
		AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
);

-- name: GetMealPlanEvents :many
SELECT
	meal_plan_events.id,
	meal_plan_events.notes,
	meal_plan_events.starts_at,
	meal_plan_events.ends_at,
	meal_plan_events.meal_name,
	meal_plan_events.belongs_to_meal_plan,
	meal_plan_events.created_at,
	meal_plan_events.last_updated_at,
	meal_plan_events.archived_at,
	(
		SELECT COUNT(meal_plan_events.id)
		FROM meal_plan_events
		WHERE meal_plan_events.archived_at IS NULL
			AND
			meal_plan_events.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND meal_plan_events.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				meal_plan_events.last_updated_at IS NULL
				OR meal_plan_events.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				meal_plan_events.last_updated_at IS NULL
				OR meal_plan_events.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR meal_plan_events.archived_at IS NULL)
			AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
	) AS filtered_count,
	(
		SELECT COUNT(meal_plan_events.id)
		FROM meal_plan_events
		WHERE meal_plan_events.archived_at IS NULL
			AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
	) AS total_count
FROM meal_plan_events
WHERE
	meal_plan_events.archived_at IS NULL
	AND meal_plan_events.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND meal_plan_events.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		meal_plan_events.last_updated_at IS NULL
		OR meal_plan_events.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		meal_plan_events.last_updated_at IS NULL
		OR meal_plan_events.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR meal_plan_events.archived_at IS NULL)
	AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id)
	AND meal_plan_events.id > COALESCE(sqlc.narg(cursor), '')
GROUP BY meal_plan_events.id
ORDER BY meal_plan_events.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetAllMealPlanEventsForMealPlan :many
SELECT
	meal_plan_events.id,
	meal_plan_events.notes,
	meal_plan_events.starts_at,
	meal_plan_events.ends_at,
	meal_plan_events.meal_name,
	meal_plan_events.belongs_to_meal_plan,
	meal_plan_events.created_at,
	meal_plan_events.last_updated_at,
	meal_plan_events.archived_at
FROM meal_plan_events
WHERE
	meal_plan_events.archived_at IS NULL
	AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(meal_plan_id);

-- name: GetMealPlanEvent :one
SELECT
	meal_plan_events.id,
	meal_plan_events.notes,
	meal_plan_events.starts_at,
	meal_plan_events.ends_at,
	meal_plan_events.meal_name,
	meal_plan_events.belongs_to_meal_plan,
	meal_plan_events.created_at,
	meal_plan_events.last_updated_at,
	meal_plan_events.archived_at
FROM meal_plan_events
WHERE meal_plan_events.archived_at IS NULL
	AND meal_plan_events.id = sqlc.arg(id)
	AND meal_plan_events.belongs_to_meal_plan = sqlc.arg(belongs_to_meal_plan);

-- name: UpdateMealPlanEvent :execrows
UPDATE meal_plan_events SET
	notes = sqlc.arg(notes),
	starts_at = sqlc.arg(starts_at),
	ends_at = sqlc.arg(ends_at),
	meal_name = sqlc.arg(meal_name),
	belongs_to_meal_plan = sqlc.arg(belongs_to_meal_plan),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);
