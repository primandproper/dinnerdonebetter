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

## D. Surface adoption is gated on store adoption, and that is the real remaining work
**Not a defect — a sequencing fact worth stating plainly.**

Every platform gRPC server takes that domain's store as its first argument:
`comments.Store`, `settings.Store`, `waitlists.Store`, `issuereports.Store`,
`billing.Store`, `notifications.Inbox` + `Registry`, `identity.Service` +
`identity.Store`. So a surface cannot be mounted over a domain whose *store*
this repo has not already adopted — there is nothing to hand it.

Where that leaves the eleven surfaces:

| surface | store adopted? | status |
|---|---|---|
| comments, settings, waitlists, issuereports, audit | yes | **adopted this run** |
| billing (payments), webhooks | yes | adoptable |
| notifications | **no** — own querier, own tables | blocked on a store adoption |
| identity | partial (2 imports, privacy only) | blocked; 34 FK re-points |
| oauth2clients, signin | **no** | blocked on a store adoption |

So "adopt the remaining surfaces" is mostly not surface work. Four of the six
outstanding need their tables migrated onto platform's schema first, which is
the identity-shaped job: rename the tables, re-point the foreign keys, delete the
local querier. The surfaces then follow in an afternoon each, as these five did.

## E. billing/grpc alone does not gate include_archived
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
