# Migrations

Migrations are plain SQL files in
[`internal/repositories/postgres/migrations/migration_files/`](../internal/repositories/postgres/migrations/migration_files/),
embedded into the binary by the compiler. They are applied by the platform's
`database/migrate`, which wraps goose: no global goose state, and on Postgres a session advisory
lock so concurrently booting replicas serialize instead of racing.

Adding one means dropping a numbered `.sql` file into that directory — the leading number orders
the sequence and there is no list to keep in sync. Files are read and validated in `NewMigrator`,
so a malformed migration fails construction rather than the first `Migrate`.

## When they run

At **app startup**, as part of building the database client: `database/config.NewDatabase` pulls
the `Migrator` out of the DI container and runs it when `Database.RunMigrations` is set. Every
generated config sets it, production included, so a deploy migrates on boot; the advisory lock
means one replica applies while the rest wait. Workers and jobs explicitly force it to `false`
(`cmd/ddb/worker.go`, `cmd/ddb/job.go`), so only the API server migrates.

`ddb migrate` runs the same migrator standalone. It is a helper for doing it out of band — not
the path a deploy takes.

## Numbering

The number sequence is shared between files on disk and platform-owned tables spliced in via
`migrate.WithGeneratedMigration` (outbox, saga, webhooks, audit, and the rest — see the version
constants in `migrations/migrate.go`). A platform package renders its own DDL rather than having
it copied here, but the version number belongs to us, because numbering belongs to whoever owns
the sequence.

Today the platform tables hold 1–25 and this application's one hand-written file,
`00026_dinnerdonebetter.sql`, comes last, because its foreign keys name platform tables and a key
cannot reference a table that does not exist yet. Identity is 1 for the same reason: nearly every
other table points at it.

So a newly adopted platform table goes before the hand-written file, and the file moves up a
number. Ordinarily that would be forbidden — goose keys applied migrations by version, so
renumbering makes an applied migration look unapplied and re-runs it — but nothing is deployed, so
the only databases that have applied anything are local ones, and the cost is the one the next
section describes: drop and recreate your local database. Once something is deployed, the rule
becomes **never renumber an applied migration**, and new tables take the next free number at the
end. A version already claimed fails `New` rather than the first `Migrate`.

## Editing an existing migration

Ordinarily you never do this; you add a new migration instead. But nothing is deployed
(see the root `CLAUDE.md`), so there is no schema anywhere that has to be migrated forward from
its current state — and editing a base migration in place is usually the better end state than
accumulating a corrective migration for a schema no database has ever held.

The catch is local, not operational: **goose keys applied migrations by version and does not
checksum their bodies.** An edit to a file some database has already applied is silently skipped
on that database — no error, the schema just quietly stays behind the file. Anyone with an
existing local Postgres volume has to drop and recreate it, and the failure mode if they don't is
confusing rather than loud.

So when you edit an applied migration, say so in the PR description, and treat "drop your local
database" as part of the change.
