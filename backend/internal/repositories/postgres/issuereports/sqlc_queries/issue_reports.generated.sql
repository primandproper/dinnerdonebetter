-- name: CreateIssueReport :exec
INSERT INTO issue_reports (
	id,
	issue_type,
	details,
	relevant_table,
	relevant_record_id,
	created_by_user,
	belongs_to_account
) VALUES (
	sqlc.arg(id),
	sqlc.arg(issue_type),
	sqlc.arg(details),
	sqlc.arg(relevant_table),
	sqlc.arg(relevant_record_id),
	sqlc.arg(created_by_user),
	sqlc.arg(belongs_to_account)
);

-- name: GetIssueReport :one
SELECT
	issue_reports.id,
	issue_reports.issue_type,
	issue_reports.details,
	issue_reports.relevant_table,
	issue_reports.relevant_record_id,
	issue_reports.created_at,
	issue_reports.last_updated_at,
	issue_reports.archived_at,
	issue_reports.created_by_user,
	issue_reports.belongs_to_account
FROM issue_reports
WHERE issue_reports.archived_at IS NULL
	AND issue_reports.id = sqlc.arg(id);

-- name: UpdateIssueReport :execrows
UPDATE issue_reports SET
	issue_type = sqlc.arg(issue_type),
	details = sqlc.arg(details),
	relevant_table = sqlc.arg(relevant_table),
	relevant_record_id = sqlc.arg(relevant_record_id),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveIssueReport :execrows
UPDATE issue_reports SET
	last_updated_at = NOW(),
	archived_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: CheckIssueReportExistence :one
SELECT EXISTS(
	SELECT issue_reports.id
	FROM issue_reports
	WHERE issue_reports.archived_at IS NULL
		AND issue_reports.id = sqlc.arg(id)
);

-- name: GetIssueReports :many
SELECT
	issue_reports.id,
	issue_reports.issue_type,
	issue_reports.details,
	issue_reports.relevant_table,
	issue_reports.relevant_record_id,
	issue_reports.created_at,
	issue_reports.last_updated_at,
	issue_reports.archived_at,
	issue_reports.created_by_user,
	issue_reports.belongs_to_account,
	(
		SELECT COUNT(issue_reports.id)
		FROM issue_reports
		WHERE issue_reports.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND issue_reports.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				issue_reports.last_updated_at IS NULL
				OR issue_reports.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				issue_reports.last_updated_at IS NULL
				OR issue_reports.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR issue_reports.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(issue_reports.id)
		FROM issue_reports
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR issue_reports.archived_at IS NULL)
	) AS total_count
FROM issue_reports
WHERE issue_reports.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND issue_reports.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		issue_reports.last_updated_at IS NULL
		OR issue_reports.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		issue_reports.last_updated_at IS NULL
		OR issue_reports.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR issue_reports.archived_at IS NULL)
	AND issue_reports.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY issue_reports.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetIssueReportsForAccount :many
SELECT
	issue_reports.id,
	issue_reports.issue_type,
	issue_reports.details,
	issue_reports.relevant_table,
	issue_reports.relevant_record_id,
	issue_reports.created_at,
	issue_reports.last_updated_at,
	issue_reports.archived_at,
	issue_reports.created_by_user,
	issue_reports.belongs_to_account,
	(
		SELECT COUNT(issue_reports.id)
		FROM issue_reports
		WHERE issue_reports.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND issue_reports.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				issue_reports.last_updated_at IS NULL
				OR issue_reports.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				issue_reports.last_updated_at IS NULL
				OR issue_reports.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR issue_reports.archived_at IS NULL)
			AND issue_reports.belongs_to_account = sqlc.arg(belongs_to_account)
	) AS filtered_count,
	(
		SELECT COUNT(issue_reports.id)
		FROM issue_reports
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR issue_reports.archived_at IS NULL)
			AND issue_reports.belongs_to_account = sqlc.arg(belongs_to_account)
	) AS total_count
FROM issue_reports
WHERE issue_reports.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND issue_reports.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		issue_reports.last_updated_at IS NULL
		OR issue_reports.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		issue_reports.last_updated_at IS NULL
		OR issue_reports.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR issue_reports.archived_at IS NULL)
	AND issue_reports.belongs_to_account = sqlc.arg(belongs_to_account)
	AND issue_reports.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY issue_reports.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetIssueReportsForTable :many
SELECT
	issue_reports.id,
	issue_reports.issue_type,
	issue_reports.details,
	issue_reports.relevant_table,
	issue_reports.relevant_record_id,
	issue_reports.created_at,
	issue_reports.last_updated_at,
	issue_reports.archived_at,
	issue_reports.created_by_user,
	issue_reports.belongs_to_account,
	(
		SELECT COUNT(issue_reports.id)
		FROM issue_reports
		WHERE issue_reports.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND issue_reports.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				issue_reports.last_updated_at IS NULL
				OR issue_reports.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				issue_reports.last_updated_at IS NULL
				OR issue_reports.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR issue_reports.archived_at IS NULL)
			AND issue_reports.relevant_table = sqlc.arg(relevant_table)
	) AS filtered_count,
	(
		SELECT COUNT(issue_reports.id)
		FROM issue_reports
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR issue_reports.archived_at IS NULL)
			AND issue_reports.relevant_table = sqlc.arg(relevant_table)
	) AS total_count
FROM issue_reports
WHERE issue_reports.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND issue_reports.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		issue_reports.last_updated_at IS NULL
		OR issue_reports.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		issue_reports.last_updated_at IS NULL
		OR issue_reports.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR issue_reports.archived_at IS NULL)
	AND issue_reports.relevant_table = sqlc.arg(relevant_table)
	AND issue_reports.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY issue_reports.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: GetIssueReportsForRecord :many
SELECT
	issue_reports.id,
	issue_reports.issue_type,
	issue_reports.details,
	issue_reports.relevant_table,
	issue_reports.relevant_record_id,
	issue_reports.created_at,
	issue_reports.last_updated_at,
	issue_reports.archived_at,
	issue_reports.created_by_user,
	issue_reports.belongs_to_account,
	(
		SELECT COUNT(issue_reports.id)
		FROM issue_reports
		WHERE issue_reports.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND issue_reports.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				issue_reports.last_updated_at IS NULL
				OR issue_reports.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				issue_reports.last_updated_at IS NULL
				OR issue_reports.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR issue_reports.archived_at IS NULL)
			AND issue_reports.relevant_table = sqlc.arg(relevant_table)
			AND issue_reports.relevant_record_id = sqlc.arg(relevant_record_id)
	) AS filtered_count,
	(
		SELECT COUNT(issue_reports.id)
		FROM issue_reports
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR issue_reports.archived_at IS NULL)
			AND issue_reports.relevant_table = sqlc.arg(relevant_table)
			AND issue_reports.relevant_record_id = sqlc.arg(relevant_record_id)
	) AS total_count
FROM issue_reports
WHERE issue_reports.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND issue_reports.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		issue_reports.last_updated_at IS NULL
		OR issue_reports.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		issue_reports.last_updated_at IS NULL
		OR issue_reports.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR issue_reports.archived_at IS NULL)
	AND issue_reports.relevant_table = sqlc.arg(relevant_table)
	AND issue_reports.relevant_record_id = sqlc.arg(relevant_record_id)
	AND issue_reports.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY issue_reports.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
