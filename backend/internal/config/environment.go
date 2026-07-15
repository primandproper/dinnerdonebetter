package config

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path"
	"strings"

	databasecfg "github.com/primandproper/platform-go/v4/database/config"
	"github.com/primandproper/platform-go/v4/observability"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

// EnvironmentConfigSet contains a way of rendering a set of every config for a given environment to a given folder.
type EnvironmentConfigSet struct {
	RootConfig                               *APIServiceConfig
	ServiceDatabaseUsers                     map[string]string
	SearchDataIndexSchedulerConfigPath       string
	MealPlanFinalizerConfigPath              string // Domain: mealplanning
	MealPlanGroceryListInitializerConfigPath string // Domain: mealplanning
	MealPlanTaskCreatorConfigPath            string // Domain: mealplanning
	DBCleanerConfigPath                      string
	MobileNotificationSchedulerConfigPath    string
	AsyncMessageHandlerConfigPath            string
	EmailDeliverabilityTestConfigPath        string
	QueueTestJobConfigPath                   string
	APIServiceConfigPath                     string
	MCPServiceConfigPath                     string
}

func stringOrDefault(s, defaultStr string) string {
	if s != "" {
		return s
	}
	return defaultStr
}

func renderJSON(obj any, pretty bool) []byte {
	var (
		b   []byte
		err error
	)
	if pretty {
		b, err = json.MarshalIndent(obj, "", "\t")
	} else {
		b, err = json.Marshal(obj)
	}

	if err != nil {
		panic(err)
	}

	return b
}

func writeFile(p string, content []byte) error {
	//nolint:gosec // I want this to be 644 I think
	return os.WriteFile(p, content, 0o0644)
}

// disableWorkerOtelMetrics turns off runtime and host metrics for worker configs to reduce cardinality.
// It clones the Otel config so the root config (API server, async message handler) is not mutated.
func disableWorkerOtelMetrics(obs *observability.Config) {
	if obs == nil || obs.Metrics.Otel == nil {
		return
	}
	copied := *obs.Metrics.Otel
	copied.EnableRuntimeMetrics = false
	copied.EnableHostMetrics = false
	obs.Metrics.Otel = &copied
}

// databaseConfigForService returns a copy of the given database config with the username
// overridden for the named service, if a mapping exists in users. Otherwise returns a copy unchanged.
func databaseConfigForService(cfg *databasecfg.Config, users map[string]string, serviceName string) databasecfg.Config {
	out := *cfg
	if username, ok := users[serviceName]; ok {
		out.ReadConnection.Username = username
		out.WriteConnection.Username = username
	}
	return out
}

const (
	apiConfigObservabilityServiceName  = "api_server"
	dbcConfigObservabilityServiceName  = "db_cleaner"
	sdisConfigObservabilityServiceName = "search_data_index_scheduler"
	mnsConfigObservabilityServiceName  = "mobile_notification_scheduler"
	amhConfigObservabilityServiceName  = "async_message_handler"
	edtConfigObservabilityServiceName  = "email_deliverability_test"
	qtConfigObservabilityServiceName   = "queue_test"
	mcpConfigObservabilityServiceName  = "dinner_done_better_mcp_server"
)

func (s *EnvironmentConfigSet) Render(outputDir string, pretty, validate bool) error {
	if err := os.MkdirAll(outputDir, 0o0750); err != nil {
		return err
	}
	errs := &multierror.Error{}

	// Ensure API server config has the correct observability name before writing.
	s.RootConfig.Observability.Tracing.ServiceName = apiConfigObservabilityServiceName
	s.RootConfig.Observability.Metrics.ServiceName = apiConfigObservabilityServiceName
	s.RootConfig.Observability.Logging.ServiceName = apiConfigObservabilityServiceName
	s.RootConfig.Observability.Profiling.ServiceName = apiConfigObservabilityServiceName
	if s.RootConfig.Routing.Chi != nil {
		s.RootConfig.Routing.Chi.ServiceName = apiConfigObservabilityServiceName
	}

	// write files
	if err := writeFile(
		path.Join(outputDir, stringOrDefault(s.APIServiceConfigPath, "api_service_config.json")),
		renderJSON(s.RootConfig, pretty),
	); err != nil {
		errs = multierror.Append(errs, err)
	}

	dbcConfig := &DBCleanerConfig{
		Observability: s.RootConfig.Observability,
		Database:      databaseConfigForService(&s.RootConfig.Database, s.ServiceDatabaseUsers, dbcConfigObservabilityServiceName),
	}
	dbcConfig.Observability.Tracing.ServiceName = dbcConfigObservabilityServiceName
	dbcConfig.Observability.Metrics.ServiceName = dbcConfigObservabilityServiceName
	dbcConfig.Observability.Logging.ServiceName = dbcConfigObservabilityServiceName
	dbcConfig.Observability.Profiling.ServiceName = dbcConfigObservabilityServiceName
	disableWorkerOtelMetrics(&dbcConfig.Observability)

	// Domain: mealplanning
	mealPlanningConfigs, mealPlanningFiles := s.renderMealPlanningConfigs(outputDir, pretty)

	sdisConfig := &SearchDataIndexSchedulerConfig{
		Observability: s.RootConfig.Observability,
		Events:        s.RootConfig.Events,
		Database:      databaseConfigForService(&s.RootConfig.Database, s.ServiceDatabaseUsers, sdisConfigObservabilityServiceName),
		Queues:        s.RootConfig.Queues,
	}
	sdisConfig.Observability.Tracing.ServiceName = sdisConfigObservabilityServiceName
	sdisConfig.Observability.Metrics.ServiceName = sdisConfigObservabilityServiceName
	sdisConfig.Observability.Logging.ServiceName = sdisConfigObservabilityServiceName
	sdisConfig.Observability.Profiling.ServiceName = sdisConfigObservabilityServiceName
	disableWorkerOtelMetrics(&sdisConfig.Observability)

	mnsConfig := &MobileNotificationSchedulerConfig{
		Observability: s.RootConfig.Observability,
		Events:        s.RootConfig.Events,
		Database:      databaseConfigForService(&s.RootConfig.Database, s.ServiceDatabaseUsers, mnsConfigObservabilityServiceName),
		Queues:        s.RootConfig.Queues,
	}
	mnsConfig.Observability.Tracing.ServiceName = mnsConfigObservabilityServiceName
	mnsConfig.Observability.Metrics.ServiceName = mnsConfigObservabilityServiceName
	mnsConfig.Observability.Logging.ServiceName = mnsConfigObservabilityServiceName
	mnsConfig.Observability.Profiling.ServiceName = mnsConfigObservabilityServiceName
	disableWorkerOtelMetrics(&mnsConfig.Observability)

	amhConfig := &AsyncMessageHandlerConfig{
		Storage:           s.RootConfig.Services.DataPrivacy.Uploads.Storage,
		Queues:            s.RootConfig.Queues,
		Email:             s.RootConfig.Email,
		Analytics:         s.RootConfig.Analytics,
		Search:            s.RootConfig.TextSearch,
		Events:            s.RootConfig.Events,
		Observability:     s.RootConfig.Observability,
		Database:          databaseConfigForService(&s.RootConfig.Database, s.ServiceDatabaseUsers, amhConfigObservabilityServiceName),
		PushNotifications: s.RootConfig.PushNotifications,
		BaseURL:           s.RootConfig.BaseURL,
	}
	amhConfig.Observability.Tracing.ServiceName = amhConfigObservabilityServiceName
	amhConfig.Observability.Metrics.ServiceName = amhConfigObservabilityServiceName
	amhConfig.Observability.Logging.ServiceName = amhConfigObservabilityServiceName
	amhConfig.Observability.Profiling.ServiceName = amhConfigObservabilityServiceName

	edtServiceEnv := "prod"
	if strings.Contains(outputDir, "localdev") {
		edtServiceEnv = "dev"
	} else if strings.Contains(outputDir, "testing") {
		edtServiceEnv = "testing"
	}
	edtConfig := &EmailDeliverabilityTestConfig{
		Observability:         s.RootConfig.Observability,
		Email:                 s.RootConfig.Email,
		RecipientEmailAddress: "verygoodsoftwarenotvirus@protonmail.com",
		ServiceEnvironment:    edtServiceEnv,
	}
	edtConfig.Observability.Tracing.ServiceName = edtConfigObservabilityServiceName
	edtConfig.Observability.Metrics.ServiceName = edtConfigObservabilityServiceName
	edtConfig.Observability.Logging.ServiceName = edtConfigObservabilityServiceName
	edtConfig.Observability.Profiling.ServiceName = edtConfigObservabilityServiceName
	disableWorkerOtelMetrics(&edtConfig.Observability)

	qtConfig := &QueueTestJobConfig{
		Observability: s.RootConfig.Observability,
		Events:        s.RootConfig.Events,
		Database:      databaseConfigForService(&s.RootConfig.Database, s.ServiceDatabaseUsers, qtConfigObservabilityServiceName),
		Queues:        s.RootConfig.Queues,
	}
	qtConfig.Observability.Tracing.ServiceName = qtConfigObservabilityServiceName
	qtConfig.Observability.Metrics.ServiceName = qtConfigObservabilityServiceName
	qtConfig.Observability.Logging.ServiceName = qtConfigObservabilityServiceName
	qtConfig.Observability.Profiling.ServiceName = qtConfigObservabilityServiceName
	disableWorkerOtelMetrics(&qtConfig.Observability)

	mcpObservability := s.RootConfig.Observability
	mcpObservability.Tracing.ServiceName = mcpConfigObservabilityServiceName
	mcpObservability.Metrics.ServiceName = mcpConfigObservabilityServiceName
	mcpObservability.Logging.ServiceName = mcpConfigObservabilityServiceName
	mcpObservability.Profiling.ServiceName = mcpConfigObservabilityServiceName
	disableWorkerOtelMetrics(&mcpObservability)

	mcpRouting := s.RootConfig.Routing
	if mcpRouting.Chi != nil {
		mcpRouting.Chi.ServiceName = mcpConfigObservabilityServiceName
		// MCP clients (e.g. the MCP inspector) run in browsers on localhost,
		// so the MCP server must always allow localhost CORS origins.
		mcpRouting.Chi.EnableCORSForLocalhost = true
	}

	mcpConfig := &MCPServiceConfig{
		Database:      databaseConfigForService(&s.RootConfig.Database, s.ServiceDatabaseUsers, mcpConfigObservabilityServiceName),
		Observability: mcpObservability,
		Routing:       mcpRouting,
		Meta:          s.RootConfig.Meta,
		HTTPServer:    s.RootConfig.HTTPServer,
	}

	if validate {
		allConfigs := []validation.ValidatableWithContext{
			s.RootConfig,
			dbcConfig,
			sdisConfig,
			mnsConfig,
			amhConfig,
			edtConfig,
			qtConfig,
			mcpConfig,
		}
		allConfigs = append(allConfigs, mealPlanningConfigs...)
		for i, cfg := range allConfigs {
			if err := cfg.ValidateWithContext(context.Background()); err != nil {
				errs = multierror.Append(errs, fmt.Errorf("validating config %d: %w", i, err))
				continue
			}
		}
	}

	if err := errs.ErrorOrNil(); err != nil {
		return err
	}

	pathToConfigMap := map[string][]byte{
		path.Join(outputDir, stringOrDefault(s.DBCleanerConfigPath, "job_db_cleaner_config.json")):                                      renderJSON(dbcConfig, pretty),
		path.Join(outputDir, stringOrDefault(s.SearchDataIndexSchedulerConfigPath, "job_search_data_index_scheduler_config.json")):      renderJSON(sdisConfig, pretty),
		path.Join(outputDir, stringOrDefault(s.MobileNotificationSchedulerConfigPath, "job_mobile_notification_scheduler_config.json")): renderJSON(mnsConfig, pretty),
		path.Join(outputDir, stringOrDefault(s.AsyncMessageHandlerConfigPath, "async_message_handler_config.json")):                     renderJSON(amhConfig, pretty),
		path.Join(outputDir, stringOrDefault(s.EmailDeliverabilityTestConfigPath, "job_email_deliverability_test_config.json")):         renderJSON(edtConfig, pretty),
		path.Join(outputDir, stringOrDefault(s.QueueTestJobConfigPath, "job_queue_test_config.json")):                                   renderJSON(qtConfig, pretty),
		path.Join(outputDir, stringOrDefault(s.MCPServiceConfigPath, "mcp_server_config.json")):                                         renderJSON(mcpConfig, pretty),
	}

	// Domain: mealplanning
	maps.Copy(pathToConfigMap, mealPlanningFiles)

	for p, b := range pathToConfigMap {
		if err := writeFile(p, b); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}
