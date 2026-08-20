-- name: CreateOAuth2Client :exec
INSERT INTO oauth2_clients (
	id,
	name,
	description,
	client_id,
	client_secret,
	redirect_uris
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description),
	sqlc.arg(client_id),
	sqlc.arg(client_secret),
	sqlc.arg(redirect_uris)
);

-- name: GetOAuth2ClientByDatabaseID :one
SELECT
	oauth2_clients.id,
	oauth2_clients.name,
	oauth2_clients.description,
	oauth2_clients.client_id,
	oauth2_clients.client_secret,
	oauth2_clients.redirect_uris,
	oauth2_clients.created_at,
	oauth2_clients.archived_at
FROM oauth2_clients
WHERE oauth2_clients.archived_at IS NULL
	AND oauth2_clients.id = sqlc.arg(id);

-- name: ArchiveOAuth2Client :execrows
UPDATE oauth2_clients SET
	archived_at = CURRENT_TIMESTAMP
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: GetOAuth2ClientByClientID :one
SELECT
	oauth2_clients.id,
	oauth2_clients.name,
	oauth2_clients.description,
	oauth2_clients.client_id,
	oauth2_clients.client_secret,
	oauth2_clients.redirect_uris,
	oauth2_clients.created_at,
	oauth2_clients.archived_at
FROM oauth2_clients
WHERE oauth2_clients.archived_at IS NULL
	AND oauth2_clients.client_id = sqlc.arg(client_id);

-- name: GetOAuth2Clients :many
SELECT
	oauth2_clients.id,
	oauth2_clients.name,
	oauth2_clients.description,
	oauth2_clients.client_id,
	oauth2_clients.client_secret,
	oauth2_clients.redirect_uris,
	oauth2_clients.created_at,
	oauth2_clients.archived_at,
	(
		SELECT COUNT(oauth2_clients.id)
		FROM oauth2_clients
		WHERE oauth2_clients.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND oauth2_clients.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR oauth2_clients.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(oauth2_clients.id)
		FROM oauth2_clients
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR oauth2_clients.archived_at IS NULL)
	) AS total_count
FROM oauth2_clients
WHERE oauth2_clients.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND oauth2_clients.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR oauth2_clients.archived_at IS NULL)
	AND oauth2_clients.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY oauth2_clients.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
