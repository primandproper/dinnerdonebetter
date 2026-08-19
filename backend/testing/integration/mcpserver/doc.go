/*
Package integration starts the MCP server the way the binary does — its rendered config
file, its dependency injection container, its migrated schema — and drives real HTTP at
it: RFC 7591 dynamic registration, RFC 8414 discovery, an authorization code with PKCE,
a token exchange, and a tool call.

Two tests in the tree already cover the halves of this that do not need a process.
internal/mcpserver.TestBuildRouter drives the same protocol through the real handler
chain over an in-process router and the memory store, and
internal/repositories/postgres/migrations.TestOAuth2Store_Conformance runs platform's
conformance suite against the tables our migrator creates. Between them the grant logic
and the schema are covered.

What is left, and what this package is for, is the wiring between a built server and its
environment: that the oauth2 stanza in the rendered config names the table prefix
migration 33 used, that the container resolves the identity and authentication
dependencies the login form needs alongside a database client, that MCP_BASE_URL becomes
the issuer and the resource the discovery documents advertise and the audience the token
verifier enforces — and, the reason the durable authorization server was adopted at all,
that a code issued by one replica is redeemable at another and a token outlives the
process that minted it.

Every replica in a run shares one base URL, because that is what a fleet behind a load
balancer looks like: the issuer and the resource identifier are properties of the
deployment, not of a process, and only the port each replica binds differs.
*/
package integration
