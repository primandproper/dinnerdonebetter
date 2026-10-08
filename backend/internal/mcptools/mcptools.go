/*
Package mcptools is what a Model Context Protocol tool surface here is built on: the Toolset a
domain hands the server, and the Gate every tool call passes before it reaches a manager or a
store.

It is a leaf on purpose. The MCP server mounts whatever Toolsets it is handed and names no
domain; a domain's tools name this package and never the server. Both can hold to that only if
what they share lives in neither, which is what this is.

# The gate is the whole of the authorization

A gRPC method here has an interceptor in front of it that resolves the caller and an enforcer
that checks the method's grant. An MCP tool has nothing in front of it but the bearer check on
the transport, which says the token is good and not who holds it or what they may read. The
Gate is therefore where a tool call is turned into a caller — the Authenticator renders the
verified token as this application's session — and where the tool's grant is checked, against
the same sessions.GrantsFromContext the API's enforcer reads. A tool that needs a grant names
the permission its gRPC counterpart declares, so a caller refused over one transport is refused
over the other.

platform-go's internal/mcptool is the same gate with reflected schemas and a per-call operation
on top. It is unexported at v15.1.0 (platform-go#1162 exports it, unreleased); when it ships,
Gate is what it replaces.
*/
package mcptools

import (
	"context"

	"github.com/primandproper/platform-go/v15/callers"
	"github.com/primandproper/primitives-go/v2/authorization"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ClaimAccountID is the access token claim the caller's account travels in. It is written by
// the authorization server's subject authenticator when a token is issued, from the account
// the operator was resolved into at /authorize, and read back by the Authenticator on every
// tool call.
const ClaimAccountID = "account_id"

// Toolset is one surface's tools, registered onto the server in one call. platform-go's own
// surfaces (webhooks/mcp, waitlists/mcp, issuereports/mcp) have the same method, so one of
// those is a Toolset too.
type Toolset interface {
	RegisterOn(server *mcp.Server)
}

// Authenticator turns one tool call into the context the principal and grants extractors read.
//
// It is asked per call, with the call's request, because the context the MCP SDK hands a tool
// handler is not the request's: over streamable HTTP it is the session's, detached from
// whichever request opened the session. The verified token arrives on each call as
// req.Extra.TokenInfo — primitives' oauth2server/mcp.Protect puts it there — so that is what an
// Authenticator reads.
//
// A call carrying no credential, or one naming somebody the directory no longer admits, is
// answered by returning ctx unchanged: the extractor then finds nobody, and the call is refused
// as unauthenticated. An error is a failure to decide — a directory that would not answer — and
// is answered as a failed call rather than as a refusal.
//
// It is the same type as platform's mcptool.Authenticator, which the three platform tool
// surfaces take.
type Authenticator = func(ctx context.Context, req *mcp.CallToolRequest) (context.Context, error)

var (
	// ErrNilAuthenticator is a gate built with no way to read a call's credential.
	ErrNilAuthenticator = platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil authenticator for the MCP tool gate")

	// ErrNilPrincipalExtractor is a gate built with no way to tell who is calling.
	ErrNilPrincipalExtractor = platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil principal extractor for the MCP tool gate")

	// ErrNilGrantsExtractor is a gate built with no way to tell what the caller may do. There is
	// no default: one granting nothing refuses every call and reads as a broken deployment, and
	// one granting everything hands every tool to anybody holding a token.
	ErrNilGrantsExtractor = platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil grants extractor for the MCP tool gate")

	// ErrNoPrincipal is a tool call that arrived with no caller on its context.
	ErrNoPrincipal = platformerrors.New("no principal on the MCP tool call context")

	// ErrPermissionDenied is a tool call by a caller who does not hold the permission the tool's
	// gRPC counterpart requires.
	ErrPermissionDenied = platformerrors.New("the caller does not hold the permission this tool requires")
)

// Gate is who is calling and what they may do, asked once per tool call.
type Gate struct {
	authenticate Authenticator
	principals   callers.PrincipalExtractor
	grants       authorization.GrantsExtractor
}

// NewGate builds a gate over an authenticator and the two extractors that read what it attached.
func NewGate(authenticate Authenticator, principals callers.PrincipalExtractor, grants authorization.GrantsExtractor) (*Gate, error) {
	if authenticate == nil {
		return nil, ErrNilAuthenticator
	}

	if principals == nil {
		return nil, ErrNilPrincipalExtractor
	}

	if grants == nil {
		return nil, ErrNilGrantsExtractor
	}

	return &Gate{authenticate: authenticate, principals: principals, grants: grants}, nil
}

// Authenticate is the gate's authenticator, for a platform tool surface built over the same
// token reading as the domain's tools.
func (g *Gate) Authenticate() Authenticator { return g.authenticate }

// Grants is the extractor the gate checks permissions with, for a platform tool surface that
// reads the same grants.
func (g *Gate) Grants() authorization.GrantsExtractor { return g.grants }

// Begin authenticates the call, resolves the caller and checks that they hold every one of
// required, in the one place every tool starts.
//
// required is the permission list the tool's gRPC method declares; a tool whose counterpart is
// public passes none, and that is a decision the tool makes by name rather than an absence. The
// context returned carries the caller, and is the one a tool reads through from here on.
func (g *Gate) Begin(ctx context.Context, req *mcp.CallToolRequest, required ...authorization.Permission) (context.Context, callers.Principal, error) {
	ctx, err := g.authenticate(ctx, req)
	if err != nil {
		return ctx, nil, platformerrors.Wrap(err, "authenticating an MCP tool call")
	}

	principal, ok := g.principals(ctx)
	if !ok || principal == nil {
		return ctx, nil, ErrNoPrincipal
	}

	if len(required) > 0 {
		grants, found := g.grants(ctx)
		if !found || !grants.HasAll(required...) {
			return ctx, nil, ErrPermissionDenied
		}
	}

	return ctx, principal, nil
}
