-- name: CreatePaymentTransaction :exec
INSERT INTO payment_transactions (
	id,
	belongs_to_account,
	subscription_id,
	purchase_id,
	external_transaction_id,
	amount_cents,
	currency,
	status
) VALUES (
	sqlc.arg(id),
	sqlc.arg(belongs_to_account),
	sqlc.arg(subscription_id),
	sqlc.arg(purchase_id),
	sqlc.arg(external_transaction_id),
	sqlc.arg(amount_cents),
	sqlc.arg(currency),
	sqlc.arg(status)
);

-- name: GetPaymentTransactionsForAccount :many
SELECT
	payment_transactions.id,
	payment_transactions.belongs_to_account,
	payment_transactions.subscription_id,
	payment_transactions.purchase_id,
	payment_transactions.external_transaction_id,
	payment_transactions.amount_cents,
	payment_transactions.currency,
	payment_transactions.status,
	payment_transactions.created_at,
	(
		SELECT COUNT(payment_transactions.id)
		FROM payment_transactions
		WHERE payment_transactions.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
			AND payment_transactions.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
			AND payment_transactions.belongs_to_account = sqlc.arg(belongs_to_account)
	) AS filtered_count,
	(
		SELECT COUNT(payment_transactions.id)
		FROM payment_transactions
		WHERE payment_transactions.belongs_to_account = sqlc.arg(belongs_to_account)
	) AS total_count
FROM payment_transactions
WHERE payment_transactions.created_at > COALESCE(sqlc.narg(created_after), (SELECT CURRENT_TIMESTAMP - '999 years'::INTERVAL))
	AND payment_transactions.created_at < COALESCE(sqlc.narg(created_before), (SELECT CURRENT_TIMESTAMP + '999 years'::INTERVAL))
	AND payment_transactions.belongs_to_account = sqlc.arg(belongs_to_account)
	AND payment_transactions.belongs_to_account = sqlc.arg(belongs_to_account)
	AND payment_transactions.id > COALESCE(sqlc.narg(page_cursor), '')
ORDER BY payment_transactions.id ASC
LIMIT COALESCE(sqlc.narg(result_limit), 50);
