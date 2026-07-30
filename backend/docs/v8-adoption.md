# platform-go v8 adoption

What v8's new packages replaced, what still runs the old way, and what has to exist in the
environment before any of it works.

v8 added four packages over v7.1.1: `jobs`, `outbox`, `idempotency`, and `authorization`. All
four are now in use, along with `distributedlock` and `retry`, which v8 made worth reaching for.

## Background work

### Queue consumers — `jobs.Pool`

`cmd/functions/async_message_handler` used to build six bare `messagequeue.Consumer`s sharing a
stop channel. Each topic is now a `jobs.Pool` (`internal/functions/datachangemessagehandler/pools.go`).

The event handlers themselves did not change: `jobs.Handler` has the signature the
`XxxEventHandler(topic)` factories already returned. What changed is everything around them.

| Before | After |
|---|---|
| Serial per topic | Bounded concurrency, per topic |
| A failure was logged and the message was gone | Retry with backoff, then dead-lettered |
| A panicking handler killed the pod | Contained as an ordinary attempt failure |

Pool shapes live in `config.WorkerPoolsConfig` and differ per topic because the work does.
`user_data_aggregation` runs at concurrency 1: a crash can lose `Concurrency+1` in-flight
messages, and a duplicated GDPR export is worse than a slow one.

Decode failures are marked `retry.Unretryable`. A payload that will not parse will not parse
three more times, and each of those attempts is latency the healthy messages behind it spend
waiting.

**One semantic shift worth knowing.** `Pool`'s handler returns as soon as it hands the payload
to a worker, so with Pub/Sub the message is acknowledged before it has been processed. Redelivery
is no longer the safety net; retry and dead-lettering are. `Concurrency` is what bounds the
exposure.

### Periodic work — `jobs.Scheduler`

Six Kubernetes CronJobs became one long-lived `scheduler` Deployment
(`cmd/workers/scheduler`, `internal/build/jobs/scheduler/`). Every replica ticks; the one that
wins a Postgres advisory lock runs the job and the rest skip.

Two CronJobs remain, and should: `db_cleaner` (`0 0 1,8,15,22 * *`) and
`email_deliverability_test` (`0 0 * * *`). `jobs.Scheduler` takes intervals, not cron
expressions, and an interval that starts whenever the pod last restarted is not "midnight UTC".

Postgres advisory locks rather than Redis: no new infrastructure, and a replica that dies drops
its connection, which releases the lock without waiting for a TTL to lapse.

**What this trades away.** Six isolated pods became one, so an OOM in the task creator now stops
search indexing too. Mitigated with per-job `Timeout` and `Recreate` rollout; the operational
simplicity and the removal of per-tick cold starts (two of these ran every 60 seconds and paid
full DI construction each time) are what buys it.

**Least privilege changed.** Those six jobs had a Postgres user each. They now share one
(`scheduler`) with the union of their grants. Separation between the six is gone; separation
from the API server's user is not.

## Events — `outbox`

Writing a row and publishing an event were two operations against two systems with no shared
commit. The row lands, the publish fails, and durable state and the event stream diverge
permanently with nothing to detect it.

The seam is `internal/repositories/postgres/events`. `outbox.Enqueue` takes the
`database.SQLQueryExecutor` that `WithTransaction` hands its callback, so the event is another
statement in the transaction that wrote the row. That executor only exists inside the
repository, which is why events are emitted there rather than from the manager — the same place,
and for the same reason, that audit log entries already are.

The relay runs in the scheduler process (`internal/build/jobs/scheduler/outbox.go`). Its table
comes from `outbox/migrations` rather than a hand-copied DDL file, registered at migration
version 22 in `internal/repositories/postgres/migrations/migrate.go`.

**Wire compatibility means this migrates incrementally.** The relay republishes the stored bytes
as `json.RawMessage`, so what reaches the broker is byte-identical to a direct publish of the
same value. Consumers need no change, and a domain can be converted one method at a time.

### What is converted

Meal plan create, update, archive, and finalize. Update and archive also gained the transaction
they never had — the row and its audit log entry were previously two independent writes.

The finalizer job's publish is gone: the event is enqueued inside
`AttemptToFinalizeMealPlan`'s transaction, which closes a real gap. A publish failure there used
to leave a finalized meal plan that no consumer heard about, with a log line as the only
evidence.

Beyond meal plans: the whole meal planning catalog (`valid_*`), meal-plan child entities, the
full recipe tree including step children, and the comments, issue reports, webhooks, settings,
notifications, waitlists, payments and identity domains. **127 of ~150 publish sites converted.**

Many of those repository methods gained the transaction they never had, so a write and its audit
log entry can no longer half-commit. Seven managers (`comments`, `issuereports`, `settings`,
`notifications`, `payments`, `waitlists`, `webhooks`) no longer construct a data-changes
publisher at all.

Five repository signatures were widened so the events could keep their full payload — the recipe
step children (`CreateRecipeStepProduct` and friends) now take the `recipeID` their events name,
rather than dropping the field. That was a `Repository` interface change and a moq regeneration.

### What is not, and why

23 sites still publish after commit, for two reasons:

**The repository cannot reproduce the payload (12 sites).** `AcceptAccountInvitation` names the
invitation's destination account, which its repository method never sees; several identity
methods (`ArchiveUserMembership`, `SetDefaultAccount`, `UpdateAccountMemberPermissions`,
`AdminUpdateUserStatus`) are similar. Emitting from the repository would silently drop fields
webhook subscribers receive. The recipe step children were fixed by widening signatures; these
need the same treatment, decided case by case.

**No single backing repository method (11 sites).** `CreateMealPlanOptionWithEventID`,
`CreateMealPlanOptionVotes`, `MealPlanTaskStatusChange`, `CreateUser`,
`CreateAccountInvitation`, `RejectAccountInvitation`, `CancelAccountInvitation` and friends
summarize several writes (`"vote_count"`, `"created"`) or compose calls the manager makes in
sequence. Each needs a decision about what the atomic unit is before it can move.

Because the relay is wire-compatible, the converted and unconverted paths coexist correctly: a
consumer cannot tell which mechanism delivered a message.

## Payments — `idempotency`

A client that sends a purchase, never sees the response, and retries is indistinguishable from a
client making a second deliberate purchase, unless it supplies a key.

`internal/build/services/api/grpc/idempotency.go` installs the server interceptor for the
subscription mutations. `pkg/client/idempotency.go` is the client half.

The key is scoped to the authenticated principal. Without that, two users who minted the same
key would collide and the second would be handed the first one's recorded response. With it,
that reuse reads as a fingerprint mismatch and is refused.

The client mints the key **once, outside its retry loop**:

```go
ctx, _ := client.NewIdempotencyContext(ctx)   // once
err := policy.Do(ctx, func(ctx context.Context) error {
    _, sendErr := c.CreateSubscription(ctx, req)
    return sendErr                             // every attempt carries the same key
})
```

A key minted inside the loop is a new key per attempt. It looks like protection and provides
none, and nothing on the server can detect the mistake.

**It is off in production.** `IdempotencyConfig.Enabled` is false there because prod has no
shared record store. The memory cache provider is per-process: with several replicas, a retry
that lands on a different pod re-executes and two concurrent requests can both claim the same
key — the exact failure this prevents. Shipping that would be worse than shipping nothing.
Provision Redis, point `Manager.Cache` at it, and flip the flag. Localdev has Redis and runs
with it enabled.

## Authorization — adopted in audit-only mode

v8's `authorization` package models exactly what `internal/authorization/` hand-rolls: roles
resolve to a permission set once, and checking is a map lookup that cannot fail. The gRPC
adapter's model also matches the existing `AuthInterceptor` exactly — declared methods require
permissions, `Public` methods opt out, and an undeclared method is denied.

`internal/authorization/platform.go` bridges the two. The policy is built from the same
permission slices the checkers are constructed from, so the two tables cannot drift, and
`internal/build/services/api/grpc/authorization.go` builds the requirements from the same
aggregated method map and the interceptor's own public-route list. Nothing is spelled twice.

**It runs with `WithAuditOnly()`, and denies nothing.** The existing interceptor is still what
refuses requests. That is the package's own documented migration path, and the reason is worth
stating plainly: turning enforcement on across a service that already has a large hand-written
permission table is a coin flip on whether the two tables agree. Audit mode turns the coin flip
into a measurement.

**To finish the migration:** deploy, watch `authorization_denied` for methods the enforcer would
have refused but the current interceptor allows. Each one is either a policy bug in
`PlatformPolicy()` or a gap in the current table — both worth knowing before either enforces.
Once it holds at zero under real traffic, drop `WithAuditOnly()` and delete the permission-check
half of `AuthInterceptor`.

## Not adopted

- **`eventstream`** — SSE/WebSocket. Live meal-plan voting is the obvious candidate, but that is
  a product decision.
- **`eventcapture`**, **`ratelimiting`**, **`compression`**, **`llm`**, **`embeddings`**,
  **`capitalism`**, **`bitmask`** — no current need.

**Circuit breaking needs no work.** The `circuitbreakingcfg.Config` values the codegen emits are
consumed by the platform packages that own them (`analytics/config`, `uploads/objectstorage`,
`distributedlock/config`), which construct the breakers themselves.

**The webhook executor needs no inner retry.** It now runs under a pool with four attempts,
backoff, a 30-second handler timeout, and a dead-letter path. An inner loop would multiply
attempts against third-party URLs. A per-webhook circuit breaker would be the right tool, but
`partitioned` needs an operator-fixed key set and cannot key on arbitrary webhook IDs.

## Before this deploys

1. **Create the `dead_letter` Pub/Sub topic.** Without it every pool drops exhausted messages
   and only increments `jobs_pool_messages_dropped`.
2. **Create the `scheduler` Postgres role** with the union of the six retired job users' grants,
   and add `DATABASE_SCHEDULER_PASSWORD` to the `api-service-config` secret.
3. **Build and push the `dinner-done-better-scheduler` image.** The six retired job images become
   unused.
4. **Run migrations** so the outbox table exists before the API server writes to it.

## Alerts

Every one of these packages is designed to fail quietly — there is no caller to hand an error
to — so the counters are the only signal.

| Instrument | Why it matters |
|---|---|
| `jobs_pool_messages_dropped` | A message nobody handled and nowhere it went. Should always be zero. |
| `jobs_pool_messages_dead_lettered` | A message nobody handled. Needs triage. |
| `jobs_pool_queue_wait_ms` p99 | Rising means `Concurrency` is too low, well before throughput visibly drops. |
| `jobs_scheduler_leases_expired` | A job outran its lease; a second replica may have started it. |
| `jobs_scheduler_runs` vs `_skipped` | Together they are the fleet's tick count. Runs alone are what happened. |
| `outbox_backlog_age_seconds` | The only instrument that separates "publishing steadily" from "publishing steadily while falling further behind". |
| `outbox_messages_quarantined` | A permanently failing message, which is a dropped event. |
| `idempotency_claims_lost` | The remaining path to a duplicate charge. Always means `InFlightTTL` is too short. |
