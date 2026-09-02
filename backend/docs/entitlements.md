# Entitlements

What an account may use, and how much of it is left.

Four things already hold pieces of that answer. `capitalism` knows what the account bought,
`metering` knows what it has consumed, `featureflags` knows about rollouts and overrides, and
`authorization` owns the yes/no seam every handler already gates on. The machinery that joins
them is `platform-go/v13/entitlements`; `internal/entitlements` supplies the three things that
package deliberately refuses to model. Read its package documentation for how the decision is
made — this document is only about the decisions that are ours.

Nothing calls it yet. Read `metering.md` first: it records that nothing is limited or billed,
so this is about having the right shape in place before that changes rather than about a live
gate.

## The three seams

| Seam | Where | Why there |
| --- | --- | --- |
| Features | `internal/entitlements/features.go` | A quota feature names a meter this application registered; a boolean one names a permission its handlers check. Neither is spellable in an environment variable, and a catalog that discovered its features from configuration would let a typo silently create a feature nobody gates on. |
| Plans | `internal/config.DefaultEntitlementsConfig`, rendered into each environment's config | What a tier includes changes when pricing changes, which is more often than a deploy and by people who do not ship one. |
| `PlanSource` | `internal/entitlements/plans.go` | The join between an account and a purchased plan is application data — it is this service's `subscriptions` table. |

## What is declared

One feature, `uploaded_media_bytes`, of kind quota, counting against the meter of the same name.
It is the only thing metered, so it is the only thing there is a quantity to gate on.

The feature key and the meter name are the same string and are deliberately two declarations. A
meter is named by whoever sets billing up and travels into provider-side idempotency keys; a
feature is named by whoever writes the gate. They agree until the first time one of them is
renamed.

## The plans

Two: `free` and `subscriber`. Both grant every feature without a bound.

That is the same product decision `metering.md` records — count first, limit once the dashboards
say what real usage looks like — expressed where a limit will go when there is one. Turning the
first limit on is replacing a grant's `unlimited: true` with a `limit` and a `behavior` in the
rendered config, and `behavior` is worth choosing rather than inheriting: the platform defaults a
quota grant to `block`, on the grounds that a gate which lets everything through is a decoration.

`free` is a plan rather than an absence. The core of this product is free, so an account that has
never paid is a customer of the free tier and entitled to what that tier includes; the platform's
other answer — reporting no plan at all — is for a product where not paying means not being a
customer, and here it would deny an unsubscribed account features it is supposed to have.

`subscriber` is one plan for every product, because there is one tier to sell. The day there are
two, the mapping goes in `SubscriptionPlanSource`: entitlements joins a plan to a catalog entry
by string equality, and a product ID cannot be that string — plan names are plain identifiers and
a minted ID is not one.

## Which plan an account is on

`SubscriptionPlanSource` reads the account's subscriptions and answers `subscriber` if any of
them is `active` or `trialing`, and `free` otherwise.

Trialing counts: a trial that silently got the free plan's grants would be a trial of a different
product. `past_due` does not, which is the lever that makes a lapsed payment degrade service
rather than end it — though with both plans granting the same thing it currently levers nothing.
An account can hold several rows, so the first live one wins rather than the only one.

It reads this service's own table rather than asking the payment provider. A provider round trip
per feature check spends a latency budget on a fact that changes a few times a year, and an
outage there would take the product down rather than the billing. The provider says when a
subscription changes; what it changed to belongs in the database by the time anybody asks.

## One limit, not two

`internal/entitlements.RegisterQuotaSource` registers the catalog's quota source twice: once as
itself, and once as `metering.QuotaSource`, which is what the enforcer resolves.

That second registration is the point of the arrangement. metering asks a quota source what a
subject may consume and enforces whatever it is told; the catalog is what knows. Wired this way
there is exactly one number, and the limit an account is shown is by construction the limit
enforced against it. Wired any other way there are two, and they will disagree — the platform's
checker notices when they do and counts it under `entitlements_limit_mismatches`.

## Failure is a degraded plan, not an open door

`CheckerConfig.FallbackPlan` is `free`. A payments database that cannot be reached therefore
degrades a paying account to the free tier rather than locking it out, and the degraded state is
a tier somebody chose and can read in the catalog rather than "everything".

It applies only to boolean features. A quota feature's limit reaches metering through the quota
source, which has no fallback, so a quota check during a payments outage errors rather than
guessing. That asymmetry is deliberate: the same source answers the exact path that records
consumption, and enforcing a guessed limit there writes usage against a plan the customer is not
on — an error during an outage costs less than a reconciliation after one.

## What has to change before the first gate

1. **Attach a cache.** No `cache.Cache[entitlements.Assignment]` is registered, so every check
   resolves the account's plan through the payments repository — a durable read on a request
   path, which is what the cache exists to avoid. Thirty seconds is the platform's default and
   the reasoning for it is in `CheckerConfig.CacheTTL`.
2. **Invalidate on subscription change.** The webhook handler that writes a subscription change
   should call `Checker.Invalidate`, so a customer who has just upgraded is not told for another
   TTL that they have not.
3. **Call it.** `Check` in front of something cheap, `CheckQuantity` in front of anything whose
   cost is known before it runs. A boolean feature can also be merged into a session's permission
   set with `Checker.Permissions`; quota features deliberately cannot, because a permission that
   meant "was entitled a moment ago" would be read by every caller as "may proceed".
