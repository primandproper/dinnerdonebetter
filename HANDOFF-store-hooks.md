# Handoff: platform-go v15 adoption

Delete this file before merging.

## What this branch is

`store-hooks-adoption` began as the port of DDB's repository wrappers onto platform's `Hooks`
interfaces (platform-go #1080). platform then shipped `recording` and a `RecordingHooks` per
store, and the branch became the first consumer of platform-go **v15**: every hand-written hook
this application had over a platform-owned store is deleted in favour of platform's — identity,
sign-in and password reset included, as of the latest commit — and the recording spine
underneath them is platform's too.

**None of the libraries are tagged.** That is deliberate: the point is to consume them here
first and find their bugs before a release. `backend/go.mod` therefore carries two local
replaces — `platform-go/v15 => ../../platform-go` and `primitives-go/v2 => ../../primitives-go`
— and builds only with those siblings checked out at or after platform-go `ea5b750a` and
primitives-go `e6509109`. `make lint` runs in a container that cannot see them; lint natively
(`golangci-lint run ./...`) and ignore the two `gomoddirectives` hits on the replaces. The
frontend and iOS still reference published `0.0.1` client libraries; moving them to local paths
is step 6 below and has not started.

## Commits, in order

- `3c7e49b32` — import path bump to `/v15`, both replaces.
- `f9960dd0e` — the recording spine. `internal/repositories/postgres/events.Emitter` is a thin
  adapter over platform's `webhooks.Emitter` and `recording.Recorder`, registered through
  platform's own `webhookscfg.NewEmitter` and `recordingcfg.Register` (`FileBySubject`,
  `sessions.PrincipalFromContext` as the `callers.PrincipalExtractor`). `indexevents` is a
  `[]searchsync.Rule` table; `datachanges.Message` implements `searchsync.Change`.
- `45430c8c0` — the webhook catalog is `webhooks.Merge` of this application's generated
  fragment with every adopted package's `EventCatalog()`; `internal/domain/audit/redaction.go`
  deleted.
- `2afb79184` — ten hook swaps: waitlists, settings, comments, issuereports,
  uploadedmedia→mediaregistry, notificationsstore, payments→billing, webhooksstore,
  oauth2clientsstore, passkeys. `localdev.Spine` for the hand-assembled processes.
- `1686c7d8b` — three upstream landings consumed: the audit privacy resolvers are platform's
  (`privacyadapters.AuditScopeResolvers`, #1135; `internal/domain/audit/privacy` deleted),
  `EmitIndex` is the writer's `EnqueueDerived` (#1132), the local `omittedByPlatform` catalog
  fragment is gone (#1133).
- (this commit) — the envelope (#1134) and the last three hook swaps. Everything on the data
  changes topic is a `webhooks.Envelope`; the async handler routes on its event type and decodes
  each payload as the type that event names. identity, sign-in and password reset record through
  `identity.NewRecordingHooks`, `signin.NewRecordingHooks` and `passwordreset.NewRecordingHooks`;
  `identitystore/hooks.go`, `authentication/signin_hooks.go`, `PasswordResetDirectory` and the
  `passwordResetHooks` are deleted, and with them every DDB identity event constant but the three
  mail requests below. The webhook catalog regenerated (34 identity entries gone).

## What stays local, and why

- **`identitystore.Hooks`** — a wrapper over `identity.RecordingHooks` that overrides four
  methods (register, register-with-invitation, update-profile, archive-user) to run the users
  index rules through `EnqueueDerived` after platform's recording. It exists because no payload
  platform's own hooks emit implements `searchsync.Change` (**platform-go #1136**), so platform's
  events cannot feed an index; the rules in `indexevents` already match platform's event names
  and only the wrapper goes when #1136 lands.
- **The three mail-request events** in `internal/domain/identity/auth_events.go`
  (`password_reset_token_created`, `username_reminder_requested`,
  `user_email_address_verification_email_requested`), emitted by `authentication.SignInMailers`.
  platform's own events for those writes deliberately carry no secret; the mail cannot be
  rendered without one, so platform hands the secret to a mailer after commit and the mailer
  puts it on one of these. All three are `Internal` in the catalog.
- **`events.Emitter` and the `recording` adapter** — `Emit`/`Record`/`RecordAndEmit` for the
  mealplanning repositories, whose events are this application's own.
- **`internal/domain/analytics`** now owns *how each reportable event's payload is read* (who to
  attribute it to, which properties, never a secret), because platform's payload types do not
  share a shape. `datachangemessagehandler.reportToAnalytics` resolves an account-only event
  (a subscription) to the account's owner.

## Behaviour that changed on purpose

- **The actor is the requester**, and platform's `RecordAs` names the sign-in's subject. An
  impersonation is recorded as the *subject's* `signin.user.authenticated` with the operator as
  the entry's `Impersonator` and the event's `actorID` (DDB filed it under the operator).
- **Erasures record nothing per store** (platform #1123 / #1115): dataprivacy records once.
- **Event vocabulary.** DDB's `user_signed_up`/`user_logged_in`/… are platform's
  `identity.user.registered`, `signin.user.authenticated`, `passwordreset.token.redeemed`, etc.
  The analytics allowlist and the handler's mail routes moved with them.
- **The whole identity, signin, passwordreset, passkeys and oauth2clients fragments are
  `Internal`** in the webhook catalog (`catalog/excluded.go`), by fragment rather than by name,
  so a new identity event nobody classified is excluded by default. This narrows what was
  deliverable before: `account_created`/`account_updated`/`account_archived` and the four
  invitation events (one of which carried the invitation token to subscribers) were subscribable.
- **The household push on an accepted invitation fires again.** DDB's hook never put the
  destination account on the event the handler read it from, so the notification had been
  skipped since the identity adoption; platform's `InvitationEvent.AccountID` restores it, and a
  registration through an invitation now also tells the household.
- **The analytics vendor no longer receives the verification link.** DDB forwarded the whole
  event context on signup, token included; `analytics.userEvent` reads the account and nothing
  else.
- **The queue-test probe** on the data changes topic is an envelope (`internalops.QueueTestProbe`,
  acknowledged by its ID); `datachanges.Message.TestID` is gone.
- A registration's three audit entries are **unattributed** rather than the registrant's
  (**platform-go #1137**, filed; DDB's adapter used `RecordAs` here and platform's hooks do not).
  `identity_users_test.go` asserts by resource until it lands.
- **A pending invitation's entry is on the global chain**, not the account's: platform gives it a
  subject only once it is answered (**platform-go #1138**, filed). Membership entries are filed
  under the member, so an account's own log shows nobody joining or leaving; the ticket asks
  about that too. `identity_accounts_test.go` asserts the invitation by resource and reads the
  ended membership as the member until it lands.
- **An anonymous password reset is unattributed** (nobody is signed in to ask for or spend a
  link), with the user as the entries' subject. `passwordreset_test.go` asserts by resource,
  reading the token's ID off the mail request.

## Found upstream on the way

Filed this round: **#1136** (no platform payload implements `searchsync.Change`; propose the
envelope answer it), **#1137** (`AfterRegister` should `RecordAs` the registrant), **#1138**
(a pending invitation's entry names no subject). Earlier: #1130, #1131 (both landed). Still
open from the original review: template-go #7.

The conformance harness reads links off the outbox: the verification link is now on platform's
`identity.user.registered` payload (`emailAddressVerificationToken`) or this application's
re-request mail, and the invitation token is on `identity.invitation.created`
(`conformance_test.go`).

Not a defect, but a consequence of building against primitives `main`: 65 `staticcheck` hits for
`filteringgrpc.FromProto`, deprecated in favour of `QueryFilterFromProto(in, archiveDecision)`.
That is the `include_archived` seam and a follow-up of its own.

## Test status (at the last commit)

- `go build ./...`, `go vet ./...`: clean.
- `RUN_CONTAINER_TESTS=true go test ./internal/repositories/postgres/... ./internal/build/...
  ./internal/localdev/...`: all pass.
- `RUN_CONTAINER_TESTS=false go test ./...`: every package passes, including the
  `testing/integration/{apiserver,mcpserver}` suites (which boot containers regardless of the
  flag).
- `golangci-lint run ./...` natively: clean apart from the two `gomoddirectives` hits on the
  replaces and the 65 `FromProto` deprecations.

## Next steps

1. Step 3: `indexstamp` + the three per-process syncer lists → `searchsync.Registry`
   (`RegisterIndex`, `PoolSpecs`, `ReindexAll`), including `cmd/tools/search_index_initializer`.
2. Step 4: the six `BuildInjector`s and `cmd/ddb/worker.go` → `service.Register/New/Run`.
   Revisit `docs/configuration.md`'s ruling against the `*/config` packages first —
   `service.Config` is that tree.
3. Step 5: `signingrpc.WithAccessTokens`, `rbac.Tiers`, `notifications/mail`, `*/mcp` +
   `oauth2server/mcp.Protect`, `metering.Unbilled`, `filtering.Observe`, `RandomQuery`,
   `billing/http`, `mediaregistry/grpc`, `QueryFilterFromProto`.
4. Step 6: frontend and iOS onto local paths for `platform-client-ts`/`-swift`.
5. When #1136 lands: `identitystore.Hooks` deletes down to `identity.NewRecordingHooks`. When
   #1137 lands: `identity_users_test.go` asserts the registration by actor again.
6. Tag the libraries; drop the replaces; delete this file.
