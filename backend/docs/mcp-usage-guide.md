# Connecting to the Dinner Done Better MCP Server

## Prerequisites

- An **admin** account on the Dinner Done Better instance
- The backend API service running (HTTP on `:8000`, gRPC on `:8001` for local dev)

## Quick Start (Local Development)

### 1. Start the MCP server

```bash
make mcp
```

This runs the MCP server on port `8888` with hot-reload (via `air`), proxied on port `9999`.

### 2. Test with the MCP Inspector

```bash
make mcp_inspector
```

This launches the `@modelcontextprotocol/inspector` UI, which connects to `http://localhost:8888`. You can browse available tools and invoke them interactively.

## Connecting an MCP Client

### Claude Desktop / Claude Code

Add the server to your MCP client configuration. For a deployed server:

```json
{
  "mcpServers": {
    "dinner-done-better": {
      "type": "streamable-http",
      "url": "https://mcp.dinnerdonebetter.com/mcp"
    }
  }
}
```

For local development:

```json
{
  "mcpServers": {
    "dinner-done-better": {
      "type": "streamable-http",
      "url": "http://localhost:8888/mcp"
    }
  }
}
```

When the client connects, it will initiate the OAuth2 flow automatically.

## Authentication Flow

The MCP server is an OAuth 2.1 authorization server — `platform-go`'s
`authentication/oauth2server` — and a protected resource in front of the same process.
Compliant MCP clients handle the whole exchange transparently, but here's what happens:

Its records live in Postgres (`ddb_oauth2_clients`, `ddb_oauth2_authorization_codes`,
`ddb_oauth2_access_tokens`, `ddb_oauth2_refresh_tokens`), so a sign-in survives a restart
and a second replica: the authorization code one replica issues is redeemable at whichever
one serves `/token`.

### 1. Discovery

The client fetches server metadata:

- `GET /.well-known/oauth-protected-resource` (RFC 9728)
- `GET /.well-known/oauth-authorization-server` (RFC 8414)

### 2. Dynamic Client Registration

The client registers itself (RFC 7591):

```bash
POST /register
Content-Type: application/json

{
  "redirect_uris": ["http://localhost:..."],
  "client_name": "My MCP Client"
}
```

Returns a `client_id` and `client_secret`. Registration is open by construction — RFC 7591
requires that for discovery to work at all — and each one lapses after **90 days**, at which
point a client re-registers on its next discovery.

`redirect_uris` are matched **exactly**, byte for byte, at `/authorize` and again at `/token`
against the URI the code was issued for. Not a prefix, not "same host", not "ignoring the
port". A client registered with a loose URI gets an `invalid_request` rather than a code.

### 3. Authorization (Login)

The client opens a browser to:

```bash
GET /authorize?response_type=code&client_id=...&redirect_uri=...&code_challenge=...&code_challenge_method=S256&state=...
```

A login form is rendered. Enter your:

| Field         | Description                |
|---------------|----------------------------|
| **Username**  | Your admin username        |
| **Password**  | Your admin password        |
| **TOTP Code** | Your 2FA code (if enabled) |

On success, the server redirects back to the client with an authorization code.

**Only admin accounts can authenticate.** Regular user accounts will be rejected, with the
same message a wrong password gets — telling the two apart would make this form an account
enumeration oracle, and it is a public endpoint.

PKCE is mandatory and **S256 only**. There is no opt-out and no support for the `plain`
method, which puts the verifier in the request PKCE exists to protect.

### 4. Token Exchange

The client exchanges the authorization code for tokens:

```bash
POST /token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code&code=...&code_verifier=...&client_id=...&redirect_uri=...
```

Returns:

```json
{
  "access_token": "...",
  "token_type": "Bearer",
  "refresh_token": "...",
  "expires_in": 900
}
```

Access tokens are **opaque**, not signed: every authenticated request costs a lookup, and
what that buys is that a revoked token stops working on the next request rather than at the
end of its lifetime. Every credential is stored as a SHA-256 digest, never as itself.

### 5. Authenticated MCP Requests

All tool calls go to `POST /mcp` with the bearer token:

```bash
Authorization: Bearer <access_token>
```

### Revocation

```bash
POST /revoke
Content-Type: application/x-www-form-urlencoded

token=...
```

RFC 7009. The token stops working immediately, which is the property the opaque access token
above is paying for.

### Token Lifetimes

| Token                 | Lifetime   |
|-----------------------|------------|
| Authorization code    | 1 minute   |
| Access token          | 15 minutes |
| Refresh token         | 7 days     |
| Client registration   | 90 days    |

Refresh tokens are single-use and **rotate**: each refresh returns a new access/refresh pair
and retires the old refresh token. Presenting an already-redeemed one is treated as a replay
and revokes the whole token family, not just the token presented — so a stolen refresh token
is good for one exchange before the theft ends the session.

The 15-minute access token is deliberate, not a tightening for its own sake. With a durable
store and a rotating refresh token behind it the session already survives; a longer-lived
bearer buys nothing and costs the entire window in which a leaked one works. Your client
should refresh transparently.

These are configurable per environment — see `oauth2` in the rendered
`mcp_server_config.json`, or the `DINNER_DONE_BETTER_OAUTH2_*` environment variables.

## Transport Modes

The server supports three transport modes (set via `--transport` flag):

| Transport        | Description               | Use Case                         |
|------------------|---------------------------|----------------------------------|
| `http` (default) | Stateless streamable HTTP | Production, API gateways         |
| `sse`            | Server-Sent Events        | Long-lived streaming connections |
| `stdio`          | Standard I/O              | CLI tools, local piping          |

## Troubleshooting

**"Access denied. Admin credentials required."**
Only admin accounts can log in. Ensure you're using admin credentials.

**Token expired / "unknown token"**
Access tokens last 15 minutes. Your MCP client should refresh automatically using the refresh
token. If the refresh also fails, you'll need to sign in again — either the refresh token
aged out (7 days) or it was replayed, which revokes the family by design.

**`invalid_request` on `/authorize` after registering**
The `redirect_uri` has to match a registered one exactly. A trailing slash, a different port,
or an added query parameter is a different URI.

**Everything worked yesterday and now the client re-registers**
Registrations lapse after 90 days, and RFC 7591 registration is automatic, so this is the
system working. It also happens once for every client on the deploy that moved this server
off its in-memory store, since those registrations lived in a map.

**Connection refused on localhost:8888**
Make sure the MCP server is running (`make mcp`) and the backend API is running on ports 8000/8001.

**PKCE verification failed**
The client must use S256 code challenge method. This is handled automatically by compliant MCP clients.
