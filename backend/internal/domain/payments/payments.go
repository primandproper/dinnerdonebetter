/*
Package payments is this application's half of the payments domain: what a
payment provider's webhook means to an account, the namespace the billing
tables carry, the tenancy they are kept under, and the data change events a
write emits.

The stored half is platform-go's. github.com/primandproper/platform-go/v13/billing
owns the catalog, the subscriptions, the one-time purchases and the ledger of
payment attempts: the schema, the paging, the tenancy column, the uniqueness that
turns a redelivered webhook into a collision instead of a second row, and the
guarded status writes that make a replayed event an answer rather than a
failure. capitalism, also platform's, is the wire to Stripe and RevenueCat. This
package neither reimplements nor wraps either.

What it holds is what platform declines to decide:

  - Which of a provider's events changes what about an account. That is
    [PaymentProcessor], which turns a verified delivery into a
    [ParsedWebhookEvent], and the manager, which applies one to the store and to
    the account's billing standing.
  - Which capitalism.SubscriptionStatus leaves an account entitled, which
    internal/entitlements writes down as the plan chooser.
  - The mapping from a subscription's status onto identity.Account's coarse
    billing status, which includes a suspension no processor reports.

# One catalog, in the global scope

Every billing table platform ships carries a tenancy scope, and this application
keeps all four in exactly one: [Scope] is tenancy.Global(). There is one catalog
of products, administered by service admins, and an account's subscriptions,
purchases and ledger rows are filed by account within it — which is what
belongs_to_account is for. Scoping per account would put the catalog out of
reach of the operator who defined it the moment they switched accounts, and would
make the account a fact stated twice in every row.

# The vocabulary that stayed, and the one that went

The subscription statuses this package used to declare are gone.
billing.Subscription.Status is capitalism.SubscriptionStatus — the closed,
documented set every adapter maps its provider's words onto — and a second
enumeration here would have been the same judgment made twice. The one casualty
is a spelling: platform writes "canceled", and the "cancelled" this package
stored is not a value the store accepts.

The product kinds and the transaction statuses went the same way, to
billing.Kind and billing.TransactionStatus.
*/
package payments

import (
	"github.com/primandproper/platform-go/v13/tenancy"
)

// TablePrefix namespaces the platform-go billing tables, rendering
// ddb_billing_products, ddb_billing_subscriptions, ddb_billing_purchases and
// ddb_billing_transactions.
//
// The platform's own default is the empty prefix. The tables this replaced were
// products, subscriptions, purchases and payment_transactions, so the prefix is
// not avoiding a collision with them — it says which application created the
// tables in a database that may hold more than one, which is the same reason
// every other adopted store here carries it.
const TablePrefix = "ddb"

// The data change events a billing write emits. They are declared in the
// webhook event catalog (internal/domain/webhooks/catalog), so a subscriber is
// already able to ask for them.
//
// Purchases and ledger rows emit nothing, as they did before the store was
// adopted: a purchase is written by a checkout flow this application does not
// yet have, and a ledger row is what a provider already told us. Both are
// recorded in the audit log.
const (
	// ProductCreatedServiceEventType indicates a product was added to the catalog.
	ProductCreatedServiceEventType = "product_created"
	// ProductUpdatedServiceEventType indicates a product's name, price, kind,
	// interval or provider-side id changed.
	ProductUpdatedServiceEventType = "product_updated"
	// ProductArchivedServiceEventType indicates a product was withdrawn from sale.
	ProductArchivedServiceEventType = "product_archived"

	// SubscriptionCreatedServiceEventType indicates an agreement was opened.
	SubscriptionCreatedServiceEventType = "subscription_created"
	// SubscriptionUpdatedServiceEventType indicates a subscription's plan,
	// status or paid period moved — whether by an administrative edit or by a
	// provider's event.
	SubscriptionUpdatedServiceEventType = "subscription_updated"
	// SubscriptionArchivedServiceEventType indicates a subscription was retired
	// administratively, which is not a cancellation.
	SubscriptionArchivedServiceEventType = "subscription_archived"
)

// Scope is the tenancy this application keeps its billing under, which is the
// global one. See the package documentation.
func Scope() tenancy.Scope { return tenancy.Global() }
