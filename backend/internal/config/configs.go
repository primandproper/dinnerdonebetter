package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	authcfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/branding"

	analyticscfg "github.com/primandproper/platform-go/v8/analytics/config"
	platformconfig "github.com/primandproper/platform-go/v8/config"
	databasecfg "github.com/primandproper/platform-go/v8/database/config"
	emailcfg "github.com/primandproper/platform-go/v8/email/config"
	"github.com/primandproper/platform-go/v8/encoding"
	featureflagscfg "github.com/primandproper/platform-go/v8/featureflags/config"
	httpclientcfg "github.com/primandproper/platform-go/v8/httpclient"
	msgconfig "github.com/primandproper/platform-go/v8/messagequeue/config"
	notificationscfg "github.com/primandproper/platform-go/v8/notifications/mobile/config"
	"github.com/primandproper/platform-go/v8/observability"
	routingcfg "github.com/primandproper/platform-go/v8/routing/config"
	textsearchcfg "github.com/primandproper/platform-go/v8/search/text/config"
	"github.com/primandproper/platform-go/v8/server/grpc"
	"github.com/primandproper/platform-go/v8/server/http"
	"github.com/primandproper/platform-go/v8/uploads/objectstorage"

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
			MealPlanFinalizerConfig |
			MealPlanGroceryListInitializerConfig |
			MealPlanTaskCreatorConfig |
			SearchDataIndexSchedulerConfig |
			MobileNotificationSchedulerConfig |
			AsyncMessageHandlerConfig |
			EmailDeliverabilityTestConfig |
			QueueTestJobConfig |
			MCPServiceConfig
	}

	// APIServiceConfig configures an instance of the service. It is composed of all the other setting structs.
	APIServiceConfig struct {
		_                 struct{}                `json:"-"`
		HTTPClient        *httpclientcfg.Config   `envPrefix:"HTTP_CLIENT_"        json:"httpClient"`
		Queues            msgconfig.QueuesConfig  `envPrefix:"QUEUES_"             json:"queues"`
		PushNotifications notificationscfg.Config `envPrefix:"PUSH_NOTIFICATIONS_" json:"pushNotifications"`
		Routing           routingcfg.Config       `envPrefix:"ROUTING_"            json:"routing"`
		Encoding          encoding.Config         `envPrefix:"ENCODING_"           json:"encoding"`
		BaseURL           string                  `env:"BASE_URL"                  json:"baseURL"`
		Events            msgconfig.Config        `envPrefix:"EVENTS_"             json:"events"`
		Observability     observability.Config    `envPrefix:"OBSERVABILITY_"      json:"observability"`
		GRPCServer        grpc.Config             `envPrefix:"GRPC_"               json:"grpc"`
		Meta              MetaSettings            `envPrefix:"META_"               json:"meta"`
		Analytics         analyticscfg.Config     `envPrefix:"ANALYTICS_"          json:"analytics"`
		Email             emailcfg.Config         `envPrefix:"EMAIL_"              json:"email"`
		FeatureFlags      featureflagscfg.Config  `envPrefix:"FEATURE_FLAGS_"      json:"featureFlags"`
		TextSearch        textsearchcfg.Config    `envPrefix:"SEARCH_"             json:"search"`
		HTTPServer        http.Config             `envPrefix:"HTTP_"               json:"http"`
		Auth              authcfg.Config          `envPrefix:"AUTH_"               json:"auth"`
		Database          databasecfg.Config      `envPrefix:"DATABASE_"           json:"database"`
		Services          ServicesConfig          `envPrefix:"SERVICE_"            json:"services"`
		validateServices  bool
	}

	// DBCleanerConfig configures an instance of the database cleaner job.
	DBCleanerConfig struct {
		_ struct{} `json:"-"`

		Observability observability.Config `envPrefix:"OBSERVABILITY_" json:"observability"`
		Database      databasecfg.Config   `envPrefix:"DATABASE_"      json:"database"`
	}

	// SearchDataIndexSchedulerConfig configures an instance of the search data index scheduler job.
	SearchDataIndexSchedulerConfig struct {
		_ struct{} `json:"-"`

		Queues        msgconfig.QueuesConfig `envPrefix:"QUEUES_"        json:"queues"`
		Events        msgconfig.Config       `envPrefix:"EVENTS_"        json:"events"`
		Observability observability.Config   `envPrefix:"OBSERVABILITY_" json:"observability"`
		Database      databasecfg.Config     `envPrefix:"DATABASE_"      json:"database"`
	}

	// MobileNotificationSchedulerConfig configures an instance of the mobile notification scheduler job.
	MobileNotificationSchedulerConfig struct {
		_ struct{} `json:"-"`

		Queues        msgconfig.QueuesConfig `envPrefix:"QUEUES_"        json:"queues"`
		Events        msgconfig.Config       `envPrefix:"EVENTS_"        json:"events"`
		Observability observability.Config   `envPrefix:"OBSERVABILITY_" json:"observability"`
		Database      databasecfg.Config     `envPrefix:"DATABASE_"      json:"database"`
	}

	// AsyncMessageHandlerConfig configures an instance of the search data index scheduler job.
	AsyncMessageHandlerConfig struct {
		_                 struct{}                `json:"-"`
		HTTPClient        *httpclientcfg.Config   `envPrefix:"HTTP_CLIENT_"        json:"httpClient"`
		Queues            msgconfig.QueuesConfig  `envPrefix:"QUEUES_"             json:"queues"`
		Storage           objectstorage.Config    `envPrefix:"STORAGE_"            json:"storage"`
		PushNotifications notificationscfg.Config `envPrefix:"PUSH_NOTIFICATIONS_" json:"pushNotifications"`
		Encoding          encoding.Config         `envPrefix:"ENCODING_"           json:"encoding"`
		BaseURL           string                  `env:"BASE_URL"                  json:"baseURL"`
		Events            msgconfig.Config        `envPrefix:"EVENTS_"             json:"events"`
		Observability     observability.Config    `envPrefix:"OBSERVABILITY_"      json:"observability"`
		Analytics         analyticscfg.Config     `envPrefix:"ANALYTICS_"          json:"analytics"`
		Email             emailcfg.Config         `envPrefix:"EMAIL_"              json:"email"`
		Search            textsearchcfg.Config    `envPrefix:"SEARCH_"             json:"search"`
		Database          databasecfg.Config      `envPrefix:"DATABASE_"           json:"database"`
	}

	// EmailDeliverabilityTestConfig configures the email deliverability test cron job.
	EmailDeliverabilityTestConfig struct {
		_                     struct{}              `json:"-"`
		HTTPClient            *httpclientcfg.Config `envPrefix:"HTTP_CLIENT_"      json:"httpClient"`
		RecipientEmailAddress string                `env:"RECIPIENT_EMAIL_ADDRESS" json:"recipientEmailAddress"`
		ServiceEnvironment    string                `env:"SERVICE_ENVIRONMENT"     json:"serviceEnvironment"`
		Observability         observability.Config  `envPrefix:"OBSERVABILITY_"    json:"observability"`
		Email                 emailcfg.Config       `envPrefix:"EMAIL_"            json:"email"`
	}

	// QueueTestJobConfig configures the queue test cron job.
	QueueTestJobConfig struct {
		_             struct{}               `json:"-"`
		Queues        msgconfig.QueuesConfig `envPrefix:"QUEUES_"        json:"queues"`
		Events        msgconfig.Config       `envPrefix:"EVENTS_"        json:"events"`
		Observability observability.Config   `envPrefix:"OBSERVABILITY_" json:"observability"`
		Database      databasecfg.Config     `envPrefix:"DATABASE_"      json:"database"`
	}

	// MCPServiceConfig configures an instance of the service. It is composed of all the other setting structs.
	MCPServiceConfig struct {
		_             struct{}             `json:"-"`
		Routing       routingcfg.Config    `envPrefix:"ROUTING_"       json:"routing"`
		Observability observability.Config `envPrefix:"OBSERVABILITY_" json:"observability"`
		Meta          MetaSettings         `envPrefix:"META_"          json:"meta"`
		HTTPServer    http.Config          `envPrefix:"HTTP_"          json:"http"`
		Database      databasecfg.Config   `envPrefix:"DATABASE_"      json:"database"`
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

var _ validation.ValidatableWithContext = (*SearchDataIndexSchedulerConfig)(nil)

// ValidateWithContext validates a SearchDataIndexSchedulerConfig struct.
func (cfg *SearchDataIndexSchedulerConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
		"Queues":        cfg.Queues.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*MobileNotificationSchedulerConfig)(nil)

// ValidateWithContext validates a MobileNotificationSchedulerConfig struct.
func (cfg *MobileNotificationSchedulerConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
		"Queues":        cfg.Queues.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
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
		"Storage":       cfg.Storage.ValidateWithContext,
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

var _ validation.ValidatableWithContext = (*QueueTestJobConfig)(nil)

// ValidateWithContext validates a QueueTestJobConfig struct.
func (cfg *QueueTestJobConfig) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	validators := map[string]func(context.Context) error{
		"Observability": cfg.Observability.ValidateWithContext,
		"Database":      cfg.Database.ValidateWithContext,
		"Queues":        cfg.Queues.ValidateWithContext,
	}

	for name, validator := range validators {
		if err := validator(ctx); err != nil {
			result = multierror.Append(fmt.Errorf("error validating %s config: %w", name, err), result)
		}
	}

	return result.ErrorOrNil()
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

	cfg, err := platformconfig.LoadFromJSONFile[T](configFilepath, envVarOptions()...)
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

	cfg, err := platformconfig.LoadFromJSONFile[T](configurationFilepath, envVarOptions()...)
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
