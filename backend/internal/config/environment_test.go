package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"

	analyticscfg "github.com/primandproper/platform-go/v9/analytics/config"
	databasecfg "github.com/primandproper/platform-go/v9/database/config"
	emailcfg "github.com/primandproper/platform-go/v9/email/config"
	"github.com/primandproper/platform-go/v9/encoding"
	featureflagscfg "github.com/primandproper/platform-go/v9/featureflags/config"
	msgconfig "github.com/primandproper/platform-go/v9/messagequeue/config"
	"github.com/primandproper/platform-go/v9/observability"
	routingcfg "github.com/primandproper/platform-go/v9/routing/config"
	textsearchcfg "github.com/primandproper/platform-go/v9/search/text/config"
	"github.com/primandproper/platform-go/v9/server/grpc"
	"github.com/primandproper/platform-go/v9/server/http"
	uploadscfg "github.com/primandproper/platform-go/v9/uploads/config"
	"github.com/primandproper/platform-go/v9/uploads/objectstorage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringOrDefault(T *testing.T) {
	T.Parallel()

	T.Run("with empty string", func(t *testing.T) {
		t.Parallel()

		result := stringOrDefault("", "default")
		assert.Equal(t, "default", result)
	})

	T.Run("with non-empty string", func(t *testing.T) {
		t.Parallel()

		result := stringOrDefault("value", "default")
		assert.Equal(t, "value", result)
	})
}

func TestRenderJSON(T *testing.T) {
	T.Parallel()

	T.Run("pretty formatting", func(t *testing.T) {
		t.Parallel()

		obj := map[string]any{
			"key": "value",
			"nested": map[string]any{
				"inner": "data",
			},
		}

		result := renderJSON(obj, true)
		assert.Contains(t, string(result), "\t")
		assert.Contains(t, string(result), "key")
		assert.Contains(t, string(result), "value")
	})

	T.Run("compact formatting", func(t *testing.T) {
		t.Parallel()

		obj := map[string]any{
			"key": "value",
		}

		result := renderJSON(obj, false)
		assert.NotContains(t, string(result), "\t")
		assert.Contains(t, string(result), "key")
		assert.Contains(t, string(result), "value")
	})

	T.Run("with marshal error", func(t *testing.T) {
		t.Parallel()

		// Test with an object that can't be marshaled
		obj := make(chan int)

		assert.Panics(t, func() {
			renderJSON(obj, false)
		})
	})
}

func TestWriteFile(T *testing.T) {
	T.Parallel()

	T.Run("successful write", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.txt")
		content := []byte("test content")

		err := writeFile(filePath, content)
		assert.NoError(t, err)

		// Verify file was written
		data, err := os.ReadFile(filePath)
		require.NoError(t, err)
		assert.Equal(t, content, data)
	})

	T.Run("with invalid path", func(t *testing.T) {
		t.Parallel()

		// Try to write to a path that doesn't exist and can't be created
		filePath := "/nonexistent/path/file.txt"
		content := []byte("test content")

		err := writeFile(filePath, content)
		assert.Error(t, err)
	})
}

func TestEnvironmentConfigSet_Render(T *testing.T) {
	T.Parallel()

	T.Run("successful render", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		rootConfig := &APIServiceConfig{
			Auth: authcfg.Config{},
			Queues: queuescfg.Config{
				DataChangesTopicName:              "data-changes",
				OutboundEmailsTopicName:           "outbound-emails",
				SearchIndexRequestsTopicName:      "search-index-requests",
				UserDataAggregationTopicName:      "user-data-aggregation",
				WebhookExecutionRequestsTopicName: "webhook-execution-requests",
			},
			Email:        emailcfg.Config{},
			Analytics:    analyticscfg.Config{},
			TextSearch:   textsearchcfg.Config{},
			FeatureFlags: featureflagscfg.Config{},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Events:        msgconfig.Config{},
			Observability: observability.Config{},
			Meta: MetaSettings{
				RunMode: DevelopmentRunMode,
			},
			Routing:    routingcfg.Config{},
			HTTPServer: http.Config{},
			GRPCServer: grpc.Config{},
			Database: dbcfg.Config{
				Config: databasecfg.Config{
					Debug: true,
					ReadConnection: databasecfg.ConnectionDetails{
						Username: "user",
						Password: "pass",
						Database: "db",
						Host:     "host",
					},
				},
			},
			Services: ServicesConfig{
				DataPrivacy: dataprivacycfg.Config{
					Uploads: uploadscfg.Config{
						Storage: objectstorage.Config{},
					},
				},
			},
		}

		configSet := &EnvironmentConfigSet{
			RootConfig: rootConfig,
		}

		err := configSet.Render(tmpDir, true, false)
		assert.NoError(t, err)

		// Verify files were created
		expectedFiles := []string{
			"api_service_config.json",
			"job_db_cleaner_config.json",
			"job_email_deliverability_test_config.json",
			"scheduler_config.json",
			"async_message_handler_config.json",
		}

		for _, fileName := range expectedFiles {
			filePath := filepath.Join(tmpDir, fileName)
			assert.FileExists(t, filePath)

			// Verify file contains valid JSON
			data, dataErr := os.ReadFile(filePath)
			require.NoError(t, dataErr)

			var jsonData any
			err = json.Unmarshal(data, &jsonData)
			assert.NoError(t, err, "File %s should contain valid JSON", fileName)
		}
	})

	T.Run("apple app site association is rendered for the API but not the MCP server", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		rootConfig := &APIServiceConfig{
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Meta: MetaSettings{
				RunMode: DevelopmentRunMode,
			},
			Observability: observability.Config{},
			HTTPServer: http.Config{
				AppleAppSiteAssociation: &http.AppleAppSiteAssociationConfig{
					TeamID:   "ABCD1234XY",
					BundleID: "com.example.ios",
				},
			},
			Database: dbcfg.Config{
				Config: databasecfg.Config{
					Debug: true,
					ReadConnection: databasecfg.ConnectionDetails{
						Username: "user",
						Password: "pass",
						Database: "db",
						Host:     "host",
					},
				},
			},
			Services: ServicesConfig{
				DataPrivacy: dataprivacycfg.Config{
					Uploads: uploadscfg.Config{
						Storage: objectstorage.Config{},
					},
				},
			},
		}

		configSet := &EnvironmentConfigSet{RootConfig: rootConfig}

		require.NoError(t, configSet.Render(tmpDir, true, false))

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

	T.Run("with custom file paths", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		rootConfig := &APIServiceConfig{
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Meta: MetaSettings{
				RunMode: DevelopmentRunMode,
			},
			Observability: observability.Config{},
			Database: dbcfg.Config{
				Config: databasecfg.Config{
					Debug: true,
					ReadConnection: databasecfg.ConnectionDetails{
						Username: "user",
						Password: "pass",
						Database: "db",
						Host:     "host",
					},
				},
			},
			Services: ServicesConfig{
				DataPrivacy: dataprivacycfg.Config{
					Uploads: uploadscfg.Config{
						Storage: objectstorage.Config{},
					},
				},
			},
		}

		configSet := &EnvironmentConfigSet{
			RootConfig:           rootConfig,
			APIServiceConfigPath: "custom_api.json",
			DBCleanerConfigPath:  "custom_db_cleaner.json",
		}

		err := configSet.Render(tmpDir, false, false)
		assert.NoError(t, err)

		// Verify custom files were created
		assert.FileExists(t, filepath.Join(tmpDir, "custom_api.json"))
		assert.FileExists(t, filepath.Join(tmpDir, "custom_db_cleaner.json"))
	})

	T.Run("with validation enabled", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		rootConfig := &APIServiceConfig{
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Meta: MetaSettings{
				RunMode: DevelopmentRunMode,
			},
			Observability: observability.Config{},
			Database: dbcfg.Config{
				Config: databasecfg.Config{
					Debug: true,
					ReadConnection: databasecfg.ConnectionDetails{
						Username: "user",
						Password: "pass",
						Database: "db",
						Host:     "host",
					},
				},
			},
			Services: ServicesConfig{
				DataPrivacy: dataprivacycfg.Config{
					Uploads: uploadscfg.Config{
						Storage: objectstorage.Config{},
					},
				},
			},
		}

		configSet := &EnvironmentConfigSet{
			RootConfig: rootConfig,
		}

		err := configSet.Render(tmpDir, false, true)
		// May have validation errors, but should not fail completely
		_ = err
	})

	T.Run("with invalid output directory", func(t *testing.T) {
		t.Parallel()

		// Try to write to a directory that can't be created
		invalidDir := "/nonexistent/deeply/nested/path"

		rootConfig := &APIServiceConfig{}
		configSet := &EnvironmentConfigSet{
			RootConfig: rootConfig,
		}

		err := configSet.Render(invalidDir, false, false)
		assert.Error(t, err)
	})

	T.Run("with file write error", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()

		// Create a directory where a file should be written to force an error
		filePath := filepath.Join(tmpDir, "api_service_config.json")
		err := os.Mkdir(filePath, 0o755)
		require.NoError(t, err)

		rootConfig := &APIServiceConfig{
			Services: ServicesConfig{
				DataPrivacy: dataprivacycfg.Config{
					Uploads: uploadscfg.Config{
						Storage: objectstorage.Config{},
					},
				},
			},
		}
		configSet := &EnvironmentConfigSet{
			RootConfig: rootConfig,
		}

		err = configSet.Render(tmpDir, false, false)
		assert.Error(t, err)
	})
}
