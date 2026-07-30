package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	databasecfg "github.com/primandproper/platform-go/v8/database/config"
	"github.com/primandproper/platform-go/v8/database/dialect"
	distributedlockcfg "github.com/primandproper/platform-go/v8/distributedlock/config"
	pglock "github.com/primandproper/platform-go/v8/distributedlock/postgres"
	"github.com/primandproper/platform-go/v8/jobs"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/outbox"
	"github.com/primandproper/platform-go/v8/retry"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

// EnvironmentConfigSet contains a way of rendering a set of every config for a given environment to a given folder.
type EnvironmentConfigSet struct {
	RootConfig                        *APIServiceConfig
	ServiceDatabaseUsers              map[string]string
	SchedulerConfigPath               string
	DBCleanerConfigPath               string
	AsyncMessageHandlerConfigPath     string
	EmailDeliverabilityTestConfigPath string
	APIServiceConfigPath              string
	MCPServiceConfigPath              string
}

// defaultScheduledJobsConfig returns the schedule for each periodic job, replacing what used to
// be one Kubernetes CronJob per job.
//
// Every LeaseTTL is set well above the job's observed worst case rather than near it: the lease
// is not renewed while a job runs, so a job that outlives its lease can be started a second time
// on another replica. Timeout is the shorter of the two — a job that hangs should be killed
// before its lease lapses, not after.
func defaultScheduledJobsConfig() ScheduledJobsConfig {
	return ScheduledJobsConfig{
		Scheduler: jobs.SchedulerConfig{
			LockKeyPrefix:   "dinner_done_better.scheduler.",
			DefaultLeaseTTL: 2 * time.Minute,
			DefaultTimeout:  time.Minute,
		},
		Lock: distributedlockcfg.Config{
			// Postgres advisory locks: no new infrastructure, and a replica that dies
			// drops its connection, which releases the lock without waiting for a TTL.
			Provider: distributedlockcfg.PostgresProvider,
			Postgres: &pglock.Config{
				ConnWaitTimeout: 5 * time.Second,
			},
		},
		SearchDataIndexScheduler: ScheduledJobConfig{
			Enabled:  true,
			Interval: 10 * time.Minute,
			Timeout:  5 * time.Minute,
			LeaseTTL: 10 * time.Minute,
		},
		MobileNotificationScheduler: ScheduledJobConfig{
			Enabled:  true,
			Interval: 2 * time.Minute,
			Timeout:  time.Minute,
			LeaseTTL: 3 * time.Minute,
		},
		QueueTest: ScheduledJobConfig{
			Enabled:  true,
			Interval: 15 * time.Minute,
			Timeout:  time.Minute,
			LeaseTTL: 2 * time.Minute,
		},
		// Domain: mealplanning
		MealPlanning: defaultMealPlanningScheduledJobsConfig(),
	}
}

// defaultOutboxRelayConfig returns the relay's knobs.
//
// ClaimSkipLocked so the relay stays correct if it is ever scaled past one replica; per-key
// ordering survives it, because the claim predicate admits a keyed message only when no older
// message with that key is still pending, and the emitter keys every event by account.
func defaultOutboxRelayConfig() outbox.RelayConfig {
	return outbox.RelayConfig{
		Dialect:   dialect.Postgres,
		TableName: outbox.DefaultTableName,
		ClaimMode: outbox.ClaimSkipLocked,
		Backoff: retry.Config{
			MaxAttempts:  8,
			InitialDelay: time.Second,
			MaxDelay:     time.Minute,
			Multiplier:   2,
			UseJitter:    true,
		},
		BatchSize:     100,
		PollInterval:  time.Second,
		LeaseDuration: 30 * time.Second,
		// Published rows are marked rather than deleted, so a duplicate or a gap can be
		// investigated after the fact. A week is long enough to answer "did that event
		// actually go out" during an incident review.
		Retention:     7 * 24 * time.Hour,
		ReapInterval:  time.Hour,
		ReapBatchSize: 1000,
	}
}

// DefaultDeadLetterTopicName is where the async message handler's pools send messages that have
// exhausted their attempts. Nothing consumes it; it exists so a permanently failing message has
// somewhere to land other than a log line.
const DefaultDeadLetterTopicName = "dead_letter"

// defaultWorkerPoolsConfig returns the pool shapes the async message handler runs with.
//
// The knobs differ per topic because the work does. Concurrency is the bound on how many messages
// can be lost to a crash, so it is smallest where a lost message is most expensive. HandlerTimeout
// bounds one attempt: without it a handler that neither returns nor honors its context occupies a
// worker permanently and holds up shutdown.
func defaultWorkerPoolsConfig() WorkerPoolsConfig {
	standard := func() jobs.PoolConfig {
		return jobs.PoolConfig{
			Concurrency:    8,
			HandlerTimeout: 30 * time.Second,
			Retry: retry.Config{
				MaxAttempts:  3,
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     5 * time.Second,
				Multiplier:   2,
				UseJitter:    true,
			},
		}
	}

	// The fan-out hub: every domain event lands here and is routed onward, so it carries the
	// most traffic of any topic.
	dataChanges := standard()
	dataChanges.Concurrency = 16

	// Third-party delivery that blips: worth more attempts and a longer ceiling than work that
	// only touches our own infrastructure.
	outboundEmails := standard()
	outboundEmails.Retry.MaxAttempts = 4
	outboundEmails.Retry.MaxDelay = 30 * time.Second

	// Outbound HTTP to admin-configured URLs. Mostly waiting on someone else's server, so the
	// concurrency is high and the timeout is what stops one slow endpoint from pinning a worker.
	webhooks := standard()
	webhooks.Concurrency = 16
	webhooks.Retry.MaxAttempts = 4
	webhooks.Retry.MaxDelay = 30 * time.Second

	// GDPR data exports: slow, heavy, and duplicated work means a user gets two archives. A
	// crash can lose Concurrency+1 messages, so this is the one topic where that bound is 1.
	userDataAggregation := standard()
	userDataAggregation.Concurrency = 1
	userDataAggregation.Retry.MaxAttempts = 2
	userDataAggregation.HandlerTimeout = 10 * time.Minute

	return WorkerPoolsConfig{
		DeadLetterTopicName:      DefaultDeadLetterTopicName,
		DataChanges:              dataChanges,
		OutboundEmails:           outboundEmails,
		SearchIndexRequests:      standard(),
		WebhookExecutionRequests: webhooks,
		UserDataAggregation:      userDataAggregation,
		MobileNotifications:      standard(),
	}
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
	apiConfigObservabilityServiceName       = "api_server"
	dbcConfigObservabilityServiceName       = "db_cleaner"
	schedulerConfigObservabilityServiceName = "scheduler"
	amhConfigObservabilityServiceName       = "async_message_handler"
	edtConfigObservabilityServiceName       = "email_deliverability_test"
	mcpConfigObservabilityServiceName       = "dinner_done_better_mcp_server"
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

	// One config for every interval-shaped periodic job, because they now share one process.
	schedulerConfig := &SchedulerConfig{
		Observability: s.RootConfig.Observability,
		Analytics:     s.RootConfig.Analytics,
		Events:        s.RootConfig.Events,
		Search:        s.RootConfig.TextSearch,
		Database:      databaseConfigForService(&s.RootConfig.Database, s.ServiceDatabaseUsers, schedulerConfigObservabilityServiceName),
		Queues:        s.RootConfig.Queues,
		Jobs:          defaultScheduledJobsConfig(),
		Outbox:        defaultOutboxRelayConfig(),
	}
	schedulerConfig.Observability.Tracing.ServiceName = schedulerConfigObservabilityServiceName
	schedulerConfig.Observability.Metrics.ServiceName = schedulerConfigObservabilityServiceName
	schedulerConfig.Observability.Logging.ServiceName = schedulerConfigObservabilityServiceName
	schedulerConfig.Observability.Profiling.ServiceName = schedulerConfigObservabilityServiceName

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
		Pools:             defaultWorkerPoolsConfig(),
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

	mcpHTTPServer := s.RootConfig.HTTPServer
	// The apple-app-site-association document describes the domain the iOS app is
	// associated with, which is the API's, not the MCP server's. Serving it from here
	// would publish an association for a host no Universal Link points at.
	mcpHTTPServer.AppleAppSiteAssociation = nil

	mcpConfig := &MCPServiceConfig{
		Database:      databaseConfigForService(&s.RootConfig.Database, s.ServiceDatabaseUsers, mcpConfigObservabilityServiceName),
		Observability: mcpObservability,
		Routing:       mcpRouting,
		Meta:          s.RootConfig.Meta,
		HTTPServer:    mcpHTTPServer,
	}

	if validate {
		allConfigs := []validation.ValidatableWithContext{
			s.RootConfig,
			dbcConfig,
			schedulerConfig,
			amhConfig,
			edtConfig,
			mcpConfig,
		}
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
		path.Join(outputDir, stringOrDefault(s.DBCleanerConfigPath, "job_db_cleaner_config.json")):                              renderJSON(dbcConfig, pretty),
		path.Join(outputDir, stringOrDefault(s.SchedulerConfigPath, "scheduler_config.json")):                                   renderJSON(schedulerConfig, pretty),
		path.Join(outputDir, stringOrDefault(s.AsyncMessageHandlerConfigPath, "async_message_handler_config.json")):             renderJSON(amhConfig, pretty),
		path.Join(outputDir, stringOrDefault(s.EmailDeliverabilityTestConfigPath, "job_email_deliverability_test_config.json")): renderJSON(edtConfig, pretty),
		path.Join(outputDir, stringOrDefault(s.MCPServiceConfigPath, "mcp_server_config.json")):                                 renderJSON(mcpConfig, pretty),
	}

	for p, b := range pathToConfigMap {
		if err := writeFile(p, b); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}
