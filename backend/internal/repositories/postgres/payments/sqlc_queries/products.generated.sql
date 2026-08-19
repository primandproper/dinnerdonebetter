-- name: CreateProduct :exec
INSERT INTO products (
	id,
	name,
	description,
	kind,
	amount_cents,
	currency,
	billing_interval_months,
	external_product_id
) VALUES (
	sqlc.arg(id),
	sqlc.arg(name),
	sqlc.arg(description),
	sqlc.arg(kind),
	sqlc.arg(amount_cents),
	sqlc.arg(currency),
	sqlc.arg(billing_interval_months),
	sqlc.arg(external_product_id)
);

-- name: GetProduct :one
SELECT
	products.id,
	products.name,
	products.description,
	products.kind,
	products.amount_cents,
	products.currency,
	products.billing_interval_months,
	products.external_product_id,
	products.created_at,
	products.last_updated_at,
	products.archived_at
FROM products
WHERE products.archived_at IS NULL
	AND products.id = sqlc.arg(id);

-- name: CheckProductExistence :one
SELECT EXISTS (
	SELECT products.id
	FROM products
	WHERE products.archived_at IS NULL
		AND products.id = sqlc.arg(id)
);

-- name: UpdateProduct :execrows
UPDATE products SET
	name = sqlc.arg(name),
	description = sqlc.arg(description),
	kind = sqlc.arg(kind),
	amount_cents = sqlc.arg(amount_cents),
	currency = sqlc.arg(currency),
	billing_interval_months = sqlc.arg(billing_interval_months),
	external_product_id = sqlc.arg(external_product_id),
	last_updated_at = NOW()
WHERE archived_at IS NULL
	AND id = sqlc.arg(id);

-- name: ArchiveProduct :execrows
UPDATE products SET archived_at = NOW(), last_updated_at = NOW() WHERE archived_at IS NULL AND id = sqlc.arg(id);

-- name: GetProductByExternalID :one
SELECT
	products.id,
	products.name,
	products.description,
	products.kind,
	products.amount_cents,
	products.currency,
	products.billing_interval_months,
	products.external_product_id,
	products.created_at,
	products.last_updated_at,
	products.archived_at
FROM products
WHERE products.archived_at IS NULL
AND products.external_product_id = sqlc.arg(external_product_id);

-- name: GetProducts :many
SELECT
	products.id,
	products.name,
	products.description,
	products.kind,
	products.amount_cents,
	products.currency,
	products.billing_interval_months,
	products.external_product_id,
	products.created_at,
	products.last_updated_at,
	products.archived_at,
	(
		SELECT COUNT(products.id)
		FROM products
		WHERE products.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND products.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				products.last_updated_at IS NULL
				OR products.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				products.last_updated_at IS NULL
				OR products.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR products.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(products.id)
		FROM products
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR products.archived_at IS NULL)
	) AS total_count
FROM products
WHERE products.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND products.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		products.last_updated_at IS NULL
		OR products.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		products.last_updated_at IS NULL
		OR products.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR products.archived_at IS NULL)
	AND products.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY products.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);

-- name: SearchForProducts :many
SELECT
	products.id,
	products.name,
	products.description,
	products.kind,
	products.amount_cents,
	products.currency,
	products.billing_interval_months,
	products.external_product_id,
	products.created_at,
	products.last_updated_at,
	products.archived_at,
	(
		SELECT COUNT(products.id)
		FROM products
		WHERE products.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
			AND products.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
			AND (
				products.last_updated_at IS NULL
				OR products.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
			)
			AND (
				products.last_updated_at IS NULL
				OR products.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
			)
			AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR products.archived_at IS NULL)
	) AS filtered_count,
	(
		SELECT COUNT(products.id)
		FROM products
		WHERE (COALESCE(sqlc.narg(include_archived), false)::boolean OR products.archived_at IS NULL)
	) AS total_count
FROM products
WHERE products.created_at > COALESCE(sqlc.narg(created_after), (SELECT NOW() - '999 years'::INTERVAL))
	AND products.created_at < COALESCE(sqlc.narg(created_before), (SELECT NOW() + '999 years'::INTERVAL))
	AND (
		products.last_updated_at IS NULL
		OR products.last_updated_at > COALESCE(sqlc.narg(updated_after), (SELECT NOW() - '999 years'::INTERVAL))
	)
	AND (
		products.last_updated_at IS NULL
		OR products.last_updated_at < COALESCE(sqlc.narg(updated_before), (SELECT NOW() + '999 years'::INTERVAL))
	)
	AND (COALESCE(sqlc.narg(include_archived), false)::boolean OR products.archived_at IS NULL)
	AND products.name ILIKE '%' || sqlc.arg(name_query)::text || '%'
	AND products.id > COALESCE(sqlc.narg(cursor), '')
ORDER BY products.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
