package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	analyticscfg "github.com/primandproper/platform-go/v10/analytics/config"
	platformconfig "github.com/primandproper/platform-go/v10/config"
	emailcfg "github.com/primandproper/platform-go/v10/email/config"
	"github.com/primandproper/platform-go/v10/encoding"
	featureflagscfg "github.com/primandproper/platform-go/v10/featureflags/config"
	httpclientcfg "github.com/primandproper/platform-go/v10/httpclient"
	idempotencycfg "github.com/primandproper/platform-go/v10/idempotency/config"
	"github.com/primandproper/platform-go/v10/jobs"
	msgconfig "github.com/primandproper/platform-go/v10/messagequeue/config"
	meteringcfg "github.com/primandproper/platform-go/v10/metering/config"
	notificationscfg "github.com/primandproper/platform-go/v10/notifications/mobile/config"
	"github.com/primandproper/platform-go/v10/observability"
	operationscfg "github.com/primandproper/platform-go/v10/operations/config"
	routingcfg "github.com/primandproper/platform-go/v10/routing/config"
	textsearchcfg "github.com/primandproper/platform-go/v10/search/text/config"
	"github.com/primandproper/platform-go/v10/server/grpc"
	"github.com/primandproper/platform-go/v10/server/http"
	webhookscfg "github.com/primandproper/platform-go/v10/webhooks/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
	"github.com/joho/godotenv"
)

const (
	// DevelopmentRunMode is the run mode for a development environment.
	DevelopmentRunMode runMode = "development"
	// TestingRunMode is the run mode for a testing environment.
	TestingRunMode runMode = "testing"
	// ProductionRunMode is the run mode for a production environment.
	ProductionRunMode runMode = "production"

	EnvVarPrefix = branding.EnvVarPrefix

	// ConfigurationFilePathEnvVarKey is the env var key we use to indicate where the config file is located.
	ConfigurationFilePathEnvVarKey = "CONFIGURATION_FILEPATH"
	// DotEnvFilePathEnvVarKey is the env var key we use to indicate where the .env file is located.
	// When set, the .env file is loaded before environment variable overrides are applied,
	// meaning .env values override JSON config but are themselves overridden by actual process env vars.
	DotEnvFilePathEnvVarKey = "DOTENV_FILEPATH"
)

type (
	// runMode describes what method of operation the server is under.
	runMode string

	// CloserFunc calls all io.Closers in the service.
	CloserFunc func()

	configurations interface {
		APIServiceConfig |
			DBCleanerConfig |
			SchedulerConfig |
			AsyncMessageHandlerConfig |
			EmailDeliverabilityTestConfig |
			MCPServiceConfig
	}

	// APIServiceConfig configures an instance of the service. It is composed of all the other setting structs.
	APIServiceConfig struct {
		_                 struct{}                `json:"-"`
		HTTPClient        *httpclientcfg.Config   `envPrefix:"HTTP_CLIENT_"        json:"httpClient,omitempty"`
		Queues            queuescfg.Config        `envPrefix:"QUEUES_"             json:"queues,omitzero"`
		Routing           routingcfg.Config       `envPrefix:"ROUTING_"            json:"routing,omitzero"`
		PushNotifications notificationscfg.Config `envPrefix:"PUSH_NOTIFICATIONS_" json:"pushNotifications,omitzero"`
		Encoding          encoding.Config         `envPrefix:"ENCODING_"           json:"encoding,omitzero"`
		BaseURL           string                  `env:"BASE_URL"                  json:"baseURL,omitempty"`
		Events            msgconfig.Config        `envPrefix:"EVENTS_"             json:"events,omitzero"`
		Observability     observability.Config    `envPrefix:"OBSERVABILITY_"      json:"observability,omitzero"`
		GRPCServer        grpc.Config             `envPrefix:"GRPC_"               json:"grpc,omitzero"`
		Meta              MetaSettings            `envPrefix:"META_"               json:"meta,omitzero"`
		Email             emailcfg.Config         `envPrefix:"EMAIL_"              json:"email,omitzero"`
		Analytics         analyticscfg.Config     `envPrefix:"ANALYTICS_"          json:"analytics,omitzero"`
		FeatureFlags      featureflagscfg.Config  `envPrefix:"FEATURE_FLAGS_"      json:"featureFlags,omitzero"`
		TextSearch        textsearchcfg.Config    `envPrefix:"SEARCH_"             json:"search,omitzero"`
		Auth              authcfg.Config          `envPrefix:"AUTH_"               json:"auth,omitzero"`
		Database          dbcfg.Config            `envPrefix:"DATABASE_"           json:"database,omitzero"`
		HTTPServer        http.Config             `envPrefix:"HTTP_"               json:"http,omitzero"`

		// Idempotency guards the mutations where running the work twice costs real money.
		// A client that never sees a response and retries is indistinguishable from a
		// deliberate second purchase unless it supplies a key, so this is opt-in per call:
		// a request without the idempotency-key metadata passes through untouched.
		// Not omitzero, and Enabled is not omitempty: a deployment with the interceptor
		// off is exactly the zero value, so omitting it would erase the distinction
		// between "deliberately off" and "nobody configured this" — the silent failure
		// the Enabled comment above exists to prevent.
		Idempotency IdempotencyConfig `envPrefix:"IDEMPOTENCY_" json:"idempotency"`

		Services ServicesConfig `envPrefix:"SERVICE_" json:"services,omitzero"`

		// Metering counts what an account consumes. The API server holds only the ingest
		// half of it — the flusher that posts usage to a billing provider runs in the
		// scheduler — but both read this same struct, so the tables one writes are by
		// construction the tables the other flushes.
		Metering meteringcfg.Config `envPrefix:"METERING_" json:"metering,omitzero"`

		// Operations is the durable record of tracked work. The API server holds the
		// enqueue-and-read half of it — data privacy requests are submitted as operations
		// here and polled for progress — while the worker that runs them lives in the
		// scheduler. Both read this same struct, so the table one writes is by construction
		// the table the other claims from.
		Operations operationscfg.Config `envPrefix:"OPERATIONS_" json:"operations,omitzero"`

		// Webhooks configures the outbound webhook tables this service writes into. Only the
		// write side lives here: dispatch rows are written inside the transactions that
		// caused them, and the worker that delivers them runs in the scheduler.
		Webhooks webhookscfg.Config `envPrefix:"WEBHOOKS_" json:"webhooks,omitzero"`

		validateServices bool
	}

	// IdempotencyConfig gates the payment idempotency interceptor on having somewhere real to
	// keep its records.
	IdempotencyConfig struct {
		_ struct{} `json:"-"`

		Manager idempotencycfg.Config `envPrefix:"MANAGER_" json:"manager,omitzero"`

		// Enabled installs the interceptor.
		//
		// It is a flag rather than a derived default because the failure mode of getting
		// this wrong is silent. The memory cache provider is per-process, so with several
		// replicas one replica's records are invisible to the others: a retry that lands
		// elsewhere re-executes, and two concurrent requests can both claim the same key.
		// That configuration looks like protection and provides none, which is worse than
		// no interceptor at all — at least an absent interceptor is legible.
		//
		// Turn it on only where Manager.Cache names a shared store.
		Enabled bool `env:"ENABLED" json:"enabled"`
	}

	// DBCleanerConfig configures an instance of the database cleaner job.
	DBCleanerConfig struct {
		_ struct{} `json:"-"`

		Observability observability.Config `envPrefix:"OBSERVABILITY_" json:"observability,omitzero"`

		Database dbcfg.Config `envPrefix:"DATABASE_" json:"database,omitzero"`
	}

	// AsyncMessageHandlerConfig configures an instance of the search data index scheduler job.
	AsyncMessageHandlerConfig struct {
		_          struct{}              `json:"-"`
		HTTPClient *httpclientcfg.Config `envPrefix:"HTTP_CLIENT_" json:"httpClient,omitempty"`
		Queues     queuescfg.Config      `envPrefix:"QUEUES_"      json:"queues,omitzero"`

		PushNotifications notificationscfg.Config `envPrefix:"PUSH_NOTIFICATIONS_" json:"pushNotifications,omitzero"`
		Encoding          encoding.Config         `envPrefix:"ENCODING_"           json:"encoding,omitzero"`
		BaseURL           string                  `env:"BASE_URL"                  json:"baseURL,omitempty"`
		Events            msgconfig.Config        `envPrefix:"EVENTS_"             json:"events,omitzero"`
		Observability     observability.Config    `envPrefix:"OBSERVABILITY_"      json:"observability,omitzero"`
		Email             emailcfg.Config         `envPrefix:"EMAIL_"              json:"email,omitzero"`
		Analytics         analyticscfg.Config     `envPrefix:"ANALYTICS_"          json:"analytics,omitzero"`
		Search            textsearchcfg.Config    `envPrefix:"SEARCH_"             json:"search,omitzero"`
		Database          dbcfg.Config            `envPrefix:"DATABASE_"           json:"database,omitzero"`
		Pools             WorkerPoolsConfig       `envPrefix:"POOLS_"              json:"pools,omitzero"`
	}

	// WorkerPoolsConfig configures the jobs.Pool draining each queue topic. Topics are not
	// repeated here: each pool's Topic is filled in from Queues at construction, so a topic
	// name lives in exactly one place.
	WorkerPoolsConfig struct {
		_ struct{} `json:"-"`

		// DeadLetterTopicName is where a message goes once it has exhausted its attempts.
		// Without it jobs.Pool has no terminal destination and silently drops exhausted
		// messages, so it is required rather than defaulted.
		DeadLetterTopicName string `env:"DEAD_LETTER_TOPIC_NAME" json:"deadLetterTopicName,omitempty"`

		DataChanges         jobs.PoolConfig `envPrefix:"DATA_CHANGES_"          json:"dataChanges,omitzero"`
		OutboundEmails      jobs.PoolConfig `envPrefix:"OUTBOUND_EMAILS_"       json:"outboundEmails,omitzero"`
		SearchIndexRequests jobs.PoolConfig `envPrefix:"SEARCH_INDEX_REQUESTS_" json:"searchIndexRequests,omitzero"`
		MobileNotifications jobs.PoolConfig `envPrefix:"MOBILE_NOTIFICATIONS_"  json:"mobileNotifications,omitzero"`
	}

	// EmailDeliverabilityTestConfig configures the email deliverability test cron job.
	EmailDeliverabilityTestConfig struct {
		_                     struct{}              `json:"-"`
		HTTPClient            *httpclientcfg.Config `envPrefix:"HTTP_CLIENT_"      json:"httpClient,omitempty"`
		RecipientEmailAddress string                `env:"RECIPIENT_EMAIL_ADDRESS" json:"recipientEmailAddress,omitempty"`
		ServiceEnvironment    string                `env:"SERVICE_ENVIRONMENT"     json:"serviceEnvironment,omitempty"`
		Observability         observability.Config  `envPrefix:"OBSERVABILITY_"    json:"observability,omitzero"`
		Email                 emailcfg.Config       `envPrefix:"EMAIL_"            json:"email,omitzero"`
	}

	// MCPServiceConfig configures an instance of the service. It is composed of all the other setting structs.
	MCPServiceConfig struct {
		_             struct{}             `json:"-"`
		Database      dbcfg.Config         `envPrefix:"DATABASE_"      json:"database,omitzero"`
		Routing       routingcfg.Config    `envPrefix:"ROUTING_"       json:"routing,omitzero"`
		Observability observability.Config `envPrefix:"OBSERVABILITY_" json:"observability,omitzero"`
		Meta          MetaSettings         `envPrefix:"META_"          json:"meta,omitzero"`
		HTTPServer    http.Config          `envPrefix:"HTTP_"          json:"http,omitzero"`
	}
)

// EncodeToFile renders your config to a file given your favorite encoder.
func (cfg *APIServiceConfig) EncodeToFile(path string, marshaller func(v any) ([]byte, error)) error {
	if cfg == nil {
		return errors.New("nil config")
	}

	byteSlice, err := marshaller(*cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, byteSlice, 0o600) //nolint:gosec // G703: path from caller; caller must pass trusted path
}

func (cfg *APIServiceConfig) Commit() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for i := range info.Settings {
			if info.Settings[i].Key == "vcs.revision" {
				return info.Settings[i].Value
			}
		}
	}

	return ""
}

var _ validation.ValidatableWithContext = (*APIServiceConfig)(nil)

// ValidateWithContext validates a APIServiceConfig struct.
func (cfg *APIServiceConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Routing":       cfg.Routing.ValidateWithContext,
		"Meta":          cfg.Meta.ValidateWithContext,
		"Queues":        cfg.Queues.ValidateWithContext,
		"Encoding":      cfg.Encoding.ValidateWithContext,
		"Analytics":     cfg.Analytics.ValidateWithContext,
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
		"HTTPServer":    cfg.HTTPServer.ValidateWithContext,
		"Email":         cfg.Email.ValidateWithContext,
		"FeatureFlags":  cfg.FeatureFlags.ValidateWithContext,
		"TextSearch":    cfg.TextSearch.ValidateWithContext,
		"Idempotency":   cfg.Idempotency.ValidateWithContext,
		"Webhooks":      cfg.Webhooks.ValidateWithContext,
		"Metering":      cfg.Metering.ValidateWithContext,
		"Operations":    cfg.Operations.ValidateWithContext,
		// no "Events" here, that's a collection of publisher/subscriber configs that can each optionally be setup
	}

	if cfg.validateServices {
		validators["Services"] = cfg.Services.ValidateWithContext
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*DBCleanerConfig)(nil)

// ValidateWithContext validates a DBCleanerConfig struct.
func (cfg *DBCleanerConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*IdempotencyConfig)(nil)

// ValidateWithContext validates an IdempotencyConfig struct. A disabled interceptor is never
// constructed, so its manager config is not validated.
func (cfg *IdempotencyConfig) ValidateWithContext(ctx context.Context) error {
	if !cfg.Enabled {
		return nil
	}

	return cfg.Manager.ValidateWithContext(ctx)
}

var _ validation.ValidatableWithContext = (*WorkerPoolsConfig)(nil)

// ValidateWithContext validates a WorkerPoolsConfig struct.
//
// The individual jobs.PoolConfig values are deliberately not validated here: their Topic is
// filled in from the queues config at construction, so they are still empty at this point and
// jobs.NewPool validates each one once it is complete.
func (cfg *WorkerPoolsConfig) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.DeadLetterTopicName, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*AsyncMessageHandlerConfig)(nil)

// ValidateWithContext validates a AsyncMessageHandlerConfig struct.
func (cfg *AsyncMessageHandlerConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Queues":        cfg.Queues.ValidateWithContext,
		"Analytics":     cfg.Analytics.ValidateWithContext,
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
		"Email":         cfg.Email.ValidateWithContext,
		"TextSearch":    cfg.Search.ValidateWithContext,
		"Pools":         cfg.Pools.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*EmailDeliverabilityTestConfig)(nil)

// ValidateWithContext validates an EmailDeliverabilityTestConfig struct.
func (cfg *EmailDeliverabilityTestConfig) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		cfg,
		validation.Field(&cfg.Observability, validation.Required),
		validation.Field(&cfg.Email, validation.Required),
		validation.Field(&cfg.RecipientEmailAddress, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*MCPServiceConfig)(nil)

// ValidateWithContext validates a MCPServiceConfig struct.
func (cfg *MCPServiceConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Database":      cfg.Database.ValidateWithContext,
		"Meta":          cfg.Meta.ValidateWithContext,
		"Observability": cfg.Observability.ValidateWithContext,
		"Routing":       cfg.Routing.ValidateWithContext,
		"HTTPServer":    cfg.HTTPServer.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

// resolveDotEnvFilePathFromDir returns the path of the .env file that should be
// loaded for the current environment, searching under baseDir.
//
// Priority and selection rules:
//  1. If DOTENV_FILEPATH is set, return it as-is (explicit override wins).
//  2. Otherwise, derive the filename from META_RUN_MODE:
//     - "development" → .env.dev
//     - "testing"     → .env.testing
//     - anything else → .env  (covers "production" and the unset case)
//  3. If the derived file does not exist under baseDir, return "" so the
//     caller can skip loading without treating a missing file as an error.
func resolveDotEnvFilePathFromDir(baseDir string) string {
	if explicit := os.Getenv(DotEnvFilePathEnvVarKey); explicit != "" {
		return explicit
	}

	var filename string
	switch runMode(strings.TrimSpace(os.Getenv(EnvVarPrefix + "META_RUN_MODE"))) {
	case DevelopmentRunMode:
		filename = ".env.dev"
	case TestingRunMode:
		filename = ".env.testing"
	default: // ProductionRunMode or unset
		filename = ".env"
	}

	path := filepath.Join(baseDir, filename)
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return ""
	}
	return path
}

// resolveDotEnvFilePath is the production entry point for resolveDotEnvFilePathFromDir,
// using the process's current working directory as the base.
func resolveDotEnvFilePath() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return resolveDotEnvFilePathFromDir(dir)
}

func LoadConfigFromEnvironment[T configurations]() (*T, error) {
	// Resolve and load the appropriate .env file before applying env var overrides.
	// godotenv.Load does not override env vars already set in the process, so
	// priority order is: JSON config < .env file < actual process environment.
	if dotEnvPath := resolveDotEnvFilePath(); dotEnvPath != "" {
		if err := godotenv.Load(dotEnvPath); err != nil {
			return nil, fmt.Errorf("loading .env file: %w", err)
		}
	}

	configFilepath := os.Getenv(ConfigurationFilePathEnvVarKey)

	// Background, not a caller's context. v10's loader takes one so that config validation
	// can be cancelled, but this runs once at process startup before there is a request or a
	// server context to cancel against, and a config load that gave up halfway would leave
	// the process with no configuration rather than with less of it.
	cfg, err := platformconfig.LoadFromJSONFile[T](context.Background(), configFilepath, envVarOptions()...)
	if err != nil {
		return nil, fmt.Errorf("loading config from environment: %w", err)
	}

	return cfg, nil
}

func LoadConfigFromPath[T configurations](configurationFilepath string) (*T, error) {
	// Resolve and load the appropriate .env file before applying env var overrides.
	// godotenv.Load does not override env vars already set in the process, so
	// priority order is: JSON config < .env file < actual process environment.
	if dotEnvPath := resolveDotEnvFilePath(); dotEnvPath != "" {
		if err := godotenv.Load(dotEnvPath); err != nil {
			return nil, fmt.Errorf("loading .env file: %w", err)
		}
	}

	// Background, not a caller's context. v10's loader takes one so that config validation
	// can be cancelled, but this runs once at process startup before there is a request or a
	// server context to cancel against, and a config load that gave up halfway would leave
	// the process with no configuration rather than with less of it.
	cfg, err := platformconfig.LoadFromJSONFile[T](context.Background(), configurationFilepath, envVarOptions()...)
	if err != nil {
		return nil, fmt.Errorf("loading config from path: %w", err)
	}

	return cfg, nil
}

// LoadConfigFromDotEnvFile loads a configuration entirely from a .env file, with no JSON config file baseline.
// Because there is no JSON baseline to fall back on, the resulting config is validated to ensure
// the caller provided enough values to produce a usable configuration.
// godotenv.Load does not override env vars already set in the process, so actual process env vars
// still take precedence over values in the file.
func LoadConfigFromDotEnvFile[T configurations](ctx context.Context, dotEnvFilepath string) (*T, error) {
	cfg, err := platformconfig.LoadFromDotEnvFile[T](dotEnvFilepath, envVarOptions()...)
	if err != nil {
		return nil, fmt.Errorf("loading config from .env file: %w", err)
	}

	if err = platformconfig.Validate(ctx, cfg); err != nil {
		return nil, fmt.Errorf("validating config loaded from .env file: %w", err)
	}

	return cfg, nil
}
