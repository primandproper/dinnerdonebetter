package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	ddbaudit "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	"github.com/primandproper/platform-go/v9/audit"
	"github.com/primandproper/platform-go/v9/database/dialect"
	distributedlockcfg "github.com/primandproper/platform-go/v9/distributedlock/config"
	pglock "github.com/primandproper/platform-go/v9/distributedlock/postgres"
	"github.com/primandproper/platform-go/v9/jobs"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/outbox"
	retrycfg "github.com/primandproper/platform-go/v9/retry/config"
	"github.com/primandproper/platform-go/v9/saga"

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
			LockKeyPrefix: "dinner_done_better.scheduler.",
			// Named rather than left empty. The default is UTC either way, but a cron
			// expression's zone is the one thing about it that cannot be read off the
			// expression, and a zone that arrives by omission is a zone nobody chose.
			// A job that wants a calendar says so with its own CRON_TZ= prefix.
			Timezone:        "UTC",
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
		// A bulk re-index competing with daytime traffic for the same tables, so it is
		// confined to the small hours — 06:00-11:59 UTC is roughly midnight to 6am US
		// Central, an hour later in summer. In UTC and not Central because the window is
		// about load rather than about people, and a fixed instant has no daylight saving
		// day where it runs twice or not at all.
		//
		// A window rather than a single nightly fire, and at the same ten-minute spacing it
		// ran at around the clock, because IndexScheduler.IndexTypes sweeps one randomly
		// chosen index type per run rather than all of them. Nine types are registered, so
		// one fire a night would sweep a given type every nine days on average; thirty-six
		// fires a night covers all nine with room to spare. The interval this replaced was
		// really a draw rate dressed up as a frequency.
		SearchDataIndexScheduler: ScheduledJobConfig{
			Enabled:  true,
			Schedule: "*/10 6-11 * * *",
			// Fires once at startup as well, because an overnight window is a long time
			// for a freshly deployed environment to have no sweep at all, and because it
			// is otherwise the whole working day before a developer running localdev sees
			// this job do anything. A sweep publishes index requests for rows that need
			// indexing, so an extra one is redundant rather than harmful.
			RunOnStart: true,
			Timeout:    5 * time.Minute,
			LeaseTTL:   10 * time.Minute,
		},
		// Push notifications, so the hours are the point: this fires on the hour from 08:00
		// to 21:00 US Central and never overnight. It carries its own zone because it is the
		// one job here whose correctness is a fact about people rather than about load, and
		// because the scheduler's own default is deliberately UTC.
		//
		// Hourly rather than once in the morning: the query is "every prep task not yet
		// notified, for an event that has not started", and each task notifies exactly once.
		// A task created after the day's last fire waits for the next one, and is dropped
		// entirely if its event starts first — so the gap between fires bounds how much
		// short notice the app can give.
		//
		// Per-user timezones are the real answer here and a much larger conversation; one
		// zone's waking hours are strictly better than every two minutes in the meantime.
		MobileNotificationScheduler: ScheduledJobConfig{
			Enabled:  true,
			Schedule: "CRON_TZ=America/Chicago 0 8-21 * * *",
			// An hour's worth of accumulated tasks per run instead of two minutes'
			// worth, and a few queries per task, so the old one-minute bound is no
			// longer generous.
			Timeout:  5 * time.Minute,
			LeaseTTL: 10 * time.Minute,
		},
		QueueTest: ScheduledJobConfig{
			Enabled:  true,
			Interval: 15 * time.Minute,
			Timeout:  time.Minute,
			LeaseTTL: 2 * time.Minute,
		},
		DataPrivacySweep: ScheduledJobConfig{
			Enabled: true,
			// Hourly is far finer than the seven-day artifact TTL needs, and a sweep with
			// nothing to do costs three indexed queries against partial indexes. It is also
			// the cadence the overdue gauge is sampled at, which is the reason not to make it
			// coarser: a deadline nobody is looking at is the same as no deadline.
			Interval:   time.Hour,
			Timeout:    5 * time.Minute,
			LeaseTTL:   10 * time.Minute,
			RunOnStart: true,
		},
		AuditRetentionSweeper: ScheduledJobConfig{
			Enabled: true,
			// Daily, in the same overnight window the bulk re-index uses and for the same
			// reason: one sweep removes a bounded batch per scope, so it is cheap, but it
			// is a DELETE against the table every write path touches.
			//
			// No RunOnStart. The other jobs' first run is catch-up work; this one's would
			// be a deletion, and a deletion from the audit log is not a thing to do a few
			// seconds into a deploy on the strength of a config file nobody has read yet.
			Schedule: "17 7 * * *",
			Timeout:  10 * time.Minute,
			LeaseTTL: 20 * time.Minute,
		},
		// Domain: mealplanning
		MealPlanning: defaultMealPlanningScheduledJobsConfig(),
	}
}

// defaultAuditSweeperConfig returns the audit log's retention knobs.
//
// Two years, against the platform's seven-year default. Seven is the window the regulations
// that ask for an audit log in the first place tend to name, and this application is under
// none of them; two still comfortably covers a dispute, an incident review, or a question
// about an account somebody closed last year, which is what this log actually gets asked.
//
// It is a knob rather than a constant because the right answer is a deployment's to make, and
// shortening it is a config change rather than a code change. Lengthening it is too — but note
// that lengthening only affects what has not already been swept. Retention deletes; a window
// that was too short is not recoverable by widening it afterwards.
//
// The batch and scope limits are the platform defaults: one sweep removes at most a thousand
// entries from any one scope and visits at most a hundred scopes, so a long-neglected log is
// trimmed over several passes rather than by one DELETE holding locks for minutes.
func defaultAuditSweeperConfig() audit.SweeperConfig {
	return audit.SweeperConfig{
		Dialect:       dialect.Postgres,
		TablePrefix:   ddbaudit.TablePrefix,
		Retention:     2 * 365 * 24 * time.Hour,
		SweepInterval: time.Hour,
		BatchSize:     audit.DefaultSweepBatchSize,
		ScopeLimit:    audit.DefaultSweepScopeLimit,
	}
}

// defaultOutboxRelayConfig returns the relay's knobs.
//
// ClaimSkipLocked so the relay stays correct if it is ever scaled past one replica; per-key
// ordering survives it, because the claim predicate admits a keyed message only when no older
// message with that key is still pending, and the emitter keys every event by account.
func defaultOutboxRelayConfig() outbox.RelayConfig {
	return outbox.RelayConfig{
		TablePrefix: outbox.DefaultTablePrefix,
		ClaimMode:   outbox.ClaimSkipLocked,
		Backoff: retrycfg.Config{
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

// defaultSagaWorkerConfig returns the settings for the loop that advances saga instances.
//
// The package's own defaults, spelled out rather than left to EnsureDefaults, because these are
// rendered into the environment config files and a knob that is blank in the file and non-blank
// in the binary is a knob nobody can reason about from the file.
//
// The one departure is StepTimeout. A meal plan finalization step reads a plan with all of its
// events, options, and votes, then generates prep tasks or a whole grocery list from it, and the
// package's thirty seconds is sized for a third-party call rather than for that. The three
// timeouts move together: a pass must fit at least one step, and both the lease and the lock
// must outlast a pass plus the step it may still have running.
func defaultSagaWorkerConfig() saga.WorkerConfig {
	return saga.WorkerConfig{
		LockKeyPrefix:        saga.DefaultLockKeyPrefix,
		IdempotencyKeyPrefix: saga.DefaultIdempotencyKeyPrefix,
		Backoff: retrycfg.Config{
			MaxAttempts:  3,
			InitialDelay: time.Second,
			MaxDelay:     time.Minute,
			Multiplier:   2,
			UseJitter:    true,
		},
		// Deliberately more attempts than the forward budget. Giving up going forward costs a
		// compensation; giving up on a compensation costs somebody's evening.
		CompensationBackoff: retrycfg.Config{
			MaxAttempts:  saga.DefaultCompensationAttempts,
			InitialDelay: time.Second,
			MaxDelay:     time.Minute,
			Multiplier:   2,
			UseJitter:    true,
		},
		PollInterval:   time.Second,
		StepTimeout:    2 * time.Minute,
		AdvanceTimeout: 5 * time.Minute,
		LeaseDuration:  10 * time.Minute,
		LockTTL:        10 * time.Minute,
		BatchSize:      saga.DefaultBatchSize,
		Concurrency:    saga.DefaultConcurrency,
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
			Retry: retrycfg.Config{
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

	// There is no webhook pool here any more. Outbound delivery is not a queue topic: a
	// dispatch row is claimed by the delivery worker, whose own concurrency, retry schedule,
	// and per-endpoint circuit breaking live in the webhooks config.

	return WorkerPoolsConfig{
		DeadLetterTopicName: DefaultDeadLetterTopicName,
		DataChanges:         dataChanges,
		OutboundEmails:      outboundEmails,
		SearchIndexRequests: standard(),
		MobileNotifications: standard(),
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
func databaseConfigForService(cfg *dbcfg.Config, users map[string]string, serviceName string) dbcfg.Config {
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
		DataPrivacy:   s.RootConfig.Services.DataPrivacy,
		Jobs:          defaultScheduledJobsConfig(),
		Outbox:        defaultOutboxRelayConfig(),
		Audit:         defaultAuditSweeperConfig(),
		Sagas:         defaultSagaWorkerConfig(),
		// The same webhook configuration the API service writes with, so the worker
		// claims from the tables the dispatch rows are written into.
		Webhooks: s.RootConfig.Webhooks,
	}
	schedulerConfig.Observability.Tracing.ServiceName = schedulerConfigObservabilityServiceName
	schedulerConfig.Observability.Metrics.ServiceName = schedulerConfigObservabilityServiceName
	schedulerConfig.Observability.Logging.ServiceName = schedulerConfigObservabilityServiceName
	schedulerConfig.Observability.Profiling.ServiceName = schedulerConfigObservabilityServiceName

	amhConfig := &AsyncMessageHandlerConfig{
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
