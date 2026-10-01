# Authentication Flow

This document describes how authentication works across the Dinner Done Better application. For
identity concepts (users, accounts, memberships), see [identity.md](identity.md).

## Overview

Signing in is platform-go's. Every door that proves who somebody is goes through one
`signin.Service`, built in `internal/authentication/do.go`, and the gRPC surfaces a client talks
to are platform's, mounted on this server:

| Surface                                          | What it is for                                                                                   | Mounted in                       |
|--------------------------------------------------|--------------------------------------------------------------------------------------------------|----------------------------------|
| `primandproper.platform.signin.v1.SignInService` | Password + TOTP sign-in, refresh, sign-out, the caller's own credentials and logins, auth status | `internal/build/signin`          |
| `SignInAdministrationService`                    | An operator listing and ending somebody else's logins                                            | `internal/build/signin`          |
| `PasswordResetService`                           | A link mailed to somebody who cannot sign in, and the password they choose with it               | `internal/build/passwordreset`   |
| `PasskeysService`                                | Passkey enrollment, listing, archiving, and sign-in                                              | `internal/build/passkeys`        |
| `auth.AuthService` (this application's)          | `RegisterUser` and `ExchangeToken`, until platform ships replacements                            | `internal/services/auth/grpc`    |
| `internalops.InternalOperations.ImpersonateUser` | An operator acting as somebody else                                                              | `internal/services/internalops`  |

There are two token systems: the tokens `signin.Service` mints (a signed access token naming a
login, and an opaque rotating refresh token), and the OAuth2 authorization server's opaque
access tokens, which the web apps and API clients exchange a sign-in token for.

What this application still decides, and where:

- **Who a registrant is** — `authentication.RegistrationPolicy`: good standing (no verification
  gate), the `service_user` role, an issued but unproven TOTP secret, ownership of their account,
  and both agreements, refused without them.
- **What a password must be** — `authentication.PasswordPolicy`, applied by every door that
  writes one: registration, change, and reset.
- **When a second factor is asked for** — `signin.SecondFactorWhenEnrolled`: only once somebody
  has proven one. The administrative door demands one whatever the policy says.
- **Who may impersonate** — `authentication.NewImpersonationPolicy`: an operator whose service
  roles grant `imitate.user`.
- **What gets recorded** — the sign-in hooks (`authentication/signin_hooks.go`) write the "logged
  in" event, the credential-change audit entries, and the impersonation record on sign-in's own
  transaction; the mailers (`authentication/signin_mailers.go`) turn platform's mails into outbox
  events the data change message handler renders.

## Tokens

### Sign-in tokens

- **Issued by**: `signin.Service` — through `SignInService.LoginForToken` / `AdminLoginForToken`,
  `PasskeysService.FinishLogin`, `AuthService.ExchangeToken`, or
  `InternalOperations.ImpersonateUser`.
- **Access token**: signed (PASETO, configured under `auth.tokens`), with `sub`, `account_id`,
  `scope` (the directory, present and empty for this application's global one), `sid` (the
  login: its refresh token family), `jti`, and on an impersonation `actor_id` and `actor_scope`.
  An hour by default; fifteen minutes on an administrative or impersonation login.
- **Refresh token**: opaque, single use, rotated on every exchange; presenting a spent one ends
  the whole login. Thirty days by default, twelve hours administrative, none on an impersonation.
  Stored in `ddb_signin_refresh_tokens`, keyed to the user so erasure takes them with it.
- **Used for**: the web apps' cookie, the bearer token a client sends directly, and the input to
  the OAuth2 authorization flow.

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

**Implementation**: [`internal/services/auth/handlers/authentication/oauth2.go`](../backend/internal/services/auth/handlers/authentication/oauth2.go), [`oauth2_authenticator.go`](../backend/internal/services/auth/handlers/authentication/oauth2_authenticator.go)

## Signing up

`AuthService.RegisterUser` is the one door somebody with no session signs up through. It is
`signin.Service.Register` with no registrar, which platform's `SignInService.Register` will not
do: there, `Register` is an operator's call, gated by `PermissionCreateUsers`. Both run through
`RegistrationPolicy`, so they register the same person. The response carries the TOTP secret and
its QR code, the one time either is handed back.

A registrant signs in at once, with their password alone: their second factor is unproven until
they call `SignInService.VerifyTOTPSecret`, and nothing is asked of an unproven one.

`RegisterUser` moves to platform when `SignInService` has an open sign-up door
([platform-go#1068](https://github.com/primandproper/platform-go/issues/1068)).

## Signing in

### Password + TOTP

1. The client calls `SignInService.LoginForToken` (or `AdminLoginForToken`) with a username or
   email address, the password, a TOTP code once the user has proven a second factor, and
   optionally the account to sign into.
2. `signin.Service` reads the user, proves the password (hashing even for a handle naming nobody),
   checks their standing, then the second factor, then resolves the account.
3. The sign-in hook writes the `user_logged_in` event on the transaction that writes the refresh
   token.
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

`AuthService.ExchangeToken(refresh_token, desired_account_id)` moves a login into another of the
user's accounts without a credential prompt. The refresh token is its whole authority. It is built
from platform's doors until `SignInService` can do it in one
([platform-go#1069](https://github.com/primandproper/platform-go/issues/1069)):

1. The refresh token is exchanged, which proves the login and retires the token.
2. With no account named, or the login's own, that successor is the answer.
3. Otherwise a new login is issued on the named account through `IssueForPrincipal`, which
   refuses an account the user is not a live member of, and the rotated login is signed out. A
   refused switch signs it out too: its successor was never handed back.

A switch is therefore a new login, listed separately by `ListSignIns`; platform's version will
keep it one.

## Impersonation

An operator acts as somebody through `InternalOperations.ImpersonateUser(subject_id, account_id)`,
which returns a fifteen-minute token with no refresh token, minted by
`signin.Service.IssueImpersonationToken`.

- **The operator is the caller**, never a request field. Platform exposes no RPC for this door
  because it takes an operator's ID on trust; this one reads it off the authenticated session.
- **Who may**: an operator holding `imitate.user` (the permission table refuses everybody else,
  and `NewImpersonationPolicy` refuses again inside the sign-in service). Somebody already acting
  through an impersonation may not start another.
- **What the token is**: the subject's — their identity, account, memberships and rows — with the
  operator named on it (`actor_id`). The interceptor carries the operator on the session
  (`ContextData.ImpersonatorID`) and gives the request the operator's service-level grants in
  place of the subject's, because an operator acting in an account is still doing operator work.
- **What is recorded**: the impersonation itself, as an audit entry filed under the operator and a
  `user_impersonated` event; and every audit entry the request writes as the subject names the
  operator as its impersonator — DDB's own repositories through
  `auditlogentries.attachImpersonator`, platform's surfaces through `callers.Delegated` on the
  session's principal.
- **Ending it**: it expires, or the subject ends it from `ListSignIns`, where it appears with the
  operator named. Suspending the operator stops it on their next request.

## Password reset

`PasswordResetService`, every RPC of which is anonymous:

1. `RequestPasswordReset(email_address)` answers every address the same way, in the same time.
   For a real one, `passwordreset.Service` issues a token and hands the mail to
   `authentication.SignInMailers`, which writes a `password_reset_token_created` event carrying the
   secret under `password_reset_token.secret`. The data change message handler renders the email.
2. `VerifyPasswordResetToken` lets the form say whether a link is still good.
3. `CompletePasswordReset(token, new_password)` checks the password against `PasswordPolicy`,
   spends the token, writes the password and revokes the user's other links, on one transaction.
   The password write goes through `authentication.PasswordResetDirectory`, which queues the
   "your password was reset" mail (`password_reset_token_redeemed`) on that same transaction.

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

1. **Login**: the user signs in (password or passkey) and the web app receives a sign-in token.
2. **Cookie**: the token is stored in a signed cookie.
3. **Per-request**: the auth middleware reads the cookie and builds an authenticated client.
4. **Client build**: the sign-in token is the bearer at `POST /authorize`; the code is exchanged
   for an OAuth2 token, which the client uses for gRPC.

The web frontend and the iOS app have not been moved off the AuthService RPCs this change deleted,
and are fixed separately.

## gRPC Auth Interceptor

Every gRPC request goes through `AuthInterceptor`
([`authn_interceptor.go`](../backend/internal/services/auth/grpc/interceptors/authn_interceptor.go)):

1. **Unauthenticated routes** go straight through: `AuthService.RegisterUser` and `ExchangeToken`;
   `SignInService`'s sign-in, refresh and sign-out doors and those whose authority is a mailed link
   (see `internal/build/signin`); every `PasswordResetService` RPC; `PasskeysService.BeginLogin`
   and `FinishLogin`; and the anonymous analytics event.
2. **Optionally authenticated routes** (`SignInService.GetAuthStatus`, the waitlist signup page)
   answer an anonymous caller, and hold a caller who sends a token to it.
3. **Resolve the caller** from `Authorization: Bearer <token>`:
   - **OAuth2 first**: a store lookup by digest. The token's audience is checked against this
     server's resource identifier (RFC 8707); one naming somewhere else — the MCP server, which
     shares this store — is refused.
   - **Otherwise a sign-in token**: its signature, its directory (`scope` has to be this one), its
     login (`sid` has to name one, and that login has to be live), and on an impersonation, that
     the operator still stands.
4. **Build the session**: `SessionBuilder.BuildSessionContextDataForUser(user, account)`, which
   refuses a user whose standing does not admit sign-in (`PermissionDenied`, not
   `Unauthenticated`: the token is genuine).
5. **Permissions**: the method's required permissions, from the aggregated table in
   `internal/build/services/api/grpc/extras.go`. A method no table names is denied.
6. **Forced password change**: a user told to change their password may only read who they are
   (`IdentityService.GetPrincipal`, `SignInService.GetAuthStatus`), change it
   (`SignInService.UpdatePassword`), or `SignOutEverywhere`.

## OAuth2 Flow (for gRPC clients)

1. The client has a sign-in token.
2. It calls `POST /authorize?client_id=X&state=Y&code_challenge=…&code_challenge_method=S256&…`
   with `Authorization: Bearer <token>`.
3. The `SubjectAuthenticator` resolves the user, and the account the authorization is granted
   against.
4. The authorization server redirects to `redirect_uri?code=Z`, echoing `state` and the RFC 9207
   `iss` parameter.
5. The client reads `code` off the `Location` header without following it, and calls
   `POST /token` with `code`, `code_verifier`, `client_id`, `client_secret`.
6. It uses the OAuth2 access token for gRPC.

**POST, not GET, at step 2.** A `GET /authorize` renders the login form — the answer for a
browser arriving without a session — and only a POST runs the authenticator that reads the
bearer token.

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

|                     | API server                         | MCP server                                          |
|---------------------|------------------------------------|-----------------------------------------------------|
| Who the subject is  | a sign-in token                    | a username, argon2 password, and TOTP — admins only |
| Client registration | administered, via the gRPC surface | RFC 7591 dynamic, open, 90-day expiry               |
| `POST /register`    | not served                         | served                                              |

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

## Sweeping

`ddb job db-cleaner` removes expired rows from the authorization server's tables, the password
reset tokens and the sign-in refresh tokens — one scheduled sweep for the fleet rather than a
sweeper goroutine in every replica. It is a garbage collector, not a security control: every read
already refuses an expired row.

## Key File Reference

| Area                                              | Path                                                                     |
|---------------------------------------------------|--------------------------------------------------------------------------|
| Sign-in service, policies, hooks, mailers         | `internal/authentication/`                                               |
| SignInService mount and its permissions           | `internal/build/signin/grpc.go`                                          |
| PasswordResetService mount                        | `internal/build/passwordreset/grpc.go`                                   |
| PasskeysService mount, relying party              | `internal/build/passkeys/grpc.go`                                        |
| AuthService (RegisterUser, ExchangeToken)         | `internal/services/auth/grpc/auth.go`                                    |
| Impersonation RPC                                 | `internal/services/internalops/grpc/impersonation.go`                    |
| Auth interceptor                                  | `internal/services/auth/grpc/interceptors/authn_interceptor.go`          |
| OAuth2 server and subject authenticator           | `internal/services/auth/handlers/authentication/`                        |
| Password reset store (audit wrapper)              | `internal/repositories/postgres/auth/password_reset_tokens.go`           |
| Sign-in refresh token store                       | `internal/repositories/postgres/auth/refresh_tokens.go`                  |
| Expired-row sweep                                 | `internal/services/oauth/workers/db_cleaner/db_cleaner.go`               |
| gRPC client                                       | `pkg/client/client.go`                                                   |

## Related Documentation

- [identity.md](identity.md) — Users, accounts, memberships, roles, permissions
- [email_verification.md](email_verification.md) — Email verification flow
- [backend/docs/adding_a_new_domain.md](../backend/docs/adding_a_new_domain.md) — Authorization permissions for new domains
