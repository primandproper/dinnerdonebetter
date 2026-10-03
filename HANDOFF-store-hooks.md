# Handoff: platform store hooks

Delete this file before merging.

## What this branch is

DDB's repository packages used to wrap platform-go stores to write an audit
entry and a data change event beside every write. platform-go now gives each
store a `Hooks` interface (one `After…` method per write, called on the
caller's transaction, updates handed the row before and after), and this
branch replaces every wrapper with a hooks implementation.

- **DDB branch:** `store-hooks-adoption`, renamed from `store-hooks` (off `cleanup` at `818eab2b7`)
  - `624b78bd9` — the port: waitlists, settings, comments, issuereports,
    uploadedmedia, notificationsstore, payments, webhooksstore, auth
    (password reset), oauth2clientsstore (now uses the Service's hooks),
    identitystore (adapted to the new signatures, no longer embeds NoopHooks).
  - `3c3932690` — **TEMPORARY** `go.mod` replace pointing at `../../platform-go`.
    Drop it and bump platform-go once the platform PR is released. On another
    machine it only builds if platform-go is checked out on `store-hooks`
    as a sibling of this repo. It also blocks `make lint`: the linter runs in a
    container where `../../platform-go` does not exist.
  - `3cdad0f47` — decisions 2 and 3 applied (see below).
- **platform-go branch:** `store-hooks`, renamed from `waitlists-hooks`,
  PR primandproper/platform-go#1080. It replaces #1079, which the rename closed
  (GitHub deleted the old head ref rather than retargeting).
  - `107df8f6` waitlists hooks (the worked example)
  - `0cd5e28a` hooks for settings, comments, issuereports, mediaregistry,
    notifications, billing, webhooks, passwordreset; before/after on updates
  - `45046fd5` identity + oauth2clients update hooks take before/after
    (**breaking** — see decisions)
  - `eb192c0a` webhook `Headers` tagged `audit:"-"`; `NoopHooks` docs name both
    choices

## Conventions the port follows

- DDB hooks types implement the platform interface **outright** (no
  `NoopHooks` embed) so a new platform write breaks the build until someone
  decides what it records. Deliberately unrecorded writes are explicit no-ops
  with a reason.
- Provider constructors kept their signatures, so DI is unchanged
  (exceptions: webhooksstore `ProvideStore`, notificationsstore lost the unused
  `ProvideInbox`/`ProvideRegistry`, oauth2clientsstore `ProvideStore` →
  `ProvideHooks`).
- Recorded entries/events are the same as before, except updates now carry
  `entry.Changes` from `platformaudit.Diff(before, after)`.
- Platform: the before row is read only when hooks are installed; erasures and
  bulk deletes get a count, never rows; writes that take no caller Tx get no hook.

## Impact

DDB repository packages, non-test Go: 4,229 → 3,494 lines (−17%). Biggest:
payments 548→366, waitlists 447→280, notificationsstore 410→291. auth and
identitystore grew slightly. notificationsstore, oauth2clientsstore and
webhooksstore gained integration tests (they had none).

## Test status (as of handoff)

- platform-go: `make lint` clean; every touched package passes on SQLite and on
  Postgres/MySQL containers (`RUN_CONTAINER_TESTS=true`).
- Pre-existing, not from this work: billing and identity SQLite tests flake
  with "database schema has changed (17)" at the same rate on `origin/main`;
  running the whole suite at once can exhaust MySQL containers (the affected
  packages pass alone).
- DDB: builds and vets; all 11 ported packages pass with
  `RUN_CONTAINER_TESTS=true go test ./internal/repositories/postgres/...`.
  `cmd/tools/codegen/converters` fails, identically on clean `cleanup`.
- DDB lint: clean when golangci-lint runs natively, apart from
  `gomoddirectives` flagging the temporary replace itself; `make lint` cannot
  typecheck until the replace is dropped.

## Decisions

1. **Accepted — breaking change in platform-go.** `identity.Hooks.AfterUpdateAccount`,
   `identity.Hooks.AfterUpdateUserAccountStatus` and
   `oauth2clients.Hooks.AfterUpdateClient` changed signature; both interfaces
   shipped in v14.2.0. It ships as a v14 minor: v14.2.0 itself made ~80
   breaking API changes against v14.1.0 (apidiff), including a hook signature
   change of exactly this kind, so platform-go does not hold v14 to Go semver.
2. **Done — personal data in audit diffs.** Webhook endpoint `Headers` is
   `audit:"-"` in platform; comment `body` and waitlist signup `notes` are hashed
   in `backend/internal/domain/audit/redaction.go`.
3. **Done — smaller calls.** Kept issuereports' before/after and settings'
   `*Definition`; `AfterConsume` records no diff; the account event's `changed`
   list drops `lastUpdatedAt` (the audit diff keeps it); `AfterUpdateProfile`
   left without a before row; `NoopHooks` docs updated. Webhooksstore's two
   behavior changes stand as described in the PR.

## Next steps

1. Review #1080; get it merged and released.
2. Drop `3c3932690`, bump platform-go in `backend/go.mod`, delete this file.

Earlier context from the same session, not yet acted on: the analytics gRPC
proxy (`internal/services/analytics/grpc`, iOS `BackendEventReporter.swift`)
is the ad-blocker-style proxy the "no speculative protective services" rule
says to delete; and platform's `webhooks.Emitter` may make DDB's local
`internal/repositories/postgres/events.Emitter` redundant.
