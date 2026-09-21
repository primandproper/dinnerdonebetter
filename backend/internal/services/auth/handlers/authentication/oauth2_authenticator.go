package authentication

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/authentication/tokens"
	"github.com/primandproper/primitives-go/v2/authentication/totp"
	"github.com/primandproper/primitives-go/v2/database"
)

// ClaimAccountID is the Subject claim naming the account a token acts on behalf of.
//
// Resolved once, at /authorize, rather than on every request the token is spent on. The gRPC
// authn interceptor reads it back out to build session context data, which is what makes an
// access token here carry the same two identifiers a JWT does.
const ClaimAccountID = "account_id"

// loginFailedMessage is what a refused sign-in says, whichever half was wrong.
//
// Deliberately uninformative, for the reason the platform's own default is: "no such user" and
// "wrong password" as distinguishable answers make a public form an account enumeration
// oracle, and rate limiting slows that down rather than closing it.
const loginFailedMessage = "Sign-in failed. Check your details and try again."

// subjectAuthenticator identifies the resource owner behind an authorization request.
//
// It answers to two kinds of caller, and that is the one place this server deliberately
// differs from the MCP server's.
//
//   - A first-party client that has already signed the user in — the web app, the iOS app,
//     pkg/client — presents the session JWT it holds in an Authorization header. There is a
//     live authenticated session already; re-rendering a password form for it would be asking
//     the user to prove something they proved a moment ago, to a page they did not ask for.
//   - A browser arriving cold gets the login form, and posts username, password and TOTP back
//     to the same URL. This is the ordinary OAuth shape, and it is what a third-party client
//     using discovery would find.
//
// The platform calls this only on POST, so both paths post: the JWT one carries no form
// fields, the form one carries no bearer token. Which is present decides which is used, and
// neither can substitute for the other — a form post cannot skip the password by adding a
// header it does not have, and a JWT holder is not asked for one.
//
// Upstream ticket: this hybrid exists because oauth2server has no seam for "the resource owner
// is already authenticated, do not render a form". A LoginRenderer that could answer a
// pre-authenticated request by returning the Subject would let the JWT path be an
// authenticator rather than a fork inside one.
type subjectAuthenticator struct {
	directory     platformidentity.SignInReader
	db            database.Client
	authenticator authentication.Authenticator
	totpVerifier  totp.Verifier
	tokenIssuer   tokens.Issuer
}

var _ oauth2server.SubjectAuthenticator = (*subjectAuthenticator)(nil)

// AuthenticateSubject implements oauth2server.SubjectAuthenticator.
func (a *subjectAuthenticator) AuthenticateSubject(ctx context.Context, req *http.Request) (*oauth2server.Subject, error) {
	if bearer := bearerTokenFrom(req); bearer != "" {
		return a.subjectFromToken(ctx, bearer)
	}

	return a.subjectFromCredentials(ctx, req)
}

// subjectFromToken resolves the user behind a session JWT.
//
// A JWT that does not parse is a LoginError rather than a hard failure: an expired session is
// the ordinary reason to be here, and the answer to it is the login form — which is the same
// answer a human would get if they had arrived without one.
func (a *subjectAuthenticator) subjectFromToken(ctx context.Context, bearer string) (*oauth2server.Subject, error) {
	claims, err := a.tokenIssuer.ParseToken(ctx, bearer)
	if err != nil {
		return nil, oauth2server.NewLoginError(loginFailedMessage, err)
	}

	userID := claims.Subject()
	if userID == "" {
		return nil, oauth2server.NewLoginError(loginFailedMessage, nil)
	}

	// The JWT already names an account when it was minted for one — LoginForToken with a
	// DesiredAccountID does exactly that — and honoring it is what keeps an OAuth2 token
	// scoped to the same account the session it came from was.
	if accountID, _ := claims.GetString(ClaimAccountID); accountID != "" {
		return &oauth2server.Subject{ID: userID, Claims: map[string]string{ClaimAccountID: accountID}}, nil
	}

	return a.subjectForUser(ctx, userID)
}

// subjectFromCredentials resolves the user behind a posted login form.
func (a *subjectAuthenticator) subjectFromCredentials(ctx context.Context, req *http.Request) (*oauth2server.Subject, error) {
	username := req.FormValue("username")
	password := req.FormValue("password")
	totpToken := req.FormValue("totp_token")

	user, err := a.directory.GetUserByUsername(ctx, a.db.Reader(), ddbidentity.Scope(), username)
	if err != nil || user == nil {
		return nil, oauth2server.NewLoginError(loginFailedMessage, err)
	}

	// No status check here. It is made below, by the read that resolves the account this
	// token will act for: GetPrincipal refuses a user whose status does not admit signing
	// in, before it reads a membership. A second check here would be a second copy of the
	// rule, and the one that drifts is the one nobody looks at.

	matches, err := a.authenticator.PasswordMatches(ctx, user.HashedPassword, password)
	if err != nil || !matches {
		return nil, oauth2server.NewLoginError(loginFailedMessage, err)
	}

	if user.TwoFactorSecretVerifiedAt != nil {
		if verifyErr := a.totpVerifier.Verify(ctx, user.TwoFactorSecret, totpToken); verifyErr != nil {
			if errors.Is(verifyErr, totp.ErrCodeRequired) {
				return nil, oauth2server.NewLoginError("TOTP code is required.", verifyErr)
			}

			return nil, oauth2server.NewLoginError(loginFailedMessage, verifyErr)
		}
	}

	return a.subjectForUser(ctx, user.ID)
}

// subjectForUser attaches the account the token will act on behalf of.
//
// Not a LoginError: the credentials were right and the account has no resolvable default,
// which is a broken record rather than a wrong password. Re-rendering the form would be asking
// the human to fix it by typing.
//
// The read is the principal rather than a default-account lookup, because it is the same
// read every authenticated request makes and it enforces the account status on the way —
// so a banned user is refused here rather than issued a token that every subsequent
// request rejects.
func (a *subjectAuthenticator) subjectForUser(ctx context.Context, userID string) (*oauth2server.Subject, error) {
	principal, err := a.directory.GetPrincipal(ctx, a.db.Reader(), ddbidentity.Scope(), userID, "")
	if err != nil {
		return nil, err
	}

	return &oauth2server.Subject{
		ID:     userID,
		Claims: map[string]string{ClaimAccountID: principal.ActiveAccountID},
	}, nil
}

// bearerTokenFrom reads a bearer credential off a request, or returns empty.
func bearerTokenFrom(req *http.Request) string {
	header := req.Header.Get("Authorization")
	if header == "" {
		return ""
	}

	if token, found := strings.CutPrefix(header, "Bearer "); found {
		return strings.TrimSpace(token)
	}

	return ""
}
