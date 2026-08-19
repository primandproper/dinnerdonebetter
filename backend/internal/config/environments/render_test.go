package environments

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	platformconfig "github.com/primandproper/platform-go/v11/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mcpServiceName is the observability service name Render gives the MCP server. It is spelled
// out rather than imported because internal/config keeps it unexported, and a test asserting on
// what landed in a file should be reading the file's own vocabulary anyway.
const mcpServiceName = "dinner_done_better_mcp_server"

// configLoaders maps each file Render writes, when the config set overrides no paths, to the
// loader the process that reads it starts up with. Rendering is only worth anything if what
// comes out is a file that loader accepts — a property to round-trip rather than to assert
// twice in two independently written functions.
func configLoaders() map[string]func(string) (any, error) {
	return map[string]func(string) (any, error){
		"api_service_config.json": func(p string) (any, error) {
			return config.LoadConfigFromPath[config.APIServiceConfig](p)
		},
		"job_db_cleaner_config.json": func(p string) (any, error) {
			return config.LoadConfigFromPath[config.DBCleanerConfig](p)
		},
		"scheduler_config.json": func(p string) (any, error) {
			return config.LoadConfigFromPath[config.SchedulerConfig](p)
		},
		"async_message_handler_config.json": func(p string) (any, error) {
			return config.LoadConfigFromPath[config.AsyncMessageHandlerConfig](p)
		},
		"job_email_deliverability_test_config.json": func(p string) (any, error) {
			return config.LoadConfigFromPath[config.EmailDeliverabilityTestConfig](p)
		},
		"mcp_server_config.json": func(p string) (any, error) {
			return config.LoadConfigFromPath[config.MCPServiceConfig](p)
		},
	}
}

func TestEnvironmentConfigSet_Render(T *testing.T) {
	T.Parallel()

	for name, build := range map[string]func() *config.APIServiceConfig{
		"localdev":          BuildLocalDevConfig,
		"integration tests": BuildIntegrationTestsConfig,
		"prod":              BuildProdConfig,
	} {
		T.Run(name+" renders files the loader reads back", func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			tmpDir := t.TempDir()

			configSet := &config.EnvironmentConfigSet{RootConfig: build()}
			require.NoError(t, configSet.Render(ctx, tmpDir))

			for fileName, load := range configLoaders() {
				loaded, err := load(filepath.Join(tmpDir, fileName))
				require.NoError(t, err, "loading %s", fileName)
				require.NotNil(t, loaded, "loading %s", fileName)
				assert.NoError(t, platformconfig.Validate(ctx, loaded), "validating %s", fileName)
			}
		})
	}

	T.Run("apple app site association is rendered for the API but not the MCP server", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		rootConfig := BuildProdConfig()
		require.NotNil(t, rootConfig.HTTPServer.AppleAppSiteAssociation)

		configSet := &config.EnvironmentConfigSet{RootConfig: rootConfig}
		require.NoError(t, configSet.Render(context.Background(), tmpDir))

		apiConfig, err := os.ReadFile(filepath.Join(tmpDir, "api_service_config.json"))
		require.NoError(t, err)
		assert.Contains(t, string(apiConfig), "appleAppSiteAssociation")

		// The MCP server copies the API's HTTP server config, so the association has to be
		// stripped from that copy: it names a domain no Universal Link points at.
		mcpConfig, err := os.ReadFile(filepath.Join(tmpDir, "mcp_server_config.json"))
		require.NoError(t, err)
		assert.NotContains(t, string(mcpConfig), "appleAppSiteAssociation")

		// Rendering the MCP config must not clear the association on the config it copied.
		assert.NotNil(t, rootConfig.HTTPServer.AppleAppSiteAssociation)
	})

	T.Run("rendering the MCP config leaves the API's chi config alone", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		// Prod, because its chi config sets EnableCORSForLocalhost false — the value the MCP
		// server's copy has to not write back.
		rootConfig := BuildProdConfig()
		require.NotNil(t, rootConfig.Routing.Chi)
		require.False(t, rootConfig.Routing.Chi.EnableCORSForLocalhost)

		configSet := &config.EnvironmentConfigSet{RootConfig: rootConfig}
		require.NoError(t, configSet.Render(context.Background(), tmpDir))

		// routingcfg.Config holds a *chi.Config, so the MCP server's copy of the routing config
		// addresses the same struct. Without a clone its two writes land here as well, and the
		// damage outlives the call: the one config set rendered twice shares this RootConfig.
		assert.Equal(t, otelServiceName, rootConfig.Routing.Chi.ServiceName)
		assert.False(t, rootConfig.Routing.Chi.EnableCORSForLocalhost)

		// ... while the MCP server's own file still gets both.
		mcpConfig, err := os.ReadFile(filepath.Join(tmpDir, "mcp_server_config.json"))
		require.NoError(t, err)
		assert.Contains(t, string(mcpConfig), mcpServiceName)
		assert.Contains(t, string(mcpConfig), "enableCORSForLocalhost")
	})

	T.Run("with custom file paths", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		configSet := &config.EnvironmentConfigSet{
			RootConfig:           BuildLocalDevConfig(),
			APIServiceConfigPath: "custom_api.json",
			DBCleanerConfigPath:  "custom_db_cleaner.json",
		}

		require.NoError(t, configSet.Render(context.Background(), tmpDir))

		assert.FileExists(t, filepath.Join(tmpDir, "custom_api.json"))
		assert.FileExists(t, filepath.Join(tmpDir, "custom_db_cleaner.json"))
		assert.NoFileExists(t, filepath.Join(tmpDir, "api_service_config.json"))
		assert.NoFileExists(t, filepath.Join(tmpDir, "job_db_cleaner_config.json"))
	})

	T.Run("an invalid config set writes nothing at all", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		// A blank username for the MCP server's database user, which leaves the root config
		// valid and the last of the six derived configs invalid. Every file is rendered by a
		// separate RenderJSONFiles call, and no call can see the ones after it, so without a
		// validation pass over the whole set first this would leave five updated files beside
		// one stale one.
		configSet := &config.EnvironmentConfigSet{
			RootConfig:           BuildProdConfig(),
			ServiceDatabaseUsers: map[string]string{mcpServiceName: ""},
		}

		err := configSet.Render(context.Background(), tmpDir)
		// Config 5 is the MCP server's, the last of the six: the failure is genuinely at the
		// end of the set rather than at the root config the other five are derived from.
		require.ErrorContains(t, err, "validating config 5")

		entries, readErr := os.ReadDir(tmpDir)
		require.NoError(t, readErr)
		assert.Empty(t, entries)
	})
}
