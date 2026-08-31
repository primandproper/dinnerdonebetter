# `ddb`

Every deployed workload is a subcommand of this one binary:

| Command                        | Shape                | Config struct                   |
|--------------------------------|----------------------|---------------------------------|
| `ddb serve`                    | long-lived           | `APIServiceConfig`              |
| `ddb serve mcp`                | long-lived           | `MCPServiceConfig`              |
| `ddb worker async-messages`    | long-lived           | `AsyncMessageHandlerConfig`     |
| `ddb worker scheduler`         | long-lived           | `SchedulerConfig`               |
| `ddb job db-cleaner`           | one-shot             | `DBCleanerConfig`               |
| `ddb job email-deliverability` | one-shot             | `EmailDeliverabilityTestConfig` |
| `ddb migrate`                  | one-shot, flags only | none                            |
| `ddb version`                  | one-shot             | none                            |

One binary means one image, built and tagged once per release, so a deployment cannot drift to a
commit its neighbors are not running. It does not mean one config: each subcommand loads only its
own struct, from the file named by `CONFIGURATION_FILEPATH`, exactly as it did when these were
separate binaries. Which workload a pod runs is its manifest's `args`.

Adding a workload means adding a subcommand here and a manifest that names it — not a new
`cmd/` directory, Dockerfile, skaffold artifact, and image.

Development-time tools that are never deployed (`codegen`, `aiagent`, `push_tester`, `bootstrap`,
and the rest) stay under `cmd/tools/`, out of the production image.

## Jobs

Jobs are any chunk of code that is expected to run on a clock. This is as opposed to acting on the back of a message put onto a queue.

Almost all of them belong in `ddb worker scheduler`. Two do not, and the reason is isolation rather than scheduling.

### `ddb worker scheduler` — clock-driven work

The process also runs two loops that are not scheduled jobs: the **outbox relay**, which moves
events written inside a caller's transaction onto the broker, and the **saga worker**, which
advances every durable saga instance this build knows how to run. Neither takes a schedule — they
poll — and both live here because this is the process that does background work. Jobs start
sagas; the worker is what steps them through, which is why `meal_plan_finalization_starter`'s
interval is the delay before a meal plan enters the finalization pipeline rather than the delay
before it comes out the other end.

It also holds two `workqueue.Queue` values over one `work_queue_items` table, told apart by
`Config.Name`: the **operations** queue the operations worker dispatches through, and the
**meal plan task notification** queue its own job fills and drains. One table serves any number
of logical queues — the name is the leading column of its primary key — so a third is a third
`Config`, not a third migration.

Because that notification job sends its own pushes rather than publishing them, this process
needs the APNs credentials the async message handler has: the env vars from `api-service-config`
and the `.p8` key mounted at `/mnt/apns`. Both kustomize patches are applied to it.

One long-lived process running `jobs.Scheduler` (from `platform-go/v13/jobs`). Every registered job fires on a schedule, and each execution is held under a `distributedlock` lease, so every replica ticks and only the one that wins the lock actually runs the job. A contended lock is the mechanism working, not an error.

Jobs are registered in `internal/build/jobs/scheduler/jobs.go` and scheduled by `config.ScheduledJobsConfig`. Adding one means writing the entrypoint, adding a `ScheduledJobConfig` field, and adding a row to `registrations`.

Two things to get right when adding a job:

- **`LeaseTTL` must exceed the job's worst case.** The lease is not renewed while the job runs. Past it, `Release` reports `ErrLockNotHeld`, `jobs_scheduler_leases_expired` increments, and a second replica may already have started the same job.
- **Set a `Timeout`, and set it below `LeaseTTL`.** A job that hangs should be killed before its lease lapses, not after.

Failed runs are not retried: the next fire is the retry. A job that cannot wait one cycle wants a tighter schedule, not an inner retry loop.

Watch `jobs_scheduler_runs` against `jobs_scheduler_skipped` (together they are the fleet's tick count), and alert on `jobs_scheduler_leases_expired`.

#### Intervals and calendars

`config.ScheduledJobConfig` carries both `Interval` and `Schedule`, and exactly one of them is set. A job carrying both is rejected at config validation rather than resolved by precedence.

Two jobs are on a calendar:

- **`meal_plan_task_notifications`** — `CRON_TZ=America/Chicago 0 8-21 * * *`. Push notifications, so the hours are the point. Hourly rather than once in the morning because each task notifies exactly once and is dropped if its event starts first, so the gap between fires bounds how much short notice the app can give. Per-user timezones are the real answer and a larger conversation.

  One pass enqueues every prep task still owed a reminder into a `workqueue.Queue`, then claims and sends under the lease. It is the only job here that both fills and drains a queue in one tick, which is deliberate: the send has to end with the same `notification_sent_at` stamp the discovery query reads, or the next pass rediscovers work that is already done. See `internal/services/mealplanning/workers/meal_plan_task_notifications`.
- **`search_data_index_scheduler`** — `*/10 6-11 * * *`. A bulk sweep, so it belongs in the small hours rather than competing with daytime traffic. A window and not a single fire because `IndexScheduler.IndexTypes` sweeps one *randomly chosen* index type per run: with nine types registered, one fire a night would sweep a given type every nine days.

The rest stay on intervals. `meal_plan_finalization_starter` is time-sensitive — a plan should finalize soon after its voting deadline, not at a fixed hour — and `queue_test` is a liveness probe, which wants a frequency by definition.

#### Timezones

`jobs.SchedulerConfig.Timezone` is set to `"UTC"` in `internal/config/environment.go`, named rather than left empty. The default is UTC either way, but a cron expression's zone is the one thing about it you cannot read off the expression, and a zone that arrives by omission is a zone nobody chose. Both calendar jobs are read in it; the notification job overrides it in its own spec, because its hours are a fact about people rather than about load, and a fixed instant is the better default for everything else.

Anything but UTC needs the zoneinfo database, and `NewScheduler` resolves the zone at construction — so a missing database is a crash loop at boot rather than a missed run. `cmd/ddb` imports `_ "time/tzdata"` and embeds it, which makes the binary independent of whatever the base image ships.

#### Overrun means something different on a calendar

It is outlasting the headroom to the next fire, not outlasting a fixed interval, and a calendar's headroom varies — a job at `0 9 * * 1-5` has three days of it on Friday night and one on Monday. `TestDefaultScheduledJobsConfig` walks a year of each cron job's fire times and asserts its `Timeout` is below the tightest gap the expression ever produces.

### `ddb job db-cleaner`, `ddb job email-deliverability` — separate pods on purpose

Kubernetes CronJobs, one short-lived pod per run. `jobs.Scheduler` takes cron expressions, so either could move in; they stay out because they are the two jobs whose isolation is worth a pod. `db_cleaner` keeps its own Postgres user and its own blast radius for a bulk delete, and a deliverability probe that shares a process with the thing it is probing is not much of a probe. That is a judgement about failure domains, not about what the scheduler can express.

Sharing an image with the API server does not weaken that: the isolation these two want is a
separate process, separate credentials, and a separate failure domain, none of which is a
property of the image they were built from.

`db-cleaner` removes a small subset of archived records. It is very nearly the only place that
runs a proper `DELETE` — right now only expired OAuth2 tokens, but it exists as a canvas for
whatever else turns out to be worth deleting regularly.

### Queue consumers are not jobs

Work driven by a message rather than a clock lives in `ddb worker async-messages`, which runs a `jobs.Pool` per topic.
