-- name: CreatePurchase :exec
INSERT INTO purchases (
	id,
	belongs_to_account,
	product_id,
	amount_cents,
	currency,
	completed_at,
	external_transaction_id
) VALUES (
	sqlc.arg(id),
	sqlc.arg(belongs_to_account),
	sqlc.arg(product_id),
	sqlc.arg(amount_cents),
	sqlc.arg(currency),
	sqlc.arg(completed_at),
	sqlc.arg(external_transaction_id)
);

-- name: GetPurchase :one
SELECT
	purchases.id,
	purchases.belongs_to_account,
	purchases.product_id,
	purchases.amount_cents,
	purchases.currency,
	purchases.completed_at,
	purchases.external_transaction_id,
	purchases.created_at,
	purchases.last_updated_at,
	purchases.archived_at
FROM purchases
WHERE purchases.archived_at IS NULL
	AND purchases.id = sqlc.arg(id);

-- name: GetPurchasesForAccount :many
SELECT
	purchases.id,
	purchases.belongs_to_account,
	purchases.product_id,
	purchases.amount_cents,
	purchases.currency,
	purchases.completed_at,
	purchases.external_transaction_id,
	purchases.created_at,
	purchases.last_updated_at,
	purchases.archived_at,
	(
		SELECT COUNT(purchases.id)
		FROM purchases
		WHERE purchases.archived_at IS NULL
			AND
			purchases.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND purchases.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				purchases.last_updated_at IS NULL
				OR purchases.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				purchases.last_updated_at IS NULL
				OR purchases.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (NOT COALESCE(sqlc.narg(include_archived), false)::boolean OR purchases.archived_at IS NULL)
			AND purchases.belongs_to_account = sqlc.arg(belongs_to_account)
	) AS filtered_count,
	(
		SELECT COUNT(purchases.id)
		FROM purchases
		WHERE purchases.archived_at IS NULL
			AND purchases.belongs_to_account = sqlc.arg(belongs_to_account)
	) AS total_count
FROM purchases
WHERE purchases.archived_at IS NULL
	AND purchases.belongs_to_account = sqlc.arg(belongs_to_account)
	AND purchases.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND purchases.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		purchases.last_updated_at IS NULL
		OR purchases.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		purchases.last_updated_at IS NULL
		OR purchases.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND purchases.belongs_to_account = sqlc.arg(belongs_to_account)
	AND purchases.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY purchases.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
