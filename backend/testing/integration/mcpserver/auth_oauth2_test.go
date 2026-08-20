package integration

import (
	"net/http"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v12/authentication/oauth2server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPServer_AuthorizationCodeFlow(T *testing.T) {
	T.Parallel()

	T.Run("signs an admin in and calls a tool", func(t *testing.T) {
		t.Parallel()

		accessToken := primary.authenticate(t)

		res, err := primary.getValidIngredient(t, accessToken, seededIngredient.ID)
		require.NoError(t, err)

		// The row, not an empty answer. Everything between the login form and this
		// assertion had to agree for it to arrive: the account the authenticator
		// resolved travels on the token as a claim, the tool handler refuses a request
		// without one, and the repository behind it is the one the MCP server's own
		// container built from its own database credentials.
		ingredient := &mealplanning.ValidIngredient{}
		requireStructuredContent(t, res, ingredient)

		assert.Equal(t, seededIngredient.ID, ingredient.ID)
		assert.Equal(t, seededIngredient.Name, ingredient.Name)
	})

	T.Run("challenges an unauthenticated MCP request", func(t *testing.T) {
		t.Parallel()

		res := primary.post(t, "/mcp", "application/json", `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
		defer closeBody(t, res)

		require.Equal(t, http.StatusUnauthorized, res.StatusCode)

		// Without this, a client that was never configured with this server gets a 401
		// and stops, rather than discovering where to authenticate. It names the fleet's
		// address rather than this process's, which is the only one a client can reach.
		assert.Contains(t, res.Header.Get("WWW-Authenticate"), fleetBaseURL+oauth2server.PathProtectedResourceMetadata)
	})

	T.Run("refuses a token minted for another resource server", func(t *testing.T) {
		t.Parallel()

		// A second authorization server over the same tables, naming itself rather than
		// the fleet as its resource. Sharing the store is what makes this a real replay:
		// the token below is not expired, not revoked, and this server can read it — the
		// only thing wrong with it is that it was minted for somewhere else, which is
		// the one refusal RFC 8707 leaves to the resource server.
		foreign := startInstanceForTest(t, "", nil)
		require.NotEqual(t, fleetBaseURL, foreign.baseURL)

		accessToken := primary.authenticate(t)

		// Good here, so that the refusal below is about the audience and not about a
		// token this server could not read in the first place.
		_, err := primary.getValidIngredient(t, accessToken, seededIngredient.ID)
		require.NoError(t, err)

		_, err = foreign.getValidIngredient(t, accessToken, seededIngredient.ID)
		require.ErrorContains(t, err, "Unauthorized")
	})

	T.Run("refuses a code redeemed with the wrong verifier", func(t *testing.T) {
		t.Parallel()

		registration := primary.registerClient(t)
		authz := primary.signIn(t, registration.ClientID, generateTOTPCode(t))

		// The verifier is what a public client has instead of a secret. An authorization
		// code intercepted on its way back through the user agent is worth nothing
		// without it, and only if this check is real.
		authz.verifier = "wrong-verifier-that-is-long-enough-to-be-legal"

		res := primary.post(t, oauth2server.PathToken, "application/x-www-form-urlencoded", tokenRequestBody(authz))
		defer closeBody(t, res)

		body := readBody(t, res)
		assert.Equal(t, http.StatusBadRequest, res.StatusCode, body)
		assert.Contains(t, body, oauth2server.ErrorCodeInvalidGrant)
	})
}

func TestMCPServer_Tools(T *testing.T) {
	T.Parallel()

	T.Run("lists the tools it serves", func(t *testing.T) {
		t.Parallel()

		accessToken := primary.authenticate(t)

		res, err := primary.listTools(t, accessToken)
		require.NoError(t, err)

		names := make([]string, 0, len(res.Tools))
		for _, tool := range res.Tools {
			names = append(names, tool.Name)
		}

		assert.Contains(t, names, "GetValidIngredient")
		assert.Contains(t, names, "SearchForRecipes")
	})

	T.Run("refuses a tool call with no token", func(t *testing.T) {
		t.Parallel()

		_, err := primary.getValidIngredient(t, "", seededIngredient.ID)
		require.ErrorContains(t, err, "Unauthorized")
	})
}
