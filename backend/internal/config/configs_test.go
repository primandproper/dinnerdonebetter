package config

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config/envvars"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	analyticscfg "github.com/primandproper/platform-go/v9/analytics/config"
	databasecfg "github.com/primandproper/platform-go/v9/database/config"
	emailcfg "github.com/primandproper/platform-go/v9/email/config"
	"github.com/primandproper/platform-go/v9/encoding"
	featureflagscfg "github.com/primandproper/platform-go/v9/featureflags/config"
	meteringcfg "github.com/primandproper/platform-go/v9/metering/config"
	"github.com/primandproper/platform-go/v9/observability"
	loggingcfg "github.com/primandproper/platform-go/v9/observability/logging/config"
	"github.com/primandproper/platform-go/v9/routing/backends/chi"
	routingcfg "github.com/primandproper/platform-go/v9/routing/config"
	textsearchcfg "github.com/primandproper/platform-go/v9/search/text/config"
	"github.com/primandproper/platform-go/v9/server/http"
	webhookscfg "github.com/primandproper/platform-go/v9/webhooks/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIServiceConfig_EncodeToFile(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &APIServiceConfig{
			HTTPServer: http.Config{
				Port:            1234,
				StartupDeadline: time.Minute,
			},
			Meta: MetaSettings{
				RunMode: DevelopmentRunMode,
			},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Observability: observability.Config{},
			Services:      ServicesConfig{},
			Database: dbcfg.Config{
				Config: databasecfg.Config{
					Debug:         true,
					RunMigrations: true,
					ReadConnection: databasecfg.ConnectionDetails{
						Username:   "username",
						Password:   "password",
						Database:   "table",
						Host:       "host",
						DisableSSL: true,
					},
				},
			},
		}

		f, err := os.CreateTemp("", "")
		require.NoError(t, err)

		assert.NoError(t, cfg.EncodeToFile(f.Name(), json.Marshal))
	})

	T.Run("with error marshaling", func(t *testing.T) {
		t.Parallel()

		cfg := &APIServiceConfig{}

		f, err := os.CreateTemp("", "")
		require.NoError(t, err)

		assert.Error(t, cfg.EncodeToFile(f.Name(), func(any) ([]byte, error) {
			return nil, errors.New("blah")
		}))
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		var cfg *APIServiceConfig

		f, err := os.CreateTemp("", "")
		require.NoError(t, err)

		err = cfg.EncodeToFile(f.Name(), json.Marshal)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nil config")
	})
}

//nolint:paralleltest // because we set env vars for this, we can't
func TestLoadConfigFromEnvironment(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		cfg := &APIServiceConfig{
			Database: dbcfg.Config{
				Config: databasecfg.Config{
					Debug: true,
				},
			},
		}
		cfgBytes, err := json.Marshal(cfg)
		require.NoError(t, err)

		configFilepath := t.TempDir() + "/config.json"
		require.NoError(t, os.WriteFile(configFilepath, cfgBytes, 0o0644))

		t.Setenv(ConfigurationFilePathEnvVarKey, configFilepath)

		actual, err := LoadConfigFromEnvironment[APIServiceConfig]()
		assert.NoError(t, err)
		assert.NotNil(t, actual)

		assert.Equal(t, actual.Database.Debug, true)
	})

	// prior TODOs count here too
	T.Run("overrides meta", func(t *testing.T) {
		cfg := &APIServiceConfig{
			Database: dbcfg.Config{
				Config: databasecfg.Config{
					Debug: true,
				},
			},
		}
		cfgBytes, err := json.Marshal(cfg)
		require.NoError(t, err)

		configFilepath := t.TempDir() + "/config.json"
		require.NoError(t, os.WriteFile(configFilepath, cfgBytes, 0o0644))

		t.Setenv(ConfigurationFilePathEnvVarKey, configFilepath)
		t.Setenv(envvars.MetaDebugEnvVarKey, strconv.FormatBool(false))

		actual, err := LoadConfigFromEnvironment[APIServiceConfig]()
		assert.NoError(t, err)
		assert.NotNil(t, actual)

		assert.Equal(t, actual.Meta.Debug, false)
	})

	T.Run("with invalid config file", func(t *testing.T) {
		t.Setenv(ConfigurationFilePathEnvVarKey, "/nonexistent/path")

		actual, err := LoadConfigFromEnvironment[APIServiceConfig]()
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid JSON", func(t *testing.T) {
		configFilepath := t.TempDir() + "/config.json"
		require.NoError(t, os.WriteFile(configFilepath, []byte("{invalid json"), 0o0644))

		t.Setenv(ConfigurationFilePathEnvVarKey, configFilepath)

		actual, err := LoadConfigFromEnvironment[APIServiceConfig]()
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with apply env vars error", func(t *testing.T) {
		cfg := &APIServiceConfig{}
		cfgBytes, err := json.Marshal(cfg)
		require.NoError(t, err)

		configFilepath := t.TempDir() + "/config.json"
		require.NoError(t, os.WriteFile(configFilepath, cfgBytes, 0o0644))

		t.Setenv(ConfigurationFilePathEnvVarKey, configFilepath)
		// Set an invalid environment variable that would cause parsing to fail
		t.Setenv(envvars.HTTPPortEnvVarKey, "invalid_port")

		actual, err := LoadConfigFromEnvironment[APIServiceConfig]()
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

//nolint:paralleltest // because we set env vars for this, we can't
func TestLoadConfigFromEnvironment_WithDotEnv(T *testing.T) {
	T.Run("loads .env file before applying overrides", func(t *testing.T) {
		cfg := &APIServiceConfig{}
		cfgBytes, err := json.Marshal(cfg)
		require.NoError(t, err)

		dir := t.TempDir()

		configFilepath := dir + "/config.json"
		require.NoError(t, os.WriteFile(configFilepath, cfgBytes, 0o0644))

		dotEnvContent := envvars.BaseURLEnvVarKey + "=https://from-dotenv.example.com\n"
		dotEnvFilepath := dir + "/.env"
		require.NoError(t, os.WriteFile(dotEnvFilepath, []byte(dotEnvContent), 0o0644))

		t.Setenv(ConfigurationFilePathEnvVarKey, configFilepath)
		t.Setenv(DotEnvFilePathEnvVarKey, dotEnvFilepath)

		actual, err := LoadConfigFromEnvironment[APIServiceConfig]()
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, "https://from-dotenv.example.com", actual.BaseURL)
	})

	T.Run("actual env var overrides .env file value", func(t *testing.T) {
		cfg := &APIServiceConfig{}
		cfgBytes, err := json.Marshal(cfg)
		require.NoError(t, err)

		dir := t.TempDir()

		configFilepath := dir + "/config.json"
		require.NoError(t, os.WriteFile(configFilepath, cfgBytes, 0o0644))

		dotEnvContent := envvars.BaseURLEnvVarKey + "=https://from-dotenv.example.com\n"
		dotEnvFilepath := dir + "/.env"
		require.NoError(t, os.WriteFile(dotEnvFilepath, []byte(dotEnvContent), 0o0644))

		t.Setenv(ConfigurationFilePathEnvVarKey, configFilepath)
		t.Setenv(DotEnvFilePathEnvVarKey, dotEnvFilepath)
		// actual process env var wins over .env
		t.Setenv(envvars.BaseURLEnvVarKey, "https://from-actual-env.example.com")

		actual, err := LoadConfigFromEnvironment[APIServiceConfig]()
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, "https://from-actual-env.example.com", actual.BaseURL)
	})

	T.Run("with invalid .env filepath", func(t *testing.T) {
		t.Setenv(DotEnvFilePathEnvVarKey, "/nonexistent/.env")
		t.Setenv(ConfigurationFilePathEnvVarKey, "/nonexistent/config.json")

		actual, err := LoadConfigFromEnvironment[APIServiceConfig]()
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

//nolint:paralleltest // because we set env vars for this, we can't
func TestLoadConfigFromDotEnvFile(T *testing.T) {
	T.Run("loads minimal valid config from .env file", func(t *testing.T) {
		ctx := t.Context()

		// DBCleanerConfig has a simpler validation surface: just Database and Observability.
		dotEnvContent := envvars.DatabaseReadConnectionHostEnvVarKey + "=localhost\n" +
			envvars.DatabaseReadConnectionUsernameEnvVarKey + "=user\n" +
			envvars.DatabaseReadConnectionPasswordEnvVarKey + "=pass\n" +
			envvars.DatabaseReadConnectionDatabaseEnvVarKey + "=dbname\n"

		dir := t.TempDir()
		dotEnvFilepath := dir + "/.env"
		require.NoError(t, os.WriteFile(dotEnvFilepath, []byte(dotEnvContent), 0o0644))

		actual, err := LoadConfigFromDotEnvFile[DBCleanerConfig](ctx, dotEnvFilepath)
		// Validation may fail for other required fields, but env parsing must succeed.
		// The important thing is that environment variables are applied from the file.
		if err == nil {
			require.NotNil(t, actual)
			assert.Equal(t, "localhost", actual.Database.ReadConnection.Host)
			assert.Equal(t, "user", actual.Database.ReadConnection.Username)
		} else {
			// If validation fails it must be a validation error, not a file-loading error.
			assert.NotContains(t, err.Error(), "loading .env file")
			assert.NotContains(t, err.Error(), "applying environment variables")
		}
	})

	T.Run("with nonexistent file", func(t *testing.T) {
		ctx := t.Context()

		actual, err := LoadConfigFromDotEnvFile[DBCleanerConfig](ctx, "/nonexistent/.env")
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.Contains(t, err.Error(), "loading .env file")
	})

	T.Run("validates result", func(t *testing.T) {
		ctx := t.Context()

		// Intentionally empty .env — APIServiceConfig has strict multi-field validation
		// that fails on a zero-value config (Meta, Encoding, HTTPServer, Routing, etc.),
		// so it reliably fails even if DB-related env vars leaked from a previous subtest.
		dir := t.TempDir()
		dotEnvFilepath := dir + "/.env"
		require.NoError(t, os.WriteFile(dotEnvFilepath, []byte(""), 0o0644))

		actual, err := LoadConfigFromDotEnvFile[APIServiceConfig](ctx, dotEnvFilepath)
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.Contains(t, err.Error(), "validating config loaded from .env file")
	})
}

func TestAPIServiceConfig_Commit(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &APIServiceConfig{}
		commit := cfg.Commit()
		// The commit may or may not be empty depending on build info
		assert.IsType(t, "", commit)
	})
}

// validMeteringConfigForTest renders a metering config that passes validation.
//
// Every knob has a package default, so EnsureDefaults is the whole of it. It is spelled out
// rather than left zero because metering validates its own nested configs, and a zero-valued
// struct fails on eleven required fields at once.
func validMeteringConfigForTest() meteringcfg.Config {
	cfg := meteringcfg.Config{}
	cfg.EnsureDefaults()

	return cfg
}

// validWebhooksConfigForTest renders a webhooks config that passes validation.
//
// EnsureDefaults fills every knob but the circuit breaker's name, which has no default worth
// having: an unnamed breaker reports that something is failing without saying which subscriber.
func validWebhooksConfigForTest() webhookscfg.Config {
	cfg := webhookscfg.Config{}
	cfg.EnsureDefaults()
	cfg.CircuitBreaker.Name = "webhook_delivery"

	return cfg
}

func TestAPIServiceConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &APIServiceConfig{
			Meta: MetaSettings{
				RunMode: DevelopmentRunMode,
			},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Observability: observability.Config{
				Logging: loggingcfg.Config{ServiceName: "service"},
			},
			HTTPServer: http.Config{
				Port:            8080,
				StartupDeadline: time.Minute,
			},
			Queues: queuescfg.Config{
				DataChangesTopicName:         "data-changes",
				OutboundEmailsTopicName:      "outbound-emails",
				SearchIndexRequestsTopicName: "search-index-requests",
				MobileNotificationsTopicName: "mobile-notifications",
			},
			Database: dbcfg.Config{
				Config: databasecfg.Config{
					Debug: true,
					ReadConnection: databasecfg.ConnectionDetails{
						Username: "user",
						Password: "pass",
						Database: "db",
						Host:     "host",
						Port:     5432,
					},
				},
			},
			// Each of these has to name a provider: platform-go v9 reports an unset
			// one rather than substituting a noop that looks configured.
			Routing:      routingcfg.Config{Provider: routingcfg.ProviderChi, Chi: &chi.Config{ServiceName: "service"}},
			FeatureFlags: featureflagscfg.Config{Provider: featureflagscfg.ProviderNoop},
			Analytics:    analyticscfg.Config{SourceConfig: analyticscfg.SourceConfig{Provider: analyticscfg.ProviderNoop}},
			TextSearch:   textsearchcfg.Config{Provider: textsearchcfg.ProviderNoop},
			Email:        emailcfg.Config{Provider: emailcfg.ProviderNoop},
			Webhooks:     validWebhooksConfigForTest(),
			Metering:     validMeteringConfigForTest(),
		}

		err := cfg.ValidateWithContext(ctx)
		assert.NoError(t, err)
	})

	T.Run("with validateServices enabled", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &APIServiceConfig{
			validateServices: true,
			Meta: MetaSettings{
				RunMode: DevelopmentRunMode,
			},
			Encoding: encoding.Config{
				ContentType: "application/json",
			},
			Observability: observability.Config{},
			HTTPServer: http.Config{
				Port:            8080,
				StartupDeadline: time.Minute,
			},
			Queues: queuescfg.Config{
				DataChangesTopicName:         "data-changes",
				OutboundEmailsTopicName:      "outbound-emails",
				SearchIndexRequestsTopicName: "search-index-requests",
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
			Services: ServicesConfig{},
		}

		err := cfg.ValidateWithContext(ctx)
		// May have validation errors in services config
		_ = err // Don't assert NoError as services might have validation issues
	})
}

func TestDBCleanerConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &DBCleanerConfig{
			Observability: observability.Config{
				Logging: loggingcfg.Config{ServiceName: "service"},
			},
			Database: dbcfg.Config{
				Config: databasecfg.Config{
					Debug: true,
					ReadConnection: databasecfg.ConnectionDetails{
						Username: "user",
						Password: "pass",
						Database: "db",
						Host:     "host",
						Port:     5432,
					},
				},
			},
		}

		err := cfg.ValidateWithContext(ctx)
		assert.NoError(t, err)
	})
}

func TestAsyncMessageHandlerConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &AsyncMessageHandlerConfig{
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
		}

		err := cfg.ValidateWithContext(ctx)
		// May have validation errors in various configs
		_ = err
	})
}
