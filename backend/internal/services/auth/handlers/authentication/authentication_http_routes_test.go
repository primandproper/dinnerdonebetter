package authentication

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationService_AuthorizeHandler(T *testing.T) {
	T.Parallel()

	T.Run("with no client_id", func(t *testing.T) {
		t.Parallel()

		// Answered in the browser rather than by redirecting: until the client and its redirect
		// URI are known, there is nowhere safe to send an error to.
		res := httptest.NewRecorder()
		buildTestService(t).AuthorizeHandler(res, httptest.NewRequest(http.MethodGet, "/authorize", http.NoBody))

		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Empty(t, res.Header().Get("Location"))
	})
}

func TestAuthenticationService_TokenHandler(T *testing.T) {
	T.Parallel()

	T.Run("with no grant type", func(t *testing.T) {
		t.Parallel()

		res := httptest.NewRecorder()
		buildTestService(t).TokenHandler(res, httptest.NewRequest(http.MethodPost, "/token", http.NoBody))

		assert.GreaterOrEqual(t, res.Code, http.StatusBadRequest)
	})
}

func TestAuthenticationService_RevokeHandler(T *testing.T) {
	T.Parallel()

	T.Run("with no token", func(t *testing.T) {
		t.Parallel()

		res := httptest.NewRecorder()
		buildTestService(t).RevokeHandler(res, httptest.NewRequest(http.MethodPost, "/revoke", http.NoBody))

		assert.GreaterOrEqual(t, res.Code, http.StatusBadRequest)
	})
}

func TestAuthenticationService_AuthorizationServerMetadataHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		res := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", http.NoBody)

		buildTestService(t).AuthorizationServerMetadataHandler(res, req)

		require.Equal(t, http.StatusOK, res.Code)

		var metadata map[string]any
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &metadata))

		assert.Equal(t, testIssuer, metadata["issuer"])
		assert.Equal(t, testIssuer+"/authorize", metadata["authorization_endpoint"])
		assert.Equal(t, testIssuer+"/token", metadata["token_endpoint"])

		// Omitted, because this server does not serve RFC 7591 registration. Advertising an
		// endpoint that answers 404 is worse than advertising nothing: a client believes the
		// document.
		assert.NotContains(t, metadata, "registration_endpoint")
	})
}
