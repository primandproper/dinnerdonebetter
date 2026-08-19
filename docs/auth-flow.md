# Authentication Flow

This document describes how authentication works across the Dinner Done Better application. The system supports multiple auth methods, multiple client types, and a layered token model. For identity concepts (users, accounts, memberships), see [identity.md](identity.md).

## Overview

Authentication in this app is **convoluted by design**: there are several ways to log in, several token types, and the same token can flow through different paths depending on the client. The main sources of complexity:

1. **Multiple auth methods**: Password+TOTP and Passkey (WebAuthn)
2. **Two token systems**: JWT (from `LoginForToken`) and OAuth2 (opaque, stored as a digest, used for gRPC)
3. **Multiple client types**: Consumer web app, Admin web app, mobile apps, API clients, integration tests
4. **Different paths for different clients**: Web apps use cookies + OAuth2 exchange; some clients send JWT directly

## Token Types

### JWT (from LoginForToken / ProcessLogin / ProcessPasskeyLogin)

- **Issued by**: `internal/authentication/manager.go` via `ProcessLogin`, `ProcessPasskeyLogin`, or `ExchangeTokenForUser`
- **Format**: JWT (or PASETO, configurable) with claims: `sub` (user ID), `account_id` (optional), `sid` (session ID), `jti` (unique token ID), `exp`, `aud`, `iss`
- **Lifetime**: Configurable (e.g. 5–10 min access, 72h refresh)
- **Used for**:
  - Stored in web app auth cookie
  - Input to OAuth2 authorization flow (Bearer token proves user identity)
  - Direct Bearer token for gRPC when using `WithBearerTokenCredentials` (e.g. localdev, tests)

**Implementation**: [`internal/authentication/tokens/jwt/jwt.go`](backend/internal/authentication/tokens/jwt/jwt.go)

### OAuth2 Access Token

- **Issued by**: the OAuth 2.1 authorization server at `/authorize` + `/token`
- **Format**: opaque string; the store holds only its hex SHA-256 digest, never the token
- **Lifetime**: 15 minutes access, refresh token rotating with reuse detection
- **Used for**: gRPC `Authorization: Bearer <token>` when clients use the full OAuth2 flow

Opaque and looked up rather than signed and verified locally: that is what makes a revoked
token stop working on the next request rather than at the end of its lifetime. The cost is a
store read per request, which is the trade `oauth2server`'s package doc argues.

**Implementation**: [`internal/services/auth/handlers/authentication/oauth2.go`](backend/internal/services/auth/handlers/authentication/oauth2.go), [`oauth2_store.go`](backend/internal/services/auth/handlers/authentication/oauth2_store.go), [`oauth2_authenticator.go`](backend/internal/services/auth/handlers/authentication/oauth2_authenticator.go)

## Auth Methods

### 1. Password + TOTP (LoginForToken / AdminLoginForToken)

**Flow**:

1. Client calls gRPC `LoginForToken` (or `AdminLoginForToken` for admin-only) with username, password, and optionally TOTP.
2. `ProcessLogin` validates credentials via `Authenticator.CredentialsAreValid` (password + TOTP if 2FA verified).
3. Manager creates a server-side session record in the `user_sessions` table with device metadata (IP, User-Agent).
4. Manager issues JWT with `IssueToken` (user ID + account ID + session ID + JTI).
5. Client receives `TokenResponse` with `AccessToken` and `RefreshToken`.

**Entry points**:

- **Consumer web app**: `POST /login/submit` → `LoginForToken` → JWT stored in cookie
- **Admin web app**: `POST /login/submit` → `AdminLoginForToken` → JWT stored in cookie
- **gRPC clients**: Call `LoginForToken` directly, use token as Bearer or exchange for OAuth2

**Implementation**: [`internal/authentication/manager.go`](backend/internal/authentication/manager.go), [`internal/services/auth/grpc/auth.go`](backend/internal/services/auth/grpc/auth.go)

### 2. Passkey (WebAuthn)

**Flow**:

1. Client calls `BeginPasskeyAuthentication` (unauthenticated) with optional username.
2. Server returns `PublicKeyCredentialRequestOptions` and challenge; challenge stored in session.
3. User completes passkey assertion in browser.
4. Client calls `FinishPasskeyAuthentication` with assertion response and challenge.
5. WebAuthn service validates assertion, returns user ID.
6. `ProcessPasskeyLogin` creates a session record and issues JWT (same as password login).
7. Client stores JWT in cookie (web app) or uses it directly.

**Entry points**:

- **Consumer web app**: `POST /auth/passkey/authentication/options`, `POST /auth/passkey/authentication/verify`
- **Admin web app**: Same routes
- **gRPC**: `BeginPasskeyAuthentication`, `FinishPasskeyAuthentication`

**Implementation**: [`internal/authentication/webauthn/service.go`](backend/internal/authentication/webauthn/service.go), [`frontend/consumer/src/routes/auth/passkey/`](frontend/consumer/src/routes/auth/passkey/)

**Where the ceremony lives**: the ceremony itself — issuing the challenge, storing it between the
two requests, verifying the response against it — is platform's
`authentication/webauthn.RelyingParty`, configured under `auth.passkey`. What this repository owns
is the `webauthn.User` adapter, where a registered credential is stored, and the sign count written
back after a login. Two things about it are worth knowing:

- **The ceremony store is a table, in every environment.** A ceremony spans two requests and
  nothing pins them to a pod, so a per-process store fails a fraction of passkey logins on a
  multi-replica deployment in a way that reads as a browser bug. `auth.passkey.provider` defaults
  to `database`; there is no in-memory option to fall into by leaving it blank. Rows are swept
  every `auth.passkey.sweepInterval` (5 minutes by default) and the challenge is *consumed* — read
  and removed in one operation — so an assertion replayed inside its TTL finds nothing.
- **One timeout, not three.** `auth.passkey.relyingParty.ceremonyTimeout` is the timeout the
  browser is asked to honor, the deadline go-webauthn enforces when the response comes back, and
  the TTL the row is stored under. It is 2 minutes, chosen to cover a cross-device prompt (scan the
  QR code, approve on the phone) rather than only a local touch.
- **The origins come from `internal/branding`.** `RPOrigins` is
  `branding.WebAppOrigins()` in prod and `branding.LocalDevWebAppOrigins()` locally, so a
  rebrand or a moved port changes them in one place. An origin the config does not name is
  every passkey ceremony on that host failing verification.

Both ceremonies are covered end to end by `testing/integration/apiserver/auth_passkey_test.go`,
which drives a virtual ES256 authenticator (`auth_passkey_authenticator.go`) against a real
Postgres: registration, username login, usernameless login, replay of both an attestation and an
assertion, and the sign count advancing between logins.

## Web App Auth Flow (Consumer / Admin)

The consumer and admin frontends use the same pattern:

1. **Login**: User submits credentials (password or passkey) → `LoginForToken` or passkey handlers → JWT returned.
2. **Cookie**: JWT is encoded and stored in a signed cookie (`AuthPayload{AccessToken}`).
3. **Per-request**: `AuthMiddleware` reads cookie, decodes JWT, calls `BuildAuthedClient(ctx, config, accessToken, developingLocally)`.
4. **Client build**:
   - **Production**: `WithOAuth2Credentials` — uses JWT as Bearer to hit `POST /authorize`, gets code, exchanges for OAuth2 token, uses OAuth2 token for gRPC.
   - **Local dev**: `BuildInsecureOAuthedGRPCClient` — same OAuth2 flow but over HTTP.
5. **gRPC calls**: Authenticated client sends OAuth2 access token (or JWT in some paths) as `Authorization: Bearer <token>`.

**Implementation**: [`internal/platform/webappauth/middleware.go`](backend/internal/platform/webappauth/middleware.go), [`internal/platform/webappauth/client_builder.go`](backend/internal/platform/webappauth/client_builder.go)

## gRPC Auth Interceptor

Every gRPC request (except unauthenticated routes) goes through `AuthInterceptor`:

1. **Extract token**: Read `Authorization: Bearer <token>` from metadata.
2. **Resolve session** (in order):
   - **OAuth2 first**: `oauth2Server.Authenticate(ctx, accessToken)` — a Store lookup by digest. If it resolves, the token's audience is checked against this server's own resource identifier (RFC 8707) and the subject's `sub` becomes the user ID. An audience naming somewhere else — the MCP server, which shares this database and therefore this store — is refused; an empty audience is accepted, because a client that sent no `resource` parameter gets one.
   - **JWT fallback**: `tokenIssuer.ParseToken(ctx, accessToken)` — if OAuth2 fails, treat as JWT.
3. **Validate session** (JWT path only): If the token has a `sid` claim, extract the `jti` and look up the session in `user_sessions`. If the session has been revoked or expired, return 401. Tokens without `sid` (pre-session-management) skip this check. Asynchronously updates `last_active_at` on the session.
4. **Build session context**: `identityDataManager.BuildSessionContextDataForUser(ctx, userID, accountID)`. The session ID is attached to `ContextData.SessionID`.
5. **Zuck mode** (optional): If `X-Zuck-Mode-User` header present and user can impersonate, override session with that user/account.
6. **Permissions**: Check method’s required permissions against session; deny if missing.
7. **Inject context**: Store `sessions.ContextData` in request context for handlers.

**Unauthenticated routes** (skip interceptor):

- `LoginForToken`, `AdminLoginForToken`
- `BeginPasskeyAuthentication`, `FinishPasskeyAuthentication`
- `CreateUser`, `VerifyTOTPSecret`
- `RequestPasswordResetToken`, `RedeemPasswordResetToken`
- `VerifyEmailAddress`

**Implementation**: [`internal/services/auth/grpc/interceptors/authn_interceptor.go`](backend/internal/services/auth/grpc/interceptors/authn_interceptor.go)

## OAuth2 Flow (for gRPC clients)

Clients that use OAuth2 (e.g. web app via `WithOAuth2Credentials`) follow this flow:

1. Client has a JWT (from `LoginForToken` or equivalent).
2. Client calls `POST /authorize?client_id=X&state=Y&code_challenge=…&code_challenge_method=S256&…`
   with `Authorization: Bearer <JWT>`.
3. The `SubjectAuthenticator` parses the JWT, extracts `sub` (user ID), and resolves the account
   the authorization is granted against.
4. The authorization server issues an authorization code and redirects to `redirect_uri?code=Z`,
   with `state` and the RFC 9207 `iss` parameter echoed back.
5. Client (with `CheckRedirect = ErrUseLastResponse`) reads `code` from the `Location` header.
6. Client calls `POST /token` with `code`, `code_verifier`, `client_id`, `client_secret` →
   receives OAuth2 access + refresh tokens.
7. Client uses the OAuth2 access token for gRPC `Authorization: Bearer <oauth2_access_token>`.

**POST, not GET, at step 2.** A `GET /authorize` renders the login form — the answer for a
browser arriving without a session — and only a POST runs the authenticator that reads the
bearer token. Both methods carry the authorization parameters in the query string, so the
request that issues the code is validated against exactly the same bytes the GET would have been.

**Endpoints** (API server):

- `GET /.well-known/oauth-authorization-server` — RFC 8414 discovery
- `GET|POST /authorize` — authorization (GET renders the login form; POST authenticates)
- `POST /token` — token exchange (`authorization_code` and `refresh_token` only)
- `POST /revoke` — RFC 7009 token revocation

`POST /register` is **not** served. RFC 7591 dynamic registration is open by construction, and an
OAuth2 client here is an administered object created through the permission-gated gRPC surface —
an anonymous endpoint writing to the same registry would be a way around those permissions. The
discovery document therefore omits `registration_endpoint` rather than advertising a 404.

### What the API server's authorization server enforces

- **Redirect URIs matched byte for byte**, at `/authorize` and again at `/token` against the URI
  the code was issued for. Not by hostname, not ignoring ports. A client registers the exact
  string it will send.
- **PKCE mandatory, S256 only.** No `plain`, and an absent `code_challenge_method` is refused
  rather than defaulted — RFC 7636 defaults it to `plain`, so silence is a request for the
  method this server does not accept.
- **Refresh token rotation with reuse detection.** A replayed refresh token revokes the whole
  family, not just itself.
- **Every credential stored as a hex SHA-256 digest**, never as itself.
- **`authorization_code` and `refresh_token` only.** No password grant, no client credentials, no
  implicit.

### Two servers, one Store

`ddb serve` and `ddb serve mcp` are different processes running the same authorization server
package over the same four `ddb_oauth2_*` tables. That is the intended shape: `oauth2server`
ships no RFC 7662 introspection endpoint, so a resource server either shares the Store or lives
in the same process, and both of these reach the same Postgres.

It is also why the audience check in the gRPC interceptor is load-bearing rather than decorative.
A token minted by the MCP server is in the same table this server reads, so what stops it being
spent here is that its audience names somewhere else.

## The MCP server's authorization server

`ddb serve mcp` runs the same `oauth2server` package, over the same tables, at the same paths.
What differs is the two things the package deliberately leaves to the application:

|                     | API server                                     | MCP server                                     |
|---------------------|------------------------------------------------|------------------------------------------------|
| Who the subject is  | a session JWT, or a username + password + TOTP | a username, argon2 password, and TOTP — admins only |
| Client registration | administered, via the gRPC surface             | RFC 7591 dynamic, open, 90-day expiry          |
| `POST /register`    | not served                                     | served                                         |

Everything else — the endpoints, the tables, exact redirect URI matching, mandatory S256 PKCE,
rotation with reuse detection, opaque 15-minute access tokens — is one implementation.

The MCP side is documented in [`backend/docs/mcp-usage-guide.md`](backend/docs/mcp-usage-guide.md).

## Session Context

After auth, handlers receive `sessions.ContextData` in the request context. It contains:

- `Requester`: User ID, username, email, account status, service role
- `ActiveAccountID`: Account for this request
- `AccountPermissions`: Map of account ID → role checker
- `SessionID`: The server-side session ID (from the `sid` claim; empty for OAuth2 tokens or pre-session JWTs)

**Implementation**: [`internal/authentication/sessions/session_context.go`](backend/internal/authentication/sessions/session_context.go)

## Session Management

Users can view and manage their active login sessions. Each login (password, passkey) creates a record in the `user_sessions` table that tracks:

- **Device metadata**: Client IP, User-Agent, friendly device name (derived from User-Agent)
- **Login method**: `password` or `passkey`
- **Activity**: `created_at`, `last_active_at` (updated asynchronously on each request), `expires_at`
- **Token linkage**: `session_token_id` (access token JTI) and `refresh_token_id` (refresh token JTI), rotated on each token refresh

### Token Refresh and Session Continuity

When a client calls `ExchangeToken` with a refresh token, the system:

1. Extracts the `jti` and `sid` from the refresh token.
2. Looks up the session by refresh token JTI — rejects if revoked.
3. Issues new access + refresh tokens with the same `sid` but new JTIs.
4. Updates the session record with the new JTIs and expiration.

This means the session ID is stable across token refreshes, while individual tokens rotate.

### gRPC Endpoints

- **`ListActiveSessions`**: Returns all active (non-revoked, non-expired) sessions for the current user with pagination. Each session includes an `is_current` flag.
- **`RevokeSession`**: Revokes a specific session by ID. The revoked session's tokens are rejected on the next request.
- **`RevokeAllOtherSessions`**: Revokes all sessions except the one making the request.

### Backward Compatibility

Tokens issued before session management (without a `sid` claim) continue to work — the interceptor skips session validation for these tokens. As old tokens expire, all active tokens will have session tracking.

**Implementation**: [`internal/domain/auth/user_session.go`](backend/internal/domain/auth/user_session.go), [`internal/repositories/postgres/auth/user_sessions.go`](backend/internal/repositories/postgres/auth/user_sessions.go), [`internal/services/auth/grpc/auth.go`](backend/internal/services/auth/grpc/auth.go)

## Key File Reference

| Area                                          | Path                                                                                |
|-----------------------------------------------|-------------------------------------------------------------------------------------|
| Auth manager (login, passkey, token exchange) | `internal/authentication/manager.go`                                                |
| JWT issuance/parsing                          | `internal/authentication/tokens/jwt/jwt.go`                                         |
| WebAuthn service (credentials, sign count)    | `internal/authentication/webauthn/service.go`                                       |
| WebAuthn user adapter                         | `internal/authentication/webauthn/user_adapter.go`                                  |
| Passkey ceremony config + wiring              | `internal/services/auth/grpc/do.go`                                                 |
| gRPC auth service                             | `internal/services/auth/grpc/auth.go`                                               |
| Auth interceptor                              | `internal/services/auth/grpc/interceptors/authn_interceptor.go`                     |
| OAuth2 server                                 | `internal/services/auth/handlers/authentication/oauth2.go`                          |
| OAuth2 client registry store                  | `internal/services/auth/handlers/authentication/oauth2_store.go`                    |
| OAuth2 subject authenticator                  | `internal/services/auth/handlers/authentication/oauth2_authenticator.go`            |
| Passkey HTTP endpoints (web)                  | `frontend/consumer/src/routes/auth/passkey/`                                        |
| Web app auth middleware                       | `internal/platform/webappauth/middleware.go`                                        |
| Client builder (OAuth2 + JWT)                 | `internal/platform/webappauth/client_builder.go`                                    |
| gRPC client (OAuth2, Bearer)                  | `pkg/client/client.go`                                                              |
| Session domain model + interfaces             | `internal/domain/auth/user_session.go`                                              |
| Session DB repository                         | `internal/repositories/postgres/auth/user_sessions.go`                              |
| Session DB migration                          | `internal/repositories/postgres/migrations/migration_files/00023_user_sessions.sql` |

## Flow Diagram

```mermaid
flowchart TB
    subgraph "Login Entry Points"
        PW[Password + TOTP]
        PK[Passkey]
    end

    subgraph "Token Issuance"
        PM[ProcessLogin / ProcessPasskeyLogin]
        SESS[(user_sessions table)]
        JWT[JWT with user_id + account_id + sid + jti]
    end

    PW --> PM
    PK --> PM

    PM --> SESS
    PM --> JWT

    subgraph "Web App Path"
        Cookie[Auth Cookie]
        MW[AuthMiddleware]
        BC[BuildAuthedClient]
        OAuth2[OAuth2 Exchange]
        OAT[OAuth2 Access Token]
    end

    JWT --> Cookie
    Cookie --> MW
    MW --> BC
    BC --> OAuth2
    OAuth2 --> OAT

    subgraph "gRPC Path"
        OAT --> gRPC[gRPC with Bearer]
        JWT --> gRPC
    end

    subgraph "Auth Interceptor"
        AI[Extract Bearer]
        O2Check{OAuth2 token?}
        JWTCheck{JWT?}
        SessCheck{Session valid?}
        SC[Session Context]
    end

    gRPC --> AI
    AI --> O2Check
    O2Check -->|yes| SC
    O2Check -->|no| JWTCheck
    JWTCheck -->|yes| SessCheck
    JWTCheck -->|no| Err[401 Unauthenticated]
    SessCheck -->|yes| SC
    SessCheck -->|revoked/expired| Err

    subgraph "Session Management"
        LS[ListActiveSessions]
        RS[RevokeSession]
        RA[RevokeAllOtherSessions]
    end

    SC --> LS
    SC --> RS
    SC --> RA
    RS --> SESS
    RA --> SESS
```

## Related Documentation

- [identity.md](identity.md) — Users, accounts, memberships, roles, permissions
- [email_verification.md](email_verification.md) — Email verification flow
- [backend/docs/adding_a_new_domain.md](backend/docs/adding_a_new_domain.md) — Authorization permissions for new domains
