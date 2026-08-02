# Data Privacy

How a GDPR/CCPA subject access request becomes a report, how that report is protected while it
exists, and how it stops existing.

## The disclosure lifecycle

A **user data disclosure** is the record of one request. It lives in `user_data_disclosures` and
moves through `pending → processing → completed → expired`, or to `failed` if the work does not
finish. The **artifact** is the object in storage that a completed disclosure points at, named by
its `report_id`.

| Step | Who | What happens |
|------|-----|--------------|
| Request | API server (`AggregateUserDataReport`) | Creates the disclosure with `expires_at = now + 7d`, publishes an aggregation request, returns the report ID. |
| Aggregation | Async message handler (`UserDataAggregationEventHandler`) | Gathers the user's data across every domain, encrypts it, writes the artifact, marks the disclosure `completed`. |
| Delivery | API server (`FetchUserDataReport`) | Reads the artifact, decrypts it, returns the collection over the authenticated call. |
| Expiry | Scheduler (`disclosure_artifact_reaper`) | Destroys the artifact, then marks the disclosure `expired`. |

## The artifact

One artifact is everything the system knows about one person, in one object. It is handled
accordingly.

**Encrypted at rest.** `internal/domain/dataprivacy/reportartifacts` owns the object: its path,
its cipher, and its destruction. Nothing else knows where an artifact lives or what format it is
in. The plaintext exists only in memory, on the way out of the aggregator and on the way back in
for the subject.

**Read-and-decrypt is the only delivery path.** There is no signed URL and no `Download`, because
the bytes behind such a URL would be base64 ciphertext the recipient has no key for. The server
reads the object, decrypts it, and returns the contents over the authenticated gRPC call. This is
the shape platform-go's `dataprivacy` package settled on independently — its `Download` refuses an
encrypted artifact with `ErrArtifactEncrypted`, and `Open` is the call that always works — so the
eventual migration to that package (#1254) should not change the delivery model.

**The bucket is private and unversioned.** Both properties are load-bearing rather than incidental:

- The bucket previously carried the same `allUsers`/`objectViewer` policy as the media bucket,
  which made every report world-readable to anyone who learned a report ID. It now grants
  `objectAdmin` to the workload identity service account and nothing else.
- Versioning is off. With versioning on, deleting an artifact only writes a tombstone and the
  object survives as a noncurrent version, so the expiry the reaper exists to enforce would not
  actually happen. A lifecycle rule purges the noncurrent versions left over from when it was on.

A 14-day lifecycle rule deletes anything the reaper never got to. It is a backstop, not the
mechanism — the seven-day expiry is enforced by the reaper.

## The reaper

`disclosure_artifact_reaper` runs hourly on the scheduler. Each pass takes a batch of disclosures
that are past `expires_at` and not yet `expired`, and for each one:

1. Deletes the artifact (both the `.json.enc` object and any pre-encryption `.json` object left
   from before artifacts were encrypted).
2. Marks the disclosure `expired`.

**Order matters.** The row must not claim the artifact is gone before it is: the query that finds
work excludes rows already marked `expired`, so a row flipped early leaves an object nothing will
ever come back for. A disclosure whose artifact cannot be deleted is left entirely alone and
retried on the next run.

The job takes more than one batch per run, stopping when a batch comes back short or when a batch
reaps nothing. That is what drains a backlog: on the first deploy carrying this job, every
artifact ever written is already past its expiry, and `RunOnStart` means the backfill happens
immediately rather than an hour later.

## Configuration

The same `internal/services/dataprivacy/config.Config` configures three processes, because three
of them touch artifacts — the API server reads, the async message handler writes, the scheduler
destroys. They must agree on the bucket and the key; a mismatch surfaces as an unreadable report
or a reaper that deletes nothing and reports success, neither of which fails at startup.

The scheduler carries the encryption key it does not strictly need, for that reason. The extra
exposure is nominal — it already holds database credentials for the data the artifacts are made of.

| Process | Env var for the key |
|---------|---------------------|
| API server | `DINNER_DONE_BETTER_SERVICE_DATA_PRIVACY_ARTIFACT_ENCRYPTION_KEY` |
| Async message handler | `DINNER_DONE_BETTER_DATA_PRIVACY_ARTIFACT_ENCRYPTION_KEY` |
| Scheduler | `DINNER_DONE_BETTER_DATA_PRIVACY_ARTIFACT_ENCRYPTION_KEY` |

The API server's spelling differs because it reaches this config through `Services`. All three
take the same value, from the `DISCLOSURE_ARTIFACT_ENCRYPTION_KEY` entry of the
`api-service-config` secret, generated by Terraform.

**Rotating the key** makes every artifact written under the old key permanently unreadable. That
is survivable — artifacts expire in seven days and a subject can request a fresh one — but it is
not a no-op, and the reaper still destroys the unreadable objects on schedule since deletion does
not involve the cipher.

## Deleting a user

`DestroyAllUserData` deletes the user and everything referencing them. Disclosure rows cascade
with the user, which means an outstanding disclosure's artifact is no longer reachable through
the reaper's query — the row that pointed at it is gone. Only the bucket lifecycle rule collects
those, so an orphaned artifact survives up to 14 days after the account it describes was erased.

This is a known gap rather than a designed behavior. It is bounded rather than urgent now that
artifacts are ciphertext in a private bucket, and #1254 addresses it properly as part of adopting
platform-go's `dataprivacy`, which models collection and erasure together.

## Related

- `docs/identity.md` — users, accounts, memberships
- Issue #1254 — replacing this machinery with platform-go's `dataprivacy` package
