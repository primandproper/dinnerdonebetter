# Completing the v14 port: what landed, and what stopped it

A run at finishing the v14 port on `v14-test-port`, taken because the settings
finding came from *doing* the port rather than reading it — two spikes had found
nothing and the third found a real hole, which is an argument for more surface
area rather than more reading.

**Landed: five surfaces adopted.** comments (earlier), settings, waitlists, issue
reports, audit. Each is platform's gRPC server over this repo's decorator store,
with the local service, proto and generated stubs deleted and the permission
constants re-exported from platform. The backend builds.

**Stopped: six surfaces.** Four are blocked on a store adoption that has to come
first (finding D). One is blocked on a product decision (F). One leaves a
remainder worth designing rather than guessing (G).

**New shared wiring**, both in `internal/authentication/sessions`:
`GrantsFromContext`, the authority half of what a platform surface reads off a
request; and `AccountScopedPrincipal`, because this application's domains do not
agree on one tenancy (finding B).

Findings are ordered by how much they matter, not by when they were found.
Findings B and D are the two worth reading first.

---


## A. waitlists: an anonymous Join writes a signup no privacy read can find
**CLOSED.** Nothing to change here — this repo already declares Join under its
own grant rather than mounting `PublicMethods()`, so every signup carries the
caller as subject. platform wrote the fork into `PublicMethods` and
`waitlists/privacy`'s docs on `pre-v14-final`.

**Where:** `waitlists/grpc/converters.go:218` (`signupFromJoin`), `permissions.go:131` (`PublicMethods`)
**Severity:** worth documenting upstream; not a defect, and fail-closed by default.

`signupFromJoin` sets `Signup.Subject` only when there is a caller, so an
anonymous Join writes a signup with an empty subject. That is deliberate and
documented — a pre-launch list's visitor has nothing to sign in to.

What is not connected anywhere: a subject-less signup is invisible to
`ListSignupsForSubject`, which is what `waitlists/privacy`'s collector pages and
what its eraser withdraws by. So a deployment that adopts `PublicMethods()`
wholesale silently creates rows that no subject access request will return and no
erasure will reach — and nothing says so at wiring time.

The safe direction is the default: `Permissions()` omits the three public methods
and the consumer's interceptor is fail-closed, so taking the map alone denies
them to everybody. The hazard is only in taking `PublicMethods()` as well,
without deciding.

**Suggested:** one sentence in `PublicMethods`' doc — or in the privacy
package's — saying an anonymous signup is outside the subject machinery. This
repo gates all three behind grants, because joining here requires a session.

## B. callers.Principal.Scope() assumes one tenancy model per consumer
**FIXED on `pre-v14-final`.** `callers/principal.go` now says the scope is the
surface's and not the consumer's, that a consumer whose domains disagree hands
each surface its own extractor, and that getting it wrong is quiet. This repo's
`sessions.AccountScopedPrincipal` is that, implemented.

**Where:** `callers/principal.go:72`, and every surface's `caller()` —
e.g. `issuereports/grpc/server.go:244`, `scope := principal.Scope()`
**Severity:** the most serious of this run. Silent cross-tenant exposure if wired wrong.

Every platform surface takes its tenant from one method on one type. The doc
frames that as the consumer's tenancy model: "An application with one directory
returns tenancy.Global here."

This consumer has two, by domain rather than by deployment. Comments, settings,
waitlists and uploaded media are `tenancy.Global()` — a recipe's discussion reads
the same for everybody, and scoping it per account would make one recipe's
comments look different depending on who was reading. Issue reports are
`tenancy.Of(accountID)`, because an account's reports are its own; that scoping
is what closed a member-visible leak in #1377.

A consumer who reads the doc and writes one `PrincipalExtractor` — which is the
natural reading — gets whichever model they wrote, everywhere. Wiring the global
one into issuereports would file every account's reports in the global scope and
serve all of them to everybody: the exact leak that scoping closed, reintroduced
silently, with no type error and no failing test unless somebody wrote one across
two accounts.

The mechanism is fine — `NewServer` takes the extractor per surface, so two
extractors is all it costs, and that is what this repo now does. What is missing
is the sentence saying a consumer may need more than one, and why.

**Suggested:** in `callers.Principal.Scope`'s doc, note that the scope is the
scope *of the surface being called*, and that a consumer whose domains differ in
tenancy supplies a different extractor per surface. One sentence; the failure it
prevents is a cross-tenant read.

## C. audit/grpc reads one scope, so "every entry about me" is not on the wire
**Where:** `audit/grpc/rpcs.go:93` — `query.Scope = pointer.To(req.scope)`
**Severity:** behaviour change to note, probably not a defect.

The surface binds the query's scope to the one the connection resolved, and the
proto argues at length why there is no scope field. Sound for an API read: you
are looking at *an account's* log.

This repo's audit chains are per-account *or* per-user — `ScopeFor` files an
entry under its account when it has one and under its actor otherwise, so that
logins do not all serialize on one chain. Its `GetAuditLogEntriesForUser` RPC
therefore filters on actor across chains, and that is not expressible here: a
user's entries live in every account they belong to plus their own chain.

The privacy collector is unaffected — it goes through the repository, not the
wire — so what is lost is the API read "everything about me", which arguably
belongs to the export anyway. Recorded because a consumer with per-user chains
will meet it, and the proto's reasoning does not mention that shape.

## D. Surface adoption is gated on store adoption
**Withdrawn as a finding.** A server taking its store is the design, not a gap.
Recorded here only because it is the shape of the remaining work: four of the six
outstanding surfaces need their tables on platform's schema before there is
anything to hand their server.

Every platform gRPC server takes that domain's store as its first argument:
`comments.Store`, `settings.Store`, `waitlists.Store`, `issuereports.Store`,
`billing.Store`, `notifications.Inbox` + `Registry`, `identity.Service` +
`identity.Store`. So a surface cannot be mounted over a domain whose *store*
this repo has not already adopted — there is nothing to hand it.

Where that leaves the eleven surfaces:

| surface | store adopted? | status |
|---|---|---|
| comments, settings, waitlists, issuereports, audit | yes | **adopted this run** |
| billing (payments) | yes | adoptable; stopped on finding F |
| webhooks | **no — see the correction below** | blocked on a store adoption |
| notifications | **no** — own querier, own tables | blocked on a store adoption |
| identity | partial (2 imports, privacy only) | blocked; 34 FK re-points |
| oauth2clients, signin | **no** | blocked on a store adoption |

### Correction: webhooks is a store migration, not a surface swap

This table first said webhooks was adoptable, on the strength of eleven
platform-webhooks imports under `internal/repositories`. Those imports are the
*delivery* machinery — the dispatcher, the endpoint rows it fans out to, the
attempt log. They are not the model.

What a client of this repo calls a webhook is a row in a **local** `webhooks`
table, with `name`, `method` and `created_by_user` columns platform's
`webhooks_endpoints` does not have, alongside local `webhook_trigger_configs`
and `webhook_trigger_events`. `repository.GetWebhook` reads it through this
repo's own generated querier. The two models run in parallel, joined by a shared
identifier, and the audit entries and outbox events for webhook writes are
written against the local tables.

So mounting platform's webhooks surface would change what a webhook *is* on the
wire, orphan three tables, and drop the audit trail for every webhook write —
because the `webhooks.Store` registered here is platform's raw store, undecorated.
There is no decorator to hand the server, unlike the five surfaces adopted this
run.

Webhooks therefore belongs with notifications, identity, oauth2clients and
signin: five surfaces blocked on a store adoption, not four. The `ListEventTypes`
ruling still stands and is still needed — it is simply needed later than this
correction implies.

So "adopt the remaining surfaces" is mostly not surface work. Four of the six
outstanding need their tables migrated onto platform's schema first, which is
the identity-shaped job: rename the tables, re-point the foreign keys, delete the
local querier. The surfaces then follow in an afternoon each, as these five did.

## E. billing/grpc alone does not gate include_archived
**FIXED on `pre-v14-final`, and it was bigger than reported.** Not one surface
but three — billing, notifications and webhooks — twelve read sites, and none of
the three had a grants seam at all. Each gains a `WithGrantsExtractor` whose
absence is fail-closed, and an `archived_test.go` that is one differential: same
rows, same request, grant present and absent. Retired webhook endpoint URLs were
among what leaked.

**Where:** `billing/grpc/options.go` (no `WithGrantsExtractor`),
`billing/grpc/products.go:114` — `filteringgrpc.FromProto(request.GetFilter())`
passed straight to the store.
**Severity:** low on its own; the inconsistency is the point.

comments/grpc, settings/grpc and waitlists/grpc each take a
`WithGrantsExtractor` and use it for exactly one decision: whether a paged read's
`include_archived` is honored or cleared. Each ships an `archived.go` saying so,
and each documents the absent-extractor default as "serves live rows to
everybody rather than removed ones to anybody".

billing/grpc has neither the option nor the file. A client-supplied
`include_archived` reaches the store, so any caller who may list products,
subscriptions, purchases or transactions may also page the archived ones.

The blast radius is small — the account-keyed reads are behind the
AccountAuthorizer and the cross-account ones behind their own `list_all` grants —
so what this really is, is four sibling surfaces that do not agree, with no
stated reason for the one that differs. A consumer wiring all four writes the
extractor three times and cannot tell whether the fourth omission is deliberate.

**Suggested:** either the option and the narrowing, or a sentence in
`billing/grpc/doc.go` saying archived billing rows are not sensitive and why.

## F. payments: two RPCs with no platform counterpart, and it is a product question
**Status: adoption stopped, decision needed. Not a platform defect.**

Ten of this repo's twelve payments RPCs map onto platform's billing surface as
renames — `GetProducts`/`ListProducts`, `GetPurchasesForAccount`/
`ListPurchasesForAccount`, and so on. Two do not:

  - `CreateSubscription`
  - `UpdateSubscription`

platform's billing surface has no client-facing way to create or change a
subscription, and that reads as deliberate rather than missing: a subscription is
what a payment provider reports, and `billing/grpc`'s writes are the archives and
the product catalog. This repo's subscription writes arrive the same way in
practice — `internal/domain/payments/manager` handles the RevenueCat webhook —
so these two RPCs are a second door onto rows the provider owns.

Dropping them is probably right and is not mine to decide. Adoption is stopped
here rather than guessed at, exactly as settings was.

Note also that `CreateCheckoutSessionPermission` and
`CancelSubscriptionPermission` exist in this repo's policy with no RPC in
proto/payments — they belong to the capitalism/RevenueCat handlers, which are
HTTP and outside this surface either way.

## G. webhooks: the event catalog is not on the wire
**Where:** `webhooks/proto/.../webhooks.proto` — ten RPCs, none of them a catalog read.
**Severity:** low, but it is the one thing that stops webhooks deleting cleanly.

Nine of this repo's ten webhook RPCs are renames of platform's —
`CreateWebhook`/`SaveEndpoint`, `AddWebhookTriggerConfig`/`AddSubscription`, and
so on. The tenth is `GetWebhookEventTypes`, and platform has no counterpart.

The catalog is a Go value supplied at construction — `webhooks.Catalog`, and
`Dispatcher.Catalog()` reads it — but nothing serves it. A client building a
subscription form has to be told what event types exist, and on this surface it
cannot ask.

So webhooks adopts to nine RPCs plus a one-RPC local service, or the catalog is
published some other way. Worth asking whether a `ListEventTypes` belongs on the
surface: the catalog is already the consumer's own value, so serving it is three
lines and it is the one question a subscription UI cannot answer without it.

Webhooks was not adopted this run — the remainder is small but it is a design
question, and the finding is the part worth having.

---

## The first store migration: notifications

Done, and it is the useful data point for the other four. Net −4,419 lines
(−5,400 / +981). One focused push, no platform change needed, nothing
finding-grade against platform.

**The shape that worked**, and should repeat:

1. `renderNotificationsDDL` — drop the local tables, render platform's at this
   application's prefix, **re-create the foreign keys**. Both principal columns
   name a user here, so the keys `belongs_to_user` carried are re-creatable, and
   re-creating them is what keeps the single identity eraser covering the tables.
   platform ships `notifications/privacy` with an eraser for each, which is the
   other way and the one a deployment with non-user principals has to take.
2. A store decorator — platform's store with the audit entry and data change
   event around its writes. Four writes record; marking read does not, because
   an entry per opened inbox buries the ones that matter, and the two erasure
   deletes do not, because the erasure records itself.
3. **An adapter presenting platform's store as the local `Repository`
   interface.** This is the piece that contained the blast radius: the manager,
   the push fanout and the privacy collector did not change at all. Without it
   the migration reaches four more packages and their tests.

**What the adapter costs**, stated so the next one is not a surprise: it is a
translation layer that exists to be deleted. Once nothing outside the store
package says `UserNotification`, the manager can take platform's types and the
adapter goes. Keeping it is what makes the migration one commit instead of five.

**Two small things platform does not offer**, neither a defect:

- No keyed device read. `Registry` lists but does not `GetDevice`, so a lookup by
  identifier is a list and a scan. A person has a handful of handsets, so the
  page is already small.
- No device update, deliberately: a handset is identified by its token, so
  changing the token is registering a different device. The adapter's
  `UpdateUserDeviceToken` refuses rather than pretending. Nothing had called it.

**For the remaining four**, in the order I would take them: webhooks (three
local tables, and a model change visible on the wire — plus the `ListEventTypes`
ruling to pull), then oauth2clients, then signin, then identity last, because
everything above touches `users` and `accounts` and doing identity first would
churn them twice.
