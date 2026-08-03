# Audit log

The audit log is the durable, queryable, tamper-evident record of who did what to
which resource. It is not logging: a log is a best-effort account written beside
the work, and this is a record written inside it, in a table nothing can edit,
chained so that removing or altering an entry is detectable after the fact.

It is `github.com/primandproper/platform-go/v9/audit`. We own the mapping onto our
domain — what a scope is, what an actor is, what is never recorded — and nothing
else. The schema belongs to the platform on purpose: the unique index that makes
a forked chain unrepresentable and the chain row that serializes concurrent
writers are the guarantee, not incidental storage details.

## Tables

Two, under our namespace, created by generated migrations 24 and 25 rather than
by DDL copied into `migration_files`:

| Table | What it holds |
|---|---|
| `ddb_audit_log_entries` | one row per recorded event, carrying its position in its scope's chain, its predecessor's hash, and its own |
| `ddb_audit_log_chains` | one row per scope: that scope's chain head and how far retention has pruned it |

The prefix is `audit.TablePrefix` in `internal/domain/audit`. Every component that
touches these tables — the Recorder, the Reader, the Sweeper — takes that same
constant, so a component pointed at the wrong tables is not expressible.

## The mapping

| Ours | Platform's | Notes |
|---|---|---|
| `BelongsToAccount` | `Entry.Scope` | the account ID, or `""` for events belonging to no account |
| `BelongsToUser` | `Entry.Actor.ID` | `"system"` when there is no requester behind the event |
| `RelevantID` | `Entry.ResourceID` | |
| `CreatedAt` | `Entry.RecordedAt` | |
| `Changes` (strings) | `Entry.Changes` (typed) | the platform stores typed values; the repository renders them as strings for the API, which has always spoken strings |

`AuditLogEntry` — the type the gRPC surface returns — is a projection, not the
stored row. The chain fields are deliberately left off it: a position and a pair
of hashes mean nothing without the neighbors that give them meaning, and
publishing them would invite a reader to believe they had checked something.
`VerifyChain` is how the chain is asked a question.

### Scope is the account, and the empty scope is real

Scope is the hash chain's partition, so entries in different scopes are unrelated
positions and two accounts writing concurrently do not serialize against each
other. Events with no account — a signup, a login, a password reset, a service
setting, a product — land in the empty scope, which is a chain like any other and
holds all of them. That is a real contention point under load and a known
tradeoff: it was chosen over falling back to the user ID so that "scope" means one
thing, and so that a future audit eraser deleting a subject's scopes cannot
delete events that merely happened to have no account attached.

`Query.Scope` is a pointer in the platform's read API for the same reason: `""` is
a real scope, so a plain string could not distinguish "only platform events" from
"every account's events", and in a multi-tenant read path that distinction is a
disclosure rather than a wrong answer. Our read methods reject an empty account ID
before they get there.

## Writing

`Record` takes the caller's `database.SQLQueryExecutor`, which is the whole
design. An audit entry that can commit while the change it describes rolls back —
or the reverse — is not a record of what happened, and no retry fixes it after the
fact.

```go
err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
	if err := r.generatedQuerier.UpdateThing(ctx, tx, arg); err != nil {
		return err
	}

	changes, err := audit.Diff(before, after)
	if err != nil {
		return err
	}

	return r.auditLogEntryRepo.Record(ctx, tx, &audit.Entry{
		Scope:        accountID,
		ResourceType: resourceTypeThings,
		ResourceID:   thing.ID,
		EventType:    audit.EventUpdated,
		Actor:        audit.UserActor(userID),
		Changes:      changes,
	})
})
```

**Always pass a transaction, never `r.writeDB`.** This is not a style preference.
`Record` locks the scope's chain-head row and holds it for the remainder of the
caller's transaction, which is what makes the second concurrent writer wait and
then read the head the first one committed. Against the connection pool there is
no remainder — the lock lapses at the end of the implicit single-statement
transaction, before the INSERT it was taken for — so two writers compute the same
position and the unique index rejects one of them, taking down a business
transaction whose only mistake was arriving second.

`Record` is variadic, and it is worth using: a transaction touching three
resources should pay one chain-head lookup and one INSERT rather than three of
each, while holding a lock every other writer to that account is queueing behind.

### `Diff` over hand-built change maps

`audit.Diff(before, after)` builds the change map by reflection, using the `json`
tag for field names so the audit log and the API speak the same vocabulary. Prefer
it: the field somebody forgets to add to a hand-built map when they add it to the
struct is exactly the field an investigation will want.

Two things to know. Diff the row that will exist after the write, not the caller's
partially-populated input — otherwise every field the UPDATE does not touch reads
as having been cleared. And a field that must never be audited anywhere carries
`audit:"-"` (see `identity.Account.Members`); a field that must not be recorded in
a given deployment is a `Redaction` instead.

### Redaction

`audit.Redactions` in `internal/domain/audit` declares what never reaches the
table, keyed by resource type, with the empty key applying to everything. A
password hash or a bearer token that lands here is in the one table designed to be
immutable and retained for years, and filtering it at query time does not un-write
it.

It is a reviewed Go declaration rather than a config knob, deliberately: "which
fields must never be recorded" should show up in a diff.

## Reading and verifying

The five read methods are unchanged in signature and semantics. `VerifyChain`
walks one account's chain over a time range and reports the first break, or that
there was none.

What a clean verification proves, stated precisely because it is easy to overstate:
every entry in the range hashes to what it claims, and each links to the one
before it, so nobody edited, removed, or reordered an entry without also rewriting
every entry after it. It does not prove the table was not replaced wholesale by a
consistent forgery — nothing self-contained can. The answer to that is to publish
head hashes somewhere this database's owner does not control; `Record` writes each
entry's `Hash` back into the value you passed, which is what you would publish. We
do not do that today.

Only the *first* break is reported. After a break, every subsequent link is
evaluated against a predecessor already known to be wrong, so the list of later
breaks says how long the chain is, not how much of it was tampered with.

**Alert on `audit_chain_breaks`.** Everything else the package emits describes
throughput. That one means the log has stopped being evidence, and a non-zero
value is an incident on its own.

## Append-only enforcement

Migration 25 installs triggers that make the database refuse an `UPDATE` on
`ddb_audit_log_entries` outright. Editing a recorded entry is not something the
chain reveals afterwards; it is something the database will not do.

`DELETE` is deliberately left permitted. Retention has to remove aged entries and
no trigger can tell that sweep apart from an attacker, so blocking deletion would
mean shipping a log that grows forever. Deletion is covered by the chain instead:
positions within a scope are contiguous, so a removed row leaves a hole `Verify`
reports, and the retention sweep records where it pruned to so its own holes are
distinguishable from everyone else's.

If the deployment can revoke `UPDATE` and `DELETE` from the application role, do
that as well — it is strictly stronger than either, because it also stops the
deletions the triggers cannot.

## Retention

The `audit_log_sweeper` job runs nightly under the shared `jobs.Scheduler`, with a
**two-year** window (`ScheduledJobsConfig.AuditRetention`).

Two years rather than the platform's seven-year default, which is pitched at
deployments carrying regulated records. Two years covers a dispute, an incident
review, or an annual compliance question with room on either side, and bounds the
table's growth. Shortening it is a decision worth making on purpose, which is why
it is written in `defaultScheduledJobsConfig` with this reasoning attached rather
than inherited silently. Config validation refuses a window under an hour: a
misplaced unit there would mean "keep nothing".

The sweep removes only a prefix of a scope's chain, never a row from the middle,
so survivors stay contiguous and verifiable against each other — and it records
the hash of the last entry it removed as that scope's prune watermark, in the same
transaction, so the oldest survivor still links to something and `Verify` can tell
retention's gap from a deletion.

## The cutover

The hand-rolled `audit_log_entries` table was **dropped**, not migrated. Its rows
were not carried across.

That was a deliberate choice over the alternative of backfilling the old rows
through `Record` in `created_at` order. A backfill would have preserved the
history in the API, but the chain it produced would attest only that the backfill
read what the table said at the moment it ran — which is exactly the claim in
question. Shipping a chain that looks like evidence and is not seemed worse than
starting clean and saying so.

**So: the audit log attests to everything from migration 23 forward, and to
nothing before it.** Pre-cutover history is gone.

### What else the old table was quietly doing

`audit_log_entries.belongs_to_account` and `.belongs_to_user` carried foreign keys
to `accounts` and `users` with `ON DELETE CASCADE`. The platform's tables carry no
such keys — an audit log that cannot outlive the rows it describes is not much of
an audit log — and two consequences follow:

- **Deleting a user no longer cascades their audit entries.** This is the
  interaction with erasure. platform-go ships `dataprivacy/auditerasure`, whose
  eraser deletes whole audit scopes belonging to a subject and reports the rest as
  retained with a legal basis, because deleting from the middle of a chain makes
  `Verify` report tampering for the rest of that scope's history. We do not use
  the platform's eraser registry yet; wiring it up belongs to the data-privacy
  work, not here.
- **Those keys were acting as accidental existence checks.** `ArchiveAccount` used
  to fail for a nonexistent account only because the audit insert violated the
  foreign key. It now checks rows affected explicitly, as every other archive path
  here does.

## Where things live

| | |
|---|---|
| Domain types, event vocabulary, redactions, actor helpers | `backend/internal/domain/audit` |
| Adapter onto the platform's Reader and Recorder | `backend/internal/repositories/postgres/auditlogentries` |
| Table DDL and append-only triggers | `backend/internal/repositories/postgres/migrations/migrate.go` |
| Sweeper wiring and schedule | `backend/internal/build/jobs/scheduler`, `backend/internal/config` |
| Chain, redaction, and append-only tests | `backend/internal/repositories/postgres/auditlogentries/audit_log_entries_test.go` |

`DataChangeMessage` lives in the same domain package and is a different concern —
it is the queue event, not the audit record — and stayed put.
