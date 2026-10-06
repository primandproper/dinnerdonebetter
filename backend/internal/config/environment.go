package config

import (
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	appentitlements "github.com/primandproper/dinnerdonebetter/backend/internal/entitlements"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"

	"github.com/primandproper/platform-go/v15/audit"
	auditcfg "github.com/primandproper/platform-go/v15/audit/config"
	oauth2database "github.com/primandproper/platform-go/v15/authentication/oauth2serverstore"
	oauth2servercfg "github.com/primandproper/platform-go/v15/authentication/oauth2serverstore/config"
	"github.com/primandproper/platform-go/v15/entitlements"
	entitlementscfg "github.com/primandproper/platform-go/v15/entitlements/config"
	"github.com/primandproper/platform-go/v15/links"
	linkscfg "github.com/primandproper/platform-go/v15/links/config"
	"github.com/primandproper/platform-go/v15/metering"
	meteringcfg "github.com/primandproper/platform-go/v15/metering/config"
	operationscfg "github.com/primandproper/platform-go/v15/operations/config"
	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/platform-go/v15/retention"
	retentioncfg "github.com/primandproper/platform-go/v15/retention/config"
	"github.com/primandproper/platform-go/v15/saga"
	sagacfg "github.com/primandproper/platform-go/v15/saga/config"
	"github.com/primandproper/platform-go/v15/service"
	waitlistsgrpc "github.com/primandproper/platform-go/v15/waitlists/grpc"
	platformconfig "github.com/primandproper/primitives-go/v2/config"
	databasecfg "github.com/primandproper/primitives-go/v2/database/config"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	distributedlockcfg "github.com/primandproper/primitives-go/v2/distributedlock/config"
	pglock "github.com/primandproper/primitives-go/v2/distributedlock/postgres"
	"github.com/primandproper/primitives-go/v2/jobs"
	jobscfg "github.com/primandproper/primitives-go/v2/jobs/config"
	"github.com/primandproper/primitives-go/v2/observability"
	retrycfg "github.com/primandproper/primitives-go/v2/retry/config"

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

// defaultJobsSchedulerConfig returns the scheduler every periodic job in the scheduler process runs
// on, and the lock that keeps each execution to one replica — this application's jobs and the
// reapers platform schedules for itself alike.
func defaultJobsSchedulerConfig() jobscfg.SchedulerConfig {
	return jobscfg.SchedulerConfig{
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
		Lock: defaultSchedulerLockConfig(),
	}
}

// defaultSchedulerLockConfig returns the lock backend the scheduler process serializes on.
//
// Postgres advisory locks: no new infrastructure, and a replica that dies drops its connection,
// which releases the lock without waiting for a TTL. The same backend backs the process's
// standalone locker, which the saga worker takes a per-instance scope of.
func defaultSchedulerLockConfig() distributedlockcfg.Config {
	return distributedlockcfg.Config{
		Provider: distributedlockcfg.PostgresProvider,
		Postgres: &pglock.Config{
			ConnWaitTimeout: 5 * time.Second,
		},
	}
}

// defaultScheduledJobsConfig returns the schedule for each of this application's periodic jobs,
// replacing what used to be one Kubernetes CronJob per job.
//
// Every LeaseTTL is set well above the job's observed worst case rather than near it: the lease
// is not renewed while a job runs, so a job that outlives its lease can be started a second time
// on another replica. Timeout is the shorter of the two — a job that hangs should be killed
// before its lease lapses, not after.
func defaultScheduledJobsConfig() ScheduledJobsConfig {
	return ScheduledJobsConfig{
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
		SearchDataIndexScheduler: jobscfg.JobConfig{
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
		QueueTest: jobscfg.JobConfig{
			Interval: 15 * time.Minute,
			Timeout:  time.Minute,
			LeaseTTL: 2 * time.Minute,
		},
		DataPrivacySweep: jobscfg.JobConfig{
			// Hourly is far finer than the seven-day artifact TTL needs, and a sweep with
			// nothing to do costs three indexed queries against partial indexes. It is also
			// the cadence the overdue gauge is sampled at, which is the reason not to make it
			// coarser: a deadline nobody is looking at is the same as no deadline.
			Interval:   time.Hour,
			Timeout:    5 * time.Minute,
			LeaseTTL:   10 * time.Minute,
			RunOnStart: true,
		},
		AuditRetentionSweeper: jobscfg.JobConfig{
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
		MeteringFlusher: jobscfg.JobConfig{
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
		TablePrefix: branding.TablePrefix,
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

// DefaultLinksConfig is the action-link registry: the two links a waitlist signup is mailed,
// pointed at the consumer web app served at webAppURL.
//
// A link's lifetime is a security policy rather than a deployment knob, which is why the
// registry is written out here and rendered into every environment's configuration rather than
// read from one. The confirmation link is short — an address that has not said yes in three
// days is one nobody should keep waiting on — and the unsubscribe link sits in an inbox until
// the day somebody wants off the list, so it lives for a year.
//
// The table prefix is not set here, and a rendered one is ignored: the store is built with the
// prefix the migration was rendered with — see internal/build/waitlists and
// docs/configuration.md.
func DefaultLinksConfig(webAppURL string) linkscfg.Config {
	cfg := linkscfg.Config{
		Actions: map[links.Action]links.ActionPolicy{
			waitlistsgrpc.ConfirmAction: {
				URL: webAppURL + "/waitlists/confirm?t=" + links.TokenPlaceholder,
				TTL: links.Duration(72 * time.Hour),
			},
			waitlistsgrpc.UnsubscribeAction: {
				URL: webAppURL + "/waitlists/unsubscribe?t=" + links.TokenPlaceholder,
				TTL: links.Duration(365 * 24 * time.Hour),
			},
		},
	}
	cfg.EnsureDefaults()

	return cfg
}

func DefaultMeteringConfig() meteringcfg.Config {
	cfg := meteringcfg.Config{
		Recorder: metering.RecorderConfig{
			BatchSize: metering.DefaultBatchSize,
			// AllowUnknownMeters stays false: every meter is recorded in the process
			// that registered it, from one build, so a record naming an unregistered
			// meter is a wiring mistake and Record should say so.
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

// DefaultEntitlementsConfig returns the plan catalog and the read path's knobs.
//
// Only the API server carries it. The plans are written out in full rather than left to the
// package defaults, because unlike every other section here there are no package defaults to
// fall back on: what a tier includes is this product's decision and nothing else can supply it,
// and a rendered config with no plans in it would deny every account with ErrUnknownPlan.
//
// Both plans grant every feature without a bound today. That is the same ordering
// internal/metering documents — count first, limit once the dashboards say what real usage looks
// like — and this is the file the first limit gets written into.
//
// FallbackPlan names the free tier deliberately. It is what a boolean feature check falls back
// to when the payments database cannot be reached, so an outage degrades a paying account to
// free rather than locking it out. It does not apply to quota features: their limit reaches
// metering through a QuotaSource that has no fallback, because the same source answers the exact
// path that records consumption, and enforcing a guessed limit there writes usage against a plan
// the customer is not on.
func DefaultEntitlementsConfig() entitlementscfg.Config {
	cfg := entitlementscfg.Config{
		Checker: entitlements.CheckerConfig{
			CachePrefix:  entitlements.DefaultCachePrefix,
			CacheTTL:     entitlements.DefaultCacheTTL,
			FallbackPlan: appentitlements.FreePlan,
		},
		Plans: appentitlements.DefaultPlans(),
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
		// New in v14: how long a quarantined row is kept before it is reaped. A
		// quarantined row is one the relay gave up on, so it is the one thing in this
		// table nobody has seen — the platform's thirty days is long enough that an
		// incident review a fortnight later still finds it.
		//
		// Spelled out rather than left to EnsureDefaults for the reason the saga config
		// below gives: these are rendered into files people read. EnsureDefaults does
		// fill it, so a process that omitted it would still run — but this struct is
		// validated as written, before any constructor sees it, so a blank one fails the
		// render rather than quietly becoming thirty days.
		QuarantineRetention: outbox.DefaultQuarantineRetention,
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
		PollInterval: time.Second,
		// New in v14: how often the worker samples the stuck level onto its gauge. Its own
		// knob rather than a multiple of PollInterval because the two reads cost different
		// things — a poll claims, and this one counts.
		StatsInterval:  saga.DefaultStatsInterval,
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

	// The mail platform's doors queue is the same third-party delivery, and is retried the same
	// way.
	queuedMail := standard()
	queuedMail.Retry.MaxAttempts = 4
	queuedMail.Retry.MaxDelay = 30 * time.Second

	// There is no webhook pool here any more. Outbound delivery is not a queue topic: a
	// dispatch row is claimed by the delivery worker, whose own concurrency, retry schedule,
	// and per-endpoint circuit breaking live in the webhooks config.

	return WorkerPoolsConfig{
		DeadLetterTopicName: DefaultDeadLetterTopicName,
		DataChanges:         dataChanges,
		OutboundEmails:      outboundEmails,
		SearchIndexRequests: standard(),
		MobileNotifications: standard(),
		QueuedMail:          queuedMail,
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
func databaseConfigForService(cfg *databasecfg.Config, users map[string]string, serviceName string) *databasecfg.Config {
	out := *cfg
	if username, ok := users[serviceName]; ok {
		out.ReadConnection.Username = username
		out.WriteConnection.Username = username
	}
	return &out
}

// observabilityFor returns obs with every pillar named for the given service.
func observabilityFor(base *observability.Config, serviceName string) observability.Config {
	obs := *base
	obs.Tracing.ServiceName = serviceName
	obs.Metrics.ServiceName = serviceName
	obs.Logging.ServiceName = serviceName
	obs.Profiling.ServiceName = serviceName

	return obs
}

// clone returns a pointer to a copy of what p points at, or nil.
//
// Every block a derived config takes from the API server's is cloned rather than shared, because
// service.Config holds its blocks by pointer and validating one normalizes and defaults them in
// place: a shared block would let rendering one process's file rewrite another's.
func clone[T any](p *T) *T {
	if p == nil {
		return nil
	}

	out := *p

	return &out
}

// schedulerShutdownTimeout bounds the scheduler process's whole shutdown: every loop it runs
// drains inside it, and a job mid-execution is the slowest of them.
const schedulerShutdownTimeout = 60 * time.Second

// apiShutdownTimeout bounds the API server's whole shutdown: draining both servers and releasing
// every client.
const apiShutdownTimeout = 10 * time.Second

const (
	apiConfigObservabilityServiceName       = "api_server"
	dbcConfigObservabilityServiceName       = "db_cleaner"
	schedulerConfigObservabilityServiceName = "scheduler"
	amhConfigObservabilityServiceName       = "async_message_handler"
	edtConfigObservabilityServiceName       = "email_deliverability_test"
	mcpConfigObservabilityServiceName       = "dinner_done_better_mcp_server"
)

// DerivedConfigs is every workload's configuration, as Derive builds them from one root.
type DerivedConfigs struct {
	DBCleaner               *DBCleanerConfig
	Scheduler               *SchedulerConfig
	AsyncMessageHandler     *AsyncMessageHandlerConfig
	EmailDeliverabilityTest *EmailDeliverabilityTestConfig
	MCPService              *MCPServiceConfig
}

// Derive builds every workload's configuration from RootConfig, the way Render writes them.
//
// It is Render's first half, exported so that a test which needs a valid configuration for a
// process other than the API server gets the one that ships rather than a fixture kept valid by
// hand. It names RootConfig's own service and observability on the way, as Render always has.
func (s *EnvironmentConfigSet) Derive() *DerivedConfigs {
	// Ensure API server config has the correct observability name before writing.
	s.RootConfig.Service.Name = apiConfigObservabilityServiceName
	// Ten seconds, the budget the API server has always shut down in: it drains two servers and
	// releases its clients, and a request still in flight past that is one the load balancer
	// stopped sending traffic to long before.
	if s.RootConfig.Service.ShutdownTimeout == 0 {
		s.RootConfig.Service.ShutdownTimeout = apiShutdownTimeout
	}
	s.RootConfig.Service.Observability = observabilityFor(&s.RootConfig.Service.Observability, apiConfigObservabilityServiceName)
	if s.RootConfig.Routing.Chi != nil {
		s.RootConfig.Routing.Chi.ServiceName = apiConfigObservabilityServiceName
	}

	root := &s.RootConfig.Service

	// Pinned in the rendered file as well as in code, so the file shows the tables the process
	// will actually use rather than a blank that platform's validation would refuse.
	dataprivacycfg.Pin(&s.RootConfig.Services.DataPrivacy.Platform)

	dbcObservability := observabilityFor(&root.Observability, dbcConfigObservabilityServiceName)
	disableWorkerOtelMetrics(&dbcObservability)

	dbcConfig := &DBCleanerConfig{
		Service: service.Config{
			Name:          dbcConfigObservabilityServiceName,
			Observability: dbcObservability,
			Database:      databaseConfigForService(root.Database, s.ServiceDatabaseUsers, dbcConfigObservabilityServiceName),
		},
		// This job sweeps the authorization server's tables, so it needs the prefix they
		// were created under and nothing else: a sweep asks the store for rows past their
		// deadlines, which needs neither an issuer nor a lifetime.
		OAuth2: oauth2servercfg.Config{
			Provider: oauth2servercfg.ProviderDatabase,
			Database: oauth2database.Config{TablePrefix: branding.TablePrefix},
		},
	}

	// One config for every interval-shaped periodic job, because they now share one process.

	// Copied out of the API server's config rather than pointed at, because service.Config
	// holds its blocks by pointer and validating it normalizes and defaults them in place: a
	// pointer into RootConfig would let rendering this file rewrite the API server's.
	// This process runs the operations worker, so it carries the whole tier. The API server
	// carries the same block for the enqueue-and-read half.
	schedulerOperations := DefaultOperationsConfig()
	// The same artifact storage the API server reads exports back from, so what this process
	// seals is what that one can open.
	schedulerDataPrivacy := s.RootConfig.Services.DataPrivacy.Platform
	dataprivacycfg.Pin(&schedulerDataPrivacy)
	// The same webhook configuration the API service writes with, so the worker claims from
	// the tables the dispatch rows are written into.
	schedulerWebhooks := s.RootConfig.Webhooks
	// Taken from the API server's config rather than rebuilt, so the tables the recorder writes
	// are by construction the tables the flusher flushes.
	schedulerMetering := s.RootConfig.Metering
	// Likewise the billing provider: the flusher posts through whichever one the payments
	// service was configured with, so enabling real usage billing is one provider setting
	// rather than two that can disagree.
	schedulerCapitalism := s.RootConfig.Services.Payments.Capitalism
	schedulerJobs := defaultJobsSchedulerConfig()
	schedulerLock := defaultSchedulerLockConfig()

	schedulerConfig := &SchedulerConfig{
		Service: service.Config{
			Name: schedulerConfigObservabilityServiceName,
			// Generous, because a job killed partway through has already done some of its
			// work and will redo it on the next tick, and the budget is shared by every
			// loop this process drains.
			ShutdownTimeout: schedulerShutdownTimeout,
			Observability:   observabilityFor(&root.Observability, schedulerConfigObservabilityServiceName),
			Database:        databaseConfigForService(root.Database, s.ServiceDatabaseUsers, schedulerConfigObservabilityServiceName),
			MessageQueue:    clone(root.MessageQueue),
			JobsScheduler:   &schedulerJobs,
			DistributedLock: &schedulerLock,
			Operations:      &schedulerOperations,
			DataPrivacy:     &schedulerDataPrivacy,
			Saga:            &sagacfg.Config{Worker: defaultSagaWorkerConfig()},
			Webhooks:        &schedulerWebhooks,
			Metering:        &schedulerMetering,
			Retention:       &retentioncfg.Config{Sweeper: defaultRetentionSweeperConfig()},
			Capitalism:      &schedulerCapitalism,
		},
		// The same sender the async message handler pushes through, so a device token this
		// process sends to is one that process would have sent to.
		PushNotifications:                s.RootConfig.PushNotifications,
		Search:                           s.RootConfig.TextSearch,
		Queues:                           s.RootConfig.Queues,
		DataPrivacyArtifactEncryptionKey: s.RootConfig.Services.DataPrivacy.ArtifactEncryptionKey,
		Jobs:                             defaultScheduledJobsConfig(),
		OutboxRelay:                      defaultOutboxRelayConfig(),
		AuditLog:                         defaultAuditSweeperConfig(),
	}

	amhEmail := s.RootConfig.Email
	amhAnalytics := s.RootConfig.Analytics

	amhConfig := &AsyncMessageHandlerConfig{
		Service: service.Config{
			Name:          amhConfigObservabilityServiceName,
			Observability: observabilityFor(&root.Observability, amhConfigObservabilityServiceName),
			Database:      databaseConfigForService(root.Database, s.ServiceDatabaseUsers, amhConfigObservabilityServiceName),
			MessageQueue:  clone(root.MessageQueue),
			// The same encoder the API server writes its messages with, which is not a
			// nicety: the handler decodes what the API published, and this section was
			// missing entirely once — a content type of "" is not a default, it is a
			// decoder that refuses to be built, so the process could not boot at all.
			Encoding:   clone(root.Encoding),
			HTTPClient: clone(root.HTTPClient),
			Email:      &amhEmail,
			Analytics:  &amhAnalytics,
		},
		Queues:            s.RootConfig.Queues,
		Search:            s.RootConfig.TextSearch,
		BaseURL:           s.RootConfig.BaseURL,
		Pools:             defaultWorkerPoolsConfig(),
		PushNotifications: s.RootConfig.PushNotifications,
	}

	edtConfig := &EmailDeliverabilityTestConfig{
		Observability:         root.Observability,
		Email:                 s.RootConfig.Email,
		RecipientEmailAddress: "verygoodsoftwarenotvirus@protonmail.com",
		ServiceEnvironment:    "prod",
	}
	edtConfig.Observability.Tracing.ServiceName = edtConfigObservabilityServiceName
	edtConfig.Observability.Metrics.ServiceName = edtConfigObservabilityServiceName
	edtConfig.Observability.Logging.ServiceName = edtConfigObservabilityServiceName
	edtConfig.Observability.Profiling.ServiceName = edtConfigObservabilityServiceName
	disableWorkerOtelMetrics(&edtConfig.Observability)

	mcpObservability := root.Observability
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

	mcpHTTPServer := *root.HTTPServer
	// The apple-app-site-association document describes the domain the iOS app is
	// associated with, which is the API's, not the MCP server's. Serving it from here
	// would publish an association for a host no Universal Link points at.
	mcpHTTPServer.AppleAppSiteAssociation = nil

	mcpConfig := &MCPServiceConfig{
		Database:      dbcfg.Config{Config: *databaseConfigForService(root.Database, s.ServiceDatabaseUsers, mcpConfigObservabilityServiceName)},
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
			Database: oauth2database.Config{TablePrefix: branding.TablePrefix},
		},
	}

	return &DerivedConfigs{
		DBCleaner:               dbcConfig,
		Scheduler:               schedulerConfig,
		AsyncMessageHandler:     amhConfig,
		EmailDeliverabilityTest: edtConfig,
		MCPService:              mcpConfig,
	}
}

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
	derived := s.Derive()
	dbcConfig, schedulerConfig, amhConfig, edtConfig, mcpConfig := derived.DBCleaner, derived.Scheduler,
		derived.AsyncMessageHandler, derived.EmailDeliverabilityTest, derived.MCPService

	switch {
	case strings.Contains(outputDir, "localdev"):
		edtConfig.ServiceEnvironment = "dev"
	case strings.Contains(outputDir, "testing"):
		edtConfig.ServiceEnvironment = "testing"
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
