package config

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	ddbaudit "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddboauth "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v13/audit"
	auditcfg "github.com/primandproper/platform-go/v13/audit/config"
	oauth2servercfg "github.com/primandproper/platform-go/v13/authentication/oauth2server/config"
	oauth2database "github.com/primandproper/platform-go/v13/authentication/oauth2server/database"
	platformconfig "github.com/primandproper/platform-go/v13/config"
	"github.com/primandproper/platform-go/v13/database/dialect"
	distributedlockcfg "github.com/primandproper/platform-go/v13/distributedlock/config"
	pglock "github.com/primandproper/platform-go/v13/distributedlock/postgres"
	"github.com/primandproper/platform-go/v13/jobs"
	"github.com/primandproper/platform-go/v13/metering"
	meteringcfg "github.com/primandproper/platform-go/v13/metering/config"
	"github.com/primandproper/platform-go/v13/observability"
	operationscfg "github.com/primandproper/platform-go/v13/operations/config"
	"github.com/primandproper/platform-go/v13/outbox"
	"github.com/primandproper/platform-go/v13/retention"
	retrycfg "github.com/primandproper/platform-go/v13/retry/config"
	"github.com/primandproper/platform-go/v13/saga"

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
		// Every five minutes, which is the cadence the metering package recommends and the
		// one its lease and timeout defaults are sized for. A pass claims a bounded batch,
		// so a longer gap does not mean a bigger pass — it means a longer tail of usage
		// the provider has not been told about, and a longer wait before a period that
		// closed overnight is settled.
		//
		// LeaseTTL is well clear of the flusher's own FlushTimeout, and for the same
		// reason that config validates the relation: two flushers posting the same total
		// concurrently is the one duplicate charge an idempotency key cannot undo, because
		// the two posts carry different sequence numbers.
		MeteringFlusher: ScheduledJobConfig{
			Enabled:  true,
			Interval: 5 * time.Minute,
			Timeout:  2 * time.Minute,
			LeaseTTL: 10 * time.Minute,
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
func defaultAuditSweeperConfig() auditcfg.Config {
	return auditcfg.Config{
		Dialect:     dialect.Postgres,
		TablePrefix: ddbaudit.TablePrefix,
		Retention: audit.RetentionConfig{
			Retention:     2 * 365 * 24 * time.Hour,
			BatchSize:     audit.DefaultRetentionBatchSize,
			ScopePageSize: audit.DefaultScopePageSize,
		},
	}
}

// DefaultMeteringConfig returns the metering knobs every process shares.
//
// The values are the platform's own defaults, written out rather than left zero so that a
// rendered config shows what a deployment is actually running — an empty object in a config file
// says nothing about whether usage events are kept for a week or a quarter.
// DefaultOperationsConfig returns the operations tier's knobs, which every process that either
// starts an operation or runs one shares.
//
// They are the platform's own defaults, obtained by asking for them rather than by copying the
// numbers out: EnsureDefaults is what the constructors call, so a rendered config produced this
// way cannot disagree with the one a process would have built for itself. The values are then
// written out in full, so an operator reading the rendered JSON can see what is in force and
// change one of them without having to know which package supplied it.
//
// The queue's Name is deliberately not set here. EnsureDefaults derives it from
// Operations.QueueName, because two names for one queue is a misconfiguration whose only symptom
// is a table of pending operations that nothing ever claims.
func DefaultOperationsConfig() operationscfg.Config {
	cfg := operationscfg.Config{}
	cfg.EnsureDefaults()

	return cfg
}

func DefaultMeteringConfig() meteringcfg.Config {
	cfg := meteringcfg.Config{
		Recorder: metering.RecorderConfig{
			BatchSize: metering.DefaultBatchSize,
			// Drop and count a record naming an unregistered meter rather than
			// returning an error to the write site. A deploy that adds a meter reaches
			// the ingest path before it reaches every replica's wiring, and a Record
			// that failed in that window would turn a rollout into an outage on a path
			// that is supposed to be incidental to the request it rides on.
			RejectUnknownMeters: false,
		},
		Enforcer: metering.EnforcerConfig{
			CachePrefix: metering.DefaultCachePrefix,
			Staleness:   metering.DefaultStaleness,
			// Nothing is enforced yet, so this decides nothing today. Fail closed is
			// the value to have inherited when that changes: a quota that guards spend
			// and fails open during an outage bills us rather than the customer.
			FailOpen: false,
		},
		Flusher: metering.FlusherConfig{
			Backoff: retrycfg.Config{
				MaxAttempts:  metering.DefaultMaxFlushAttempts,
				InitialDelay: time.Second,
				MaxDelay:     5 * time.Minute,
				Multiplier:   2,
				UseJitter:    true,
			},
			LeaseDuration: metering.DefaultFlushLeaseDuration,
			FlushTimeout:  metering.DefaultFlushTimeout,
			// Ninety days of the event ledger, which is what makes ingest idempotent
			// for as long as anything could plausibly re-present a key: a dead-letter
			// redelivery, a batch replayed by hand after a bad deploy. It is also the
			// only record of what a total is made of when somebody disputes it.
			EventRetention: metering.DefaultEventRetention,
			BatchSize:      metering.DefaultFlushBatchSize,
			Concurrency:    metering.DefaultFlushConcurrency,
			MaxAttempts:    metering.DefaultMaxFlushAttempts,
			ReapBatchSize:  metering.DefaultReapBatchSize,
			DisableReap:    false,
		},
		TablePrefix: metering.DefaultTablePrefix,
	}

	cfg.EnsureDefaults()

	return cfg
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
		// New in v10, and required: the floor between two wake-driven cycles. Without it a
		// busy table can wake the relay faster than it can drain one, which spends the
		// cycle budget on wakeups rather than on publishing.
		MinWakeInterval: outbox.DefaultMinWakeInterval,
	}
}

// defaultRetentionSweeperConfig returns the bounds the retention sweep runs under.
//
// The platform's own defaults, asked for rather than copied, for the same reason
// DefaultOperationsConfig does it that way: EnsureDefaults is what the constructor calls, so a
// rendered config produced this way cannot disagree with the one the process would have built.
func defaultRetentionSweeperConfig() retention.SweeperConfig {
	cfg := retention.SweeperConfig{}
	cfg.EnsureDefaults()

	return cfg
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

// Render writes one config file per workload into outputDir.
//
// The files go out through platform's config.RenderJSONFiles, the documented inverse of
// LoadFromJSONFile: what this writes, that reads back. There is one call per config type
// because Environment[T] is generic over a single T.
//
// Indentation is not a parameter. RenderJSONFiles fixes it at one tab, deliberately — these
// files are checked in and read in diffs, and a file whose indentation depends on its last
// call site produces a diff that is all whitespace. Neither is validation: every config is
// validated, every time.
func (s *EnvironmentConfigSet) Render(ctx context.Context, outputDir string) error {
	// Ensure API server config has the correct observability name before writing.
	s.RootConfig.Observability.Tracing.ServiceName = apiConfigObservabilityServiceName
	s.RootConfig.Observability.Metrics.ServiceName = apiConfigObservabilityServiceName
	s.RootConfig.Observability.Logging.ServiceName = apiConfigObservabilityServiceName
	s.RootConfig.Observability.Profiling.ServiceName = apiConfigObservabilityServiceName
	if s.RootConfig.Routing.Chi != nil {
		s.RootConfig.Routing.Chi.ServiceName = apiConfigObservabilityServiceName
	}

	dbcConfig := &DBCleanerConfig{
		Observability: s.RootConfig.Observability,
		Database:      databaseConfigForService(&s.RootConfig.Database, s.ServiceDatabaseUsers, dbcConfigObservabilityServiceName),
		// This job sweeps the authorization server's tables, so it needs the prefix they
		// were created under and nothing else: a sweep asks the store for rows past their
		// deadlines, which needs neither an issuer nor a lifetime.
		OAuth2: oauth2servercfg.Config{
			Provider: oauth2servercfg.ProviderDatabase,
			Database: oauth2database.Config{TablePrefix: ddboauth.TablePrefix},
		},
	}
	dbcConfig.Observability.Tracing.ServiceName = dbcConfigObservabilityServiceName
	dbcConfig.Observability.Metrics.ServiceName = dbcConfigObservabilityServiceName
	dbcConfig.Observability.Logging.ServiceName = dbcConfigObservabilityServiceName
	dbcConfig.Observability.Profiling.ServiceName = dbcConfigObservabilityServiceName
	disableWorkerOtelMetrics(&dbcConfig.Observability)

	// One config for every interval-shaped periodic job, because they now share one process.
	schedulerConfig := &SchedulerConfig{
		Observability: s.RootConfig.Observability,
		// The same sender the async message handler pushes through, so a device token
		// this process sends to is one that process would have sent to.
		PushNotifications: s.RootConfig.PushNotifications,
		Analytics:         s.RootConfig.Analytics,
		Events:            s.RootConfig.Events,
		Search:            s.RootConfig.TextSearch,
		Database:          databaseConfigForService(&s.RootConfig.Database, s.ServiceDatabaseUsers, schedulerConfigObservabilityServiceName),
		Queues:            s.RootConfig.Queues,
		DataPrivacy:       s.RootConfig.Services.DataPrivacy,
		Jobs:              defaultScheduledJobsConfig(),
		Outbox:            defaultOutboxRelayConfig(),
		Audit:             defaultAuditSweeperConfig(),
		Retention:         defaultRetentionSweeperConfig(),
		Sagas:             defaultSagaWorkerConfig(),
		// This process runs the operations worker, so it carries the whole tier. The API
		// server carries the same struct for the enqueue-and-read half.
		Operations: DefaultOperationsConfig(),
		// The same webhook configuration the API service writes with, so the worker
		// claims from the tables the dispatch rows are written into.
		Webhooks: s.RootConfig.Webhooks,
		// Taken from the API server's config rather than rebuilt, so the tables the
		// recorder writes are by construction the tables the flusher flushes.
		Metering: s.RootConfig.Metering,
		// Likewise the billing provider: the flusher posts through whichever one the
		// payments service was configured with, so enabling real usage billing is one
		// provider setting rather than two that can disagree.
		Capitalism: s.RootConfig.Services.Payments.Capitalism,
	}
	schedulerConfig.Observability.Tracing.ServiceName = schedulerConfigObservabilityServiceName
	schedulerConfig.Observability.Metrics.ServiceName = schedulerConfigObservabilityServiceName
	schedulerConfig.Observability.Logging.ServiceName = schedulerConfigObservabilityServiceName
	schedulerConfig.Observability.Profiling.ServiceName = schedulerConfigObservabilityServiceName

	amhConfig := &AsyncMessageHandlerConfig{
		Queues: s.RootConfig.Queues,
		// The same encoder the API server writes its messages with, which is not a
		// nicety: the handler decodes what the API published, and this section was
		// missing entirely — a content type of "" is not a default, it is a decoder
		// that refuses to be built, so the process could not boot at all.
		Encoding:          s.RootConfig.Encoding,
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
		// Cloned, not shared — the same reason disableWorkerOtelMetrics clones the Otel config.
		// routingcfg.Config holds a *chi.Config, so assigning the struct above copies the
		// pointer and both configs address one chi.Config: the two writes below land on the API
		// server's routing config too, renaming its service and turning on localhost CORS.
		//
		// Nothing repairs that afterwards, and RootConfig belongs to the caller, so the write
		// outlives Render. Back when the generator rendered localdev twice from one RootConfig,
		// the second pass read the corrupted value and got away with it only because localdev
		// asks for EnableCORSForLocalhost true anyway. Every environment renders once now, but
		// the builders are exported and callable from anywhere, so the clone stays.
		chiConfig := *mcpRouting.Chi
		chiConfig.ServiceName = mcpConfigObservabilityServiceName
		// MCP clients (e.g. the MCP inspector) run in browsers on localhost,
		// so the MCP server must always allow localhost CORS origins.
		chiConfig.EnableCORSForLocalhost = true
		mcpRouting.Chi = &chiConfig
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
		// The authorization server the MCP server runs. Two fields are absent on purpose:
		// Issuer and Resources are the server's own public URL, which only the deployment
		// knows — it arrives as MCP_BASE_URL — so rendering a guess here would produce a
		// discovery document pointing somewhere nothing is listening. mcpserver.NewService
		// fills both from that URL.
		//
		// The table prefix is not optional in the same way. It has to be the one migration
		// 33 created the tables under, and a prefix that differs between the DDL and the
		// store is a server that comes up clean and cannot find a table.
		OAuth2: oauth2servercfg.Config{
			Provider: oauth2servercfg.ProviderDatabase,
			Database: oauth2database.Config{TablePrefix: ddboauth.TablePrefix},
		},
	}

	// RenderJSONFiles validates every environment it is handed before writing any of that
	// call's files, but it cannot see across the six calls below. Validating the whole set
	// here first restores the guarantee for the set as a whole: one invalid config leaves
	// every file as it was, rather than the ones rendered before it updated and the rest stale.
	for i, cfg := range []any{s.RootConfig, dbcConfig, schedulerConfig, amhConfig, edtConfig, mcpConfig} {
		if err := platformconfig.Validate(ctx, cfg); err != nil {
			return fmt.Errorf("validating config %d: %w", i, err)
		}
	}

	// The rendered files are checked in and read by everyone; platform's owner-only default
	// is the wrong mode for them. It applies on creation only, so this is what a fresh
	// checkout's first render gets.
	renderOpts := []platformconfig.RenderOption{platformconfig.WithFileMode(0o644)}

	renderErr := multierror.Append(
		nil,
		platformconfig.RenderJSONFiles(ctx, []platformconfig.Environment[APIServiceConfig]{{
			Name:   apiConfigObservabilityServiceName,
			Path:   path.Join(outputDir, stringOrDefault(s.APIServiceConfigPath, "api_service_config.json")),
			Config: s.RootConfig,
		}}, renderOpts...),
		platformconfig.RenderJSONFiles(ctx, []platformconfig.Environment[DBCleanerConfig]{{
			Name:   dbcConfigObservabilityServiceName,
			Path:   path.Join(outputDir, stringOrDefault(s.DBCleanerConfigPath, "job_db_cleaner_config.json")),
			Config: dbcConfig,
		}}, renderOpts...),
		platformconfig.RenderJSONFiles(ctx, []platformconfig.Environment[SchedulerConfig]{{
			Name:   schedulerConfigObservabilityServiceName,
			Path:   path.Join(outputDir, stringOrDefault(s.SchedulerConfigPath, "scheduler_config.json")),
			Config: schedulerConfig,
		}}, renderOpts...),
		platformconfig.RenderJSONFiles(ctx, []platformconfig.Environment[AsyncMessageHandlerConfig]{{
			Name:   amhConfigObservabilityServiceName,
			Path:   path.Join(outputDir, stringOrDefault(s.AsyncMessageHandlerConfigPath, "async_message_handler_config.json")),
			Config: amhConfig,
		}}, renderOpts...),
		platformconfig.RenderJSONFiles(ctx, []platformconfig.Environment[EmailDeliverabilityTestConfig]{{
			Name:   edtConfigObservabilityServiceName,
			Path:   path.Join(outputDir, stringOrDefault(s.EmailDeliverabilityTestConfigPath, "job_email_deliverability_test_config.json")),
			Config: edtConfig,
		}}, renderOpts...),
		platformconfig.RenderJSONFiles(ctx, []platformconfig.Environment[MCPServiceConfig]{{
			Name:   mcpConfigObservabilityServiceName,
			Path:   path.Join(outputDir, stringOrDefault(s.MCPServiceConfigPath, "mcp_server_config.json")),
			Config: mcpConfig,
		}}, renderOpts...),
	)

	return renderErr.ErrorOrNil()
}
