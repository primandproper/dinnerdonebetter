package authentication

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v12/authentication/oauth2server"
	oauth2servercfg "github.com/primandproper/platform-go/v12/authentication/oauth2server/config"
	"github.com/primandproper/platform-go/v12/database"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/metrics"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

// ProvideOAuth2Server builds the API server's OAuth 2.1 authorization server.
//
// The store is the platform's, with its client half redirected at oauth2_clients — see
// clientRegistryStore for why that split rather than one table or the other.
func ProvideOAuth2Server(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	cfg *oauth2servercfg.Config,
	dbClient database.Client,
	authenticator oauth2server.SubjectAuthenticator,
	clients oauth.OAuth2ClientDataManager,
) (*oauth2server.Server, error) {
	store, err := oauth2servercfg.NewStore(ctx, cfg, dbClient,
		oauth2servercfg.WithLogger(logger),
		oauth2servercfg.WithTracerProvider(tracerProvider),
		oauth2servercfg.WithMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("building oauth2 store: %w", err)
	}

	srv, err := oauth2server.NewServer(cfg.Issuer, &clientRegistryStore{Store: store, clients: clients}, authenticator,
		oauth2server.WithLogger(logger),
		oauth2server.WithTracerProvider(tracerProvider),
		oauth2server.WithMetricsProvider(metricsProvider),
		oauth2server.WithAccessTokenTTL(cfg.AccessTokenTTL),
		oauth2server.WithRefreshTokenTTL(cfg.RefreshTokenTTL),
		oauth2server.WithAuthorizationCodeTTL(cfg.AuthorizationCodeTTL),
		oauth2server.WithRefreshReuseDetection(!cfg.DisableRefreshReuseDetection),
		oauth2server.WithResources(cfg.Resources...),
		oauth2server.WithScopes(cfg.Scopes...),
		oauth2server.WithServiceDocumentation(cfg.ServiceDocumentation),
		oauth2server.WithLoginRenderer(newLoginRenderer(logger)),
	)
	if err != nil {
		return nil, fmt.Errorf("building oauth2 authorization server: %w", err)
	}

	return srv, nil
}

// loginTemplateData is what the login form renders from: the platform's view of the
// authorization request, plus the one thing branding owns.
type loginTemplateData struct {
	_ struct{} `json:"-"`

	CompanyName string

	oauth2server.LoginView
}

// loginTemplate draws the sign-in form a browser arriving without a session sees.
//
// html/template rather than concatenation, because ClientName is a registered client's name
// and a renderer that builds this by hand is choosing to be an XSS.
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
func newLoginRenderer(logger logging.Logger) oauth2server.LoginRenderer {
	return oauth2server.LoginRendererFunc(func(_ context.Context, res http.ResponseWriter, view oauth2server.LoginView) {
		res.Header().Set("Content-Type", "text/html; charset=utf-8")

		if view.Error != "" {
			res.WriteHeader(http.StatusUnauthorized)
		}

		if err := loginTemplate.Execute(res, &loginTemplateData{
			LoginView:   view,
			CompanyName: branding.CompanyName,
		}); err != nil {
			logging.EnsureLogger(logger).WithValue("client_name", view.ClientName).Error("rendering login form", err)
		}
	})
}
