# Authentication Flow

This document describes how authentication works across the Dinner Done Better application. For
identity concepts (users, accounts, memberships), see [identity.md](identity.md).

## Overview

Signing in is platform-go's. Every door that proves who somebody is goes through one
`signin.Service`, built in `internal/authentication/do.go`, and the gRPC surfaces a client talks
to are platform's, mounted on this server:

| Surface                                          | What it is for                                                                                                               | Mounted in                      |
|--------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------|---------------------------------|
| `primandproper.platform.signin.v1.SignInService` | Sign-up, password + TOTP sign-in, refresh, account switching, sign-out, the caller's own credentials and logins, auth status | `internal/build/signin`         |
| `SignInAdministrationService`                    | An operator listing and ending somebody else's logins                                                                        | `internal/build/signin`         |
| `PasswordResetService`                           | A link mailed to somebody who cannot sign in, and the password they choose with it                                           | `internal/build/passwordreset`  |
| `PasskeysService`                                | Passkey enrollment, listing, archiving, and sign-in                                                                          | `internal/build/passkeys`       |
| `internalops.InternalOperations.ImpersonateUser` | An operator acting as somebody else                                                                                          | `internal/services/internalops` |

There are two token systems: the tokens `signin.Service` mints (a signed access token naming a
login, and an opaque rotating refresh token), and the OAuth2 authorization server's opaque
access tokens, which API clients exchange a sign-in token for. The web apps send the sign-in
token itself.

What this application still decides, and where:

- **Who a registrant is** — `authentication.RegistrationPolicy`: good standing (no verification
  gate), the `service_user` role, an issued but unproven TOTP secret, ownership of their account,
  and both agreements, refused without them.
- **What a password must be** — `authentication.PasswordPolicy`, applied by every door that
  writes one: registration, change, and reset. `authentication.AccountPasswordPolicy` adds the
  rule that needs the account: a change may not keep the password it is changing.
- **When a second factor is asked for** — `signin.SecondFactorWhenEnrolled`: only once somebody
  has proven one. The administrative door demands one whatever the policy says.
- **Who may impersonate** — `authentication.NewImpersonationPolicy`: an operator whose service
  roles grant `imitate.user`.
- **Who is an operator** — `authorization.AdministrativeServiceRoleNames` (`service_admin` and
  `service_data_admin`), admitted only through the administrative door. A token from any other
  door keeps `authorization.OrdinaryServiceRoles`: the person (`service_user`), never the
  operator.
- **What gets recorded** — nothing this application decides any more. The sign-in service is
  built with platform's `signin.RecordingHooks`, which write every door's audit entry and event
  (`signin.user.authenticated`, `signin.password.updated`, the second-factor and revocation
  events) on the door's own transaction, filed under the user. What stays this application's is
  the mail: the mailers (`authentication/signin_mailers.go`) turn platform's mails into outbox
  events the data change message handler renders.

## Tokens

### Sign-in tokens

- **Issued by**: `signin.Service` — through `SignInService.LoginForToken` / `AdminLoginForToken`,
  `ExchangeRefreshToken` / `SwitchAccount`, `PasskeysService.FinishLogin`, or
  `InternalOperations.ImpersonateUser`.
- **Access token**: signed (PASETO, configured under `auth.tokens`), with `sub`, `account_id`,
  `scope` (the directory, present and empty for this application's global one), `sid` (the
  login: its refresh token family), `jti`, and on an impersonation `actor_id` and `actor_scope`.
  An hour by default; fifteen minutes on an administrative or impersonation login.
- **Refresh token**: opaque, single use, rotated on every exchange; presenting a spent one ends
  the whole login. Thirty days by default, twelve hours administrative, none on an impersonation.
  Stored in `ddb_signin_refresh_tokens`, keyed to the user so erasure takes them with it.
- **Used for**: the web apps' cookie and every call they make, the bearer token any other
  client sends directly, and the input to the OAuth2 authorization flow.

The login is read on every request (`signin.Service.CheckSignIn`), so a sign-out, an ended
login, or a detected refresh-token reuse stops the access token on its next request rather than
when it expires.

### OAuth2 access tokens

- **Issued by**: the OAuth 2.1 authorization server at `/authorize` + `/token`
- **Format**: opaque string; the store holds only its hex SHA-256 digest, never the token
- **Lifetime**: 15 minutes access, refresh token rotating with reuse detection
- **Used for**: gRPC `Authorization: Bearer <token>` when clients use the full OAuth2 flow

Opaque and looked up rather than signed and verified locally: that is what makes a revoked
token stop working on the next request rather than at the end of its lifetime.

**Implementation**: [`internal/services/auth/handlers/authentication/oauth2.go`](../backend/internal/services/auth/handlers/authentication/oauth2.go), [`oauth2_session_resolver.go`](../backend/internal/services/auth/handlers/authentication/oauth2_session_resolver.go)

## Signing up

`SignInService.Register` is open to anybody, signed in or not. What stands in front of it is
`RegistrationPolicy`, which signin runs on every registration before anything is written: no
agreements, no registration. The response carries the TOTP secret and its provisioning URI, the
one time either is handed back; a client renders the QR code from the URI. A registration may
answer an invitation, in which case it joins the inviter's account rather than creating one.

A registrant signs in at once, with their password alone: their second factor is unproven until
they call `SignInService.VerifyTOTPSecret`, and nothing is asked of an unproven one.

## Signing in

### Password + TOTP

1. The client calls `SignInService.LoginForToken` (or `AdminLoginForToken`) with a username or
   email address, the password, a TOTP code once the user has proven a second factor, and
   optionally the account to sign into.
2. `signin.Service` reads the user, proves the password (hashing even for a handle naming nobody),
   checks their standing, then the second factor, then resolves the account.
3. platform's recording hook writes the `signin.user.authenticated` audit entry and event on the
   transaction that writes the refresh token, naming the credential kind and the account.
4. The client receives an `IssuedToken`: the access token, the refresh token, and their deadlines.

**Refreshing**: `SignInService.ExchangeRefreshToken`. **Signing out**: `SignOut` (this login, by
its refresh token), `SignOutEverywhere`, `EndSignIn` / `EndOtherSignIns` (from `ListSignIns`).

### Passkey

1. A signed-in user enrolls through `PasskeysService.BeginRegistration` and `FinishRegistration`.
2. To sign in, the client calls `BeginLogin` — with a username, or without one for a discoverable
   credential — and answers the options with the authenticator's assertion in `FinishLogin`.
3. `FinishLogin` mints through `signin.Service.IssueForPrincipal`, so a passkey login is a login
   like any other: the same token, refresh token, hooks and event. A passkey asserted with user
   verification counts as both factors; one without is asked for a TOTP code beside it.

A credential's WebAuthn user handle is the user's ID. The ceremony state lives in a table
(`webauthn_sessions`), consumed on use, so a ceremony works across replicas and a replayed
response finds nothing:

- **The ceremony store is a table, in every environment.** `auth.passkey.provider` defaults to
  `database`; there is no in-memory option to fall into by leaving it blank. Rows are swept every
  `auth.passkey.sweepInterval` (5 minutes by default).
- **One timeout, not three.** `auth.passkey.relyingParty.ceremonyTimeout` is the timeout the
  browser is asked to honor, the deadline go-webauthn enforces, and the TTL the row is stored
  under: 2 minutes, to cover a cross-device prompt.
- **The origins come from `internal/branding`.** `RPOrigins` is `branding.WebAppOrigins()` in
  prod and `branding.LocalDevWebAppOrigins()` locally. An origin the config does not name is
  every passkey ceremony on that host failing verification.

Archiving a user's last passkey is refused only when they hold no password to fall back on.

### Switching households

`SignInService.SwitchAccount(refresh_token, account_id)` moves a login into another of the user's
accounts without a credential prompt; the refresh token is its whole authority. The successor
stays in the same login — one entry in `ListSignIns`, ended by one `EndSignIn` — and an account the
user is not a live member of is refused with the login left where it was. The sign-in hook records
the move as `changed_active_account`.

## Impersonation

An operator acts as somebody through `InternalOperations.ImpersonateUser(subject_id, account_id)`,
which returns a fifteen-minute token with no refresh token, minted by
`signin.Service.IssueImpersonationToken`.

- **The operator is the caller**, never a request field. Platform exposes no RPC for this door
  because it takes an operator's ID on trust; this one reads it off the authenticated session.
- **Who may**: an operator holding `imitate.user` (the permission table refuses everybody else,
  and `NewImpersonationPolicy` refuses again inside the sign-in service). Somebody already acting
  through an impersonation may not start another.
- **What the token is**: the subject's — their identity, account, memberships, rows and grants —
  with the operator named on it (`actor_id`). The interceptor carries the operator on the session
  (`ContextData.ImpersonatorID`) and nothing of the operator's grants: an impersonation is not an
  administrative login, so it carries no service role, and operator work is done with the
  operator's own token. Verifying an account's audit chain is an account admin's grant for this
  reason, so impersonating the owner can still ask.
- **What is recorded**: the impersonation itself, as the subject's `signin.user.authenticated`
  entry and event with the operator named on both — the entry's actor is the subject with the
  operator as its `Impersonator`, the event's `actorID` is the operator — written by platform's
  recording hook through `RecordAs`, since the request that mints the token carries no principal
  yet; and every audit entry the request then writes as the subject names the operator as its
  impersonator — DDB's own repositories through `auditlogentries.attachImpersonator`, platform's
  surfaces through `callers.Delegated` on the session's principal.
- **Ending it**: it expires, or the subject ends it from `ListSignIns`, where it appears with the
  operator named. Suspending the operator stops it on their next request.

## Password reset

`PasswordResetService`, every RPC of which is anonymous:

1. `RequestPasswordReset(email_address)` answers every address the same way, in the same time.
   For a real one, `passwordreset.Service` issues a token and hands the mail, secret included, to
   platform's `notifications/mail` `QueuedMailer`, which queues it on the mail topic once the token
   has committed. The async message handler's mail `Drainer` renders the email and sends it.
2. `VerifyPasswordResetToken` lets the form say whether a link is still good.
3. `CompletePasswordReset(token, new_password)` checks the password against `PasswordPolicy`,
   spends the token, writes the password through the identity store and revokes the user's other
   links, on one transaction. Spending the token fires platform's recording hook on the token
   store, which writes the `passwordreset.token.redeemed` audit entry and event on that same
   transaction; the data change message handler renders the "your password was reset" mail from
   that event. Issuing a link is recorded the same way, as `passwordreset.token.issued`; neither
   event carries the secret.

What the store guarantees:

- **The token is stored as a digest, never as itself.** `ddb_password_reset_tokens.token_digest`
  holds SHA-256 of the secret. The secret exists once, in the issuance, and travels only on the
  event the mail is rendered from — never on a span, a log line, or an audit entry.
- **Single use is the store's decision.** Two requests answering one link at once both find the
  row live; exactly one spends it.
- **Expiry is refused rather than swept.** The `ddb job db-cleaner` sweep reclaims rows; it is not
  what makes a link expire.

A link lives thirty minutes. Issuing and spending one are audited as `password_reset_tokens`
created/updated, by a wrapper around the platform store.

## Web App Auth Flow (Consumer / Admin)

Both web apps are SvelteKit servers that hold the login in `@primandproper/platform-client`'s
`Session`: one per request, over the library's `encryptedCredentialStore`, which seals the token
with AES-256-GCM under `COOKIE_ENCRYPTION_KEY` into an HTTP-only cookie. The cookie holds the whole `IssuedToken`, refresh token included, and never reaches the
browser's scripts. Every call, to platform's services and this repository's alike, goes through
`Session.call` with the access token as `Authorization: Bearer <token>`; neither app exchanges it
for an OAuth2 token.

1. **Sign-in**: the consumer app calls `signIn` (`LoginForToken`), the admin app `adminSignIn`
   (`AdminLoginForToken`). A second-factor refusal is branched on by its reason,
   `SECOND_FACTOR_REQUIRED`, and the form sends the password again with the code. A passkey
   sign-in in either app is `beginPasskeySignIn` and `passkeySignIn` through `PasskeysService`,
   which mints the same login a password does, with the browser's half of the ceremony carried
   by the library's WebAuthn bridge (`parseAssertionOptions`, `serializeAssertion`). A key tapped
   with no user verification is one factor, so a person with a second factor is asked for their
   code and taps again.
2. **Per request**: the hook builds the `Session` and gates the request with `resolveOrRedirect`,
   which redirects to `/login` when no login is held. It doesn't call the server.
3. **Calls**: `Session.call` refreshes within thirty seconds of expiry, and on `Unauthenticated`
   refreshes and retries once. A successor is written back to the cookie. Requests that arrive
   together with the same cookie exchange its refresh token once between them, through the
   process's `InMemoryExchangeCoordinator`; more than one replica needs a
   `SharedExchangeCoordinator`.
4. **Ended logins**: a refresh the server refuses clears the cookie. A page navigation
   redirects to `/login`, and a form action or API endpoint answers with its own error.
5. **Forced password changes**: somebody an operator has made change their password is signed in
   anyway, and every other call is refused with `PASSWORD_CHANGE_REQUIRED`. Sign-in reads
   `GetAuthStatus` and sends them to `/change_password` (`UpdatePassword`) when the change is
   owed, and a call refused for it later redirects there too.
6. **Sign-out**: `signOut`, which ends the login through `SignOut` and then clears the cookie.

The consumer's account pages are platform's surfaces too: `/account/sessions` is `ListSignIns`,
`EndSignIn` and `EndOtherSignIns`; `/account/passkeys` is `PasskeysService`; the reset and
verification links are `PasswordResetService` and `SignInService.VerifyEmailAddress`. The admin
app's per-user sessions page is `SignInAdministrationService`. A control the caller may or may
not use, such as editing the household or inviting to it, is shown from the effective
permissions `IdentityService.GetPrincipal` answers, never from the names of the caller's roles.

**Implementation**: [`frontend/consumer/src/hooks.server.ts`](../frontend/consumer/src/hooks.server.ts),
[`frontend/consumer/src/lib/auth/session.ts`](../frontend/consumer/src/lib/auth/session.ts),
[`frontend/consumer/src/lib/grpc/clients.ts`](../frontend/consumer/src/lib/grpc/clients.ts), and
their counterparts under `frontend/admin/src`.

The iOS app has not been moved off the deleted AuthService, and is fixed separately.

## gRPC Auth Interceptor

Every gRPC request goes through `AuthInterceptor`
([`authn_interceptor.go`](../backend/internal/services/auth/grpc/interceptors/authn_interceptor.go)):

Who is calling is platform's: the interceptor runs `signingrpc.PrincipalExtractor`'s own
interceptor and only renders what it resolved.

1. **Unauthenticated routes** go straight through: `SignInService`'s sign-in, refresh, account
   switch and sign-out doors and those whose authority is a mailed link (see
   `internal/build/signin`); every `PasswordResetService` RPC; `PasskeysService.BeginLogin`
   and `FinishLogin`; and the anonymous analytics event. The six that test a credential or send
   mail are throttled first — see [Rate limiting](#rate-limiting).
2. **Optionally authenticated routes** (`SignInService.GetAuthStatus` and `Register`, the waitlist
   signup page) answer an anonymous caller, and hold a caller who sends a token to it: a token
   that no longer works is `Unauthenticated`, not a visitor.
3. **Resolve the caller** from `Authorization: Bearer <token>`, through the extractor:
   - **A sign-in token first**: its signature, its login (`sid` has to name one, and that login has
     to be live), the principal (refusing a user whose standing does not admit sign-in with
     `PermissionDenied`: the token is genuine), and on an impersonation, that the operator still
     stands. Service roles ride only on a token whose `administrative` claim is set — the
     administrative door's.
   - **Otherwise an OAuth2 access token**, through `oauth2server.Verifier`: a store lookup by
     digest, and an audience that has to name this server (RFC 8707). A token naming no resource
     is refused as well as one naming somewhere else — the MCP server shares this store. Its
     subject keeps the person's service roles and none of the operator's.
4. **Forced password change**: the extractor's gate holds a user told to change their password
   at platform's `PasswordChangeMethods` — reading who they are, changing it, resetting it, and
   ending their logins.
5. **Build the session**: `SessionBuilder.SessionForPrincipal`, which turns the resolved
   principal's role names into permissions. It reads nothing; the extractor already did.
6. **Permissions**: primitives' authorization `Enforcer`, the next interceptor in both the unary
   and the stream chain, checks the method's required permissions against the session's grants,
   from the aggregated table in `internal/build/services/api/grpc/extras.go`. A method no table
   names is denied. `AuthInterceptor` itself decides only who is calling.

The HTTP routes resolve their caller the same way, through `AuthInterceptor.HTTPMiddleware`.

## OAuth2 Flow (for gRPC clients)

1. The client has a sign-in token.
2. It calls `POST /authorize?client_id=X&state=Y&code_challenge=…&code_challenge_method=S256&…`
   with `Authorization: Bearer <token>`.
3. The session resolver (`NewSessionResolver`, wrapped in `authserver.NewGuardedResolver`)
   resolves the user and their account through the same extractor every API request is resolved
   by. A signed-out session, a banned user, and an impersonation token are sent to the login
   form instead, which does not take a token. The login form itself is
   `authserver.NewAuthenticator`, which signs in through `signin.Service.Authenticate`. Both
   check `oauth2clients.Client.Admits` against the registration.
4. The authorization server redirects to `redirect_uri?code=Z`, echoing `state` and the RFC 9207
   `iss` parameter.
5. The client reads `code` off the `Location` header without following it, and calls
   `POST /token` with `code`, `code_verifier`, `client_id`, `client_secret`.
6. It uses the OAuth2 access token for gRPC.

**Name the resource.** Both legs carry `resource=<this server's identifier>` (RFC 8707). The API
server refuses an access token with no audience.

**GET or POST at step 2.** The session resolver is consulted on both, so a request carrying a live
sign-in gets a code either way. A request carrying none gets the login form on GET, and is signed
in from the form's fields on POST.

**Endpoints** (API server): `GET /.well-known/oauth-authorization-server`, `GET|POST /authorize`,
`POST /token` (`authorization_code` and `refresh_token` only), `POST /revoke` (RFC 7009).
`POST /register` is **not** served: an OAuth2 client here is an administered object created
through the permission-gated gRPC surface.

What the authorization server enforces: redirect URIs matched byte for byte, at `/authorize` and
again at `/token`; PKCE mandatory, S256 only; refresh token rotation with reuse detection; every
credential stored as a digest; no password, client-credentials or implicit grants. Platform's
`oauth2server` conformance suite asserts the storage half of this against this deployment.

### Two servers, one Store

`ddb serve` and `ddb serve mcp` run the same authorization server package over the same
`ddb_oauth2_*` tables. That is why the audience check in the interceptor is load-bearing: a token
minted by the MCP server is in the table this server reads, and what stops it being spent here is
that its audience names somewhere else.

|                     | API server                         | MCP server                                                                                 |
|---------------------|------------------------------------|--------------------------------------------------------------------------------------------|
| Who the subject is  | a sign-in token, or the login form | `signin.Service.AdminAuthenticate` — `service_admin` only, a proven second factor required |
| Client registration | administered, via the gRPC surface | RFC 7591 dynamic, open, 90-day expiry                                                      |
| `POST /register`    | not served                         | served                                                                                     |

The MCP side is documented in [`backend/docs/mcp-usage-guide.md`](../backend/docs/mcp-usage-guide.md).

## Session Context

After auth, handlers receive `sessions.ContextData` in the request context:

- `Requester`: user ID, username, email, account status, service permissions
- `ActiveAccountID`: the account this request is against
- `AccountPermissions`: account ID → role checker, for every account the user belongs to
- `SignInFamilyID`: the login a sign-in token belongs to (empty on an OAuth2 token)
- `ImpersonatorID`: the operator acting through an impersonation token (empty otherwise)

Platform's surfaces read the same session through `sessions.PrincipalFromContext`, whose
principal reports the impersonator as `callers.Delegated`.

## Where you're signed in

Every token issued for a login records the device it was issued to — the address, the user agent,
and a device name the client gave — on the token's own transaction, in `sign_in_devices`, keyed by
the login. `ListSignIns` and `ListSignInsForUser` answer each login with it as `attributes`
(`ip_address`, `user_agent`, `device_name`). The web apps forward the browser they are serving in
`x-client-address` and `x-client-user-agent`; the iOS app names itself in `x-device-name`.
An impersonation records nothing: the request behind it is the operator's. What a client forwards
is display and decides nothing. The rows are exported with the rest of a person's data, erased with
them, and swept once the login could no longer be alive.

**Implementation**: [`internal/authentication/devices`](../backend/internal/authentication/devices).

## Rate limiting

The doors a caller reaches with no credential and that test one or send mail —
`LoginForToken`, `AdminLoginForToken`, `Register`, `RequestHandleReminder`,
`RequestVerificationEmailByAddress`, `PasswordResetService.RequestPasswordReset`, and the OAuth2
login form's `POST /authorize`, on both the API and the MCP server — are throttled per address and per door, ahead of authentication,
by `primitives-go/ratelimiting`. The address is the last `X-Forwarded-For` entry, which Caddy
writes and a client cannot, or the connection's when there is none; nothing a client writes is a
key. Caddy sees the caller's own address only because its load balancer Service is
`externalTrafficPolicy: Local`; under the default, a node rewrites the source to its own address
and every caller would share that node's budget. The web apps reach the API through Caddy too, so their sign-ins share the cluster's egress
address and its budget. A refusal is `RESOURCE_EXHAUSTED` (429 over HTTP) with when to retry.
Configured under `services.auth.rateLimiting` (the MCP server's `rateLimiting` is rendered from
the same block); the limiter is in memory, so each replica holds a
budget of its own.

**Implementation**: [`ratelimit.go`](../backend/internal/services/auth/grpc/interceptors/ratelimit.go).

## Sweeping

`ddb job db-cleaner` removes expired rows from the authorization server's tables, the password
reset tokens, the sign-in refresh tokens and the sign-in devices — one scheduled sweep for the
fleet rather than a sweeper goroutine in every replica. It is a garbage collector, not a security
control: every read already refuses an expired row.

## Key File Reference

| Area                                              | Path                                                                     |
|---------------------------------------------------|--------------------------------------------------------------------------|
| Sign-in service, policies, hooks, mailers         | `internal/authentication/`                                               |
| SignInService mount and its permissions           | `internal/build/signin/grpc.go`                                          |
| PasswordResetService mount                        | `internal/build/passwordreset/grpc.go`                                   |
| PasskeysService mount, relying party              | `internal/build/passkeys/grpc.go`                                        |
| Impersonation RPC                                 | `internal/services/internalops/grpc/impersonation.go`                    |
| Auth interceptor                                  | `internal/services/auth/grpc/interceptors/authn_interceptor.go`          |
| OAuth2 server and session resolver                | `internal/services/auth/handlers/authentication/`                        |
| Sign-in devices                                   | `internal/authentication/devices/`                                       |
| Anonymous door rate limiting                      | `internal/services/auth/grpc/interceptors/ratelimit.go`                  |
| Password reset store (audit wrapper)              | `internal/repositories/postgres/auth/password_reset_tokens.go`           |
| Sign-in refresh token store                       | `internal/repositories/postgres/auth/refresh_tokens.go`                  |
| Expired-row sweep                                 | `internal/services/oauth/workers/db_cleaner/db_cleaner.go`               |
| gRPC client                                       | `pkg/client/client.go`                                                   |

## Related Documentation

- [identity.md](identity.md) — Users, accounts, memberships, roles, permissions
- [email_verification.md](email_verification.md) — Email verification flow
- [backend/docs/adding_a_new_domain.md](../backend/docs/adding_a_new_domain.md) — Authorization permissions for new domains
