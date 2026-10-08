package mcpserver

import (
	"context"
	"errors"
	"html/template"
	"net/http"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	"github.com/primandproper/platform-go/v15/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/authentication/totp"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/modelcontextprotocol/go-sdk/auth"
)

// claimAccountID is the Subject claim carrying the account a token acts on behalf of.
//
// It is the one thing beside the user ID that every tool handler needs, and it is
// resolved once at /authorize rather than per request. Subject.Claims is
// map[string]string by construction, so it round-trips through the store as the
// same Go type it went in as.
const claimAccountID = mcptools.ClaimAccountID

// accessDeniedMessage is what a failed sign-in says, whichever half was wrong.
//
// Deliberately uninformative: "no such user" and "wrong password" as separate
// answers make this form an account enumeration oracle, and the endpoint is
// public. It also does not distinguish "not an admin" from "does not exist",
// because who holds an admin account is not something an anonymous caller
// should be able to probe for.
const accessDeniedMessage = "Access denied. Admin credentials required."

// AdminAuthenticator proves an operator's credentials — signin.Service.AdminAuthenticate.
type AdminAuthenticator interface {
	AdminAuthenticate(ctx context.Context, scope tenancy.Scope, credentials *signin.Credentials) (*platformidentity.Principal, error)
}

var _ AdminAuthenticator = (*signin.Service)(nil)

// subjectAuthenticator identifies the human at /authorize.
//
// It is the seam the platform's authorization server deliberately does not fill, and it is one
// call: signin's administrative door, which is what decides who may sign in here. The protocol
// around it — PKCE, redirect URI matching, code redemption, token rotation — is the platform's.
//
// The door is what makes this login safe to put on the public internet, and it is why there is
// nothing else here. It reads the user, hashes the password — a decoy one for a handle that names
// nobody — and only then checks the account's standing, that it holds an administrative role, and
// that it holds a proven second factor whatever its own policy says. Every refusal before the
// password is proven costs what a wrong password does, so the time an answer takes says neither
// whether a username exists nor whether it is an administrator's. The check this replaced read
// the role before it hashed anything, and an unknown username before that.
type subjectAuthenticator struct {
	signIn AdminAuthenticator
}

var _ oauth2server.SubjectAuthenticator = (*subjectAuthenticator)(nil)

// AuthenticateSubject implements oauth2server.SubjectAuthenticator.
//
// Every refusal is a *oauth2server.LoginError, which re-renders the form with a message rather
// than failing the authorization request: the human is still there and can try again. A refusal
// is whatever signin's own mapper names, which is every sentinel its doors refuse a credential
// with. Anything else — a directory that is down — is returned as itself, because retrying a form
// against a database that is down produces a user who tries four times and then files a support
// ticket.
func (a *subjectAuthenticator) AuthenticateSubject(ctx context.Context, req *http.Request) (*oauth2server.Subject, error) {
	principal, err := a.signIn.AdminAuthenticate(ctx, tenancy.Global(), &signin.Credentials{
		Username: req.FormValue(oauth2server.FieldUsername),
		Password: req.FormValue(oauth2server.FieldPassword),
		TOTPCode: req.FormValue(oauth2server.FieldTOTPCode),
	})
	if err != nil {
		if _, refused := signin.GRPCMapper.Map(err); !refused {
			return nil, err
		}

		// Asked for only once the password is proven and the account is an administrator's,
		// so saying so tells nobody anything they did not already know.
		if errors.Is(err, signin.ErrSecondFactorRequired) {
			return nil, oauth2server.NewLoginError("TOTP code is required.", err)
		}

		return nil, oauth2server.NewLoginError(accessDeniedMessage, err)
	}

	return &oauth2server.Subject{
		ID:     principal.User.ID,
		Claims: map[string]string{claimAccountID: principal.ActiveAccountID},
	}, nil
}

// refusingIssuer is the token issuer of a sign-in service that never mints a token.
//
// signin.Service requires one, and the MCP server's only use of the service is AdminAuthenticate,
// which proves a credential and issues nothing: the authorization server mints this server's
// tokens. An issuer that refused rather than one that worked is what makes a future call to a
// minting door here fail loudly instead of handing out a token nothing checks.
type refusingIssuer struct{}

// errNoTokensHere is every refusingIssuer answer.
var errNoTokensHere = errors.New("the MCP server issues no sign-in tokens")

func (refusingIssuer) IssueToken(context.Context, string, time.Duration, map[string]any) (tokenStr, jti string, err error) {
	return "", "", errNoTokensHere
}

// NewAdminSignIn builds the sign-in service the MCP login form proves an operator through.
//
// Only service_admin is admitted, as only service_admin always was here: a data administrator
// administers reference data from the admin app, and has no business holding an agent's token
// that can reach everything the MCP tools can.
func NewAdminSignIn(
	client database.Client,
	directory platformidentity.Store,
	authenticator authentication.Authenticator,
	totpVerifier totp.Verifier,
	logger logging.Logger,
	tracerProvider tracing.Provider,
) (*signin.Service, error) {
	return signin.NewService(client, directory, authenticator, refusingIssuer{},
		[]string{authorization.AccountAdminRoleName},
		signin.WithAdminServiceRoles(authorization.ServiceAdminRoleName),
		signin.WithSecondFactorPolicy(signin.SecondFactorWhenEnrolled),
		signin.WithTOTPVerifier(totpVerifier),
		signin.WithLogger(logger),
		signin.WithTracerProvider(tracerProvider),
	)
}

// loginTemplateData is what the login form renders from: the platform's view of
// the authorization request, plus the one thing branding owns.
type loginTemplateData struct {
	_ struct{} `json:"-"`

	CompanyName string

	// TOTPCodeField is the field the code is posted in, which is the one the platform names —
	// oauth2server.FieldTOTPCode — rather than a spelling of this form's own.
	TOTPCodeField string

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

            <label for="{{.TOTPCodeField}}">TOTP Code</label>
            <input type="text" id="{{.TOTPCodeField}}" name="{{.TOTPCodeField}}" autocomplete="one-time-code" inputmode="numeric" pattern="[0-9]*">

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
			LoginView:     view,
			CompanyName:   branding.CompanyName,
			TOTPCodeField: oauth2server.FieldTOTPCode,
		}); err != nil {
			logging.EnsureLogger(logger).WithValue("client_name", view.ClientName).Error("rendering MCP login form", err)
		}
	})
}

// newTokenVerifier adapts the authorization server's resource-server half to the MCP SDK's bearer
// middleware.
//
// oauth2server.Verifier makes both checks: the lookup, which is the whole point of an opaque
// access token — a revoked token stops working on the next request rather than at the end of its
// lifetime — and the audience, which RFC 8707 leaves to the resource server. It refuses a token
// with no audience as well as one naming somewhere else, where the check this replaced let the
// first through: a token minted with no resource indicator is spendable at every resource server
// sharing this store, the API among them. A client asking for a token for this server names it
// in the resource parameter, as the MCP specification requires it to.
//
// Every refusal is auth.ErrInvalidToken, which the SDK answers 401 with the resource metadata's
// challenge, so a client whose token stopped working is sent to get another.
func newTokenVerifier(verifier *oauth2server.Verifier) auth.TokenVerifier {
	return func(ctx context.Context, token string, _ *http.Request) (*auth.TokenInfo, error) {
		accessToken, err := verifier.Verify(ctx, token)
		if err != nil {
			return nil, errors.Join(auth.ErrInvalidToken, err)
		}

		return &auth.TokenInfo{
			UserID:     accessToken.Subject.ID,
			Scopes:     accessToken.Scopes,
			Expiration: accessToken.ExpiresAt,
			Extra:      map[string]any{claimAccountID: accessToken.Subject.Claims[claimAccountID]},
		}, nil
	}
}
