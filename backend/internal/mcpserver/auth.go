package mcpserver

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"slices"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v11/authentication/oauth2server"
	"github.com/primandproper/platform-go/v11/authentication/totp"
	"github.com/primandproper/platform-go/v11/observability/logging"

	"github.com/modelcontextprotocol/go-sdk/auth"
)

// claimAccountID is the Subject claim carrying the account a token acts on behalf of.
//
// It is the one thing beside the user ID that every tool handler needs, and it is
// resolved once at /authorize rather than per request. Subject.Claims is
// map[string]string by construction, so it round-trips through the store as the
// same Go type it went in as.
const claimAccountID = "account_id"

// accessDeniedMessage is what a failed sign-in says, whichever half was wrong.
//
// Deliberately uninformative: "no such user" and "wrong password" as separate
// answers make this form an account enumeration oracle, and the endpoint is
// public. It also does not distinguish "not an admin" from "does not exist",
// because who holds an admin account is not something an anonymous caller
// should be able to probe for.
const accessDeniedMessage = "Access denied. Admin credentials required."

// subjectAuthenticator identifies the human at /authorize.
//
// This is the seam the platform's authorization server deliberately does not
// fill, and everything application-shaped about MCP authentication lives here:
// the account must be an admin, it must not be banned, the password is argon2,
// and TOTP is checked when the user has verified a secret. The protocol around
// it — PKCE, redirect URI matching, code redemption, token rotation — is the
// platform's.
type subjectAuthenticator struct {
	identityRepo  identity.Repository
	authenticator authentication.Authenticator
	totpVerifier  totp.Verifier
}

var _ oauth2server.SubjectAuthenticator = (*subjectAuthenticator)(nil)

// AuthenticateSubject implements oauth2server.SubjectAuthenticator.
//
// Every refusal is a *oauth2server.LoginError, which re-renders the form with a
// message rather than failing the authorization request: the human is still
// there and can try again. A broken identity store returns a plain error
// instead, because retrying a form against a database that is down produces a
// user who tries four times and then files a support ticket.
func (a *subjectAuthenticator) AuthenticateSubject(ctx context.Context, req *http.Request) (*oauth2server.Subject, error) {
	username := req.FormValue("username")
	password := req.FormValue("password")
	totpToken := req.FormValue("totp_token")

	// Admin-only, and the lookup is what enforces it: there is no non-admin
	// branch to fall through to.
	user, err := a.identityRepo.GetAdminUserByUsername(ctx, username)
	if err != nil || user == nil {
		return nil, oauth2server.NewLoginError(accessDeniedMessage, err)
	}

	if user.IsBanned() {
		return nil, oauth2server.NewLoginError("Access denied. Account is banned.", nil)
	}

	matches, err := a.authenticator.PasswordMatches(ctx, user.HashedPassword, password)
	if err != nil || !matches {
		return nil, oauth2server.NewLoginError(accessDeniedMessage, err)
	}

	if user.TwoFactorSecretVerifiedAt != nil {
		if verifyErr := a.totpVerifier.Verify(ctx, user.TwoFactorSecret, totpToken); verifyErr != nil {
			if errors.Is(verifyErr, totp.ErrCodeRequired) {
				return nil, oauth2server.NewLoginError("TOTP code is required.", verifyErr)
			}

			return nil, oauth2server.NewLoginError(accessDeniedMessage, verifyErr)
		}
	}

	// Not a LoginError. The credentials were right and the account still has no
	// resolvable default account, which is a broken record rather than a wrong
	// password — re-rendering the form would ask the human to fix it by typing.
	accountID, err := a.identityRepo.GetDefaultAccountIDForUser(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return &oauth2server.Subject{
		ID:     user.ID,
		Claims: map[string]string{claimAccountID: accountID},
	}, nil
}

// loginTemplateData is what the login form renders from: the platform's view of
// the authorization request, plus the one thing branding owns.
type loginTemplateData struct {
	_ struct{} `json:"-"`

	CompanyName string

	oauth2server.LoginView
}

// loginTemplate draws the sign-in form.
//
// html/template rather than concatenation, because ClientName is whatever an
// anonymous /register call said it was — a renderer that builds this by hand is
// choosing to be an XSS.
var loginTemplate = template.Must(template.New("login").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.CompanyName}} — Sign In</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f5f5f5; display: flex; justify-content: center; align-items: center; min-height: 100vh; }
        .card { background: white; border-radius: 12px; padding: 2rem; width: 100%; max-width: 400px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        h1 { font-size: 1.5rem; margin-bottom: 0.5rem; text-align: center; }
        .client { font-size: 0.875rem; color: #666; text-align: center; margin-bottom: 1.5rem; }
        label { display: block; font-size: 0.875rem; font-weight: 500; margin-bottom: 0.25rem; color: #333; }
        input[type="text"], input[type="password"] { width: 100%; padding: 0.625rem; border: 1px solid #ddd; border-radius: 6px; font-size: 1rem; margin-bottom: 1rem; }
        input:focus { outline: none; border-color: #4a90d9; box-shadow: 0 0 0 2px rgba(74,144,217,0.2); }
        button { width: 100%; padding: 0.75rem; background: #4a90d9; color: white; border: none; border-radius: 6px; font-size: 1rem; font-weight: 500; cursor: pointer; }
        button:hover { background: #3a7bc8; }
        .error { background: #fee; color: #c33; padding: 0.75rem; border-radius: 6px; margin-bottom: 1rem; font-size: 0.875rem; }
        .scopes { font-size: 0.8125rem; color: #666; margin-bottom: 1rem; }
    </style>
</head>
<body>
    <div class="card">
        <h1>Sign In</h1>
        {{with .ClientName}}<p class="client">to continue to {{.}}</p>{{end}}
        {{with .Error}}<div class="error">{{.}}</div>{{end}}
        {{with .Scopes}}<p class="scopes">Requested access: {{range $i, $s := .}}{{if $i}}, {{end}}{{$s}}{{end}}</p>{{end}}
        <form method="POST" action="{{.Action}}">
            <label for="username">Username</label>
            <input type="text" id="username" name="username" required autofocus>

            <label for="password">Password</label>
            <input type="password" id="password" name="password" required>

            <label for="totp_token">TOTP Code</label>
            <input type="text" id="totp_token" name="totp_token" autocomplete="one-time-code" inputmode="numeric" pattern="[0-9]*">

            <button type="submit">Sign In</button>
        </form>
    </div>
</body>
</html>`))

// newLoginRenderer draws the platform's login form in this application's brand.
//
// The authorization parameters are not hidden form fields here: Action is the
// /authorize URL with the original query string intact, so the POST is
// validated against exactly the same request the GET was. That is what makes
// carrying them in the form unnecessary rather than merely redundant.
func newLoginRenderer(logger logging.Logger) oauth2server.LoginRenderer {
	return oauth2server.LoginRendererFunc(func(ctx context.Context, res http.ResponseWriter, view oauth2server.LoginView) {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")

		// A renderer owns the status, and a re-render with a message is the
		// answer to a refused credential.
		if view.Error != "" {
			res.WriteHeader(http.StatusUnauthorized)
		}

		if err := loginTemplate.Execute(res, &loginTemplateData{
			LoginView:   view,
			CompanyName: branding.CompanyName,
		}); err != nil {
			logging.EnsureLogger(logger).WithValue("client_name", view.ClientName).Error("rendering MCP login form", err)
		}
	})
}

// newTokenVerifier adapts the authorization server to the MCP SDK's bearer
// middleware.
//
// Two things happen here that the SDK cannot do for us. The lookup is the whole
// point of an opaque access token — a revoked token stops working on the next
// request rather than at the end of its lifetime — and the audience check is the
// one RFC 8707 leaves to the resource server: a token minted for somewhere else
// must not be spendable here, and no authorization server can make that
// determination on our behalf.
func newTokenVerifier(srv *oauth2server.Server, resource string) auth.TokenVerifier {
	return func(ctx context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		accessToken, err := srv.Authenticate(ctx, token)
		if err != nil {
			return nil, errors.Join(auth.ErrInvalidToken, err)
		}

		// An empty audience is a token no resource was named for, which this
		// server does not refuse: a client that sends no RFC 8707 resource
		// parameter gets one, and refusing those would make every such client
		// unable to sign in. What must not be accepted is an audience naming
		// somewhere that is not here — that token was minted for another
		// resource server and arriving at this one is the replay resource
		// indicators exist to stop.
		if len(accessToken.Audience) > 0 && !slices.Contains(accessToken.Audience, resource) {
			return nil, errors.Join(auth.ErrInvalidToken, errWrongAudience)
		}

		return &auth.TokenInfo{
			UserID:     accessToken.Subject.ID,
			Scopes:     accessToken.Scopes,
			Expiration: accessToken.ExpiresAt,
			Extra:      map[string]any{claimAccountID: accessToken.Subject.Claims[claimAccountID]},
		}, nil
	}
}

// errWrongAudience is the refusal the store cannot make. Expiry and revocation
// are already its answer — GetAccessToken returns ErrExpired rather than an
// inactive token — so this is the only condition left for the resource server to
// check for itself.
var errWrongAudience = errors.New("token audience does not name this resource server")
