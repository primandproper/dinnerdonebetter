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

	"github.com/primandproper/platform-go/v12/authentication/oauth2server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// The API server's OAuth 2.1 authorization server is the platform's, mounted at the paths its
// discovery document names. This file drives those routes directly, so the negative cases and
// the parameters the login helper never varies are reachable.
//
// It began as a characterization suite against the go-oauth2 server this replaced (#1339), and
// several of its cases pinned behavior that was wrong: a lookup miss answered 500 with
// "sql: no rows in result set" on the wire, `plain` PKCE was accepted, a rejected redirect_uri
// was reported by redirecting to the very URI that had just been rejected, and the refresh grant
// never checked the client secret. Each of those is now asserted in its corrected form, with the
// old behavior named in the comment — the point of pinning them was that the replacement would
// have to arrive as a visible diff to this file, and this is that diff.

const (
	oauth2AuthorizePath = "/authorize"
	oauth2TokenPath     = "/token"
	oauth2MetadataPath  = "/.well-known/oauth-authorization-server"

	// codeChallengeMethodPlain is the method this server refuses. RFC 7636 defines it, and it
	// puts the verifier in the authorization request — the request PKCE exists to protect.
	codeChallengeMethodPlain = "plain"
	codeChallengeMethodS256  = "S256"
)

// oauth2TokenResponse is the subset of a /token response either outcome can carry: the tokens on
// success, the error fields on failure. Both arrive as JSON, so one shape reads both.
type oauth2TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

// oauth2RedirectURIForTest is the redirect URI the suite authorizes against: the API server's own
// address, which init.go registers on the integration client. Nothing listens for the redirect —
// the code is read off the Location header.
//
// It is now matched byte for byte, so this string and the registered one have to be identical
// rather than merely compatible.
func oauth2RedirectURIForTest() string {
	return httpTestServerAddress
}

// authorizeForTest drives POST /authorize with an arbitrary query and returns the response
// alongside the parsed Location of the redirect it answered with. It never follows the redirect,
// because the authorization code is the redirect.
//
// POST, not GET. A GET renders the login form — the answer for a browser arriving without a
// session — and only a POST runs the authenticator that reads the bearer token below. The
// authorization parameters stay in the query string either way, so the request that issues the
// code is validated against the same bytes.
//
// The returned URL is nil when the server answered with something that carried no Location.
func authorizeForTest(t *testing.T, jwt string, query url.Values) (*http.Response, *url.URL) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, httpTestServerAddress+oauth2AuthorizePath+"?"+query.Encode(), http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

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

	if codeChallenge != "" {
		query.Set("code_challenge", codeChallenge)
		query.Set("code_challenge_method", codeChallengeMethod)
	}

	return query
}

// requestTokenForTest posts to /token and decodes whichever of the two response shapes came back.
// It returns the status code alongside, because several of these cases are only meaningfully
// distinguished by it.
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
	// A non-JSON body means an error the authorization server never saw — the router's 405, say.
	// The zero value is the right answer there: no token, no oauth2 error code.
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

// assertGrantRefused is the answer to a grant naming a record that is absent, spent or expired.
//
// It replaces assertLookupMissLeaksInternalError, which pinned what this used to do: answer 500
// and echo the repository's own "sql: no rows in result set" to an unauthenticated caller. Both
// halves were wrong, and both are asserted fixed here — a 400 with invalid_grant, and no SQL
// anywhere in the body.
func assertGrantRefused(t *testing.T, status int, token *oauth2TokenResponse) {
	t.Helper()

	assert.Equal(t, http.StatusBadRequest, status)
	assert.Equal(t, "invalid_grant", token.Error)
	assert.Empty(t, token.AccessToken)
	assert.NotContains(t, token.Error+token.ErrorDescription, "sql:")
}

// createUserAndJWTForTest makes a user that has never been through the authorization code flow,
// and returns the JWT that authenticates it at /authorize.
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
		// RFC 9207. A client holding more than one authorization server cannot detect a mix-up
		// without it, and this server sets it on every authorization response.
		assert.NotEmpty(t, location.Query().Get("iss"))

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
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "invalid_grant", token.Error)
		assert.Empty(t, token.AccessToken)
	})

	T.Run("a failed verifier check burns the code", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		status, token := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, oauth2.GenerateVerifier()))
		require.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "invalid_grant", token.Error)

		// The code is consumed before the challenge is checked, which is what stops a wrong
		// verifier from being retried. The correct verifier cannot rescue it afterwards — and
		// the refusal is now a protocol error rather than the repository's error text.
		status, retried := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		assertGrantRefused(t, status, retried)
	})

	T.Run("an omitted code verifier is rejected when a challenge was sent", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		code, _ := fetchAuthorizationCodeForTest(t, jwt)

		// A 400, where the go-oauth2 server answered 500: ErrMissingCodeVerifier was not one of
		// the three errors it mapped to invalid_grant, so it fell through to the internal error
		// handler. No token was issued either way; the status is what this corrects.
		status, token := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, ""))
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Empty(t, token.AccessToken)
		assert.NotEmpty(t, token.Error)
	})

	T.Run("an authorization code is single use", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		status, token := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		require.Equal(t, http.StatusOK, status, "first exchange failed: %s", token.ErrorDescription)
		require.NotEmpty(t, token.AccessToken)

		status, replayed := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		assertGrantRefused(t, status, replayed)
	})

	T.Run("a replayed authorization code is refused but does not revoke what it issued", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, jwt := createUserAndJWTForTest(t)

		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		status, token := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		require.Equal(t, http.StatusOK, status, "first exchange failed: %s", token.ErrorDescription)
		require.NotEmpty(t, token.AccessToken)

		status, _ = requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		require.Equal(t, http.StatusBadRequest, status)

		// RFC 6749 §4.1.2: a code presented twice SHOULD revoke what it previously issued — a
		// second redemption is either a client retrying or somebody else holding the code, and
		// the server cannot tell which, so it assumes the worse of the two.
		//
		// This assertion used to be its own inverse, characterizing a platform gap:
		// AuthorizationCode carried no FamilyID, so oauth2server could count a replay and log
		// it but had no way to name the tokens the first redemption minted. platform-go v12
		// closed that, and the characterization turned red on the version bump exactly as it
		// was pinned to.
		//
		// Refresh reuse detection, which has always had a family, is asserted in
		// TestAuth_OAuth2RefreshTokenGrant.
		c, err := buildAuthedGRPCClientWithBearerToken(token.AccessToken)
		require.NoError(t, err)

		_, err = c.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		assert.Error(t, err, "a replayed code must revoke the tokens it issued")
	})

	T.Run("an expired authorization code is rejected", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, jwt := createUserAndJWTForTest(t)

		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		// Codes are short-lived but not short enough for a test to wait out. The consuming
		// UPDATE's predicate is `redeemed_at IS NULL AND expires_at > now`, so moving expires_at
		// into the past is exactly what an elapsed lifetime looks like.
		//
		// The row is addressable by its own hash rather than by user, because the platform's
		// table stores the digest of the code the client holds — which is the one thing a test
		// holding the plaintext can compute.
		result, err := databaseClient.Writer().ExecContext(ctx,
			`UPDATE ddb_oauth2_authorization_codes SET expires_at = NOW() - INTERVAL '1 hour' WHERE hash = $1`,
			oauth2HashForTest(code),
		)
		require.NoError(t, err)

		affected, err := result.RowsAffected()
		require.NoError(t, err)
		require.Equal(t, int64(1), affected, "expected exactly one row for the code just issued")

		status, token := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		assertGrantRefused(t, status, token)
	})

	T.Run("the redirect_uri at the token endpoint must match the one from the authorize request", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		form := exchangeCodeFormForTest(code, verifier)
		form.Set("redirect_uri", oauth2RedirectURIForTest()+"1")

		status, token := requestTokenForTest(t, http.MethodPost, form)
		assert.Equal(t, http.StatusBadRequest, status)
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

		// invalid_client with a 401, where this used to be a 500 carrying the repository's SQL
		// error. Translating the registry's lookup miss into the protocol's "no such client" is
		// what the store decorator exists to do.
		status, token := requestTokenForTest(t, http.MethodPost, form)
		assert.Equal(t, http.StatusUnauthorized, status)
		assert.Equal(t, "invalid_client", token.Error)
		assert.NotContains(t, token.Error+token.ErrorDescription, "sql:")
		assert.Empty(t, token.AccessToken)
	})

	T.Run("an authorize request without a bearer token renders the login form", func(t *testing.T) {
		t.Parallel()

		verifier := oauth2.GenerateVerifier()

		// The old server answered this by redirecting to the client with access_denied. This one
		// asks: a POST with neither a session token nor credentials is a failed sign-in, and the
		// answer to a failed sign-in is the form again, because the human is still there.
		res, _ := authorizeForTest(t, "", authorizeQueryForTest(t.Name(), oauth2.S256ChallengeFromVerifier(verifier), codeChallengeMethodS256))
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
		assert.Contains(t, res.Header.Get("Content-Type"), "text/html")
	})

	T.Run("GET on the authorize endpoint renders the login form rather than a code", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		verifier := oauth2.GenerateVerifier()
		query := authorizeQueryForTest(t.Name(), oauth2.S256ChallengeFromVerifier(verifier), codeChallengeMethodS256)

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, httpTestServerAddress+oauth2AuthorizePath+"?"+query.Encode(), http.NoBody)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+jwt)

		httpClient, err := localdev.NewNonRedirectingHTTPClient()
		require.NoError(t, err)

		res, err := httpClient.Do(req)
		require.NoError(t, err)
		t.Cleanup(func() { assert.NoError(t, res.Body.Close()) })

		// A GET is a browser asking to sign in, and no credential in a header changes that. This
		// is the property every first-party client had to be changed for.
		assert.Equal(t, http.StatusOK, res.StatusCode)
		assert.Contains(t, res.Header.Get("Content-Type"), "text/html")
	})

	T.Run("GET on the token endpoint returns no token", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		code, verifier := fetchAuthorizationCodeForTest(t, jwt)

		status, token := requestTokenForTest(t, http.MethodGet, exchangeCodeFormForTest(code, verifier))
		assert.Equal(t, http.StatusMethodNotAllowed, status)
		assert.Empty(t, token.AccessToken)

		// The code must still be redeemable — the GET must not have consumed it.
		status, exchanged := requestTokenForTest(t, http.MethodPost, exchangeCodeFormForTest(code, verifier))
		assert.Equal(t, http.StatusOK, status)
		assert.NotEmpty(t, exchanged.AccessToken)
	})

	T.Run("the plain code challenge method is refused", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		// The go-oauth2 server accepted this, and the suite's own login helper sent it until
		// #1339. `plain` puts the verifier in the authorization request — the request PKCE
		// exists to protect — so this server has no configuration that turns it back on.
		verifier := oauth2.GenerateVerifier()

		res, location := authorizeForTest(t, jwt, authorizeQueryForTest(t.Name(), verifier, codeChallengeMethodPlain))
		require.Equal(t, http.StatusFound, res.StatusCode)
		require.NotNil(t, location)

		assert.Empty(t, location.Query().Get("code"))
		assert.Equal(t, "invalid_request", location.Query().Get("error"))
	})

	T.Run("an omitted code challenge is refused", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		// PKCE is mandatory, and an absent method defaults to `plain` under RFC 7636 — so
		// silence is not agreement, it is a request for the method this server refuses.
		res, location := authorizeForTest(t, jwt, authorizeQueryForTest(t.Name(), "", ""))
		require.Equal(t, http.StatusFound, res.StatusCode)
		require.NotNil(t, location)

		assert.Empty(t, location.Query().Get("code"))
		assert.Equal(t, "invalid_request", location.Query().Get("error"))
	})

	T.Run("an unsupported code challenge method issues no code", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		verifier := oauth2.GenerateVerifier()

		res, location := authorizeForTest(t, jwt, authorizeQueryForTest(t.Name(), oauth2.S256ChallengeFromVerifier(verifier), "S512"))
		require.Equal(t, http.StatusFound, res.StatusCode)
		require.NotNil(t, location)

		assert.Empty(t, location.Query().Get("code"))
		assert.Equal(t, "invalid_request", location.Query().Get("error"))
	})
}

// TestAuth_OAuth2RedirectURIValidation covers redirect URI matching, which the happy path never
// exercises because it only ever passes the exact registered value.
//
// Every case but the first used to pass: the old validateRedirectURI matched on hostname with a
// dot boundary and deliberately ignored ports, so a subdomain and a different port were both
// accepted. Matching is now byte for byte, which is what OAuth 2.1 requires and what makes the
// registered list mean what it says.
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

		if res.StatusCode == http.StatusFound && location != nil {
			if code := location.Query().Get("code"); code != "" {
				return true
			}
		}

		// Answered in the browser, with no Location at all. The old server reported a rejected
		// redirect_uri *by redirecting to it*, which RFC 6749 §4.1.2.1 names as the one thing
		// not to do: the URI it is redirecting to is the URI it has just decided is not the
		// client's.
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		assert.Nil(t, location)

		return false
	}

	T.Run("the registered URI is accepted", func(t *testing.T) {
		t.Parallel()

		assert.True(t, authorizeWithRedirectURIForTest(t, oauth2RedirectURIForTest()))
	})

	T.Run("a different port on the registered host is rejected", func(t *testing.T) {
		t.Parallel()

		assert.False(t, authorizeWithRedirectURIForTest(t, "http://localhost:1"))
	})

	T.Run("a subdomain of the registered host is rejected", func(t *testing.T) {
		t.Parallel()

		assert.False(t, authorizeWithRedirectURIForTest(t, "http://sub.localhost:1"))
	})

	T.Run("a trailing slash on the registered URI is rejected", func(t *testing.T) {
		t.Parallel()

		// The case byte-exact matching is really about: a string a human would call the same
		// address, and a comparison that does not.
		assert.False(t, authorizeWithRedirectURIForTest(t, oauth2RedirectURIForTest()+"/"))
	})

	T.Run("a host merely ending in the registered host is rejected", func(t *testing.T) {
		t.Parallel()

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

	T.Run("an omitted redirect_uri is rejected", func(t *testing.T) {
		t.Parallel()

		_, jwt := createUserAndJWTForTest(t)

		// Not defaulted to the single registered URI, though RFC 6749 permits that when there is
		// exactly one. A client that omits it is a client that has not decided.
		verifier := oauth2.GenerateVerifier()
		query := authorizeQueryForTest(t.Name(), oauth2.S256ChallengeFromVerifier(verifier), codeChallengeMethodS256)
		query.Del("redirect_uri")

		res, location := authorizeForTest(t, jwt, query)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		assert.Nil(t, location)
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
		assertGrantRefused(t, status, replayed)

		oldClient, err := buildAuthedGRPCClientWithBearerToken(token.AccessToken)
		require.NoError(t, err)

		_, err = oldClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		assert.Error(t, err, "the access token the refresh replaced must stop working")
	})

	T.Run("replaying a refresh token revokes its whole family", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, token := fetchTokenPairForTest(t)

		status, refreshed := requestTokenForTest(t, http.MethodPost, refreshTokenFormForTest(token.RefreshToken))
		require.Equal(t, http.StatusOK, status, "refresh failed: %s", refreshed.ErrorDescription)
		require.NotEmpty(t, refreshed.AccessToken)

		// Rotation without reuse detection detects nothing: the replay is refused and the copy
		// the attacker is actually using keeps working. Presenting the spent token ends the
		// family, so the token the legitimate client is now holding stops working too — which is
		// the point. Somebody has a copy, and the server cannot tell which of the two is which.
		status, _ = requestTokenForTest(t, http.MethodPost, refreshTokenFormForTest(token.RefreshToken))
		require.Equal(t, http.StatusBadRequest, status)

		c, err := buildAuthedGRPCClientWithBearerToken(refreshed.AccessToken)
		require.NoError(t, err)

		_, err = c.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		assert.Error(t, err, "a detected refresh replay must revoke the whole family")
	})

	T.Run("an unknown refresh token is rejected", func(t *testing.T) {
		t.Parallel()

		status, token := requestTokenForTest(t, http.MethodPost, refreshTokenFormForTest("not-a-refresh-token"))
		assertGrantRefused(t, status, token)
	})

	T.Run("a wrong client secret stops the refresh grant", func(t *testing.T) {
		t.Parallel()

		_, token := fetchTokenPairForTest(t)

		// This is the defect #1339 pinned, now fixed. go-oauth2 sent the refresh grant straight
		// to Manager.RefreshAccessToken, which looked the client up by the *token's* client ID
		// and never verified the secret the request presented — so a leaked refresh token was
		// redeemable without the client credential that scoped it. Client authentication now
		// happens once, before the grant is dispatched.
		form := refreshTokenFormForTest(token.RefreshToken)
		form.Set("client_secret", "not-the-client-secret")

		status, refreshed := requestTokenForTest(t, http.MethodPost, form)
		assert.Equal(t, http.StatusUnauthorized, status)
		assert.Equal(t, "invalid_client", refreshed.Error)
		assert.Empty(t, refreshed.AccessToken)
	})
}

// TestAuth_OAuth2PasswordGrant pins that the password grant cannot mint a token. OAuth 2.1 removes
// it, and this server has no configuration that brings it back.
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
		assert.Equal(t, http.StatusBadRequest, status)
		assert.Equal(t, "unsupported_grant_type", token.Error)
		assert.Empty(t, token.AccessToken)
	})
}

// TestAuth_OAuth2Metadata covers the RFC 8414 discovery document.
func TestAuth_OAuth2Metadata(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, httpTestServerAddress+oauth2MetadataPath, http.NoBody)
		require.NoError(t, err)

		res, err := (&http.Client{}).Do(req)
		require.NoError(t, err)
		defer func() { assert.NoError(t, res.Body.Close()) }()

		require.Equal(t, http.StatusOK, res.StatusCode)

		var metadata struct {
			Issuer                        string   `json:"issuer"`
			AuthorizationEndpoint         string   `json:"authorization_endpoint"`
			TokenEndpoint                 string   `json:"token_endpoint"`
			RegistrationEndpoint          string   `json:"registration_endpoint"`
			GrantTypesSupported           []string `json:"grant_types_supported"`
			CodeChallengeMethodsSupported []string `json:"code_challenge_methods_supported"`
		}
		require.NoError(t, json.NewDecoder(res.Body).Decode(&metadata))

		// The endpoints have to be where they are actually mounted, which is what mounting at
		// the paths the document derives buys.
		assert.Equal(t, httpTestServerAddress, metadata.Issuer)
		assert.Equal(t, httpTestServerAddress+oauth2AuthorizePath, metadata.AuthorizationEndpoint)
		assert.Equal(t, httpTestServerAddress+oauth2TokenPath, metadata.TokenEndpoint)

		// Absent, because this server does not serve RFC 7591 registration: a client here is
		// created through the permission-gated gRPC surface. Advertising an endpoint that
		// answers 404 is worse than advertising nothing.
		assert.Empty(t, metadata.RegistrationEndpoint)

		assert.ElementsMatch(t, []string{"authorization_code", "refresh_token"}, metadata.GrantTypesSupported)
		assert.Equal(t, []string{codeChallengeMethodS256}, metadata.CodeChallengeMethodsSupported)
	})
}

// oauth2HashForTest is how a credential appears in the authorization server's tables: the
// hex-encoded SHA-256 digest of the value the client holds, never the value itself. A test that
// wants to reach the row behind a code it is holding has to compute the same digest.
func oauth2HashForTest(credential string) string {
	return oauth2server.Hash(credential)
}
