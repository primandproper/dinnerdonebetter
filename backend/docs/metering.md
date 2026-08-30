# Usage metering

We count what accounts consume. We do not yet limit it, and we do not yet bill for it.

That ordering is deliberate. A limit set before there is usage data to set it from is a guess,
and the guess that is too low is an outage for a customer who did nothing wrong. So the counting
goes in first, the totals feed dashboards, and the limits get set from what the dashboards say.
Turning one on later is an edit to `planLimits` rather than a migration.

The machinery is `platform-go/v13/metering`; this repo supplies the three things that package
deliberately refuses to model. Read its package documentation for how the guarantees work — this
document is only about the decisions that are ours.

## What is counted

| Meter                  | Unit  | Aggregation | Period         | Recorded at                 |
|------------------------|-------|-------------|----------------|-----------------------------|
| `uploaded_media_bytes` | bytes | sum         | calendar month | `uploadedmedia/grpc.Upload` |

One meter, on purpose. It is the one that maps most directly onto a bill we actually receive —
object storage charges for what is held, and the upload endpoint is the only place anything gets
held — and one meter proves the whole path end to end without spreading half-wired call sites
across the codebase.

It counts **bytes accepted, not bytes resident**. A sum over the period answers "how much did
this account add in March", which is what a write site can honestly observe. Residency is a
different question: it needs a periodic sweep of the bucket rather than a record at the write
site, and it would be `AggregationLast` over a gauge. Deleting media does not decrement this
meter.

## The subject is the account

Everything else here is scoped to an account, subscriptions belong to accounts, and an invoice
goes to an account rather than to whoever happened to be signed in. Usage is recorded against
`sessionContextData.GetActiveAccountID()`.

## The period is the calendar month

UTC calendar month, not the subscription's billing anchor. The anchor is more correct for an
invoice line, and `metering.PeriodBillingPeriod` exists for exactly that — but it is answerable
only for an account that holds a subscription, and most accounts do not. A resolver would have
to fall back to the calendar for everyone else, so an account's bucketing would change on the
day it subscribed and split one month's usage across two rows keyed on different period starts.

Switching later means passing a `metering.PeriodResolver` to both the recorder and the enforcer.
It is not a schema change, but it does change what every stored total means, so it belongs at a
period boundary.

## Idempotency keys

The key is the created row's ID — for uploads, the `uploaded_media` ID.

Not a request ID, despite that being the usual advice, because a request ID is not stable here.
A client that retries a timed-out upload sends the bytes again, gets a new ID, and stores a
second object. That is genuinely new usage, and a request-scoped key would have deduped it away
into an object nobody is charged for.

Dedupe in the ledger is keyed on `(meter, idempotency_key)`, so one row's ID feeding several
meters records against each of them.

## Where each piece runs

| Component | Process | Wired in |
| --- | --- | --- |
| `Recorder` | API server | `internal/build/services/api/grpc/build.go` |
| `Enforcer` | API server | registered, consulted by nothing yet |
| `Flusher` | scheduler | `internal/build/jobs/scheduler/metering.go`, job `metering_flusher` |

The flusher is in the scheduler because a flush is a scheduled pass over a backlog under a
lease, and because the credentials it posts usage with are not credentials a request path should
hold. It runs every five minutes.

The tables come from the platform's DDL, rendered at migration version 30 in
`internal/repositories/postgres/migrations/migrate.go` rather than copied into a numbered file —
same arrangement as outbox, saga, webhooks, and audit. They are therefore invisible to `sqlc`;
do not reference them from generated queries.

## Recording failures are swallowed

`recordUploadUsage` logs and continues. By the time it runs the file is in the bucket and its row
is in the database, so failing the call would tell a client an upload did not happen that did.
Nothing enforces this meter, so an uncounted record costs a gap in a dashboard rather than a
wrong invoice.

If a meter ever gates something, that call site should use `Enforcer.ConsumeUsage` instead, which
decides and records in one transaction against the true total.

## Nothing is billed yet

`NewProviderMapper` reports no provider reference for anything, and the capitalism provider is
`noop` in every environment. The flusher reads that as "nothing to post" and settles each total
rather than re-claiming it every interval forever.

The consequence to be clear about: **usage settled this way is not retro-billable.** The durable
`quantity` is untouched — the totals stay complete and the dashboards see every byte — but
`flushed_quantity` advances with it, so a real mapper wired in later starts from the totals as
they stand rather than replaying the months before it existed. That is the intended shape of
"count now, bill later".

The flush pass still does real work meanwhile: it keeps the backlog gauge honest and reaps
event-ledger rows past their ninety-day retention.

## What has to change before the first limit goes on

1. **Put the limit in `planLimits`** (`internal/metering/plans.go`), keyed by meter and by
   product ID. The quota source short-circuits to unlimited for any meter absent from that map,
   without reading anything — so an entry appearing there is also what turns the subscription
   lookup on. `Behavior` is required: platform's `NewPlanLimitSource` refuses a zero one at
   construction rather than picking a default, and it refuses a meter the registry does not know.
2. **Attach a cache to the enforcer.** It is built with a nil `cache.Cache[metering.CachedTotal]`
   today, so `Check` reads the durable total on every call. That is a durable read on a request
   path, which is the thing `Check` exists not to be.
3. **Decide `EnforcerConfig.FailOpen`.** It is false — refuse when the store is unreachable —
   which is right for a quota that guards spend and wrong for one protecting a shared dependency
   from a noisy neighbor. Nobody has been asked which this is.
4. **Call it.** `Check` in front of something cheap, `ConsumeUsage` in front of anything whose
   overage is worth more than a write.

## What has to change before billing

1. **Implement `ProviderMapper`** against `identity.Account.PaymentProcessorCustomerID` and the
   provider-side meter name — which does not exist yet, and which whoever owns pricing has to
   name at the provider first.
2. **Set the capitalism provider to `stripe`** in the scheduler's config. It is taken from the
   payments service's own capitalism config, so there is one provider setting rather than two
   that can disagree.
