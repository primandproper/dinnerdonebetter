package integration

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/localdev"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// The API server's OAuth2 authorization server is two routes — GET /oauth2/authorize and
// POST /oauth2/token — and until this file existed they were only ever driven by
// localdev.FetchOAuth2TokenForUser, on the happy path, from every authed test's login. That
// helper is deliberately rigid: one redirect URI, one PKCE method, no way to send a bad
// verifier. Everything here talks to the two routes directly instead, so the negative cases and
// the parameters the happy path never varies are reachable.
//
// Several of these tests pin behavior that is wrong, or right only incidentally. They say so
// where they do, and assert the exact status and error rather than merely "not a token", so
// that #1288 — which replaces this implementation with platform's OAuth 2.1 server — changes
// them as a visible diff instead of silently.

const (
	oauth2AuthorizePath = "/oauth2/authorize"
	oauth2TokenPath     = "/oauth2/token"

	// codeChallengeMethodPlain is the method the suite's login helper used to send, and the one
	// the OAuth 2.1 server in #1288 does not accept. It appears here only in the test that pins
	// today's acceptance of it.
	codeChallengeMethodPlain = "plain"
	codeChallengeMethodS256  = "S256"
)

// oauth2TokenResponse is the subset of an /oauth2/token response either outcome can carry: the
// tokens on success, the error fields on failure. Both arrive as JSON, so one shape reads both.
type oauth2TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// oauth2RedirectURIForTest is the redirect URI the suite authorizes against: the API server's
// own address. Nothing listens for the redirect — the code is read off the Location header.
func oauth2RedirectURIForTest() string {
	return httpTestServerAddress
}

// authorizeForTest drives GET /oauth2/authorize with an arbitrary query and returns the response
// alongside the parsed Location of the redirect it answered with. It never follows the redirect,
// because the authorization code is the redirect.
//
// The returned URL is nil when the server answered with something that carried no Location.
func authorizeForTest(t *testing.T, jwt string, query url.Values) (*http.Response, *url.URL) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, httpTestServerAddress+oauth2AuthorizePath+"?"+query.Encode(), http.NoBody)
	require.NoError(t, err)

	if jwt != "" {
		req.Header.Set("Authorization", "Bearer "+jwt)
	}

	httpClient, err := localdev.NewNonRedirectingHTTPClient()
	require.NoError(t, err)

	res, err := httpClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() {
		assert.NoError(t, res.Body.Close())
	})

	location, err := res.Location()
	if err != nil {
		return res, nil
	}

	return res, location
}

// authorizeQueryForTest builds the query for an authorization request that would succeed, so a
// caller can vary exactly the one parameter it is testing.
func authorizeQueryForTest(state, codeChallenge, codeChallengeMethod string) url.Values {
	query := url.Values{}
	query.Set("response_type", "code")
	query.Set("client_id", createdClientID)
	query.Set("redirect_uri", oauth2RedirectURIForTest())
	query.Set("state", state)
	query.Set("scope", "anything")

	if codeChallenge != "" {
		query.Set("code_challenge", codeChallenge)
		query.Set("code_challenge_method", codeChallengeMethod)
	}

	return query
}

// requestTokenForTest posts to /oauth2/token and decodes whichever of the two response shapes
// came back. It returns the status code alongside, because several of these cases are only
// meaningfully distinguished by it.
func requestTokenForTest(t *testing.T, method string, form url.Values) (int, *oauth2TokenResponse) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), method, httpTestServerAddress+oauth2TokenPath, strings.NewReader(form.Encode()))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := (&http.Client{}).Do(req)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, res.Body.Close())
	}()

	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	parsed := &oauth2TokenResponse{}
	// A non-JSON body means an error the oauth2 library never saw — the router's 405, say. The
	// zero value is the right answer there: no token, no oauth2 error code.
	if json.Valid(body) {
		require.NoError(t, json.Unmarshal(body, parsed))
	}

	return res.StatusCode, parsed
}

// exchangeCodeFormForTest builds the form for an authorization code exchange that would succeed.
func exchangeCodeFormForTest(code, verifier string) url.Values {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", createdClientID)
	form.Set("client_secret", createdClientSecret)
	form.Set("redirect_uri", oauth2RedirectURIForTest())
	form.Set("code", code)

	if verifier != "" {
		form.Set("code_verifier", verifier)
	}

	return form
}

// assertLookupMissLeaksInternalError pins the response the token endpoint gives when the record
// a grant names — a code, a refresh token, a client — simply is not there.
//
// It is a 500 carrying the repository's own error text, "sql: no rows in result set", to an
// unauthenticated caller. The token store reports a lookup miss as an error rather than as the
// (nil, nil) the oauth2 library reads as "not found", so the library has no oauth2 error to map,
// hands it to InternalErrorHandler, and echoes the description back. Both halves are wrong: the
// status should be a 400 with invalid_grant (or a 401 with invalid_client), and the repository's
// error text should never reach the wire.
//
// Pinned rather than fixed because #1288 replaces this token store and this error plumbing
// wholesale. It is here so that replacement has to account for it.
func assertLookupMissLeaksInternalError(t *testing.T, status int, token *oauth2TokenResponse) {
	t.Helper()

	assert.Equal(t, http.StatusInternalServerError, status)
	assert.Empty(t, token.AccessToken)
	assert.Contains(t, token.Error, "sql: no rows in result set")
}

// createUserAndJWTForTest makes a user that has never been through the authorization code flow,
// and returns the JWT that authenticates it to /oauth2/authorize.
//
// The suite's usual createUserAndClientForTest cannot be used here: building its client runs the
// flow, which leaves oauth2_client_tokens rows behind and makes "the row this test just created"
// ambiguous for the tests that reach into the database.
func createUserAndJWTForTest(t *testing.T) (user *identity.User, jwt string) {
	t.Helper()

	user = createServiceUserForTest(t, true, buildUserRegistrationInputForTest(t))

	return user, fetchLoginTokenForUserForTest(t, user)
}

// fetchAuthorizationCodeForTest runs the authorize leg with S256 PKCE and returns the code along
// with the verifier that will redeem it.
func fetchAuthorizationCodeForTest(t *testing.T, jwt string) (code, verifier string) {
	t.Helper()

	verifier = oauth2.GenerateVerifier()

	res, location := authorizeForTest(t, jwt, authorizeQueryForTest(t.Name(), oauth2.S256ChallengeFromVerifier(verifier), codeChallengeMethodS256))
	require.Equal(t, http.StatusFound, res.StatusCode)
	require.NotNil(t, location)

	code = location.Query().Get("code")
	require.NotEmpty(t, code)

	return code, verifier
}

func TestAuth_OAuth2AuthorizationCodeFlow(T *testing.T) {
	T.Parallel()

	T.Run("S256 PKCE round trip yields a usable access token", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, jwt := createUserAndJWTForTest(t)

		verifier := oauth2.GenerateVerifier()
		state := t.Name()

		res, location := authorizeForTest(t, jwt, authorizeQueryForTest(state, oauth2.S256ChallengeFromVerifier(verifier), codeChallengeMethodS256))
		require.Equal(t, http.StatusFound, res.StatusCode)
		require.NotNil(t, location)

		assert.Equal(t, state, location.Query().Get("state"), "state must be echoed back on the redirect")

		code := location.Query().Get("code")
		require.NotEmpty(t, code)

		status, token := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		require.Equal(t, http.StatusOK, status, "token exchange failed: %s", token.ErrorDescription)
		require.NotEmpty(t, token.AccessToken)
		assert.NotEmpty(t, token.RefreshToken)
		assert.Equal(t, "Bearer", token.TokenType)

		c, err := buildAuthedGRPCClientWithBearerToken(token.AccessToken)
		require.NoError(t, err)

		authStatus, err := c.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.NoError(t, err)
		assert.Equal(t, user.ID, authStatus.UserId)
	})

	T.Run("a code verifier that does not match the challenge is rejected", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		code, _ := fetchAuthorizationCodeForTest(t, jwt)

		status, token := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, oauth2.GenerateVerifier()))
		assert.Equal(t, http.StatusUnauthorized, status)
		assert.Equal(t, "invalid_grant", token.Error)
		assert.Empty(t, token.AccessToken)
	})

	T.Run("a failed verifier check burns the code", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		status, token := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, oauth2.GenerateVerifier()))
		require.Equal(t, http.StatusUnauthorized, status)
		assert.Equal(t, "invalid_grant", token.Error)

		// The code is deleted before the challenge is checked, so the correct verifier cannot
		// rescue it afterwards. That ordering is what stops a wrong verifier from being retried,
		// and it is the reason this second attempt reports a missing record rather than a bad
		// grant.
		status, retried := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		assertLookupMissLeaksInternalError(t, status, retried)
	})

	T.Run("an omitted code verifier is rejected when a challenge was sent", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		code, _ := fetchAuthorizationCodeForTest(t, jwt)

		// ErrMissingCodeVerifier is not one of the three errors GetAccessToken maps to
		// invalid_grant, so it falls through to InternalErrorHandler and comes back a 500. No
		// token is issued either way; the status is the part #1288 should correct to a 400.
		status, token := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, ""))
		assert.Equal(t, http.StatusInternalServerError, status)
		assert.Equal(t, "missing code verifier", token.Error)
		assert.Empty(t, token.AccessToken)
	})

	T.Run("an authorization code is single use", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		status, token := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		require.Equal(t, http.StatusOK, status, "first exchange failed: %s", token.ErrorDescription)
		require.NotEmpty(t, token.AccessToken)

		status, replayed := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		assertLookupMissLeaksInternalError(t, status, replayed)
	})

	T.Run("an expired authorization code is rejected", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, jwt := createUserAndJWTForTest(t)

		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		// Codes live ten minutes, which no test is going to wait out. The check is
		// code_created_at + (code_expires_at - code_created_at) < now, so sliding both stamps
		// into the past expires the code without changing the lifetime it was issued with.
		//
		// The user is addressable here precisely because it has been through the flow once and
		// only once: createUserAndJWTForTest does not build an OAuth2 client.
		result, err := databaseClient.Writer().ExecContext(ctx,
			`UPDATE oauth2_client_tokens SET code_created_at = NOW() - INTERVAL '2 hours', code_expires_at = NOW() - INTERVAL '110 minutes' WHERE belongs_to_user = $1`,
			user.ID,
		)
		require.NoError(t, err)

		affected, err := result.RowsAffected()
		require.NoError(t, err)
		require.Equal(t, int64(1), affected, "expected exactly one token row for a user that has authorized once")

		status, token := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		assert.Equal(t, http.StatusUnauthorized, status)
		assert.Equal(t, "invalid_grant", token.Error)
		assert.Empty(t, token.AccessToken)
	})

	T.Run("the redirect_uri at the token endpoint must match the one from the authorize request", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		// A different port on the same host: accepted by validateRedirectURI, which ignores
		// ports, so what rejects this is the code's own record of the URI it was issued for and
		// not client registration. That is the check being pinned.
		form := exchangeCodeFormForTest(code, verifier)
		form.Set("redirect_uri", oauth2RedirectURIForTest()+"1")

		status, token := requestTokenForTest(t, http.MethodPost, form)
		assert.Equal(t, http.StatusUnauthorized, status)
		assert.Equal(t, "invalid_grant", token.Error)
		assert.Empty(t, token.AccessToken)
	})

	T.Run("a wrong client secret is rejected at the token endpoint", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		form := exchangeCodeFormForTest(code, verifier)
		form.Set("client_secret", "not-the-client-secret")

		status, token := requestTokenForTest(t, http.MethodPost, form)
		assert.Equal(t, http.StatusUnauthorized, status)
		assert.Equal(t, "invalid_client", token.Error)
		assert.Empty(t, token.AccessToken)
	})

	T.Run("an unknown client is rejected at the token endpoint", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		form := exchangeCodeFormForTest(code, verifier)
		form.Set("client_id", nonexistentID)

		status, token := requestTokenForTest(t, http.MethodPost, form)
		assertLookupMissLeaksInternalError(t, status, token)
	})

	T.Run("an authorize request without a bearer token issues no code", func(t *testing.T) {
		t.Parallel()

		verifier := oauth2.GenerateVerifier()

		res, location := authorizeForTest(t, "", authorizeQueryForTest(t.Name(), oauth2.S256ChallengeFromVerifier(verifier), codeChallengeMethodS256))
		require.Equal(t, http.StatusFound, res.StatusCode)
		require.NotNil(t, location)

		assert.Empty(t, location.Query().Get("code"))
		assert.Equal(t, "access_denied", location.Query().Get("error"))
	})

	T.Run("GET on the token endpoint returns no token", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		// SetAllowGetAccessRequest(true) tells the oauth2 library a GET token request is fine.
		// Only POST /oauth2/token is routed, so the router is what actually prevents it — a
		// property of the route table rather than of the intent, and worth a test that fails if
		// someone adds the GET route.
		status, token := requestTokenForTest(t, http.MethodGet, exchangeCodeFormForTest(code, verifier))
		assert.Equal(t, http.StatusMethodNotAllowed, status)
		assert.Empty(t, token.AccessToken)

		// The code must still be redeemable — the GET must not have consumed it.
		status, exchanged := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		assert.Equal(t, http.StatusOK, status)
		assert.NotEmpty(t, exchanged.AccessToken)
	})

	T.Run("the plain code challenge method is accepted today", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		// Characterization, not endorsement. AllowedCodeChallengeMethods still lists plain, and
		// plain is what the login helper sent until this branch. #1288 adopts an OAuth 2.1
		// server that accepts S256 only, and should flip this to an assertion of rejection.
		verifier := oauth2.GenerateVerifier()

		res, location := authorizeForTest(t, jwt, authorizeQueryForTest(t.Name(), verifier, codeChallengeMethodPlain))
		require.Equal(t, http.StatusFound, res.StatusCode)
		require.NotNil(t, location)

		code := location.Query().Get("code")
		require.NotEmpty(t, code)

		status, token := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		assert.Equal(t, http.StatusOK, status)
		assert.NotEmpty(t, token.AccessToken)
	})

	T.Run("an unsupported code challenge method issues no code", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		verifier := oauth2.GenerateVerifier()

		res, location := authorizeForTest(t, jwt, authorizeQueryForTest(t.Name(), oauth2.S256ChallengeFromVerifier(verifier), "S512"))
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		assert.Nil(t, location)
	})
}

// TestAuth_OAuth2RedirectURIValidation covers validateRedirectURI, which the happy path never
// exercises because it only ever passes the exact registered value.
//
// The registered domain for the integration client is http://localhost:9000, so the hostname
// being matched against is "localhost". Every expectation here is today's behavior: #1288
// replaces this function with a byte-exact comparison, at which point only the first case still
// passes — deliberately, and as a diff to this file rather than a silent change.
func TestAuth_OAuth2RedirectURIValidation(T *testing.T) {
	T.Parallel()

	// authorizeWithRedirectURIForTest runs the authorize leg with one redirect URI substituted
	// in, and reports whether a code came back.
	authorizeWithRedirectURIForTest := func(t *testing.T, redirectURI string) bool {
		t.Helper()

		_, jwt := createUserAndJWTForTest(t)

		verifier := oauth2.GenerateVerifier()
		query := authorizeQueryForTest(t.Name(), oauth2.S256ChallengeFromVerifier(verifier), codeChallengeMethodS256)
		query.Set("redirect_uri", redirectURI)

		res, location := authorizeForTest(t, jwt, query)
		require.Equal(t, http.StatusFound, res.StatusCode)
		require.NotNil(t, location)

		if code := location.Query().Get("code"); code != "" {
			return true
		}

		// A rejected redirect URI is reported *by redirecting to it*, which is the one thing
		// RFC 6749 §4.1.2.1 says not to do when the URI is the invalid part. No code rides
		// along, so nothing is granted, but an unvalidated URI still gets to be a Location on a
		// 302 from this server. Pinned so #1288 has to decide about it.
		assert.Equal(t, "invalid redirect uri", location.Query().Get("error"))

		return false
	}

	T.Run("the registered host is accepted", func(t *testing.T) {
		t.Parallel()

		assert.True(t, authorizeWithRedirectURIForTest(t, oauth2RedirectURIForTest()))
	})

	T.Run("a different port on the registered host is accepted", func(t *testing.T) {
		t.Parallel()

		// Ports are ignored on purpose, so localdev can register one port and take the callback
		// on another.
		assert.True(t, authorizeWithRedirectURIForTest(t, "http://localhost:1"))
	})

	T.Run("a subdomain of the registered host is accepted", func(t *testing.T) {
		t.Parallel()

		assert.True(t, authorizeWithRedirectURIForTest(t, "http://sub.localhost:1"))
	})

	T.Run("a host merely ending in the registered host is rejected", func(t *testing.T) {
		t.Parallel()

		// The dot boundary in the suffix check is the entire reason this is not a plain
		// strings.HasSuffix.
		assert.False(t, authorizeWithRedirectURIForTest(t, "http://notlocalhost:1"))
	})

	T.Run("the registered host as a subdomain of somewhere else is rejected", func(t *testing.T) {
		t.Parallel()

		assert.False(t, authorizeWithRedirectURIForTest(t, "http://localhost.example.com"))
	})

	T.Run("an unrelated host is rejected", func(t *testing.T) {
		t.Parallel()

		assert.False(t, authorizeWithRedirectURIForTest(t, "http://example.com"))
	})
}

func TestAuth_OAuth2RefreshTokenGrant(T *testing.T) {
	T.Parallel()

	// refreshTokenFormForTest builds a refresh grant request that would succeed.
	refreshTokenFormForTest := func(refreshToken string) url.Values {
		form := url.Values{}
		form.Set("grant_type", "refresh_token")
		form.Set("client_id", createdClientID)
		form.Set("client_secret", createdClientSecret)
		form.Set("refresh_token", refreshToken)

		return form
	}

	// fetchTokenPairForTest runs the whole authorization code flow and hands back the tokens.
	fetchTokenPairForTest := func(t *testing.T) (*identity.User, *oauth2TokenResponse) {
		t.Helper()

		user, jwt := createUserAndJWTForTest(t)
		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		status, token := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		require.Equal(t, http.StatusOK, status, "token exchange failed: %s", token.ErrorDescription)
		require.NotEmpty(t, token.RefreshToken)

		return user, token
	}

	T.Run("a refresh token is redeemable for a new access token", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, token := fetchTokenPairForTest(t)

		status, refreshed := requestTokenForTest(t, http.MethodPost, refreshTokenFormForTest(token.RefreshToken))
		require.Equal(t, http.StatusOK, status, "refresh failed: %s", refreshed.ErrorDescription)
		require.NotEmpty(t, refreshed.AccessToken)
		assert.NotEqual(t, token.AccessToken, refreshed.AccessToken)

		c, err := buildAuthedGRPCClientWithBearerToken(refreshed.AccessToken)
		require.NoError(t, err)

		authStatus, err := c.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.NoError(t, err)
		assert.Equal(t, user.ID, authStatus.UserId)
	})

	T.Run("redeeming a refresh token invalidates the tokens it replaces", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, token := fetchTokenPairForTest(t)

		status, refreshed := requestTokenForTest(t, http.MethodPost, refreshTokenFormForTest(token.RefreshToken))
		require.Equal(t, http.StatusOK, status, "refresh failed: %s", refreshed.ErrorDescription)
		require.NotEmpty(t, refreshed.RefreshToken)
		require.NotEqual(t, token.RefreshToken, refreshed.RefreshToken)

		status, replayed := requestTokenForTest(t, http.MethodPost, refreshTokenFormForTest(token.RefreshToken))
		assertLookupMissLeaksInternalError(t, status, replayed)

		oldClient, err := buildAuthedGRPCClientWithBearerToken(token.AccessToken)
		require.NoError(t, err)

		_, err = oldClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		assert.Error(t, err, "the access token the refresh replaced must stop working")
	})

	T.Run("an unknown refresh token is rejected", func(t *testing.T) {
		t.Parallel()

		status, token := requestTokenForTest(t, http.MethodPost, refreshTokenFormForTest("not-a-refresh-token"))
		assertLookupMissLeaksInternalError(t, status, token)
	})

	T.Run("a wrong client secret does not stop the refresh grant", func(t *testing.T) {
		t.Parallel()

		_, token := fetchTokenPairForTest(t)

		// Characterization of a defect, recorded so #1288 has to be explicit about fixing it.
		// GetAccessToken sends the refresh grant straight to Manager.RefreshAccessToken, which
		// looks the client up by the *token's* client ID and never verifies the secret the
		// request presented. The authorization code grant checks it; this one does not, so a
		// leaked refresh token is redeemable without the client credential that scoped it.
		form := refreshTokenFormForTest(token.RefreshToken)
		form.Set("client_secret", "not-the-client-secret")

		status, refreshed := requestTokenForTest(t, http.MethodPost, form)
		assert.Equal(t, http.StatusOK, status)
		assert.NotEmpty(t, refreshed.AccessToken)
	})
}

// TestAuth_OAuth2PasswordGrant pins that the password grant cannot mint a token. It is absent
// from AllowedGrantTypes, so GetAccessToken rejects it — but ValidationTokenRequest runs the
// PasswordAuthorizationHandler first, which is why this branch removed ours: the endpoint was
// verifying credentials against the database for a grant that could never succeed.
func TestAuth_OAuth2PasswordGrant(T *testing.T) {
	T.Parallel()

	T.Run("valid credentials do not yield a token", func(t *testing.T) {
		t.Parallel()

		user := createServiceUserForTest(t, true, buildUserRegistrationInputForTest(t))

		form := url.Values{}
		form.Set("grant_type", "password")
		form.Set("client_id", createdClientID)
		form.Set("client_secret", createdClientSecret)
		form.Set("username", user.Username)
		form.Set("password", user.HashedPassword)

		status, token := requestTokenForTest(t, http.MethodPost, form)
		assert.Equal(t, http.StatusForbidden, status)
		assert.Equal(t, "access_denied", token.Error)
		assert.Empty(t, token.AccessToken)
	})
}
