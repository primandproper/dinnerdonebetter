// Package fakes builds the randomized products, subscriptions, purchases and
// ledger rows this application's tests write.
package fakes

import (
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v13/billing"
	"github.com/primandproper/platform-go/v13/capitalism"
	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// currency is the one currency every fake here is priced in. The store checks
// only that a currency is three characters, so any three would do; one is chosen
// so that a test comparing what it wrote to what it read back is not comparing
// two random codes that happen to differ in case.
const currency = "USD"

// BuildFakeProduct builds a faked recurring Product.
//
// Four fields are fixed rather than randomized, and each of them is a value the
// store refuses or a value that would make the row incoherent:
//
//   - The kind, because a random string is not one of the two the store admits.
//     Recurring is chosen because it is the kind with the extra rule — its
//     billing interval is required — so it is the kind worth exercising.
//   - The billing interval, because a recurring product needs a positive one.
//   - The currency and the amount, because the store checks the one and refuses
//     a negative other.
//   - The scope, because this application's catalog is global.
func BuildFakeProduct() *billing.Product {
	product := fake.BuildFakeRecord[billing.Product]()

	product.Kind = billing.KindRecurring
	product.BillingIntervalMonths = int64(gofakeit.Number(1, 12))
	product.Currency = currency
	product.AmountCents = int64(gofakeit.Number(100, 10_000))
	product.Scope = payments.Scope()

	return product
}

// BuildFakeOneTimeProduct builds a faked one-time Product, which is the kind
// with no billing interval and the one a purchase points at.
func BuildFakeOneTimeProduct() *billing.Product {
	product := BuildFakeProduct()

	product.Kind = billing.KindOneTime
	product.BillingIntervalMonths = 0

	return product
}

// BuildFakeProductList builds a faked page of Products.
func BuildFakeProductList() *filtering.QueryFilteredResult[billing.Product] {
	return fake.BuildFakePage(BuildFakeProduct)
}

// BuildFakeSubscription builds a faked active Subscription whose paid period
// covers now.
//
// It takes the account and the product because a subscription is the row joining
// them, and a subscription pointing at rows no test created is one the schema's
// foreign keys refuse. The period covers the present so that the subscription is
// current — which is the read every entitlement check makes, and the shape an
// unqualified "a subscription" means everywhere else.
func BuildFakeSubscription(accountID, productID string) *billing.Subscription {
	subscription := fake.BuildFakeRecord[billing.Subscription]()

	subscription.BelongsToAccount = accountID
	subscription.ProductID = productID
	subscription.Status = capitalism.SubscriptionStatusActive
	subscription.Scope = payments.Scope()

	// Truncated to the second, which is what a TIMESTAMPTZ read back holds
	// through the driver, so a test comparing the period it wrote to the one it
	// read is not defeated by nanoseconds.
	subscription.CurrentPeriodStart = time.Now().UTC().Add(-24 * time.Hour).Truncate(time.Second)
	subscription.CurrentPeriodEnd = subscription.CurrentPeriodStart.AddDate(0, gofakeit.Number(1, 12), 0)

	return subscription
}

// BuildFakeSubscriptionList builds a faked page of Subscriptions.
func BuildFakeSubscriptionList(accountID, productID string) *filtering.QueryFilteredResult[billing.Subscription] {
	return fake.BuildFakePage(func() *billing.Subscription {
		return BuildFakeSubscription(accountID, productID)
	})
}

// BuildFakePurchase builds a faked, not yet completed Purchase.
func BuildFakePurchase(accountID, productID string) *billing.Purchase {
	purchase := fake.BuildFakeRecord[billing.Purchase]()

	purchase.BelongsToAccount = accountID
	purchase.ProductID = productID
	purchase.Currency = currency
	purchase.AmountCents = int64(gofakeit.Number(100, 10_000))
	purchase.CompletedAt = nil
	purchase.Scope = payments.Scope()

	return purchase
}

// BuildFakeTransaction builds a faked succeeded Transaction that names neither a
// subscription nor a purchase, which is the shape a refund of something since
// removed has and the only shape that points at nothing a test has to create.
func BuildFakeTransaction(accountID string) *billing.Transaction {
	transaction := fake.BuildFakeRecord[billing.Transaction]()

	transaction.BelongsToAccount = accountID
	transaction.SubscriptionID = ""
	transaction.PurchaseID = ""
	transaction.Status = billing.TransactionSucceeded
	transaction.Currency = currency
	transaction.AmountCents = int64(gofakeit.Number(100, 10_000))
	transaction.Scope = payments.Scope()

	return transaction
}
