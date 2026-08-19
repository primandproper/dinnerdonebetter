package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	"github.com/primandproper/platform-go/v11/authentication/oauth2server"
	oauth2memory "github.com/primandproper/platform-go/v11/authentication/oauth2server/memory"
	"github.com/primandproper/platform-go/v11/observability"
	"github.com/primandproper/platform-go/v11/routing/backends/chi"
	routingcfg "github.com/primandproper/platform-go/v11/routing/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// exampleCodeVerifier is 43 characters of the unreserved set, which is the
// shortest RFC 7636 permits.
const exampleCodeVerifier = "abcdefghijklmnopqrstuvwxyz0123456789-._~ABC"

// buildTestRouter stands up the whole MCP surface — the six authorization server
// endpoints, the resource metadata document, and a stub MCP handler behind the
// bearer middleware — over a memory store.
//
// It is the wiring that is under test here, not the protocol: the platform holds
// the conformance suite for the latter. What this catches is a route mounted at
// the wrong path, a verifier handed the wrong resource identifier, or a challenge
// header pointing somewhere a client cannot follow.
func buildTestRouter(t *testing.T, subject *oauth2server.Subject) (handler http.Handler, resource string) {
	t.Helper()

	ctx := t.Context()

	srv, err := oauth2server.NewServer(exampleResource, oauth2memory.NewStore(),
		oauth2server.SubjectAuthenticatorFunc(func(context.Context, *http.Request) (*oauth2server.Subject, error) {
			return subject, nil
		}),
		oauth2server.WithLoginRenderer(newLoginRenderer(nil)),
		oauth2server.WithResources(exampleResource),
	)
	require.NoError(t, err)

	resourceMetadata, err := oauth2server.NewResourceMetadata(exampleResource, []string{exampleResource})
	require.NoError(t, err)

	mcpHandler := http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
		res.WriteHeader(http.StatusTeapot)
	})

	router, err := buildRouter(ctx, mcpHandler, srv, resourceMetadata, &observability.Pillars{},
		&routingcfg.Config{Provider: routingcfg.ProviderChi, Chi: &chi.Config{ServiceName: t.Name()}},
		exampleResource,
	)
	require.NoError(t, err)

	return router.Handler(), exampleResource
}

func TestBuildRouter(T *testing.T) {
	T.Parallel()

	T.Run("publishes both discovery documents", func(t *testing.T) {
		t.Parallel()

		handler, resource := buildTestRouter(t, nil)

		for path, key := range map[string]string{
			oauth2server.PathAuthorizationServerMetadata: "issuer",
			oauth2server.PathProtectedResourceMetadata:   "resource",
		} {
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))

			require.Equal(t, http.StatusOK, res.Code, path)

			doc := map[string]any{}
			require.NoError(t, json.Unmarshal(res.Body.Bytes(), &doc))
			assert.Equal(t, resource, doc[key], path)
		}
	})

	T.Run("challenges an unauthenticated MCP request", func(t *testing.T) {
		t.Parallel()

		handler, resource := buildTestRouter(t, nil)

		res := httptest.NewRecorder()
		handler.ServeHTTP(res, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/mcp", nil))

		require.Equal(t, http.StatusUnauthorized, res.Code)

		// Without this header a client that was never configured with this server
		// gets a 401 and stops, rather than discovering where to authenticate.
		assert.Contains(t, res.Header().Get("WWW-Authenticate"),
			resource+oauth2server.PathProtectedResourceMetadata)
	})

	T.Run("register, authorize, token, and call", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUser := identityfakes.BuildFakeUser()
		exampleAccountID := identityfakes.BuildFakeAccount().ID

		handler, resource := buildTestRouter(t, &oauth2server.Subject{
			ID:     exampleUser.ID,
			Claims: map[string]string{claimAccountID: exampleAccountID},
		})

		post := func(path, contentType, body string) *httptest.ResponseRecorder {
			t.Helper()
			req := httptest.NewRequestWithContext(ctx, http.MethodPost, path, strings.NewReader(body))
			req.Header.Set("Content-Type", contentType)
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)

			return res
		}

		// 1. RFC 7591 dynamic registration.
		const redirectURI = "http://127.0.0.1:33418/callback"
		res := post(oauth2server.PathRegister, "application/json",
			fmt.Sprintf(`{"redirect_uris":[%q],"client_name":"Test Client","token_endpoint_auth_method":"none"}`, redirectURI))
		require.Equal(t, http.StatusCreated, res.Code, res.Body.String())

		var registration oauth2server.RegistrationResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &registration))
		require.NotEmpty(t, registration.ClientID)

		authorizeQuery := url.Values{
			"response_type":         {oauth2server.ResponseTypeCode},
			"client_id":             {registration.ClientID},
			"redirect_uri":          {redirectURI},
			"code_challenge":        {oauth2server.S256Challenge(exampleCodeVerifier)},
			"code_challenge_method": {oauth2server.CodeChallengeMethodS256},
			"resource":              {resource},
			"state":                 {"opaque-state"},
		}.Encode()

		// 2. The login form, rendered by this application's renderer.
		res = httptest.NewRecorder()
		handler.ServeHTTP(res, httptest.NewRequestWithContext(ctx, http.MethodGet,
			oauth2server.PathAuthorize+"?"+authorizeQuery, nil))
		require.Equal(t, http.StatusOK, res.Code)
		assert.Contains(t, res.Body.String(), "Test Client")
		assert.Contains(t, res.Body.String(), `name="totp_token"`)

		// 3. Signing in, which redirects back with an authorization code.
		res = post(oauth2server.PathAuthorize+"?"+authorizeQuery, "application/x-www-form-urlencoded", "")
		require.Equal(t, http.StatusFound, res.Code, res.Body.String())

		location, err := url.Parse(res.Header().Get("Location"))
		require.NoError(t, err)
		assert.Equal(t, "opaque-state", location.Query().Get("state"))

		code := location.Query().Get("code")
		require.NotEmpty(t, code)

		// 4. Redeeming it.
		res = post(oauth2server.PathToken, "application/x-www-form-urlencoded", url.Values{
			"grant_type":    {oauth2server.GrantTypeAuthorizationCode},
			"code":          {code},
			"code_verifier": {exampleCodeVerifier},
			"client_id":     {registration.ClientID},
			"redirect_uri":  {redirectURI},
		}.Encode())
		require.Equal(t, http.StatusOK, res.Code, res.Body.String())

		var token oauth2server.TokenResponse
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &token))
		require.NotEmpty(t, token.AccessToken)

		// 5. The token reaches the MCP handler behind the bearer middleware.
		req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/mcp", nil)
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)
		res = httptest.NewRecorder()
		handler.ServeHTTP(res, req)

		assert.Equal(t, http.StatusTeapot, res.Code, readAll(t, res.Body))
	})
}

func readAll(t *testing.T, r io.Reader) string {
	t.Helper()

	b, err := io.ReadAll(r)
	require.NoError(t, err)

	return string(b)
}
