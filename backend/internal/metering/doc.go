/*
Package metering names what this application counts, and fills the three seams
the platform's metering package leaves to its consumer.

The platform owns the hard parts — idempotent ingest, the concurrent fold into a
period total, the idempotent push to a billing provider. What it deliberately
does not own is anything that would duplicate a billing provider's product
catalog, so it asks an application three questions and this package answers them:

  - Registry: which meters exist, and how do their records combine.
  - QuotaSource: what is this account allowed on this meter.
  - ProviderMapper: which provider-side customer and meter does this account's
    usage post against.

# Count now, enforce and bill later

Nothing is enforced. Every quota this package hands out is BehaviorAllowOverage
against a limit nobody reaches, which is how the platform spells "unlimited" —
it refuses to treat an absent quota as an unlimited one, because unmetered and
unlimited are different facts. Nothing is billed either: NewProviderMapper
reports no provider reference for anything, so the flusher settles totals without
posting them.

That ordering is the point. The counting has to exist and be trustworthy before
any limit can be set from data rather than from a guess, and a limit that arrives
later is then a change to planLimits rather than a migration. The consequence to
be clear about is that usage counted before a real ProviderMapper is wired is not
retro-billable: the flusher marks those totals flushed as it settles them. The
durable quantity is untouched, so the dashboards this exists to feed keep seeing
every byte.

# The subject is the account

Everything else in this service is scoped to an account, subscriptions are held
by accounts, and an invoice goes to an account rather than to whichever member of
it happened to be signed in. So metering subjects are account IDs.

# The period is the calendar month

PeriodMonth, in UTC, rather than the subscription's own billing anchor.

The anchor is more correct for an invoice line and it is what
metering.PeriodBillingPeriod exists for, but it is answerable only for an account
that holds a subscription — and most accounts do not. A resolver would therefore
have to fall back to the calendar for everyone else, which means an account's
bucketing would change on the day it subscribed, splitting one month's usage
across two rows keyed on different period starts. A uniform calendar month is
worse for exactly one thing we do not do yet and better for everything we do.

Switching later is a PeriodResolver passed to both the recorder and the enforcer;
it is not a schema change. It is, however, a change to what every stored total
means, so it belongs at a period boundary and with the old rows understood.
*/
package metering
