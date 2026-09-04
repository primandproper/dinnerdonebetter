# Identity System Documentation

This document describes the identity and authentication system used in this application. The system is built around three core concepts: **Users**, **Accounts**, and **Account Memberships**.

## Core Concepts

### Users

Users represent individual people in the system. Each user has:

- A unique identifier (`ID`)
- Authentication credentials (username, email, password)
- Personal information (first name, last name, birthday)
- A service-level role (admin, user, etc.)
- Account status (active, banned, terminated)

**Domain Definition**: [`internal/domain/identity/user.go`](internal/domain/identity/user.go)

### Accounts

Accounts represent organizations or groups that users can belong to. Most data in the system is associated with accounts rather than individual users. Each account has:

- A unique identifier (`ID`)
- A name and contact information
- A billing status
- A webhook encryption key (unused; webhooks sign per endpoint — see `docs/webhooks.md`)
- A list of members

**Domain Definition**: [`internal/domain/identity/account.go`](internal/domain/identity/account.go)

### Account Memberships

Account memberships define the relationship between users and accounts. Each membership has:

- A unique identifier (`ID`)
- A user ID (`BelongsToUser`)
- An account ID (`BelongsToAccount`)
- An account role (admin or member)
- A flag indicating if this is the user's default account

**Domain Definition**: [`internal/domain/identity/account_user_membership.go`](internal/domain/identity/account_user_membership.go)

## Data Ownership Model

The system uses a clear data ownership model where most data is associated with accounts rather than individual users. This is indicated by the presence of `BelongsToUser` or `BelongsToAccount` fields in data structures.

**Examples**:

- User profile data: `BelongsToUser`
- Account settings: `BelongsToAccount`
- Meal plans: `BelongsToAccount`
- Webhooks: `BelongsToAccount`

## Authentication and Session Management

**For a complete description of the auth flow, including password, passkey, OAuth2, and gRPC interceptor behavior, see [auth-flow.md](auth-flow.md).**

### Token-Based Authentication

The system uses token-based authentication with JWT and PASETO support (PASETO is currently configured). Sessions are not stored server-side but are persisted through tokens.

### Authentication Flow (Summary)

1. User provides credentials (username/email + password + TOTP if 2FA is enabled), or uses passkey
2. System validates credentials and retrieves user information
3. System issues a token (JWT/PASETO) containing user ID and account information
4. Client uses this token to obtain OAuth2 credentials via the OAuth2 exchange process (web app) or sends JWT directly (some clients)
5. All subsequent gRPC requests use Bearer token (OAuth2 access token or JWT)

### OAuth2 Integration

The system runs an OAuth 2.1 authorization server for service authentication:

- **Discovery**: `GET /.well-known/oauth-authorization-server`
- **Authorization Endpoint**: `GET|POST /authorize` — a GET renders the login form, a POST
  authenticates (either a session JWT in an `Authorization` header, or posted credentials)
- **Token Endpoint**: `POST /token`
- **Revocation Endpoint**: `POST /revoke`
- Clients authenticate via `Authorization` header in gRPC requests
- HTTP endpoints only support the OAuth2 flow (legacy HTTP auth routes should be deprecated)

See [`backend/docs/auth-flow.md`](../backend/docs/auth-flow.md) for what the server enforces —
byte-exact redirect URIs, mandatory S256 PKCE, refresh rotation with reuse detection.

**OAuth2 Implementation**: [`pkg/client/client.go:WithOAuth2Credentials`](pkg/client/client.go)

### Session Context

The session context contains:

- User information (ID, username, email, account status)
- Active account ID (initially the default account)
- Account permissions map (role for each account the user belongs to)
- Service-level permissions

**Domain Definition**: [`internal/authentication/sessions/session_context.go`](internal/authentication/sessions/session_context.go)

### Two-Factor Authentication (2FA)

- All users are issued a TOTP secret during registration
- Users must verify their TOTP secret by submitting a valid TOTP code with their current password
- Once verified, TOTP is required for all login attempts
- Passwords are hashed using scrypt before storage

### Admin-Only Login

The system supports a special admin-only login mode that has stricter requirements:

1. **Service Role Restriction**: Only users with `service_role = 'service_admin'` can use admin login
2. **2FA Requirement**: Admin login **requires** a valid TOTP token (6-digit code)
3. **Verified 2FA**: The user must have a verified 2FA secret (`two_factor_secret_verified_at IS NOT NULL`)
4. **Database Query**: Uses `GetAdminUserByUsername` instead of `GetUserByUsername` which includes additional filters

**Key Differences from Regular Login**:

- Regular users can have unverified 2FA secrets and login without TOTP
- Admin login **always** requires TOTP validation
- Admin login only works for users with service admin privileges
- Admin login uses a separate database query with stricter filtering

**Implementation**: [`internal/authentication/manager.go:ProcessLogin`](internal/authentication/manager.go) and [`internal/services/auth/handlers/authentication/authentication_http_routes.go:BuildLoginHandler`](internal/services/auth/handlers/authentication/authentication_http_routes.go)

### Logout

There is currently no server-side logout mechanism. Tokens expire naturally, and logout is handled client-side by discarding stored OAuth2 credentials.

## Account Roles and Permissions

Five roles, declared once in [`internal/authorization/platform.go`](../backend/internal/authorization/platform.go)
as `PlatformPolicy()`. That declaration is the only one: the migrator seeds it into
platform-go's `authorization/database` tables, and every permission check at runtime resolves
against what it seeded. There is no second list in SQL to keep in step — there used to be, and
the two had drifted on three of five roles.

Roles inherit, and inheritance is where most of a role's authority comes from:

```text
account_member       —
account_admin        inherits account_member
service_data_admin   —
service_admin        inherits account_admin, service_data_admin
service_user         —
```

### Account-level roles

- **Account Member** — the authority an ordinary user has, held **per account**. Meal
  planning, webhooks they can read, their own data privacy requests. 131 permissions.
- **Account Admin** — account settings, invitations, membership changes, ownership transfer,
  plus everything a member holds. 169 permissions.

### Service-level roles

- **Service User** — assigned to every user at signup, service-wide, and holds **nothing**.
  That is deliberate rather than an omission: a permission granted here would be granted in
  every account, and account authority is what `account_member` carries per account.
- **Service Data Admin** — the reference-data catalog: instruments, ingredients,
  preparations, measurement units and their bridges. 42 permissions.
- **Service Admin** — user administration, impersonation, session management, arbitrary queue
  messages and worker runs, plus everything an account admin and a data admin hold. 239
  permissions, which is every permission the service declares.

### Which roles a principal holds

Assignments live in `user_role_assignments`, which is this repository's table rather than the
platform's, because an assignment names a user and an account and no platform package can
model those without owning them. A row with a NULL `account_id` is a service-wide assignment;
anything else is scoped to that account. The row names its role by name, and a foreign key
onto the roles table refuses a name nothing declares.

Which roles may be assigned where is enforced in Go: `ModifyUserPermissionsInput` accepts only
the two account roles, because the role it names is written into an account-scoped assignment
and resolved within that account.

## Account Creation and User Registration

### Standard Registration

When a user registers without an invitation:

1. User account is created
2. A default account is automatically created for the user
3. User is made an admin of their default account
4. This account becomes their default account

### Registration with Invitation

When a user registers with an invitation token:

1. User account is created
2. User is added to the invited account as a member
3. The invitation is marked as accepted
4. A default account is still created for the user
5. The invited account becomes their default account

## Account Invitations

### Invitation Types

1. **Email-based invitations**: Sent to a specific email address
2. **Token-based invitations**: Created with a token that can be used during registration

### Invitation Process

1. Account admin creates an invitation
2. Invitation can be sent via email or shared as a token
3. When a user registers with the invitation token, they're automatically added to the account
4. If the invitation was sent to an email, it's automatically associated when that email registers

**Domain Definition**: [`internal/domain/identity/account_invitation.go`](internal/domain/identity/account_invitation.go)

## Account Switching

Users can switch between accounts they're members of:

1. User requests to switch to a different account via the `SetDefaultAccount` gRPC method
2. System validates the user is a member of that account
3. The account is permanently set as the user's default account
4. All subsequent requests use the new account context

**TODO**: The current implementation permanently changes the default account. Consider implementing session-based account switching that doesn't permanently change the user's default account.

## Account Membership Management

### Removing Users from Accounts

Users can be removed from accounts by account admins. When a user is removed:

1. Their membership is archived
2. If they have no remaining accounts, a new default account is created
3. If they have remaining accounts, one is set as their new default

**Important**: Users cannot remove themselves from accounts - this must be done by an account admin.

## System Architecture

```mermaid
graph TB
    User[User] --> Auth[Authentication]
    Auth --> Session[Session Context]
    Session --> ActiveAccount[Active Account]
    
    User --> Memberships[Account Memberships]
    Memberships --> Account1[Account 1]
    Memberships --> Account2[Account 2]
    Memberships --> AccountN[Account N]
    
    Account1 --> Data1[Account Data]
    Account2 --> Data2[Account Data]
    AccountN --> DataN[Account Data]
    
    User --> DefaultAccount[Default Account]
    DefaultAccount --> DefaultData[Default Account Data]
    
    Account1 --> Invitations[Account Invitations]
    Invitations --> NewUser[New User Registration]
    NewUser --> NewMembership[New Membership]
```

## Key Data Flow

```mermaid
sequenceDiagram
    participant U as User
    participant A as Auth Service
    participant I as Identity Service
    participant D as Database
    
    U->>A: Login (username, password)
    A->>D: Validate credentials
    D-->>A: User data
    A->>D: Get user's default account
    D-->>A: Default account ID
    A->>A: Create session context
    A-->>U: Session with user ID + account ID
    
    U->>I: Request data
    I->>I: Check session context
    I->>D: Query data for active account
    D-->>I: Account-specific data
    I-->>U: Response
```

⚠️ **Important Clarifications**:

- When a user registers with an invitation, they're added as a **member** (not admin) of the invited account
- The invited account becomes their default account, but they still get their own personal account created
- Account switching **permanently** changes the user's default account (not just the active session account)
- The system uses token-based authentication with OAuth2, not traditional server-side sessions
- All users have 2FA secrets but must verify them before 2FA becomes required

## Security Considerations

1. **Token-Based Authentication**: Uses JWT/PASETO tokens with OAuth2 for service authentication
2. **Password Security**: Passwords are hashed using scrypt before storage
3. **Two-Factor Authentication**: TOTP-based 2FA is available and can be required for login
4. **Permission Checking**: Every request validates the user has access to the active account
5. **Account Isolation**: Data is strictly isolated by account membership
6. **Self-Removal Prevention**: Users cannot remove themselves from accounts to prevent lockout
7. **Default Account Guarantee**: Users always have at least one account (their personal account)

## Known Issues and TODOs

### Critical Issues

- **TODO**: If a user's default account is deleted, the system likely breaks. Need to implement proper handling for this scenario.

### Future Improvements

- **TODO**: Implement session-based account switching that doesn't permanently change the user's default account

### gRPC Services

- **Auth Service**: [`internal/services/auth/grpc/`](internal/services/auth/grpc/) - Authentication and authorization
- **Identity Service**: [`internal/services/identity/grpc/`](internal/services/identity/grpc/) - User and account management

## Related Files

- **Domain Models**: [`internal/domain/identity/`](internal/domain/identity/)
- **Authentication**: [`internal/services/auth/`](internal/services/auth/)
- **Authorization**: [`internal/authorization/`](internal/authorization/)
- **Session Management**: [`internal/authentication/sessions/`](internal/authentication/sessions/)
