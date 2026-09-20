# platform-go v14: test port findings

A `replace`-directive port of this repo onto `platform-go` at `main` (untagged
v14) plus `primitives-go/v2` v2.2.0, done on the branch `v14-test-port` to
answer one question: **is there anything in v14 that would force a v15?**

Non-test code builds. Test code does not yet — 58 packages, all the same two
mechanical shapes — and mocks want regenerating. Nothing here is a bump: both
modules are wired with `replace`.

## What v14 is

Two modules where there was one.

- **`primitives-go/v2`** — the substrate: `database`, `filtering`, `errors`,
  `observability`, `encoding`, `uploads`, `search`, `routing`, `server`,
  `tenancy`, `pointer`, `identifiers`, `messagequeue`, `capitalism`. 131 of the
  200 platform packages this repo imported.
- **`platform-go/v14`** — the domains and their transports: `identity`,
  `comments`, `settings`, `waitlists`, `issuereports`, `billing`, `webhooks`,
  `audit`, `notifications`, `dataprivacy`, `metering`, `entitlements`,
  `mediaregistry`, `rbac`, `sessions`, `saga`, `outbox`, `workqueue`. 69.

Every one of the 200 paths this repo used maps to exactly one of them, with 11
renames where a package moved tier: `uploads/registry` → `mediaregistry`,
`search/sync` → `searchsync`, `authorization/database` → `rbac`,
`authentication/oauth2server/database` → `authentication/oauth2serverstore`,
`authentication/webauthn/database` → `authentication/webauthnsessions`.

## The two conventions that changed every call site

**A store holds no database handle.** Reads take a
`database.SQLQueryExecutor`, writes take a `database.Tx`, and the caller
supplies both. Measured across all 34 exported `Store`/`Reader`/`Dispatcher`
interfaces: 18 are fully converted, 6 are split, 3 take none. The split ones
(`dataprivacy`, `metering`, `webhooks`, `operations`, `saga`,
`notifications.Registry`) take no executor on exactly their worker-path methods
— `Claim`, `Reap`, `MarkFlushed`, `Stranded`, `Release`, `Requeue` — and the
rule is stated per method (`webhooks/doc.go:40`, `metering/doc.go:118`). The
three that take none at all are `links.Store` (cache-backed), `shredding.Store`
(may not be on the caller's database at all) and `dataprivacy.Service` (not a
store). **This is coherent and documented. It is not v15 pressure.**

**`tenancy.Scope` is a value type.** It was a string; it is now a struct that
cannot be built from nothing, with `Global()` as the explicit way to say "no
tenant". A read that lost its scope fails to compile rather than quietly
widening. This is the single largest source of mechanical churn in the port and
is worth every line of it.

## What the port bought

These are not speculative: each is a workaround this repo was carrying that came
out with the port.

- **platform-go #458** (waitlists erasure). The local eraser documented two
  departures from the `Eraser` contract — `Withdraw` owned its own transaction,
  and archived signups were unreachable. v14's `Withdraw` takes the caller's
  `Tx` and `ListSignupsForSubject` pages archived rows. Both caveats deleted.
- **platform-go #456** (oauth2 sweep interval). `SweepInterval` is a
  `*time.Duration`, so an explicit zero now means "no sweeper" instead of being
  rewritten to the ten-minute default. The comment explaining that the config
  did not take effect is gone.
- **Audit entries now commit with the writes they describe.** Seven repository
  decorators (`settings`, `comments`, `waitlists`, `payments`, `issuereports`,
  `uploadedmedia`, password-reset tokens) each carried a `record` helper that
  opened a transaction of its own, because the platform store's write had
  already committed. Every one of those comments said the same thing: the pair
  is not atomic, and the gap loses the entry rather than the write. v14 closes
  it — the write, the audit entry and the outbox event are one transaction.
- **Read-backs deleted.** Writes return the row they wrote, so
  `waitlists.recordedTransition`, `comments.ArchiveComment` and
  `mediaregistry.ArchiveObject` no longer need a read before or after the write
  to name what changed, along with the window each read opened.
- **`EnsurePackaging` and the fulfiller's `encrypted bool`** are gone, replaced
  by `WithCompressor`/`WithEncryptor` on the shared option set. Whether
  artifacts are encrypted is now read off the encryptor rather than off a flag
  beside it that could disagree with it.
- **Fenced work-queue completion.** `workqueue.Complete` takes the claimed
  `Item` rather than its key, so a completion cannot land against a lease
  another replica now holds. Handing back the key no longer compiles.

## Findings

### 1. Three constructors kept their path and name and lost their SQL backend

**Closed in primitives-go v2.3.0** (#16). Kept here for the record, because it is
the one finding this port hit by accident rather than by looking.

| path (unchanged) | v13 | primitives v2 | SQL version now at |
|---|---|---|---|
| `authentication/oauth2server/config` | `NewStore(ctx, cfg, db, opts...)` | `NewStore(ctx, cfg, opts...)` → **memory** | `platform/authentication/oauth2serverstore/config` |
| `authentication/webauthn/config` | `NewSessionStore(ctx, cfg, db, opts...)` | `NewSessionStore(ctx, cfg, opts...)` → **cache** | `platform/authentication/webauthnsessions/config` |
| `authorization/config` | `NewPolicyResolver(ctx, cfg, db, cache, opts...)` | `NewPolicyResolver(ctx, cfg, cache, opts...)` → **static** | `platform/rbac/config` |

A full-signature diff of every package present at the same path in v13 and in
primitives v2 finds exactly three exported constructors that lost a parameter,
and all three lost the database. The same three `Config` types lost their
`Provider string \`env:"PROVIDER"\`` field, so the environment variable that
selects the backend is read by nobody at those paths — and caarlos0/env does not
error on a variable no field claims.

In all three the *call* fails to compile, because the dropped parameter's type
does not match the one that took its place. The hazard is the fix: deleting the
argument compiles, and silently switches the backend. For `oauth2server` that
means OAuth2 codes and tokens in process memory — fine on one replica in dev,
and every login failing across two. For `authorization/config` it means
authorization decisions come from `cfg.Roles` rather than from the policy table,
and PR #1418 already found those two drifting apart on three of five roles, so
the direction of the change is not predictable.

This port hit the oauth2 one and was saved by the argument count. That is luck.

**Fixed as:** `Provider` is back on all three primitives-tier `Config`s with its
`env:"PROVIDER"` tag, and each constructor refuses anything but the provider it
actually builds — with an error naming the platform package that builds the SQL
one:

    oauth2 server store provider "database": this config builds the in-memory
    store only, and a SQL-backed one is built by platform-go's
    authentication/oauth2serverstore/config: unknown provider

Verified against the published tag: this repo now requires
`primitives-go/v2 v2.3.0` with no replace directive, builds clean, and all three
constructors refuse this application's own `PROVIDER=database` config when handed
to the primitives tier. The substitution that nearly happened during pass 1 is
now unrepresentable, and no consumer-side guard is needed — the refusal is the
guard.

### 2. `audit.NewReader` takes a `database.Client`; `audit.NewRecorder` takes a `dialect.Dialect`

`audit/reader.go:354` and `audit/recorder.go:114`. Both are stateless now —
every method takes the caller's executor — and the Reader uses its `Client`
argument for nothing but `client.Dialect()`. Same package, same release, two
answers to the same question, and the `Client` argument implies a handle the
Reader does not keep. Changing it after the tag is a rewrite at every
construction site. `NewReader(d dialect.Dialect, opts ...ReaderOption)`.

### 3. `WithServerOptions` exists at both config tiers with different element types

`oauth2serverstore/config.WithServerOptions` takes
`oauth2servercfg.Option`, which is the *other* `WithServerOptions`'s own option
type, so a wiring site that wants a server option writes
`WithServerOptions(WithServerOptions(x))` against two identically-named
functions. It reads as a mistake and is not one. Cosmetic, fixable in a patch
release — noted, not blocking.

### Checked and cleared

- The executor split (see above) — principled and documented per method.
- `oauth2serverstore.Store` keeping its client — it backs an external library's
  interface it does not control.
- `billing/privacy.AccountResolver` being a defined type where six siblings
  alias `dataprivacy.ScopeResolver` — a billing row carries both a scope and an
  account id and neither is inferable from the other.
- `dataprivacy.Collector.Collect` taking no executor, so every collector must
  capture a reader at construction — stated as a deliberate ruling in
  `settings/privacy/privacy.go:75` and applied consistently, including by
  platform's own collectors.

## What this means for the "mostly platform imports" goal

v14 ships 11 `.proto` files and complete gRPC servers — 134 RPCs across
`identity`, `billing`, `comments`, `settings`, `waitlists`, `issuereports`,
`webhooks`, `audit`, `notifications`, `signin`, `oauth2clients`.

Comparing RPC sets against this repo's own protos, **they are the same
operations under different names**: `GetSettingDefinitions`/`ListDefinitions`,
`GetWaitlists`/`ListLists`, `CreateWaitlist`/`CreateList`,
`RegisterDeviceToken`/`RegisterDevice`. This repo says `VerbDomainNoun`;
platform says `VerbNoun` with the domain implied by the service. Nothing local
is missing upstream except a short list of genuine extensions — avatars and
service-role admin on identity, trigger configs and the event catalog on
webhooks, `WaitlistIsOpen`, and audit's by-account/by-user read variants that
platform expresses as one `ListEntries` with a `Query`.

**The audit and outbox decorators are not in the way.** Every platform server
takes the `Store` *interface*, not the concrete `SQLStore`, and opens the
transaction itself before calling into it. This repo's decorator repositories
already implement those interfaces, so handing one to a platform server puts the
audit entry and the data-change event inside that server's transaction — which
is exactly what the decorators want and could not have before v14.

So the pass-2 shape is: **keep the decorator repositories, delete the local
proto, converters and gRPC services, mount platform's servers.** The domains
platform also ships account for ~85,700 lines here against mealplanning's
~158,300. Not all of it goes — identity and webhooks carry real local extension
— but the thin ones (comments, settings, waitlists, issuereports, uploadedmedia,
audit, notifications) are mostly forwarding and translation.

The cost is client-facing: adopting platform's protos renames RPCs and messages
across the generated TypeScript and Swift clients. Nothing is deployed, so that
is a regenerate and a frontend sweep rather than a compatibility problem.

## Verdict

Nothing found here forces a v15. The one thing worth fixing before the tag is
finding 1, and it is a rename.

---

# Pass 2 spike: comments

Done on the same branch. The standalone comments service, its proto and its
generated code are gone; platform's `comments/grpc.Server` is mounted over this
repo's existing repository. The backend builds and the comments suite passes.

## What the spike had to prove

The whole pass-2 plan rests on one claim: that this repo's audit entry and
outbox event survive adoption, because platform's server takes the
`comments.Store` interface and opens the transaction before calling into it. If
that were wrong — if the server committed the row and left the decorator's two
statements outside — then adopting any of the thirteen surfaces would mean
trading the audit log for the deletion, which is not a trade this application
can make.

`internal/repositories/postgres/comments/server_transaction_test.go` settles it
against a real Postgres, driving platform's server with this repo's repository
behind it and a real outbox writer:

- **It commits together.** One `CreateComment` through platform's surface leaves
  the comment, its audit entry and its outbox message behind.
- **It rolls back together.** With the audit repository made to fail, the call
  errors and *neither the comment nor the outbox row exists*.
- **A refused write records nothing.** An unknown target type is refused by the
  store, and no entry and no event are left behind.

The middle test is the load-bearing one, so it was checked against a mutant
rather than trusted for being green. Restoring the pre-v14 shape — the store
opening a transaction of its own, which is exactly what v13 did — makes it fail
on precisely the right assertion:

    Error:    Should be zero, but was 1
    Messages: the comment should have rolled back with the audit entry

So the test is not passing vacuously: it distinguishes v14's shape from v13's,
which is the only thing it was written to do.

## What came out, and what went in

| | lines |
|---|---|
| deleted: local service, converters, permissions, generated stubs, `proto/comments` | −2,611 |
| deleted: the four `AddCommentTo…` RPCs and their messages from the mealplanning and issue_reports protos | −722 |
| added: `sessions.Principal` + extractor (one-time, serves all thirteen) | +57 |
| added: `internal/build/comments` server wiring | +55 |
| added: the transaction tests | +296 |

Net −2,925 for the domain, and the only new production code that is not
comments-specific is the 57-line principal adapter.

## Four things the spike found that reading could not

**The convenience RPCs go too.** `AddCommentToRecipe`, `AddCommentToMeal`,
`AddCommentToMealPlan` and `AddCommentToIssueReport` lived on two *other*
services and looked like separate surface. Each is sugar: it names the target
from the RPC rather than the body and forwards to `CreateComment`. The existence
check they appear to add is the store's, through the target catalog. Platform's
`CreateComment` takes the target as a field, so all four are redundant — but
they mean comments cannot be adopted without regenerating two neighbouring
protos, which is the first thing that makes this a cross-domain change.

**`Permission` had to become an alias.** Platform's surfaces ship their own
method→permission maps, and `commentsgrpc.Permissions()` returns
`map[string][]primitives.Permission`. This repo's `Permission` was a defined
type, so the maps would not compose — one conversion per domain, thirteen times.
Platform documents the fix in `authorization/authorization.go:18`: declare the
local type as an alias. One line, and every existing constant, map key and
switch kept compiling.

**The permission strings change.** Local `"create.comments"` becomes platform's
`"comments.create"` — domain first, so that composed domains cannot collide on a
bare verb. `internal/authorization/comments_permissions.go` now re-exports
platform's constants under the names the policy already spells, so only the
seeded strings move. Free here because nothing is deployed; a deployed service
would need a policy migration per adopted domain.

**Two of the seams need nothing.** Platform's default `AuthorAuthorizer`,
`OwnCommentsOnly`, is exactly what the deleted service's `ownedComment` did, and
the absent `GrantsExtractor` clears `include_archived`, which is the behaviour
the deleted service had (it exposed no archived read at all). Both are left at
their defaults, with the reasons written down where the server is built.

## Two findings fixed upstream mid-spike

`audit.NewReader` now takes a `dialect.Dialect` rather than a `database.Client`,
and `oauth2serverstore/config.WithServerOptions` is now
`WithServerConfigOptions`. Both call sites here have been moved over. That
clears findings 2 and 3; finding 1 — the three constructors that silently lose
their SQL backend — is the one still open, and it is still the only thing worth
blocking the tag on.

## What this says about the remaining twelve

The shape holds and the cost is now measured rather than estimated. Per domain,
expect: delete the service, converters, proto and generated stubs; add a
build-layer registration of about fifty lines; re-export platform's permission
constants; and regenerate any neighbouring proto that imported the domain's
messages. The repository decorator — the thing that makes the write auditable —
is untouched, which is the whole reason this is worth doing.

What will not be uniform is the cross-domain proto coupling. Comments was
imported by two neighbours; identity and webhooks will be worse, and the
generated TypeScript and Swift clients have to be regenerated and swept either
way.

---

# Pass 2 spike: identity

Identity was the domain most likely to surprise us, and it did — but not in the
seam, which is the part that matters. Nothing here was ported: this spike
establishes the shape and proves the one mechanism the adoption would rest on.

## The seam is different, and it is better

Comments adopts by decorating `comments.Store`: platform's server opens the
transaction and calls the decorated store inside it, so this repo's audit entry
and outbox event are statements of that transaction.

That does not work for identity. An identity write goes through
`identity.Service`, which orchestrates several store calls in one transaction —
`Register` writes a user, an account and an owner membership — so decorating the
store would record one row of an operation that wrote three.

platform's answer is `identity.Hooks`: twenty-four methods, one per operation,
each handed the operation's `database.Tx`, with the documentation stating the
reason outright (`identity/hooks.go:11`):

> Every application that adopts this package has companions for an identity
> write — an audit entry, a data change event, a search index stamp, an outbox
> row — and those companions are the same fact as the row.

`internal/repositories/postgres/identityspike` proves it against real Postgres,
with platform's identity DDL migrated alongside this application's own schema:

- **Commits together.** A `Register` leaves the user, the account, the
  membership, the hook's row and the hook's outbox message.
- **Rolls back together.** With the hook made to fail, *none of the five exist* —
  including the three rows the operation had already written before the hook ran.

Checked against a mutant, as the comments test was: a hook that writes on a
connection of its own rather than the operation's `Tx` fails it on the right
assertion. So the test distinguishes a hook inside the transaction from one
beside it, which is the only thing it is for.

Hooks is a better seam than the store decorator — it sees whole operations
rather than individual rows — and the nine `recordAndEmit` sites in this repo's
identity repository map onto it directly.

## The surface maps almost exactly

Twenty-nine RPCs on each side, and the same operations under different names:
`GetAccounts`/`ListAccounts`, `ArchiveUserMembership`/`RemoveMembership`,
`UpdateAccountMemberPermissions`/`SetMembershipRoles`, and so on. Three of this
repo's RPCs collapse into one — `UpdateUserDetails`, `UpdateUserEmailAddress`
and `UpdateUserUsername` are all `UpdateProfile`, whose `ProfileUpdate` carries
username, display name, email address, first and last name.

Three do not map, and each for a stated reason:

- **`UploadUserAvatar`** stays local. Platform is explicit
  (`identity/proto/…/identity.proto:71`): "identity has no avatar column and
  joining one is a contract between two packages that has not been designed."
  A consumer with columns of its own puts them in a side table keyed by user id —
  which is exactly what `user_avatars` already is.
- **`CreateAccount`** stays local. `Store.CreateAccount` exists and platform's
  own documentation shows composing it with `CreateMembership` in one
  transaction; it is simply not a `Service` operation or an RPC.
- **`AdminSetPasswordChangeRequired`** keeps a thin local RPC over
  `Service.SetUserRequiresPasswordChange`, which exists but is deliberately off
  the wire — the proto argues credentials do not belong in the transport.

Platform adds four this repo has no equivalent for: `GetPrincipal`,
`GetMembership`, `ListMembershipsForUser`, `RecordAgreement`.

## What makes identity genuinely harder than comments

**Identity is the root of the schema.** Comments was a leaf on a table this repo
had already adopted. `users` and `accounts` are this application's own tables,
and **34 of the schema's 77 tables hold a foreign key to one of them**. Adopting
identity renames them to `ddb_identity_users` and `ddb_identity_accounts` and
re-points all 34 constraints. That is mechanical DDL, but it is the whole schema
rather than one corner of it.

The query layer is better than that sounds: SQL that actually reads `users` or
`accounts` is confined to four files, all in the identity package, all deleted
with the adoption. Every other domain carries `belongs_to_user` /
`belongs_to_account` as plain columns and never joins.

**The erasure model changes, and it is a product decision.** This repo's
`accounts` table carries
`belongs_to_user TEXT NOT NULL REFERENCES users("id") ON DELETE CASCADE`, and
twelve tables cascade from `accounts` in turn — so erasing a user today destroys
the households they owned and everything in them. That cascade is what the single
`EraserKeyIdentity` eraser relies on.

Platform's `identity_accounts.owner_user_id` has **no** foreign key, deliberately
(`identity/migrations/postgres.sql:117`): a cascade "would destroy an
organization, its invoices and every other member's work because one member
exercised a right to be forgotten," and `RESTRICT` would refuse an erasure that
has to commit. `Store.EraseUser` therefore leaves owned accounts standing with an
`owner_user_id` that resolves to nothing, and says so.

For an organization that is plainly right. For a household whose owner has left
it is a question this application has not had to answer, because the cascade
answered it. Adoption forces the answer, and it is not a mechanical one.

**Two schema improvements come free.** `user_account_status` and its explanation
are renamed; `service_role` is already dead here (the rbac adoption normalized
roles into `user_role_assignments`, which splits cleanly into platform's
`identity_user_roles` and `identity_membership_roles` — assignments, where rbac's
`authz_*` tables hold the policy, so the two do not overlap). And
`email_address_verification_token` becomes
`email_address_verification_token_digest`: this repo stores the **raw** token,
**indexes it**, and selects it into the generated rows for account and invitation
reads. Platform stores a SHA-256 digest, indexes the digest, and argues the case
in `identity/token_digest.go:8` — a database copy otherwise hands out every
outstanding verification link. That is one of the verification-token defects the
blocked links adoption (#1385) left standing.

`birthday` and `last_indexed_at` have no platform home and become the side table
platform names as the intended pattern.

## Verdict

The seam works and is better than comments'. The surface maps. What identity
costs is not code but schema: 34 foreign keys to re-point, and one product
decision about what happens to a household when its owner is erased.

**That decision is made**, and the blocker is cleared — see below.

## The succession rule

An erasure transfers the household to the longest-tenured remaining member, and
deletes it only if the owner was alone. platform leaves the account standing;
this application decides what standing means. No platform change, so identity
can start.

`internal/domain/identity/succession` is that rule, and it is production code
rather than spike code: it depends only on platform's `identity.Store` and a
`database.Tx`, so it is finished and tested now and gets wired when the port
lands. `internal/repositories/postgres/identityspike/succession_test.go` runs it
against a real database, composed the way the erasure will be — the rule, then
`EraseUser`, on one transaction:

- **Transfer.** Three members added newest-first; the earliest membership
  inherits, and the other two keep theirs.
- **Deletion.** A household the owner was alone in is gone.
- **The invariant.** One subject owning two households, one shared and one solo,
  so both paths run in a single erasure — and afterwards *no surviving household
  names an owner who no longer exists*.
- **Rollback.** A failure after `EraseUser` takes the transfer and the deletion
  back with it, so a retried erasure does not find households already moved.

Mutation-checked twice. Sorting by newest instead of earliest fails the transfer
test on "the longest-tenured member should have inherited the household". Making
the rule do nothing at all — which is exactly the state platform's `EraseUser`
leaves — fails the invariant on "every surviving household should name an owner
who still exists".

### Two decisions inside the rule worth knowing about

**Tenure ties break on the membership identifier.** Two people added in one
transaction share a `created_at`, and an erasure that picked between them
differently on a retry would hand the household to a different person the second
time. Arbitrary, but total.

**Deleting a solo household is a `DELETE` this application issues against
identity's own table**, because platform's `Store` offers `ArchiveAccount` and no
delete. Archiving would not be an erasure: the row keeps the name the erased
person chose, and every one of the twelve tables that cascade from an account —
the meal plans, the recipes, the webhooks, the subscriptions — survives, because
they cascade from a deletion rather than from a flag. A solo household's contents
are the erased person's data by definition. A `Store.DeleteAccount` would be the
tidier home and is worth asking platform for later; it is one visible statement
against a table platform owns, named in the package documentation so that a
schema change upstream lands on a comment rather than a surprise.

---

# Pass 2 spike: settings

Taken as insurance before the tag, on the expectation that the third thin domain
would repeat comments. It did not. **The adoption was stopped before any
deletion**, because mounting platform's settings surface silently drops an
authorization check this application performs today.

## What settings does that comments did not

`settings.Definition.AdminOnly` marks a setting only an administrator may write.
platform records it and deliberately does not enforce it, and says so
(`settings/settings.go:282`):

> It is recorded rather than enforced — this package has no notion of who is
> calling, and a store that pretended to would be an authorization check in the
> wrong layer. What it is for is the caller's own check.

This application is that caller. `internal/services/settings/grpc` is where the
check lives: `writableSetting` reads the definition before every value write and
refuses a non-administrator reaching for an admin-only one, and the read path
does the same.

## The gap

platform's own gRPC surface gives that check nowhere to stand.

- The method grant does not cover it. `SetValue` requires
  `settings.values.write`, which every self-service user must hold to set any of
  their own settings at all. An admin-only setting's value write is the same RPC
  with the same permission.
- The `SubjectAuthorizer` does not cover it, and cannot. It is handed the
  *subject* and never the *definition*, and it runs **before** the definition is
  read — `settings/grpc/values.go:69` calls `authorizeSubject` above the
  transaction that then does `GetDefinitionByName`.
- There is no other seam. The server takes `store`, `client`, `principals` and
  `subjects` positionally, and its options are a grants extractor and the three
  observability providers.

So: **a caller holding `settings.values.write` can set a value for a definition
marked `AdminOnly`, for themselves.** That is what the flag exists to prevent,
and what this repo prevents today.

`internal/repositories/postgres/settingsspike` demonstrates it rather than
asserting it — an ordinary service-user principal, platform's server wired with
the self-service `SubjectAuthorizer` platform's own documentation supplies
(`settings/grpc/authorizer.go:60`), and an admin-only definition. The write
succeeds and the value is read back. The test passing *is* the finding.

## How bad, honestly

The mechanism is real and reachable. The blast radius here today is small: the
only definitions this repo marks `AdminOnly` are two examples in the localdev
seed — a theme preference and a notification frequency — and both are arguably
mis-marked anyway. Nothing in production depends on it, because nothing is in
production.

It also does not generalize. "Recorded rather than enforced" appears in settings
and nowhere else in platform, so this is one seam on one surface rather than a
systemic hole.

## What it means for the tag

It does **not** force a `/v15`. An additive `WithDefinitionAuthorizer` option,
called just after `GetDefinitionByName` inside the write's transaction, closes it
in a minor release.

The reason to settle it before the tag anyway is the same reason findings 2 and 3
were settled before the tag: it is a question about the *shape* of a seam, and
the shape is free to change now and expensive later. If the right answer is one
authorizer that sees the subject and the definition together, that is a breaking
change to `SubjectAuthorizer` once consumers have written one. If the answer is a
second authorizer, consumers write two where one would have done, forever. Today
nobody has written either.

The failure mode also argues for settling it early: a consumer adopts the
surface, the check quietly stops happening, and nothing fails.

## What was not done

No deletion. The local settings service, proto and converters are untouched, and
the 1,901-line service still enforces `AdminOnly`. Adopting settings should wait
on the seam, at which point it looks like comments did — the store decorator
already implements `settings.Store`, so the audit entry and outbox event carry
over unchanged.

The workaround that does not need platform is enforcing `AdminOnly` in this
repo's store decorator, reading the session off the context. It works and it is
the wrong layer, which is precisely the objection platform raises about the store
doing it. Better to ask for the seam.
