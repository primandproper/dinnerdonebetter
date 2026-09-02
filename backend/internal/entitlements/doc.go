/*
Package entitlements says what an account may use, and fills the three seams the
platform's entitlements package leaves to its consumer.

The platform owns the hard parts — the catalog, the decision, the plan cache,
the arrangement that makes the limit an account is shown the limit enforced
against it. What it deliberately does not own is anything that would duplicate a
billing provider's product catalog or an application's own subscription table,
so it asks three questions and this package answers them:

  - Features: which features exist and what kind each one is. Code, because a
    quota feature names a meter this application registered.
  - Plans: what each plan includes. Configuration, because pricing changes more
    often than a deploy — see internal/config.DefaultEntitlementsConfig for the
    shipped catalog.
  - PlanSource: which plan an account is on. See NewSubscriptionPlanSource.

# What this replaced

The gate used to be a table of limits keyed by meter and by product ID, read
through metering's own NewPlanLimitSource. That answered "how much of this meter
may this account consume" and nothing else: a boolean feature — an export
button, a support tier, anything an account either has or does not — had nowhere
to go, and the limit an account would have been shown lived in a different place
from the limit enforced against it.

The catalog answers both kinds with one call, and the QuotaSource this package
hands metering is the same catalog, so there is exactly one number.

# Nothing is limited yet

Both shipped plans grant every feature without a bound. That is the same product
decision internal/metering documents at length: a limit set before there is
usage data to set it from is a guess, and the guess that is too low is an outage
for a customer who did nothing wrong. The counting exists first.

What has changed is what it costs to turn one on. It used to be an edit to a Go
map and a deploy; it is now an edit to the plan's grant in the rendered config.

# The subject is the account

Subscriptions belong to accounts, an invoice goes to an account rather than to
whoever happened to be signed in, and metering records usage against the account
too. So an entitlement is an account's, and every account has one — see
NewSubscriptionPlanSource for why there is a free plan rather than an absence.
*/
package entitlements
