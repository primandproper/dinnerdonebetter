# Jobs

Jobs are any chunk of code that is expected to run at some regular interval. This is as opposed to acting on the back of a message put onto a queue.

There are two shapes here, and a job's schedule decides which one it belongs in.

## `scheduler/` — interval-shaped work

One long-lived process running `jobs.Scheduler` (from `platform-go/v8/jobs`). Every registered job fires on an interval, and each execution is held under a `distributedlock` lease, so every replica ticks and only the one that wins the lock actually runs the job. A contended lock is the mechanism working, not an error.

Jobs are registered in `internal/build/jobs/scheduler/jobs.go` and scheduled by `config.ScheduledJobsConfig`. Adding one means writing the entrypoint, adding a `ScheduledJobConfig` field, and adding a row to `registrations`.

Two things to get right when adding a job:

- **`LeaseTTL` must exceed the job's worst case.** The lease is not renewed while the job runs. Past it, `Release` reports `ErrLockNotHeld`, `jobs_scheduler_leases_expired` increments, and a second replica may already have started the same job.
- **Set a `Timeout`, and set it below `LeaseTTL`.** A job that hangs should be killed before its lease lapses, not after.

Failed runs are not retried: the next tick is the retry. A job that cannot wait one interval wants a shorter interval, not an inner retry loop.

Watch `jobs_scheduler_runs` against `jobs_scheduler_skipped` (together they are the fleet's tick count), and alert on `jobs_scheduler_leases_expired`.

## `db_cleaner/`, `email_deliverability_test/` — calendar-shaped work

Kubernetes CronJobs, one short-lived pod per run. `jobs.Scheduler` takes intervals, not cron expressions, and an interval that starts whenever the pod last restarted is not the same thing as "midnight UTC" or "the 1st of the month". Work that has to land at a wall-clock time stays here.

## Queue consumers are not jobs

Work driven by a message rather than a clock lives in `cmd/functions/async_message_handler`, which runs a `jobs.Pool` per topic.
