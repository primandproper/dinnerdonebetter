package mcptools

import (
	"context"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	oauth2mcp "github.com/primandproper/primitives-go/v2/authentication/oauth2server/mcp"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// PrincipalDirectory is the directory read every authenticated tool call makes: platform's
// identity.Store.GetPrincipal, which is where standing is enforced. A user whose status admits
// no sign-in is refused there rather than answered, so a ban takes effect on the next call
// whatever minted the token it arrived with.
type PrincipalDirectory interface {
	GetPrincipal(
		ctx context.Context,
		q database.SQLQueryExecutor,
		scope tenancy.Scope,
		userID, activeAccountID string,
	) (*platformidentity.Principal, error)
}

// SessionRenderer turns a resolved principal into this application's session: what their roles
// grant. It is identitybuild.SessionBuilder, the same one the API's interceptor renders a
// request through.
type SessionRenderer interface {
	SessionForPrincipal(ctx context.Context, principal *platformidentity.Principal) (*sessions.ContextData, error)
}

var (
	// ErrNilDirectory is an authenticator built with nothing to resolve a token's subject in.
	ErrNilDirectory = platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil principal directory for the MCP authenticator")

	// ErrNilDatabaseClient is an authenticator built with no handle to read the directory on.
	ErrNilDatabaseClient = platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil database client for the MCP authenticator")

	// ErrNilSessionRenderer is an authenticator built with no way to say what a caller's roles
	// grant.
	ErrNilSessionRenderer = platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil session renderer for the MCP authenticator")
)

// NewAuthenticator builds the Authenticator every tool surface on this server reads its caller
// through.
//
// It is the access-token half of platform's sign-in extractor, for a transport that extractor
// ships no middleware for: the token the bearer middleware already verified names a subject and
// the account they were resolved into at /authorize, the directory answers who that is today,
// and the session renderer says what their roles grant. The directory is read on every call,
// which is what makes a ban, a membership that ended or an account that was archived take
// effect on the next call rather than at the end of the token's lifetime.
//
// The service roles are kept. The MCP login form is platform's administrative door —
// signin.Service.AdminAuthenticate, service_admin only, a proven second factor required — so a
// token minted here is an administrative token in the sense the API's interceptor means it,
// and the grants an operator holds are the grants the platform tool surfaces check. An
// ordinary-door token, which the API narrows to service_user, cannot be minted here at all.
func NewAuthenticator(directory PrincipalDirectory, client database.Client, renderer SessionRenderer) (Authenticator, error) {
	if directory == nil {
		return nil, ErrNilDirectory
	}

	if client == nil {
		return nil, ErrNilDatabaseClient
	}

	if renderer == nil {
		return nil, ErrNilSessionRenderer
	}

	return func(ctx context.Context, req *mcp.CallToolRequest) (context.Context, error) {
		if req == nil || req.Extra == nil {
			return ctx, nil
		}

		token, ok := oauth2mcp.AccessTokenFrom(req.Extra.TokenInfo)
		if !ok || token.Subject.ID == "" {
			return ctx, nil
		}

		// Every user is in the one global directory; the account is a grouping inside it.
		// See sessions.Principal.Scope.
		principal, err := directory.GetPrincipal(ctx, client.Reader(), tenancy.Global(), token.Subject.ID, token.Subject.Claims[ClaimAccountID])
		if err != nil {
			if refusesTheCaller(err) {
				return ctx, nil
			}

			return ctx, platformerrors.Wrap(err, "resolving an access token's caller")
		}

		if principal == nil || principal.User == nil {
			return ctx, nil
		}

		session, err := renderer.SessionForPrincipal(ctx, principal)
		if err != nil {
			return ctx, platformerrors.Wrap(err, "rendering an access token's session")
		}

		return sessions.AttachToContext(ctx, session), nil
	}, nil
}

// refusesTheCaller reports whether a directory error is an answer about the caller rather than
// a failure to give one: the same set platform's sign-in extractor reads as a refusal.
func refusesTheCaller(err error) bool {
	return errors.Is(err, platformidentity.ErrSignInNotAdmitted) ||
		errors.Is(err, platformidentity.ErrUserNotFound) ||
		errors.Is(err, platformidentity.ErrMembershipNotFound) ||
		errors.Is(err, platformerrors.ErrInvalidIDProvided) ||
		errors.Is(err, tenancy.ErrNoScope)
}
