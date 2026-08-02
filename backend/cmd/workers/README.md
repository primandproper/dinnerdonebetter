# Jobs

Jobs are any chunk of code that is expected to run on a clock. This is as opposed to acting on the back of a message put onto a queue.

Almost all of them belong in `scheduler/`. Two do not, and the reason is isolation rather than scheduling.

## `scheduler/` — clock-driven work

One long-lived process running `jobs.Scheduler` (from `platform-go/v9/jobs`). Every registered job fires on a schedule, and each execution is held under a `distributedlock` lease, so every replica ticks and only the one that wins the lock actually runs the job. A contended lock is the mechanism working, not an error.

Jobs are registered in `internal/build/jobs/scheduler/jobs.go` and scheduled by `config.ScheduledJobsConfig`. Adding one means writing the entrypoint, adding a `ScheduledJobConfig` field, and adding a row to `registrations`.

Two things to get right when adding a job:

- **`LeaseTTL` must exceed the job's worst case.** The lease is not renewed while the job runs. Past it, `Release` reports `ErrLockNotHeld`, `jobs_scheduler_leases_expired` increments, and a second replica may already have started the same job.
- **Set a `Timeout`, and set it below `LeaseTTL`.** A job that hangs should be killed before its lease lapses, not after.

Failed runs are not retried: the next fire is the retry. A job that cannot wait one cycle wants a tighter schedule, not an inner retry loop.

Watch `jobs_scheduler_runs` against `jobs_scheduler_skipped` (together they are the fleet's tick count), and alert on `jobs_scheduler_leases_expired`.

### Intervals and calendars

`config.ScheduledJobConfig` carries both `Interval` and `Schedule`, and exactly one of them is set. A job carrying both is rejected at config validation rather than resolved by precedence.

Two jobs are on a calendar:

- **`mobile_notification_scheduler`** — `CRON_TZ=America/Chicago 0 8-21 * * *`. Push notifications, so the hours are the point. Hourly rather than once in the morning because each task notifies exactly once and is dropped if its event starts first, so the gap between fires bounds how much short notice the app can give. Per-user timezones are the real answer and a larger conversation.
- **`search_data_index_scheduler`** — `*/10 6-11 * * *`. A bulk sweep, so it belongs in the small hours rather than competing with daytime traffic. A window and not a single fire because `IndexScheduler.IndexTypes` sweeps one *randomly chosen* index type per run: with nine types registered, one fire a night would sweep a given type every nine days.

The rest stay on intervals. The meal plan jobs are time-sensitive — a plan should finalize soon after its voting deadline, not at a fixed hour — and `queue_test` is a liveness probe, which wants a frequency by definition.

### Timezones

`jobs.SchedulerConfig.Timezone` is set to `"UTC"` in `internal/config/environment.go`, named rather than left empty. The default is UTC either way, but a cron expression's zone is the one thing about it you cannot read off the expression, and a zone that arrives by omission is a zone nobody chose. Both calendar jobs are read in it; the notification job overrides it in its own spec, because its hours are a fact about people rather than about load, and a fixed instant is the better default for everything else.

Anything but UTC needs the zoneinfo database, and `NewScheduler` resolves the zone at construction — so a missing database is a crash loop at boot rather than a missed run. `cmd/workers/scheduler` imports `_ "time/tzdata"` and embeds it, which makes the binary independent of whatever the base image ships.

### Overrun means something different on a calendar

It is outlasting the headroom to the next fire, not outlasting a fixed interval, and a calendar's headroom varies — a job at `0 9 * * 1-5` has three days of it on Friday night and one on Monday. `TestDefaultScheduledJobsConfig` walks a year of each cron job's fire times and asserts its `Timeout` is below the tightest gap the expression ever produces.

## `db_cleaner/`, `email_deliverability_test/` — separate pods on purpose

Kubernetes CronJobs, one short-lived pod per run. `jobs.Scheduler` takes cron expressions, so either could move in; they stay out because they are the two jobs whose isolation is worth a pod. `db_cleaner` keeps its own Postgres user and its own blast radius for a bulk delete, and a deliverability probe that shares a process with the thing it is probing is not much of a probe. That is a judgement about failure domains, not about what the scheduler can express.

## Queue consumers are not jobs

Work driven by a message rather than a clock lives in `cmd/functions/async_message_handler`, which runs a `jobs.Pool` per topic.
