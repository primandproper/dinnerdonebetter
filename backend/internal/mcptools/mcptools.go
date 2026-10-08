/*
Package mcptools is what a Model Context Protocol tool surface here is built on: the Toolset a
domain hands the server, the hand-written schema vocabulary its tools describe themselves with,
and the one read every tool makes of its caller.

It is a leaf on purpose. The MCP server mounts whatever Toolsets it is handed and names no
domain; a domain's tools name this package and never the server. Both can hold to that only if
what they share lives in neither, which is what this is.

The schema helpers are a hand-written vocabulary, and schema.go says what that exposes: every
type's schema is written out beside a struct that can move without it. platform-go's mcptool
package — unreleased at v15.1.0 — reflects a tool's schema off the type itself and is what
replaces this package when it ships; see platform-go#1162.
*/
package mcptools

import (
	"errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ClaimAccountID is the access token claim the caller's account travels in. It is written by
// the authorization server when a token is issued and read by every tool through
// AccountFromRequest.
const ClaimAccountID = "account_id"

// Toolset is one domain's tools, registered onto the server in one call. platform-go's own
// surfaces (webhooks/mcp, issuereports/mcp) have the same method, so one of those is a
// Toolset too.
type Toolset interface {
	RegisterOn(server *mcp.Server)
}

var (
	errNotAuthenticated = errors.New("not authenticated")
	errNoAccountOnToken = errors.New("no account on token")
)

// AccountFromRequest resolves the caller's account from the tool call's bearer token.
//
// The account is read off the token rather than looked up again: it was resolved once at
// /authorize and travels in the access token's claims, so a tool call costs the one store read
// the bearer middleware already made.
func AccountFromRequest(req *mcp.CallToolRequest) (accountID string, err error) {
	if req == nil || req.Extra == nil || req.Extra.TokenInfo == nil {
		return "", errNotAuthenticated
	}

	accountID, ok := req.Extra.TokenInfo.Extra[ClaimAccountID].(string)
	if !ok || accountID == "" {
		return "", errNoAccountOnToken
	}

	return accountID, nil
}
