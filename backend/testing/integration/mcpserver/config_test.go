package integration

import (
	"fmt"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	ddboauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v11/authentication/oauth2server"
	oauth2servercfg "github.com/primandproper/platform-go/v11/authentication/oauth2server/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPServer_Configuration(T *testing.T) {
	T.Parallel()

	T.Run("advertises the base URL it was started with", func(t *testing.T) {
		t.Parallel()

		authorizationServer := primary.authorizationServerMetadata(t)

		// Every endpoint in the RFC 8414 document is derived from the issuer, so an
		// issuer resolved from the wrong place is a document a client can follow all the
		// way to a host that is not this one.
		assert.Equal(t, fleetBaseURL, authorizationServer.Issuer)
		assert.Equal(t, fleetBaseURL+oauth2server.PathAuthorize, authorizationServer.AuthorizationEndpoint)
		assert.Equal(t, fleetBaseURL+oauth2server.PathToken, authorizationServer.TokenEndpoint)
		assert.Equal(t, fleetBaseURL+oauth2server.PathRegister, authorizationServer.RegistrationEndpoint)

		// PKCE is not optional here, and this is where a client discovers that.
		assert.Equal(t, []string{oauth2server.CodeChallengeMethodS256}, authorizationServer.CodeChallengeMethodsSupported)

		resource := primary.protectedResourceMetadata(t)
		assert.Equal(t, fleetBaseURL, resource.Resource)
		assert.Contains(t, resource.AuthorizationServers, fleetBaseURL)
	})

	T.Run("prefers a configured issuer over the base URL", func(t *testing.T) {
		t.Parallel()

		// A deployment fronting this server somewhere else is what the issuer setting is
		// for, and the rule it encodes is that the configured value wins. Getting the
		// precedence backwards is invisible until a client follows the document: the
		// server serves every request it is sent and advertises an address it is not at.
		const configuredIssuer = "https://mcp.example.com"

		fronted := startInstanceForTest(t, fleetBaseURL, func(cfg *config.MCPServiceConfig) {
			cfg.OAuth2.Issuer = configuredIssuer
		})

		authorizationServer := fronted.authorizationServerMetadata(t)
		assert.Equal(t, configuredIssuer, authorizationServer.Issuer)
		assert.Equal(t, configuredIssuer+oauth2server.PathToken, authorizationServer.TokenEndpoint)

		// The resource document is the protected resource's, not the authorization
		// server's, and it still names the base URL this replica serves.
		assert.Equal(t, fleetBaseURL, fronted.protectedResourceMetadata(t).Resource)
	})

	T.Run("stores its records under the prefix the rendered config names", func(t *testing.T) {
		t.Parallel()

		// Read off the file rather than off a config built in Go. This stanza is written
		// by internal/config/environments and read by nothing else at runtime, so a
		// prefix that drifted from the one migration 33 rendered its DDL with would be a
		// server that comes up clean and cannot find a table.
		require.Equal(t, oauth2servercfg.ProviderDatabase, mcpServiceConfig.OAuth2.Provider)
		require.Equal(t, ddboauth.TablePrefix, mcpServiceConfig.OAuth2.Database.TablePrefix)

		registration := primary.registerClient(t)

		// The table name is built from the same constant the migration renders its DDL
		// from, which is the point: this asserts the row landed where a reader looking
		// only at the config would expect to find it.
		var found int
		require.NoError(t, rawDB.QueryRowContext(t.Context(),
			fmt.Sprintf(`SELECT COUNT(*) FROM %s_oauth2_clients WHERE id = $1`, ddboauth.TablePrefix),
			registration.ClientID).Scan(&found))

		assert.Equal(t, 1, found)
	})
}
