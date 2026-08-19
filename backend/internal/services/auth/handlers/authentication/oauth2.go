package authentication

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v11/authentication/tokens"
	"github.com/primandproper/platform-go/v11/observability"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
)

func ProvideOAuth2ClientManager(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	cfg *OAuth2Config,
	dataManager types.Repository,
) *manage.Manager {
	manager := manage.NewManager()

	tracer := tracing.NewNamedTracer(tracerProvider, "oauth2_client_manager")

	manager.SetValidateURIHandler(validateRedirectURI)

	manager.MapAuthorizeGenerate(generates.NewAuthorizeGenerate())
	manager.MapAccessGenerate(generates.NewAccessGenerate())
	manager.MapClientStorage(newOAuth2ClientStore(cfg.Domain, logger, tracer, dataManager))
	manager.MapTokenStorage(&oauth2TokenStoreImpl{
		tracer:      tracer,
		logger:      logging.EnsureLogger(logger),
		dataManager: dataManager,
	})

	return manager
}

func ProvideOAuth2ServerImplementation(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	tokenIssuer tokens.Issuer,
	manager *manage.Manager,
) *server.Server {
	oauth2ServerConfig := &server.Config{
		TokenType: "Bearer",
		AllowedResponseTypes: []oauth2.ResponseType{
			oauth2.Code,
		},
		AllowedGrantTypes: []oauth2.GrantType{
			oauth2.AuthorizationCode,
			oauth2.Refreshing,
		},
		AllowedCodeChallengeMethods: []oauth2.CodeChallengeMethod{
			oauth2.CodeChallengeS256,
			oauth2.CodeChallengePlain,
		},
	}

	oauth2Server := server.NewServer(oauth2ServerConfig, manager)

	tracer := tracing.NewNamedTracer(tracerProvider, "oauth2_server_impl")

	oauth2Server.AuthorizeScopeHandler = AuthorizeScopeHandler(logger)
	oauth2Server.AccessTokenExpHandler = AccessTokenExpHandler(logger)
	oauth2Server.ClientScopeHandler = ClientScopeHandler(logger)
	oauth2Server.UserAuthorizationHandler = buildUserAuthorizationHandler(tracer, logger, tokenIssuer)
	// No PasswordAuthorizationHandler: the password grant is absent from AllowedGrantTypes above,
	// so GetAccessToken rejects it — but ValidationTokenRequest runs the handler first, which meant
	// POST /oauth2/token was verifying usernames and passwords against the database for a grant it
	// could never issue a token for. The library's default returns access_denied without touching
	// anything. See TestAuth_OAuth2PasswordGrant.
	//
	// this allows GET requests to retrieve tokens
	oauth2Server.SetAllowGetAccessRequest(true)
	oauth2Server.ClientInfoHandler = buildClientInfoHandler()
	oauth2Server.InternalErrorHandler = buildInternalErrorHandler(logger)
	oauth2Server.ResponseErrorHandler = buildOAuth2ErrorHandler(logger)

	return oauth2Server
}

// validateRedirectURI rejects redirect URIs whose hostname is neither the client's
// registered domain nor a subdomain of it. Ports are deliberately ignored so localdev
// setups (registered domain and callback on different local ports) keep working.
// A plain suffix check would accept e.g. "evildinnerdonebetter.com", so the subdomain
// match requires a dot boundary.
func validateRedirectURI(baseURI, redirectURI string) error {
	base, err := url.Parse(baseURI)
	if err != nil {
		return err
	}

	redirect, err := url.Parse(redirectURI)
	if err != nil {
		return err
	}

	baseHost, redirectHost := base.Hostname(), redirect.Hostname()
	if baseHost == "" || redirectHost == "" {
		return errors.ErrInvalidRedirectURI
	}

	if redirectHost != baseHost && !strings.HasSuffix(redirectHost, "."+baseHost) {
		return errors.ErrInvalidRedirectURI
	}

	return nil
}

func buildOAuth2ErrorHandler(logger logging.Logger) func(*errors.Response) {
	return func(res *errors.Response) {
		observability.AcknowledgeError(res.Error, logger, nil, "oauth2 response error")
	}
}

func buildInternalErrorHandler(logger logging.Logger) func(error) *errors.Response {
	return func(err error) *errors.Response {
		observability.AcknowledgeError(err, logger, nil, "internal oauth2 error")
		return &errors.Response{
			Error:       err,
			ErrorCode:   -1,
			Description: err.Error(),
			URI:         "",
			StatusCode:  http.StatusInternalServerError,
			Header:      nil,
		}
	}
}

// this determines how we identify clients from HTTP requests.
func buildClientInfoHandler() func(*http.Request) (string, string, error) {
	return func(req *http.Request) (string, string, error) {
		clientID, clientSecret := req.Form.Get("client_id"), req.Form.Get("client_secret")
		if clientID == "" || clientSecret == "" {
			username, password, ok := req.BasicAuth()
			if !ok {
				return "", "", errors.ErrInvalidClient
			}

			return username, password, nil
		}

		return clientID, clientSecret, nil
	}
}

func buildUserAuthorizationHandler(tracer tracing.Tracer, logger logging.Logger, tokenIssuer tokens.Issuer) func(http.ResponseWriter, *http.Request) (string, error) {
	return func(res http.ResponseWriter, req *http.Request) (userID string, err error) {
		ctx, span := tracer.StartCustomSpan(req.Context(), "oauth2_server.UserAuthorizationHandler")
		defer span.End()

		l := logger.WithRequest(req)
		l.Info("UserAuthorizationHandler invoked")

		rawToken := req.Header.Get("Authorization")
		token := strings.TrimPrefix(rawToken, "Bearer ")

		claims, err := tokenIssuer.ParseToken(ctx, token)
		if err != nil {
			l.Error("parsing token in UserAuthorizationHandler", err)
			return "", errors.ErrAccessDenied
		}

		return claims.Subject(), nil
	}
}

func AuthorizeScopeHandler(_ logging.Logger) func(http.ResponseWriter, *http.Request) (string, error) {
	return func(_ http.ResponseWriter, req *http.Request) (scope string, err error) {
		return req.URL.Query().Get("scope"), nil
	}
}

func AccessTokenExpHandler(_ logging.Logger) func(http.ResponseWriter, *http.Request) (time.Duration, error) {
	return func(_ http.ResponseWriter, _ *http.Request) (time.Duration, error) {
		return 24 * time.Hour, nil
	}
}

func ClientScopeHandler(_ logging.Logger) func(_ *oauth2.TokenGenerateRequest) (allowed bool, err error) {
	return func(_ *oauth2.TokenGenerateRequest) (allowed bool, err error) {
		return true, nil
	}
}
