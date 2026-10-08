/*
Package payments is this application's half of the payments domain: where a
payment provider's subscription lands among this application's accounts and
products, and the tenancy the billing tables are kept under.

The stored half is platform-go's. github.com/primandproper/platform-go/v15/billing
owns the catalog, the subscriptions, the one-time purchases and the ledger of
payment attempts: the schema, the paging, the tenancy column, the uniqueness that
turns a redelivered webhook into a collision instead of a second row, and the
guarded status writes that make a replayed event an answer rather than a
failure. capitalism, also platform's, is the wire to Stripe and RevenueCat, and
billing/sync and billing/http are the webhook endpoint between the two. This
package neither reimplements nor wraps any of them.

What it holds is what platform declines to decide:

  - Which account and which product a subscription nobody holds yet belongs to.
    That is [StripePlace] and [RevenueCatPlace], the billing/sync Place for each
    provider; see placement.go.
  - Which capitalism.SubscriptionStatus leaves an account entitled, which
    internal/entitlements writes down as the plan chooser.

# One catalog, in the global scope

Every billing table platform ships carries a tenancy scope, and this application
keeps all four in exactly one, tenancy.Global(). There is one catalog
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
